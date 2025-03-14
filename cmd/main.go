package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	choice int
	url    string
)

func main() {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Ошибка при создании файла логов: %v\n", err)
		return
	}
	defer logFile.Close()
	logger := log.New(logFile, "log: ", log.Ldate|log.Ltime)

	printMenu()
	for {
		switch choice {
		case 4:
			os.Exit(1)
		case 3:
			readAndFormatLogs("app.log")
			resetMenu()
			printMenu()
		case 2:
			// Повторить последний запрос
			lastURL, err := getLastURLFromLogs("app.log")
			if err != nil {
				fmt.Println("Ошибка при получении последнего URL:", err)
				resetMenu()
				printMenu()
				continue
			}

			fmt.Printf("Повторяю запрос к URL: %s\n", lastURL)

			resp, err := http.Get(lastURL)
			if err != nil {
				fmt.Printf("Ошибка при выполнении запроса: %v\n", err)
				resetMenu()
				printMenu()
				continue
			}
			defer resp.Body.Close()

			// Получаем данные из ответа
			headers := resp.Header
			statusCode := resp.StatusCode
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Ошибка при чтении тела ответа: %v\n", err)
				resetMenu()
				printMenu()
				continue
			}

			headersText := formatHeaders(headers)
			prettyJSON, err := formatJSON(body)
			if err != nil {
				fmt.Printf("Ошибка при форматировании JSON: %v\n", err)
				resetMenu()
				printMenu()
				continue
			}

			// Выводим результат
			fmt.Println(statusCode)
			fmt.Println(headersText)
			fmt.Println(prettyJSON)

			// Логируем повторный запрос
			logger.Printf("%s %d", lastURL, statusCode)

			resetMenu()
			printMenu()
		case 1:
			// Создаем логгер, который пишет в файл
			styleUrl := lipgloss.NewStyle().Bold(true).
				Border(lipgloss.NormalBorder()).
				Foreground(lipgloss.Color("#FAFAFA")).
				Background(lipgloss.Color("#52E8d2")).
				Padding(0, 4)
			fmt.Println(styleUrl.Render("Введите URL:"))
			fmt.Scanln(&url)

			resp, err := http.Get(url)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			// Получаем данные из ответа
			headers := resp.Header
			statusCode := resp.StatusCode
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}

			headersText := formatHeaders(headers)
			prettyJSON, err := formatJSON(body)
			if err != nil {
				log.Fatalf("Ошибка при форматировании JSON: %v", err)
			}

			// Выводим результат
			fmt.Println(statusCode)
			fmt.Println(headersText)
			fmt.Println(prettyJSON)

			// Логируем запрос
			logger.Printf("%s %d", url, statusCode)

			resetMenu()
			printMenu()
		default:
			resetMenu()
			printMenu()
		}
	}
}

// Функция для получения последнего URL из логов
func getLastURLFromLogs(filePath string) (string, error) {
	// Открываем файл для чтения
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("ошибка при открытии файла: %v", err)
	}
	defer file.Close()

	// Создаем сканер для построчного чтения файла
	scanner := bufio.NewScanner(file)
	var lastLine string

	// Проходим по каждой строке файла
	for scanner.Scan() {
		lastLine = scanner.Text()
	}

	// Проверяем ошибки при сканировании
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("ошибка при чтении файла: %v", err)
	}

	// Разделяем последнюю строку на части (log: date url statuscode)
	parts := strings.Fields(lastLine)
	if len(parts) < 4 {
		return "", fmt.Errorf("некорректный формат последней строки: %s", lastLine)
	}

	// Извлекаем URL
	return parts[3], nil
}

func readAndFormatLogs(filePath string) {
	// Открываем файл для чтения
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Ошибка при открытии файла: %v", err)
	}
	defer file.Close()

	// Создаем сканер для построчного чтения файла
	scanner := bufio.NewScanner(file)

	// Переменная для хранения отформатированных логов
	var formattedLogs []string

	// Проходим по каждой строке файла
	for scanner.Scan() {
		line := scanner.Text()

		// Разделяем строку на части (log: date url statuscode)
		parts := strings.Fields(line)
		if len(parts) < 4 {
			fmt.Printf("Некорректный формат строки: %s\n", line)
			continue
		}

		// Извлекаем дату, URL и статус-код
		date := parts[1] + " " + parts[2] // Объединяем дату и время
		url := parts[3]
		statusCode := parts[4]

		// Форматируем строку в нужном стиле
		formattedLog := fmt.Sprintf("(%s) (%s) - (%s)", date, url, statusCode)
		formattedLogs = append(formattedLogs, formattedLog)
	}

	// Проверяем ошибки при сканировании
	if err := scanner.Err(); err != nil {
		log.Fatalf("Ошибка при чтении файла: %v", err)
	}

	// Выводим отформатированные логи
	for _, logLine := range formattedLogs {
		fmt.Println(logLine)
	}
}

func resetMenu() {
	choice = 0
}

func printMenu() {
	styleTitle := lipgloss.NewStyle().Bold(true).
		Border(lipgloss.NormalBorder()).
		AlignHorizontal(lipgloss.Center).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(1, 4)
	title := "My Requests"
	fmt.Println(styleTitle.Render(title))

	styleMenu := lipgloss.NewStyle().Bold(true).
		Border(lipgloss.NormalBorder()).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#52E8d2")).
		Padding(0, 4)
	menu := "Menu"
	fmt.Println(styleMenu.Render(menu))

	styleOptions := lipgloss.NewStyle().Bold(true).
		Border(lipgloss.NormalBorder()).
		Width(40).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#a75dd9")).
		Padding(0, 3)
	options := []string{
		"1. Do request",
		"2. Repeat last request",
		"3. History",
		"4. Exit",
	}
	for _, option := range options {
		fmt.Println(styleOptions.Render(option))
	}

	styleChoice := lipgloss.NewStyle().Bold(true)
	c := "Выберите:"
	fmt.Print(styleChoice.Render(c) + " ")
	fmt.Scan(&choice)
}

// Функция для форматирования заголовков
func formatHeaders(headers http.Header) string {
	var result string
	for key, values := range headers {
		result += fmt.Sprintf("%s: %s\n", key, values)
	}
	return result
}

func formatJSON(data []byte) (string, error) {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, data, "", "  ") // Форматируем JSON с отступами
	if err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}

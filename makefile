.PHONY: build
build:
	@go build -o ./bin/my-requests cmd/main.go

.PHONY: run
run: build
	@./bin/my-requests
.PHONY: build run tidy

build:
	go build -o bin/htui ./cmd/app

run:
	go run ./cmd/app

tidy:
	go mod tidy

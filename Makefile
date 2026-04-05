.PHONY: build run tidy test

build:
	go build -o bin/htui ./cmd/app

run:
	go run ./cmd/app

test:
	go test ./...

tidy:
	go mod tidy

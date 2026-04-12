BINARY := bin/htui
CMD    := ./cmd/app

.PHONY: run build test clean

run:
	go run $(CMD)

build:
	mkdir -p bin
	go build -o $(BINARY) $(CMD)

test:
	go test ./...

clean:
	rm -rf bin/

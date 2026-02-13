BINARY  := mcp-server-microsoft-todo
MODULE  := github.com/michMartineau/mcp-server-microsoft-todo
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: build clean tidy lint test

build:
	go build -ldflags="-s -w" -o $(BINARY) .

clean:
	rm -f $(BINARY)

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

test:
	go test ./...
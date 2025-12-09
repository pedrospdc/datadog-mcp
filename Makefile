.PHONY: build test lint clean run deps

BINARY_NAME=datadog-mcp
BUILD_DIR=./build

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/datadog-mcp

test:
	go test -v ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR)

run: build
	$(BUILD_DIR)/$(BINARY_NAME)

deps:
	go mod download
	go mod tidy

.PHONY: all build test lint clean check

BINARY_NAME=nova
BUILD_DIR=bin
MODULE=github.com/awanmh/Nova

all: check build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/nova

test:
	@echo "Running tests..."
	go test -v ./...

test-race:
	@echo "Running tests with race detector..."
	go test -race ./...

lint:
	@echo "Running static analysis..."
	go vet ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)

check: test lint

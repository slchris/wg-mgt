.PHONY: all build test lint security clean dev frontend help

BINARY_NAME=wg-mgt
BUILD_DIR=./build
GO_FILES=$(shell find . -name '*.go' -not -path "./vendor/*" -not -path "./web/*")

all: lint security test build

build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/wg-mgt

test:
	@echo "Running tests..."
	@go test -v -race -cover ./...

lint:
	@echo "Running linter..."
	@golangci-lint run --config .golangci.yml

security:
	@echo "Running security scan..."
	@gosec ./...

complexity:
	@echo "Checking complexity..."
	@gocyclo -over 15 $(GO_FILES)

frontend:
	@echo "Building frontend..."
	@cd web && npm install && npm run build

dev:
	@echo "Starting development server..."
	@go run ./cmd/wg-mgt

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f gosec-report.json

deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

help:
	@echo "Available targets:"
	@echo "  all       - Run lint, security, test, and build"
	@echo "  build     - Build the binary"
	@echo "  test      - Run tests"
	@echo "  lint      - Run golangci-lint"
	@echo "  security  - Run gosec security scan"
	@echo "  frontend  - Build frontend"
	@echo "  dev       - Start development server"
	@echo "  clean     - Clean build artifacts"
	@echo "  deps      - Download dependencies"

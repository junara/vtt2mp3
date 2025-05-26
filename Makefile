# Makefile for the project

.PHONY: all build build-windows build-linux test clean lint lint-fix

# Default target
all: lint build test

# Build the application
build:
	@echo "Building the application..."
	go build -o bin/vtt2mp3 ./cmd/vtt2mp3

# Build the application for Windows
build-windows:
	@echo "Building the application for Windows..."
	GOOS=windows GOARCH=amd64 go build -o bin/vtt2mp3.exe ./cmd/vtt2mp3

# Build the application for Linux
build-linux:
	@echo "Building the application for Linux..."
	GOOS=linux GOARCH=amd64 go build -o bin/vtt2mp3-linux ./cmd/vtt2mp3

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/

# Install golangci-lint if not installed
install-lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)

# Run golangci-lint
lint: install-lint
	@echo "Running golangci-lint..."
	golangci-lint run --no-config --timeout=5m ./...

# Run golangci-lint with auto-fix
lint-fix: install-lint
	@echo "Running golangci-lint with auto-fix..."
	golangci-lint run --no-config --timeout=5m --fix ./...

# Help command
help:
	@echo "Available targets:"
	@echo "  all           - Run lint, build, and test"
	@echo "  build         - Build the application"
	@echo "  build-windows - Build the application for Windows (creates bin/vtt2mp3.exe)"
	@echo "  build-linux   - Build the application for Linux (creates bin/vtt2mp3-linux)"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  lint          - Run golangci-lint"
	@echo "  lint-fix      - Run golangci-lint with auto-fix"
	@echo "  help          - Show this help message"

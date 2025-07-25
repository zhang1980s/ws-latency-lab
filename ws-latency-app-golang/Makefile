# Makefile for ws-latency-app-golang

# Variables
BINARY_NAME=ws-latency-app
VERSION=1.0.0
BUILD_DIR=build
MAIN_PATH=cmd/ws-latency-app/main.go
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

# Default target
.PHONY: all
all: clean build

# Clean build directory
.PHONY: clean
clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)
	@mkdir -p $(BUILD_DIR)

# Build for current platform
.PHONY: build
build:
	@echo "Building for current platform..."
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Binary built at $(BUILD_DIR)/$(BINARY_NAME)"

# Build for all supported platforms
.PHONY: build-all
build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64

# Build for Linux AMD64
.PHONY: build-linux-amd64
build-linux-amd64:
	@echo "Building for Linux AMD64..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@echo "Binary built at $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

# Build for Linux ARM64
.PHONY: build-linux-arm64
build-linux-arm64:
	@echo "Building for Linux ARM64..."
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	@echo "Binary built at $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64"

# Build for macOS AMD64
.PHONY: build-darwin-amd64
build-darwin-amd64:
	@echo "Building for macOS AMD64..."
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@echo "Binary built at $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64"

# Build for macOS ARM64
.PHONY: build-darwin-arm64
build-darwin-arm64:
	@echo "Building for macOS ARM64..."
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "Binary built at $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64"

# Build for Windows AMD64
.PHONY: build-windows-amd64
build-windows-amd64:
	@echo "Building for Windows AMD64..."
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Binary built at $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	@go mod download

# Run the application
.PHONY: run
run:
	@echo "Running application..."
	@go run $(MAIN_PATH)

# Create Docker image
.PHONY: docker
docker:
	@echo "Building Docker image..."
	@docker build -t ws-latency-app:$(VERSION) .

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all          : Clean and build for current platform"
	@echo "  clean        : Clean build directory"
	@echo "  build        : Build for current platform"
	@echo "  build-all    : Build for all supported platforms"
	@echo "  build-linux-amd64  : Build for Linux AMD64"
	@echo "  build-linux-arm64  : Build for Linux ARM64"
	@echo "  build-darwin-amd64 : Build for macOS AMD64"
	@echo "  build-darwin-arm64 : Build for macOS ARM64"
	@echo "  build-windows-amd64: Build for Windows AMD64"
	@echo "  test         : Run tests"
	@echo "  deps         : Install dependencies"
	@echo "  run          : Run the application"
	@echo "  docker       : Create Docker image"
	@echo "  help         : Show this help message"
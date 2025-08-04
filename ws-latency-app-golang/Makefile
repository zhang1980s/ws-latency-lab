# WebSocket Latency Test Application Makefile

# Variables
BINARY_NAME=ws-latency-app
GO=go
GOFLAGS=-v
MAIN_PACKAGE=./cmd/ws-latency-test
BUILD_DIR=build

# Default target
.PHONY: all
all: clean build

# Build the application
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@$(GO) build $(GOFLAGS) -o $(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Build complete: $(BINARY_NAME)"

# Build for multiple platforms
.PHONY: build-all
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	@GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)
	@echo "Multi-platform build complete. Binaries available in $(BUILD_DIR)/"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@$(GO) test -v ./...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	@$(GO) mod tidy
	@echo "Dependencies installed"

# Help target
.PHONY: help
help:
	@echo "WebSocket Latency Test Application Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make              Build the application (same as 'make build')"
	@echo "  make build        Build the application"
	@echo "  make build-all    Build for multiple platforms"
	@echo "  make test         Run tests"
	@echo "  make clean        Remove build artifacts"
	@echo "  make deps         Install dependencies"
	@echo "  make help         Show this help message"

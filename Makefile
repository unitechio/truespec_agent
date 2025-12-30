.PHONY: all build build-all test clean install

# Build variables
BINARY_NAME=agent
VERSION=1.0.0
BUILD_DIR=build
GO=go
GOFLAGS=-ldflags="-s -w -X main.version=$(VERSION)"

# Cross-compilation targets
PLATFORMS=windows/amd64 linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

all: test build

# Build for current platform
build:
	@echo "Building $(BINARY_NAME) for current platform..."
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/agent

# Cross-compile for all platforms
build-all:
	@echo "Cross-compiling for all platforms..."
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/} ./cmd/agent; \
	done

# Run tests
test:
	@echo "Running tests..."
	$(GO) test ./internal/... -cover -race

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	$(GO) test ./tests/integration/... -tags=integration

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GO) mod download
	$(GO) mod tidy

# Generate protobuf files
proto:
	@echo "Generating protobuf files..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		pkg/api/proto/*.proto

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Install agent (requires sudo on Linux/macOS)
install: build
	@echo "Installing agent..."
	@if [ "$(shell uname)" = "Windows_NT" ]; then \
		echo "Use installer for Windows"; \
	else \
		sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/your-agent; \
		sudo chmod +x /usr/local/bin/your-agent; \
	fi

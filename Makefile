.PHONY: build test lint clean install run

BINARY_NAME=fenrir
VERSION=$(shell git describe --tags --always 2>/dev/null || echo "v0.1.0")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build directories
BUILD_DIR=./bin
DIST_DIR=./dist

# Default target
all: clean test build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(GO_LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/fenrir

# Cross-compile for multiple platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(GO_LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/fenrir
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(GO_LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/fenrir
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(GO_LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/fenrir
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(GO_LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/fenrir
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(GO_LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/fenrir
	@echo "Done!"

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Run with hot reload (requires air)
dev:
	air

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -cover ./...

# Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Verify dependencies
verify:
	@echo "Verifying dependencies..."
	$(GOMOD) verify

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run ./...

# Vet code
vet:
	@echo "Vetting code..."
	$(GOCMD) vet ./...

# Build and install locally
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)

# Initialize a new project with fenrir
init:
	@echo "Initializing Fenrir..."
	./$(BUILD_DIR)/$(BINARY_NAME) init

# Generate mock data for testing
mock-data:
	@echo "Generating mock data..."
	$(GOCMD) run scripts/mock_data.go

# Database operations
db-reset:
	@echo "Resetting database..."
	rm -rf ~/.fenrir

# Run MCP server
mcp:
	./$(BUILD_DIR)/$(BINARY_NAME) mcp

# Run HTTP server
serve:
	./$(BUILD_DIR)/$(BINARY_NAME) serve

# Run TUI
tui:
	./$(BUILD_DIR)/$(BINARY_NAME) tui

# Release (requires goreleaser)
release:
	goreleaser release --config .goreleaser.yaml

# Snapshot release
snapshot:
	goreleaser release --config .goreleaser.yaml --snapshot --clean

# Help
help:
	@echo "Available targets:"
	@echo "  all          - Clean, test and build"
	@echo "  build        - Build the binary"
	@echo "  build-all    - Cross-compile for all platforms"
	@echo "  run          - Build and run"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage"
	@echo "  bench        - Run benchmarks"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download dependencies"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  vet          - Vet code"
	@echo "  install      - Build and install locally"
	@echo "  init         - Initialize Fenrir in project"
	@echo "  mcp          - Run MCP server"
	@echo "  serve        - Run HTTP server"
	@echo "  tui          - Run TUI"
	@echo "  release      - Create release"
	@echo "  help         - Show this help"

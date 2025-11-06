.PHONY: build install uninstall clean test run help init

# Binary name
BINARY_NAME=todo

# Installation directory
INSTALL_DIR=$(HOME)/.local/bin

# Build flags
LDFLAGS=-ldflags="-s -w"

# Default target
.DEFAULT_GOAL := help

help: ## Show this help message
	@echo '╔════════════════════════════════════════╗'
	@echo '║         Todo-Go Makefile               ║'
	@echo '╚════════════════════════════════════════╝'
	@echo ''
	@echo 'Usage:'
	@echo '  make <target>'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@go build $(LDFLAGS) -o $(BINARY_NAME) main.go
	@echo "✓ Build complete: ./$(BINARY_NAME)"

install: build ## Build and install to ~/.local/bin
	@echo "Installing to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@cp $(BINARY_NAME) $(INSTALL_DIR)/
	@chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✓ Installed to $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo ""
	@echo "Run 'todo init' to initialize your todo environment"

init: install ## Install and initialize the application
	@echo ""
	@echo "Initializing todo environment..."
	@$(INSTALL_DIR)/$(BINARY_NAME) init

uninstall: ## Remove the installed binary
	@echo "Uninstalling $(BINARY_NAME)..."
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✓ Uninstalled $(BINARY_NAME)"
	@echo ""
	@echo "Note: Todo data in ~/.todo has been preserved"
	@echo "To remove data: rm -rf ~/.todo"

clean: ## Remove build artifacts
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@echo "✓ Clean complete"

clean-all: clean ## Remove build artifacts and todo data
	@echo "Removing todo data..."
	@rm -rf $(HOME)/.todo
	@echo "✓ All data removed"

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

run: build ## Build and run the application
	@./$(BINARY_NAME)

dev: ## Run in development mode (with verbose logging)
	@LOG_LEVEL=debug go run main.go

# Build for multiple platforms
build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 main.go
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 main.go
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 main.go
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 main.go
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe main.go
	@echo "✓ Built for all platforms in ./dist/"

version: ## Show version information
	@echo "Todo-Go v1.3.0"
	@echo "Go version: $$(go version)"

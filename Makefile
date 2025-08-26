# RedTriage Makefile
# Build and package the RedTriage incident response triage tool

# Variables
BINARY_NAME = redtriage
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS = -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_DATE)

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod

# Directories
DIST_DIR = dist
BUILD_DIR = build
CMD_DIR = cmd/redtriage

# Platform-specific settings
UNAME_S := $(shell uname -s)
ifeq ($(OS),Windows_NT)
	BINARY_EXTENSION = .exe
	PLATFORM = windows
	ARCH = amd64
else
	BINARY_EXTENSION = 
	PLATFORM = $(shell uname -s | tr '[:upper:]' '[:lower:]')
	ARCH = $(shell uname -m)
endif

# Default target
.PHONY: all
all: clean build

# Build the binary
.PHONY: build
build: 
	@echo "Building RedTriage $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXTENSION) ./$(CMD_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXTENSION)"

# Build for specific platform
.PHONY: build-windows
build-windows:
	@echo "Building RedTriage for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME).exe ./$(CMD_DIR)

.PHONY: build-linux
build-linux:
	@echo "Building RedTriage for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

.PHONY: build-macos
build-macos:
	@echo "Building RedTriage for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

# Build all platforms
.PHONY: build-all
build-all: build-windows build-linux build-macos
	@echo "Multi-platform build complete"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@echo "Clean complete"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...
	@echo "Tests complete"

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies installed"

# Update dependencies
.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@echo "Dependencies updated"

# Lint and format code
.PHONY: lint
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping linting"; \
	fi

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...
	@echo "Code formatting complete"

# Package for distribution
.PHONY: package-windows
package-windows: build-windows
	@echo "Packaging for Windows..."
	@mkdir -p $(DIST_DIR)
	@cd $(BUILD_DIR) && zip -r ../$(DIST_DIR)/redtriage-$(VERSION)-windows-amd64.zip $(BINARY_NAME).exe
	@echo "Windows package created: $(DIST_DIR)/redtriage-$(VERSION)-windows-amd64.zip"

.PHONY: package-linux
package-linux: build-linux
	@echo "Packaging for Linux..."
	@mkdir -p $(DIST_DIR)
	@cd $(BUILD_DIR) && tar -czf ../$(DIST_DIR)/redtriage-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)
	@echo "Linux package created: $(DIST_DIR)/redtriage-$(VERSION)-linux-amd64.tar.gz"

.PHONY: package-macos
package-macos: build-macos
	@echo "Packaging for macOS..."
	@mkdir -p $(DIST_DIR)
	@cd $(BUILD_DIR) && tar -czf ../$(DIST_DIR)/redtriage-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)
	@echo "macOS package created: $(DIST_DIR)/redtriage-$(VERSION)-darwin-amd64.tar.gz"

# Package all platforms
.PHONY: package-all
package-all: package-windows package-linux package-macos
	@echo "All packages created in $(DIST_DIR)"

# Install locally
.PHONY: install
install: build
	@echo "Installing RedTriage..."
	@if [ "$(PLATFORM)" = "windows" ]; then \
		cp $(BUILD_DIR)/$(BINARY_NAME).exe /usr/local/bin/ || cp $(BUILD_DIR)/$(BINARY_NAME).exe $(GOPATH)/bin/; \
	else \
		cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/ || cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/; \
	fi
	@echo "Installation complete"

# Uninstall
.PHONY: uninstall
uninstall:
	@echo "Uninstalling RedTriage..."
	@rm -f /usr/local/bin/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME).exe
	@rm -f $(GOPATH)/bin/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME).exe
	@echo "Uninstallation complete"

# Run the binary
.PHONY: run
run: build
	@echo "Running RedTriage..."
	@$(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXTENSION) --help

# Development helpers
.PHONY: dev-setup
dev-setup: deps
	@echo "Setting up development environment..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint already installed"; \
	else \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2; \
	fi
	@echo "Development environment ready"

.PHONY: preflight
preflight: fmt lint test
	@echo "Preflight checks complete"

# Documentation
.PHONY: docs
docs:
	@echo "Generating documentation..."
	@if command -v godoc >/dev/null 2>&1; then \
		echo "Starting godoc server at http://localhost:6060"; \
		godoc -http=:6060; \
	else \
		echo "godoc not found, install with: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# Release preparation
.PHONY: release-prep
release-prep: clean build-all package-all
	@echo "Release preparation complete"
	@echo "Packages available in $(DIST_DIR):"
	@ls -la $(DIST_DIR)

# Help
.PHONY: help
help:
	@echo "RedTriage Makefile"
	@echo "=================="
	@echo ""
	@echo "Available targets:"
	@echo "  build          - Build for current platform"
	@echo "  build-windows  - Build for Windows"
	@echo "  build-linux    - Build for Linux"
	@echo "  build-macos    - Build for macOS"
	@echo "  build-all      - Build for all platforms"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  deps           - Install dependencies"
	@echo "  deps-update    - Update dependencies"
	@echo "  lint           - Lint code"
	@echo "  fmt            - Format code"
	@echo "  package-windows- Package for Windows"
	@echo "  package-linux  - Package for Linux"
	@echo "  package-macos  - Package for macOS"
	@echo "  package-all    - Package for all platforms"
	@echo "  install        - Install locally"
	@echo "  uninstall      - Uninstall"
	@echo "  run            - Build and run"
	@echo "  dev-setup      - Setup development environment"
	@echo "  preflight      - Run all checks (fmt, lint, test)"
	@echo "  docs           - Generate documentation"
	@echo "  release-prep   - Prepare release packages"
	@echo "  help           - Show this help"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION        - Version to build (default: git describe)"
	@echo "  COMMIT         - Git commit hash (default: git rev-parse)"
	@echo "  BUILD_DATE     - Build date (default: current UTC time)"
	@echo ""
	@echo "Examples:"
	@echo "  make build VERSION=v1.0.0"
	@echo "  make package-all VERSION=v1.0.0"
	@echo "  make release-prep VERSION=v1.0.0"

# Default target
.DEFAULT_GOAL := help

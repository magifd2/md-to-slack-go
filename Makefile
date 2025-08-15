# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

# Project details
BINARY_NAME=md-to-slack
OUTPUT_DIR=bin

# Versioning
VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-s -w -X 'main.version=$(VERSION)'"

# Installation paths
PREFIX?=/usr/local
BIN_DIR=$(PREFIX)/bin

.PHONY: all build clean test cross-compile install uninstall build-mac-universal build-linux build-windows package-all vulncheck help

help:
	@echo "Usage: make <command> [PREFIX=/path/to/install]"
	@echo ""
	@echo "Commands:"
	@echo "  all            : Builds for current OS/Arch and cross-compiles for all platforms."
	@echo "  build          : Builds the binary for the current OS and architecture."
	@echo "  test           : Runs all tests."
	@echo "  lint           : Runs linters (golangci-lint)."
	@echo "  vulncheck      : Runs vulnerability check (govulncheck)."
	@echo "  clean          : Cleans up build artifacts."
	@echo "  install        : Builds for the current architecture and installs the binary."
	@echo "  uninstall      : Uninstalls the binary."
	@echo "  cross-compile  : Cross-compiles for all target platforms (macOS, Linux, Windows)."
	@echo "  help           : Displays this help message."
	@echo ""
	@echo "Variables:"
	@echo "  PREFIX         : Installation prefix for 'install' and 'uninstall' commands."
	@echo "                   Defaults to /usr/local. Use PREFIX=~ for user-local installation."

all: vulncheck build cross-compile

# Build for the current OS/Arch
build:
	@echo "Building for $(shell go env GOOS)/$(shell go env GOARCH)..."
	@mkdir -p $(OUTPUT_DIR)/$(shell go env GOOS)-$(shell go env GOARCH)
	@$(GOBUILD) $(LDFLAGS) -o $(OUTPUT_DIR)/$(shell go env GOOS)-$(shell go env GOARCH)/$(BINARY_NAME) ./cmd/md-to-slack

# Run tests
test:
	@echo "Running tests..."
	@$(GOTEST) -v ./...

# Run linters
lint:
	@echo "Running linters..."
	@$(GOCMD) run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run ./...

# Run vulnerability check
vulncheck:
	@echo "Running vulnerability check..."
	@$(GOCMD) run golang.org/x/vuln/cmd/govulncheck@latest ./...

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	@$(GOCLEAN)
	@rm -f $(BINARY_NAME)
	@rm -rf $(OUTPUT_DIR)

# Install the binary
install: build
	@echo "Installing $(BINARY_NAME) to $(BIN_DIR)..."
	@mkdir -p $(BIN_DIR)
	@cp $(OUTPUT_DIR)/$(shell go env GOOS)-$(shell go env GOARCH)/$(BINARY_NAME) $(BIN_DIR)/
	@echo "Installation complete."

# Uninstall the binary
uninstall:
	@echo "Uninstalling $(BINARY_NAME) from $(BIN_DIR)..."
	@rm -f $(BIN_DIR)/$(BINARY_NAME)
	@echo "Uninstallation complete."

# Cross-compile for all target platforms
cross-compile: build-mac-universal build-linux build-windows package-all
	@echo "Cross-compilation and packaging finished. Release assets are in the $(OUTPUT_DIR)/ directory."

# Build for Linux (amd64)
build-linux:
	@echo "Building for Linux (amd64)..."
	@mkdir -p $(OUTPUT_DIR)/linux-amd64
	@GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(OUTPUT_DIR)/linux-amd64/$(BINARY_NAME) ./cmd/md-to-slack

# Build for Windows (amd64)
build-windows:
	@echo "Building for Windows (amd64)..."
	@mkdir -p $(OUTPUT_DIR)/windows-amd64
	@GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(OUTPUT_DIR)/windows-amd64/$(BINARY_NAME).exe ./cmd/md-to-slack

# Build macOS Universal Binary
build-mac-universal:
	@echo "Building for macOS (Universal)..."
	@mkdir -p $(OUTPUT_DIR)/darwin-universal
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/md-to-slack
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/md-to-slack
	@lipo -create -output $(OUTPUT_DIR)/darwin-universal/$(BINARY_NAME) $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-amd64 $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-arm64
	@codesign -s - $(OUTPUT_DIR)/darwin-universal/$(BINARY_NAME)
	@rm $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-amd64 $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-arm64
	@echo "Created Universal binary at $(OUTPUT_DIR)/darwin-universal/$(BINARY_NAME)"

# Package all binaries into archives
package-all: package-darwin package-linux package-windows

# Package macOS binary
package-darwin:
	@echo "Packaging macOS binary..."
	@cd $(OUTPUT_DIR)/darwin-universal && tar -czvf ../$(BINARY_NAME)-$(VERSION)-darwin-universal.tar.gz $(BINARY_NAME)

# Package Linux binary
package-linux:
	@echo "Packaging Linux binary..."
	@cd $(OUTPUT_DIR)/linux-amd64 && tar -czvf ../$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)

# Package Windows binary
package-windows:
	@echo "Packaging Windows binary..."
	@cd $(OUTPUT_DIR)/windows-amd64 && zip -r ../$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME).exe
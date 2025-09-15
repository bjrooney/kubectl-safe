# Makefile for kubectl-safe - Interactive safety net for dangerous kubectl commands
#
# This Makefile provides targets for building, testing, and installing the kubectl-safe plugin.
# The plugin is designed to be distributed via Krew (the kubectl plugin manager) but can also
# be installed manually.
#
# Available targets:
#   build       - Build the binary for the current platform (default target)
#   test        - Run all unit tests with verbose output
#   clean       - Remove all build artifacts
#   install     - Build and install to /usr/local/bin (requires sudo)
#   dev-install - Build and install to $HOME/.local/bin (for development)
#
# Usage examples:
#   make               # Build the binary
#   make test          # Run tests
#   make dev-install   # Install for development use

.PHONY: build test clean install

# Binary name - must match the kubectl plugin naming convention
BINARY_NAME=kubectl-safe

# Directory where compiled binaries are placed
BUILD_DIR=bin

# Default target - builds the binary for the current platform
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/kubectl-safe

# Run all unit tests with verbose output to see detailed test results
test:
	@echo "Running tests..."
	go test -v ./pkg/...

# Clean up all build artifacts
clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)

# Install the binary to /usr/local/bin for system-wide access
# Note: This requires sudo privileges
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

# Install the binary to $HOME/.local/bin for development use
# This is useful for testing without requiring sudo privileges
# Make sure $HOME/.local/bin is in your PATH
dev-install: build
	@echo "Installing $(BINARY_NAME) to $$HOME/.local/bin..."
	@mkdir -p $$HOME/.local/bin
	cp $(BUILD_DIR)/$(BINARY_NAME) $$HOME/.local/bin/

# Default target when just running 'make'
.DEFAULT_GOAL := build
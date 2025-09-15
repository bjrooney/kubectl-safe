.PHONY: build test clean install version

BINARY_NAME=kubectl-safe
BUILD_DIR=bin
VERSION_FILE=VERSION

# Read version from VERSION file, or use "dev" if file doesn't exist
VERSION := $(shell cat $(VERSION_FILE) 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/bjrooney/kubectl-safe/pkg/safe.Version=$(VERSION)"

build:
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/kubectl-safe

version:
	@echo "Current version: $(VERSION)"

# Increment patch version
bump-patch:
	@current=$$(cat $(VERSION_FILE) 2>/dev/null || echo "0.0.0"); \
	major=$$(echo $$current | cut -d. -f1); \
	minor=$$(echo $$current | cut -d. -f2); \
	patch=$$(echo $$current | cut -d. -f3); \
	new_patch=$$((patch + 1)); \
	new_version="$$major.$$minor.$$new_patch"; \
	echo "$$new_version" > $(VERSION_FILE); \
	echo "Version bumped from $$current to $$new_version"

test:
	@echo "Running tests..."
	go test -v ./pkg/...

clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)

install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

# For development
dev-install: build
	@echo "Installing $(BINARY_NAME) to $$HOME/.local/bin..."
	@mkdir -p $$HOME/.local/bin
	cp $(BUILD_DIR)/$(BINARY_NAME) $$HOME/.local/bin/

.DEFAULT_GOAL := build
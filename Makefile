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

# Cross-platform build and packaging for Krew release assets
DIST_DIR=dist

build-all: clean
	@echo "Building release assets for all platforms..."
	@mkdir -p $(DIST_DIR)
	GOOS=linux   GOARCH=amd64  go build $(LDFLAGS) -o $(DIST_DIR)/kubectl-safe-linux-amd64/kubectl-safe ./cmd/kubectl-safe
	GOOS=linux   GOARCH=arm64  go build $(LDFLAGS) -o $(DIST_DIR)/kubectl-safe-linux-arm64/kubectl-safe ./cmd/kubectl-safe
	GOOS=darwin  GOARCH=amd64  go build $(LDFLAGS) -o $(DIST_DIR)/kubectl-safe-darwin-amd64/kubectl-safe ./cmd/kubectl-safe
	GOOS=darwin  GOARCH=arm64  go build $(LDFLAGS) -o $(DIST_DIR)/kubectl-safe-darwin-arm64/kubectl-safe ./cmd/kubectl-safe
	GOOS=windows GOARCH=amd64  go build $(LDFLAGS) -o $(DIST_DIR)/kubectl-safe-windows-amd64/kubectl-safe.exe ./cmd/kubectl-safe

	@echo "Packaging release tarballs..."
	cd $(DIST_DIR)/kubectl-safe-linux-amd64   && tar -czvf ../kubectl-safe-linux-amd64.tar.gz   kubectl-safe
	cd $(DIST_DIR)/kubectl-safe-linux-arm64   && tar -czvf ../kubectl-safe-linux-arm64.tar.gz   kubectl-safe
	cd $(DIST_DIR)/kubectl-safe-darwin-amd64  && tar -czvf ../kubectl-safe-darwin-amd64.tar.gz  kubectl-safe
	cd $(DIST_DIR)/kubectl-safe-darwin-arm64  && tar -czvf ../kubectl-safe-darwin-arm64.tar.gz  kubectl-safe
	cd $(DIST_DIR)/kubectl-safe-windows-amd64 && tar -czvf ../kubectl-safe-windows-amd64.tar.gz kubectl-safe.exe

	@echo "Release assets are ready in $(DIST_DIR)"

.DEFAULT_GOAL := build
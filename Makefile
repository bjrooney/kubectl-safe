.PHONY: build test clean install

BINARY_NAME=kubectl-safe
BUILD_DIR=bin

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/kubectl-safe

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
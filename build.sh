#!/bin/bash
# Updated to support Apple Silicon (arm64) Macs.

set -e

DIST_DIR="dist"
echo "Creating distribution directory..."
mkdir -p "$DIST_DIR"

# Build for Linux (Intel/AMD)
echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -o "$DIST_DIR/kubectl-safe-linux-amd64" main.go

# Build for macOS (Intel)
echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -o "$DIST_DIR/kubectl-safe-darwin-amd64" main.go

# Build for macOS (Apple Silicon)
echo "Building for macOS (arm64)..."
GOOS=darwin GOARCH=arm64 go build -o "$DIST_DIR/kubectl-safe-darwin-arm64" main.go

# Build for Windows (Intel/AMD)
echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -o "$DIST_DIR/kubectl-safe-windows-amd64.exe" main.go

echo "Build complete! Binaries are in the '$DIST_DIR/' directory."
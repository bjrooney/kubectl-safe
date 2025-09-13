#!/bin/bash
# Updated to build AND package the binaries into .tar.gz and .zip archives.

set -e

DIST_DIR="dist"
echo "Cleaning and creating distribution directory..."
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

# --- Build and Package for Linux ---
echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -o "$DIST_DIR/kubectl-safe" main.go
echo "Packaging for Linux (amd64)..."
tar -czf "$DIST_DIR/kubectl-safe-linux-amd64.tar.gz" -C "$DIST_DIR" kubectl-safe
rm "$DIST_DIR/kubectl-safe" # Clean up the raw binary

# --- Build and Package for macOS (Intel) ---
echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -o "$DIST_DIR/kubectl-safe" main.go
echo "Packaging for macOS (amd64)..."
tar -czf "$DIST_DIR/kubectl-safe-darwin-amd64.tar.gz" -C "$DIST_DIR" kubectl-safe
rm "$DIST_DIR/kubectl-safe"

# --- Build and Package for macOS (Apple Silicon) ---
echo "Building for macOS (arm64)..."
GOOS=darwin GOARCH=arm64 go build -o "$DIST_DIR/kubectl-safe" main.go
echo "Packaging for macOS (arm64)..."
tar -czf "$DIST_DIR/kubectl-safe-darwin-arm64.tar.gz" -C "$DIST_DIR" kubectl-safe
rm "$DIST_DIR/kubectl-safe"

# --- Build and Package for Windows ---
echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -o "$DIST_DIR/kubectl-safe.exe" main.go
echo "Packaging for Windows (amd64)..."
zip -j "$DIST_DIR/kubectl-safe-windows-amd64.zip" "$DIST_DIR/kubectl-safe.exe"
rm "$DIST_DIR/kubectl-safe.exe"

echo "Build and packaging complete! Archives are in the '$DIST_DIR/' directory."
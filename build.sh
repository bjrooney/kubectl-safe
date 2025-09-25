#!/bin/bash
# Updated to build AND package the binaries into .tar.gz and .zip archives with version support.

set -e

# Read version from VERSION file
if [ -f "VERSION" ]; then
    VERSION=$(cat VERSION)
else
    VERSION="dev"
fi

echo "Building kubectl-safe version $VERSION"

DIST_DIR="dist"
echo "Cleaning and creating distribution directory..."
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

# Build flags with version
LDFLAGS="-X github.com/bjrooney/kubectl-safe/pkg/safe.Version=$VERSION"

# --- Build and Package for Linux ---
echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "$DIST_DIR/kubectl-safe" ./cmd/kubectl-safe
echo "Packaging for Linux (amd64)..."
tar -czf "$DIST_DIR/kubectl-safe-linux-amd64.tar.gz" -C "$DIST_DIR" kubectl-safe
rm "$DIST_DIR/kubectl-safe" # Clean up the raw binary

# --- Build and Package for Linux ARM64 ---
echo "Building for Linux (arm64)..."
GOOS=linux GOARCH=arm64 go build -ldflags "$LDFLAGS" -o "$DIST_DIR/kubectl-safe" ./cmd/kubectl-safe
echo "Packaging for Linux (arm64)..."
tar -czf "$DIST_DIR/kubectl-safe-linux-arm64.tar.gz" -C "$DIST_DIR" kubectl-safe
rm "$DIST_DIR/kubectl-safe"

# --- Build and Package for macOS (Intel) ---
echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "$DIST_DIR/kubectl-safe" ./cmd/kubectl-safe
echo "Packaging for macOS (amd64)..."
tar -czf "$DIST_DIR/kubectl-safe-darwin-amd64.tar.gz" -C "$DIST_DIR" kubectl-safe
rm "$DIST_DIR/kubectl-safe"

# --- Build and Package for macOS (Apple Silicon) ---
echo "Building for macOS (arm64)..."
GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o "$DIST_DIR/kubectl-safe" ./cmd/kubectl-safe
echo "Packaging for macOS (arm64)..."
tar -czf "$DIST_DIR/kubectl-safe-darwin-arm64.tar.gz" -C "$DIST_DIR" kubectl-safe
rm "$DIST_DIR/kubectl-safe"

# --- Build and Package for Windows ---
echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "$DIST_DIR/kubectl-safe.exe" ./cmd/kubectl-safe
echo "Packaging for Windows (amd64)..."
zip -j "$DIST_DIR/kubectl-safe-windows-amd64.zip" "$DIST_DIR/kubectl-safe.exe"
rm "$DIST_DIR/kubectl-safe.exe"

echo "Build and packaging complete! Archives are in the '$DIST_DIR/' directory."
echo "Version: $VERSION"
# Copy safe.yaml to plugins directory
echo "Copying safe.yaml to plugins directory..."
cp "$DIST_DIR/safe.yaml" "plugins/safe.yaml"
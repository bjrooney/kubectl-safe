#!/bin/bash
#
# build.sh - Multi-platform build script for kubectl-safe
#
# This script builds the kubectl-safe binary for multiple platforms and architectures,
# then packages them into compressed archives suitable for distribution via GitHub releases
# and Krew plugin installation.
#
# The script creates the following artifacts:
#   - kubectl-safe-linux-amd64.tar.gz   (Linux 64-bit Intel/AMD)
#   - kubectl-safe-darwin-amd64.tar.gz  (macOS Intel)
#   - kubectl-safe-darwin-arm64.tar.gz  (macOS Apple Silicon)
#   - kubectl-safe-windows-amd64.zip    (Windows 64-bit)
#
# Each archive contains a single binary file named appropriately for the platform.
# The archives are placed in the 'dist' directory which is created/cleaned by this script.
#
# Usage:
#   ./build.sh
#
# Requirements:
#   - Go 1.19 or later
#   - tar and zip utilities
#   - Cross-compilation support (automatically available in recent Go versions)
#
# Environment variables that can be set to customize behavior:
#   DIST_DIR: Directory for build artifacts (default: "dist")
#
# Exit codes:
#   0: Success
#   1: Build failure

# Exit immediately if any command exits with a non-zero status
# This ensures we don't continue with partial builds
set -e

# Distribution directory for build artifacts
DIST_DIR="${DIST_DIR:-dist}"

echo "kubectl-safe multi-platform build starting..."
echo "Target directory: $DIST_DIR"

# Clean up any previous builds and create fresh distribution directory
echo "Cleaning and creating distribution directory..."
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

# Build for Linux AMD64 (most common server architecture)
echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -o "$DIST_DIR/kubectl-safe" ./cmd/kubectl-safe
echo "Packaging Linux binary..."
tar -czf "$DIST_DIR/kubectl-safe-linux-amd64.tar.gz" -C "$DIST_DIR" kubectl-safe
rm "$DIST_DIR/kubectl-safe" # Clean up the raw binary to avoid confusion

# Build for Linux ARM64 (increasingly common, especially on Apple Silicon and some servers)
echo "Building for Linux (arm64)..."
GOOS=linux GOARCH=arm64 go build -o "$DIST_DIR/kubectl-safe" ./cmd/kubectl-safe
echo "Packaging Linux ARM64 binary..."
tar -czf "$DIST_DIR/kubectl-safe-linux-arm64.tar.gz" -C "$DIST_DIR" kubectl-safe
rm "$DIST_DIR/kubectl-safe"

# Build for macOS Intel (traditional Mac architecture)
echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -o "$DIST_DIR/kubectl-safe" ./cmd/kubectl-safe
echo "Packaging macOS Intel binary..."
tar -czf "$DIST_DIR/kubectl-safe-darwin-amd64.tar.gz" -C "$DIST_DIR" kubectl-safe
rm "$DIST_DIR/kubectl-safe"

# Build for macOS Apple Silicon (M1/M2/M3 Macs)
echo "Building for macOS (arm64)..."
GOOS=darwin GOARCH=arm64 go build -o "$DIST_DIR/kubectl-safe" ./cmd/kubectl-safe
echo "Packaging macOS Apple Silicon binary..."
tar -czf "$DIST_DIR/kubectl-safe-darwin-arm64.tar.gz" -C "$DIST_DIR" kubectl-safe
rm "$DIST_DIR/kubectl-safe"

# Build for Windows (use .exe extension as required on Windows)
echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -o "$DIST_DIR/kubectl-safe.exe" ./cmd/kubectl-safe
echo "Packaging Windows binary..."
# Use zip format for Windows as it's more commonly supported there
zip -j "$DIST_DIR/kubectl-safe-windows-amd64.zip" "$DIST_DIR/kubectl-safe.exe"
rm "$DIST_DIR/kubectl-safe.exe"

echo ""
echo "✅ Build and packaging complete!"
echo "📦 Archives created in the '$DIST_DIR/' directory:"
ls -la "$DIST_DIR"/*.tar.gz "$DIST_DIR"/*.zip 2>/dev/null || true
echo ""
echo "Next steps:"
echo "  1. Run './checksum.sh' to generate SHA256 checksums for Krew manifest"
echo "  2. Test the binaries on target platforms"
echo "  3. Create a GitHub release with these artifacts"
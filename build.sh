#!/bin/bash
# A simple script to cross-compile the plugin for all target platforms.

# Exit immediately if a command exits with a non-zero status.
set -e

# Create a directory to store the binaries, if it doesn't exist.
echo "Creating distribution directory..."
mkdir -p dist

# Build for Linux
echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -o dist/kubectl-safe-linux main.go

# Build for macOS
echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -o dist/kubectl-safe-darwin main.go

# Build for Windows
echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -o dist/kubectl-safe-windows.exe main.go

echo "Build complete! Binaries are in the 'dist/' directory."
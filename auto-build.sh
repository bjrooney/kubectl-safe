#!/bin/bash
# Auto-build script that bumps patch version, builds, and prepares for release

set -e

echo "=== kubectl-safe Auto-Build ==="

# Bump patch version
echo "1. Bumping patch version..."
make bump-patch

# Get the new version
VERSION=$(cat VERSION)
echo "New version: $VERSION"

# Run tests
echo "2. Running tests..."
make test

# Build the application
echo "3. Building kubectl-safe..."
make build

# Build distribution packages
echo "4. Building distribution packages..."
./build.sh

# Generate checksums
echo "5. Generating checksums..."
./checksum.sh

echo "=== Build Complete ==="
echo "Version: $VERSION"
echo "Distribution files available in dist/"
echo ""
echo "Next steps:"
echo "1. Create a git tag: git tag v$VERSION"
echo "2. Push tag: git push origin v$VERSION"
echo "3. Create GitHub release with files from dist/"
echo "4. Update safe.yaml with new version and checksums"
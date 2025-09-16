#!/bin/bash
# Update safe.yaml with new version and checksums from GitHub release

set -e

if [ $# -ne 1 ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 0.1.2"
    exit 1
fi

VERSION="$1"
REPO="bjrooney/kubectl-safe"
BASE_URL="https://github.com/$REPO/releases/download/v$VERSION"

echo "Updating safe.yaml for version $VERSION..."

# Update version in safe.yaml
sed -i "s/version: v.*/version: v$VERSION/" safe.yaml

# Function to get SHA256 from GitHub release
get_sha256() {
    local platform="$1"
    local filename="kubectl-safe-$platform.tar.gz"
    local url="$BASE_URL/$filename"
    
    echo "Fetching checksum for $filename..."
    # Download file temporarily to calculate hash
    curl -L -s "$url" | sha256sum | cut -d' ' -f1
}

# Update URLs and checksums for each platform
echo "Updating platform URLs and checksums..."

# Linux AMD64
LINUX_AMD64_SHA=$(get_sha256 "linux-amd64")
sed -i "/os: linux/,/arch: amd64/{
    s|uri: .*|uri: $BASE_URL/kubectl-safe-linux-amd64.tar.gz|
    s|sha256: .*|sha256: \"$LINUX_AMD64_SHA\"|
}" safe.yaml

# Linux ARM64
LINUX_ARM64_SHA=$(get_sha256 "linux-arm64")
sed -i "/os: linux/,/arch: arm64/{
    s|uri: .*|uri: $BASE_URL/kubectl-safe-linux-arm64.tar.gz|
    s|sha256: .*|sha256: \"$LINUX_ARM64_SHA\"|
}" safe.yaml

# Darwin AMD64
DARWIN_AMD64_SHA=$(get_sha256 "darwin-amd64")
sed -i "/os: darwin/,/arch: amd64/{
    s|uri: .*|uri: $BASE_URL/kubectl-safe-darwin-amd64.tar.gz|
    s|sha256: .*|sha256: \"$DARWIN_AMD64_SHA\"|
}" safe.yaml

# Darwin ARM64
DARWIN_ARM64_SHA=$(get_sha256 "darwin-arm64")
sed -i "/os: darwin/,/arch: arm64/{
    s|uri: .*|uri: $BASE_URL/kubectl-safe-darwin-arm64.tar.gz|
    s|sha256: .*|sha256: \"$DARWIN_ARM64_SHA\"|
}" safe.yaml

# Windows AMD64
WINDOWS_AMD64_SHA=$(get_sha256 "windows-amd64")
sed -i "/os: windows/,/arch: amd64/{
    s|uri: .*|uri: $BASE_URL/kubectl-safe-windows-amd64.tar.gz|
    s|sha256: .*|sha256: \"$WINDOWS_AMD64_SHA\"|
}" safe.yaml

echo "safe.yaml updated successfully!"
echo ""
echo "Updated checksums:"
echo "  Linux AMD64:   $LINUX_AMD64_SHA"
echo "  Linux ARM64:   $LINUX_ARM64_SHA"
echo "  Darwin AMD64:  $DARWIN_AMD64_SHA"
echo "  Darwin ARM64:  $DARWIN_ARM64_SHA"
echo "  Windows AMD64: $WINDOWS_AMD64_SHA"
echo ""
echo "Review the changes in safe.yaml and commit if correct."
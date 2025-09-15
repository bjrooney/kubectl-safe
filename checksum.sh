#!/bin/bash
#
# checksum.sh - Generate SHA256 checksums for Krew plugin manifest
#
# This script generates SHA256 checksums for the compiled kubectl-safe binaries
# and formats them for easy copying into the Krew plugin manifest (safe.yaml).
# 
# The checksums are required by Krew to verify the integrity of downloaded
# plugin binaries. Each platform-specific binary archive must have its
# checksum specified in the manifest.
#
# Usage:
#   ./checksum.sh
#
# Prerequisites:
#   - Run ./build.sh first to create the binary archives
#   - The 'sha256sum' utility must be available (standard on Linux/macOS)
#
# Output:
#   Formatted SHA256 checksums ready to paste into safe.yaml
#
# Environment variables:
#   DIST_DIR: Directory containing build artifacts (default: "dist")

# Exit immediately if a command exits with a non-zero status
set -e

# The directory where build.sh places the compiled binaries
DIST_DIR="${DIST_DIR:-dist}"

# Array of the binary archive names we need to process
# This matches the artifacts created by build.sh and expected by the release workflow
BINARIES=(
  "kubectl-safe-linux-amd64.tar.gz"
  "kubectl-safe-linux-arm64.tar.gz"
  "kubectl-safe-darwin-amd64.tar.gz"
  "kubectl-safe-darwin-arm64.tar.gz"
  "kubectl-safe-windows-amd64.zip"
)

echo "🔐 Generating SHA256 Checksums for Krew Manifest"
echo "============================================="
echo ""

# Verify that the distribution directory exists
if [ ! -d "$DIST_DIR" ]; then
  echo "❌ Error: Distribution directory '$DIST_DIR' not found"
  echo "   Please run './build.sh' first to create the binaries."
  exit 1
fi

echo "📁 Checking binaries in: $DIST_DIR"
echo ""

# Process each binary archive
for binary in "${BINARIES[@]}"; do
  file_path="$DIST_DIR/$binary"

  # Verify that the binary file exists
  if [ ! -f "$file_path" ]; then
    echo "❌ Error: Binary not found at '$file_path'"
    echo "   Please run './build.sh' first to create the binaries."
    exit 1
  fi

  # Calculate the SHA256 checksum
  # awk '{print $1}' extracts just the checksum hash from sha256sum output
  checksum=$(sha256sum "$file_path" | awk '{print $1}')

  # Display file info
  file_size=$(stat -c%s "$file_path" 2>/dev/null || stat -f%z "$file_path" 2>/dev/null || echo "unknown")
  echo "📦 ${binary}"
  echo "   Size: ${file_size} bytes"
  echo "   SHA256: ${checksum}"
  echo ""
done

echo "✅ Checksum generation complete!"
echo ""
echo "📋 Copy these values into your safe.yaml manifest:"
echo "================================================="

# Generate formatted output for the YAML manifest
for binary in "${BINARIES[@]}"; do
  file_path="$DIST_DIR/$binary"
  if [ -f "$file_path" ]; then
    checksum=$(sha256sum "$file_path" | awk '{print $1}')
    echo "# ${binary}"
    printf "sha256: \"%s\"\n\n" "$checksum"
  fi
done

echo "💡 Tips:"
echo "  - Update the version number in safe.yaml if needed"
echo "  - Verify the download URLs match your GitHub release"
echo "  - Test the plugin installation: kubectl krew install --manifest=safe.yaml"
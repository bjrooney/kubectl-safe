#!/bin/bash
#
# This script generates the SHA256 checksums for the Krew plugin binaries
# and formats them for easy copying into the plugin manifest (e.g., safe.yaml).

# Exit immediately if a command exits with a non-zero status.
set -e

# The directory where build.sh places the compiled binaries.
DIST_DIR="dist"

# An array of the binary names we need to process.
# This makes it easy to add more platforms in the future.
BINARIES=(
  "kubectl-safe-linux-amd64.tar.gz"
  "kubectl-safe-darwin-arm64.tar.gz"
  "kubectl-safe-darwin-amd64.tar.gz"
  "kubectl-safe-windows-amd64.zip"
)

echo "--- Generating SHA256 Checksums for Krew Manifest ---"
echo ""

# Loop through each binary in our list.
for binary in "${BINARIES[@]}"; do
  file_path="$DIST_DIR/$binary"

  # First, check if the binary file actually exists.
  if [ ! -f "$file_path" ]; then
    echo "❌ Error: Binary not found at '$file_path'"
    echo "   Please run './build.sh' first to create the binaries."
    exit 1
  fi

  # Calculate the checksum.
  # awk '{print $1}' extracts just the checksum hash from the command's output.
  checksum=$(sha256sum "$file_path" | awk '{print $1}')

  # Print the nicely formatted output for the YAML file.
  echo "# Checksum for ${binary}"
  printf "sha256: \"%s\"\n\n" "$checksum"
done

echo "--- ✅ Done! Copy the sha256 lines into your safe.yaml ---"
#!/bin/bash
#
# This script generates the SHA256 checksums for the Krew plugin binaries
# and updates them in the specified YAML file or prints them for manual copying.

# Exit immediately if a command exits with a non-zero status.
set -e

# The directory where build.sh places the compiled binaries.
DIST_DIR="dist"

# Check if a YAML file was provided as argument
YAML_FILE=""
if [ $# -eq 1 ]; then
    YAML_FILE="$1"
    echo "--- Updating SHA256 Checksums in $YAML_FILE ---"
else
    echo "--- Generating SHA256 Checksums for Krew Manifest ---"
fi

echo ""

# Function to update SHA256 in YAML file for a specific platform
update_sha256_in_yaml() {
    local file="$1"
    local platform="$2"
    local arch="$3"
    local checksum="$4"
    local filename="$5"
    
    # Use awk to find and update the sha256 for the specific platform
    awk -v os="$platform" -v arch="$arch" -v sha="$checksum" '
    BEGIN { in_platform = 0; found_os = 0; found_arch = 0 }
    /- selector:/ { in_platform = 1; found_os = 0; found_arch = 0; next }
    in_platform && /os: / && $2 == os { found_os = 1; next }
    in_platform && /arch: / && $2 == arch && found_os { found_arch = 1; next }
    in_platform && /sha256:/ && found_os && found_arch { 
        print "    sha256: \"" sha "\""
        in_platform = 0
        found_os = 0
        found_arch = 0
        next 
    }
    /- selector:/ && in_platform { in_platform = 0; found_os = 0; found_arch = 0 }
    { print }
    ' "$file" > "${file}.tmp" && mv "${file}.tmp" "$file"
}

# An array of the binary names and their corresponding platform info
declare -A BINARIES
BINARIES["kubectl-safe-linux-amd64.tar.gz"]="linux amd64"
BINARIES["kubectl-safe-linux-arm64.tar.gz"]="linux arm64"
BINARIES["kubectl-safe-darwin-amd64.tar.gz"]="darwin amd64"
BINARIES["kubectl-safe-darwin-arm64.tar.gz"]="darwin arm64"
BINARIES["kubectl-safe-windows-amd64.zip"]="windows amd64"

# Loop through each binary in our list.
for binary in "${!BINARIES[@]}"; do
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
  
  # Get platform and architecture
  platform_arch=(${BINARIES[$binary]})
  platform=${platform_arch[0]}
  arch=${platform_arch[1]}

  if [ -n "$YAML_FILE" ]; then
    echo "Updating checksum for $platform-$arch: $checksum"
    update_sha256_in_yaml "$YAML_FILE" "$platform" "$arch" "$checksum" "$binary"
  else
    # Print the nicely formatted output for manual copying to YAML file.
    echo "# Checksum for ${binary} ($platform-$arch)"
    printf "sha256: \"%s\"\n\n" "$checksum"
  fi
done

if [ -n "$YAML_FILE" ]; then
  echo "--- ✅ Done! Checksums updated in $YAML_FILE ---"
else
  echo "--- ✅ Done! Copy the sha256 lines into your safe.yaml ---"
fi
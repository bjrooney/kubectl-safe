#!/bin/bash
#
# update-manifest.sh - Update Krew manifest with checksums and version
#
# This script automates the process of updating the safe.yaml Krew manifest
# with the correct version and SHA256 checksums from built release artifacts.
# It's designed to be run after building release artifacts but before submitting
# to the Krew index.
#
# Usage:
#   ./update-manifest.sh [version]
#
# Parameters:
#   version: Optional version string (e.g., "v1.2.3"). If not provided,
#            the script will attempt to detect it from git tags or prompt for input.
#
# Prerequisites:
#   - Release artifacts must exist in the DIST_DIR (run ./build.sh first)
#   - The yq utility for YAML manipulation (or manual editing as fallback)
#
# Environment variables:
#   DIST_DIR: Directory containing build artifacts (default: "dist")
#   DRY_RUN: If set to "true", shows what would be changed without modifying files

set -e

# Configuration
DIST_DIR="${DIST_DIR:-dist}"
MANIFEST_FILE="safe.yaml"
DRY_RUN="${DRY_RUN:-false}"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Utility functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Function to detect version from git tags
detect_version() {
    if git describe --tags --exact-match HEAD 2>/dev/null; then
        return 0
    elif git describe --tags --abbrev=0 2>/dev/null; then
        log_warning "Using latest tag (not on a tagged commit)"
        return 0
    else
        return 1
    fi
}

# Function to get checksums for all platform archives
get_checksums() {
    local platform_checksums=""
    
    # Array of expected platform archives
    local platforms=(
        "linux-amd64:kubectl-safe-linux-amd64"
        "linux-arm64:kubectl-safe-linux-arm64"  
        "darwin-amd64:kubectl-safe-darwin-amd64"
        "darwin-arm64:kubectl-safe-darwin-arm64"
        "windows-amd64:kubectl-safe-windows-amd64.exe"
    )
    
    log_info "Calculating checksums for platform archives..."
    
    for platform_info in "${platforms[@]}"; do
        IFS=':' read -r platform binary_name <<< "$platform_info"
        archive_file="$DIST_DIR/kubectl-safe-${platform}.tar.gz"
        
        if [[ "$platform" == "windows-amd64" ]]; then
            archive_file="$DIST_DIR/kubectl-safe-windows-amd64.zip"
        fi
        
        if [ ! -f "$archive_file" ]; then
            log_error "Archive not found: $archive_file"
            log_error "Please run './build.sh' first to create release artifacts"
            exit 1
        fi
        
        local checksum=$(sha256sum "$archive_file" | awk '{print $1}')
        log_info "  $platform: $checksum"
        
        # Store for later use
        eval "checksum_${platform//-/_}=\"$checksum\""
    done
}

# Function to update manifest (manual approach if yq is not available)
update_manifest_manual() {
    local version="$1"
    local backup_file="${MANIFEST_FILE}.backup"
    
    log_info "Creating backup: $backup_file"
    cp "$MANIFEST_FILE" "$backup_file"
    
    if [ "$DRY_RUN" = "true" ]; then
        log_warning "DRY RUN: Would update $MANIFEST_FILE with:"
        echo "  Version: $version"
        echo "  Linux AMD64 checksum: $checksum_linux_amd64"
        echo "  Linux ARM64 checksum: $checksum_linux_arm64"
        echo "  Darwin AMD64 checksum: $checksum_darwin_amd64"
        echo "  Darwin ARM64 checksum: $checksum_darwin_arm64"
        echo "  Windows AMD64 checksum: $checksum_windows_amd64"
        return 0
    fi
    
    # Update version
    sed -i.tmp "s/version: v[0-9].*/version: $version/" "$MANIFEST_FILE"
    
    # Update download URLs
    sed -i.tmp "s|/download/v[0-9][^/]*/|/download/$version/|g" "$MANIFEST_FILE"
    
    # Update checksums (this is a simplified approach - for production use yq or manual editing)
    log_warning "Manual checksum update required!"
    log_info "Please update the sha256 values in $MANIFEST_FILE with these checksums:"
    echo ""
    echo "Linux AMD64:   sha256: \"$checksum_linux_amd64\""
    echo "Linux ARM64:   sha256: \"$checksum_linux_arm64\""
    echo "Darwin AMD64:  sha256: \"$checksum_darwin_amd64\""
    echo "Darwin ARM64:  sha256: \"$checksum_darwin_arm64\""
    echo "Windows AMD64: sha256: \"$checksum_windows_amd64\""
    echo ""
    
    # Clean up temp files
    rm -f "${MANIFEST_FILE}.tmp"
    
    log_success "Version and URLs updated in $MANIFEST_FILE"
    log_warning "Remember to manually update the SHA256 checksums!"
}

# Main execution
main() {
    local version="$1"
    
    log_info "kubectl-safe Krew Manifest Updater"
    log_info "===================================="
    
    # Verify prerequisites
    if [ ! -f "$MANIFEST_FILE" ]; then
        log_error "Manifest file not found: $MANIFEST_FILE"
        exit 1
    fi
    
    if [ ! -d "$DIST_DIR" ]; then
        log_error "Distribution directory not found: $DIST_DIR"
        log_error "Please run './build.sh' first to create release artifacts"
        exit 1
    fi
    
    # Determine version
    if [ -z "$version" ]; then
        log_info "No version specified, attempting to detect from git..."
        if version=$(detect_version); then
            log_success "Detected version: $version"
        else
            log_warning "Could not detect version from git tags"
            echo -n "Please enter the version (e.g., v1.2.3): "
            read -r version
            if [ -z "$version" ]; then
                log_error "Version is required"
                exit 1
            fi
        fi
    fi
    
    # Validate version format
    if [[ ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+.*$ ]]; then
        log_error "Invalid version format: $version"
        log_error "Expected format: v1.2.3 or v1.2.3-beta1"
        exit 1
    fi
    
    log_info "Using version: $version"
    
    # Get checksums for all platforms
    get_checksums
    
    # Update the manifest
    update_manifest_manual "$version"
    
    echo ""
    log_success "Manifest update process completed!"
    echo ""
    log_info "Next steps:"
    echo "  1. Review the changes in $MANIFEST_FILE"
    echo "  2. Manually update SHA256 checksums if not done automatically"
    echo "  3. Test the manifest: kubectl krew install --manifest=$MANIFEST_FILE"
    echo "  4. Submit a PR to kubernetes-sigs/krew-index"
    echo ""
    log_info "For Krew index submission, see:"
    echo "  https://krew.sigs.k8s.io/docs/developer-guide/plugin-manifest/"
}

# Execute main function with all arguments
main "$@"
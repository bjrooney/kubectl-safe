# Versioning and Release Process

This document describes the automated versioning and release process for kubectl-safe.

## Version Management

The project uses semantic versioning (MAJOR.MINOR.PATCH) stored in the `VERSION` file.

### Automatic Version Bumping

1. **Patch Version Bumping**: Use `make bump-patch` to increment the patch version
2. **Manual Version Setting**: Edit the `VERSION` file directly for major/minor version changes

### Building with Version

All builds automatically embed the version from the `VERSION` file into the binary:

```bash
# Standard build
make build

# Check current version
make version

# Build all distribution packages
./build.sh
```

## Automated Build Process

Use the `auto-build.sh` script for a complete build cycle:

```bash
./auto-build.sh
```

This script:
1. Bumps the patch version
2. Runs tests
3. Builds the binary
4. Creates distribution packages for all platforms
5. Generates checksums

## Release Process

### Creating a Release

1. **Automated Build**: Run `./auto-build.sh` to prepare the release
2. **Create Git Tag**: Create and push a git tag matching the version
   ```bash
   git tag v$(cat VERSION)
   git push origin v$(cat VERSION)
   ```
3. **GitHub Actions**: The push of a tag triggers the release workflow automatically

### GitHub Actions Workflows

- **Build Workflow** (`.github/workflows/build.yml`): Runs on every push/PR to main
- **Release Workflow** (`.github/workflows/release.yml`): Runs on tag push, creates GitHub release

### Updating Krew Plugin Manifest

After a GitHub release is created, update the krew plugin manifest:

```bash
# Update safe.yaml with new version and checksums
./update-safe-yaml.sh $(cat VERSION)
```

This script:
- Downloads binaries from the GitHub release
- Calculates SHA256 checksums
- Updates `safe.yaml` with new version and checksums

## Manual Process

If you need to handle releases manually:

1. **Build**: `./auto-build.sh`
2. **Tag**: `git tag v$(cat VERSION) && git push origin v$(cat VERSION)`
3. **Release**: Create GitHub release with files from `dist/`
4. **Update Krew**: `./update-safe-yaml.sh $(cat VERSION)`

## Files

- `VERSION`: Contains the current version number
- `auto-build.sh`: Complete automated build script
- `build.sh`: Multi-platform build script
- `checksum.sh`: Generates SHA256 checksums
- `update-safe-yaml.sh`: Updates krew manifest
- `safe.yaml`: Krew plugin manifest
# Krew Index Publishing Automation

This document describes the automated publishing system for the kubectl-safe plugin to the `@bjrooney/krew-index` repository.

## Overview

The kubectl-safe project now includes automated publishing to a custom Krew index with version control and change tracking. This ensures that new releases are automatically made available through Krew while maintaining a proper audit trail.

## How It Works

### 1. Release Workflow
When a new tag is pushed (e.g., `v1.2.3`), the release workflow:
1. Builds binaries for all supported platforms
2. Creates release archives (`.tar.gz` files)
3. Generates SHA256 checksums for all archives
4. Creates a GitHub release with all assets and checksums
5. Triggers the Krew index publishing workflow

### 2. Krew Publishing Workflow
The Krew publishing workflow (`.github/workflows/publish-krew.yml`):
1. Downloads the release assets from the GitHub release
2. Calculates SHA256 checksums for verification
3. Generates a changelog from git commits since the last release
4. Updates the Krew plugin manifest (`safe.yaml`) with new version and checksums
5. Creates a pull request to the `@bjrooney/krew-index` repository

### 3. Version Control & Change Tracking
Each update to the Krew index includes:
- **Version information**: Extracted from git tags
- **Changelog**: Generated from commit messages between releases
- **Technical details**: SHA256 checksums, platform support, release URLs
- **Automated PR**: Proper description with all relevant information

## Setup Requirements

### Required Secrets
The following GitHub secret must be configured in the kubectl-safe repository:

- `KREW_INDEX_TOKEN`: A GitHub Personal Access Token with repository access to `bjrooney/krew-index`

### Repository Structure
The automation expects the following repository structure for `bjrooney/krew-index`:
```
krew-index/
├── plugins/
│   └── safe.yaml    # Krew plugin manifest
└── README.md        # Index documentation
```

## Usage

### Automatic Publishing (Recommended)
1. Create and push a new tag:
   ```bash
   git tag v1.2.3
   git push origin v1.2.3
   ```
2. The release workflow will automatically trigger
3. A few minutes later, the Krew publishing workflow will create a PR to the index repository
4. Review and merge the PR to make the new version available

### Manual Publishing
You can also trigger publishing manually:
1. Go to the "Actions" tab in the kubectl-safe repository
2. Select "Publish to Krew Index"
3. Click "Run workflow"
4. Enter the version to publish (e.g., `v1.2.3`)

## Workflow Files

### `.github/workflows/release.yml`
- **Trigger**: Git tags matching `v*`
- **Purpose**: Build binaries, create GitHub release, trigger Krew publishing
- **Enhanced**: Now includes checksum generation and Krew publishing trigger

### `.github/workflows/publish-krew.yml`
- **Triggers**: Release creation, repository dispatch, manual trigger
- **Purpose**: Update Krew index with new plugin version
- **Features**: Automatic changelog, checksum verification, PR creation

## Changelog Generation

The automation generates changelogs by analyzing git commits between releases:
- Commits since the previous tag are included
- Merge commits are excluded to keep the changelog clean
- Each commit becomes a bullet point in the changelog

For better changelog quality, use descriptive commit messages following conventional commits format.

## Troubleshooting

### Common Issues

1. **Missing KREW_INDEX_TOKEN**
   - Ensure the secret is configured with proper repository access
   - Token needs read/write access to the krew-index repository

2. **Checksum Verification Failures**
   - Usually indicates a problem with the release assets
   - Check that all platform binaries were built correctly

3. **PR Creation Failures**
   - Verify the krew-index repository exists and is accessible
   - Check that the repository structure matches expectations

### Manual Recovery

If the automation fails, you can manually update the index:
1. Download the release assets
2. Calculate SHA256 checksums: `sha256sum *.tar.gz`
3. Update `plugins/safe.yaml` with new version and checksums
4. Create a PR to the krew-index repository

## Security Considerations

- The automation uses minimal required permissions
- The GitHub token is scoped only to the krew-index repository
- All operations are logged and auditable through GitHub Actions
- Checksums ensure integrity of published binaries

## Monitoring

Monitor the automation through:
- GitHub Actions logs in the kubectl-safe repository
- Pull requests created in the krew-index repository
- Release notes and changelogs
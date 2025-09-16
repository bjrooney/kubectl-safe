# Krew Index Setup Example

This document provides step-by-step instructions for setting up the `@bjrooney/krew-index` repository to work with the automated publishing workflow.

## 1. Create the Krew Index Repository

Create a new repository named `krew-index` under the `bjrooney` organization/user:

```bash
# Using GitHub CLI
gh repo create bjrooney/krew-index --public --description "Custom Krew index for kubectl-safe plugin"

# Or create manually through GitHub web interface
```

## 2. Initialize Repository Structure

Clone and set up the basic structure:

```bash
git clone https://github.com/bjrooney/krew-index.git
cd krew-index

# Create directory structure
mkdir -p plugins

# Create initial README
cat > README.md << 'EOF'
# @bjrooney/krew-index

Custom Krew index for kubectl plugins.

## Available Plugins

- **safe**: Interactive safety net for dangerous kubectl commands
  - Repository: https://github.com/bjrooney/kubectl-safe
  - Installation: `kubectl krew index add bjrooney https://github.com/bjrooney/krew-index.git && kubectl krew install bjrooney/safe`

## Usage

Add this index to your Krew installation:

```bash
kubectl krew index add bjrooney https://github.com/bjrooney/krew-index.git
kubectl krew install bjrooney/safe
```

## Updates

This index is automatically updated when new versions of plugins are released.
EOF

# Create initial plugin manifest (will be overwritten by automation)
cat > plugins/safe.yaml << 'EOF'
apiVersion: krew.googlecode.com/v1alpha2
kind: Plugin
metadata:
  name: safe
spec:
  version: v0.1.0
  homepage: https://github.com/bjrooney/kubectl-safe
  shortDescription: Interactive safety net for dangerous kubectl commands
  description: |
    kubectl-safe provides an interactive safety net for dangerous kubectl commands.
    
    It acts as a wrapper around destructive kubectl operations to prevent common
    mistakes by requiring explicit --context and --namespace flags and showing
    interactive confirmation prompts.
  platforms:
  - selector:
      matchLabels:
        os: linux
        arch: amd64
    uri: https://github.com/bjrooney/kubectl-safe/releases/download/v0.1.0/kubectl-safe-linux-amd64.tar.gz
    sha256: ""
    bin: kubectl-safe
  # Additional platforms will be added by automation
EOF

# Commit initial structure
git add .
git commit -m "Initial krew-index setup"
git push origin main
```

## 3. Configure GitHub Token

Create a Personal Access Token with repository permissions:

1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Give it a descriptive name: "kubectl-safe krew-index automation"
4. Set expiration as needed
5. Select scopes:
   - `repo` (Full control of private repositories)
   - `workflow` (Update GitHub Action workflows)

6. Copy the token

## 4. Add Secret to kubectl-safe Repository

In the `kubectl-safe` repository:

1. Go to Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Name: `KREW_INDEX_TOKEN`
4. Value: Paste the Personal Access Token from step 3
5. Click "Add secret"

## 5. Test the Automation

### Option A: Create a Test Release

```bash
# In kubectl-safe repository
git tag v0.1.1
git push origin v0.1.1
```

This will trigger:
1. Release workflow (builds binaries, creates GitHub release)
2. Krew publishing workflow (updates index, creates PR)

### Option B: Manual Trigger

1. Go to kubectl-safe repository → Actions
2. Select "Publish to Krew Index" workflow
3. Click "Run workflow"
4. Enter version: `v0.1.0`
5. Click "Run workflow"

## 6. Verify Setup

After running the workflow:

1. Check the krew-index repository for a new pull request
2. Review the PR content (version, checksums, changelog)
3. Merge the PR to publish the plugin

Users can then install with:
```bash
kubectl krew index add bjrooney https://github.com/bjrooney/krew-index.git
kubectl krew install bjrooney/safe
```

## Troubleshooting

### Common Issues

1. **Permission denied when pushing to krew-index**
   - Verify `KREW_INDEX_TOKEN` has correct permissions
   - Ensure token hasn't expired

2. **Workflow fails to download release assets**
   - Check that the release was created successfully
   - Verify all platform binaries are present

3. **Checksum mismatches**
   - Usually indicates build issues in release workflow
   - Check release workflow logs

### Manual Recovery

If automation fails, you can manually update the index:

```bash
# Clone krew-index
git clone https://github.com/bjrooney/krew-index.git
cd krew-index

# Download and verify checksums
VERSION="v1.2.3"
curl -L -o /tmp/kubectl-safe-linux-amd64.tar.gz \
  "https://github.com/bjrooney/kubectl-safe/releases/download/${VERSION}/kubectl-safe-linux-amd64.tar.gz"
sha256sum /tmp/kubectl-safe-linux-amd64.tar.gz

# Update plugins/safe.yaml with new version and checksums
# Create PR manually
```
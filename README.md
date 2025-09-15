# kubectl-safe

A Krew plugin that provides an interactive safety net for dangerous kubectl commands.

## Overview

kubectl-safe acts as a simple, interactive wrapper around destructive kubectl commands to prevent common, high-impact mistakes. It's designed to be a final checkpoint before you make a change you might regret.

## Features

- **Enforces Best Practices**: Requires the mandatory use of `--context` and `--namespace` flags, forcing you to be explicit about your target
- **Interactive Confirmation**: Shows detailed information about the target cluster and namespace before executing dangerous commands
- **Transparent Pass-through**: Safe commands like `get`, `describe`, `logs` etc. are passed through without any checks
- **Comprehensive Coverage**: Protects against dangerous operations including delete, apply, create, replace, patch, edit, scale, rollout, drain, cordon, uncordon, and taint

## Installation

### Via Krew (Recommended)

```bash
kubectl krew install safe
```

### Manual Installation

1. Download the latest release from the [releases page](https://github.com/bjrooney/kubectl-safe/releases)
2. Extract the binary and place it in your PATH
3. Ensure the binary is named `kubectl-safe`

### Build from Source

```bash
git clone https://github.com/bjrooney/kubectl-safe.git
cd kubectl-safe
make build
# Binary will be available at bin/kubectl-safe
```

## Usage

Replace your dangerous kubectl commands with `kubectl safe`:

```bash
# Instead of: kubectl delete pod mypod
kubectl safe delete pod mypod --context=prod --namespace=myapp

# Instead of: kubectl apply -f deployment.yaml  
kubectl safe apply -f deployment.yaml --context=staging --namespace=myapp
```

### Examples

```bash
# This will prompt for confirmation and show target details
kubectl safe delete deployment myapp --context=production --namespace=default

# This will fail - missing required flags
kubectl safe delete pod mypod
# Error: dangerous command requires explicit --context and --namespace flag(s)

# Safe commands pass through without checks
kubectl safe get pods
kubectl safe describe deployment myapp
```

## Dangerous Commands

The following kubectl commands trigger safety checks:

- `delete` - Delete resources
- `apply` - Apply configuration changes
- `create` - Create new resources  
- `replace` - Replace existing resources
- `patch` - Patch existing resources
- `edit` - Edit resources in-place
- `scale` - Scale deployments/replicasets
- `rollout` - Manage rollouts
- `drain` - Drain nodes
- `cordon` - Cordon nodes
- `uncordon` - Uncordon nodes  
- `taint` - Taint nodes

## Safety Features

When executing a dangerous command, kubectl-safe will:

1. **Validate Required Flags**: Ensure both `--context` and `--namespace` are provided
2. **Show Target Information**: Display the target cluster context and namespace
3. **Request Confirmation**: Ask for explicit confirmation before proceeding
4. **Execute Safely**: Only proceed if the user confirms with "yes" or "y"

Example safety prompt:

```
⚠️  DANGEROUS COMMAND DETECTED ⚠️

You are about to execute: kubectl delete pod mypod --context=prod --namespace=default

Target Details:
  Context:   prod
  Namespace: default

This operation may cause data loss or service disruption.
Are you sure you want to continue? (yes/no):
```

## Development

```bash
# Run tests
make test

# Build binary
make build

# Install locally for development  
make dev-install
```

## Krew Index Publishing

This project includes automated publishing to `@bjrooney/krew-index` with version control and change tracking. See [docs/krew-publishing.md](docs/krew-publishing.md) for details on how the automation works and how to set it up.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

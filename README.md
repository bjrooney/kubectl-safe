# kubectl-safe

A Krew plugin that provides an interactive safety net for modifying kubectl commands.

## Overview

kubectl-safe acts as a simple, interactive wrapper around modifying kubectl commands to prevent common, high-impact mistakes. It's designed to be a final checkpoint before you make a change you might regret.

## Features

- **Enforces Best Practices**: Requires the mandatory use of `--context` and `--namespace` flags, forcing you to be explicit about your target
- **Context Validation**: Validates that the specified context exists in your kubeconfig to prevent typos and targeting non-existent clusters
- **Interactive Confirmation**: Shows detailed information about the target cluster and namespace before executing modifying commands
- **Transparent Pass-through**: Safe commands like `get`, `describe`, `logs` etc. are passed through without any checks
- **Comprehensive Coverage**: Protects against 13 always-modifying commands, 4 commands with specific modifying subcommands, and 2 commands that are modifying only with certain flags (see [Commands Covered](#commands-covered) for complete details)

## Installation

### Via Krew (Recommended)

Install using krew with the local plugin manifest:

```bash
kubectl krew install --manifest=https://raw.githubusercontent.com/bjrooney/kubectl-safe/main/safe.yaml
```

Or if you have cloned the repository:

```bash
kubectl krew install --manifest=safe.yaml
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

Replace your modifying kubectl commands with `kubectl safe`:

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
# Error: modifying command requires explicit --context and --namespace flag(s)

# This will fail - invalid context
kubectl safe delete pod mypod --context=invalid-cluster --namespace=default
# Error: context 'invalid-cluster' not found in kubeconfig. Available contexts: production, staging, development

# Safe commands pass through without checks
kubectl safe get pods
kubectl safe describe deployment myapp
```

## Commands Covered

kubectl-safe provides safety checks for the following kubectl commands that can modify cluster state:

### Always-Modifying Commands

These commands always require safety checks regardless of arguments:

| Command | Description | Example |
|---------|-------------|---------|
| `apply` | Apply configuration changes | `kubectl safe apply -f deployment.yaml --context=prod --namespace=myapp` |
| `create` | Create new resources | `kubectl safe create deployment myapp --image=nginx --context=prod --namespace=myapp` |
| `delete` | Delete resources | `kubectl safe delete pod mypod --context=prod --namespace=myapp` |
| `edit` | Edit resources in-place | `kubectl safe edit deployment myapp --context=prod --namespace=myapp` |
| `patch` | Patch existing resources | `kubectl safe patch deployment myapp -p '{"spec":{"replicas":3}}' --context=prod --namespace=myapp` |
| `replace` | Replace existing resources | `kubectl safe replace -f deployment.yaml --context=prod --namespace=myapp` |
| `scale` | Scale deployments/replicasets | `kubectl safe scale deployment myapp --replicas=5 --context=prod --namespace=myapp` |
| `cordon` | Mark node as unschedulable | `kubectl safe cordon node1 --context=prod` |
| `uncordon` | Mark node as schedulable | `kubectl safe uncordon node1 --context=prod` |
| `drain` | Drain node for maintenance | `kubectl safe drain node1 --ignore-daemonsets --context=prod` |
| `taint` | Add/remove taints from nodes | `kubectl safe taint nodes node1 key=value:NoSchedule --context=prod` |
| `autoscale` | Create horizontal pod autoscaler | `kubectl safe autoscale deployment myapp --min=1 --max=10 --context=prod --namespace=myapp` |
| `expose` | Expose resource as service | `kubectl safe expose deployment myapp --port=80 --context=prod --namespace=myapp` |

### Commands with Modifying Subcommands

These commands only require safety checks when used with specific subcommands:

| Command | Modifying Subcommands | Example |
|---------|----------------------|---------|
| `auth` | `reconcile` | `kubectl safe auth reconcile -f rbac.yaml --context=prod --namespace=myapp` |
| `certificate` | `approve`, `deny` | `kubectl safe certificate approve csr-12345 --context=prod` |
| `rollout` | `restart`, `undo` | `kubectl safe rollout restart deployment/myapp --context=prod --namespace=myapp` |
| `set` | `env`, `image`, `resources`, `selector`, `subject`, `serviceaccount` | `kubectl safe set image deployment/myapp container=image:v2 --context=prod --namespace=myapp` |

### Commands Modifying by Flag

These commands only require safety checks when used with specific flags:

| Command | Modifying Flags | Description | Example |
|---------|----------------|-------------|---------|
| `annotate` | `-f`, `--filename` | Annotate resources from file | `kubectl safe annotate -f deployment.yaml key=value --context=prod --namespace=myapp` |
| `label` | `-f`, `--filename` | Label resources from file | `kubectl safe label -f deployment.yaml app=myapp --context=prod --namespace=myapp` |

### Safe Commands (Pass-through)

These commands are considered safe and pass through without any safety checks:

- `get` - Retrieve resources
- `describe` - Show detailed resource information  
- `logs` - View container logs
- `exec` - Execute commands in containers
- `port-forward` - Forward local ports to pods
- `proxy` - Run proxy to Kubernetes API server
- `top` - Display resource usage
- `explain` - Show resource documentation
- `api-resources` - List available API resources
- `api-versions` - List available API versions
- `config` - Manage kubeconfig files
- `version` - Show client/server versions
- `cluster-info` - Show cluster information

### Non-Modifying Subcommands

When using commands with mixed subcommands, these specific subcommands are safe:

| Command | Safe Subcommands | Example |
|---------|-----------------|---------|
| `auth` | `can-i`, `whoami` | `kubectl safe auth can-i create pods` |
| `certificate` | `list` | `kubectl safe certificate list` |
| `rollout` | `status`, `history` | `kubectl safe rollout status deployment/myapp` |
| `set` | Any subcommand other than the modifying ones | Safe unless using `env`, `image`, `resources`, `selector`, `subject`, or `serviceaccount` |

## Safety Features

When executing a modifying command, kubectl-safe will:

1. **Validate Required Flags**: Ensure both `--context` and `--namespace` are provided
2. **Validate Context Existence**: Verify the specified context exists in your kubeconfig
3. **Show Target Information**: Display the target cluster context and namespace
4. **Request Confirmation**: Ask for explicit confirmation before proceeding
5. **Execute Safely**: Only proceed if the user confirms with "yes" or "y"

Example safety prompt:

```
⚠️  MODIFYING COMMAND DETECTED ⚠️

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

# Check version
make version

# Build release packages
./auto-build.sh

# Install locally for development  
make dev-install
```


### Versioning and Releases

See [VERSIONING.md](VERSIONING.md) for details on the automated versioning and release process.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

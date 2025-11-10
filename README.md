<p align="center">
  <img alt="GoReleaser Logo" src="https://avatars2.githubusercontent.com/u/24697112?v=3&s=200" height="200" />
  <h3 align="center">GoReleaser</h3>
  <p align="center">Release engineering, simplified.</p>
  <p align="center">
    <img alt="Go" src="./www/docs/static/go-light.svg#gh-light-mode-only" height="30" width="30" />
    <img alt="Go" src="./www/docs/static/go-dark.svg#gh-dark-mode-only" height="30" width="30" />
    <img alt="Rust" src="./www/docs/static/rust-light.svg#gh-light-mode-only" height="30" width="30" />
    <img alt="Rust" src="./www/docs/static/rust-dark.svg#gh-dark-mode-only" height="30" width="30" />
    <img alt="Zig" src="./www/docs/static/zig-light.svg#gh-light-mode-only" height="30" width="30" />
    <img alt="Zig" src="./www/docs/static/zig-dark.svg#gh-dark-mode-only" height="30" width="30" />
    <img alt="Bun" src="./www/docs/static/bun-light.svg#gh-light-mode-only" height="30" width="30" />
    <img alt="Bun" src="./www/docs/static/bun-dark.svg#gh-dark-mode-only" height="30" width="30" />
    <img alt="Deno" src="./www/docs/static/deno-light.svg#gh-light-mode-only" height="30" width="30" />
    <img alt="Deno" src="./www/docs/static/deno-dark.svg#gh-dark-mode-only" height="30" width="30" />
    <img alt="Python" src="./www/docs/static/python-light.svg#gh-light-mode-only" height="30" width="30" />
    <img alt="Python" src="./www/docs/static/python-dark.svg#gh-dark-mode-only" height="30" width="30" />
  </p>
</p>

A Krew plugin that provides an interactive safety net for modifying kubectl commands.

We handle the complexities of releasing so you can focus in building what really
matters: **your software**.

kubectl-safe acts as a simple, interactive wrapper around modifying kubectl commands to prevent common, high-impact mistakes. It's designed to be a final checkpoint before you make a change you might regret.

---

- **Enforces Best Practices**: Requires the mandatory use of `--context` and `--namespace` flags, forcing you to be explicit about your target
- **Context Validation**: Validates that the specified context exists in your kubeconfig to prevent typos and targeting non-existent clusters
- **Interactive Confirmation**: Shows detailed information about the target cluster and namespace before executing modifying commands
- **Transparent Pass-through**: Safe commands like `get`, `describe`, `logs` etc. are passed through without any checks
- **Comprehensive Coverage**: Protects against 13 always-modifying commands, 4 commands with specific modifying subcommands, and 2 commands that are modifying only with certain flags (see [Commands Covered](#commands-covered) for complete details)

- [On your machine](https://goreleaser.com/install/);
- [On CI/CD systems](https://goreleaser.com/ci/).

## Documentation

Install using krew with the local plugin manifest:

```bash
kubectl krew install --manifest=https://raw.githubusercontent.com/bjrooney/kubectl-safe/main/safe.yaml
```

Or if you have cloned the repository:

```bash
kubectl krew install --manifest=safe.yaml
```

## Community

You have questions, need support and or just want to talk about GoReleaser?

Here are ways to get in touch with the GoReleaser community:

[![Join Discord](https://img.shields.io/badge/Join_our_Discord_server-5865F2?style=for-the-badge&logo=discord&logoColor=white)](https://discord.gg/RGEBtg8vQ6)
[![Follow Twitter](https://img.shields.io/badge/follow_on_twitter-1DA1F2?style=for-the-badge&logo=twitter&logoColor=white)](https://twitter.com/goreleaser)
[![GitHub Discussions](https://img.shields.io/badge/GITHUB_DISCUSSION-181717?style=for-the-badge&logo=github&logoColor=white)](https://github.com/goreleaser/goreleaser/discussions)

You can find the links above and all others [here](https://goreleaser.com/links/).

Replace your modifying kubectl commands with `kubectl safe`:

This project adheres to the Contributor Covenant [code of conduct](https://github.com/goreleaser/.github/blob/main/CODE_OF_CONDUCT.md).
By participating, you are expected to uphold this code.
We appreciate your contribution.
Please refer to our [contributing guidelines](CONTRIBUTING.md) for further information.

## Badges

[![Release](https://img.shields.io/github/release/goreleaser/goreleaser.svg?style=for-the-badge)](https://github.com/goreleaser/goreleaser/releases/latest)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge)](/LICENSE.md)
[![Build status](https://img.shields.io/github/actions/workflow/status/goreleaser/goreleaser/build.yml?style=for-the-badge&branch=main)](https://github.com/goreleaser/goreleaser/actions?workflow=build)
[![Codecov branch](https://img.shields.io/codecov/c/github/goreleaser/goreleaser/main.svg?style=for-the-badge)](https://codecov.io/gh/goreleaser/goreleaser)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/goreleaser&style=for-the-badge)](https://artifacthub.io/packages/search?repo=goreleaser)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge)](http://godoc.org/github.com/goreleaser/goreleaser)
[![Powered By: GoReleaser](https://img.shields.io/badge/powered%20by-goreleaser-green.svg?style=for-the-badge)](https://github.com/goreleaser)
[![Backers on Open Collective](https://opencollective.com/goreleaser/backers/badge.svg?style=for-the-badge)](https://opencollective.com/goreleaser/backers/)
[![Sponsors on Open Collective](https://opencollective.com/goreleaser/sponsors/badge.svg?style=for-the-badge)](https://opencollective.com/goreleaser/sponsors/)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg?style=for-the-badge)](https://conventionalcommits.org)
[![CII Best Practices](https://img.shields.io/cii/summary/5420?label=openssf%20best%20practices&style=for-the-badge)](https://bestpractices.coreinfrastructure.org/projects/5420)
[![GoReportCard](https://goreportcard.com/badge/github.com/goreleaser/goreleaser?style=for-the-badge)](https://goreportcard.com/report/github.com/goreleaser/goreleaser)

## GitHub Sponsors

# This will fail - missing required flags
kubectl safe delete pod mypod
# Error: modifying command requires explicit --context and --namespace flag(s)

## OpenCollective

### Sponsors

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

Love our work and community? [Become a backer](https://opencollective.com/goreleaser).

When executing a modifying command, kubectl-safe will:

### Contributors

This project exists thanks to all the people who contribute.
[Contribution guide](CONTRIBUTING.md).

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

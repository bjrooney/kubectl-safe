# kubectl-safe Architecture

## Current Structure (v1.0.3+)

```
kubectl-safe/
├── cmd/kubectl-safe/main.go          # Entry point - calls safe.Execute()
├── pkg/safe/
│   ├── safe.go                       # Main implementation with Solarized colors
│   └── safe_test.go                  # Tests
├── main.go.legacy                    # Old implementation (not used)
└── Makefile                          # Builds ./cmd/kubectl-safe
```

## Active Components

### `cmd/kubectl-safe/main.go`
- **Purpose**: Simple entry point
- **Function**: Calls `safe.Execute()` and handles top-level errors
- **Status**: ✅ Active (built by Makefile)

### `pkg/safe/safe.go`
- **Purpose**: Main kubectl-safe implementation
- **Features**:
  - ✅ Solarized color scheme
  - ✅ Production context detection (`prod` in context name)
  - ✅ Interactive confirmations
  - ✅ Context/namespace validation with colored lists
  - ✅ Safety checks for dangerous commands
- **Status**: ✅ Active

### `pkg/safe/safe_test.go`
- **Purpose**: Comprehensive test suite
- **Coverage**: Validation, flag parsing, context detection
- **Status**: ✅ Active and passing

## Legacy Components

### `main.go.legacy`
- **Purpose**: Original implementation (renamed from `main.go`)
- **Status**: ❌ Inactive (not built)
- **Note**: Keep for reference, but not used in builds

## Build Process

```bash
make build          # Builds ./cmd/kubectl-safe -> pkg/safe
make dev-install    # Installs to ~/.local/bin
make test           # Runs pkg/safe tests
```

## Color Scheme

The current implementation uses a beautiful Solarized color palette:
- 🔴 **Red**: Critical errors and warnings
- 🟠 **Orange**: Error details and warnings  
- 🟡 **Yellow**: Highlights and user guidance
- 🟢 **Green**: Success messages and safe operations
- 🔵 **Cyan**: Commands and important info
- 🔵 **Blue**: General information labels
- 🟣 **Violet**: Interactive prompts
- 🟣 **Magenta**: Special emphasis

## Production Context Detection

Contexts containing "prod" (case-insensitive) trigger enhanced security:
- 🚨 Special production warning
- 🔐 Must type exact context name (not just "yes")
- ❌ Immediate abort on mismatch
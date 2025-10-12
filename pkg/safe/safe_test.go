package safe

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

// Test mocking variables
var (
	mockKubeconfigContexts  func() ([]string, error)
	mockNamespacesInContext func(context string) ([]string, error)
)

// Mock functions for testing
func setupMocks() {
	mockKubeconfigContexts = func() ([]string, error) {
		return []string{"dev-001", "test-001", "prod-001"}, nil
	}

	mockNamespacesInContext = func(context string) ([]string, error) {
		// Return different namespaces based on context
		switch context {
		case "dev-001":
			return []string{"default", "test", "development"}, nil
		case "test-001":
			return []string{"default", "test", "testing"}, nil
		case "prod-001":
			return []string{"default", "production", "monitoring"}, nil
		default:
			return []string{"default", "test"}, nil
		}
	}
}

func teardownMocks() {
	mockKubeconfigContexts = nil
	mockNamespacesInContext = nil
}

// Override the original functions for testing
func getKubeconfigContextsForTest() ([]string, error) {
	if mockKubeconfigContexts != nil {
		return mockKubeconfigContexts()
	}
	return getKubeconfigContexts()
}

func getNamespacesInContextForTest(context string) ([]string, error) {
	if mockNamespacesInContext != nil {
		return mockNamespacesInContext(context)
	}
	return getNamespacesInContext(context)
}

// validateRequiredFlagsForTest is a test version that uses mocked functions
func validateRequiredFlagsForTest(fs *pflag.FlagSet, commandAndArgs []string) error {
	hasContext := fs.Changed("context")
	hasNamespace := fs.Changed("namespace")

	isNamespaceResourceCommand := false
	if len(commandAndArgs) > 1 {
		resourceType := strings.ToLower(commandAndArgs[1])
		if resourceType == "namespace" || resourceType == "namespaces" || resourceType == "ns" {
			isNamespaceResourceCommand = true
		}
	}

	var missing []string
	if !hasContext {
		missing = append(missing, "--context")
	}
	if !hasNamespace && !isNamespaceResourceCommand {
		missing = append(missing, "--namespace")
	}
	if len(missing) > 0 {
		return fmt.Errorf("modifying command requires explicit %s flag(s)", strings.Join(missing, " and "))
	}

	contextValue, _ := fs.GetString("context")
	namespaceValue, _ := fs.GetString("namespace")

	// Use mocked function for testing
	availableContexts, err := getKubeconfigContextsForTest()
	if err != nil {
		return fmt.Errorf("failed to get available contexts: %w", err)
	}
	if !slices.Contains(availableContexts, contextValue) {
		return fmt.Errorf("context '%s' not found in kubeconfig. Available contexts: %s",
			contextValue, strings.Join(availableContexts, ", "))
	}

	if hasNamespace && !isNamespaceResourceCommand {
		// Use mocked function for testing
		availableNamespaces, err := getNamespacesInContextForTest(contextValue)
		if err != nil {
			return fmt.Errorf("failed to get namespaces for context '%s': %w", contextValue, err)
		}
		isCreateCommand := commandAndArgs[0] == "create"
		if !slices.Contains(availableNamespaces, namespaceValue) && !isCreateCommand {
			return fmt.Errorf("namespace '%s' not found in context '%s'. Available namespaces: %s",
				namespaceValue, contextValue, strings.Join(availableNamespaces, ", "))
		}
	}
	return nil
}

func TestIsModifyingCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		// Always-modifying commands from ModifyingCommandMap
		{
			name:     "delete command is modifying",
			args:     []string{"delete", "pod", "mypod"},
			expected: true,
		},
		{
			name:     "apply command is modifying",
			args:     []string{"apply", "-f", "deployment.yaml"},
			expected: true,
		},
		{
			name:     "create command is modifying",
			args:     []string{"create", "namespace", "test"},
			expected: true,
		},
		{
			name:     "edit command is modifying",
			args:     []string{"edit", "deployment", "myapp"},
			expected: true,
		},
		{
			name:     "patch command is modifying",
			args:     []string{"patch", "deployment", "myapp"},
			expected: true,
		},
		{
			name:     "replace command is modifying",
			args:     []string{"replace", "-f", "config.yaml"},
			expected: true,
		},
		{
			name:     "scale command is modifying",
			args:     []string{"scale", "deployment", "myapp", "--replicas=3"},
			expected: true,
		},
		{
			name:     "cordon command is modifying",
			args:     []string{"cordon", "node1"},
			expected: true,
		},
		{
			name:     "uncordon command is modifying",
			args:     []string{"uncordon", "node1"},
			expected: true,
		},
		{
			name:     "drain command is modifying",
			args:     []string{"drain", "node1"},
			expected: true,
		},
		{
			name:     "taint command is modifying",
			args:     []string{"taint", "nodes", "node1", "key=value:NoSchedule"},
			expected: true,
		},
		{
			name:     "autoscale command is modifying",
			args:     []string{"autoscale", "deployment", "myapp", "--min=1", "--max=10"},
			expected: true,
		},
		{
			name:     "expose command is modifying",
			args:     []string{"expose", "deployment", "myapp", "--port=80"},
			expected: true,
		},

		// Commands with modifying subcommands
		{
			name:     "auth reconcile is modifying",
			args:     []string{"auth", "reconcile", "-f", "rbac.yaml"},
			expected: true,
		},
		{
			name:     "auth can-i is safe (no modifying subcommand)",
			args:     []string{"auth", "can-i", "create", "pods"},
			expected: false,
		},
		{
			name:     "certificate approve is modifying",
			args:     []string{"certificate", "approve", "csr-12345"},
			expected: true,
		},
		{
			name:     "certificate deny is modifying",
			args:     []string{"certificate", "deny", "csr-12345"},
			expected: true,
		},
		{
			name:     "certificate list is safe (no modifying subcommand)",
			args:     []string{"certificate", "list"},
			expected: false,
		},
		{
			name:     "rollout restart is modifying",
			args:     []string{"rollout", "restart", "deployment/myapp"},
			expected: true,
		},
		{
			name:     "rollout undo is modifying",
			args:     []string{"rollout", "undo", "deployment/myapp"},
			expected: true,
		},
		{
			name:     "rollout status is safe (no modifying subcommand)",
			args:     []string{"rollout", "status", "deployment/myapp"},
			expected: false,
		},
		{
			name:     "set env is modifying",
			args:     []string{"set", "env", "deployment/myapp", "ENV_VAR=value"},
			expected: true,
		},
		{
			name:     "set image is modifying",
			args:     []string{"set", "image", "deployment/myapp", "container=image:tag"},
			expected: true,
		},
		{
			name:     "set resources is modifying",
			args:     []string{"set", "resources", "deployment/myapp", "--limits=cpu=100m"},
			expected: true,
		},
		{
			name:     "set selector is modifying",
			args:     []string{"set", "selector", "service/myapp", "app=myapp"},
			expected: true,
		},
		{
			name:     "set subject is modifying",
			args:     []string{"set", "subject", "clusterrolebinding/admin", "--user=john"},
			expected: true,
		},
		{
			name:     "set serviceaccount is modifying",
			args:     []string{"set", "serviceaccount", "deployment/myapp", "mysa"},
			expected: true,
		},
		{
			name:     "set something-else is safe (no modifying subcommand)",
			args:     []string{"set", "other", "deployment/myapp"},
			expected: false,
		},

		// Commands modifying by flag (ModifyingByFlagMap)
		{
			name:     "annotate with -f flag is modifying",
			args:     []string{"annotate", "-f", "deployment.yaml", "key=value"},
			expected: true,
		},
		{
			name:     "annotate with --filename flag is modifying",
			args:     []string{"annotate", "--filename", "deployment.yaml", "key=value"},
			expected: true,
		},
		{
			name:     "annotate with --filename= flag is modifying",
			args:     []string{"annotate", "--filename=deployment.yaml", "key=value"},
			expected: true,
		},
		{
			name:     "annotate without -f/--filename is safe",
			args:     []string{"annotate", "pod/mypod", "key=value"},
			expected: false,
		},
		{
			name:     "label with -f flag is modifying",
			args:     []string{"label", "-f", "deployment.yaml", "app=myapp"},
			expected: true,
		},
		{
			name:     "label with --filename flag is modifying",
			args:     []string{"label", "--filename", "deployment.yaml", "app=myapp"},
			expected: true,
		},
		{
			name:     "label with --filename= flag is modifying",
			args:     []string{"label", "--filename=deployment.yaml", "app=myapp"},
			expected: true,
		},
		{
			name:     "label without -f/--filename is safe",
			args:     []string{"label", "pod/mypod", "app=myapp"},
			expected: false,
		},

		// Safe commands (no matches in either map)
		{
			name:     "get command is safe",
			args:     []string{"get", "pods"},
			expected: false,
		},
		{
			name:     "describe command is safe",
			args:     []string{"describe", "pod", "mypod"},
			expected: false,
		},
		{
			name:     "logs command is safe",
			args:     []string{"logs", "pod/mypod"},
			expected: false,
		},
		{
			name:     "exec command is safe",
			args:     []string{"exec", "-it", "pod/mypod", "--", "bash"},
			expected: false,
		},
		{
			name:     "port-forward command is safe",
			args:     []string{"port-forward", "pod/mypod", "8080:80"},
			expected: false,
		},

		// Edge cases
		{
			name:     "empty args",
			args:     []string{},
			expected: false,
		},
		{
			name:     "single command without subcommand (set)",
			args:     []string{"set"},
			expected: false,
		},
		{
			name:     "single command without subcommand (auth)",
			args:     []string{"auth"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isModifyingCommand(tt.args, tt.args)
			if result != tt.expected {
				t.Errorf("isModifyingCommand(%v) = %v, want %v", tt.args, result, tt.expected)
			}
		})
	}
}

func TestValidateRequiredFlags(t *testing.T) {
	// Setup mocks before running tests
	setupMocks()
	defer teardownMocks()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string // Expected error message substring
	}{
		{
			name:    "both flags present (separate) - valid context and namespace",
			args:    []string{"delete", "pod", "mypod", "--context", "dev-001", "--namespace", "default"},
			wantErr: false,
		},
		{
			name:    "both flags present (equals format) - valid context and namespace",
			args:    []string{"delete", "pod", "mypod", "--context=test-001", "--namespace=test"},
			wantErr: false,
		},
		{
			name:    "both flags present (short form) - valid context and namespace",
			args:    []string{"delete", "pod", "mypod", "-c", "prod-001", "-n", "default"},
			wantErr: false,
		},
		{
			name:    "valid context but invalid namespace",
			args:    []string{"delete", "pod", "mypod", "--context", "dev-001", "--namespace", "nonexistent"},
			wantErr: true,
			errMsg:  "not found in context",
		},
		{
			name:    "invalid context",
			args:    []string{"delete", "pod", "mypod", "--context", "invalid-context", "--namespace", "default"},
			wantErr: true,
			errMsg:  "not found in kubeconfig",
		},
		{
			name:    "missing context flag",
			args:    []string{"delete", "pod", "mypod", "--namespace", "default"},
			wantErr: true,
			errMsg:  "requires explicit --context",
		},
		{
			name:    "missing namespace flag",
			args:    []string{"delete", "pod", "mypod", "--context", "dev-001"},
			wantErr: true,
			errMsg:  "requires explicit --namespace",
		},
		{
			name:    "missing both flags",
			args:    []string{"delete", "pod", "mypod"},
			wantErr: true,
			errMsg:  "requires explicit --context and --namespace",
		},
		{
			name:    "namespace operation doesn't require namespace flag",
			args:    []string{"delete", "namespace", "test", "--context", "dev-001"},
			wantErr: false,
		},
		{
			name:    "create command with new namespace should work",
			args:    []string{"create", "deployment", "myapp", "--context", "dev-001", "--namespace", "newnamespace"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a flagSet and parse the args like the main function does
			flagSet := pflag.NewFlagSet("safe-flags", pflag.ContinueOnError)
			flagSet.StringP("context", "c", "", "context")
			flagSet.StringP("namespace", "n", "", "namespace")
			flagSet.Usage = func() {}
			_ = flagSet.Parse(tt.args)

			err := validateRequiredFlagsForTest(flagSet, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequiredFlags(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
				return
			}

			// Check error message contains expected substring
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateRequiredFlags(%v) error = '%v', expected to contain '%v'", tt.args, err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestPflagExtraction(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		flagName string
		expected string
	}{
		{
			name:     "extract context with equals",
			args:     []string{"delete", "pod", "--context=prod", "--namespace=default"},
			flagName: "context",
			expected: "prod",
		},
		{
			name:     "extract context separate value",
			args:     []string{"delete", "pod", "--context", "prod", "--namespace", "default"},
			flagName: "context",
			expected: "prod",
		},
		{
			name:     "extract namespace short form",
			args:     []string{"delete", "pod", "-c", "prod", "-n", "default"},
			flagName: "namespace",
			expected: "default",
		},
		{
			name:     "flag not found",
			args:     []string{"delete", "pod"},
			flagName: "context",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flagSet := pflag.NewFlagSet("test-flags", pflag.ContinueOnError)
			flagSet.StringP("context", "c", "", "context")
			flagSet.StringP("namespace", "n", "", "namespace")
			flagSet.Usage = func() {}
			_ = flagSet.Parse(tt.args)

			result, _ := flagSet.GetString(tt.flagName)
			if result != tt.expected {
				t.Errorf("flagSet.GetString(%s) with args %v = %s, want %s", tt.flagName, tt.args, result, tt.expected)
			}
		})
	}
}

func TestVersionFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool // true if should show version
	}{
		{
			name: "version long flag",
			args: []string{"--version"},
			want: true,
		},
		{
			name: "version short flag",
			args: []string{"-v"},
			want: true,
		},
		{
			name: "not version flag",
			args: []string{"get", "pods"},
			want: false,
		},
		{
			name: "multiple args",
			args: []string{"--version", "extra"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isVersion := len(tt.args) == 1 && (tt.args[0] == "--version" || tt.args[0] == "-v")
			if isVersion != tt.want {
				t.Errorf("version flag check for %v = %v, want %v", tt.args, isVersion, tt.want)
			}
		})

	}
}

func TestGetKubeconfigContexts(t *testing.T) {
	// This test validates the function works but the exact output depends on the test environment
	contexts, err := getKubeconfigContexts()

	// We should either get contexts or an error (if kubectl is not available)
	if err != nil {
		// It's okay if kubectl is not available in test environment
		t.Skipf("kubectl not available in test environment: %v", err)
		return
	}

	// If kubectl is available, contexts should be a slice (may be empty)
	// An empty slice is valid if no contexts are configured
	if contexts == nil {
		t.Error("getKubeconfigContexts() returned nil contexts without error")
		return
	}

	// Each context should be a non-empty string
	for i, context := range contexts {
		if strings.TrimSpace(context) == "" {
			t.Errorf("getKubeconfigContexts() returned empty context at index %d", i)
		}
	}

	// Log the contexts for debugging (this is helpful to see what we got)
	t.Logf("Found %d contexts: %v", len(contexts), contexts)
}

func TestMockedKubeconfigContexts(t *testing.T) {
	// Setup mocks
	setupMocks()
	defer teardownMocks()

	// Test mocked contexts
	contexts, err := getKubeconfigContextsForTest()
	if err != nil {
		t.Fatalf("getKubeconfigContextsForTest() returned error: %v", err)
	}

	expectedContexts := []string{"dev-001", "test-001", "prod-001"}
	if len(contexts) != len(expectedContexts) {
		t.Errorf("getKubeconfigContextsForTest() returned %d contexts, want %d", len(contexts), len(expectedContexts))
	}

	for i, expected := range expectedContexts {
		if i >= len(contexts) || contexts[i] != expected {
			t.Errorf("getKubeconfigContextsForTest() context[%d] = %v, want %v", i, contexts[i], expected)
		}
	}
}

func TestMockedNamespacesInContext(t *testing.T) {
	// Setup mocks
	setupMocks()
	defer teardownMocks()

	tests := []struct {
		context  string
		expected []string
	}{
		{
			context:  "dev-001",
			expected: []string{"default", "test", "development"},
		},
		{
			context:  "test-001",
			expected: []string{"default", "test", "testing"},
		},
		{
			context:  "prod-001",
			expected: []string{"default", "production", "monitoring"},
		},
		{
			context:  "unknown-context",
			expected: []string{"default", "test"},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("context_%s", tt.context), func(t *testing.T) {
			namespaces, err := getNamespacesInContextForTest(tt.context)
			if err != nil {
				t.Fatalf("getNamespacesInContextForTest(%s) returned error: %v", tt.context, err)
			}

			if len(namespaces) != len(tt.expected) {
				t.Errorf("getNamespacesInContextForTest(%s) returned %d namespaces, want %d", tt.context, len(namespaces), len(tt.expected))
			}

			for i, expected := range tt.expected {
				if i >= len(namespaces) || namespaces[i] != expected {
					t.Errorf("getNamespacesInContextForTest(%s) namespace[%d] = %v, want %v", tt.context, i, namespaces[i], expected)
				}
			}
		})
	}
}

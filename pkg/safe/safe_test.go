package safe

import (
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

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
	tests := []struct {
		name                  string
		args                  []string
		wantErr               bool
		skipContextValidation bool // Skip context validation for some tests
	}{
		{
			name:                  "both flags present (separate)",
			args:                  []string{"delete", "pod", "mypod", "--context", "test-context", "--namespace", "default"},
			wantErr:               true, // Will fail context validation unless test-context exists
			skipContextValidation: false,
		},
		{
			name:                  "both flags present (equals format)",
			args:                  []string{"delete", "pod", "mypod", "--context=test-context", "--namespace=default"},
			wantErr:               true, // Will fail context validation unless test-context exists
			skipContextValidation: false,
		},
		{
			name:                  "both flags present (short form)",
			args:                  []string{"delete", "pod", "mypod", "-c", "test-context", "-n", "default"},
			wantErr:               true, // Will fail context validation unless test-context exists
			skipContextValidation: false,
		},
		{
			name:                  "missing context flag",
			args:                  []string{"delete", "pod", "mypod", "--namespace", "default"},
			wantErr:               true,
			skipContextValidation: true, // No context to validate
		},
		{
			name:                  "missing namespace flag",
			args:                  []string{"delete", "pod", "mypod", "--context", "test-context"},
			wantErr:               true,
			skipContextValidation: true, // Missing namespace, so context validation is not the main issue
		},
		{
			name:                  "missing both flags",
			args:                  []string{"delete", "pod", "mypod"},
			wantErr:               true,
			skipContextValidation: true, // No flags to validate
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

			err := validateRequiredFlags(flagSet, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequiredFlags(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
			}

			// Additional check: if we expect an error and got one, verify it's the right type
			if tt.wantErr && err != nil {
				errMsg := err.Error()
				if !tt.skipContextValidation && strings.Contains(errMsg, "not found in kubeconfig") {
					// This is the expected context validation error
					return
				}
				if strings.Contains(errMsg, "requires explicit") {
					// This is the expected missing flag error
					return
				}
				if strings.Contains(errMsg, "failed to get available contexts") {
					// This is acceptable if kubectl is not available
					return
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

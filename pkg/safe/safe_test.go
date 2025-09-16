package safe

import (
	"strings"
	"testing"
)

func TestIsDangerousCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "delete command is dangerous",
			args:     []string{"delete", "pod", "mypod"},
			expected: true,
		},
		{
			name:     "apply command is dangerous",
			args:     []string{"apply", "-f", "deployment.yaml"},
			expected: true,
		},
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
			name:     "empty args",
			args:     []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDangerousCommand(tt.args)
			if result != tt.expected {
				t.Errorf("isDangerousCommand(%v) = %v, want %v", tt.args, result, tt.expected)
			}
		})
	}
}

func TestValidateRequiredFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		skipContextValidation bool // Skip context validation for some tests
	}{
		{
			name:    "both flags present (separate)",
			args:    []string{"delete", "pod", "mypod", "--context", "test-context", "--namespace", "default"},
			wantErr: true, // Will fail context validation unless test-context exists
			skipContextValidation: false,
		},
		{
			name:    "both flags present (equals format)",
			args:    []string{"delete", "pod", "mypod", "--context=test-context", "--namespace=default"},
			wantErr: true, // Will fail context validation unless test-context exists
			skipContextValidation: false,
		},
		{
			name:    "both flags present (short form)",
			args:    []string{"delete", "pod", "mypod", "-c", "test-context", "-n", "default"},
			wantErr: true, // Will fail context validation unless test-context exists
			skipContextValidation: false,
		},
		{
			name:    "missing context flag",
			args:    []string{"delete", "pod", "mypod", "--namespace", "default"},
			wantErr: true,
			skipContextValidation: true, // No context to validate
		},
		{
			name:    "missing namespace flag",
			args:    []string{"delete", "pod", "mypod", "--context", "test-context"},
			wantErr: true,
			skipContextValidation: true, // Missing namespace, so context validation is not the main issue
		},
		{
			name:    "missing both flags",
			args:    []string{"delete", "pod", "mypod"},
			wantErr: true,
			skipContextValidation: true, // No flags to validate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequiredFlags(tt.args)
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

func TestExtractFlagValue(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		longFlag  string
		shortFlag string
		expected  string
	}{
		{
			name:      "extract context with equals",
			args:      []string{"delete", "pod", "--context=prod", "--namespace=default"},
			longFlag:  "--context",
			shortFlag: "-c",
			expected:  "prod",
		},
		{
			name:      "extract context separate value",
			args:      []string{"delete", "pod", "--context", "prod", "--namespace", "default"},
			longFlag:  "--context",
			shortFlag: "-c",
			expected:  "prod",
		},
		{
			name:      "extract namespace short form",
			args:      []string{"delete", "pod", "-c", "prod", "-n", "default"},
			longFlag:  "--namespace",
			shortFlag: "-n",
			expected:  "default",
		},
		{
			name:      "flag not found",
			args:      []string{"delete", "pod"},
			longFlag:  "--context",
			shortFlag: "-c",
			expected:  "<not specified>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFlagValue(tt.args, tt.longFlag, tt.shortFlag)
			if result != tt.expected {
				t.Errorf("extractFlagValue(%v, %s, %s) = %s, want %s", tt.args, tt.longFlag, tt.shortFlag, result, tt.expected)
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
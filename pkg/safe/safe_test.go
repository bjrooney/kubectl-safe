package safe

import (
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
	}{
		{
			name:    "both flags present (separate)",
			args:    []string{"delete", "pod", "mypod", "--context", "prod", "--namespace", "default"},
			wantErr: false,
		},
		{
			name:    "both flags present (equals format)",
			args:    []string{"delete", "pod", "mypod", "--context=prod", "--namespace=default"},
			wantErr: false,
		},
		{
			name:    "both flags present (short form)",
			args:    []string{"delete", "pod", "mypod", "-c", "prod", "-n", "default"},
			wantErr: false,
		},
		{
			name:    "missing context flag",
			args:    []string{"delete", "pod", "mypod", "--namespace", "default"},
			wantErr: true,
		},
		{
			name:    "missing namespace flag",
			args:    []string{"delete", "pod", "mypod", "--context", "prod"},
			wantErr: true,
		},
		{
			name:    "missing both flags",
			args:    []string{"delete", "pod", "mypod"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequiredFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequiredFlags(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
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
// Package safe provides an interactive safety net for dangerous kubectl commands.
//
// This package implements the core functionality of kubectl-safe, a plugin that
// acts as a wrapper around destructive kubectl operations to prevent common
// mistakes by requiring explicit context and namespace flags and showing
// interactive confirmation prompts.
package safe

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

// DangerousCommands defines the list of kubectl commands that can cause data loss 
// or service disruption and therefore require safety checks.
//
// These commands will trigger the safety validation process which includes:
// - Requiring explicit --context and --namespace flags
// - Displaying an interactive confirmation prompt
// - Showing target cluster and namespace information
//
// Commands not in this list are considered safe and will pass through to kubectl
// without any safety checks.
var DangerousCommands = []string{
	"delete",    // Delete resources by file names, stdin, resources and names, or by resources and label selector
	"apply",     // Apply a configuration to a resource by file name or stdin
	"create",    // Create a resource from a file or from stdin
	"replace",   // Replace a resource by file name or stdin
	"patch",     // Update fields of a resource using strategic merge patch, a JSON merge patch, or a JSON patch
	"edit",      // Edit a resource on the server (opens an editor)
	"scale",     // Set a new size for a deployment, replica set, or replication controller
	"rollout",   // Manage the rollout of a resource (deployments, daemonsets, statefulsets)
	"drain",     // Drain node in preparation for maintenance
	"cordon",    // Mark node as unschedulable
	"uncordon",  // Mark node as schedulable
	"taint",     // Update the taints on one or more nodes
}

// Execute is the main entry point for the kubectl-safe plugin.
//
// This function processes command line arguments and determines whether to:
// 1. Show usage information (when no arguments provided)
// 2. Pass through to kubectl directly (for safe commands)
// 3. Apply safety checks and confirmation (for dangerous commands)
//
// The safety check process includes:
// - Validating that required --context and --namespace flags are present
// - Displaying an interactive confirmation prompt with target details
// - Only proceeding if the user explicitly confirms the operation
//
// Returns an error if:
// - Required flags are missing for dangerous commands
// - User cancels the operation
// - The underlying kubectl command fails
func Execute() error {
	args := os.Args[1:] // Skip the program name

	// If no command is given, show usage information
	if len(args) == 0 {
		return showUsage()
	}

	// Check if this is a dangerous command that requires safety checks
	if !isDangerousCommand(args) {
		// For safe commands, just pass through to kubectl without any checks
		return executeKubectl(args)
	}

	// For dangerous commands, enforce safety checks
	if err := validateRequiredFlags(args); err != nil {
		return err
	}

	// Show interactive confirmation with target details
	if err := showConfirmation(args); err != nil {
		return err
	}

	// All safety checks passed, execute the kubectl command
	return executeKubectl(args)
}

// isDangerousCommand checks if the command contains dangerous operations.
//
// This function examines the first argument (the kubectl command) to determine
// if it appears in the DangerousCommands list. Only exact matches are considered.
//
// Args:
//   args: slice of command line arguments, where args[0] should be the kubectl command
//
// Returns:
//   bool: true if the command is dangerous and requires safety checks, false otherwise
//
// Examples:
//   isDangerousCommand([]string{"delete", "pod", "mypod"}) -> true
//   isDangerousCommand([]string{"get", "pods"}) -> false
//   isDangerousCommand([]string{}) -> false
func isDangerousCommand(args []string) bool {
	if len(args) == 0 {
		return false
	}

	command := args[0]
	return slices.Contains(DangerousCommands, command)
}

// validateRequiredFlags ensures --context and --namespace are provided for dangerous commands.
//
// This function scans the command line arguments to verify that both the --context
// and --namespace flags are present. This is a critical safety requirement to ensure
// users are explicit about which cluster and namespace they are targeting.
//
// Supported flag formats:
//   --context=value or --context value
//   --namespace=value or --namespace value or -n=value or -n value
//
// Args:
//   args: slice of command line arguments to validate
//
// Returns:
//   error: nil if both required flags are present, otherwise an error describing
//          which flags are missing
//
// Examples:
//   validateRequiredFlags([]string{"delete", "pod", "--context=prod", "--namespace=default"}) -> nil
//   validateRequiredFlags([]string{"delete", "pod"}) -> error (both flags missing)
//   validateRequiredFlags([]string{"delete", "pod", "--context=prod"}) -> error (namespace missing)
func validateRequiredFlags(args []string) error {
	hasContext := false
	hasNamespace := false

	// Scan through all arguments looking for required flags
	for _, arg := range args {
		// Check for standalone flag names
		if arg == "--context" || arg == "-c" {
			hasContext = true
		}
		if arg == "--namespace" || arg == "-n" {
			hasNamespace = true
		}
		// Check for flag=value format
		if strings.HasPrefix(arg, "--context=") || strings.HasPrefix(arg, "-c=") {
			hasContext = true
		}
		if strings.HasPrefix(arg, "--namespace=") || strings.HasPrefix(arg, "-n=") {
			hasNamespace = true
		}
	}

	// Build list of missing required flags
	var missing []string
	if !hasContext {
		missing = append(missing, "--context")
	}
	if !hasNamespace {
		missing = append(missing, "--namespace")
	}

	// Return error if any required flags are missing
	if len(missing) > 0 {
		return fmt.Errorf("dangerous command requires explicit %s flag(s). This ensures you're targeting the correct cluster and namespace", strings.Join(missing, " and "))
	}

	return nil
}

// showConfirmation displays an interactive prompt for dangerous commands.
//
// This function presents a detailed confirmation dialog that includes:
// - A warning about the dangerous nature of the command
// - The full command that will be executed
// - Target cluster context and namespace information
// - A yes/no prompt requiring explicit user confirmation
//
// The user must respond with "yes" or "y" (case insensitive) to proceed.
// Any other response will cancel the operation.
//
// Args:
//   args: slice of command line arguments representing the kubectl command to execute
//
// Returns:
//   error: nil if user confirms, otherwise an error if user cancels or input fails
//
// Example output:
//   ⚠️  DANGEROUS COMMAND DETECTED ⚠️
//   
//   You are about to execute: kubectl delete pod mypod --context=prod --namespace=default
//   
//   Target Details:
//     Context:   prod
//     Namespace: default
//   
//   This operation may cause data loss or service disruption.
//   Are you sure you want to continue? (yes/no):
func showConfirmation(args []string) error {
	fmt.Printf("⚠️  DANGEROUS COMMAND DETECTED ⚠️\n\n")
	fmt.Printf("You are about to execute: kubectl %s\n\n", strings.Join(args, " "))
	
	// Extract context and namespace for display
	context := extractFlagValue(args, "--context", "-c")
	namespace := extractFlagValue(args, "--namespace", "-n")
	
	fmt.Printf("Target Details:\n")
	fmt.Printf("  Context:   %s\n", context)
	fmt.Printf("  Namespace: %s\n\n", namespace)
	
	fmt.Printf("This operation may cause data loss or service disruption.\n")
	fmt.Printf("Are you sure you want to continue? (yes/no): ")

	// Read user response from stdin
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read user input: %w", err)
	}

	// Check if user confirmed the operation
	response = strings.TrimSpace(strings.ToLower(response))
	if response != "yes" && response != "y" {
		fmt.Println("Operation cancelled.")
		return fmt.Errorf("operation cancelled by user")
	}

	fmt.Println("Proceeding with operation...")
	return nil
}

// extractFlagValue extracts the value for a given flag from args.
//
// This helper function searches through command line arguments to find a specific
// flag and return its value. It supports both the equals format (--flag=value)
// and the space-separated format (--flag value).
//
// Args:
//   args: slice of command line arguments to search through
//   longFlag: the long form of the flag (e.g., "--context")
//   shortFlag: the short form of the flag (e.g., "-c")
//
// Returns:
//   string: the value of the flag if found, otherwise "<not specified>"
//
// Supported formats:
//   --flag=value or -f=value (equals format)
//   --flag value or -f value (space-separated format)
//
// Examples:
//   extractFlagValue([]string{"--context=prod"}, "--context", "-c") -> "prod"
//   extractFlagValue([]string{"--context", "prod"}, "--context", "-c") -> "prod"
//   extractFlagValue([]string{"-n", "default"}, "--namespace", "-n") -> "default"
//   extractFlagValue([]string{"delete", "pod"}, "--context", "-c") -> "<not specified>"
func extractFlagValue(args []string, longFlag, shortFlag string) string {
	for i, arg := range args {
		// Check for --flag=value format
		if strings.HasPrefix(arg, longFlag+"=") {
			return strings.TrimPrefix(arg, longFlag+"=")
		}
		if strings.HasPrefix(arg, shortFlag+"=") {
			return strings.TrimPrefix(arg, shortFlag+"=")
		}
		// Check for --flag value format (flag followed by value in next argument)
		if (arg == longFlag || arg == shortFlag) && i+1 < len(args) {
			return args[i+1]
		}
	}
	return "<not specified>"
}

// executeKubectl runs the actual kubectl command with the provided arguments.
//
// This function creates a new process to execute kubectl with the given arguments,
// ensuring that stdin, stdout, and stderr are properly connected so that interactive
// kubectl commands work correctly and output is displayed to the user.
//
// Args:
//   args: slice of arguments to pass to kubectl
//
// Returns:
//   error: nil if kubectl execution succeeds, otherwise the error from kubectl
//
// The function preserves the exit code of kubectl, so if kubectl fails,
// this function will also return an error.
func executeKubectl(args []string) error {
	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	return cmd.Run()
}

// showUsage displays help information for the kubectl-safe plugin.
//
// This function prints comprehensive usage information including:
// - Basic usage syntax and examples
// - Explanation of the safety features
// - List of dangerous commands that trigger safety checks
// - Examples of safe commands that pass through
//
// The help text is designed to be informative for both new users learning
// about the plugin and experienced users who need a quick reference.
//
// Returns:
//   error: always returns nil (help display cannot fail)
func showUsage() error {
	fmt.Printf(`kubectl-safe: Interactive safety net for dangerous kubectl commands

Usage:
  kubectl safe <kubectl-command> [flags]

This plugin acts as a safety wrapper around kubectl commands. For dangerous operations,
it will:
  - Require explicit --context and --namespace flags
  - Show an interactive confirmation prompt
  - Display target cluster and namespace information

Examples:
  kubectl safe delete pod mypod --context=prod --namespace=default
  kubectl safe apply -f deployment.yaml --context=staging --namespace=myapp

Dangerous commands that trigger safety checks:
  %s

For safe commands, this plugin acts as a transparent pass-through to kubectl.

`, strings.Join(DangerousCommands, ", "))
	return nil
}
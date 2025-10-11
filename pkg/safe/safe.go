package safe

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/fatih/color"
)

// Solarized color scheme
var (
	// Solarized base colors
	solarizedRed     = color.New(color.FgHiRed).Add(color.Bold)    // #dc322f - for warnings/errors
	solarizedOrange  = color.New(color.FgHiYellow).Add(color.Bold) // #cb4b16 - for warnings
	solarizedYellow  = color.New(color.FgYellow)                   // #b58900 - for highlights
	solarizedGreen   = color.New(color.FgGreen)                    // #859900 - for success/safe
	solarizedCyan    = color.New(color.FgCyan)                     // #2aa198 - for info/commands
	solarizedBlue    = color.New(color.FgBlue)                     // #268bd2 - for info
	solarizedViolet  = color.New(color.FgMagenta)                  // #6c71c4 - for special
	solarizedMagenta = color.New(color.FgHiMagenta)                // #d33682 - for emphasis
)

// printColoredList prints a list of items with alternating Solarized colors
func printColoredList(items []string) {
	colors := []*color.Color{solarizedCyan, solarizedGreen, solarizedYellow, solarizedViolet, solarizedBlue, solarizedMagenta}

	for i, item := range items {
		colorIndex := i % len(colors)
		if i > 0 {
			fmt.Print(", ")
		}
		colors[colorIndex].Print(item)
	}
	fmt.Print("\n")
}

// Version will be set at build time
var Version = "dev"

// DangerousCommands are kubectl commands that can cause data loss or service disruption
var DangerousCommands = []string{
	"delete",
	"apply",
	"create",
	"replace",
	"patch",
	"edit",
	"scale",
	"rollout",
	"drain",
	"cordon",
	"uncordon",
	"taint",
}

// Execute is the main entry point for the kubectl-safe plugin
func Execute() error {
	args := os.Args[1:] // Skip the program name

	if len(args) == 0 {
		return showUsage()
	}

	// Handle version flag
	if len(args) == 1 && (args[0] == "--version" || args[0] == "-v") {
		solarizedBlue.Print("kubectl-safe version ")
		solarizedCyan.Printf("%s\n", Version)
		return nil
	}

	// Check if this is a dangerous command
	if !isDangerousCommand(args) {
		// For safe commands, show a message and pass through to kubectl
		solarizedGreen.Print("→ Safe command detected. ")
		solarizedCyan.Println("Passing directly to kubectl...")
		return executeKubectl(args)
	}

	// For dangerous commands, enforce safety checks
	if err := validateRequiredFlags(args); err != nil {
		return err
	}

	// Show interactive confirmation
	if err := showConfirmation(args); err != nil {
		return err
	}

	// Execute the kubectl command
	return executeKubectl(args)
}

// isDangerousCommand checks if the command contains dangerous operations
func isDangerousCommand(args []string) bool {
	if len(args) == 0 {
		return false
	}

	command := args[0]
	return slices.Contains(DangerousCommands, command)
}

// validateRequiredFlags ensures --context and --namespace are provided for dangerous commands
func validateRequiredFlags(args []string) error {
	hasContext := false
	hasNamespace := false
	contextValue := ""

	for _, arg := range args {
		if arg == "--context" || arg == "-c" {
			hasContext = true
		}
		if arg == "--namespace" || arg == "-n" {
			hasNamespace = true
		}
		// Check for flag=value format
		if strings.HasPrefix(arg, "--context=") || strings.HasPrefix(arg, "-c=") {
			hasContext = true
			contextValue = extractFlagValue(args, "--context", "-c")
		}
		if strings.HasPrefix(arg, "--namespace=") || strings.HasPrefix(arg, "-n=") {
			hasNamespace = true
		}
	}

	// If context was found but not extracted yet (separate flag format), extract it
	if hasContext && contextValue == "" {
		contextValue = extractFlagValue(args, "--context", "-c")
	}

	var missing []string
	if !hasContext {
		missing = append(missing, "--context")
	}
	if !hasNamespace {
		missing = append(missing, "--namespace")
	}

	if len(missing) > 0 {
		solarizedRed.Print("❌ ERROR: ")
		solarizedOrange.Printf("Dangerous command requires explicit %s flag(s).\n", strings.Join(missing, " and "))
		solarizedYellow.Print("This ensures you're targeting the correct cluster and namespace.\n")
		os.Exit(1)
	}

	// Validate that the provided context exists in kubeconfig
	if hasContext && contextValue != "" && contextValue != "<not specified>" {
		availableContexts, err := getKubeconfigContexts()
		if err != nil {
			solarizedRed.Print("❌ ERROR: ")
			solarizedOrange.Printf("Failed to get available contexts from kubeconfig: %v\n", err)
			os.Exit(1)
		}

		if !slices.Contains(availableContexts, contextValue) {
			solarizedYellow.Print("✋ WARNING: ")
			solarizedOrange.Printf("Context '%s' not found in kubeconfig.\n", contextValue)
			solarizedBlue.Print("Available contexts: ")
			printColoredList(availableContexts)
			os.Exit(1)
		}
	}

	// Validate that the provided namespace exists in the specified context
	if hasNamespace && hasContext && contextValue != "" && contextValue != "<not specified>" {
		namespaceValue := extractFlagValue(args, "--namespace", "-n")
		if namespaceValue != "" && namespaceValue != "<not specified>" {
			availableNamespaces, err := getNamespacesInContext(contextValue)
			if err != nil {
				solarizedRed.Print("❌ ERROR: ")
				solarizedOrange.Printf("Failed to get available namespaces for context '%s': %v\n", contextValue, err)
				os.Exit(1)
			}

			if !slices.Contains(availableNamespaces, namespaceValue) {
				solarizedYellow.Print("✋ WARNING: ")
				solarizedOrange.Printf("Namespace '%s' not found in context '%s'.\n", namespaceValue, contextValue)
				solarizedBlue.Print("Available namespaces: ")
				printColoredList(availableNamespaces)
				os.Exit(1)
			}
		}
	}

	return nil
}

// showConfirmation displays an interactive prompt for dangerous commands
func showConfirmation(args []string) error {
	solarizedRed.Print("⚠️  DANGEROUS COMMAND DETECTED ⚠️\n\n")
	solarizedOrange.Print("You are about to execute: ")
	solarizedCyan.Printf("kubectl %s\n\n", strings.Join(args, " "))

	// Extract context and namespace for display
	context := extractFlagValue(args, "--context", "-c")
	namespace := extractFlagValue(args, "--namespace", "-n")

	solarizedBlue.Print("Target Details:\n")
	solarizedBlue.Print("  Context:   ")
	solarizedCyan.Printf("%s\n", context)
	solarizedBlue.Print("  Namespace: ")
	solarizedCyan.Printf("%s\n\n", namespace)

	solarizedYellow.Print("This operation may cause data loss or service disruption.\n")
	solarizedViolet.Print("Are you sure you want to continue? (yes/no): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		solarizedRed.Print("❌ ERROR: ")
		solarizedOrange.Printf("Failed to read user input: %v\n", err)
		return fmt.Errorf("failed to read user input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "yes" && response != "y" {
		solarizedRed.Print("Operation cancelled.")
		return fmt.Errorf("operation cancelled by user")
	}

	solarizedGreen.Println("Proceeding with operation...")
	return nil
}

// extractFlagValue extracts the value for a given flag from args
func extractFlagValue(args []string, longFlag, shortFlag string) string {
	for i, arg := range args {
		// Check for --flag=value format
		if strings.HasPrefix(arg, longFlag+"=") {
			return strings.TrimPrefix(arg, longFlag+"=")
		}
		if strings.HasPrefix(arg, shortFlag+"=") {
			return strings.TrimPrefix(arg, shortFlag+"=")
		}
		// Check for --flag value format
		if (arg == longFlag || arg == shortFlag) && i+1 < len(args) {
			return args[i+1]
		}
	}
	return "<not specified>"
}

// executeKubectl runs the actual kubectl command
func executeKubectl(args []string) error {
	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// getKubeconfigContexts returns the list of available contexts from kubeconfig
func getKubeconfigContexts() ([]string, error) {
	cmd := exec.Command("kubectl", "config", "get-contexts", "--output=name")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute kubectl config get-contexts: %w", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		// No contexts available - return empty slice
		return []string{}, nil
	}

	contexts := strings.Split(outputStr, "\n")

	// Filter out empty strings
	var validContexts []string
	for _, context := range contexts {
		if strings.TrimSpace(context) != "" {
			validContexts = append(validContexts, strings.TrimSpace(context))
		}
	}

	return validContexts, nil
}

// getNamespacesInContext returns the list of available namespaces in the given context
func getNamespacesInContext(context string) ([]string, error) {
	cmd := exec.Command("kubectl", "get", "namespaces", "--output=name", "--context", context)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get namespaces for context '%s': %w", context, err)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		// No namespaces available - return empty slice
		return []string{}, nil
	}

	namespaceLines := strings.Split(outputStr, "\n")

	// Extract namespace names from the "namespace/name" format
	var namespaces []string
	for _, line := range namespaceLines {
		line = strings.TrimSpace(line)
		if line != "" && strings.HasPrefix(line, "namespace/") {
			nsName := strings.TrimPrefix(line, "namespace/")
			if nsName != "" {
				namespaces = append(namespaces, nsName)
			}
		}
	}

	return namespaces, nil
}

// showUsage displays help information
func showUsage() error {
	solarizedBlue.Print("kubectl-safe: ")
	solarizedCyan.Print("Interactive safety net for dangerous kubectl commands\n")
	solarizedYellow.Printf("Version: %s\n\n", Version)

	solarizedViolet.Print("Usage:\n")
	fmt.Print("  ")
	solarizedCyan.Print("kubectl safe ")
	fmt.Print("<kubectl-command> [flags]\n")
	fmt.Print("  ")
	solarizedCyan.Print("kubectl safe ")
	fmt.Print("--version\n\n")

	fmt.Print("This plugin acts as a safety wrapper around kubectl commands. For dangerous operations,\nit will:\n")
	solarizedGreen.Print("  - Require explicit --context and --namespace flags\n")
	solarizedGreen.Print("  - Show an interactive confirmation prompt\n")
	solarizedGreen.Print("  - Display target cluster and namespace information\n\n")

	solarizedViolet.Print("Examples:\n")
	fmt.Print("  ")
	solarizedCyan.Print("kubectl safe delete pod mypod --context=prod --namespace=default\n")
	fmt.Print("  ")
	solarizedCyan.Print("kubectl safe apply -f deployment.yaml --context=staging --namespace=myapp\n\n")

	solarizedOrange.Print("Dangerous commands that trigger safety checks:\n")
	fmt.Printf("  %s\n\n", strings.Join(DangerousCommands, ", "))

	fmt.Print("For safe commands, this plugin acts as a transparent pass-through to kubectl.\n\n")
	return nil
}

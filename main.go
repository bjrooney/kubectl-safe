package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

// Solarized color palette
var (
	// Solarized colors
	solarizedRed     = color.New(color.FgHiRed)     // #dc322f - for general errors
	solarizedOrange  = color.New(color.FgHiYellow)  // #cb4b16 - for context validation
	solarizedYellow  = color.New(color.FgYellow)    // #b58900 - for warnings
	solarizedGreen   = color.New(color.FgGreen)     // #859900 - for success
	solarizedCyan    = color.New(color.FgCyan)      // #2aa198 - for info
	solarizedBlue    = color.New(color.FgBlue)      // #268bd2 - for namespace validation
	solarizedViolet  = color.New(color.FgMagenta)   // #6c71c4 - for special cases
	solarizedMagenta = color.New(color.FgHiMagenta) // #d33682 - for highlights
)

// dangerousCommands lists kubectl commands that require extra safety checks.
var dangerousCommands = map[string]bool{
	"delete":  true,
	"apply":   true,
	"edit":    true,
	"patch":   true,
	"rollout": true,
	"scale":   true,
	"create":  true,
	"replace": true,
}

// executeKubectl runs the actual kubectl command with the given arguments.
func executeKubectl(args ...string) {
	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}

// askForConfirmation prompts the user for y/n confirmation and returns true if 'y' is entered.
func askForConfirmation() bool {
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		color.Red("ERROR: Failed to read user input: %v", err)
		return false
	}
	return strings.ToLower(strings.TrimSpace(response)) == "y"
}

// contextExists checks if the given context exists in the user's kubeconfig.
func contextExists(context string) bool {
	if context == "" {
		return false
	}
	cmd := exec.Command("kubectl", "config", "get-contexts", "-o", "name")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	contexts := strings.Split(string(output), "\n")
	for _, ctx := range contexts {
		if ctx == context {
			return true
		}
	}
	return false
}

// getAvailableContexts returns a list of available contexts from kubeconfig
func getAvailableContexts() []string {
	cmd := exec.Command("kubectl", "config", "get-contexts", "-o", "name")
	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	var contexts []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			contexts = append(contexts, line)
		}
	}
	return contexts
}

// namespaceExists checks if the given namespace exists in the specified context.
func namespaceExists(context, namespace string) bool {
	if namespace == "" || context == "" {
		return false
	}
	cmd := exec.Command("kubectl", "get", "namespaces", "-o", "name", "--context", context)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	namespaces := strings.Split(string(output), "\n")
	for _, ns := range namespaces {
		// Output format is "namespace/namespacename", so we need to extract the name
		if strings.HasPrefix(ns, "namespace/") {
			nsName := strings.TrimPrefix(ns, "namespace/")
			if nsName == namespace {
				return true
			}
		}
	}
	return false
}

// getAvailableNamespaces returns a list of available namespaces in the specified context
func getAvailableNamespaces(context string) []string {
	if context == "" {
		return []string{}
	}
	cmd := exec.Command("kubectl", "get", "namespaces", "-o", "name", "--context", context)
	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	var namespaces []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// Output format is "namespace/namespacename", so we need to extract the name
		if strings.HasPrefix(line, "namespace/") {
			nsName := strings.TrimPrefix(line, "namespace/")
			if nsName != "" {
				namespaces = append(namespaces, nsName)
			}
		}
	}
	return namespaces
}

// colorizeItems returns a string with each item colored using different Solarized colors
func colorizeItems(items []string) string {
	if len(items) == 0 {
		return ""
	}

	// Solarized colors for cycling through items
	colors := []*color.Color{
		solarizedGreen,   // #859900
		solarizedCyan,    // #2aa198
		solarizedBlue,    // #268bd2
		solarizedViolet,  // #6c71c4
		solarizedMagenta, // #d33682
		solarizedRed,     // #dc322f
		solarizedOrange,  // #cb4b16
		solarizedYellow,  // #b58900
	}

	var colorizedItems []string
	for i, item := range items {
		colorIndex := i % len(colors)
		colorizedItem := colors[colorIndex].Sprint(item)
		colorizedItems = append(colorizedItems, colorizedItem)
	}

	return strings.Join(colorizedItems, ", ")
}

// parseContextAndNamespace extracts context and namespace from kubectl arguments.
func parseContextAndNamespace(args []string) (context, namespace string, contextIsSet, namespaceIsSet bool) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--context=") {
			contextIsSet = true
			context = strings.SplitN(arg, "=", 2)[1]
		} else if strings.HasPrefix(arg, "--namespace=") {
			namespaceIsSet = true
			namespace = strings.SplitN(arg, "=", 2)[1]
		} else if arg == "--context" || arg == "-n" || arg == "--namespace" {
			if (i + 1) < len(args) {
				if arg == "--context" {
					contextIsSet = true
					context = args[i+1]
				} else {
					namespaceIsSet = true
					namespace = args[i+1]
				}
			}
		}
	}
	return
}

// printCommandSummary displays the command and context/namespace info to the user.
func printCommandSummary(allArgs []string, context, namespace string) {
	color.Red("⚠️  DANGEROUS COMMAND DETECTED ⚠️")
	color.Red("You are about to run the following command:")
	fullCommandStr := fmt.Sprintf("kubectl %s", strings.Join(allArgs, " "))
	cyan := color.New(color.FgCyan)
	cyan.Printf("  %s\n", fullCommandStr)
	color.Red("on context ")
	cyan.Printf("%s", context)
	color.Red(" in namespace ")
	cyan.Printf("%s\n", namespace)
}

// confirmProductionContext prompts for context name if prod, else y/n confirmation.
func confirmProductionContext(context string) bool {
	if strings.Contains(strings.ToLower(context), "prod") {
		color.Red("WARNING: You are about to run a command on a PRODUCTION context!")
		fmt.Printf("To proceed, please type the context name ('%s') and press Enter: ", context)
		reader := bufio.NewReader(os.Stdin)
		confirmation, err := reader.ReadString('\n')
		if err != nil {
			color.Red("ERROR: Failed to read user input: %v", err)
			return false
		}
		if strings.TrimSpace(confirmation) != context {
			color.Red("Aborted: Context name did not match. Command will not be executed.")
			return false
		}
		return true
	} else {
		yellow := color.New(color.FgYellow)
		yellow.Print("Do you want to continue? (y/n): ")
		return askForConfirmation()
	}
}

func main() {
	if len(os.Args) < 2 {
		// If no command is given, just run kubectl help.
		executeKubectl()
		return
	}

	command := os.Args[1]
	kubectlArgs := os.Args[2:]
	allArgs := os.Args[1:]

	// Check if the command is NOT in our dangerous list.
	if !dangerousCommands[command] {
		solarizedGreen.Printf("--> Safe command detected. Passing directly to kubectl...\n")
		executeKubectl(allArgs...)
		return
	}

	// Parse context and namespace from args.
	foundContext, foundNamespace, contextIsSet, namespaceIsSet := parseContextAndNamespace(kubectlArgs)

	// Enforce that flags are set for dangerous commands.
	missingArgs := false
	if !contextIsSet {
		solarizedRed.Printf("ERROR: The --context flag is mandatory for the dangerous command '%s'.\n", command)
		missingArgs = true
	}
	if !namespaceIsSet {
		solarizedRed.Printf("ERROR: The --namespace (-n) flag is mandatory for the dangerous command '%s'.\n", command)
		missingArgs = true
	}
	if missingArgs {
		solarizedRed.Printf("\nPlease specify the cluster and namespace and try again.\n")
		os.Exit(1)
	}

	// Check if the context exists in kubeconfig before confirmation prompt
	if !contextExists(foundContext) {
		availableContexts := getAvailableContexts()
		solarizedOrange.Printf("WARNING: The specified context '%s' does not exist in your kubeconfig.\n", foundContext)
		if len(availableContexts) > 0 {
			solarizedOrange.Printf("Available contexts: ")
			fmt.Printf("%s\n", colorizeItems(availableContexts))
		}
		solarizedOrange.Printf("Please check your --context value and try again.\n")
		os.Exit(1)
	}

	// Check if the namespace exists in the specified context
	if !namespaceExists(foundContext, foundNamespace) {
		availableNamespaces := getAvailableNamespaces(foundContext)
		solarizedOrange.Printf("WARNING: The specified namespace '%s' does not exist in context '%s'.\n", foundNamespace, foundContext)
		if len(availableNamespaces) > 0 {
			solarizedOrange.Printf("Available namespaces in context '%s': ", foundContext)
			fmt.Printf("%s\n", colorizeItems(availableNamespaces))
		}
		solarizedOrange.Printf("Please check your --namespace value and try again.\n")
		os.Exit(1)
	}

	// Print command summary and prompt for confirmation
	printCommandSummary(allArgs, foundContext, foundNamespace)
	if !confirmProductionContext(foundContext) {
		os.Exit(1)
	}

	// If all checks pass, execute the command.
	solarizedCyan.Printf("--- All checks passed. Executing command. ---\n")
	executeKubectl(allArgs...)
}

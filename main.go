package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
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
		solarizedRed.Print("❌ ERROR: ")
		solarizedOrange.Printf("Failed to read user input: %v\n", err)
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
	solarizedRed.Print("⚠️  DANGEROUS COMMAND DETECTED ⚠️\n")
	solarizedOrange.Print("You are about to run the following command:\n")
	fullCommandStr := fmt.Sprintf("kubectl %s", strings.Join(allArgs, " "))
	solarizedCyan.Printf("  %s\n", fullCommandStr)
	solarizedBlue.Print("on context ")
	solarizedCyan.Printf("%s", context)
	solarizedBlue.Print(" in namespace ")
	solarizedCyan.Printf("%s\n", namespace)
}

// confirmProductionContext prompts for context name if prod, else y/n confirmation.
func confirmProductionContext(context string) bool {
	if strings.Contains(strings.ToLower(context), "prod") {
		solarizedRed.Print("⚠️  WARNING: ")
		solarizedOrange.Print("You are about to run a command on a PRODUCTION context!\n")
		solarizedViolet.Printf("To proceed, please type the context name ('%s') and press Enter: ", context)
		reader := bufio.NewReader(os.Stdin)
		confirmation, err := reader.ReadString('\n')
		if err != nil {
			solarizedRed.Print("❌ ERROR: ")
			solarizedOrange.Printf("Failed to read user input: %v\n", err)
			return false
		}
		if strings.TrimSpace(confirmation) != context {
			solarizedRed.Print("❌ Aborted: ")
			solarizedOrange.Print("Context name did not match. Command will not be executed.\n")
			return false
		}
		return true
	} else {
		solarizedYellow.Print("Do you want to continue? (y/n): ")
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
		solarizedGreen.Print("→ Safe command detected. ")
		solarizedCyan.Println("Passing directly to kubectl...")
		executeKubectl(allArgs...)
		return
	}

	// Parse context and namespace from args.
	foundContext, foundNamespace, contextIsSet, namespaceIsSet := parseContextAndNamespace(kubectlArgs)

	// Enforce that flags are set for dangerous commands.
	missingArgs := false
	if !contextIsSet {
		solarizedRed.Print("❌ ERROR: ")
		solarizedOrange.Printf("The --context flag is mandatory for the dangerous command '%s'.\n", command)
		missingArgs = true
	}
	if !namespaceIsSet {
		solarizedRed.Print("❌ ERROR: ")
		solarizedOrange.Printf("The --namespace (-n) flag is mandatory for the dangerous command '%s'.\n", command)
		missingArgs = true
	}
	if missingArgs {
		solarizedYellow.Print("\nPlease specify the cluster and namespace and try again.\n")
		os.Exit(1)
	}

	// Check if the context exists in kubeconfig before confirmation prompt
	if !contextExists(foundContext) {
		solarizedRed.Print("❌ ERROR: ")
		solarizedOrange.Printf("The specified context '%s' does not exist in your kubeconfig.\n", foundContext)
		solarizedYellow.Print("Please check your --context value and try again.\n")
		os.Exit(1)
	}

	// Check if the namespace exists in the specified context
	if !namespaceExists(foundContext, foundNamespace) {
		solarizedRed.Print("❌ ERROR: ")
		solarizedOrange.Printf("The specified namespace '%s' does not exist in context '%s'.\n", foundNamespace, foundContext)
		solarizedYellow.Print("Please check your --namespace value and try again.\n")
		solarizedBlue.Printf("You can list available namespaces with: ")
		solarizedCyan.Printf("kubectl get namespaces --context %s\n", foundContext)
		os.Exit(1)
	}

	// Print command summary and prompt for confirmation
	printCommandSummary(allArgs, foundContext, foundNamespace)
	if !confirmProductionContext(foundContext) {
		os.Exit(1)
	}

	// If all checks pass, execute the command.
	color.Cyan("--- All checks passed. Executing command. ---")
	executeKubectl(allArgs...)
}

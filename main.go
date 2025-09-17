package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
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
	yellow := color.New(color.FgYellow)
	cyan := color.New(color.FgCyan)
	yellow.Println("You are about to run the following command:")
	fullCommandStr := fmt.Sprintf("kubectl %s", strings.Join(allArgs, " "))
	cyan.Printf("  %s\n", fullCommandStr)
	fmt.Printf("on context ")
	cyan.Printf("%s", context)
	fmt.Printf(" in namespace ")
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
		color.New(color.FgGreen).Printf("--> Safe command detected. Passing directly to kubectl...\n")
		executeKubectl(allArgs...)
		return
	}

	// Parse context and namespace from args.
	foundContext, foundNamespace, contextIsSet, namespaceIsSet := parseContextAndNamespace(kubectlArgs)

	// Enforce that flags are set for dangerous commands.
	missingArgs := false
	if !contextIsSet {
		color.Red("ERROR: The --context flag is mandatory for the dangerous command '%s'.", command)
		missingArgs = true
	}
	if !namespaceIsSet {
		color.New(color.FgYellow).Printf("WARNING: The --namespace (-n) flag is mandatory for the dangerous command '%s'.\n", command)
		missingArgs = true
	}
	if missingArgs {
		fmt.Println("\nPlease specify the cluster and namespace and try again.")
		os.Exit(1)
	}

	// Check if the context exists in kubeconfig before confirmation prompt
	if !contextExists(foundContext) {
		color.Red("ERROR: The specified context '%s' does not exist in your kubeconfig.", foundContext)
		fmt.Println("Please check your --context value and try again.")
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

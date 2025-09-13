package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.comcom/fatih/color"
)

// Define the list of dangerous commands.
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

func main() {
	// os.Args contains all command-line arguments.
	// os.Args[0] is the program name ("kubectl-safe").
	// os.Args[1] should be the dangerous command ("apply", "delete", etc.).
	if len(os.Args) < 2 {
		color.Red("Error: 'safe' plugin requires a command to run.")
		fmt.Println("Example: kubectl safe apply -f my-app.yaml")
		os.Exit(1)
	}

	dangerousCmd := os.Args[1]
	// The rest of the arguments are passed to kubectl.
	kubectlArgs := os.Args[2:]

	// Check if the command is in our list.
	if !dangerousCommands[dangerousCmd] {
		color.Red("Error: '%s' is not a supported dangerous command.", dangerousCmd)
		os.Exit(1)
	}

	// --- The same logic as our script, now in Go ---

	// 1. Parse arguments to find context and namespace.
	var foundContext, foundNamespace string
	contextIsSet, namespaceIsSet := false, false

	for i := 0; i < len(kubectlArgs); i++ {
		arg := kubectlArgs[i]
		// Handle --flag=value format
		if strings.HasPrefix(arg, "--context=") {
			contextIsSet = true
			foundContext = strings.SplitN(arg, "=", 2)[1]
		} else if strings.HasPrefix(arg, "--namespace=") {
			namespaceIsSet = true
			foundNamespace = strings.SplitN(arg, "=", 2)[1]
		} else if arg == "--context" || arg == "-n" || arg == "--namespace" {
			// Handle --flag value format, making sure we don't go out of bounds.
			if (i + 1) < len(kubectlArgs) {
				if arg == "--context" {
					contextIsSet = true
					foundContext = kubectlArgs[i+1]
				} else { // -n or --namespace
					namespaceIsSet = true
					foundNamespace = kubectlArgs[i+1]
				}
			}
		}
	}

	// 2. Enforce that flags are set.
	missingArgs := false
	if !contextIsSet {
		color.Red("ERROR: The --context flag is mandatory.")
		missingArgs = true
	}
	if !namespaceIsSet {
		color.New(color.FgYellow).Println("WARNING: The --namespace (-n) flag is mandatory.")
		missingArgs = true
	}
	if missingArgs {
		fmt.Println("\nPlease specify the cluster and namespace and try again.")
		os.Exit(1)
	}

	// 3. Final confirmation prompt.
	yellow := color.New(color.FgYellow)
	cyan := color.New(color.FgCyan)

	yellow.Println("You are about to run the following command:")
	// Reconstruct the full command string for display.
	fullCommandStr := fmt.Sprintf("kubectl %s %s", dangerousCmd, strings.Join(kubectlArgs, " "))
	cyan.Printf("  %s\n", fullCommandStr)
	fmt.Printf("on context ")
	cyan.Printf("%s\n", foundContext)
	yellow.Print("Do you want to continue? (y/n): ")

	if !askForConfirmation() {
		color.Red("Aborted by user.")
		os.Exit(1)
	}

	// 4. Special production alert.
	if strings.Contains(strings.ToLower(foundContext), "prod") {
		color.New(color.FgRed).Add(color.Bold).Println("#################### PRODUCTION ALERT ####################")
		fmt.Println("This is a PRODUCTION context. Please type the context name to confirm:")
		cyan.Printf("%s\n> ", foundContext)

		reader := bufio.NewReader(os.Stdin)
		prodConfirmation, _ := reader.ReadString('\n')
		if strings.TrimSpace(prodConfirmation) != foundContext {
			color.Red("Confirmation failed. Aborted by user.")
			os.Exit(1)
		}
		fmt.Println("Production confirmation received. Proceeding...")
	}

	// 5. If all checks pass, execute the real kubectl command.
	fmt.Println("--- All checks passed. Executing command. ---")
	// Prepend the dangerous command back to the front of the arguments list.
	finalArgs := append([]string{dangerousCmd}, kubectlArgs...)
	cmd := exec.Command("kubectl", finalArgs...)

	// Connect the command's output/error to our terminal so we can see it.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Run the command and exit with its status code.
	if err := cmd.Run(); err != nil {
		// This handles cases where kubectl itself returns an error.
		os.Exit(1)
	}
}

// askForConfirmation reads a single y/n from the console.
func askForConfirmation() bool {
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	return strings.ToLower(strings.TrimSpace(response)) == "y"
}

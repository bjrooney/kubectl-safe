package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
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
	if len(os.Args) < 2 {
		color.Red("Error: 'safe' plugin requires a command to run.")
		fmt.Println("Example: kubectl safe apply -f my-app.yaml")
		os.Exit(1)
	}

	dangerousCmd := os.Args[1]
	kubectlArgs := os.Args[2:]

	if !dangerousCommands[dangerousCmd] {
		color.Red("Error: '%s' is not a supported dangerous command.", dangerousCmd)
		os.Exit(1)
	}

	// 1. Parse arguments to find context and namespace.
	var foundContext, foundNamespace string
	contextIsSet, namespaceIsSet := false, false

	for i := 0; i < len(kubectlArgs); i++ {
		arg := kubectlArgs[i]
		if strings.HasPrefix(arg, "--context=") {
			contextIsSet = true
			foundContext = strings.SplitN(arg, "=", 2)[1]
		} else if strings.HasPrefix(arg, "--namespace=") {
			namespaceIsSet = true
			foundNamespace = strings.SplitN(arg, "=", 2)[1]
		} else if arg == "--context" || arg == "-n" || arg == "--namespace" {
			if (i + 1) < len(kubectlArgs) {
				if arg == "--context" {
					contextIsSet = true
					foundContext = kubectlArgs[i+1]
				} else {
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
	fullCommandStr := fmt.Sprintf("kubectl %s %s", dangerousCmd, strings.Join(kubectlArgs, " "))
	cyan.Printf("  %s\n", fullCommandStr)
	// UPDATED SECTION: Now displays context AND namespace
	fmt.Printf("on context ")
	cyan.Printf("%s", foundContext)
	fmt.Printf(" in namespace ")
	cyan.Printf("%s\n", foundNamespace)
	yellow.Print("Do you want to continue? (y/n): ")

	if !askForConfirmation() {
		color.Red("Aborted by user.")
		os.Exit(1)
	}

	// 4. Special production alert.
	if strings.Contains(strings.ToLower(foundContext), "prod") {
		redBold := color.New(color.FgRed).Add(color.Bold)

		redBold.Println("#################### PRODUCTION ALERT ####################")
		fmt.Println("This is a PRODUCTION context. Please review the details carefully.")
		// UPDATED SECTION: Also shows command, context, and namespace in prod alert
		fmt.Printf("  Command:   ")
		cyan.Printf("%s %s\n", dangerousCmd, strings.Join(kubectlArgs, " "))
		fmt.Printf("  Context:   ")
		cyan.Printf("%s\n", foundContext)
		fmt.Printf("  Namespace: ")
		cyan.Printf("%s\n\n", foundNamespace)
		fmt.Println("To proceed, please type the full context name to confirm:")
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
	finalArgs := append([]string{dangerousCmd}, kubectlArgs...)
	cmd := exec.Command("kubectl", finalArgs...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}

// askForConfirmation reads a single y/n from the console.
func askForConfirmation() bool {
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	return strings.ToLower(strings.TrimSpace(response)) == "y"
}

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

func main() {
	if len(os.Args) < 2 {
		// If no command is given, just run kubectl help.
		executeKubectl()
		return
	}

	command := os.Args[1]
	kubectlArgs := os.Args[2:]
	allArgs := os.Args[1:] // The full command and args.

	// --- MODIFIED SECTION: The "Smart" Logic ---
	// Check if the command is NOT in our dangerous list.
	if !dangerousCommands[command] {
		// This is a safe command, so we execute it directly and exit.
		color.New(color.FgGreen).Printf("--> Safe command detected. Passing directly to kubectl...\n")
		executeKubectl(allArgs...)
		return // Exit successfully.
	}
	// --- END OF MODIFIED SECTION ---

	// If we reach here, it's a dangerous command. All the safety checks now apply.
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

	// 2. Enforce that flags are set for dangerous commands.
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

	// 3. Final confirmation prompt.
	yellow := color.New(color.FgYellow)
	cyan := color.New(color.FgCyan)

	yellow.Println("You are about to run the following command:")
	fullCommandStr := fmt.Sprintf("kubectl %s", strings.Join(allArgs, " "))
	cyan.Printf("  %s\n", fullCommandStr)
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
		// ... (production alert logic is the same)
	}

	// 5. If all checks pass, execute the command.
	color.Cyan("--- All checks passed. Executing command. ---")
	executeKubectl(allArgs...)
}

// askForConfirmation reads a single y/n from the console.
func askForConfirmation() bool {
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	return strings.ToLower(strings.TrimSpace(response)) == "y"
}

package safe

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/pflag"
)

// Solarized color scheme
var (
	solarizedRed    = color.New(color.FgHiRed).Add(color.Bold)
	solarizedOrange = color.New(color.FgHiYellow).Add(color.Bold)
	solarizedYellow = color.New(color.FgYellow)
	solarizedGreen  = color.New(color.FgGreen)
	solarizedCyan   = color.New(color.FgCyan)
	solarizedBlue   = color.New(color.FgBlue)
	solarizedViolet = color.New(color.FgMagenta)
)

// Version will be set at build time
var Version = "dev"

// ModifyingCommandMap defines commands that are modifying by themselves or with specific subcommands.
var ModifyingCommandMap = map[string][]string{
	"apply": {}, "create": {}, "delete": {}, "edit": {}, "patch": {}, "replace": {}, "scale": {},
	"cordon": {}, "uncordon": {}, "drain": {}, "taint": {},
	"autoscale": {}, "expose": {},
	"auth":        {"reconcile"},
	"certificate": {"approve", "deny"},
	"rollout":     {"restart", "undo"},
	"set":         {"env", "image", "resources", "selector", "subject", "serviceaccount"},
}

// ModifyingByFlagMap defines commands that are ONLY modifying when specific flags are used.
var ModifyingByFlagMap = map[string][]string{
	"annotate": {"-f", "--filename"},
	"label":    {"-f", "--filename"},
}

// printColoredList prints a list of items with alternating Solarized colors
func printColoredList(items []string) {
	colors := []*color.Color{solarizedCyan, solarizedGreen, solarizedYellow, solarizedViolet, solarizedBlue}
	for i, item := range items {
		if i > 0 {
			fmt.Print(", ")
		}
		colors[i%len(colors)].Print(item)
	}
	fmt.Println()
}

// Execute is the main entry point for the plugin.
func Execute() error {
	args := os.Args[1:]

	if len(args) == 0 {
		return showUsage()
	}
	if len(args) == 1 && (args[0] == "--version" || args[0] == "-v") {
		solarizedBlue.Print("kubectl-safe version ")
		solarizedCyan.Printf("%s\n", Version)
		return nil
	}

	flagSet := pflag.NewFlagSet("safe-flags", pflag.ContinueOnError)
	var context, namespace string
	flagSet.StringVarP(&context, "context", "c", "", "Kubernetes context")
	flagSet.StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace")
	flagSet.Usage = func() {}
	_ = flagSet.Parse(args)

	commandAndArgs := flagSet.Args()

	if !isModifyingCommand(commandAndArgs, args) {
		solarizedGreen.Print("→ Read-only command detected. ")
		solarizedCyan.Println("Passing directly to kubectl...")
		return executeKubectl(args)
	}

	if err := validateRequiredFlags(flagSet, commandAndArgs); err != nil {
		handleValidationError(err)
		return err
	}

	context, _ = flagSet.GetString("context")
	namespace, _ = flagSet.GetString("namespace")

	if err := showConfirmation(args, context, namespace); err != nil {
		return err
	}

	return executeKubectl(args)
}

// isModifyingCommand checks if a command will modify cluster state.
func isModifyingCommand(commandAndArgs []string, rawArgs []string) bool {
	if len(commandAndArgs) == 0 {
		return false
	}
	command := commandAndArgs[0]

	// 1. Check for modifying by subcommand
	if modifyingSubcommands, exists := ModifyingCommandMap[command]; exists {
		if len(modifyingSubcommands) == 0 {
			return true // Command is always modifying
		}
		if len(commandAndArgs) > 1 {
			subcommand := commandAndArgs[1]
			if slices.Contains(modifyingSubcommands, subcommand) {
				return true // Found a modifying subcommand
			}
		}
	}

	// 2. Check for modifying by flag
	if modifyingFlags, exists := ModifyingByFlagMap[command]; exists {
		for _, arg := range rawArgs {
			for _, flag := range modifyingFlags {
				if arg == flag || strings.HasPrefix(arg, flag+"=") {
					return true
				}
			}
		}
	}

	return false
}

// validateRequiredFlags ensures required flags are present for modifying commands.
func validateRequiredFlags(fs *pflag.FlagSet, commandAndArgs []string) error {
	hasContext := fs.Changed("context")
	hasNamespace := fs.Changed("namespace")

	isNamespaceResourceCommand := false
	if len(commandAndArgs) > 1 {
		resourceType := strings.ToLower(commandAndArgs[1])
		if resourceType == "namespace" || resourceType == "namespaces" || resourceType == "ns" {
			isNamespaceResourceCommand = true
		}
	}

	var missing []string
	if !hasContext {
		missing = append(missing, "--context")
	}
	if !hasNamespace && !isNamespaceResourceCommand {
		missing = append(missing, "--namespace")
	}
	if len(missing) > 0 {
		return fmt.Errorf("modifying command requires explicit %s flag(s)", strings.Join(missing, " and "))
	}

	contextValue, _ := fs.GetString("context")
	namespaceValue, _ := fs.GetString("namespace")

	availableContexts, err := getKubeconfigContexts()
	if err != nil {
		return fmt.Errorf("failed to get available contexts: %w", err)
	}
	if !slices.Contains(availableContexts, contextValue) {
		return fmt.Errorf("context '%s' not found in kubeconfig. Available contexts: %s",
			contextValue, strings.Join(availableContexts, ", "))
	}

	if hasNamespace && !isNamespaceResourceCommand {
		availableNamespaces, err := getNamespacesInContext(contextValue)
		if err != nil {
			return fmt.Errorf("failed to get namespaces for context '%s': %w", contextValue, err)
		}
		isCreateCommand := commandAndArgs[0] == "create"
		if !slices.Contains(availableNamespaces, namespaceValue) && !isCreateCommand {
			return fmt.Errorf("namespace '%s' not found in context '%s'. Available namespaces: %s",
				namespaceValue, contextValue, strings.Join(availableNamespaces, ", "))
		}
	}
	return nil
}

// handleValidationError displays colored error messages and exits.
func handleValidationError(err error) {
	errMsg := err.Error()
	if strings.Contains(errMsg, "requires explicit") {
		solarizedRed.Print("❌ ERROR: ")
		solarizedOrange.Print("Modifying command requires explicit --context and --namespace flags.\n")
		solarizedYellow.Print("This ensures you're targeting the correct cluster and namespace.\n")
	} else if strings.Contains(errMsg, "not found in kubeconfig") {
		parts := strings.Split(errMsg, ". Available contexts: ")
		if len(parts) == 2 {
			solarizedYellow.Print("✋ WARNING: ")
			contextPart := strings.Replace(parts[0], "context '", "", 1)
			contextPart = strings.Replace(contextPart, "' not found in kubeconfig", "", 1)
			solarizedOrange.Printf("Context '%s' not found in kubeconfig.\n", contextPart)
			solarizedBlue.Print("Available contexts: ")
			printColoredList(strings.Split(parts[1], ", "))
		}
	} else if strings.Contains(errMsg, "not found in context") {
		parts := strings.Split(errMsg, ". Available namespaces: ")
		if len(parts) == 2 {
			solarizedYellow.Print("✋ WARNING: ")
			solarizedOrange.Print(parts[0] + ".\n")
			solarizedBlue.Print("Available namespaces: ")
			printColoredList(strings.Split(parts[1], ", "))
		}
	} else {
		solarizedRed.Print("❌ ERROR: ")
		solarizedOrange.Printf("%s\n", errMsg)
	}
	os.Exit(1)
}

// showConfirmation displays an interactive prompt for modifying commands.
func showConfirmation(args []string, context, namespace string) error {
	solarizedOrange.Print("⚠️  MODIFYING COMMAND DETECTED ⚠️\n\n")
	solarizedYellow.Print("You are about to execute: ")
	solarizedCyan.Printf("kubectl %s\n\n", strings.Join(args, " "))

	solarizedBlue.Print("Target Details:\n")
	solarizedBlue.Print("  Context:   ")
	solarizedCyan.Printf("%s\n", context)
	if namespace != "" {
		solarizedBlue.Print("  Namespace: ")
		solarizedCyan.Printf("%s\n", namespace)
	}
	fmt.Println()

	if strings.Contains(strings.ToLower(context), "prod") {
		solarizedRed.Print("🚨 PRODUCTION CONTEXT WARNING! 🚨\n")
		solarizedOrange.Print("This command targets a PRODUCTION context!\n\n")
		solarizedViolet.Printf("To proceed, type the context name ('%s') and press Enter: ", context)
		reader := bufio.NewReader(os.Stdin)
		confirmation, _ := reader.ReadString('\n')
		if strings.TrimSpace(confirmation) != context {
			solarizedRed.Println("❌ Aborted: Context name mismatch.")
			return fmt.Errorf("operation cancelled by user")
		}
		solarizedGreen.Println("Production context confirmed.")
		return nil
	}

	solarizedYellow.Print("This operation will modify the state of the cluster.\n")
	solarizedViolet.Print("Are you sure you want to continue? (yes/no): ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	if response != "yes" && response != "y" {
		solarizedRed.Println("Operation cancelled.")
		return fmt.Errorf("operation cancelled by user")
	}
	solarizedGreen.Println("Proceeding with operation...")
	return nil
}

// executeKubectl runs the actual kubectl command.
func executeKubectl(args []string) error {
	cmd := exec.Command("kubectl", args...)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
	return cmd.Run()
}

// getKubeconfigContexts returns the list of available contexts.
func getKubeconfigContexts() ([]string, error) {
	cmd := exec.Command("kubectl", "config", "get-contexts", "--output=name")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var validContexts []string
	outputStr := strings.TrimSpace(string(output))

	// If there's no output, return empty slice (not nil)
	if outputStr == "" {
		return []string{}, nil
	}

	for _, c := range strings.Split(outputStr, "\n") {
		if t := strings.TrimSpace(c); t != "" {
			validContexts = append(validContexts, t)
		}
	}

	// Ensure we return empty slice instead of nil if no valid contexts found
	if validContexts == nil {
		validContexts = []string{}
	}

	return validContexts, nil
}

// getNamespacesInContext returns the list of namespaces in a context.
func getNamespacesInContext(context string) ([]string, error) {
	cmd := exec.Command("kubectl", "get", "namespaces", "--output=name", "--context", context)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var namespaces []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if trimmed := strings.TrimPrefix(strings.TrimSpace(line), "namespace/"); trimmed != "" {
			namespaces = append(namespaces, trimmed)
		}
	}
	return namespaces, nil
}

// showUsage displays help information.
func showUsage() error {
	solarizedBlue.Print("kubectl-safe: ")
	solarizedCyan.Print("A safety net for kubectl commands\n")
	solarizedYellow.Printf("Version: %s\n\n", Version)
	solarizedViolet.Print("Usage:\n")
	fmt.Print("  kubectl safe <kubectl-command> [flags]\n\n")
	fmt.Print("This plugin wraps kubectl to enforce safety checks for operations that modify cluster state.\n\n")

	var always, conditional, byFlag []string
	for cmd, subCmds := range ModifyingCommandMap {
		if len(subCmds) == 0 {
			always = append(always, cmd)
		} else {
			conditional = append(conditional, fmt.Sprintf("%s (%s)", cmd, strings.Join(subCmds, ", ")))
		}
	}
	for cmd, flags := range ModifyingByFlagMap {
		byFlag = append(byFlag, fmt.Sprintf("%s (with %s)", cmd, strings.Join(flags, "/")))
	}
	sort.Strings(always)
	sort.Strings(conditional)
	sort.Strings(byFlag)

	solarizedOrange.Print("Always-modifying commands:\n")
	fmt.Printf("  %s\n\n", strings.Join(always, ", "))
	solarizedOrange.Print("Commands with modifying subcommands:\n")
	fmt.Printf("  %s\n\n", strings.Join(conditional, ", "))
	solarizedOrange.Print("Conditionally modifying commands (by flag):\n")
	fmt.Printf("  %s\n\n", strings.Join(byFlag, ", "))
	return nil
}

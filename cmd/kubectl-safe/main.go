// Package main is the entry point for the kubectl-safe plugin.
//
// kubectl-safe is a Krew plugin that provides an interactive safety net for
// dangerous kubectl commands. It acts as a wrapper around destructive kubectl
// operations to prevent common mistakes by requiring explicit context and
// namespace flags and showing interactive confirmation prompts.
//
// The plugin integrates seamlessly with kubectl as a plugin, allowing users
// to run:
//   kubectl safe <command> [flags]
//
// instead of:
//   kubectl <command> [flags]
//
// For dangerous operations, the plugin will enforce safety checks before
// executing the actual kubectl command.
package main

import (
	"fmt"
	"os"

	"github.com/bjrooney/kubectl-safe/pkg/safe"
)

// main is the entry point for the kubectl-safe binary.
//
// This function simply delegates to the safe.Execute() function which contains
// the main plugin logic. Any errors from the execution are printed to stderr
// and cause the program to exit with status code 1.
func main() {
	if err := safe.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
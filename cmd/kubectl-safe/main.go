package main

import (
	"fmt"
	"os"

	"github.com/bjrooney/kubectl-safe/pkg/safe"
)

func main() {
	if err := safe.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
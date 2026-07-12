package main

import (
	"fmt"
	"os"
)

// runGCL dispatches `vet gcl <run|gate|trace>`. Implemented in M3.
func runGCL(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "vet gcl: missing subcommand (run|gate|trace)")
		os.Exit(2)
	}
	sub := args[0]
	// TODO(M3): dispatch to internal/gclrun|gclgate|gcltrace.
	fmt.Fprintf(os.Stderr, "vet gcl %s: not implemented yet (M3 pending)\n", sub)
	os.Exit(3)
}

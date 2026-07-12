package main

import (
	"fmt"
	"os"
)

// runCheck dispatches `vet check <name>`. Individual checks are implemented
// in internal/check/* and wired here as M2 progresses.
func runCheck(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "vet check: missing subcommand (frontmatter|aiops|assessment|gcl|links|eval)")
		os.Exit(2)
	}
	name := args[0]
	rest := args[1:]
	// TODO(M2): dispatch to internal/check/* implementations.
	fmt.Fprintf(os.Stderr, "vet check %s: not implemented yet (M2 pending)\n", name)
	_ = rest
	os.Exit(3)
}

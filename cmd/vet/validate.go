package main

import (
	"fmt"
	"os"
)

// runValidate dispatches `vet validate`. Implemented in M2.7.
func runValidate(args []string) {
	// TODO(M2.7): delegate to internal/validate.
	fmt.Fprintln(os.Stderr, "vet validate: not implemented yet (M2.7 pending)")
	_ = args
	os.Exit(3)
}

// Command vet is the unified CLI for the ve-skills repository.
//
// It consolidates the Python verification/CI scripts (scripts/*.py) into a
// single statically-linked Go binary so that engineers and AI agents can run
// the checks without a Python interpreter.
//
// Subcommands:
//
//	vet version                 - print version information
//	vet check <name>            - run a validation check (frontmatter/aiops/assessment/gcl/links/eval)
//	vet validate                - run the local validation suite
//	vet gcl run|gate|trace      - run GCL loop operations
package main

import (
	"flag"
	"fmt"
	"os"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "0.0.0-dev"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "version":
		runVersion(args)
	case "check":
		runCheck(args)
	case "validate":
		runValidate(args)
	case "gcl":
		runGCL(args)
	default:
		fmt.Fprintf(os.Stderr, "vet: unknown command %q\n", cmd)
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `vet - ve-skills unified CLI

Usage:
  vet version
  vet check <frontmatter|aiops|assessment|gcl|links|eval> [flags]
  vet validate [flags]
  vet gcl <run|gate|trace> [flags]

Use "vet <command> -h" for more information.`)
}

func runVersion(args []string) {
	fs := flag.NewFlagSet("version", flag.ExitOnError)
	fs.Parse(args)
	fmt.Printf("vet %s\n", version)
}

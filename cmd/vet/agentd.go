package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/agentd"
)

func runAgentd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "vet agentd: missing subcommand (serve)")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  vet agentd serve [flags]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Flags:")
		fmt.Fprintln(os.Stderr, "  --port              HTTP port (default: 8080)")
		fmt.Fprintln(os.Stderr, "  --max-concurrent    Max concurrent runs (default: 5)")
		fmt.Fprintln(os.Stderr, "  --root              Repo root directory")
		os.Exit(2)
	}

	sub := args[0]
	rest := args[1:]
	switch sub {
	case "serve":
		runAgentdServe(rest)
	default:
		fmt.Fprintf(os.Stderr, "vet agentd: unknown subcommand %q\n", sub)
		os.Exit(2)
	}
}

func runAgentdServe(args []string) {
	fs := flag.NewFlagSet("agentd serve", flag.ExitOnError)
	port := fs.Int("port", 8080, "HTTP port")
	maxConcurrent := fs.Int("max-concurrent", 5, "max concurrent runs")
	root := fs.String("root", repoRoot(), "repo root")
	fs.Parse(args)

	addr := fmt.Sprintf(":%d", *port)
	server := agentd.NewServer(*root, addr, *maxConcurrent)

	ctx := context.Background()
	fmt.Fprintf(os.Stderr, "[INFO] agentd | starting server on %s (max_concurrent=%d)\n", addr, *maxConcurrent)

	if err := server.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] agentd | server failed: %v\n", err)
		os.Exit(1)
	}
}

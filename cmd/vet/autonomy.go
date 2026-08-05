package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/autonomy"
)

func runAutonomy(args []string) {
	fs := flag.NewFlagSet("autonomy", flag.ExitOnError)
	envelopePath := fs.String("envelope", "", "path to autonomy-envelope.yaml")
	n := fs.Int("n", 5, "number of synthetic incidents to run")
	timeout := fs.Duration("timeout", 60*time.Second, "test timeout")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `vet autonomy - L4 autonomous domain testing

Usage:
  vet autonomy test --envelope <path> [--n 5] [--timeout 60s]

Flags:`)
		fs.PrintDefaults()
	}

	subcmd := ""
	flagArgs := args
	for i, arg := range args {
		if arg[0] != '-' {
			subcmd = arg
			flagArgs = append(args[:i:i], args[i+1:]...)
			break
		}
	}

	fs.Parse(flagArgs)

	if subcmd == "" {
		fs.Usage()
		os.Exit(2)
	}

	switch subcmd {
	case "test":
		runAutonomyTest(*envelopePath, *n, *timeout)
	case "loadtest":
		runAutonomyLoadTest(*envelopePath, *n, *timeout)
	default:
		fmt.Fprintf(os.Stderr, "vet autonomy: unknown subcommand %q\n", subcmd)
		fs.Usage()
		os.Exit(2)
	}
}

func runAutonomyTest(envelopePath string, n int, timeout time.Duration) {
	if envelopePath == "" {
		fmt.Fprintln(os.Stderr, "error: --envelope is required")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fmt.Printf("Running %d synthetic incidents against envelope: %s\n", n, envelopePath)
	fmt.Println("---------------------------------------------------")

	report, err := autonomy.RunNConsecutiveIncidentsPath(ctx, n, envelopePath, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Results:\n")
	fmt.Printf("  Total incidents:  %d\n", report.TotalIncidents)
	fmt.Printf("  Passed:           %d\n", report.Passed)
	fmt.Printf("  Failed:           %d\n", report.Failed)
	fmt.Printf("  Prompts (L4=0):   %d\n", report.Prompts)
	fmt.Printf("  SLO Violations:   %d\n", report.SLOViolations)
	fmt.Printf("  Duration:         %v\n", report.Duration.Round(time.Millisecond))
	fmt.Println("---------------------------------------------------")

	if report.Failed > 0 || report.Prompts > 0 || report.SLOViolations > 0 {
		fmt.Println("FAIL — L4 criteria not met")
		os.Exit(1)
	}

	fmt.Printf("PASS — %d/%d incidents completed, 0 prompts, SLO maintained\n", report.Passed, report.TotalIncidents)
}

func runAutonomyLoadTest(envelopePath string, n int, timeout time.Duration) {
	if envelopePath == "" {
		fmt.Fprintln(os.Stderr, "error: --envelope is required")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	env, err := autonomy.LoadEnvelope(envelopePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading envelope: %v\n", err)
		os.Exit(1)
	}

	ev, err := autonomy.RunL4LoadTest(ctx, n, env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("L4 runtime evidence (synthetic):")
	fmt.Println("---------------------------------------------------")
	for _, it := range ev.Items {
		status := "PASS"
		if !it.Passed {
			status = "FAIL"
		}
		fmt.Printf("  [%s] item %s — %s\n         %s\n", status, it.ID, it.Description, it.Detail)
	}
	if ev.Rollbacks > 0 {
		fmt.Printf("  rollbacks applied: %d\n", ev.Rollbacks)
	}
	fmt.Println("---------------------------------------------------")

	if !ev.Passed {
		fmt.Println("FAIL — L4 runtime evidence incomplete")
		os.Exit(1)
	}
	fmt.Println("PASS — L4 runtime evidence items ③⑤⑥⑦ met")
}

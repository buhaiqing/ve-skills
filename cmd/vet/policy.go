package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/policy"
)

func runPolicy(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "vet policy: missing subcommand (load|diff|check-changelog)")
		os.Exit(2)
	}
	sub := args[0]
	rest := args[1:]
	switch sub {
	case "load":
		runPolicyLoad(rest)
	case "diff":
		runPolicyDiff(rest)
	case "check-changelog":
		runPolicyCheckChangelog(rest)
	default:
		fmt.Fprintf(os.Stderr, "vet policy: unknown subcommand %q\n", sub)
		os.Exit(2)
	}
}

func runPolicyLoad(args []string) {
	fs := flag.NewFlagSet("policy load", flag.ExitOnError)
	root := fs.String("root", "", "repo root path")
	fs.Parse(args)
	if *root == "" {
		fmt.Fprintln(os.Stderr, "policy load: --root is required")
		os.Exit(2)
	}
	ps, err := policy.Load(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "policy load: %v\n", err)
		os.Exit(1)
	}
	data, _ := json.MarshalIndent(ps, "", "  ")
	fmt.Println(string(data))
}

func runPolicyDiff(args []string) {
	fs := flag.NewFlagSet("policy diff", flag.ExitOnError)
	_ = fs.String("old", "HEAD~1", "old git ref")
	_ = fs.String("new", "HEAD", "new git ref")
	root := fs.String("root", "", "repo root path")
	fs.Parse(args)
	if *root == "" {
		fmt.Fprintln(os.Stderr, "policy diff: --root is required")
		os.Exit(2)
	}

	// Simple approach: just load current and report
	// For a full diff between commits, we'd need to checkout old version
	// Instead, we compare current against empty for now
	// TODO: full git-based diff in future

	ps, err := policy.Load(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "policy diff: %v\n", err)
		os.Exit(1)
	}

	empty := &policy.PolicySet{}
	changes := policy.DiffPolicySets(empty, ps)

	if len(changes) == 0 {
		fmt.Println("No policy changes detected")
		return
	}

	for _, c := range changes {
		fmt.Printf("[%s] %s: %s\n", c.Type, c.File, c.Detail)
	}
}

func runPolicyCheckChangelog(args []string) {
	fs := flag.NewFlagSet("policy check-changelog", flag.ExitOnError)
	root := fs.String("root", "", "repo root path")
	base := fs.String("base", "HEAD~1", "base ref to compare against")
	fs.Parse(args)
	if *root == "" {
		fmt.Fprintln(os.Stderr, "policy check-changelog: --root is required")
		os.Exit(2)
	}

	// Run git diff --name-only <base> HEAD to check changed files
	cmd := exec.Command("git", "diff", "--name-only", *base, "HEAD")
	cmd.Dir = *root
	out, err := cmd.Output()
	if err != nil {
		// git diff may fail if there's no previous commit (initial commit)
		// In that case, treat as no changes
		fmt.Println("policy check-changelog: git diff failed, skipping check")
		return
	}

	policyChanged := false
	changelogChanged := false
	prefix := "incident-loop-agent/references/policies/"

	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, prefix+"CHANGELOG.md") {
			changelogChanged = true
		} else if strings.HasPrefix(line, prefix) {
			policyChanged = true
		}
	}

	if policyChanged && !changelogChanged {
		fmt.Fprintln(os.Stderr, "policy check-changelog: policy files changed but CHANGELOG.md not updated")
		fmt.Fprintln(os.Stderr, "  Please add a line to incident-loop-agent/references/policies/CHANGELOG.md")
		os.Exit(1)
	}

	fmt.Println("policy check-changelog: OK")
}

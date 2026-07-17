package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/reflexion/promote"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/reflexion/transpile"
)

func runReflexion(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "vet reflexion: missing subcommand (promote|check|transpile)")
		os.Exit(2)
	}
	sub := args[0]
	rest := args[1:]
	switch sub {
	case "promote":
		runReflexionPromote(rest)
	case "check":
		runReflexionCheck(rest)
	case "transpile":
		runReflexionTranspile(rest)
	default:
		fmt.Fprintf(os.Stderr, "vet reflexion: unknown subcommand %q\n", sub)
		os.Exit(2)
	}
}

func runReflexionPromote(args []string) {
	fs := flag.NewFlagSet("reflexion promote", flag.ExitOnError)
	patternsPath := fs.String("patterns", "", "path to failure-patterns.md")
	fs.Parse(args)

	if *patternsPath == "" {
		fmt.Fprintln(os.Stderr, "reflexion promote: --patterns is required")
		os.Exit(2)
	}

	patterns := readFailurePatterns(*patternsPath)
	if len(patterns) == 0 {
		fmt.Fprintln(os.Stderr, "reflexion promote: no patterns found in file")
		os.Exit(1)
	}

	fmt.Println("category | skill | pattern | count | level")
	for _, p := range patterns {
		lvl := promote.LevelOf(promote.Pattern{
			Category: p.Category,
			Skill:    p.Skill,
			Pattern:  p.Pattern,
			Count:    p.Count,
		})
		fmt.Printf("%s | %s | %s | %d | %s\n",
			p.Category, p.Skill, p.Pattern, p.Count, levelName(lvl))
	}
}

func runReflexionCheck(args []string) {
	fs := flag.NewFlagSet("reflexion check", flag.ExitOnError)
	patternsPath := fs.String("patterns", "", "path to failure-patterns.md")
	planPath := fs.String("plan", "", "path to plan JSON")
	fs.Parse(args)

	if *patternsPath == "" || *planPath == "" {
		fmt.Fprintln(os.Stderr, "reflexion check: --patterns and --plan are required")
		os.Exit(2)
	}

	patterns := readFailurePatterns(*patternsPath)
	hasHard := false

	for _, p := range patterns {
		// Only check patterns with count >= 10 (Constraint or Hard)
		if p.Count < 10 {
			continue
		}
		pat := promote.Pattern{
			Category: p.Category,
			Skill:    p.Skill,
			Pattern:  p.Pattern,
			Count:    p.Count,
		}
		lvl, err := promote.Enforce(pat, 1)
		switch lvl {
		case promote.LevelHard:
			fmt.Printf("HARD: %s/%s (count=%d): %v\n", p.Category, p.Pattern, p.Count, err)
			hasHard = true
		case promote.LevelConstraint:
			fmt.Printf("WARNING: constraint violated: %s/%s (count=%d)\n", p.Category, p.Pattern, p.Count)
		}
	}

	if hasHard {
		os.Exit(1)
	}
}

func runReflexionTranspile(args []string) {
	fs := flag.NewFlagSet("reflexion transpile", flag.ExitOnError)
	patternsPath := fs.String("patterns", "", "path to failure-patterns.md")
	outPath := fs.String("out", "", "output path for guardrails.yaml")
	fs.Parse(args)

	if *patternsPath == "" || *outPath == "" {
		fmt.Fprintln(os.Stderr, "reflexion transpile: --patterns and --out are required")
		os.Exit(2)
	}

	n, err := transpile.TranspileFile(*patternsPath, *outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "reflexion transpile: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Upgraded %d patterns to guardrails\n", n)
}

// readFailurePatterns reads a markdown file and returns FailurePattern entries.
// Reuses the transpile package's pattern extraction.
func readFailurePatterns(path string) []transpile.FailurePattern {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "reflexion: cannot read patterns file: %v\n", err)
		os.Exit(1)
	}
	return transpile.ExtractPatterns(string(data))
}

func levelName(l promote.Level) string {
	switch l {
	case promote.LevelHard:
		return "Hard"
	case promote.LevelConstraint:
		return "Constraint"
	case promote.LevelHint:
		return "Hint"
	case promote.LevelPruned:
		return "Pruned"
	default:
		return "Unknown"
	}
}

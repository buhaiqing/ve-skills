package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/check/frontmatter"
)

// checkReport is the machine-readable shape emitted with --json.
type checkReport struct {
	Skill  string   `json:"skill"`
	OK     bool     `json:"ok"`
	Errors []string `json:"errors"`
}

type checkSummary struct {
	Passing int            `json:"passing"`
	Total   int            `json:"total"`
	Reports []checkReport `json:"reports"`
}

func runCheck(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "vet check: missing subcommand (frontmatter|aiops|assessment|gcl|links|eval)")
		os.Exit(2)
	}
	name := args[0]
	rest := args[1:]

	fs := flag.NewFlagSet("check "+name, flag.ExitOnError)
	root := fs.String("root", repoRoot(), "repo root to scan")
	jsonOut := fs.Bool("json", false, "emit machine-readable JSON")
	fs.Parse(rest)

	var summary checkSummary
	switch name {
	case "frontmatter":
		summary = frontmatterCheck(*root)
	default:
		fmt.Fprintf(os.Stderr, "vet check %s: not implemented yet\n", name)
		os.Exit(3)
	}

	if *jsonOut {
		b, _ := json.MarshalIndent(summary, "", "  ")
		fmt.Println(string(b))
	} else {
		passing := summary.Passing
		total := summary.Total
		if total == 0 {
			fmt.Println("0 skills checked")
		} else if passing == total {
			fmt.Printf("OK: %d/%d skills passed\n", passing, total)
		} else {
			fmt.Printf("FAIL: %d/%d skills passed\n", passing, total)
			for _, r := range summary.Reports {
				if !r.OK {
					fmt.Printf("  FAIL %s:\n", r.Skill)
					for _, e := range r.Errors {
						fmt.Printf("    - %s\n", e)
					}
				}
			}
		}
	}
	if summary.Passing != summary.Total {
		os.Exit(1)
	}
}

func frontmatterCheck(root string) checkSummary {
	results, skills := frontmatter.CheckDir(root)
	var reports []checkReport
	for _, sk := range skills {
		errs, hasErrors := results[sk]
		if errs == nil {
			errs = []string{} // emit [] not null in JSON for clean skills
		}
		reports = append(reports, checkReport{
			Skill:  filepath.Base(filepath.Dir(sk)),
			OK:     !hasErrors,
			Errors: errs,
		})
	}
	sort.Slice(reports, func(i, j int) bool { return reports[i].Skill < reports[j].Skill })
	passing := 0
	for _, r := range reports {
		if r.OK {
			passing++
		}
	}
	return checkSummary{Passing: passing, Total: len(reports), Reports: reports}
}

// repoRoot returns the repository root: two levels up from cmd/vet.
func repoRoot() string {
	exe, err := os.Executable()
	if err == nil {
		// cmd/vet/vet -> ../../..
		if p := up(up(up(exe))); dirExists(p) {
			return p
		}
	}
	return "."
}

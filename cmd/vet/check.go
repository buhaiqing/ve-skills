package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/check/aiops"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/check/assessment"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/check/eval"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/check/frontmatter"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/check/gcl"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/check/links"
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

	switch name {
	case "frontmatter":
		frontmatterCheck(*root, *jsonOut)
	case "aiops":
		aiopsCheck(*root, *jsonOut)
	case "assessment":
		assessmentCheck(*root, *jsonOut)
	case "gcl":
		gclRes, gclSkills := gcl.CheckDir(*root)
		perSkillCheck("gcl", gclRes, gclSkills, *jsonOut)
	case "links":
		linkRes, linkSkills := links.CheckDir(*root)
		perSkillCheck("links", linkRes, linkSkills, *jsonOut)
	case "eval":
		evalRes, evalSkills := eval.CheckDir(*root)
		perSkillCheck("eval", evalRes, evalSkills, *jsonOut)
	default:
		fmt.Fprintf(os.Stderr, "vet check %s: not implemented yet\n", name)
		os.Exit(3)
	}
}

func frontmatterCheck(root string, jsonOut bool) {
	results, skills := frontmatter.CheckDir(root)
	perSkillCheck("frontmatter", results, skills, jsonOut)
}

// perSkillCheck renders a results map (only failing skills present) + sorted
// skill list into either human or JSON form. Passing == skill not in map.
func perSkillCheck(name string, results map[string][]string, skills []string, jsonOut bool) {
	var reports []checkReport
	for _, sk := range skills {
		errs, hasErrors := results[sk]
		if errs == nil {
			errs = []string{}
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
	summary := checkSummary{Passing: passing, Total: len(reports), Reports: reports}
	emitCheck(name, summary, jsonOut)
}

func aiopsCheck(root string, jsonOut bool) {
	rep := aiops.CheckDir(root)
	if jsonOut {
		b, _ := json.MarshalIndent(rep, "", "  ")
		fmt.Println(string(b))
		return
	}
	fmt.Println("AIOps Coverage Report")
	fmt.Println("==================================================")
	fmt.Printf("Total skills: %d\n", rep.TotalSkills)
	fmt.Printf("advanced/aiops.md:  R+R tier %s\n", rep.RRCoverage)
	fmt.Printf("advanced/finops.md: R+R tier %s\n", rep.RRFinOpsCoverage)
	fmt.Printf("eval_queries.json:  %s\n", rep.EvalCoverage)
	if len(rep.SkillsMissingAIOps) > 0 {
		fmt.Println("\nMissing advanced/aiops.md:")
		for _, s := range rep.SkillsMissingAIOps {
			fmt.Printf("  - %s\n", s)
		}
	}
	if len(rep.SkillsMissingFinOps) > 0 {
		fmt.Println("\nMissing advanced/finops.md (R+R tier):")
		for _, s := range rep.SkillsMissingFinOps {
			fmt.Printf("  - %s\n", s)
		}
	}
	if len(rep.SkillsMissingEval) > 0 {
		fmt.Println("\nMissing eval_queries.json:")
		for _, s := range rep.SkillsMissingEval {
			fmt.Printf("  - %s\n", s)
		}
	}
	if len(rep.EvalQualityBad) > 0 {
		fmt.Println("\nLow-quality eval_queries.json (need trigger>=5, non_trigger>=2):")
		for _, r := range rep.EvalQualityBad {
			fmt.Printf("  - %s: trigger=%d, non_trigger=%d\n", r.Skill, r.EvalTrigger, r.EvalNonTrigger)
		}
	}
	if !rep.OK {
		os.Exit(1)
	}
}

func assessmentCheck(root string, jsonOut bool) {
	errs, files, examples := assessment.CheckDir(root)
	if jsonOut {
		b, _ := json.MarshalIndent(map[string]any{
			"ok":               len(errs) == 0,
			"files_checked":    files,
			"examples_checked": examples,
			"errors":           errs,
		}, "", "  ")
		fmt.Println(string(b))
		return
	}
	if len(errs) > 0 {
		fmt.Printf("FAIL: %d error(s) in %d files (%d examples)\n\n", len(errs), files, examples)
		for _, e := range errs {
			fmt.Printf("  - %s\n", e)
		}
		os.Exit(1)
	}
	fmt.Printf("OK: %d files, %d example JSON blocks validated\n", files, examples)
}

func emitCheck(name string, summary checkSummary, jsonOut bool) {
	if jsonOut {
		b, _ := json.MarshalIndent(summary, "", "  ")
		fmt.Println(string(b))
		return
	}
	if summary.Total == 0 {
		fmt.Println("0 skills checked")
		return
	}
	if summary.Passing == summary.Total {
		fmt.Printf("OK: %d/%d skills passed (%s)\n", summary.Passing, summary.Total, name)
		return
	}
	fmt.Printf("FAIL: %d/%d skills passed (%s)\n", summary.Passing, summary.Total, name)
	for _, r := range summary.Reports {
		if !r.OK {
			fmt.Printf("  FAIL %s:\n", r.Skill)
			for _, e := range r.Errors {
				fmt.Printf("    - %s\n", e)
			}
		}
	}
	os.Exit(1)
}

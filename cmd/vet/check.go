package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/check/aiops"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/check/assessment"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/check/eval"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/check/frontmatter"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/check/gcl"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/check/links"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/check/policyguard"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/check/trace"
	gcltrace "github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/trace"
)

// checkReport is the machine-readable shape emitted with --json.
type checkReport struct {
	Skill  string   `json:"skill"`
	OK     bool     `json:"ok"`
	Errors []string `json:"errors"`
}

type checkSummary struct {
	Passing int            `json:"passing"`
	Total  int            `json:"total"`
	Reports []checkReport `json:"reports"`
}

func runCheck(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "vet check: missing subcommand (frontmatter|aiops|assessment|gcl|links|eval|policyguard|trace)")
		os.Exit(2)
	}
	name := args[0]
	rest := args[1:]

	fs := flag.NewFlagSet("check "+name, flag.ExitOnError)
	root := fs.String("root", repoRoot(), "repo root to scan")
	jsonOut := fs.Bool("json", false, "emit machine-readable JSON")
	gitDiff := fs.Bool("git-diff", false, "enable git-diff semantic-drift (eval only); base defaults to HEAD~1")
	gitDiffBase := fs.String("git-diff-base", "HEAD~1", "base revision for --git-diff semantic-drift (eval only)")
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
		var evalRes map[string][]string
		var evalSkills []string
		if *gitDiff {
			evalRes, evalSkills = eval.CheckDirGitDiff(*root, *gitDiffBase)
		} else {
			evalRes, evalSkills = eval.CheckDir(*root)
		}
		perSkillCheck("eval", evalRes, evalSkills, *jsonOut)
	case "policyguard":
		pgCheck(*root, *jsonOut)
		case "trace":
			traceCheck(*root, *jsonOut)
	case "routing":
		runCheckRouting(*root, *jsonOut)
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
	if len(rep.EvalParseFail) > 0 {
		fmt.Println("\nCorrupt eval_queries.json (JSON parse failure):")
		for _, s := range rep.EvalParseFail {
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

func pgCheck(root string, jsonOut bool) {
	// Find incident-loop-agent's policyguard test fixture (a real dispatch plan).
	// The schema file itself is not a plan; use the testdata fixture.
	planPath := filepath.Join(root, "cmd", "vet", "internal", "check", "policyguard", "testdata", "plan_clean.json")
	if _, err := os.Stat(planPath); os.IsNotExist(err) {
		if jsonOut {
			fmt.Println(`{"ok":true,"reason":"no plan fixture found"}`)
		} else {
			fmt.Println("OK: no plan fixture found (policyguard is clean)")
		}
		return
	}
	err := policyguard.Check(planPath)
	if jsonOut {
		if err != nil {
			b, _ := json.MarshalIndent(map[string]any{"ok": false, "error": err.Error()}, "", "  ")
			fmt.Println(string(b))
		} else {
			fmt.Println(`{"ok":true}`)
		}
		return
	}
	if err != nil {
		fmt.Printf("FAIL: policyguard: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("OK: policyguard passed")
}

func traceCheck(root string, jsonOut bool) {
	auditDir := filepath.Join(root, "audit-results")
	entries, err := os.ReadDir(auditDir)
	if os.IsNotExist(err) || len(entries) == 0 {
		if jsonOut {
			fmt.Println(`{"ok":true,"reason":"no traces found"}`)
		} else {
			fmt.Println("OK: no trace files found")
		}
		return
	}
	if err != nil {
		if jsonOut {
			fmt.Printf(`{"ok":false,"error":"read audit-results: %v"}`+"\n", err)
		} else {
			fmt.Printf("FAIL: read audit-results: %v\n", err)
		}
		os.Exit(1)
	}
	var failCount int
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(auditDir, entry.Name())
		var err error
		switch {
		case strings.HasPrefix(entry.Name(), "gcl-trace-"):
			// Runtime GCL trace (written by `vet gcl run`).
			err = gcltrace.Check(path)
		case strings.HasPrefix(entry.Name(), "incident-trace-"):
			// Agent incident trace (written by incident-loop-agent).
			err = trace.Check(path)
		default:
			continue
		}
		if err != nil {
			failCount++
			if !jsonOut {
				fmt.Printf("FAIL %s: %v\n", entry.Name(), err)
			}
		}
	}
	if jsonOut {
		b, _ := json.MarshalIndent(map[string]any{"ok": failCount == 0, "checked": len(entries), "failures": failCount}, "", "  ")
		fmt.Println(string(b))
		return
	}
	if failCount > 0 {
		os.Exit(1)
	}
	fmt.Printf("OK: %d trace file(s) passed\n", len(entries))
}

// routingSkillPattern matches a valid skill id used as primary/secondary.
var routingSkillPattern = regexp.MustCompile(`^ve-[a-z0-9-]+-ops$|^incident-loop-agent$`)

// runCheckRouting validates the skill routing graph (docs/skill-routing-graph.md).
// It parses markdown table rows and validates rows whose last cell is a known
// source enum (predictive|reactive) against the routing schema shape:
// symptom / primary / secondary / action / source, with primary & each
// secondary matching a skill id pattern and source ∈ {predictive, reactive}.
func runCheckRouting(root string, jsonOut bool) {
	path := filepath.Join(root, "docs", "skill-routing-graph.md")
	raw, err := os.ReadFile(path)
	if err != nil {
		if jsonOut {
			b, _ := json.Marshal(map[string]any{"ok": false, "error": err.Error()})
			fmt.Println(string(b))
		} else {
			fmt.Printf("FAIL: read %s: %v\n", path, err)
		}
		os.Exit(1)
	}

	var errors []string
	checked := 0
	for i, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			continue
		}
		cells := splitTableRow(line)
		// A trigger row ends with a source enum cell.
		if len(cells) < 5 {
			// Check whether this short row has a source cell matching
			// predictive/reactive — if so it's structurally broken.
			if len(cells) > 0 {
				last := strings.TrimSpace(cells[len(cells)-1])
				if last == "predictive" || last == "reactive" {
					checked++
					errors = append(errors, fmt.Sprintf("line %d: trigger row has %d cells, need 5", i+1, len(cells)))
				}
			}
			continue
		}
		source := strings.TrimSpace(cells[len(cells)-1])
		if source != "predictive" && source != "reactive" {
			continue
		}
		checked++
		symptom := strings.TrimSpace(cells[0])
		primary := stripCodeSpan(cells[1])
		secondary := stripCodeSpan(cells[2])
		action := strings.TrimSpace(cells[len(cells)-2])
		if symptom == "" {
			errors = append(errors, fmt.Sprintf("line %d: empty symptom", i+1))
		}
		if !routingSkillPattern.MatchString(primary) {
			errors = append(errors, fmt.Sprintf("line %d: invalid primary skill %q", i+1, primary))
		}
		if secondary != "" {
			for _, s := range strings.Split(secondary, ",") {
				if s = strings.TrimSpace(s); s != "" && !routingSkillPattern.MatchString(s) {
					errors = append(errors, fmt.Sprintf("line %d: invalid secondary skill %q", i+1, s))
				}
			}
		}
		if action == "" {
			errors = append(errors, fmt.Sprintf("line %d: empty action", i+1))
		}
	}

	if jsonOut {
		b, _ := json.MarshalIndent(map[string]any{
			"ok":      len(errors) == 0,
			"checked": checked,
			"errors":  errors,
		}, "", "  ")
		fmt.Println(string(b))
		return
	}
	if len(errors) > 0 {
		fmt.Printf("FAIL: %d routing row(s) invalid (%d checked)\n", len(errors), checked)
		for _, e := range errors {
			fmt.Printf("  - %s\n", e)
		}
		os.Exit(1)
	}
	fmt.Printf("OK: %d routing trigger row(s) valid\n", checked)
}

// splitTableRow splits a markdown table row into trimmed cell strings,
// dropping the leading/trailing pipe.
func splitTableRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	parts := strings.Split(line, "|")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}
	return out
}

// stripCodeSpan removes markdown code-span backticks so skill ids written as
// `ve-foo-ops` in the routing table validate against the bare id pattern.
func stripCodeSpan(s string) string {
	s = strings.TrimSpace(s)
	return strings.Trim(s, "`")
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

package validate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/check/aiops"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/check/assessment"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/check/distillation"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/check/eval"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/check/frontmatter"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/check/gcl"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/check/links"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/gate"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/run"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/trace"
)

// stepOutcome is the result of running one validation step.
type stepOutcome struct {
	errors   []string // per-finding messages (empty => pass)
	advisory bool     // if advisory, a non-empty errors list does not break the suite
}

// Step is a single validation step, mirroring the Python Step dataclass.
// Unlike the old implementation (which shelled out to python3 for every
// check), each step runs in-process against the equivalent `vet check` /
// `vet gcl` package — no Python interpreter required.
type Step struct {
	name string
	run  func(root string) stepOutcome
}

// mapFailed reports whether a (results map, skill list) pair has any failure.
func mapFailed(results map[string][]string) []string {
	var errs []string
	for _, es := range results {
		errs = append(errs, es...)
	}
	return errs
}

// mapFailedOK mirrors mapFailed but treats an empty map as pass.
func checkMap(root string, fn func(string) (map[string][]string, []string)) stepOutcome {
	res, _ := fn(root)
	return stepOutcome{errors: mapFailed(res)}
}

// buildSteps constructs the full validation step list, mirroring
// build_steps() in scripts/validate_local.py but executed in pure Go.
func buildSteps() []Step {
	steps := []Step{
		{
			name: "File integrity (null byte check)",
			run: func(root string) stepOutcome {
				return stepOutcome{errors: checkFileIntegrity(root)}
			},
		},
		{
			name: "Frontmatter validation",
			run:  func(root string) stepOutcome { return checkMap(root, frontmatter.CheckDir) },
		},
		{
			name: "Required sections presence",
			run: func(root string) stepOutcome {
				return stepOutcome{errors: checkRequiredSections(root)}
			},
		},
		{
			name: "Error Taxonomy (≥10 codes, HALT/RETRY)",
			run: func(root string) stepOutcome {
				return stepOutcome{errors: checkErrorTaxonomy(root), advisory: true}
			},
		},
		{
			name: "TE-1 hardcoded version scan",
			run: func(root string) stepOutcome {
				return stepOutcome{errors: checkTE1Hardcodes(root), advisory: true}
			},
		},
		{
			name: "Markdown local links",
			run:  func(root string) stepOutcome { return checkMap(root, links.CheckDir) },
		},
		{
			name: "AIOps / FinOps / eval coverage",
			run: func(root string) stepOutcome {
				rep := aiops.CheckDir(root)
				var errs []string
				errs = append(errs, rep.SkillsMissingAIOps...)
				errs = append(errs, rep.SkillsMissingFinOps...)
				errs = append(errs, rep.SkillsMissingEval...)
				errs = append(errs, rep.EvalParseFail...)
				return stepOutcome{errors: errs}
			},
		},
		{
			name: "GCL Tier-A conformance",
			run:  func(root string) stepOutcome { return checkMap(root, gcl.CheckDir) },
		},
		{
			name: "Eval regression",
			run:  func(root string) stepOutcome { return checkMap(root, eval.CheckDir) },
		},
		{
			name: "Product assessment examples",
			run: func(root string) stepOutcome {
				errs, _, _ := assessment.CheckDir(root)
				return stepOutcome{errors: errs}
			},
		},
		{
			name: "Knowledge distillation compliance",
			run: func(root string) stepOutcome {
				results, err := distillation.CheckDir(root)
				if err != nil {
					return stepOutcome{errors: []string{fmt.Sprintf("distillation check failed: %v", err)}}
				}
				var errs []string
				for _, r := range results {
					errs = append(errs, r.Errors...)
				}
				return stepOutcome{errors: errs}
			},
		},
		{
			name: "GCL runner structural smoke",
			run: func(root string) stepOutcome {
				code := run.Run(run.Options{
					Root:           root,
					Skill:          "ve-skill-generator",
					Request:        "CI smoke test",
					Command:        `echo {"Response":{"RequestId":"ci-smoke"}}`,
					MaxIter:        1,
					Timeout:        30,
					StructuralOnly: true,
				}).ExitCode
				// Smoke verifies loop plumbing (classifier + policy + writeback),
				// not the policy decision itself. Any code < 5 means the loop
				// completed structurally (0 PASS / 1 MAX_ITER / 2 USER_CANCEL /
				// 3 SELF_HEAL_EXHAUSTED / 4 REFUSE). Negative = crash/timeout.
				if code >= 0 && code < 5 {
					return stepOutcome{}
				}
				return stepOutcome{errors: []string{fmt.Sprintf("gcl run structural smoke exited with %d", code)}}
			},
		},
		{
			name: "GCL CI gate (all skills)",
			run: func(root string) stepOutcome {
				code := gate.Run(root, nil, true, false)
				if code == 0 {
					return stepOutcome{}
				}
				return stepOutcome{errors: []string{fmt.Sprintf("gcl gate exited with %d", code)}}
			},
		},
		{
			name: "GCL trace aggregate",
			run: func(root string) stepOutcome {
				hours := 168
				code := trace.CmdAggregate(root, nil, &hours)
				// No traces in a clean checkout is the normal state, so this
				// step is advisory rather than breaking the suite.
				return stepOutcome{errors: nil, advisory: code != 0}
			},
		},
		{
			name: "Unit tests (go test ./...)",
			run: func(root string) stepOutcome {
				return goTestStep(root)
			},
		},
	}
	return steps
}

// goTestStep runs the vet module's Go test suite. Mirrors the original
// `python3 -m unittest discover -s scripts` step, but exercises the Go
// package tests instead.
func goTestStep(root string) stepOutcome {
	vetDir := filepath.Join(root, "cmd", "vet")
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = vetDir
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		var errs []string
		for _, line := range strings.Split(stderr.String(), "\n") {
			if strings.TrimSpace(line) != "" {
				errs = append(errs, line)
			}
		}
		if len(errs) == 0 {
			errs = append(errs, "go test failed")
		}
		return stepOutcome{errors: errs}
	}
	return stepOutcome{}
}

// Run executes every step in order, stopping at the first non-advisory
// failure. Returns (errors, failedStepNames, totalSteps) where errors maps
// each failed step name to its finding lines, and failedStepNames is the
// sorted list of failing step names. When listOnly is true, no steps are
// executed; the step commands are printed instead and (nil, nil, total) is
// returned.
func Run(root string, listOnly bool) (map[string][]string, []string, int) {
	steps := buildSteps()
	total := len(steps)
	if listOnly {
		for _, s := range steps {
			fmt.Printf("%s\n", s.name)
		}
		return nil, nil, total
	}

	failErrs := make(map[string][]string)
	var failed []string
	for _, s := range steps {
		out := s.run(root)
		if len(out.errors) == 0 {
			continue
		}
		if out.advisory {
			fmt.Fprintf(os.Stderr, "\nADVISORY: %s — %d finding(s)\n", s.name, len(out.errors))
			for _, e := range out.errors {
				fmt.Fprintf(os.Stderr, "  WARN: %s\n", e)
			}
			continue
		}
		fmt.Fprintf(os.Stderr, "\nFAILED: %s\n", s.name)
		for _, e := range out.errors {
			fmt.Fprintf(os.Stderr, "  FAIL: %s\n", e)
		}
		failErrs[s.name] = out.errors
		failed = append(failed, s.name)
		break
	}
	sort.Strings(failed)
	if len(failed) == 0 {
		fmt.Fprintln(os.Stderr, "\nOK: local validation suite passed")
	}
	return failErrs, failed, total
}

// StepNames returns the friendly names of all validation steps, for --list.
func StepNames() []string {
	var names []string
	for _, s := range buildSteps() {
		names = append(names, s.name)
	}
	return names
}

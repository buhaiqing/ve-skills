// Package run implements `vet gcl run` — the GCL Orchestrator loop.
//
// Faithful Go port of scripts/gcl_runner.py: executes the generator command,
// feeds (sanitized) output to a Critic (structural-only, --critic-json,
// --critic-stdin, or an isolated --critic-command process), decides
// PASS / RETRY / SAFETY_FAIL, persists the trace, and performs Reflexion
// failure-pattern write-back. Credential masking is applied throughout.
package run

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/critic"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/secret"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/trace"
)

// SkillMaxIter mirrors gcl_runner.SKILL_MAX_ITER.
var SkillMaxIter = map[string]int{
	"ve-ecs-ops": 2, "ve-redis-ops": 2, "ve-rds-mysql-ops": 2, "ve-rds-ops": 2,
	"ve-rds-pg-ops": 2, "ve-polar-mysql-ops": 2, "ve-mongodb-ops": 2, "ve-elasticsearch-ops": 2,
	"ve-tos-ops": 2, "ve-iam-ops": 2, "ve-kms-ops": 2, "ve-eip-ops": 2, "ve-security-group-ops": 2,
	"ve-vpc-ops": 3, "ve-nat-ops": 3, "ve-vpn-ops": 3, "ve-clb-ops": 3, "ve-alb-ops": 3,
	"ve-vke-ops": 3, "ve-nas-ops": 3, "ve-cms-ops": 3, "ve-fg-ops": 3, "ve-ark-ops": 3,
	"ve-cdn-ops": 5, "ve-dns-ops": 5, "ve-kafka-ops": 5, "ve-sls-ops": 5, "ve-billing-ops": 5,
	"ve-skill-generator": 3,
}

// failureSignatures mirrors gcl_runner._FAILURE_SIGNATURES.
var failureSignatures = []struct {
	category string
	re       *regexp.Regexp
}{
	{"cli_parameter", regexp.MustCompile(`InvalidParameter|MissingParameter|AuthFailure|UnauthorizedOperation`)},
	{"runtime", regexp.MustCompile(`TIMEOUT|RequestLimitExceeded|InternalError|ConnectionError`)},
	{"cross_skill", regexp.MustCompile(`delegate-to|not found in target skill|cross-skill`)},
	{"token_efficiency", regexp.MustCompile(`token budget|exceeds.*token|too long|truncated`)},
	{"skill_generation", regexp.MustCompile(`frontmatter missing|missing rubric|broken link`)},
}

var (
	destructiveVerbs = regexp.MustCompile(`\b(Delete|Remove|Terminate|Destroy|Stop|Shutdown|PowerOff|Release|Revoke|Disable|Deactivate|Flush|Purge|Drop|Truncate|Detach|Disassociate|Revoke\w*Access|Revoke\w*Permission)\b`)
	mutatingVerbs    = regexp.MustCompile(`\b(Create|Add|Allocate|Attach|Assign|Authorize|Enable|Activate|Modify|Update|Set|Change|Resize|Rebuild|Reboot|Restart)\b`)
	negationPattern  = regexp.MustCompile(`\b(Enable|Activate|Allow|Grant|Create)\w*(Protection|Policy|Rule|Firewall)\b`)
)

// Options is the resolved configuration for a run.
type Options struct {
	Root              string
	Skill             string
	Request           string
	Command           string
	MaxIter           int
	Timeout           int
	StructuralOnly    bool
	CriticJSON        string
	CriticStdin       bool
	CriticCommand     string
}

// deriveOperationIntent mirrors gcl_runner.derive_operation_intent.
func deriveOperationIntent(skill, command string) map[string]any {
	if command == "" {
		return map[string]any{"operation": "unknown", "resource_scope": []string{}, "expected_state": "unknown", "safety_class": "read_only"}
	}
	resource := strings.TrimSuffix(strings.TrimPrefix(skill, "ve-"), "-ops")
	cmdStripped := regexp.MustCompile(`#.*$`).ReplaceAllString(command, "")
	if negationPattern.MatchString(cmdStripped) {
		return map[string]any{"operation": "enable_" + resource, "resource_scope": []string{}, "expected_state": "ACTIVE", "safety_class": "mutating"}
	}
	if destructiveVerbs.MatchString(cmdStripped) {
		return map[string]any{"operation": "destructive_" + resource, "resource_scope": []string{}, "expected_state": "DELETED", "safety_class": "destructive"}
	}
	if mutatingVerbs.MatchString(cmdStripped) {
		return map[string]any{"operation": "modify_" + resource, "resource_scope": []string{}, "expected_state": "MODIFIED", "safety_class": "mutating"}
	}
	return map[string]any{"operation": "describe", "resource_scope": []string{}, "expected_state": "UNCHANGED", "safety_class": "read_only"}
}

// runCommand mirrors gcl_runner.run_command with masking.
func runCommand(command string, timeout int, extraEnv map[string]string) trace.GeneratorResult {
	env := os.Environ()
	for k, v := range extraEnv {
		if v != "" {
			env = append(env, k+"="+v)
		}
	}
	cmd := exec.Command("sh", "-c", command)
	cmd.Env = env
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	exitCode := 0
	done := make(chan error, 1)
	go func() { done <- cmd.Run() }()
	select {
	case err := <-done:
		if err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				exitCode = ee.ExitCode()
			} else {
				exitCode = -1
			}
		}
	case <-time.After(time.Duration(timeout) * time.Second):
		_ = cmd.Process.Kill()
		return trace.GeneratorResult{
			Command:       secret.MaskSecrets(command),
			ExitCode:      -1,
			ResultExcerpt: "TIMEOUT after " + itoa(timeout) + "s",
			StdoutLen:     len(stdout.String()),
			StderrLen:     len(stderr.String()),
		}
	}
	combined := secret.MaskSecrets(stdout.String() + stderr.String())
	excerpt := combined
	if len(excerpt) > 2000 {
		excerpt = excerpt[:2000] + "..."
	}
	return trace.GeneratorResult{
		Command:       secret.MaskSecrets(command),
		ExitCode:      exitCode,
		ResultExcerpt: excerpt,
		StdoutLen:     len(stdout.String()),
		StderrLen:     len(stderr.String()),
	}
}

// loadCritic mirrors gcl_runner.load_critic (--critic-json / --critic-stdin).
func loadCritic(path string, stdin bool) (*critic.CriticResult, error) {
	var raw []byte
	var err error
	if path != "" {
		raw, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
	} else if stdin && !isTerminal() {
		raw, err = readAllStdin()
		if err != nil {
			return nil, err
		}
	} else {
		return nil, nil
	}
	var c critic.CriticResult
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// runIsolatedCritic mirrors gcl_runner.run_isolated_critic — separate process,
// never sees the raw request.
func runIsolatedCritic(opts Options, operationIntent map[string]any, gen trace.GeneratorResult, iterations []trace.Iteration) (*critic.CriticResult, error) {
	rubricPath := "AGENTS.md"
	rp := joinPath(opts.Root, opts.Skill, "references", "rubric.md")
	if fileExists(rp) {
		rel, _ := relPath(opts.Root, rp)
		rubricPath = rel
	}
	input := map[string]any{
		"skill":           opts.Skill,
		"operation_intent": operationIntent,
		"generator_output": map[string]any{
			"command":        gen.Command,
			"exit_code":      gen.ExitCode,
			"result_excerpt": gen.ResultExcerpt,
		},
		"trace":       map[string]any{"iterations": iterations},
		"rubric_path": rubricPath,
	}
	inBytes, _ := json.Marshal(input)
	cmd := exec.Command("sh", "-c", opts.CriticCommand)
	cmd.Stdin = strings.NewReader(string(inBytes))
	var out, errBuf strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			fmt.Fprintf(os.Stderr, "ERROR: Critic command failed (%d): %s\n", ee.ExitCode(), firstLine(errBuf.String()))
		} else {
			fmt.Fprintf(os.Stderr, "ERROR: Critic command error: %v\n", err)
		}
		return nil, err
	}
	var c critic.CriticResult
	if err := json.Unmarshal([]byte(out.String()), &c); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Critic output is not valid JSON: %v\n", err)
		return nil, err
	}
	return &c, nil
}

// extractFailurePattern mirrors gcl_runner.extract_failure_pattern.
func extractFailurePattern(skill, command string, gen trace.GeneratorResult, c *critic.CriticResult) *trace.FailurePattern {
	corpusParts := []string{command, gen.ResultExcerpt}
	if c != nil {
		corpusParts = append(corpusParts, c.Suggestions...)
	}
	corpus := strings.Join(corpusParts, "\n")
	for _, fs := range failureSignatures {
		m := fs.re.FindString(corpus)
		if m == "" {
			continue
		}
		fix := "Investigate failure pattern and add fix"
		if c != nil && len(c.Suggestions) > 0 {
			fix = c.Suggestions[0]
		}
		if len(fix) > 200 {
			fix = fix[:200]
		}
		cmd := command
		if len(cmd) > 200 {
			cmd = cmd[:200]
		}
		return &trace.FailurePattern{
			Category: fs.category,
			Skill:    skill,
			Command:  cmd,
			Error:    m,
			Fix:      fix,
			Count:    1,
			Reusable: fs.category == "cli_parameter" || fs.category == "runtime",
		}
	}
	return nil
}

// loadKnownFailurePatterns mirrors gcl_runner.load_known_failure_patterns.
func loadKnownFailurePatterns(root, skill string, limit int) string {
	fp := joinPath(root, "docs", "failure-patterns.md")
	if !fileExists(fp) {
		return ""
	}
	b, _ := os.ReadFile(fp)
	var out []string
	for _, line := range strings.Split(string(b), "\n") {
		s := strings.TrimSpace(line)
		if !strings.HasPrefix(s, "|") || !strings.Contains(s, skill) {
			continue
		}
		if strings.Contains(s, "---") || (strings.Contains(strings.ToLower(s), "skill") && strings.Contains(strings.ToLower(s), "command")) {
			continue
		}
		out = append(out, s)
		if len(out) >= limit {
			break
		}
	}
	return strings.Join(out, "\n")
}

// writebackFailurePattern mirrors gcl_runner._writeback_failure_pattern.
func writebackFailurePattern(root, skill string, fp *trace.FailurePattern) {
	if fp == nil {
		return
	}
	summary := trace.Summary{
		FailurePatterns: []map[string]any{
			{
				"skill":    skill,
				"pattern":  fp.Error,
				"category": fp.Category,
				"source":   orString(fp.Command, "gcl-runner"),
			},
		},
	}
	if _, err := trace.UpdateFailurePatternsFile(root, &summary); err != nil {
		fmt.Fprintf(os.Stderr, "WARN: Reflexion write-back skipped: %v\n", err)
	}
}

// Run executes the GCL loop. Returns the process exit code.
func Run(opts Options) int {
	if opts.MaxIter <= 0 {
		opts.MaxIter = SkillMaxIter[opts.Skill]
		if opts.MaxIter == 0 {
			opts.MaxIter = 3
		}
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 120
	}

	operationIntent := deriveOperationIntent(opts.Skill, opts.Command)
	maskedFields := secret.DetectCredentialFields(opts.Command)
	knownPatterns := loadKnownFailurePatterns(opts.Root, opts.Skill, 10)

	tr := &trace.Trace{
		TraceSchemaVersion: "v1",
		Skill:              opts.Skill,
		Request:            opts.Request,
		RubricVersion:      "v1",
		OperationIntent:    operationIntent,
		MaskedFields:       maskedFields,
		RedactionPass:      true,
		Iterations:         []trace.Iteration{},
	}

	criticFeedback := ""
	for iter := 1; iter <= opts.MaxIter; iter++ {
		gen := runCommand(opts.Command, opts.Timeout, map[string]string{
			"GCL_CRITIC_FEEDBACK":      criticFeedback,
			"GCL_KNOWN_FAILURE_PATTERNS": knownPatterns,
		})
		gen.Args = map[string]any{"iter": iter, "critic_feedback": orEmpty(criticFeedback)}

		var c *critic.CriticResult
		if opts.StructuralOnly {
			g := critic.GeneratorResult{Command: gen.Command, ExitCode: gen.ExitCode, ResultExcerpt: gen.ResultExcerpt}
			res := critic.StructuralCritic(g)
			c = &res
		} else {
			if secret.HasCredentialLeak(gen.ResultExcerpt) {
				tr.Iterations = append(tr.Iterations, trace.Iteration{
					Iter:      iter,
					Generator: gen,
					Critic: trace.CriticRecord{
						Scores:      map[string]float64{"correctness": 0, "safety": 0, "idempotency": 0.5, "traceability": 0.5, "spec_compliance": 0.5},
						Suggestions: []string{"Credential leak in generator output — mask secrets and re-run"},
						Blocking:    true,
					},
					Decision: "SAFETY_FAIL",
				})
				fp := extractFailurePattern(opts.Skill, opts.Command, gen, nil)
				tr.Final = trace.Final{Status: "SAFETY_FAIL", Iter: iter, Output: nil, FailurePattern: fp}
				path, _ := trace.PersistTrace(opts.Root, "", tr)
				writebackFailurePattern(opts.Root, opts.Skill, fp)
				fmt.Fprintf(os.Stderr, "SAFETY_FAIL — credential leak detected — trace: %s\n", path)
				return 3
			}
			loaded, err := loadCritic(opts.CriticJSON, opts.CriticStdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: invalid critic JSON: %v\n", err)
				return 2
			}
			if loaded == nil && opts.CriticCommand != "" {
				loaded2, err := runIsolatedCritic(opts, operationIntent, gen, tr.Iterations)
				if err != nil {
					return 2
				}
				loaded = loaded2
			}
			if loaded == nil {
				fmt.Fprintln(os.Stderr, "ERROR: No Critic payload. Pass --critic-json, pipe JSON to stdin, --critic-command <cmd>, or use --structural-critic-only for rule-based audit.")
				return 2
			}
			if errs := critic.ValidatePayload(*loaded); len(errs) > 0 {
				fmt.Fprintf(os.Stderr, "ERROR: Invalid critic JSON: %s\n", strings.Join(errs, "; "))
				return 2
			}
			c = loaded
		}

		decision := critic.Decide(c.Scores)
		tr.Iterations = append(tr.Iterations, trace.Iteration{
			Iter:      iter,
			Generator: gen,
			Critic:    trace.CriticRecord{Scores: c.Scores, Suggestions: c.Suggestions, Blocking: c.Blocking},
			Decision: decision,
		})

		if decision == "SAFETY_FAIL" {
			fp := extractFailurePattern(opts.Skill, opts.Command, gen, c)
			tr.Final = trace.Final{Status: "SAFETY_FAIL", Iter: iter, Output: nil, FailurePattern: fp}
			path, _ := trace.PersistTrace(opts.Root, "", tr)
			writebackFailurePattern(opts.Root, opts.Skill, fp)
			fmt.Fprintf(os.Stderr, "SAFETY_FAIL — trace: %s\n", path)
			return 3
		}
		if decision == "PASS" {
			out := gen.ResultExcerpt
			tr.Final = trace.Final{Status: "PASS", Iter: iter, Output: &out}
			path, _ := trace.PersistTrace(opts.Root, "", tr)
			fmt.Printf("PASS (iter %d) — trace: %s\n", iter, path)
			return 0
		}
		criticFeedback = strings.Join(firstN(c.Suggestions, 3), "; ")
	}

	lastIter := tr.Iterations[len(tr.Iterations)-1]
	var unresolved []string
	for dim, th := range critic.RubricThresholds {
		if lastIter.Critic.Scores[dim] < th {
			unresolved = append(unresolved, dim)
		}
	}
	fp := extractFailurePattern(opts.Skill, opts.Command, lastIter.Generator, &critic.CriticResult{Scores: lastIter.Critic.Scores, Suggestions: lastIter.Critic.Suggestions})
	out := lastIter.Generator.ResultExcerpt
	tr.Final = trace.Final{Status: "MAX_ITER", Iter: opts.MaxIter, Output: &out, Unresolved: unresolved, FailurePattern: fp}
	path, _ := trace.PersistTrace(opts.Root, "", tr)
	writebackFailurePattern(opts.Root, opts.Skill, fp)
	fmt.Fprintf(os.Stderr, "MAX_ITER — trace: %s\n", path)
	return 1
}

// Package run implements `vet gcl run` — the GCL Orchestrator loop.
//
// Faithful Go port of scripts/gcl_runner.py: executes the generator command,
// feeds (sanitized) output to a Critic (structural-only, --critic-json,
// --critic-stdin, or an isolated --critic-command process), decides
// PASS / RETRY / SAFETY_FAIL, persists the trace, and performs Reflexion
// failure-pattern write-back. Credential masking is applied throughout.
package run

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/critic"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/heal"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/secret"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/trace"
	vlog "github.com/buhaiqing/ve-skills/cmd/vet/internal/log"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/memory"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/costgate"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/reflexion/transpile"
	"gopkg.in/yaml.v3"
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
	// Verbs may be written as standalone tokens (Delete --x) or CamelCase
	// action names (DeleteInstances), hence the optional ([A-Z]\w*)? suffix.
	destructiveVerbs = regexp.MustCompile(`\b(Delete|Remove|Terminate|Destroy|Stop|Shutdown|PowerOff|Release|Revoke|Disable|Deactivate|Flush|Purge|Drop|Truncate|Detach|Disassociate)([A-Z]\w*)?\b`)
	mutatingVerbs    = regexp.MustCompile(`\b(Create|Add|Allocate|Attach|Assign|Authorize|Enable|Activate|Modify|Update|Set|Change|Resize|Rebuild|Reboot|Restart)([A-Z]\w*)?\b`)
	negationPattern  = regexp.MustCompile(`\b(Enable|Activate|Allow|Grant|Create)\w*(Protection|Policy|Rule|Firewall)\b`)
)

var allowedSkills = map[string]bool{
	"ve-cms-ops":       true,
	"ve-ecs-ops":       true,
	"ve-rds-mysql-ops": true,
	"ve-redis-ops":     true,
	"ve-vpc-ops":       true,
	"ve-iam-ops":       true,
	"ve-kms-ops":       true,
	"ve-billing-ops":   true,
}

// OpDecision is the policy decision for one operation.
type OpDecision int

const (
	OpRefuse OpDecision = iota
	OpAsk
	OpAuto
)

func (d OpDecision) String() string {
	switch d {
	case OpRefuse:
		return "REFUSE"
	case OpAsk:
		return "ASK"
	case OpAuto:
		return "AUTO"
	}
	return "UNKNOWN"
}

// scoreDecision applies the execution-risk policy to a single operation.
// Returns the policy verdict: AUTO, ASK, or REFUSE.
func scoreDecision(skill, safetyClass, blastRadius, confidence string, safety float64, metadataOK bool) OpDecision {
	if safety == 0 {
		return OpRefuse
	}
	if safetyClass == "destructive" {
		return OpAsk
	}
	if !metadataOK {
		return OpAsk
	}
	if !allowedSkills[skill] {
		return OpAsk
	}
	if safetyClass == "read_only" && confidence == "high" {
		return OpAuto
	}
	if safetyClass == "mutating" && blastRadius == "single" && confidence == "high" {
		return OpAuto
	}
	return OpAsk
}

// Options is the resolved configuration for a run.
type Options struct {
	Root           string
	Skill          string
	Request        string
	Command        string
	MaxIter        int
	Timeout        int
	StructuralOnly bool
	CriticJSON     string
	CriticStdin    bool
	CriticCommand  string
	// Confirmed lets an external caller vouch for ASK-class operations
	// (e.g. a human gate upstream). In the non-interactive `vet gcl run`
	// runtime ASK is otherwise treated as REFUSE (no human to ask).
	Confirmed bool
	// ConfirmedBy records the provenance of that vouch (ticket id, human
	// handle, or upstream loop id from the Step 5 {{user.confirm}} gate) so
	// the trace can answer "who authorized this ASK-class op".
	ConfirmedBy string
	// Heal selects the retry strategy for the generator command:
	//   "smart" (default) — error-classification-driven retry (L4/T09)
	//   "none"           — legacy fixed-count loop (no smart retry)
	Heal string
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

// policyInputs derives the scoreDecision arguments from the operation intent
// and (optional) Critic scores. The execution-risk gate runs BEFORE the
// generator command, so on the first iteration no Critic scores exist yet —
// we fail-safe to low confidence and a passable safety, which keeps
// non-read-only ops out of AUTO until evidence accrues.
func policyInputs(skill string, intent map[string]any, scores map[string]float64) (safetyClass, blastRadius, confidence string, safety float64, metadataOK bool) {
	safetyClass, _ = intent["safety_class"].(string)
	if safetyClass == "" {
		safetyClass = "read_only"
	}
	// A single `ve <svc> <Action>` maps to a single resource by default.
	blastRadius = "single"
	// Leaf skills now expose machine-readable safety_class/blast_radius; if
	// the intent resolved to a known class we treat metadata as present.
	metadataOK = safetyClass != "" && allowedSkills[skill]
	if scores == nil {
		// No Critic evidence yet. Read-only ops are inherently safe → treat as
		// high confidence so they AUTO-execute (L3 happy path). Mutating /
		// destructive ops stay low confidence → fall to ASK/REFUSE (conservative).
		if safetyClass == "read_only" {
			confidence = "high"
		} else {
			confidence = "low"
		}
		safety = 1.0
		return
	}
	if s, ok := scores["safety"]; ok {
		safety = s
	} else {
		safety = 1.0
	}
	// Map the lowest non-passing rubric dimension to a confidence bucket.
	confidence = "high"
	for _, dim := range []string{"correctness", "idempotency", "traceability", "spec_compliance"} {
		if v, ok := scores[dim]; ok && v < 1.0 {
			confidence = "medium"
			break
		}
	}
	if v, ok := scores["safety"]; ok && v < 1.0 {
		confidence = "low"
	}
	return
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
			StderrExcerpt: firstLine(secret.MaskSecrets(stderr.String())),
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
		StderrExcerpt: firstLine(secret.MaskSecrets(stderr.String())),
	}
}

// runGeneratorWithHeal runs the generator command, applying the L4 multi-path
// self-healing policy (heal.Run) when opts.Heal == "smart". Under "none" it
// falls back to a single call (legacy fixed-count loop — the outer MaxIter
// bounds critic-driven retries). The chosen error class and the self-healing
// result are returned so the caller can stamp them into the trace iteration
// for audit/telemetry (T10 + T11).
//
// metrics (may be nil) accumulates each healing attempt for the self-healing
// telemetry; logPath ("" = no logging) appends a framework §6.2 row per
// attempt so T11's `vet gcl heal-stats` can aggregate offline.
func runGeneratorWithHeal(opts Options, criticFeedback, knownPatterns string, metrics *heal.Metrics, logPath string, runID string) (trace.GeneratorResult, string, *trace.SelfHealingRecord) {
	env := map[string]string{
		"GCL_CRITIC_FEEDBACK":        criticFeedback,
		"GCL_KNOWN_FAILURE_PATTERNS": knownPatterns,
	}
	if opts.Heal != "smart" {
		return runCommand(opts.Command, opts.Timeout, env), "", nil
	}
	// First attempt: observe the outcome so we can classify the error and
	// select the best healing path (Classify → Select → Execute, per T10).
	first := runCommand(opts.Command, opts.Timeout, env)
	if first.ExitCode == 0 {
		// Succeeded on first try — no healing needed.
		return first, "", nil
	}
	var last trace.GeneratorResult
	op := func() error {
		last = runCommand(opts.Command, opts.Timeout, env)
		if last.ExitCode != 0 {
			return &generatorError{exit: last.ExitCode, excerpt: last.ResultExcerpt}
		}
		return nil
	}
	class := heal.Classify(first.ResultExcerpt)
	start := time.Now()
	res, _ := heal.Run(context.Background(), class, op, metrics)
	if metrics != nil && res.Fallback {
		metrics.FallbackUsed++
	}
	recordHealPath(metrics, logPath, res, start, runID)
	shr := &trace.SelfHealingRecord{
		Class:      res.Class,
		PathName:   res.Name,
		Cost:       res.Cost,
		Result:     res.Result,
		DurationMs: res.DurationMs,
	}
	return last, res.Class, shr
}

// recordHealPath folds one heal.Run PathResult into the telemetry sink
// (Metrics.PerPath) and the §6.2 log so T11's `vet gcl heal-stats` can
// re-aggregate offline.
func recordHealPath(metrics *heal.Metrics, logPath string, res heal.PathResult, start time.Time, runID string) {
	if metrics == nil && logPath == "" {
		return
	}
	durationMs := time.Since(start).Milliseconds()
	ev := heal.HealEvent{
		ISO:        time.Now().UTC().Format(time.RFC3339),
		EventType:  "multi-path",
		ErrorCode:  res.Class,
		Action:     res.Name,
		Result:     res.Result,
		DurationMs: durationMs,
	}
	if metrics != nil {
		metrics.Record(ev)
	}
	if logPath != "" {
		// Use structured log package for rotation-aware heal event logging
		if err := vlog.Append(logPath, runID, vlog.INFO, "gcl.heal", "multi-path",
			vlog.KV("event_type", ev.EventType),
			vlog.KV("error_code", ev.ErrorCode),
			vlog.KV("action", ev.Action),
			vlog.KV("result", ev.Result),
			vlog.KV("duration_ms", fmt.Sprintf("%d", ev.DurationMs)),
		); err != nil {
			fmt.Fprintf(os.Stderr, "[%s] [WARN] gcl.heal | heal log write skipped | %v\n", runID, err)
		}
	}
}

// generatorError is the synthetic failure returned by the SmartRetry op when
// the generator command exits non-zero or times out. Its Error() message is
// the ResultExcerpt verbatim, so the error-class signal is preserved for
// heal.Classify.
type generatorError struct {
	exit    int
	excerpt string
}

func (e *generatorError) Error() string {
	if e.excerpt != "" {
		return e.excerpt
	}
	return fmt.Sprintf("generator exit %d", e.exit)
}

// parseRequestID extracts the cloud API RequestId from `ve` CLI JSON output
// ({"Response":{"RequestId":"..."}}). Returns "" if absent or unparseable —
// the runtime can still emit a trace, it just won't be request_id-traceable.
func parseRequestID(output string) string {
	// Prefer structured parse: ve emits {"Response":{"RequestId":"..."}}.
	var doc struct {
		Response struct {
			RequestId string `json:"RequestId"`
		} `json:"Response"`
	}
	if err := json.Unmarshal([]byte(output), &doc); err == nil {
		if id := strings.TrimSpace(doc.Response.RequestId); id != "" {
			return id
		}
	}
	// Fallback: scan for the first "RequestId" string value (legacy/non-JSON output).
	idx := strings.Index(output, "\"RequestId\"")
	if idx < 0 {
		return ""
	}
	rest := output[idx+len("\"RequestId\""):]
	// Find the first quoted value after the key.
	start := strings.Index(rest, "\"")
	if start < 0 {
		return ""
	}
	end := strings.Index(rest[start+1:], "\"")
	if end < 0 {
		return ""
	}
	id := rest[start+1 : start+1+end]
	if strings.TrimSpace(id) == "" {
		return ""
	}
	return id
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
func runIsolatedCritic(opts Options, operationIntent map[string]any, gen trace.GeneratorResult, iterations []trace.Iteration, runID string) (*critic.CriticResult, error) {
	rubricPath := "AGENTS.md"
	rp := joinPath(opts.Root, opts.Skill, "references", "rubric.md")
	if fileExists(rp) {
		rel, _ := relPath(opts.Root, rp)
		rubricPath = rel
	}
	input := map[string]any{
		"skill":            opts.Skill,
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
			fmt.Fprintf(os.Stderr, "[%s] [ERROR] gcl.run | critic command failed | exit=%d stderr=%s\n",
				runID, ee.ExitCode(), firstLine(errBuf.String()))
		} else {
			fmt.Fprintf(os.Stderr, "[%s] [ERROR] gcl.run | critic command error | %v\n", runID, err)
		}
		return nil, err
	}
	var c critic.CriticResult
	if err := json.Unmarshal([]byte(out.String()), &c); err != nil {
		fmt.Fprintf(os.Stderr, "[%s] [ERROR] gcl.run | critic JSON parse failed | %v\n", runID, err)
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

// loadKnownFailurePatterns loads failure patterns from the structured JSON store
// and formats them as HINT lines for injection into the Generator context.
// Distinguishes "file not found" (first run, silent fallback) from
// "JSON corrupted" (diagnostic issue, ERROR log).
func loadKnownFailurePatterns(root, skill string, limit int) string {
	entries, err := memory.GetPatternsBySkill(root, skill, limit)
	if err != nil {
		if os.IsNotExist(err) {
			// First run — JSON store doesn't exist yet, fallback to markdown silently
			return loadKnownFailurePatternsFromMarkdown(root, skill, limit)
		}
		// JSON store exists but is corrupted or unreadable — log ERROR
		fmt.Fprintf(os.Stderr, "ERROR: gcl.run | failure-patterns.json corrupted, falling back to markdown | %v\n", err)
		return loadKnownFailurePatternsFromMarkdown(root, skill, limit)
	}
	if len(entries) == 0 {
		// No patterns in JSON store yet — try markdown fallback
		return loadKnownFailurePatternsFromMarkdown(root, skill, limit)
	}
	var out []string
	for _, e := range entries {
		out = append(out, fmt.Sprintf("- %s (count=%d): %s → %s", e.Category, e.Count, e.Pattern, e.Fix))
	}
	return strings.Join(out, "\n")
}

// loadKnownFailurePatternsFromMarkdown is the legacy fallback that reads
// docs/failure-patterns.md line-by-line. Used when the JSON store is empty.
func loadKnownFailurePatternsFromMarkdown(root, skill string, limit int) string {
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

// writebackFailurePattern persists a failure pattern to both the legacy
// markdown store (docs/failure-patterns.md) and the structured JSON store
// (.runtime/memory/failure-patterns.json). If the pattern's count reaches
// the threshold (>= 5, matches promote.LevelConstraint), it triggers
// auto-transpile to guardrails.yaml.
func writebackFailurePattern(root, skill string, fp *trace.FailurePattern) {
	if fp == nil {
		return
	}
	// 1. Write to markdown (legacy, keep compatibility — best-effort, non-blocking)
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
		fmt.Fprintf(os.Stderr, "WARN: Reflexion markdown write-back skipped: %v\n", err)
	}

	// 2. Write to structured JSON store (count++ on dedup)
	//    This is the authoritative store — if it fails, skip transpile check.
	entry := memory.FailurePatternEntry{
		Skill:    skill,
		Pattern:  fp.Error,
		Category: fp.Category,
		Fix:      fp.Fix,
		Source:   orString(fp.Command, "gcl-runner"),
		Count:    1, // AppendFailurePattern will increment if exists
	}
	if err := memory.AppendFailurePattern(root, entry); err != nil {
		fmt.Fprintf(os.Stderr, "WARN: Reflexion JSON write-back skipped: %v\n", err)
		return // JSON is authoritative; if it fails, skip transpile
	}

	// 3. Check if count reached threshold → auto-transpile (in-process)
	entries, err := memory.GetPatternsBySkill(root, skill, 0) // 0 = all
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARN: Reflexion count check failed: %v\n", err)
		return
	}
	for _, e := range entries {
		if e.Pattern == fp.Error && e.Count >= 5 {
			fmt.Fprintf(os.Stderr, "INFO: pattern count=%d reached threshold, triggering auto-transpile\n", e.Count)
			triggerTranspile(root)
			break
		}
	}
}

// triggerTranspile regenerates guardrails.yaml in-process from the JSON memory
// store when a failure pattern reaches the count >= 5 threshold
// (matches promote.LevelConstraint).
// This avoids shelling out to `vet` (fragile PATH dependency) and uses the
// authoritative JSON store (which has count data) instead of the markdown file
// (which lacks count).
func triggerTranspile(root string) {
	entries, err := memory.LoadFailurePatterns(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARN: auto-transpile: load patterns failed: %v\n", err)
		return
	}

	// Convert memory entries to transpile.FailurePattern and filter count >= 5
	var guardrails []transpile.Guardrail
	for _, e := range entries {
		fp := transpile.FailurePattern{
			Skill:   e.Skill,
			Pattern: e.Pattern,
			Fix:     e.Fix,
			Count:   e.Count,
		}
		if g, ok := transpile.Transpile(fp); ok {
			guardrails = append(guardrails, g)
		}
	}

	if len(guardrails) == 0 {
		return
	}

	outPath := joinPath(root, "incident-loop-agent", "references", "policies", "guardrails.yaml")
	doc := struct {
		Guardrails []transpile.Guardrail `yaml:"guardrails"`
	}{Guardrails: guardrails}
	data, err := yaml.Marshal(doc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARN: auto-transpile: marshal failed: %v\n", err)
		return
	}
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "WARN: auto-transpile: write failed: %v\n", err)
		return
	}
	fmt.Fprintf(os.Stderr, "INFO: auto-transpile: %d guardrails written to %s\n", len(guardrails), outPath)
}

// persistCriticFailureTrace writes a stub trace when the critic infrastructure
// fails (exit code 2), ensuring every failure has a trace record for analysis.
func persistCriticFailureTrace(root string, tr *trace.Trace, skill, command, reason, runID string) {
	tr.Final = trace.Final{
		Status: "CRITIC_FAILURE",
		FailurePattern: &trace.FailurePattern{
			Category: "critic_infrastructure",
			Skill:    skill,
			Command:  command,
			Error:    reason,
			Fix:      "Check critic configuration: --critic-json, --critic-stdin, --critic-command, or --structural-critic-only",
		},
	}
	path, err := trace.PersistTrace(root, "", tr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s] [ERROR] gcl.run | CRITIC_FAILURE trace persist failed: %v\n", runID, err)
		return
	}
	fmt.Fprintf(os.Stderr, "[%s] [ERROR] gcl.run | CRITIC_FAILURE | skill=%s reason=%s trace=%s\n",
		runID, skill, reason, path)
}

// Result is the outcome of a Run, carrying the process exit code plus the
// timing/error signals needed by upstream reporters (vet gcl gate).
type Result struct {
	ExitCode   int
	TimedOut   bool
	TraceLine  string
	StderrLine string
}

// Run executes the GCL loop and returns a Result. The exit code matches the
// legacy Python runner: 0 PASS, 1 MAX_ITER, 2 invalid critic, 3 SAFETY_FAIL;
// -1 on an unexpected internal error.
func Run(opts Options) Result {
	if opts.MaxIter <= 0 {
		opts.MaxIter = SkillMaxIter[opts.Skill]
		if opts.MaxIter == 0 {
			opts.MaxIter = 3
		}
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 120
	}

	runID := fmt.Sprintf("%08x", time.Now().UnixNano()%0x100000000) // 8-char hex run ID
	startTime := time.Now()
	fmt.Fprintf(os.Stderr, "[%s] [INFO] gcl.run | GCL run start | skill=%s max_iter=%d heal=%s\n",
		runID, opts.Skill, opts.MaxIter, opts.Heal)

	operationIntent := deriveOperationIntent(opts.Skill, opts.Command)
	maskedFields := secret.DetectCredentialFields(opts.Command)
	knownPatterns := loadKnownFailurePatterns(opts.Root, opts.Skill, 10)

	// P0-3: Pre-flight Cost Gate (advisory — logs warning, never blocks)
	costImpact := costgate.EstimateCostImpact(opts.Skill, opts.Command)
	if costImpact != nil {
		fmt.Fprintf(os.Stderr, "[%s] [INFO] gcl.run | cost_gate | operation=%s billing_model=%s est_monthly=%.2f refund=%.2f delta=%.2f warning=%q\n",
			runID, costImpact.Operation, costImpact.BillingModel,
			costImpact.EstMonthlyCost, costImpact.RefundOnDelete,
			costImpact.NetMonthlyDelta, costImpact.Warning)
	}

	// L4 self-healing telemetry (T11): accumulate retry outcomes when smart
	// heal is active. FallbackUsed is incremented by runGeneratorWithHeal
	// when a secondary path is tried (T10 multi-path healing).
	var healMetrics *heal.Metrics
	healLogPath := ""
	if opts.Heal == "smart" {
		healMetrics = &heal.Metrics{}
		healLogPath = heal.DefaultLogPath
	}

	tr := &trace.Trace{
		TraceSchemaVersion: "v1",
		RunID:              runID,
		Skill:              opts.Skill,
		Request:            opts.Request,
		RubricVersion:      "v1",
		OperationIntent:    operationIntent,
		MaskedFields:       maskedFields,
		RedactionPass:      true,
		Iterations:         []trace.Iteration{},
	}

	criticFeedback := ""
	var lastGen trace.GeneratorResult
	for iter := 1; iter <= opts.MaxIter; iter++ {
		// Execution-risk gate (L3): score the operation BEFORE running it.
		sClass, bRadius, conf, safety, metaOK := policyInputs(opts.Skill, operationIntent, nil)
		policy := scoreDecision(opts.Skill, sClass, bRadius, conf, safety, metaOK)
		// Block unless AUTO, or ASK with external confirmation + provenance.
		askOK := policy == OpAsk && opts.Confirmed && strings.TrimSpace(opts.ConfirmedBy) != ""
		if policy != OpAuto && !askOK {
			blocked := policy
			// ASK without confirmation in a non-interactive runtime has no
			// human to ask → degrade to REFUSE for the recorded decision.
			if policy == OpAsk {
				blocked = OpRefuse
			}
			tr.Iterations = append(tr.Iterations, trace.Iteration{
				Iter:           iter,
				Timestamp:      time.Now().UTC().Format(time.RFC3339),
				Generator:      trace.GeneratorResult{Command: secret.MaskSecrets(opts.Command)},
				Decision:       "POLICY_BLOCK",
				PolicyDecision: blocked.String(),
			})
			// A blocked (ASK-without-confirmation) op is a human-intervention
			// event for the telemetry SLO (T11 spec §2.2).
			if healMetrics != nil {
				healMetrics.UserInterventions++
			}
			tr.Final = trace.Final{Status: "POLICY_BLOCK", Iter: iter, Output: nil,
				FailurePattern: &trace.FailurePattern{
					Category: "execution_risk", Skill: opts.Skill, Command: opts.Command,
						Error: "operation blocked by execution-risk policy: " + blocked.String(), Fix: "escalate to human or supply --confirmed with --confirmed-by for ASK class",
					}}
			path, _ := trace.PersistTrace(opts.Root, "", tr)
			writebackFailurePattern(opts.Root, opts.Skill, tr.Final.FailurePattern)
			fmt.Fprintf(os.Stderr, "[%s] [WARN] gcl.run | POLICY_BLOCK | skill=%s decision=%s trace=%s\n",
				runID, opts.Skill, blocked.String(), path)
			return Result{ExitCode: 4, TraceLine: "blocked:" + blocked.String(), StderrLine: "blocked:" + blocked.String()}
		}

		iterStart := time.Now()
		gen, healClass, selfHeal := runGeneratorWithHeal(opts, criticFeedback, knownPatterns, healMetrics, healLogPath, runID)
		iterDurationMs := time.Since(iterStart).Milliseconds()
		// When this iteration runs an ASK-class op that was authorized by an
		// external confirmation, stamp the confirmation provenance into the
		// trace so the audit trail answers "who authorized this op". AUTO ops
		// need no confirmation; REFUSE never reaches here.
		confirmedBy := ""
		if policy == OpAsk && opts.Confirmed {
			confirmedBy = opts.ConfirmedBy
		}
		gen.Args = map[string]any{"iter": iter, "critic_feedback": orEmpty(criticFeedback), "policy_decision": policy.String()}
		if confirmedBy != "" {
			gen.Args["confirmed_by"] = confirmedBy
		}
		lastGen = gen

		var c *critic.CriticResult
		if opts.StructuralOnly {
			g := critic.GeneratorResult{Command: gen.Command, ExitCode: gen.ExitCode, ResultExcerpt: gen.ResultExcerpt}
			res := critic.StructuralCritic(g)
			c = &res
		} else {
			if secret.HasCredentialLeak(gen.ResultExcerpt) {
				tr.Iterations = append(tr.Iterations, trace.Iteration{
					Iter:      iter,
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					Generator: gen,
					Critic: trace.CriticRecord{
						Scores:      map[string]float64{"correctness": 0, "safety": 0, "idempotency": 0.5, "traceability": 0.5, "spec_compliance": 0.5},
						Suggestions: []string{"Credential leak in generator output — mask secrets and re-run"},
						Blocking:    true,
					},
					Decision:       "SAFETY_FAIL",
					PolicyDecision: "AUTO",
				})
				fp := extractFailurePattern(opts.Skill, opts.Command, gen, nil)
				tr.Final = trace.Final{Status: "SAFETY_FAIL", Iter: iter, Output: nil, FailurePattern: fp}
				path, _ := trace.PersistTrace(opts.Root, "", tr)
				writebackFailurePattern(opts.Root, opts.Skill, fp)
				fmt.Fprintf(os.Stderr, "[%s] [ERROR] gcl.run | SAFETY_FAIL | credential_leak skill=%s iter=%d trace=%s\n",
				runID, opts.Skill, iter, path)
				return Result{ExitCode: 3, TraceLine: gen.ResultExcerpt, StderrLine: gen.StderrExcerpt}
			}
		loaded, err := loadCritic(opts.CriticJSON, opts.CriticStdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[%s] [ERROR] gcl.run | invalid critic JSON | %v\n", runID, err)
			persistCriticFailureTrace(opts.Root, tr, opts.Skill, opts.Command, "invalid critic JSON: "+err.Error(), runID)
			return Result{ExitCode: 2, TraceLine: gen.ResultExcerpt, StderrLine: gen.StderrExcerpt}
		}
		if loaded == nil && opts.CriticCommand != "" {
			loaded2, err := runIsolatedCritic(opts, operationIntent, gen, tr.Iterations, runID)
			if err != nil {
				persistCriticFailureTrace(opts.Root, tr, opts.Skill, opts.Command, "critic command failed: "+err.Error(), runID)
				return Result{ExitCode: 2, TraceLine: gen.ResultExcerpt, StderrLine: gen.StderrExcerpt}
			}
			loaded = loaded2
		}
		if loaded == nil {
			fmt.Fprintf(os.Stderr, "[%s] [ERROR] gcl.run | no critic payload | Pass --critic-json, pipe JSON to stdin, --critic-command <cmd>, or use --structural-critic-only\n", runID)
			persistCriticFailureTrace(opts.Root, tr, opts.Skill, opts.Command, "no critic payload", runID)
			return Result{ExitCode: 2, TraceLine: gen.ResultExcerpt, StderrLine: gen.StderrExcerpt}
		}
		if errs := critic.ValidatePayload(*loaded); len(errs) > 0 {
			fmt.Fprintf(os.Stderr, "[%s] [ERROR] gcl.run | invalid critic payload | %s\n", runID, strings.Join(errs, "; "))
			persistCriticFailureTrace(opts.Root, tr, opts.Skill, opts.Command, "invalid critic payload: "+strings.Join(errs, "; "), runID)
			return Result{ExitCode: 2, TraceLine: gen.ResultExcerpt, StderrLine: gen.StderrExcerpt}
			}
			c = loaded
		}

		decision := critic.Decide(c.Scores)
		// Capture the cloud API RequestId from this iteration's `ve` call so
		// the runtime trace is end-to-end traceable (P5).
		requestID := parseRequestID(gen.ResultExcerpt)
		tr.Iterations = append(tr.Iterations, trace.Iteration{
			Iter:           iter,
			Timestamp:      iterStart.UTC().Format(time.RFC3339),
			DurationMs:     iterDurationMs,
			Generator:      gen,
			Critic:         trace.CriticRecord{Scores: c.Scores, Suggestions: c.Suggestions, Blocking: c.Blocking},
			Decision:       decision,
			PolicyDecision: policy.String(),
			ConfirmedBy:    confirmedBy,
			RequestID:      requestID,
			HealClass:      healClass,
			SelfHealing:    selfHeal,
			CostImpact: func() *trace.CostImpactRecord {
				if costImpact == nil {
					return nil
				}
				return &trace.CostImpactRecord{
					Operation:       costImpact.Operation,
					BillingModel:    costImpact.BillingModel,
					EstMonthlyCost:  costImpact.EstMonthlyCost,
					RefundOnDelete:  costImpact.RefundOnDelete,
					NetMonthlyDelta: costImpact.NetMonthlyDelta,
					Warning:         costImpact.Warning,
				}
			}(),
		})

		if decision == "SAFETY_FAIL" {
			fp := extractFailurePattern(opts.Skill, opts.Command, gen, c)
			tr.Final = trace.Final{Status: "SAFETY_FAIL", Iter: iter, Output: nil, FailurePattern: fp}
			path, _ := trace.PersistTrace(opts.Root, "", tr)
			writebackFailurePattern(opts.Root, opts.Skill, fp)
			fmt.Fprintf(os.Stderr, "[%s] [ERROR] gcl.run | SAFETY_FAIL | skill=%s iter=%d trace=%s\n",
				runID, opts.Skill, iter, path)
			return Result{ExitCode: 3, TraceLine: gen.ResultExcerpt, StderrLine: gen.StderrExcerpt}
		}
		if decision == "PASS" {
			out := gen.ResultExcerpt
			tr.Final = trace.Final{Status: "PASS", Iter: iter, Output: &out}
			path, _ := trace.PersistTrace(opts.Root, "", tr)
			fmt.Fprintf(os.Stderr, "[%s] [INFO] gcl.run | PASS | skill=%s iter=%d total_duration_ms=%d trace=%s\n",
				runID, opts.Skill, iter, time.Since(startTime).Milliseconds(), path)
			return Result{ExitCode: 0, TraceLine: gen.ResultExcerpt, StderrLine: gen.StderrExcerpt}
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
	fmt.Fprintf(os.Stderr, "[%s] [WARN] gcl.run | MAX_ITER | skill=%s iter=%d total_duration_ms=%d trace=%s\n",
		runID, opts.Skill, opts.MaxIter, time.Since(startTime).Milliseconds(), path)
	return Result{
		ExitCode:   1,
		TimedOut:   strings.HasPrefix(lastGen.ResultExcerpt, "TIMEOUT"),
		TraceLine:  lastGen.ResultExcerpt,
		StderrLine: lastGen.StderrExcerpt,
	}
}

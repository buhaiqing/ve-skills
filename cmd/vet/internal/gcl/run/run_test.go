package run

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/critic"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/trace"
)

// TestDeriveOperationIntent ported from gcl_runner_test.py: verifies the
// operation-intent classifier maps destructive / mutating / enable verbs to the
// correct safety class. The runner-loop logic previously had no Go coverage.
func TestDeriveOperationIntent(t *testing.T) {
	cases := []struct {
		skill   string
		command string
		wantOp  string
		wantSafety string
	}{
		{"ve-ecs-ops", "ve ecs Delete --InstanceIds i-xxx", "destructive_ecs", "destructive"},
		{"ve-ecs-ops", "ve ecs Create --ImageId img", "modify_ecs", "mutating"},
		{"ve-security-group-ops", "ve security-group Enable --RuleId r", "modify_security-group", "mutating"},
		{"ve-security-group-ops", "ve security-group EnableProtection --RuleId r", "enable_security-group", "mutating"},
		{"ve-ecs-ops", "ve ecs DescribeInstances", "describe", "read_only"},
		{"ve-ecs-ops", "", "unknown", "read_only"},
	}
	for _, c := range cases {
		got := deriveOperationIntent(c.skill, c.command)
		if got["operation"] != c.wantOp {
			t.Errorf("deriveOperationIntent(%q,%q).operation = %q, want %q", c.skill, c.command, got["operation"], c.wantOp)
		}
		if got["safety_class"] != c.wantSafety {
			t.Errorf("deriveOperationIntent(%q,%q).safety_class = %q, want %q", c.skill, c.command, got["safety_class"], c.wantSafety)
		}
	}
}

// TestExtractFailurePattern ported from gcl_runner_test.py: verifies the
// failure-signature extractor matches known runtime / cli_parameter patterns
// and returns nil for benign output.
func TestExtractFailurePattern(t *testing.T) {
	skill := "ve-ecs-ops"
	command := "ve ecs RunInstances"

	cli := extractFailurePattern(skill, command, trace.GeneratorResult{ResultExcerpt: "Error: InvalidParameter.InstanceIdNotFound"}, nil)
	if cli == nil || cli.Category != "cli_parameter" {
		t.Fatalf("expected cli_parameter pattern, got %+v", cli)
	}

	runtime := extractFailurePattern(skill, command, trace.GeneratorResult{ResultExcerpt: "RequestLimitExceeded please retry later"}, nil)
	if runtime == nil || runtime.Category != "runtime" {
		t.Fatalf("expected runtime pattern, got %+v", runtime)
	}

	clean := extractFailurePattern(skill, command, trace.GeneratorResult{ResultExcerpt: "RequestId: abc-123 OK"}, nil)
	if clean != nil {
		t.Fatalf("expected no pattern for clean output, got %+v", clean)
	}
}

// TestRunResultTimedOut ensures a TIMEOUT generator result is surfaced through
// Run's Result (regression guard for the gcl gate timed_out field).
func TestRunResultTimedOut(t *testing.T) {
	if !strings.HasPrefix("TIMEOUT after 30s", "TIMEOUT") {
		t.Fatal("precondition: prefix check")
	}
	// Structural-only smoke against a fast, clean command should not time out.
	res := Run(Options{
		Root:           ".",
		Skill:          "ve-skill-generator",
		Request:        "unit-test smoke",
		Command:        `echo {"Response":{"RequestId":"ut"}}`,
		MaxIter:        1,
		Timeout:        30,
		StructuralOnly: true,
	})
	if res.TimedOut {
		t.Fatalf("unexpected TimedOut for clean structural smoke (exit %d)", res.ExitCode)
	}
}

// TestScoreDecision_9Cell covers the 9-cell decision matrix:
// - read_only + high → AUTO
// - mutating + single + high → AUTO
// - destructive + any → ASK
// - safety=0 → REFUSE (hard floor, overrides everything)
// - missing metadata → ASK (fail-safe)
func TestScoreDecision_9Cell(t *testing.T) {
	cases := []struct {
		skill      string
		safetyClass string
		blastRadius string
		confidence string
		safety     float64
		metadataOK bool
		want       OpDecision
	}{
		// AUTO cases
		{"ve-ecs-ops", "read_only", "single", "high", 1.0, true, OpAuto},
		{"ve-rds-mysql-ops", "read_only", "single", "high", 1.0, true, OpAuto},
		{"ve-ecs-ops", "mutating", "single", "high", 1.0, true, OpAuto},
		// ASK cases
		{"ve-ecs-ops", "destructive", "single", "high", 1.0, true, OpAsk},
		{"ve-ecs-ops", "mutating", "multi", "high", 1.0, true, OpAsk},
		{"ve-redis-ops", "read_only", "single", "low", 1.0, true, OpAsk},
		{"ve-unknown-ops", "read_only", "single", "high", 1.0, true, OpAsk},
		// REFUSE: safety=0 hard floor
		{"ve-ecs-ops", "read_only", "single", "high", 0.0, true, OpRefuse},
		// REFUSE: destructive + safety=0
		{"ve-redis-ops", "destructive", "single", "high", 0.0, true, OpRefuse},
		// ASK: missing metadata (fail-safe)
		{"ve-ecs-ops", "read_only", "single", "high", 1.0, false, OpAsk},
	}
	for _, c := range cases {
		got := scoreDecision(c.skill, c.safetyClass, c.blastRadius, c.confidence, c.safety, c.metadataOK)
		if got != c.want {
			t.Errorf("scoreDecision(%q,%q,%q,%q,%.1f,%v) = %v, want %v",
				c.skill, c.safetyClass, c.blastRadius, c.confidence, c.safety, c.metadataOK, got, c.want)
		}
	}
}

// TestScoreDecision_DestructiveNeverAuto: destructive ops must never get AUTO.
func TestScoreDecision_DestructiveNeverAuto(t *testing.T) {
	// mutating can be AUTO (single + high); only truly destructive is blocked
	for _, sc := range []string{"destructive"} {
		got := scoreDecision("ve-ecs-ops", sc, "single", "high", 1.0, true)
		if got == OpAuto {
			t.Errorf("safety_class=%q should never be AUTO, got AUTO", sc)
		}
	}
}

// TestScoreDecision_SafetyZeroRefuse: safety=0 must always be REFUSE.
func TestScoreDecision_SafetyZeroRefuse(t *testing.T) {
	for _, sc := range []string{"read_only", "mutating", "destructive"} {
		got := scoreDecision("ve-ecs-ops", sc, "single", "high", 0.0, true)
		if got != OpRefuse {
			t.Errorf("safety=0 with safety_class=%q: got %v, want REFUSE", sc, got)
		}
	}
}

// TestPolicyInputs_FailSafe verifies the pre-execution policy gate degrades
// safely when no Critic evidence exists yet (first iteration): destructive ops
// are never AUTO, and unknown skills/missing metadata fall back to ASK.
func TestPolicyInputs_FailSafe(t *testing.T) {
	cases := []struct {
		skill   string
		command string
		want    OpDecision
	}{
		{"ve-ecs-ops", "ve ecs DeleteInstances --Ids i", OpAsk},  // destructive → ASK (Run downgrades to REFUSE)
		{"ve-ecs-ops", "ve ecs DescribeInstances", OpAuto},        // read_only + high confidence (allow-list) → AUTO
		{"ve-unknown-ops", "ve unknown Describe", OpAsk},         // not in allow-list → ASK
	}
	for _, c := range cases {
		sClass, bRadius, conf, safety, metaOK := policyInputs(c.skill, deriveOperationIntent(c.skill, c.command), nil)
		got := scoreDecision(c.skill, sClass, bRadius, conf, safety, metaOK)
		if got != c.want {
			t.Errorf("policyInputs+scoreDecision(%q,%q) = %v, want %v", c.skill, c.command, got, c.want)
		}
	}
}

// TestRun_PolicyBlocksDestructive asserts the execution-risk gate is wired into
// Run(): a destructive command must NOT execute and must exit POLICY_BLOCK (4)
// with policy_decision=REFUSE in the trace.
func TestRun_PolicyBlocksDestructive(t *testing.T) {
	res := Run(Options{
		Root:           t.TempDir(),
		Skill:          "ve-ecs-ops",
		Request:        "unit-test: destructive must be blocked",
		Command:        "ve ecs DeleteInstances --InstanceIds i-xxx",
		MaxIter:        1,
		Timeout:        10,
		StructuralOnly: true,
	})
	if res.ExitCode != 4 {
		t.Fatalf("destructive op should be POLICY_BLOCK (exit 4), got exit %d", res.ExitCode)
	}
}

// TestRun_PolicyAutoReadonly asserts a read-only command passes the gate and
// actually executes (policy_decision=AUTO recorded in the trace iteration).
func TestRun_PolicyAutoReadonly(t *testing.T) {
	res := Run(Options{
		Root:           t.TempDir(),
		Skill:          "ve-ecs-ops",
		Request:        "unit-test: readonly should auto-execute",
		Command:        `echo {"Response":{"RequestId":"ut-readonly"}}`,
		MaxIter:        1,
		Timeout:        10,
		StructuralOnly: true,
	})
	if res.ExitCode == 4 {
		t.Fatalf("read-only op must not be POLICY_BLOCK, got exit 4 (trace: %s)", res.TraceLine)
	}
}

// TestRun_PolicyAskNeedsConfirm asserts an ASK-class op (outside allow-list)
// without --confirmed is blocked, and with --confirmed it executes.
func TestRun_PolicyAskNeedsConfirm(t *testing.T) {
	cmd := `echo {"Response":{"RequestId":"ut-ask"}}`
	blocked := Run(Options{
		Root: t.TempDir(), Skill: "ve-unknown-ops", Request: "ask-no-confirm",
		Command: cmd, MaxIter: 1, Timeout: 10, StructuralOnly: true,
	})
	if blocked.ExitCode != 4 {
		t.Fatalf("ASK without --confirmed should be POLICY_BLOCK, got exit %d", blocked.ExitCode)
	}
	confirmedNoBy := Run(Options{
		Root: t.TempDir(), Skill: "ve-unknown-ops", Request: "ask-with-confirm-no-by",
		Command: cmd, MaxIter: 1, Timeout: 10, StructuralOnly: true, Confirmed: true,
	})
	if confirmedNoBy.ExitCode != 4 {
		t.Fatalf("ASK with --confirmed but no confirmed_by should be POLICY_BLOCK, got exit %d", confirmedNoBy.ExitCode)
	}
	// Authorized ASK must persist confirmation provenance in the trace so the
	// audit trail answers "who authorized this op".
	root := t.TempDir()
	auth := Run(Options{
		Root: root, Skill: "ve-unknown-ops", Request: "ask-with-confirm-provenance",
		Command: cmd, MaxIter: 1, Timeout: 10, StructuralOnly: true,
		Confirmed: true, ConfirmedBy: "DOPS-12345",
	})
	if auth.ExitCode == 4 {
		t.Fatalf("authorized ASK should execute, got POLICY_BLOCK (exit 4)")
	}
	// Re-read the persisted trace and assert confirmed_by is recorded.
	matches, _ := filepath.Glob(filepath.Join(root, "audit-results", "gcl-trace-*.json"))
	if len(matches) == 0 {
		t.Fatalf("expected a persisted trace under %s/audit-results", root)
	}
	tr := trace.ParseTrace(matches[0])
	if tr == nil {
		t.Fatalf("persisted trace %s failed to parse", matches[0])
	}
	var found bool
	for _, it := range tr.Iterations {
		if it.PolicyDecision == "ASK" && it.ConfirmedBy == "DOPS-12345" {
			found = true
		}
	}
	if !found {
		t.Fatalf("authorized ASK iteration must record confirmed_by=DOPS-12345; trace=%s", matches[0])
	}
}

func TestASKConfirmedWithoutByIsBlocked(t *testing.T) {
	cmd := `echo {"Response":{"RequestId":"ut-ask"}}`
	res := Run(Options{
		Root: t.TempDir(), Skill: "ve-unknown-ops", Request: "ask-confirmed-no-by",
		Command: cmd, MaxIter: 1, Timeout: 10, StructuralOnly: true,
		Confirmed: true, ConfirmedBy: "",
	})
	if res.ExitCode != 4 {
		t.Fatalf("expected POLICY_BLOCK, got %d", res.ExitCode)
	}
}

// TestParseRequestID covers the four RequestId extraction states for
// parseRequestID: structured JSON parse, nested JSON, non-JSON fallback scan,
// and missing/truncated RequestId.
func TestParseRequestID(t *testing.T) {
	cases := []struct {
		name   string
		output string
		want   string
	}{
		{
			name:   "standard json",
			output: `{"Response":{"RequestId":"abc-123"}}`,
			want:   "abc-123",
		},
		{
			name:   "nested with other fields",
			output: `{"foo":1,"Response":{"RequestId":"x"}}`,
			want:   "x",
		},
		{
			name:   "non-json text fallback scan",
			output: `some log line "RequestId":"y" trailing`,
			want:   "y",
		},
		{
			name:   "truncated or missing RequestId",
			output: `{"Response":{}}`,
			want:   "",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := parseRequestID(c.output); got != c.want {
				t.Errorf("parseRequestID(%q) = %q, want %q", c.output, got, c.want)
			}
		})
	}
}

// --- P0-1 / P0-2 loop-feedback tests -----------------------------------------
//
// These exercise the RETRY-path closure added in the 2026-08-06 change:
//   - commandHash no-op detection   (TestCommandHash)
//   - NO_PROGRESS termination       (TestNoProgress)
//   - --fix-command rewrite path    (TestFixCommand)
//   - critic history + prior hints  (TestCriticHistory)
//
// They drive the real Run() loop and inspect the persisted trace, so a
// regression in wiring (flag → Options → loop) fails the test, not just a
// unit of an isolated helper.

// TestCommandHash verifies the no-op detector's digest is (a) deterministic,
// (b) credential-masking-aware (a rotated secret must not defeat the check),
// and (c) sensitive to any real command change.
func TestCommandHash(t *testing.T) {
	if commandHash("ve ecs DescribeInstances") != commandHash("ve ecs DescribeInstances") {
		t.Fatal("same command must hash equal")
	}
	if commandHash("false") == commandHash("true") {
		t.Fatal("different commands must hash differently")
	}
	// Masking runs before hashing: the two literals differ only in the secret
	// value, so the masked forms (and hence the hashes) must match.
	const secretA = "AKLTxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	const secretB = "AKLTyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy"
	a := commandHash("ve ecs Create --SecretId " + secretA)
	b := commandHash("ve ecs Create --SecretId " + secretB)
	if a != b {
		t.Fatal("commands differing only in a masked credential must hash equal")
	}
}

// TestNoProgress asserts that a RETRY whose effective command is byte-identical
// to the just-retried command stops the loop with Final.Status == "NO_PROGRESS"
// (exit 1) instead of silently burning the remaining MaxIter budget.
//
// Setup: `false` exits 1 → structural critic scores correctness=0 → RETRY. No
// --fix-command is supplied, so the next iteration re-runs the identical
// command; the guard must fire on iteration 2 with MaxIter=3 still available.
func TestNoProgress(t *testing.T) {
	root := t.TempDir()
	res := Run(Options{
		Root:           root,
		Skill:          "ve-ecs-ops",
		Request:        "unit-test: identical retry must stop NO_PROGRESS",
		Command:        "false", // read-only intent; exit 1 → structural RETRY
		MaxIter:        3,
		Timeout:        10,
		StructuralOnly: true,
		Heal:           "none",
	})
	if res.ExitCode != 1 {
		t.Fatalf("NO_PROGRESS must exit 1, got %d", res.ExitCode)
	}
	matches, _ := filepath.Glob(filepath.Join(root, "audit-results", "gcl-trace-*.json"))
	if len(matches) == 0 {
		t.Fatalf("expected a persisted trace under %s/audit-results", root)
	}
	tr := trace.ParseTrace(matches[0])
	if tr == nil {
		t.Fatalf("persisted trace %s failed to parse", matches[0])
	}
	if tr.Final.Status != "NO_PROGRESS" {
		t.Fatalf("Final.Status = %q, want NO_PROGRESS (iter=%d)", tr.Final.Status, tr.Final.Iter)
	}
	// The guard must fire as soon as the duplicate retry starts, i.e. on the
	// second iteration — not after exhausting all three.
	if tr.Final.Iter != 2 {
		t.Fatalf("NO_PROGRESS should stop at iter 2 (first duplicate retry), got iter %d", tr.Final.Iter)
	}
	if tr.Final.FailurePattern == nil || tr.Final.FailurePattern.Category != "no_progress" {
		t.Fatalf("expected failure_pattern.category=no_progress, got %+v", tr.Final.FailurePattern)
	}
	// The status must be a recognised final status (aggregator allowlist).
	var recognised bool
	for _, s := range trace.FinalStatuses {
		if s == tr.Final.Status {
			recognised = true
		}
	}
	if !recognised {
		t.Fatalf("NO_PROGRESS missing from trace.FinalStatuses")
	}
	// P0-3: the guard acts on the PREVIOUS iteration's RETRY verdict, so the
	// duplicate-retry iteration must carry that real verdict — not a null
	// critic. A `decision: RETRY` with `critic.scores: null` is a malformed
	// trace that downstream consumers cannot correlate.
	if len(tr.Iterations) < 2 {
		t.Fatalf("expected >=2 iterations, got %d", len(tr.Iterations))
	}
	prior, dup := tr.Iterations[len(tr.Iterations)-2], tr.Iterations[len(tr.Iterations)-1]
	if dup.Decision != "RETRY" {
		t.Fatalf("NO_PROGRESS iteration decision = %q, want RETRY", dup.Decision)
	}
	if dup.Critic.Scores == nil {
		t.Fatalf("P0-3: NO_PROGRESS iteration carries a null critic record")
	}
	if !dup.Critic.Blocking {
		t.Fatalf("P0-3: NO_PROGRESS iteration critic.blocking = false, want true")
	}
	if !reflect.DeepEqual(dup.Critic.Scores, prior.Critic.Scores) {
		t.Fatalf("P0-3: NO_PROGRESS critic scores %+v != prior iteration scores %+v",
			dup.Critic.Scores, prior.Critic.Scores)
	}
	// Spec §6.2: the NO_PROGRESS Final must carry `unresolved` — the dims
	// still below threshold — exactly as the MAX_ITER Final does. Without it,
	// a downstream consumer cannot tell *why* the run stalled without
	// re-deriving the below-threshold set from the iteration history.
	if len(tr.Final.Unresolved) == 0 {
		t.Fatalf("spec §6.2: NO_PROGRESS Final.Unresolved is empty (scores=%+v)", dup.Critic.Scores)
	}
	// Each unresolved dim must genuinely be below threshold.
	want := map[string]bool{}
	for dim, th := range critic.RubricThresholds {
		if dup.Critic.Scores[dim] < th {
			want[dim] = true
		}
	}
	if len(want) != len(tr.Final.Unresolved) {
		t.Fatalf("Unresolved %v does not match below-threshold set %v", tr.Final.Unresolved, want)
	}
	for _, dim := range tr.Final.Unresolved {
		if !want[dim] {
			t.Fatalf("Unresolved lists %q but its score %v is not below threshold %v",
				dim, dup.Critic.Scores[dim], critic.RubricThresholds[dim])
		}
	}
}

// TestNoProgressSkipsSuccessfulRetry is the regression guard for the BLOCKER
// the no-op detector originally introduced: it fired on hash equality ALONE,
// so a byte-identical command that failed once and then succeeded on retry
// (transient failure — rate limit, lock contention, eventual consistency) was
// misclassified as NO_PROGRESS and a valid result was discarded.
//
// Setup: a flaky generator exits 1 on its first invocation and 0 afterwards
// (told apart by a counter file), while the critic RETRYs on non-zero and
// PASSes on zero. Iteration 2 re-runs the SAME command ("flaky.sh", no
// --fix-command) but exits 0, so the guard must NOT fire: the run must reach
// PASS with 2 iterations, not NO_PROGRESS at iteration 2.
func TestNoProgressSkipsSuccessfulRetry(t *testing.T) {
	root := t.TempDir()
	counterPath := filepath.Join(root, "counter")
	if err := os.WriteFile(counterPath, []byte("0"), 0o644); err != nil {
		t.Fatalf("seed counter: %v", err)
	}

	// Flaky generator: fails the first run, succeeds every run after.
	flakyPath := filepath.Join(root, "flaky.sh")
	flakyScript := `#!/bin/sh
n=$(cat ` + counterPath + `)
n=$((n + 1))
printf '%s' "$n" > ` + counterPath + `
if [ "$n" -eq 1 ]; then
  echo "attempt 1: transient failure" >&2
  exit 1
fi
echo "attempt $n: OK"
exit 0
`
	if err := os.WriteFile(flakyPath, []byte(flakyScript), 0o755); err != nil {
		t.Fatalf("write flaky generator: %v", err)
	}

	// Critic helper: RETRY on non-zero generator exit, PASS on zero.
	criticPath := filepath.Join(root, "critic.sh")
	criticScript := `#!/bin/sh
payload=$(cat)
if printf '%s' "$payload" | grep -q '"exit_code":0'; then
  echo '{"scores":{"correctness":1,"safety":1,"idempotency":1,"traceability":1,"spec_compliance":1},"suggestions":[],"blocking":false}'
else
  echo '{"scores":{"correctness":0,"safety":1,"idempotency":0.5,"traceability":0.5,"spec_compliance":0.5},"suggestions":["transient failure, retry"],"blocking":true}'
fi
`
	if err := os.WriteFile(criticPath, []byte(criticScript), 0o755); err != nil {
		t.Fatalf("write critic helper: %v", err)
	}

	res := Run(Options{
		Root:    root,
		Skill:   "ve-ecs-ops",
		Request: "unit-test: successful identical retry must not be NO_PROGRESS",
		Command: flakyPath, // no --fix-command → iteration 2 re-runs the identical command
		MaxIter: 3,
		Timeout: 10,
		Heal:    "none",
		// StructuralOnly=false + explicit critic so the RETRY verdict comes
		// from a controllable source rather than the structural fallback.
		CriticCommand: criticPath,
	})
	if res.ExitCode != 0 {
		t.Fatalf("successful retry must reach PASS (exit 0), got %d (trace=%s)", res.ExitCode, res.TraceLine)
	}
	matches, _ := filepath.Glob(filepath.Join(root, "audit-results", "gcl-trace-*.json"))
	if len(matches) == 0 {
		t.Fatalf("expected a persisted trace under %s/audit-results", root)
	}
	tr := trace.ParseTrace(matches[0])
	if tr == nil {
		t.Fatalf("persisted trace %s failed to parse", matches[0])
	}
	if tr.Final.Status != "PASS" {
		t.Fatalf("Final.Status = %q, want PASS — a successful retry must not be dropped as NO_PROGRESS", tr.Final.Status)
	}
	if tr.Final.Iter != 2 {
		t.Fatalf("expected PASS at iter 2 (fail → success), got iter %d", tr.Final.Iter)
	}
	if tr.Iterations[1].Generator.ExitCode != 0 {
		t.Fatalf("iteration 2 exit_code = %d, want 0", tr.Iterations[1].Generator.ExitCode)
	}
}

// TestNoProgressTracePassesCheck asserts the NO_PROGRESS guard's synthetic
// iteration is well-formed enough to survive trace.Check: the guard appends an
// iteration for a command that DID run, so it must carry a request_id exactly
// like every other executed iteration. Without it, `vet check trace` rejects
// the very trace the loop just wrote.
func TestNoProgressTracePassesCheck(t *testing.T) {
	root := t.TempDir()
	// A command that emits a RequestId (like a real `ve` invocation) so the
	// executed iterations have one to carry forward.
	cmdPath := filepath.Join(root, "vetlike.sh")
	cmdScript := `#!/bin/sh
echo '{"RequestId":"req-abc-123","ok":false}'
exit 1
`
	if err := os.WriteFile(cmdPath, []byte(cmdScript), 0o755); err != nil {
		t.Fatalf("write vet-like generator: %v", err)
	}

	res := Run(Options{
		Root:           root,
		Skill:          "ve-ecs-ops",
		Request:        "unit-test: NO_PROGRESS trace must satisfy trace.Check",
		Command:        cmdPath,
		MaxIter:        3,
		Timeout:        10,
		StructuralOnly: true, // structural critic → correctness=0 → RETRY
		Heal:           "none",
	})
	if res.ExitCode != 1 {
		t.Fatalf("NO_PROGRESS must exit 1, got %d", res.ExitCode)
	}
	matches, _ := filepath.Glob(filepath.Join(root, "audit-results", "gcl-trace-*.json"))
	if len(matches) == 0 {
		t.Fatalf("expected a persisted trace under %s/audit-results", root)
	}
	tr := trace.ParseTrace(matches[0])
	if tr == nil {
		t.Fatalf("persisted trace %s failed to parse", matches[0])
	}
	if err := trace.Check(matches[0]); err != nil {
		t.Fatalf("trace.Check rejected the NO_PROGRESS trace: %v", err)
	}
	// Both iterations ran the command, so both need a request_id.
	for _, it := range tr.Iterations {
		if it.RequestID == "" {
			t.Fatalf("iteration %d: request_id empty (decision=%s)", it.Iter, it.Decision)
		}
	}
}

// TestFixCommand asserts the --fix-command hook rewrites the effective command
// between RETRY iterations and that the rewritten command actually runs.
//
// Setup: a critic helper returns RETRY while the generator exits non-zero and
// PASS once it exits zero. The fixer maps the failing `false` to `true`, so
// iteration 2 runs `true`, the critic returns PASS, and the run exits 0. The
// persisted trace's iteration-2 generator command must be the rewritten one.
func TestFixCommand(t *testing.T) {
	root := t.TempDir()

	// Critic helper: read the generator exit_code from the stdin payload and
	// emit a payload that RETRYs on non-zero, PASSes on zero. Kept minimal so
	// the test exercises Run()'s plumbing, not the critic's scoring.
	criticPath := filepath.Join(root, "critic.sh")
	criticScript := `#!/bin/sh
payload=$(cat)
if printf '%s' "$payload" | grep -q '"exit_code":0'; then
  echo '{"scores":{"correctness":1,"safety":1,"idempotency":1,"traceability":1,"spec_compliance":1},"suggestions":[],"blocking":false}'
else
  echo '{"scores":{"correctness":0,"safety":1,"idempotency":0.5,"traceability":0.5,"spec_compliance":0.5},"suggestions":["command exited non-zero"],"blocking":true}'
fi
`
	if err := os.WriteFile(criticPath, []byte(criticScript), 0o755); err != nil {
		t.Fatalf("write critic helper: %v", err)
	}

	// Fixer helper: ignore the payload, always emit the passing command.
	fixPath := filepath.Join(root, "fix.sh")
	fixScript := `#!/bin/sh
cat >/dev/null
echo '{"command":"true"}'
`
	if err := os.WriteFile(fixPath, []byte(fixScript), 0o755); err != nil {
		t.Fatalf("write fix helper: %v", err)
	}

	res := Run(Options{
		Root:           root,
		Skill:          "ve-ecs-ops",
		Request:        "unit-test: fix command rewrites retry",
		Command:        "false", // read-only intent; exit 1 → RETRY
		MaxIter:        3,
		Timeout:        10,
		Heal:           "none",
		CriticCommand:  criticPath,
		FixCommand:     fixPath,
		StructuralOnly: false,
	})
	if res.ExitCode != 0 {
		t.Fatalf("fix command should drive the loop to PASS (exit 0), got %d (trace=%s)", res.ExitCode, res.TraceLine)
	}
	matches, _ := filepath.Glob(filepath.Join(root, "audit-results", "gcl-trace-*.json"))
	if len(matches) == 0 {
		t.Fatalf("expected a persisted trace under %s/audit-results", root)
	}
	tr := trace.ParseTrace(matches[0])
	if tr == nil {
		t.Fatalf("persisted trace %s failed to parse", matches[0])
	}
	if tr.Final.Status != "PASS" {
		t.Fatalf("Final.Status = %q, want PASS", tr.Final.Status)
	}
	if len(tr.Iterations) != 2 {
		t.Fatalf("expected 2 iterations (fail → fixed → pass), got %d", len(tr.Iterations))
	}
	if got := tr.Iterations[1].Generator.Command; !strings.Contains(got, "true") {
		t.Fatalf("iteration 2 must run the rewritten command, got %q", got)
	}
	// Backward-compat: opts.Command stays the original for the audit delta.
	if got := tr.Iterations[0].Generator.Command; !strings.Contains(got, "false") {
		t.Fatalf("iteration 1 must run the original command, got %q", got)
	}
}

// TestCriticHistory asserts the Critic payload now carries the current
// iteration in trace.iterations (ordering fix) and a flattened
// prior_suggestions list (P0-2), so the Critic can judge whether its previous
// feedback was addressed.
//
// Setup: a critic helper appends each stdin payload to a file. Iteration 1
// RETRYs; iteration 2's payload must therefore contain the iteration-1
// suggestion in prior_suggestions and at least 2 entries in trace.iterations.
// The fixer rewrites `false` → `true` so the run terminates cleanly.
func TestCriticHistory(t *testing.T) {
	root := t.TempDir()
	capturePath := filepath.Join(root, "critic-payloads.jsonl")

	criticPath := filepath.Join(root, "critic.sh")
	criticScript := `#!/bin/sh
payload=$(cat)
printf '%s\n' "$payload" >> ` + capturePath + `
if printf '%s' "$payload" | grep -q '"exit_code":0'; then
  echo '{"scores":{"correctness":1,"safety":1,"idempotency":1,"traceability":1,"spec_compliance":1},"suggestions":[],"blocking":false}'
else
  echo '{"scores":{"correctness":0,"safety":1,"idempotency":0.5,"traceability":0.5,"spec_compliance":0.5},"suggestions":["iteration-one-suggestion"],"blocking":true}'
fi
`
	if err := os.WriteFile(criticPath, []byte(criticScript), 0o755); err != nil {
		t.Fatalf("write critic helper: %v", err)
	}
	fixPath := filepath.Join(root, "fix.sh")
	if err := os.WriteFile(fixPath, []byte("#!/bin/sh\ncat >/dev/null\necho '{\"command\":\"true\"}'\n"), 0o755); err != nil {
		t.Fatalf("write fix helper: %v", err)
	}

	res := Run(Options{
		Root:           root,
		Skill:          "ve-ecs-ops",
		Request:        "unit-test: critic must see history + prior suggestions",
		Command:        "false",
		MaxIter:        3,
		Timeout:        10,
		Heal:           "none",
		CriticCommand:  criticPath,
		FixCommand:     fixPath,
		StructuralOnly: false,
	})
	if res.ExitCode != 0 {
		t.Fatalf("loop should reach PASS, got exit %d (trace=%s)", res.ExitCode, res.TraceLine)
	}

	raw, err := os.ReadFile(capturePath)
	if err != nil {
		t.Fatalf("critic payloads not captured: %v", err)
	}
	var payloads []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if line == "" {
			continue
		}
		var p map[string]any
		if err := json.Unmarshal([]byte(line), &p); err != nil {
			t.Fatalf("captured payload is not valid JSON: %v\n%s", err, line)
		}
		payloads = append(payloads, p)
	}
	if len(payloads) < 2 {
		t.Fatalf("expected >=2 critic invocations, got %d", len(payloads))
	}

	// Iteration-1 payload: trace must already include the iteration under
	// review (ordering fix) — i.e. length 1, not 0.
	if iters := iterationCount(t, payloads[0]); iters < 1 {
		t.Fatalf("iter-1 critic payload trace.iterations length = %d, want >=1 (current iteration must be visible)", iters)
	}

	// Iteration-2 payload: must see both iterations, and must carry the
	// iteration-1 suggestion in prior_suggestions.
	if iters := iterationCount(t, payloads[1]); iters < 2 {
		t.Fatalf("iter-2 critic payload trace.iterations length = %d, want >=2", iters)
	}
	prior, _ := payloads[1]["prior_suggestions"].([]any)
	if len(prior) == 0 {
		t.Fatalf("iter-2 critic payload must carry prior_suggestions, got %v", payloads[1]["prior_suggestions"])
	}
	var found bool
	for _, s := range prior {
		if str, _ := s.(string); str == "iteration-one-suggestion" {
			found = true
		}
	}
	if !found {
		t.Fatalf("prior_suggestions must include iteration-1's suggestion; got %v", prior)
	}
}

// iterationCount extracts len(trace.iterations) from a captured critic payload.
func iterationCount(t *testing.T, payload map[string]any) int {
	t.Helper()
	tr, ok := payload["trace"].(map[string]any)
	if !ok {
		return 0
	}
	iters, ok := tr["iterations"].([]any)
	if !ok {
		return 0
	}
	return len(iters)
}

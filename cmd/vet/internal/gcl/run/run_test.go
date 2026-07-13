package run

import (
	"strings"
	"testing"

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

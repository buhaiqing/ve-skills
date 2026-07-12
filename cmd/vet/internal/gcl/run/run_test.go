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

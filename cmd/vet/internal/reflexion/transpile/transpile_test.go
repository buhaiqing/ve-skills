package transpile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTranspile_BelowThreshold(t *testing.T) {
	p := FailurePattern{
		Skill:   "ve-ecs-ops",
		Pattern: "MissingParameter",
		Count:   4,
	}
	_, ok := Transpile(p)
	if ok {
		t.Error("expected false for count below threshold")
	}
}

func TestTranspile_AboveThreshold(t *testing.T) {
	p := FailurePattern{
		Skill:   "ve-ecs-ops",
		Pattern: "evidence_overfetch",
		Fix:     "cap at 15",
		Count:   10,
	}
	g, ok := Transpile(p)
	if !ok {
		t.Fatal("expected true for count >= 5")
	}
	if g.Skill != "ve-ecs-ops" {
		t.Errorf("Skill = %q, want %q", g.Skill, "ve-ecs-ops")
	}
	if g.Trigger != "evidence_overfetch" {
		t.Errorf("Trigger = %q, want %q", g.Trigger, "evidence_overfetch")
	}
	if g.Severity != "medium" {
		t.Errorf("Severity = %q, want %q", g.Severity, "medium")
	}
	if g.Action != "auto-ASK" {
		t.Errorf("Action = %q, want %q", g.Action, "auto-ASK")
	}
	if g.SourceCount != 10 {
		t.Errorf("SourceCount = %d, want %d", g.SourceCount, 10)
	}
	if g.ID == "" {
		t.Error("ID should not be empty")
	}
}

func TestTranspile_ExactlyThreshold(t *testing.T) {
	p := FailurePattern{
		Skill:   "ve-redis-ops",
		Pattern: "DeleteDBInstance",
		Count:   5,
	}
	_, ok := Transpile(p)
	if !ok {
		t.Error("expected true for count exactly at threshold")
	}
}

func TestTranspile_HardLevel(t *testing.T) {
	p := FailurePattern{
		Skill:   "ve-ecs-ops",
		Pattern: "retry_storm",
		Count:   20,
	}
	g, ok := Transpile(p)
	if !ok {
		t.Fatal("expected true for count >= 15")
	}
	if g.Severity != "high" {
		t.Errorf("Severity = %q, want %q", g.Severity, "high")
	}
	if g.Action != "auto-REFUSE" {
		t.Errorf("Action = %q, want %q", g.Action, "auto-REFUSE")
	}
}

func TestTranspile_IDStable(t *testing.T) {
	p := FailurePattern{
		Skill:   "ve-ecs-ops",
		Pattern: "evidence_overfetch",
		Count:   15,
	}
	g1, _ := Transpile(p)
	g2, _ := Transpile(p)
	if g1.ID != g2.ID {
		t.Errorf("IDs differ: %q vs %q", g1.ID, g2.ID)
	}
}

func TestTranspileFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	inPath := filepath.Join(dir, "empty.md")
	outPath := filepath.Join(dir, "out.yaml")

	content := `# No tables here
Just some text.
`
	if err := os.WriteFile(inPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	n, err := TranspileFile(inPath, outPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 guardrails, got %d", n)
	}
}

func TestTranspileFile_WithPatterns(t *testing.T) {
	dir := t.TempDir()
	inPath := filepath.Join(dir, "patterns.md")
	outPath := filepath.Join(dir, "guardrails.yaml")

	content := `# Failure Patterns

## Incident Response

| Scenario | Failure Pattern | Root Cause | Fix | Count |
|----------|-----------------|------------|-----|-------|
| Alarm triage | routing_mismatch | wrong skill | verify graph | 10 |
| Diagnosis | evidence_overfetch | >20 calls | cap at 15 | 20 |
| Reflexion | pattern_undercount | forgotten | scan every run | 5 |
| Alarm triage | correlated_alarm_missed | only loudest | run correlation | 3 |
`
	if err := os.WriteFile(inPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	n, err := TranspileFile(inPath, outPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 3 {
		t.Errorf("expected 3 guardrails (counts 10,20,5), got %d", n)
	}

	// Verify output YAML
	out, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	outStr := string(out)
	if !strings.Contains(outStr, "guardrails:") {
		t.Error("output should contain 'guardrails:'")
	}
	if !strings.Contains(outStr, "routing_mismatch") {
		t.Error("output should contain 'routing_mismatch'")
	}
	if !strings.Contains(outStr, "evidence_overfetch") {
		t.Error("output should contain 'evidence_overfetch'")
	}
	if !strings.Contains(outStr, "auto-REFUSE") {
		t.Error("output should contain 'auto-REFUSE' for count>=15")
	}
	if strings.Contains(outStr, "correlated_alarm_missed") {
		t.Error("output should NOT contain 'correlated_alarm_missed' (count=5 < threshold)")
	}
}

func TestTranspileFile_MissingFile(t *testing.T) {
	_, err := TranspileFile("/nonexistent/path.md", "/tmp/out.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

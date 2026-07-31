package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestComputeValueMTTA(t *testing.T) {
	alerted := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	started := alerted.Add(5 * time.Second)
	resolved := started.Add(10 * time.Second)

	m := ComputeValue(ValueInput{
		RunID:      "r1",
		Success:    true,
		AlertedAt:  alerted,
		StartedAt:  started,
		ResolvedAt: resolved,
	})
	if m.MTTAMs != 5000 {
		t.Fatalf("MTTAMs: got %d want 5000", m.MTTAMs)
	}
}

func TestComputeValueMTTR(t *testing.T) {
	alerted := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	started := alerted.Add(2 * time.Second)
	resolved := alerted.Add(45 * time.Second)

	m := ComputeValue(ValueInput{
		RunID:      "r2",
		Success:    true,
		AlertedAt:  alerted,
		StartedAt:  started,
		ResolvedAt: resolved,
	})
	if m.MTTRMs != 45000 {
		t.Fatalf("MTTRMs: got %d want 45000", m.MTTRMs)
	}
	if m.AgentDurationMs != 43000 {
		t.Fatalf("AgentDurationMs: got %d want 43000", m.AgentDurationMs)
	}

	// Failure still records MTTR elapsed.
	fail := ComputeValue(ValueInput{
		RunID:      "r2f",
		Success:    false,
		AlertedAt:  alerted,
		StartedAt:  started,
		ResolvedAt: resolved,
	})
	if fail.MTTRMs != 45000 {
		t.Fatalf("failed MTTRMs: got %d want 45000", fail.MTTRMs)
	}
}

func TestComputeValueLabor(t *testing.T) {
	alerted := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	started := alerted
	// Agent ran 2 minutes → labor = 30 - 2 = 28
	resolved := started.Add(2 * time.Minute)

	m := ComputeValue(ValueInput{
		RunID:      "r3",
		Success:    true,
		AlertedAt:  alerted,
		StartedAt:  started,
		ResolvedAt: resolved,
	})
	if m.BaselineManualMin != DefaultBaselineManualMin {
		t.Fatalf("baseline: got %v want %v", m.BaselineManualMin, DefaultBaselineManualMin)
	}
	if m.LaborMinutesSaved != 28.0 {
		t.Fatalf("LaborMinutesSaved: got %v want 28", m.LaborMinutesSaved)
	}

	fail := ComputeValue(ValueInput{
		RunID:      "r3f",
		Success:    false,
		AlertedAt:  alerted,
		StartedAt:  started,
		ResolvedAt: resolved,
	})
	if fail.LaborMinutesSaved != 0 {
		t.Fatalf("failed LaborMinutesSaved: got %v want 0", fail.LaborMinutesSaved)
	}

	// Agent longer than baseline → labor clamped to 0
	long := ComputeValue(ValueInput{
		RunID:      "r3l",
		Success:    true,
		AlertedAt:  alerted,
		StartedAt:  started,
		ResolvedAt: started.Add(60 * time.Minute),
	})
	if long.LaborMinutesSaved != 0 {
		t.Fatalf("long LaborMinutesSaved: got %v want 0", long.LaborMinutesSaved)
	}
}

func TestPersistValueRoundtrip(t *testing.T) {
	root := t.TempDir()
	m := ComputeValue(ValueInput{
		RunID:      "persist-run",
		TicketID:   "T-1",
		Success:    true,
		AlertedAt:  time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC),
		StartedAt:  time.Date(2026, 8, 1, 10, 0, 5, 0, time.UTC),
		ResolvedAt: time.Date(2026, 8, 1, 10, 1, 0, 0, time.UTC),
	})
	if err := PersistValue(root, m); err != nil {
		t.Fatalf("PersistValue: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, ".runtime", "agent", "runs", "persist-run", "value.json"))
	if err != nil {
		t.Fatalf("read value.json: %v", err)
	}
	var got ValueMetrics
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.RunID != m.RunID || got.MTTAMs != m.MTTAMs || got.MTTRMs != m.MTTRMs {
		t.Fatalf("roundtrip mismatch: got %+v want %+v", got, m)
	}

	jsonl, err := os.ReadFile(filepath.Join(root, "audit-results", "value-metrics.jsonl"))
	if err != nil {
		t.Fatalf("read jsonl: %v", err)
	}
	if !strings.Contains(string(jsonl), `"run_id":"persist-run"`) {
		t.Fatalf("jsonl missing run_id: %s", jsonl)
	}
}

func TestFileTicketWriter(t *testing.T) {
	dir := t.TempDir()
	w := FileTicketWriter{Dir: dir}
	body := FormatValueComment(ValueMetrics{RunID: "r", Success: true, MTTAMs: 100})
	if err := w.WriteValueComment("TICKET-9", body); err != nil {
		t.Fatalf("WriteValueComment: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "TICKET-9.md"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(got), "MTTA: 100 ms") {
		t.Fatalf("unexpected body: %s", got)
	}
}

func TestConfirmStubHealForcesASK(t *testing.T) {
	root := t.TempDir()
	policyDir := filepath.Join(root, "incident-loop-agent", "references", "policies")
	if err := os.MkdirAll(policyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	mustWrite := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(policyDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite("execution-risk.md", `# Execution Risk Policy

## 2. Decision matrix

| risk | blast_radius | decision |
|------|--------------|----------|
| read-only | single | AUTO |
| destructive | single | ASK |
`)
	mustWrite("domain-allowlist.md", "# Domain Allow-list\n\n## 1. Eligible skills\n\n`ve-ecs-ops`\n")

	plan := &DispatchPlan{
		Operations: []DispatchOp{
			{
				Skill:       "ve-ecs-ops",
				Command:     "ve ecs DescribeInstances",
				SafetyClass: "read_only",
				BlastRadius: "single",
				Confidence:  "high",
				Safety:      1.0,
			},
		},
		BlastRadius:      "single",
		HealIncidentType: "cpu_high",
	}

	res := Confirm(root, plan)
	if res.Decision != "ASK" {
		t.Fatalf("Decision: got %s want ASK (reason=%s)", res.Decision, res.Reason)
	}
	if !strings.Contains(strings.ToLower(res.Reason), "stub") {
		t.Fatalf("Reason should mention stub: %s", res.Reason)
	}

	// Without HealIncidentType, same ops → AUTO
	plan.HealIncidentType = ""
	res2 := Confirm(root, plan)
	if res2.Decision != "AUTO" {
		t.Fatalf("no heal type: got %s want AUTO (reason=%s)", res2.Decision, res2.Reason)
	}
}

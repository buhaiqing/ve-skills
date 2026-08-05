package trace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeIncidentTraceRaw writes a single incident-trace-*.json fixture from a raw body.
func writeIncidentTraceRaw(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatalf("write fixture %s: %v", name, err)
	}
}

func TestAggregateIncident(t *testing.T) {
	root := t.TempDir()
	audit := filepath.Join(root, "audit-results")
	if err := os.MkdirAll(audit, 0o755); err != nil {
		t.Fatalf("mkdir audit-results: %v", err)
	}

	// Two tickets, each with ve_calls and request_ids + a policy_decision.
	t1 := `{"ticket_id":"T1","policy_decision":"AUTO","iterations":[{"ve_calls":[{"request_id":"r1a"},{"request_id":"r1b"},{"request_id":""}]}]}`
	t2 := `{"ticket_id":"T2","policy_decision":"ASK","iterations":[{"ve_calls":[{"request_id":"r2a"}]},{"ve_calls":[{"request_id":"r2b"},{"request_id":"r2c"}]}]}`
	writeIncidentTraceRaw(t, audit, "incident-trace-1.json", t1)
	writeIncidentTraceRaw(t, audit, "incident-trace-2.json", t2)

	s := AggregateIncident(root)

	if got := len(s.ByTicket); got != 2 {
		t.Fatalf("ByTicket len = %d, want 2", got)
	}

	t1stat, ok := s.ByTicket["T1"]
	if !ok {
		t.Fatal("missing ticket T1")
	}
	if t1stat.VeCalls != 3 {
		t.Errorf("T1 ve_calls = %d, want 3", t1stat.VeCalls)
	}
	if t1stat.RequestIDs != 2 {
		t.Errorf("T1 request_ids = %d, want 2", t1stat.RequestIDs)
	}
	if t1stat.PolicyDecision != "AUTO" {
		t.Errorf("T1 policy_decision = %q, want AUTO", t1stat.PolicyDecision)
	}

	t2stat, ok := s.ByTicket["T2"]
	if !ok {
		t.Fatal("missing ticket T2")
	}
	if t2stat.VeCalls != 3 {
		t.Errorf("T2 ve_calls = %d, want 3", t2stat.VeCalls)
	}
	if t2stat.RequestIDs != 3 {
		t.Errorf("T2 request_ids = %d, want 3", t2stat.RequestIDs)
	}
	if t2stat.PolicyDecision != "ASK" {
		t.Errorf("T2 policy_decision = %q, want ASK", t2stat.PolicyDecision)
	}

	if s.Totals["tickets"] != 2 {
		t.Errorf("totals.tickets = %d, want 2", s.Totals["tickets"])
	}
	if s.Totals["ve_calls"] != 6 {
		t.Errorf("totals.ve_calls = %d, want 6", s.Totals["ve_calls"])
	}
	if s.Totals["request_ids_covered"] != 5 {
		t.Errorf("totals.request_ids_covered = %d, want 5", s.Totals["request_ids_covered"])
	}

	if s.PolicyDecision["AUTO"] != 1 || s.PolicyDecision["ASK"] != 1 {
		t.Errorf("policy_decision distribution = %v, want AUTO:1 ASK:1", s.PolicyDecision)
	}
}

func TestPersistIncident(t *testing.T) {
	root := t.TempDir()
	s := &IncidentSummary{
		TraceSchemaVersion: "v1-incident",
		Totals:             map[string]int{"tickets": 1, "ve_calls": 2, "request_ids_covered": 2},
		ByTicket: map[string]TicketStat{
			"T1": {VeCalls: 2, RequestIDs: 2, PolicyDecision: "AUTO"},
		},
		PolicyDecision: map[string]int{"AUTO": 1},
	}
	path, err := PersistIncident(root, s)
	if err != nil {
		t.Fatalf("PersistIncident: %v", err)
	}
	if !strings.Contains(filepath.Base(path), "incident-summary-") {
		t.Fatalf("output path %q missing incident-summary- prefix", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read persisted: %v", err)
	}
	var got IncidentSummary
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal persisted: %v", err)
	}
	if got.Totals["ve_calls"] != 2 {
		t.Errorf("round-trip totals.ve_calls = %d, want 2", got.Totals["ve_calls"])
	}
	if got.ByTicket["T1"].RequestIDs != 2 {
		t.Errorf("round-trip T1 request_ids = %d, want 2", got.ByTicket["T1"].RequestIDs)
	}
}

func TestCmdIncidentNoAuditDir(t *testing.T) {
	root := t.TempDir() // no audit-results dir
	code := CmdIncident(root)
	if code != 0 {
		t.Fatalf("CmdIncident exit code = %d, want 0", code)
	}
	// When there is no audit-results dir, the incident_path must be empty.
	matches, _ := filepath.Glob(filepath.Join(root, "audit-results", "incident-summary-*.json"))
	if len(matches) != 0 {
		t.Errorf("expected no incident-summary file, got %v", matches)
	}
}

func TestIterationRollbackApplied(t *testing.T) {
	it := Iteration{Iter: 1, Timestamp: "2026-08-06T00:00:00Z", RollbackApplied: true}
	b, err := json.Marshal(it)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(b), "\"rollback_applied\":true") {
		t.Errorf("expected rollback_applied=true in JSON, got %s", string(b))
	}
	var got Iteration
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !got.RollbackApplied {
		t.Error("round-trip RollbackApplied = false, want true")
	}
}

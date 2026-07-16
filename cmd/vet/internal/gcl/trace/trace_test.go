package trace

import (
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func sampleTrace(skill, status string, scores map[string]float64) *Trace {
	return &Trace{
		TraceSchemaVersion: "v1",
		Skill:              skill,
		Request:            "req",
		RubricVersion:      "v1",
		OperationIntent:    map[string]any{"operation": "describe"},
		MaskedFields:       []string{},
		RedactionPass:      true,
		Iterations: []Iteration{
			{
				Iter:      1,
				Generator: GeneratorResult{Command: "ve x", ExitCode: 0, ResultExcerpt: "ok"},
				Critic:    CriticRecord{Scores: scores, Suggestions: []string{}, Blocking: false},
				Decision:  "PASS",
			},
		},
		Final: Final{Status: status, Iter: 1, Output: ptr("ok")},
	}
}

func ptr(s string) *string { return &s }

func TestAggregate(t *testing.T) {
	traces := []*Trace{
		sampleTrace("ve-ecs-ops", "PASS", map[string]float64{"correctness": 1, "safety": 1, "idempotency": 0.5, "traceability": 1, "spec_compliance": 1}),
		sampleTrace("ve-ecs-ops", "PASS", map[string]float64{"correctness": 1, "safety": 1, "idempotency": 0.5, "traceability": 1, "spec_compliance": 1}),
		sampleTrace("ve-redis-ops", "SAFETY_FAIL", map[string]float64{"correctness": 1, "safety": 0, "idempotency": 0.5, "traceability": 1, "spec_compliance": 1}),
	}
	sum := Aggregate("/tmp", traces)
	if sum.Totals["total_runs"] != 3 {
		t.Errorf("total_runs want 3 got %d", sum.Totals["total_runs"])
	}
	if sum.Totals["PASS"] != 2 {
		t.Errorf("PASS want 2 got %d", sum.Totals["PASS"])
	}
	if sum.Totals["SAFETY_FAIL"] != 1 {
		t.Errorf("SAFETY_FAIL want 1 got %d", sum.Totals["SAFETY_FAIL"])
	}
	if sum.PassRate < 0.6666 || sum.PassRate > 0.6668 {
		t.Errorf("pass_rate want ~%v got %v", 2.0/3.0, sum.PassRate)
	}
}

func TestCollectPaths(t *testing.T) {
	dir := t.TempDir()
	writeTrace(t, filepath.Join(dir, "audit-results", "gcl-trace-20260101-000000.json"))
	paths := CollectPaths(dir, nil, nil)
	if len(paths) != 1 {
		t.Fatalf("expected 1 path, got %d", len(paths))
	}
}

func writeTrace(t *testing.T, p string) {
	t.Helper()
	if err := writeFile(p, "{\"skill\":\"ve-ecs-ops\",\"final\":{\"status\":\"PASS\"}}"); err != nil {
		t.Fatal(err)
	}
}

// writeRuntimeTrace serializes a runtime Trace (gcl-trace-*.json shape) to a
// temp file for Check tests.
func writeRuntimeTrace(t *testing.T, dir string, tr *Trace) string {
	t.Helper()
	b, err := json.MarshalIndent(tr, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "gcl-trace-"+strconv.Itoa(len(t.Name()))+".json")
	if err := writeFile(p, string(b)); err != nil {
		t.Fatal(err)
	}
	return p
}

// A runtime trace that actually ran a `ve` call but forgot request_id.
func TestCheck_RuntimeMissingRequestID(t *testing.T) {
	dir := t.TempDir()
	tr := sampleTrace("ve-ecs-ops", "PASS", map[string]float64{"correctness": 1, "safety": 1})
	tr.Iterations[0].RequestID = "" // ve call ran, but no RequestId recorded
	p := writeRuntimeTrace(t, dir, tr)
	if err := Check(p); err == nil {
		t.Fatal("expected missing request_id to fail, got nil")
	} else if !strings.Contains(err.Error(), "request_id") {
		t.Errorf("expected error about request_id, got: %v", err)
	}
}

// A runtime trace with request_id populated + redaction_pass passes.
func TestCheck_RuntimeValid(t *testing.T) {
	dir := t.TempDir()
	tr := sampleTrace("ve-ecs-ops", "PASS", map[string]float64{"correctness": 1, "safety": 1})
	tr.Iterations[0].RequestID = "req-abc-123"
	p := writeRuntimeTrace(t, dir, tr)
	if err := Check(p); err != nil {
		t.Fatalf("valid runtime trace should pass, got: %v", err)
	}
}

// A POLICY_BLOCK first iteration never ran a `ve` call -> empty request_id is OK.
func TestCheck_RuntimePolicyBlockSkipsRequestID(t *testing.T) {
	dir := t.TempDir()
	tr := sampleTrace("ve-ecs-ops", "POLICY_BLOCK", map[string]float64{"correctness": 1, "safety": 1})
	tr.Iterations[0].Decision = "POLICY_BLOCK"
	tr.Iterations[0].RequestID = "" // no ve call ran
	p := writeRuntimeTrace(t, dir, tr)
	if err := Check(p); err != nil {
		t.Fatalf("POLICY_BLOCK iteration needs no request_id, got: %v", err)
	}
}

// A file without trace_schema_version is not a runtime trace -> skipped (nil).
func TestCheck_RuntimeSkipsNonRuntimeTrace(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "incident-trace-X.json")
	if err := writeFile(p, `{"ticket_id":"T1","started_at":"s","finished_at":"f","policy_decision":"AUTO","iterations":[],"redaction_pass":true}`); err != nil {
		t.Fatal(err)
	}
	if err := Check(p); err != nil {
		t.Fatalf("non-runtime trace should be skipped, got: %v", err)
	}
}

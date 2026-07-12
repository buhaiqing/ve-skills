package trace

import (
	"path/filepath"
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

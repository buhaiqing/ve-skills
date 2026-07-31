package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAggregateValueMetricsP50AndAutoRate(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "audit-results")
	_ = os.MkdirAll(dir, 0o755)
	lines := []string{
		`{"run_id":"1","success":true,"policy_decision":"AUTO","mtta_ms":100,"mttr_ms":200,"labor_minutes_saved":10}`,
		`{"run_id":"2","success":true,"policy_decision":"ASK","mtta_ms":300,"mttr_ms":400,"labor_minutes_saved":5}`,
		`{"run_id":"3","success":false,"policy_decision":"AUTO","mtta_ms":500,"mttr_ms":600,"labor_minutes_saved":0}`,
	}
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(l + "\n")
	}
	_ = os.WriteFile(filepath.Join(dir, "value-metrics.jsonl"), []byte(b.String()), 0o644)
	d := AggregateValueMetrics(root)
	if d.Samples != 3 {
		t.Fatalf("samples=%d", d.Samples)
	}
	if d.P50MTTAMs != 300 {
		t.Fatalf("p50 mtta=%d", d.P50MTTAMs)
	}
	if d.P50MTTRMs != 400 {
		t.Fatalf("p50 mttr=%d", d.P50MTTRMs)
	}
	if d.AutoRate < 0.66 || d.AutoRate > 0.67 {
		t.Fatalf("auto_rate=%v", d.AutoRate)
	}
	if d.LaborMinutesSum != 15 {
		t.Fatalf("labor=%v", d.LaborMinutesSum)
	}
}

func TestAggregateValueMetricsMissingFile(t *testing.T) {
	root := t.TempDir()
	d := AggregateValueMetrics(root)
	if d.Samples != 0 || d.P50MTTAMs != 0 || d.AutoRate != 0 {
		t.Fatalf("expected zero dashboard, got %+v", d)
	}
}

func TestAggregateValueMetricsSkipsBadJSONLLines(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "audit-results")
	_ = os.MkdirAll(dir, 0o755)
	body := strings.Join([]string{
		`{"run_id":"1","success":true,"policy_decision":"AUTO","mtta_ms":100,"mttr_ms":200,"labor_minutes_saved":1}`,
		`{bad json`,
		`{"run_id":"2","success":true,"policy_decision":"ASK","mtta_ms":300,"mttr_ms":400,"labor_minutes_saved":2}`,
	}, "\n") + "\n"
	_ = os.WriteFile(filepath.Join(dir, "value-metrics.jsonl"), []byte(body), 0o644)
	d := AggregateValueMetrics(root)
	if d.Samples != 2 {
		t.Fatalf("samples=%d want 2", d.Samples)
	}
	if d.LaborMinutesSum != 3 {
		t.Fatalf("labor=%v", d.LaborMinutesSum)
	}
}

func TestAggregateValueMetricsFallsBackToRunValueJSON(t *testing.T) {
	root := t.TempDir()
	runDir := filepath.Join(root, ".runtime", "agent", "runs", "r1")
	_ = os.MkdirAll(runDir, 0o755)
	_ = os.WriteFile(filepath.Join(runDir, "value.json"), []byte(
		`{"run_id":"r1","success":true,"policy_decision":"AUTO","mtta_ms":50,"mttr_ms":80,"labor_minutes_saved":7}`,
	), 0o644)
	d := AggregateValueMetrics(root)
	if d.Samples != 1 || d.P50MTTAMs != 50 || d.LaborMinutesSum != 7 || d.AutoRate != 1 {
		t.Fatalf("unexpected fallback dashboard: %+v", d)
	}
}

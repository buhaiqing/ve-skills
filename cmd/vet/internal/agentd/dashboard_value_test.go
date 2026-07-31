package agentd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/agent"
)

func TestRenderDashboardValueKPIs(t *testing.T) {
	stats := &DashboardStats{
		Value: agent.ValueDashboard{
			P50MTTAMs:       300,
			P50MTTRMs:       400,
			LaborMinutesSum: 15,
			AutoRate:        0.67,
			Samples:         3,
		},
	}

	got := renderDashboard(stats)
	for _, want := range []string{"Value KPIs", "P50 MTTA", "P50 MTTR", "Labor Saved", "AUTO %", "300ms", "15"} {
		if !strings.Contains(got, want) {
			t.Errorf("dashboard missing %q", want)
		}
	}
}

func TestAggregateStatsIncludesValueMetrics(t *testing.T) {
	server, tmpDir := setupTestServer(t)

	dir := filepath.Join(tmpDir, "audit-results")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	line := `{"run_id":"1","success":true,"policy_decision":"AUTO","mtta_ms":100,"mttr_ms":200,"labor_minutes_saved":10}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, "value-metrics.jsonl"), []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}

	stats := server.aggregateStats()
	if stats.Value.Samples != 1 {
		t.Fatalf("value samples=%d", stats.Value.Samples)
	}
	if stats.Value.P50MTTAMs != 100 {
		t.Fatalf("p50 mtta=%d", stats.Value.P50MTTAMs)
	}
	if stats.Value.AutoRate != 1.0 {
		t.Fatalf("auto_rate=%v", stats.Value.AutoRate)
	}
}

func TestAggregateStatsValueMetricsNoFile(t *testing.T) {
	server, _ := setupTestServer(t)
	stats := server.aggregateStats()
	if stats.Value.Samples != 0 {
		t.Fatalf("expected zero value metrics, got samples=%d", stats.Value.Samples)
	}
}

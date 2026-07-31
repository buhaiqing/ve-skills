package agent

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

// ValueDashboard holds aggregated value KPIs across runs.
type ValueDashboard struct {
	P50MTTAMs       int64
	P50MTTRMs       int64
	LaborMinutesSum float64
	AutoRate        float64
	Samples         int
}

// AggregateValueMetrics prefers audit-results/value-metrics.jsonl; if that
// yields no samples, falls back to .runtime/agent/runs/*/value.json.
// Missing sources or scanner errors return a zero ValueDashboard.
func AggregateValueMetrics(root string) ValueDashboard {
	metrics := loadValueMetricsJSONL(filepath.Join(root, "audit-results", "value-metrics.jsonl"))
	if len(metrics) == 0 {
		metrics = loadValueMetricsFromRuns(root)
	}
	return dashboardFromMetrics(metrics)
}

func loadValueMetricsJSONL(path string) []ValueMetrics {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var out []ValueMetrics
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var m ValueMetrics
		if err := json.Unmarshal(line, &m); err != nil {
			continue
		}
		out = append(out, m)
	}
	if err := sc.Err(); err != nil {
		return nil
	}
	return out
}

func loadValueMetricsFromRuns(root string) []ValueMetrics {
	runDir := filepath.Join(root, ".runtime", "agent", "runs")
	entries, err := os.ReadDir(runDir)
	if err != nil {
		return nil
	}
	var out []ValueMetrics
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(runDir, e.Name(), "value.json"))
		if err != nil {
			continue
		}
		var m ValueMetrics
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}
		out = append(out, m)
	}
	return out
}

func dashboardFromMetrics(metrics []ValueMetrics) ValueDashboard {
	if len(metrics) == 0 {
		return ValueDashboard{}
	}
	mttas := make([]int64, 0, len(metrics))
	mttrs := make([]int64, 0, len(metrics))
	var laborSum float64
	autoCnt := 0
	for _, m := range metrics {
		mttas = append(mttas, m.MTTAMs)
		mttrs = append(mttrs, m.MTTRMs)
		laborSum += m.LaborMinutesSaved
		if m.PolicyDecision == "AUTO" {
			autoCnt++
		}
	}
	n := len(metrics)
	return ValueDashboard{
		P50MTTAMs:       p50Int64(mttas),
		P50MTTRMs:       p50Int64(mttrs),
		LaborMinutesSum: laborSum,
		AutoRate:        float64(autoCnt) / float64(n),
		Samples:         n,
	}
}

func p50Int64(vals []int64) int64 {
	if len(vals) == 0 {
		return 0
	}
	cp := append([]int64(nil), vals...)
	sort.Slice(cp, func(i, j int) bool { return cp[i] < cp[j] })
	n := len(cp)
	if n%2 == 1 {
		return cp[n/2]
	}
	return (cp[n/2-1] + cp[n/2]) / 2
}

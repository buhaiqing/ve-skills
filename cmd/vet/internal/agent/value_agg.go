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

// AggregateValueMetrics reads audit-results/value-metrics.jsonl and computes KPIs.
// Missing file returns zero ValueDashboard.
func AggregateValueMetrics(root string) ValueDashboard {
	path := filepath.Join(root, "audit-results", "value-metrics.jsonl")
	f, err := os.Open(path)
	if err != nil {
		return ValueDashboard{}
	}
	defer f.Close()

	var (
		mttas    []int64
		mttrs    []int64
		laborSum float64
		autoCnt  int
		samples  int
	)

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
		samples++
		mttas = append(mttas, m.MTTAMs)
		mttrs = append(mttrs, m.MTTRMs)
		laborSum += m.LaborMinutesSaved
		if m.PolicyDecision == "AUTO" {
			autoCnt++
		}
	}

	if samples == 0 {
		return ValueDashboard{}
	}

	var autoRate float64
	if samples > 0 {
		autoRate = float64(autoCnt) / float64(samples)
	}

	return ValueDashboard{
		P50MTTAMs:       p50Int64(mttas),
		P50MTTRMs:       p50Int64(mttrs),
		LaborMinutesSum: laborSum,
		AutoRate:        autoRate,
		Samples:         samples,
	}
}

func p50Int64(vals []int64) int64 {
	if len(vals) == 0 {
		return 0
	}
	sort.Slice(vals, func(i, j int) bool { return vals[i] < vals[j] })
	n := len(vals)
	if n%2 == 1 {
		return vals[n/2]
	}
	return (vals[n/2-1] + vals[n/2]) / 2
}

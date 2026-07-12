package trace

import (
	"path/filepath"
	"sort"
	"time"
)

// Summary is the aggregated quality summary (mirrors gcl_trace_aggregate.aggregate).
type Summary struct {
	Version          string                    `json:"version"`
	GeneratedAt      string                    `json:"generated_at"`
	Window           map[string]int           `json:"window"`
	Totals           map[string]int           `json:"totals"`
	PassRate         float64                   `json:"pass_rate"`
	AvgRubricScores  map[string]any           `json:"avg_rubric_scores"`
	FailurePatterns  []map[string]any         `json:"failure_patterns"`
	BySkill          map[string]map[string]any `json:"by_skill"`
	TraceFiles       []string                  `json:"trace_files"`
}

// Aggregate computes the quality summary from a set of parsed traces.
func Aggregate(root string, traces []*Trace) *Summary {
	totals := map[string]int{"PASS": 0, "SAFETY_FAIL": 0, "MAX_ITER": 0, "total_runs": len(traces)}
	scoreSums := map[string]float64{}
	scoreCount := 0
	bySkill := map[string]map[string]any{}
	var failurePatterns []map[string]any
	var traceFiles []string

	for _, t := range traces {
		status := t.Final.Status
		if _, ok := totals[status]; ok {
			totals[status]++
		}
		bucket := bySkill[t.Skill]
		if bucket == nil {
			bucket = map[string]any{"total": 0, "PASS": 0, "SAFETY_FAIL": 0, "MAX_ITER": 0, "avg_iterations": 0.0}
			bySkill[t.Skill] = bucket
		}
		bucket["total"] = toInt(bucket["total"]) + 1
		if _, ok := bucket[status]; ok {
			bucket[status] = toInt(bucket[status]) + 1
		}
		iters := len(t.Iterations)
		prevAvg := toFloat(bucket["avg_iterations"])
		total := toInt(bucket["total"])
		bucket["avg_iterations"] = (prevAvg*float64(total-1) + float64(iters)) / float64(total)

		scores := t.LastScores()
		if len(scores) > 0 {
			scoreCount++
			for _, d := range RubricDims {
				scoreSums[d] += scores[d]
			}
		}
		if t.Final.FailurePattern != nil {
			fp := t.Final.FailurePattern
			failurePatterns = append(failurePatterns, map[string]any{
				"skill":    t.Skill,
				"pattern":  fp.Error,
				"category": fp.Category,
				"source":   t.SourcePath,
			})
		}
		if t.SourcePath != "" {
			traceFiles = append(traceFiles, t.SourcePath)
		}
	}

	passRate := 0.0
	if totals["total_runs"] > 0 {
		passRate = float64(totals["PASS"]) / float64(totals["total_runs"])
	}
	avgScores := map[string]any{}
	for _, d := range RubricDims {
		if scoreCount > 0 {
			avgScores[d] = round3(scoreSums[d] / float64(scoreCount))
		} else {
			avgScores[d] = nil
		}
	}
	_ = root
	return &Summary{
		Version:         "1.1",
		GeneratedAt:     time.Now().UTC().Format(time.RFC3339),
		Window:          map[string]int{"trace_count": totals["total_runs"]},
		Totals:          totals,
		PassRate:        round4(passRate),
		AvgRubricScores: avgScores,
		FailurePatterns: failurePatterns,
		BySkill:         bySkill,
		TraceFiles:      traceFiles,
	}
}

// UpdateFailurePatternsFile appends/refreshes the auto-generated failure-pattern
// block in docs/failure-patterns.md. Mirrors gcl_trace_aggregate.update_failure_patterns_file.
func UpdateFailurePatternsFile(root string, sum *Summary) (string, error) {
	patterns := sum.FailurePatterns
	if len(patterns) == 0 {
		return "", nil
	}
	lines := []string{"", "---", "", "## Extracted from GCL Traces (auto-generated)", "",
		"| Skill | Pattern | Category | Source |", "|-------|---------|----------|--------|"}
	seen := map[string]bool{}
	for _, p := range patterns {
		key := p["skill"].(string) + "|" + toString(p["pattern"])
		if seen[key] {
			continue
		}
		seen[key] = true
		lines = append(lines, "| `"+toString(p["skill"])+"` | `"+toString(p["pattern"])+"` | "+toString(p["category"])+" | `"+toString(p["source"])+"` |")
	}
	table := joinLines(lines)
	fpPath := filepath.Join(root, "docs", "failure-patterns.md")
	existing := ""
	if b, err := readFileIfExists(fpPath); err == nil {
		existing = b
	}
	marker := "## Extracted from GCL Traces (auto-generated)"
	if contains(existing, marker) {
		existing = replaceBlock(existing, marker, table)
	} else {
		existing += table + "\n"
	}
	if err := writeFile(fpPath, existing); err != nil {
		return "", err
	}
	return fpPath, nil
}

// PersistSummary writes the quality summary JSON. Mirrors persist_summary.
func PersistSummary(root string, sum *Summary) (string, error) {
	outDir := filepath.Join(root, "audit-results")
	if err := mkdirAll(outDir); err != nil {
		return "", err
	}
	ts := time.Now().UTC().Format("20060102-150405")
	path := filepath.Join(outDir, "gcl-quality-summary-"+ts+".json")
	return path, writeJSON(path, sum)
}

func toInt(v any) int {
	switch x := v.(type) {
	case int:
		return x
	case float64:
		return int(x)
	}
	return 0
}

func toFloat(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	}
	return 0
}

func round3(f float64) float64 { return float64(int(f*1000+0.5)) / 1000 }
func round4(f float64) float64 { return float64(int(f*10000+0.5)) / 10000 }

// sortBySkill returns traces sorted by skill name for stable output.
func sortBySkill(traces []*Trace) {
	sort.Slice(traces, func(i, j int) bool { return traces[i].Skill < traces[j].Skill })
}

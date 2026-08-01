package trace

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/heal"
)

// Summary is the aggregated quality summary (mirrors gcl_trace_aggregate.aggregate).
type Summary struct {
	Version         string                    `json:"version"`
	GeneratedAt     string                    `json:"generated_at"`
	Window          Window                    `json:"window"`
	Totals          map[string]int            `json:"totals"`
	PassRate        float64                   `json:"pass_rate"`
	AvgRubricScores map[string]any            `json:"avg_rubric_scores"`
	FailurePatterns []map[string]any          `json:"failure_patterns"`
	BySkill         map[string]map[string]any `json:"by_skill"`
	TraceFiles      []string                  `json:"trace_files"`
	Heal            *HealSummary              `json:"heal,omitempty"`
}

// HealSummary surfaces L4 self-healing telemetry in the quality report. nil
// (omitted via omitempty) when no heal log exists, so legacy summaries are
// unchanged.
type HealSummary struct {
	SuccessRate          float64 `json:"success_rate"`
	AvgDurationMs        float64 `json:"avg_duration_ms"`
	UserInterventionRate float64 `json:"user_intervention_rate"`
	FallbackRate         float64 `json:"fallback_rate"`
	TotalEvents          int64   `json:"total_events"`
}

// Window describes the real time range the aggregated traces cover. Until is
// always the aggregation moment; Since is non-nil only when a --since window
// was requested, in which case it is the window's lower bound.
type Window struct {
	Since      *string `json:"since"`
	Until      string  `json:"until"`
	TraceCount int     `json:"trace_count"`
}

// Aggregate computes the quality summary from a set of parsed traces.
// sinceHours is the --since window in hours (nil for full scan); it is used to
// populate the real Window range rather than the legacy trace-count total.
func Aggregate(root string, traces []*Trace, sinceHours *int) *Summary {
	totals := map[string]int{"PASS": 0, "SAFETY_FAIL": 0, "MAX_ITER": 0, "total_runs": len(traces)}
	scoreSums := map[string]float64{}
	scoreCount := 0
	bySkill := map[string]map[string]any{}
	var failurePatterns []map[string]any
	var traceFiles []string

	for _, t := range traces {
		// Warn on suspicious data that may skew aggregate metrics
		if len(t.Iterations) == 0 {
			fmt.Fprintf(os.Stderr, "WARN: trace %s has 0 iterations, skipping\n", t.SourcePath)
			continue
		}
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
			// Skip smoke / structural-only traces — they emit deterministic fake
			// REFUSE blocks that would otherwise pollute the pattern store and
			// auto-promote REFUSE itself into a Guardrail on every CI run.
			if strings.HasPrefix(t.Request, "CI smoke test") ||
				strings.HasPrefix(t.Request, "CI gate smoke:") {
				continue
			}
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
	// Fold L4 self-healing telemetry into the report when the heal log exists.
	// heal.HealSummary is nil when the file is absent or holds no parseable
	// events, leaving all other summary fields untouched (backward compat).
	healSummary := aggregateHeal(root)

	win := Window{TraceCount: totals["total_runs"], Until: time.Now().UTC().Format(time.RFC3339)}
	if sinceHours != nil {
		cutoff := time.Now().UTC().Add(-time.Duration(*sinceHours) * time.Hour)
		s := cutoff.Format(time.RFC3339)
		win.Since = &s
	}
	return &Summary{
		Version:         "1.1",
		GeneratedAt:     time.Now().UTC().Format(time.RFC3339),
		Window:          win,
		Totals:          totals,
		PassRate:        round4(passRate),
		AvgRubricScores: avgScores,
		FailurePatterns: failurePatterns,
		BySkill:         bySkill,
		TraceFiles:      traceFiles,
		Heal:            healSummary,
	}
}

// aggregateHeal reads audit-results/ve-self-healing.log (relative to root) and
// rolls its events into a HealSummary. Returns nil when the file is missing or
// yields no parseable events, so existing summaries are unaffected.
func aggregateHeal(root string) *HealSummary {
	path := filepath.Join(root, "audit-results", "ve-self-healing.log")
	events, _, err := heal.ParseFile(path, time.Time{})
	if err != nil || len(events) == 0 {
		return nil
	}
	var m heal.Metrics
	for _, e := range events {
		m.Record(heal.HealEvent{
			ISO:        e.ISO,
			EventType:  e.EventType,
			ErrorCode:  e.ErrorCode,
			Action:     e.Action,
			Result:     e.Result,
			DurationMs: e.DurationMs,
		})
	}
	return &HealSummary{
		SuccessRate:          m.SuccessRate(),
		AvgDurationMs:        m.AvgDurationMs(),
		UserInterventionRate: m.UserInterventionRate(),
		FallbackRate:         m.FallbackRate(),
		TotalEvents:          m.TotalCount,
	}
}

// UpdateFailurePatternsFile appends/refreshes the auto-generated failure-pattern
// block in docs/failure-patterns.md. Mirrors gcl_trace_aggregate.update_failure_patterns_file.
func UpdateFailurePatternsFile(root string, sum *Summary) (string, error) {
	patterns := sum.FailurePatterns
	if len(patterns) == 0 {
		return "", nil
	}
	lines := []string{"## Extracted from GCL Traces (auto-generated)", "",
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
		existing = strings.TrimRight(existing, "\n") + "\n\n---\n\n" + table + "\n"
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

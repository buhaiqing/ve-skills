// Package cmd wires `vet gcl trace` — the aggregator CLI over the trace package.
//
// Faithful Go port of scripts/gcl_trace_aggregate.py main(): collects trace
// files, aggregates, persists the summary, and updates docs/failure-patterns.md.
package trace

import (
	"encoding/json"
	"fmt"
	"path/filepath"
)

// CmdAggregate runs the trace aggregation. Returns the process exit code.
func CmdAggregate(root string, inputs []string, sinceHours *int) int {
	paths := CollectPaths(root, inputs, sinceHours)
	if len(paths) == 0 {
		fmt.Fprintln(stderr, "No gcl-trace files found.")
		return 1
	}
	var traces []*Trace
	for _, p := range paths {
		t := ParseTrace(p)
		if t == nil {
			fmt.Fprintf(stderr, "WARN: skip %s\n", p)
			continue
		}
		if rel, err := filepath.Rel(root, p); err == nil {
			t.SourcePath = rel
		} else {
			t.SourcePath = p
		}
		traces = append(traces, t)
	}
	if len(traces) == 0 {
		fmt.Fprintln(stderr, "No valid traces parsed.")
		return 1
	}
	sortBySkill(traces)
	summary := Aggregate(root, traces, sinceHours)
	out, err := PersistSummary(root, summary)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: persist summary: %v\n", err)
		return 1
	}
	fpOut, err := UpdateFailurePatternsFile(root, summary)
	if err != nil {
		fmt.Fprintf(stderr, "WARN: failure-patterns update skipped: %v\n", err)
	} else if fpOut != "" {
		fmt.Fprintf(stderr, "INFO: failure-patterns updated: %d patterns written to %s\n", len(summary.FailurePatterns), fpOut)
	}
	result := map[string]any{
		"summary_path":                 out,
		"pass_rate":                    summary.PassRate,
		"total_runs":                   summary.Totals["total_runs"],
		"failure_patterns_extracted":   len(summary.FailurePatterns),
	}
	if fpOut != "" {
		result["failure_patterns_updated"] = fpOut
	}
	b, _ := json.Marshal(result)
	fmt.Println(string(b))
	return 0
}

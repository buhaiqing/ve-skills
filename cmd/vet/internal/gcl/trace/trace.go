// Package trace defines the GCL trace schema shared by gcl run (writer) and
// gcl trace (aggregator). Field names match gcl_runner.py / gcl_trace_aggregate.py
// output so existing audit-results/*.json remain readable.
package trace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// RubricDims mirrors gcl_trace_aggregate.RUBRIC_DIMS.
var RubricDims = []string{"correctness", "safety", "idempotency", "traceability", "spec_compliance"}

// FinalStatuses mirrors gcl_trace_aggregate.FINAL_STATUSES.
var FinalStatuses = []string{"PASS", "SAFETY_FAIL", "MAX_ITER"}

// GeneratorResult is the per-iteration generator record (masked).
type GeneratorResult struct {
	Command       string `json:"command"`
	ExitCode      int    `json:"exit_code"`
	ResultExcerpt string `json:"result_excerpt"`
	StdoutLen     int    `json:"stdout_len"`
	StderrLen     int    `json:"stderr_len"`
	StderrExcerpt string `json:"stderr_excerpt,omitempty"`
	Args          map[string]any `json:"args,omitempty"`
}

// CriticRecord is the per-iteration critic record.
type CriticRecord struct {
	Scores      map[string]float64 `json:"scores"`
	Suggestions []string           `json:"suggestions"`
	Blocking    bool               `json:"blocking"`
}

// Iteration is one GCL loop iteration.
type Iteration struct {
	Iter    int             `json:"iter"`
	Generator GeneratorResult `json:"generator"`
	Critic  CriticRecord    `json:"critic"`
	Decision string         `json:"decision"`
}

// FailurePattern mirrors the Reflexion failure-pattern schema.
type FailurePattern struct {
	Category string `json:"category"`
	Skill    string `json:"skill"`
	Command  string `json:"command,omitempty"`
	Error    string `json:"error"`
	Fix      string `json:"fix"`
	Count    int    `json:"count"`
	Reusable bool   `json:"reusable"`
}

// Final is the trace's terminal record.
type Final struct {
	Status        string         `json:"status"`
	Iter          int            `json:"iter"`
	Output        *string        `json:"output,omitempty"`
	Unresolved    []string       `json:"unresolved,omitempty"`
	FailurePattern *FailurePattern `json:"failure_pattern,omitempty"`
}

// Trace is the top-level GCL trace document.
type Trace struct {
	TraceSchemaVersion string         `json:"trace_schema_version"`
	Skill             string         `json:"skill"`
	Request           string         `json:"request"`
	RubricVersion     string         `json:"rubric_version"`
	OperationIntent   map[string]any `json:"operation_intent"`
	MaskedFields      []string       `json:"masked_fields"`
	RedactionPass     bool           `json:"redaction_pass"`
	Iterations        []Iteration    `json:"iterations"`
	Final             Final          `json:"final"`
	SourcePath        string         `json:"_source_path,omitempty"`
}

// ParseTrace reads and validates a trace file. Returns nil on parse error.
func ParseTrace(path string) *Trace {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var t Trace
	if err := json.Unmarshal(data, &t); err != nil {
		return nil
	}
	if t.Skill == "" || t.Final.Status == "" {
		return nil
	}
	return &t
}

// LastScores returns the rubric scores of the final iteration, if any.
func (t *Trace) LastScores() map[string]float64 {
	if len(t.Iterations) == 0 {
		return map[string]float64{}
	}
	return t.Iterations[len(t.Iterations)-1].Critic.Scores
}

// PersistTrace writes the trace to audit-results/gcl-trace-YYYYMMDD-HHMMSS.json
// and returns the path. Mirrors gcl_runner.persist_trace.
func PersistTrace(root, relPath string, t *Trace) (string, error) {
	outDir := filepath.Join(root, "audit-results")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}
	ts := time.Now().UTC().Format("20060102-150405")
	path := filepath.Join(outDir, "gcl-trace-"+ts+".json")
	b, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return "", err
	}
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// CollectPaths returns sorted trace paths under audit-results matching the
// given input globs (relative to root) or all gcl-trace-*.json within
// sinceHours. Mirrors gcl_trace_aggregate.collect_paths.
func CollectPaths(root string, inputs []string, sinceHours *int) []string {
	var paths []string
	if len(inputs) > 0 {
		for _, pat := range inputs {
			matches, _ := filepath.Glob(filepath.Join(root, pat))
			paths = append(paths, matches...)
		}
	} else {
		all, _ := filepath.Glob(filepath.Join(root, "audit-results", "gcl-trace-*.json"))
		paths = append(paths, all...)
	}
	sort.Strings(paths)
	if sinceHours == nil {
		return filterFiles(paths)
	}
	cutoff := time.Now().UTC().Add(-time.Duration(*sinceHours) * time.Hour)
	var filtered []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if info.ModTime().UTC().After(cutoff) {
			filtered = append(filtered, p)
		}
	}
	return filterFiles(filtered)
}

func filterFiles(paths []string) []string {
	var out []string
	for _, p := range paths {
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			out = append(out, p)
		}
	}
	return out
}

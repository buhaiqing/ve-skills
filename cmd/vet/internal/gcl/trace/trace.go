// Package trace defines the GCL trace schema shared by gcl run (writer) and
// gcl trace (aggregator). Field names match gcl_runner.py / gcl_trace_aggregate.py
// output so existing audit-results/*.json remain readable.
package trace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// RubricDims mirrors gcl_trace_aggregate.RUBRIC_DIMS.
var RubricDims = []string{"correctness", "safety", "idempotency", "traceability", "spec_compliance"}

// FinalStatuses mirrors gcl_trace_aggregate.FINAL_STATUSES.
var FinalStatuses = []string{"PASS", "SAFETY_FAIL", "MAX_ITER"}

// GeneratorResult is the per-iteration generator record (masked).
type GeneratorResult struct {
	Command       string         `json:"command"`
	ExitCode      int            `json:"exit_code"`
	ResultExcerpt string         `json:"result_excerpt"`
	StdoutLen     int            `json:"stdout_len"`
	StderrLen     int            `json:"stderr_len"`
	StderrExcerpt string         `json:"stderr_excerpt,omitempty"`
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
	Iter      int             `json:"iter"`
	Generator GeneratorResult `json:"generator"`
	Critic    CriticRecord    `json:"critic"`
	Decision  string          `json:"decision"`
	// PolicyDecision records the execution-risk verdict (AUTO/ASK/REFUSE)
	// from scoreDecision, applied before the generator command runs. Empty
	// for pre-policy traces.
	PolicyDecision string `json:"policy_decision,omitempty"`
	// ConfirmedBy records the provenance of an external confirmation that
	// authorized an ASK-class op to execute (e.g. ticket id / human handle
	// from the Step 5 {{user.confirm}} gate). Empty when no confirmation was
	// supplied. Provides the audit trail for "who authorized this op".
	ConfirmedBy string `json:"confirmed_by,omitempty"`
	// RequestID is the cloud API request id returned by the `ve` CLI for the
	// generator command run in this iteration ({"Response":{"RequestId":"..."}}).
	// Empty on the first iteration when the execution-risk gate blocks before
	// any `ve` call runs, or when the command produced no RequestId. Used by
	// P5 to prove every runtime `ve` call is traceable end-to-end.
	RequestID string `json:"request_id,omitempty"`
	// HealClass records the error-classification verdict (retryable /
	// rate_limit / fatal / unknown) that drove the retry decision on this
	// iteration, when `--heal=smart` is active. Empty under `--heal=none`
	// (legacy fixed-count loop). Consumed by the L4 telemetry layer (T11)
	// and the audit trail.
	HealClass string `json:"heal_class,omitempty"`
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
	Status         string          `json:"status"`
	Iter           int             `json:"iter"`
	Output         *string         `json:"output,omitempty"`
	Unresolved     []string        `json:"unresolved,omitempty"`
	FailurePattern *FailurePattern `json:"failure_pattern,omitempty"`
}

// Trace is the top-level GCL trace document.
type Trace struct {
	TraceSchemaVersion string         `json:"trace_schema_version"`
	Skill              string         `json:"skill"`
	Request            string         `json:"request"`
	RubricVersion      string         `json:"rubric_version"`
	OperationIntent    map[string]any `json:"operation_intent"`
	MaskedFields       []string       `json:"masked_fields"`
	RedactionPass      bool           `json:"redaction_pass"`
	Iterations         []Iteration    `json:"iterations"`
	Final              Final          `json:"final"`
	SourcePath         string         `json:"_source_path,omitempty"`
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

// Check validates a runtime GCL trace file (written by PersistTrace as
// gcl-trace-*.json). It enforces the P5 invariants:
//   - redaction_pass must be true (credentials masked)
//   - every iteration that actually ran a `ve` call must carry a non-empty
//     request_id (the cloud API RequestId), so the call is traceable end-to-end
//
// The runtime trace and the incident trace (incident-trace-*.json, validated by
// the check/trace package) use different schemas. This Check only handles the
// runtime shape, identified by a non-empty trace_schema_version. Files without
// that field are not runtime traces and are skipped (returning nil) so the two
// checkers stay orthogonal.
func Check(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("trace check: read %s: %w", path, err)
	}
	var t Trace
	if err := json.Unmarshal(data, &t); err != nil {
		return fmt.Errorf("trace check: parse %s: %w", path, err)
	}
	// Not a runtime trace (no trace_schema_version) → leave to the incident checker.
	if t.TraceSchemaVersion == "" {
		return nil
	}
	if !t.RedactionPass {
		return fmt.Errorf("trace check: redaction_pass must be true (credentials must be masked)")
	}
	for i, iter := range t.Iterations {
		// POLICY_BLOCK iterations never ran a `ve` call → no RequestId expected.
		if iter.Decision == "POLICY_BLOCK" {
			continue
		}
		if strings.TrimSpace(iter.Generator.Command) == "" {
			continue
		}
		if strings.TrimSpace(iter.RequestID) == "" {
			return fmt.Errorf("trace check: iterations[%d].request_id is required (ve call not traceable)", i)
		}
	}
	return nil
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

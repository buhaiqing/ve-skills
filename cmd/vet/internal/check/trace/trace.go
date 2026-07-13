// Package trace checks GCL trace files for schema compliance: request_id presence,
// policy_decision, and redaction_pass.
package trace

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Trace is the top-level incident-loop trace shape validated by Check.
type Trace struct {
	TicketID      string      `json:"ticket_id"`
	StartedAt     string     `json:"started_at"`
	FinishedAt    string     `json:"finished_at"`
	PolicyDecision string     `json:"policy_decision"`
	Iterations    []Iteration `json:"iterations"`
	RedactionPass bool       `json:"redaction_pass"`
}

// Iteration is one GCL loop iteration.
type Iteration struct {
	VeCalls []VeCall `json:"ve_calls"`
}

// VeCall is one ve CLI invocation within an iteration.
type VeCall struct {
	RequestID string `json:"request_id"`
	Action   string `json:"action"`
	Status   string `json:"status"`
}

// Check validates a single trace file.
// Returns nil if the trace passes all checks:
//   - ticket_id, started_at, finished_at, policy_decision are present
//   - every VeCall has a non-empty request_id
//   - redaction_pass is true
func Check(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("trace check: read %s: %w", path, err)
	}
	var t Trace
	if err := json.Unmarshal(data, &t); err != nil {
		return fmt.Errorf("trace check: parse %s: %w", path, err)
	}

	// Top-level required fields
	if t.TicketID == "" {
		return fmt.Errorf("trace check: ticket_id is required")
	}
	if t.StartedAt == "" {
		return fmt.Errorf("trace check: started_at is required")
	}
	if t.FinishedAt == "" {
		return fmt.Errorf("trace check: finished_at is required")
	}
	if t.PolicyDecision == "" {
		return fmt.Errorf("trace check: policy_decision is required")
	}
	validPolicyDecision := map[string]bool{"AUTO": true, "ASK": true, "REFUSE": true}
	if !validPolicyDecision[t.PolicyDecision] {
		return fmt.Errorf("trace check: policy_decision must be AUTO|ASK|REFUSE, got %q", t.PolicyDecision)
	}

	// Every VeCall must have a request_id
	for i, iter := range t.Iterations {
		for j, call := range iter.VeCalls {
			if strings.TrimSpace(call.RequestID) == "" {
				return fmt.Errorf("trace check: iterations[%d].ve_calls[%d].request_id is required", i, j)
			}
		}
	}

	// Redaction must always be true
	if !t.RedactionPass {
		return fmt.Errorf("trace check: redaction_pass must be true (credentials must be masked)")
	}

	return nil
}

// CheckDir validates incident-trace-*.json files under root/audit-results.
// Skips other trace formats (e.g. gcl-trace-*.json from the GCL runner)
// which use a different schema.
func CheckDir(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		// No audit-results dir → nothing to check
		return nil
	}
	var failures int
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		// Only check incident-trace-*.json (new schema); skip gcl-trace-*.json (legacy)
		if !strings.HasPrefix(entry.Name(), "incident-trace-") {
			continue
		}
		path := root + string(os.PathSeparator) + entry.Name()
		if err := Check(path); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", entry.Name(), err)
			failures++
		}
	}
	if failures > 0 {
		return fmt.Errorf("%d trace file(s) failed", failures)
	}
	return nil
}

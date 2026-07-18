package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// RunState is persisted to .runtime/agent/runs/<runID>/state.json for checkpoint/resume.
type RunState struct {
	RunID       string             `json:"run_id"`
	CurrentStep Step               `json:"current_step"`
	Payload     IncidentPayload    `json:"payload"`
	Triage      *TriageResult      `json:"triage,omitempty"`
	Evidence    *DiagnosisEvidence `json:"evidence,omitempty"`
	Plan        *DispatchPlan      `json:"plan,omitempty"`
	Confirm     *ConfirmResult     `json:"confirm,omitempty"`
	Result      *ExecuteResult     `json:"result,omitempty"`
	Error       string             `json:"error,omitempty"`
}

func runDir(root, runID string) string {
	return filepath.Join(root, ".runtime", "agent", "runs", runID)
}

func statePath(root, runID string) string {
	return filepath.Join(runDir(root, runID), "state.json")
}

// SaveState persists the run state as JSON.
func SaveState(root string, state *RunState) error {
	dir := runDir(root, state.RunID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(statePath(root, state.RunID), data, 0o644)
}

// LoadState reads and unmarshals a run state from disk.
// Returns nil, nil if the state file does not exist.
func LoadState(root, runID string) (*RunState, error) {
	data, err := os.ReadFile(statePath(root, runID))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var state RunState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

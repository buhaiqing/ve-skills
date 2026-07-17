package agent

import (
	"fmt"
	"os"
)

// RunResult is the final result of an agent run.
type RunResult struct {
	Success   bool   `json:"success"`
	FinalStep Step   `json:"final_step"`
	Error     string `json:"error,omitempty"`
	RunID     string `json:"run_id"`
}

// Run executes the full 7-step agent loop for an incident payload.
// Steps: INGEST → TRIAGE → DIAGNOSE → PROPOSE → CONFIRM → EXECUTE → REFLEXION
// On error at any step, the state is persisted and the error is returned.
// Structured logging via os.Stderr at every lifecycle point per AGENTS.md.
func Run(root string, payload *IncidentPayload, runID string) *RunResult {
	state := &RunState{
		RunID:       runID,
		CurrentStep: StepIngest,
		Payload:     *payload,
	}

	logStep(runID, "INGEST", "start", "product=%s symptom=%s", payload.ProductHint, payload.Symptom)

	// Step 1: INGEST — payload is already provided by caller (CLI handles parsing).
	// Save checkpoint for consistency (all 7 steps have checkpoints).
	if err := SaveState(root, state); err != nil {
		logError(runID, "INGEST", "save failed: %v", err)
		return &RunResult{Success: false, FinalStep: StepIngest, Error: fmt.Sprintf("ingest save: %v", err), RunID: runID}
	}

	// Step 2: TRIAGE.
	state.CurrentStep = StepTriage
	logStep(runID, "TRIAGE", "start", "")
	triage := Triage(payload)
	state.Triage = triage
	if err := SaveState(root, state); err != nil {
		logError(runID, "TRIAGE", "save failed: %v", err)
		return &RunResult{Success: false, FinalStep: StepTriage, Error: fmt.Sprintf("triage save: %v", err), RunID: runID}
	}
	logStep(runID, "TRIAGE", "done", "primary=%s confidence=%s", triage.PrimarySkill, triage.Confidence)

	// Step 3: DIAGNOSE.
	state.CurrentStep = StepDiagnose
	logStep(runID, "DIAGNOSE", "start", "skill=%s", triage.PrimarySkill)
	evidence := Diagnose(root, triage.PrimarySkill, payload)
	state.Evidence = evidence
	if err := SaveState(root, state); err != nil {
		logError(runID, "DIAGNOSE", "save failed: %v", err)
		return &RunResult{Success: false, FinalStep: StepDiagnose, Error: fmt.Sprintf("diagnose save: %v", err), RunID: runID}
	}
	logStep(runID, "DIAGNOSE", "done", "findings=%d partial=%v", len(evidence.Findings), evidence.Partial)

	// Step 4: PROPOSE.
	state.CurrentStep = StepPropose
	logStep(runID, "PROPOSE", "start", "")
	plan := ProposeFix(evidence, payload)
	state.Plan = plan
	if err := SaveState(root, state); err != nil {
		logError(runID, "PROPOSE", "save failed: %v", err)
		return &RunResult{Success: false, FinalStep: StepPropose, Error: fmt.Sprintf("propose save: %v", err), RunID: runID}
	}
	logStep(runID, "PROPOSE", "done", "ops=%d blast_radius=%s", len(plan.Operations), plan.BlastRadius)

	// Step 5: CONFIRM.
	state.CurrentStep = StepConfirm
	logStep(runID, "CONFIRM", "start", "")
	confirm := Confirm(root, plan)
	state.Confirm = confirm
	if err := SaveState(root, state); err != nil {
		logError(runID, "CONFIRM", "save failed: %v", err)
		return &RunResult{Success: false, FinalStep: StepConfirm, Error: fmt.Sprintf("confirm save: %v", err), RunID: runID}
	}
	logStep(runID, "CONFIRM", "done", "decision=%s reason=%s", confirm.Decision, confirm.Reason)

	if confirm.Decision == "REFUSE" {
		state.CurrentStep = StepReflexion
		logStep(runID, "REFLEXION", "policy_refused", "reason=%s", confirm.Reason)
		if err := Reflect(root, triage.PrimarySkill, "policy_refused", "execution_risk",
			"plan was refused by policy: "+confirm.Reason); err != nil {
			logError(runID, "REFLEXION", "writeback failed: %v", err)
		}
		_ = SaveState(root, state)
		return &RunResult{Success: false, FinalStep: StepConfirm, Error: "policy refused: " + confirm.Reason, RunID: runID}
	}

	// Step 6: EXECUTE.
	state.CurrentStep = StepExecute
	logStep(runID, "EXECUTE", "start", "ops=%d", len(plan.Operations))
	result := Execute(root, plan, payload.TicketID)
	state.Result = result
	if err := SaveState(root, state); err != nil {
		logError(runID, "EXECUTE", "save failed: %v", err)
		return &RunResult{Success: false, FinalStep: StepExecute, Error: fmt.Sprintf("execute save: %v", err), RunID: runID}
	}
	logStep(runID, "EXECUTE", "done", "success=%v", result.Success)

	if !result.Success {
		state.CurrentStep = StepReflexion
		logStep(runID, "REFLEXION", "failure", "error_class=%s", result.ErrorClass)
		if err := Reflect(root, triage.PrimarySkill, result.ErrorClass, "agent_execute",
			"execution failed with error class: "+result.ErrorClass); err != nil {
			logError(runID, "REFLEXION", "writeback failed: %v", err)
		}
		_ = SaveState(root, state)
		return &RunResult{Success: false, FinalStep: StepExecute, Error: "execution failed: " + result.ErrorClass, RunID: runID}
	}

	// Step 7: REFLEXION — successful run.
	state.CurrentStep = StepReflexion
	logStep(runID, "REFLEXION", "done", "success=true")
	_ = SaveState(root, state)

	return &RunResult{Success: true, FinalStep: StepReflexion, RunID: runID}
}

// logStep writes a structured lifecycle log line to stderr.
func logStep(runID, step, event, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[%s] [INFO] agent.engine | %s %s", runID, step, event)
	if format != "" {
		fmt.Fprintf(os.Stderr, " | "+format, args...)
	}
	fmt.Fprintln(os.Stderr)
}

// logError writes a structured error log line to stderr.
func logError(runID, step, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[%s] [ERROR] agent.engine | %s ", runID, step)
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintln(os.Stderr)
}

package agent

import (
	"fmt"
	"os"
)

type RunResult struct {
	Success   bool   `json:"success"`
	FinalStep Step   `json:"final_step"`
	Error     string `json:"error,omitempty"`
	RunID     string `json:"run_id"`
}

func Run(root string, payload *IncidentPayload, runID string) *RunResult {
	return runLoop(root, payload, runID, false)
}

func RunDry(root string, payload *IncidentPayload, runID string) *RunResult {
	return runLoop(root, payload, runID, true)
}

func runLoop(root string, payload *IncidentPayload, runID string, dryRun bool) *RunResult {
	state := &RunState{
		RunID:       runID,
		CurrentStep: StepIngest,
		Payload:     *payload,
	}

	// Resume: load existing state and skip completed steps
	existing, err := LoadState(root, runID)
	if err == nil && existing != nil {
		state = existing
	}

	if dryRun {
		logStep(runID, "DRY-RUN", "start", "product=%s symptom=%s", payload.ProductHint, payload.Symptom)
	} else {
		logStep(runID, "INGEST", "start", "product=%s symptom=%s", payload.ProductHint, payload.Symptom)
	}

	if state.CurrentStep <= StepIngest {
		if err := SaveState(root, state); err != nil {
			logError(runID, "INGEST", "save failed: %v", err)
			return &RunResult{Success: false, FinalStep: StepIngest, Error: fmt.Sprintf("ingest save: %v", err), RunID: runID}
		}
		state.CurrentStep = StepTriage
	}

	// Step 2: TRIAGE
	if state.CurrentStep <= StepTriage {
		logStep(runID, "TRIAGE", "start", "")
		triage := Triage(payload)
		state.Triage = triage
		if err := SaveState(root, state); err != nil {
			logError(runID, "TRIAGE", "save failed: %v", err)
			return &RunResult{Success: false, FinalStep: StepTriage, Error: fmt.Sprintf("triage save: %v", err), RunID: runID}
		}
		logStep(runID, "TRIAGE", "done", "primary=%s confidence=%s", triage.PrimarySkill, triage.Confidence)
		state.CurrentStep = StepDiagnose
	}

	// Step 3: DIAGNOSE
	if state.CurrentStep <= StepDiagnose {
		logStep(runID, "DIAGNOSE", "start", "skill=%s", state.Triage.PrimarySkill)
		evidence := Diagnose(root, state.Triage.PrimarySkill, payload)
		state.Evidence = evidence
		if err := SaveState(root, state); err != nil {
			logError(runID, "DIAGNOSE", "save failed: %v", err)
			return &RunResult{Success: false, FinalStep: StepDiagnose, Error: fmt.Sprintf("diagnose save: %v", err), RunID: runID}
		}
		logStep(runID, "DIAGNOSE", "done", "findings=%d partial=%v", len(evidence.Findings), evidence.Partial)
		state.CurrentStep = StepPropose
	}

	// Step 4: PROPOSE
	if state.CurrentStep <= StepPropose {
		logStep(runID, "PROPOSE", "start", "")
		plan := ProposeFix(state.Evidence, payload)
		state.Plan = plan
		if err := SaveState(root, state); err != nil {
			logError(runID, "PROPOSE", "save failed: %v", err)
			return &RunResult{Success: false, FinalStep: StepPropose, Error: fmt.Sprintf("propose save: %v", err), RunID: runID}
		}
		logStep(runID, "PROPOSE", "done", "ops=%d blast_radius=%s", len(plan.Operations), plan.BlastRadius)
		state.CurrentStep = StepConfirm
	}

	// Step 5: CONFIRM
	if state.CurrentStep <= StepConfirm {
		logStep(runID, "CONFIRM", "start", "")
		confirm := Confirm(root, state.Plan)
		state.Confirm = confirm
		if err := SaveState(root, state); err != nil {
			logError(runID, "CONFIRM", "save failed: %v", err)
			return &RunResult{Success: false, FinalStep: StepConfirm, Error: fmt.Sprintf("confirm save: %v", err), RunID: runID}
		}
		logStep(runID, "CONFIRM", "done", "decision=%s reason=%s", confirm.Decision, confirm.Reason)
		state.CurrentStep = StepExecute
	}

	if state.Confirm.Decision == "REFUSE" {
		state.CurrentStep = StepReflexion
		logStep(runID, "REFLEXION", "policy_refused", "reason=%s", state.Confirm.Reason)
		if err := Reflect(root, state.Triage.PrimarySkill, "policy_refused", "execution_risk",
			"plan was refused by policy: "+state.Confirm.Reason); err != nil {
			logError(runID, "REFLEXION", "writeback failed: %v", err)
		}
		_ = SaveState(root, state)
		return &RunResult{Success: false, FinalStep: StepConfirm, Error: "policy refused: " + state.Confirm.Reason, RunID: runID}
	}

	// Step 6: EXECUTE
	if state.CurrentStep <= StepExecute {
		if dryRun {
			logStep(runID, "DRY-RUN", "skip_execute", "ops=%d", len(state.Plan.Operations))
			_ = SaveState(root, state)
			return &RunResult{Success: true, FinalStep: StepExecute, RunID: runID}
		}

		logStep(runID, "EXECUTE", "start", "ops=%d", len(state.Plan.Operations))
		result := Execute(root, state.Plan, state.Payload.TicketID)
		state.Result = result
		if err := SaveState(root, state); err != nil {
			logError(runID, "EXECUTE", "save failed: %v", err)
			return &RunResult{Success: false, FinalStep: StepExecute, Error: fmt.Sprintf("execute save: %v", err), RunID: runID}
		}
		logStep(runID, "EXECUTE", "done", "success=%v", result.Success)
		state.CurrentStep = StepReflexion

		if !result.Success {
			logStep(runID, "REFLEXION", "failure", "error_class=%s", result.ErrorClass)
			if err := Reflect(root, state.Triage.PrimarySkill, result.ErrorClass, "agent_execute",
				"execution failed with error class: "+result.ErrorClass); err != nil {
				logError(runID, "REFLEXION", "writeback failed: %v", err)
			}
			_ = SaveState(root, state)
			return &RunResult{Success: false, FinalStep: StepExecute, Error: "execution failed: " + result.ErrorClass, RunID: runID}
		}
	}

	// Step 7: REFLEXION — successful run
	state.CurrentStep = StepReflexion
	logStep(runID, "REFLEXION", "done", "success=true")
	_ = SaveState(root, state)

	return &RunResult{Success: true, FinalStep: StepReflexion, RunID: runID}
}

func logStep(runID, step, event, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[%s] [INFO] agent.engine | %s %s", runID, step, event)
	if format != "" {
		fmt.Fprintf(os.Stderr, " | "+format, args...)
	}
	fmt.Fprintln(os.Stderr)
}

func logError(runID, step, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[%s] [ERROR] agent.engine | %s | ", runID, step)
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintln(os.Stderr)
}

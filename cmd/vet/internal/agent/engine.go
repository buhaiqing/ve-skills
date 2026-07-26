package agent

import (
	"context"
	"fmt"
	"os"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/observability"
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
	rootTrace := observability.NewRootTrace("agent", "run")
	ctx := observability.WithTrace(context.Background(), rootTrace)

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
		span := observability.StartSpan(observability.FromContext(ctx), "agent", "triage")
		logStep(runID, "TRIAGE", "start", "")
		triage := Triage(payload)
		state.Triage = triage
		saveErr := SaveState(root, state)
		if saveErr != nil {
			span.End(saveErr)
			logError(runID, "TRIAGE", "save failed: %v", saveErr)
			return &RunResult{Success: false, FinalStep: StepTriage, Error: fmt.Sprintf("triage save: %v", saveErr), RunID: runID}
		}
		span.End(nil)
		logStep(runID, "TRIAGE", "done", "primary=%s confidence=%s", triage.PrimarySkill, triage.Confidence)
		state.CurrentStep = StepDiagnose
	}

	// Step 3: DIAGNOSE
	if state.CurrentStep <= StepDiagnose {
		span := observability.StartSpan(observability.FromContext(ctx), "agent", "diagnose")
		logStep(runID, "DIAGNOSE", "start", "skill=%s", state.Triage.PrimarySkill)
		evidence := Diagnose(root, state.Triage.PrimarySkill, payload)
		state.Evidence = evidence
		saveErr := SaveState(root, state)
		if saveErr != nil {
			span.End(saveErr)
			logError(runID, "DIAGNOSE", "save failed: %v", saveErr)
			return &RunResult{Success: false, FinalStep: StepDiagnose, Error: fmt.Sprintf("diagnose save: %v", saveErr), RunID: runID}
		}
		span.End(nil)
		logStep(runID, "DIAGNOSE", "done", "findings=%d partial=%v", len(evidence.Findings), evidence.Partial)
		state.CurrentStep = StepPropose
	}

	// Step 4: PROPOSE
	if state.CurrentStep <= StepPropose {
		span := observability.StartSpan(observability.FromContext(ctx), "agent", "propose")
		logStep(runID, "PROPOSE", "start", "")
		plan := ProposeFix(state.Evidence, payload)
		state.Plan = plan
		saveErr := SaveState(root, state)
		if saveErr != nil {
			span.End(saveErr)
			logError(runID, "PROPOSE", "save failed: %v", saveErr)
			return &RunResult{Success: false, FinalStep: StepPropose, Error: fmt.Sprintf("propose save: %v", saveErr), RunID: runID}
		}
		span.End(nil)
		logStep(runID, "PROPOSE", "done", "ops=%d blast_radius=%s", len(plan.Operations), plan.BlastRadius)
		state.CurrentStep = StepConfirm
	}

	// Step 5: CONFIRM
	if state.CurrentStep <= StepConfirm {
		span := observability.StartSpan(observability.FromContext(ctx), "agent", "confirm")
		logStep(runID, "CONFIRM", "start", "")
		confirm := Confirm(root, state.Plan)
		state.Confirm = confirm
		saveErr := SaveState(root, state)
		if saveErr != nil {
			span.End(saveErr)
			logError(runID, "CONFIRM", "save failed: %v", saveErr)
			return &RunResult{Success: false, FinalStep: StepConfirm, Error: fmt.Sprintf("confirm save: %v", saveErr), RunID: runID}
		}
		span.End(nil)
		logStep(runID, "CONFIRM", "done", "decision=%s reason=%s", confirm.Decision, confirm.Reason)
		state.CurrentStep = StepExecute
	}

	if state.Confirm.Decision == "REFUSE" {
		state.CurrentStep = StepReflexion
		span := observability.StartSpan(observability.FromContext(ctx), "agent", "reflexion")
		logStep(runID, "REFLEXION", "policy_refused", "reason=%s", state.Confirm.Reason)
			reflectErr := Reflect(root, state.Triage.PrimarySkill, "policy_refused", "execution_risk",
				"plan was refused by policy: "+state.Confirm.Reason)
		if reflectErr != nil {
			span.End(reflectErr)
			logError(runID, "REFLEXION", "writeback failed: %v", reflectErr)
		} else {
			span.End(nil)
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

		span := observability.StartSpan(observability.FromContext(ctx), "agent", "execute")
		logStep(runID, "EXECUTE", "start", "ops=%d", len(state.Plan.Operations))
		result := Execute(root, state.Plan, state.Payload.TicketID)
		state.Result = result
		saveErr := SaveState(root, state)
		if saveErr != nil {
			span.End(saveErr)
			logError(runID, "EXECUTE", "save failed: %v", saveErr)
			return &RunResult{Success: false, FinalStep: StepExecute, Error: fmt.Sprintf("execute save: %v", saveErr), RunID: runID}
		}
		span.End(nil)
		logStep(runID, "EXECUTE", "done", "success=%v", result.Success)
		state.CurrentStep = StepReflexion

		if !result.Success {
			state.CurrentStep = StepReflexion
			reflexSpan := observability.StartSpan(observability.FromContext(ctx), "agent", "reflexion")
			logStep(runID, "REFLEXION", "failure", "error_class=%s", result.ErrorClass)
			reflectErr := Reflect(root, state.Triage.PrimarySkill, result.ErrorClass, "agent_execute",
				"execution failed with error class: "+result.ErrorClass)
			if reflectErr != nil {
				reflexSpan.End(reflectErr)
				logError(runID, "REFLEXION", "writeback failed: %v", reflectErr)
			} else {
				reflexSpan.End(nil)
			}
			_ = SaveState(root, state)
			return &RunResult{Success: false, FinalStep: StepExecute, Error: "execution failed: " + result.ErrorClass, RunID: runID}
		}
	}

	// Step 7: REFLEXION — successful run
	state.CurrentStep = StepReflexion
	span := observability.StartSpan(observability.FromContext(ctx), "agent", "reflexion")
	logStep(runID, "REFLEXION", "done", "success=true")
	span.End(nil)
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

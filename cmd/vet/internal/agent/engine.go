package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	vlog "github.com/buhaiqing/ve-skills/cmd/vet/internal/log"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/observability"
)

const agentLogPath = "audit-results/agent-execution.log"

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
	started := time.Now()
	alertedAt := started
	if payload.AlertedAt != "" {
		if t, err := time.Parse(time.RFC3339Nano, payload.AlertedAt); err == nil {
			alertedAt = t
		} else if t, err := time.Parse(time.RFC3339, payload.AlertedAt); err == nil {
			alertedAt = t
		}
	}

	rootTrace := observability.NewRootTrace("agent", "run")
	ctx := observability.WithTrace(context.Background(), rootTrace)

	state := &RunState{
		RunID:       runID,
		CurrentStep: StepIngest,
		Payload:     *payload,
		StartedAt:   started.UTC().Format(time.RFC3339Nano),
	}

	// Resume: load existing state and skip completed steps
	existing, err := LoadState(root, runID)
	if err == nil && existing != nil {
		state = existing
		if state.StartedAt == "" {
			state.StartedAt = started.UTC().Format(time.RFC3339Nano)
		} else if t, err := time.Parse(time.RFC3339Nano, state.StartedAt); err == nil {
			started = t
		}
	}

	finish := func(result *RunResult) *RunResult {
		EmitValue(root, state, started, alertedAt, result, nil)
		return result
	}

	if dryRun {
		logStep(runID, "DRY-RUN", "start", "product=%s symptom=%s", payload.ProductHint, payload.Symptom)
	} else {
		logStep(runID, "INGEST", "start", "product=%s symptom=%s", payload.ProductHint, payload.Symptom)
	}

	if state.CurrentStep <= StepIngest {
		if err := SaveState(root, state); err != nil {
			logError(runID, "INGEST", "save failed: %v", err)
			return finish(&RunResult{Success: false, FinalStep: StepIngest, Error: fmt.Sprintf("ingest save: %v", err), RunID: runID})
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
			return finish(&RunResult{Success: false, FinalStep: StepTriage, Error: fmt.Sprintf("triage save: %v", saveErr), RunID: runID})
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
			return finish(&RunResult{Success: false, FinalStep: StepDiagnose, Error: fmt.Sprintf("diagnose save: %v", saveErr), RunID: runID})
		}
		span.End(nil)
		logStep(runID, "DIAGNOSE", "done", "findings=%d partial=%v", len(evidence.Findings), evidence.Partial)
		state.CurrentStep = StepPropose
	}

	// Step 4: PROPOSE
	if state.CurrentStep <= StepPropose {
		span := observability.StartSpan(observability.FromContext(ctx), "agent", "propose")
		logStep(runID, "PROPOSE", "start", "")
		plan := ProposeFixWithRoot(root, state.Evidence, payload)
		state.Plan = plan
		saveErr := SaveState(root, state)
		if saveErr != nil {
			span.End(saveErr)
			logError(runID, "PROPOSE", "save failed: %v", saveErr)
			return finish(&RunResult{Success: false, FinalStep: StepPropose, Error: fmt.Sprintf("propose save: %v", saveErr), RunID: runID})
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
			return finish(&RunResult{Success: false, FinalStep: StepConfirm, Error: fmt.Sprintf("confirm save: %v", saveErr), RunID: runID})
		}
		span.End(nil)
		logStep(runID, "CONFIRM", "done", "decision=%s reason=%s", confirm.Decision, confirm.Reason)
		state.CurrentStep = StepExecute
	}

	if state.Confirm != nil && state.Confirm.Decision == "REFUSE" {
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
		return finish(&RunResult{Success: false, FinalStep: StepConfirm, Error: "policy refused: " + state.Confirm.Reason, RunID: runID})
	}

	// ASK without provenance: pause for agentd/CLI confirm (ADR-0006). Do not Execute.
	if state.Confirm != nil && state.Confirm.Decision == "ASK" && strings.TrimSpace(state.Confirm.ConfirmedBy) == "" {
		state.CurrentStep = StepConfirm
		state.Result = nil
		_ = SaveState(root, state)
		logStep(runID, "CONFIRM", "awaiting_human", "reason=%s", state.Confirm.Reason)
		return finish(&RunResult{
			Success:   false,
			FinalStep: StepConfirm,
			Error:     "awaiting human confirmation",
			RunID:     runID,
		})
	}

	// Step 6: EXECUTE
	if state.CurrentStep <= StepExecute {
		if state.Plan == nil {
			return finish(&RunResult{Success: false, FinalStep: StepExecute, Error: "missing plan for execute", RunID: runID})
		}
		if dryRun {
			logStep(runID, "DRY-RUN", "skip_execute", "ops=%d", len(state.Plan.Operations))
			_ = SaveState(root, state)
			return finish(&RunResult{Success: true, FinalStep: StepExecute, RunID: runID})
		}

		span := observability.StartSpan(observability.FromContext(ctx), "agent", "execute")
		logStep(runID, "EXECUTE", "start", "ops=%d", len(state.Plan.Operations))
		confirmedBy := ""
		if state.Confirm != nil {
			confirmedBy = state.Confirm.ConfirmedBy
		}
		result := Execute(root, state.Plan, state.Payload.TicketID, confirmedBy)
		state.Result = result
		saveErr := SaveState(root, state)
		if saveErr != nil {
			span.End(saveErr)
			logError(runID, "EXECUTE", "save failed: %v", saveErr)
			return finish(&RunResult{Success: false, FinalStep: StepExecute, Error: fmt.Sprintf("execute save: %v", saveErr), RunID: runID})
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
			return finish(&RunResult{Success: false, FinalStep: StepExecute, Error: "execution failed: " + result.ErrorClass, RunID: runID})
		}
	}

	// Step 7: REFLEXION — successful run
	state.CurrentStep = StepReflexion
	span := observability.StartSpan(observability.FromContext(ctx), "agent", "reflexion")
	logStep(runID, "REFLEXION", "done", "success=true")
	span.End(nil)
	_ = SaveState(root, state)

	return finish(&RunResult{Success: true, FinalStep: StepReflexion, RunID: runID})
}

// EmitValue computes, persists, and optionally writebacks value metrics.
// Writer failures are logged only — Success is unchanged.
// nil writer → FileTicketWriter{Dir: runDir(root, runID)} (ADR-0003).
func EmitValue(root string, state *RunState, started, alertedAt time.Time, result *RunResult, writer TicketWriter) {
	policy := ""
	if state.Confirm != nil {
		policy = state.Confirm.Decision
	}
	m := ComputeValue(ValueInput{
		RunID:          state.RunID,
		TicketID:       state.Payload.TicketID,
		PolicyDecision: policy,
		Success:        result.Success,
		AlertedAt:      alertedAt,
		StartedAt:      started,
		ResolvedAt:     time.Now(),
	})
	if err := PersistValue(root, m); err != nil {
		logError(state.RunID, "VALUE", "persist failed: %v", err)
	}
	if state.Payload.TicketID != "" {
		w := writer
		if w == nil {
			w = FileTicketWriter{Dir: runDir(root, state.RunID)}
		}
		if err := w.WriteValueComment(state.Payload.TicketID, FormatValueComment(m)); err != nil {
			logError(state.RunID, "VALUE", "ticket writeback failed: %v", err)
		}
	}
	state.Value = &m
	state.StartedAt = started.UTC().Format(time.RFC3339Nano)
	_ = SaveState(root, state)
}

func logStep(runID, step, event, format string, args ...interface{}) {
	msg := step + " " + event
	var kvs []string
	if format != "" {
		kvs = append(kvs, vlog.KV("detail", fmt.Sprintf(format, args...)))
	}
	_ = vlog.Append(agentLogPath, runID, vlog.INFO, "agent.engine", msg, kvs...)
}

func logError(runID, step, format string, args ...interface{}) {
	_ = vlog.Append(agentLogPath, runID, vlog.ERROR, "agent.engine", step,
		vlog.KV("detail", fmt.Sprintf(format, args...)))
}

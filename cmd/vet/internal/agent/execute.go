package agent

import (
	"strings"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/run"
)

const executeTimeout = 300

type ExecuteResult struct {
	Success    bool   `json:"success"`
	Output     string `json:"output,omitempty"`
	ErrorClass string `json:"error_class,omitempty"`
}

// Execute runs each plan op via GCL. When confirmedBy is non-empty, ASK-class
// ops are authorized with that provenance (ADR-0006).
func Execute(root string, plan *DispatchPlan, ticketID, confirmedBy string) *ExecuteResult {
	by := strings.TrimSpace(confirmedBy)
	confirmed := by != ""
	for _, op := range plan.Operations {
		structuralOnly := op.SafetyClass == "read_only"

		result := run.Run(run.Options{
			Root:           root,
			Skill:          op.Skill,
			Request:        "agent dispatch",
			Command:        op.Command,
			Timeout:        executeTimeout,
			StructuralOnly: structuralOnly,
			Heal:           "smart",
			Confirmed:      confirmed,
			ConfirmedBy:    by,
		})
		if result.ExitCode != 0 {
			return &ExecuteResult{
				Success:    false,
				ErrorClass: classifyExecError(result),
			}
		}
	}
	return &ExecuteResult{Success: true}
}

func classifyExecError(result run.Result) string {
	switch result.ExitCode {
	case 1:
		return "max_iter"
	case 2:
		return "critic_failure"
	case 3:
		return "safety_fail"
	case 4:
		return "policy_block"
	default:
		return "unknown"
	}
}

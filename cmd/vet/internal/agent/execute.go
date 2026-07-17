package agent

import "github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/run"

// ExecuteResult is the outcome of delegated execution.
type ExecuteResult struct {
	Success    bool   `json:"success"`
	Output     string `json:"output,omitempty"`
	ErrorClass string `json:"error_class,omitempty"`
}

// Execute runs the dispatch plan via the GCL runner.
// Each operation is executed through run.Run() for safety-gated execution with healing.
// Safety note: StructuralOnly critic is used for read_only operations.
// State-changing/destructive operations require external critic configuration.
func Execute(root string, plan *DispatchPlan, ticketID string) *ExecuteResult {
	for _, op := range plan.Operations {
		// For state-changing or destructive operations, StructuralOnly is not sufficient.
		// These should be executed with a real Critic or confirmed by human.
		structuralOnly := op.SafetyClass == "read_only"

		result := run.Run(run.Options{
			Root:           root,
			Skill:          op.Skill,
			Request:        "agent dispatch",
			Command:        op.Command,
			StructuralOnly: structuralOnly,
			Heal:           "smart",
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

// classifyExecError maps GCL exit codes to error classes for reflexion.
func classifyExecError(result run.Result) string {
	switch result.ExitCode {
	case 1:
		return "max_iter"
	case 2:
		return "critic_failure" // forward-compatible: will be used when real critic is configured
	case 3:
		return "safety_fail"
	case 4:
		return "policy_block"
	default:
		return "unknown"
	}
}

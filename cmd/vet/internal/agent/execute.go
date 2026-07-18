package agent

import "github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/run"

const executeTimeout = 300

type ExecuteResult struct {
	Success    bool   `json:"success"`
	Output     string `json:"output,omitempty"`
	ErrorClass string `json:"error_class,omitempty"`
}

func Execute(root string, plan *DispatchPlan, ticketID string) *ExecuteResult {
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

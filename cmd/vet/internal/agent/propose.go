package agent

import (
	"fmt"
	"strings"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/strategy"
)

// DispatchOp is a single operation in the dispatch plan.
type DispatchOp struct {
	Skill       string  `json:"skill"`
	Command     string  `json:"command"`
	SafetyClass string  `json:"safety_class"`
	BlastRadius string  `json:"blast_radius"`
	Confidence  string  `json:"confidence"`
	Safety      float64 `json:"safety"`
}

// DispatchPlan is the proposed fix plan.
type DispatchPlan struct {
	Operations       []DispatchOp `json:"operations"`
	BlastRadius      string       `json:"blast_radius"`
	RollbackPlan     string       `json:"rollback_plan,omitempty"`
	HealIncidentType string       `json:"heal_incident_type,omitempty"` // e.g. "cpu_high"
}

// healIncidentType maps symptom text to a known heal recovery plan name.
func healIncidentType(symptom string) string {
	s := strings.ToLower(symptom)
	switch {
	case strings.Contains(s, "cpu"):
		return "cpu_high"
	case strings.Contains(s, "redis") && strings.Contains(s, "slow"):
		return "redis_slow_query"
	case strings.Contains(s, "mysql") || strings.Contains(s, "connection"):
		return "mysql_connection_pool"
	case strings.Contains(s, "vpc") && strings.Contains(s, "route"):
		return "vpc_route_table"
	default:
		return ""
	}
}

// ProposeFix builds a dispatch plan from diagnosis evidence and the original payload.
// Uses the diagnosed skill to build product-specific commands.
func ProposeFix(evidence *DiagnosisEvidence, payload *IncidentPayload) *DispatchPlan {
	healType := healIncidentType(payload.Symptom)
	kb := strategy.NewKnowledgeBase()
	if pattern := kb.Query(payload.Symptom); pattern != nil {
		return &DispatchPlan{
			Operations: []DispatchOp{
				{
					Skill:       pattern.Skill,
					Command:     pattern.Solution,
					SafetyClass: "read_only",
					BlastRadius: "single",
					Confidence:  fmt.Sprintf("%.0f%%", pattern.Confidence*100),
					Safety:      1.0,
				},
			},
			BlastRadius:      "single",
			RollbackPlan:     "no rollback needed (diagnostic only)",
			HealIncidentType: healType,
		}
	}

	symptom := strings.ToLower(payload.Symptom)
	skill := evidence.Skill
	service := skillToService(skill) // e.g., ve-ecs-ops → ecs

	hasCPU := strings.Contains(symptom, "cpu")
	hasMem := strings.Contains(symptom, "mem")
	hasDisk := strings.Contains(symptom, "disk")
	hasLatency := strings.Contains(symptom, "latency")
	hasConnection := strings.Contains(symptom, "connection")
	hasErrors := strings.Contains(symptom, "error") || strings.Contains(symptom, "failure")

	var ops []DispatchOp
	switch {
	case hasCPU:
		ops = append(ops, newReadOp(skill, "ve "+service+" DescribeInstances", "high"))
	case hasMem:
		ops = append(ops, newReadOp(skill, "ve "+service+" DescribeInstances", "high"))
	case hasDisk:
		ops = append(ops, newReadOp(skill, "ve "+service+" DescribeDisks", "high"))
	case hasLatency:
		ops = append(ops, newReadOp(skill, "ve "+service+" DescribeInstances", "medium"))
	case hasConnection:
		ops = append(ops, newReadOp(skill, "ve "+service+" DescribeInstances", "medium"))
	case hasErrors:
		ops = append(ops, newReadOp(skill, "ve "+service+" DescribeInstances", "low"))
	default:
		ops = append(ops, newReadOp(skill, "ve "+service+" DescribeInstances", "low"))
	}

	return &DispatchPlan{
		Operations:       ops,
		BlastRadius:      "single",
		RollbackPlan:     "no rollback needed (diagnostic only)",
		HealIncidentType: healType,
	}
}

func newReadOp(skill, command, confidence string) DispatchOp {
	return DispatchOp{
		Skill:       skill,
		Command:     command,
		SafetyClass: "read_only",
		BlastRadius: "single",
		Confidence:  confidence,
		Safety:      1.0,
	}
}

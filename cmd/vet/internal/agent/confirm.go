package agent

import (
	"strings"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/heal"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/policy"
)

// ConfirmResult is the policy decision for a dispatch plan.
type ConfirmResult struct {
	Decision    string `json:"decision"` // AUTO / ASK / REFUSE
	Reason      string `json:"reason"`
	ConfirmedBy string `json:"confirmed_by,omitempty"`
}

// Confirm evaluates the dispatch plan against execution-risk policies.
// Reuses policy.Load() to read execution-risk.md + domain-allowlist.md + guardrails.yaml.
func Confirm(root string, plan *DispatchPlan) *ConfirmResult {
	ps, err := policy.Load(root)
	if err != nil {
		return &ConfirmResult{Decision: "ASK", Reason: "policy load failed: " + err.Error()}
	}

	// Build the allowlist set for fast lookup.
	allowlist := map[string]bool{}
	for _, s := range ps.DomainAllowlist {
		allowlist[s] = true
	}

	// Build guardrail map by trigger pattern.
	// Guardrails may be empty (no patterns reached count≥10 yet) — that's normal.
	// Guardrails being nil means guardrails.yaml doesn't exist — also normal.
	guardrails := ps.Guardrails
	if guardrails == nil {
		guardrails = []policy.GuardrailEntry{} // normalize nil to empty for safe iteration
	}

	for _, op := range plan.Operations {
		// Safety=0 is hard floor — always REFUSE.
		if op.Safety == 0 {
			return &ConfirmResult{Decision: "REFUSE", Reason: "operation " + op.Command + " has safety=0 (hard floor)"}
		}

		// Destructive operations always require human confirmation.
		if op.SafetyClass == "destructive" {
			return &ConfirmResult{Decision: "ASK", Reason: "operation " + op.Command + " is destructive"}
		}

		// Blast radius must be single for AUTO.
		if op.BlastRadius != "single" {
			return &ConfirmResult{Decision: "ASK", Reason: "operation " + op.Command + " has blast radius " + op.BlastRadius + ", not single"}
		}

		// Confidence must be high for AUTO.
		if op.Confidence != "high" {
			return &ConfirmResult{Decision: "ASK", Reason: "operation " + op.Command + " has confidence " + op.Confidence + ", not high"}
		}

		// Must be in the domain allowlist.
		if !allowlist[op.Skill] {
			return &ConfirmResult{Decision: "ASK", Reason: "skill " + op.Skill + " is not in the domain allowlist"}
		}

		// Check guardrails for this skill+command combination.
		for _, g := range guardrails {
			if g.Skill == op.Skill && strings.Contains(op.Command, g.Trigger) {
				switch strings.ToUpper(g.Action) {
				case "REFUSE":
					return &ConfirmResult{Decision: "REFUSE", Reason: "guardrail " + g.ID + " triggered: " + g.Trigger}
				case "ASK":
					return &ConfirmResult{Decision: "ASK", Reason: "guardrail " + g.ID + " triggered: " + g.Trigger}
				case "AUTO":
					// AUTO guardrail means "allow", continue to next check.
				}
			}
		}
	}

	// Stub heal plans must not enter production AUTO (P0-2 / spec §3.2).
	if plan.HealIncidentType != "" {
		hp := heal.NewOrchestrator().Plan(plan.HealIncidentType)
		if hp != nil && hp.IsStub() {
			return &ConfirmResult{
				Decision: "ASK",
				Reason:   "heal plan " + plan.HealIncidentType + " is stub (no real probe); production AUTO forbidden",
			}
		}
	}

	return &ConfirmResult{Decision: "AUTO", Reason: "all operations pass policy checks"}
}

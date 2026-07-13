// Package policyguard checks the policy-invariant that SAFETY=0 never AUTO.
package policyguard

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Decision is the policy decision for a single operation.
type Decision int

const (
	DECISION_REFUSE Decision = iota
	DECISION_ASK
	DECISION_AUTO
)

func (d Decision) String() string {
	switch d {
	case DECISION_REFUSE:
		return "REFUSE"
	case DECISION_ASK:
		return "ASK"
	case DECISION_AUTO:
		return "AUTO"
	}
	return "UNKNOWN"
}

// Operation describes a single operation in a dispatch plan.
type Operation struct {
	Skill        string  `json:"skill"`
	Command      string  `json:"command"`
	SafetyClass  string  `json:"safety_class"`  // "read_only" | "mutating" | "destructive"
	BlastRadius  string  `json:"blast_radius"`  // "single" | "multi" | "account-or-region"
	Confidence   string  `json:"confidence"`    // "high" | "medium" | "low"
	Safety       float64 `json:"safety"`        // 0.0–1.0; 0 = hard floor → REFUSE
	MetadataOK   bool    `json:"metadata_ok"`   // false → fail-safe → ASK (never AUTO)
}

// DispatchPlan is the input to Check.
type DispatchPlan struct {
	Operations []Operation `json:"operations"`
}

// Check enforces three invariants on the dispatch plan.
// Returns nil if all invariants hold; returns a descriptive error otherwise.
func Check(planPath string) error {
	data, err := os.ReadFile(planPath)
	if err != nil {
		return fmt.Errorf("policyguard: read %s: %w", planPath, err)
	}
	var plan DispatchPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return fmt.Errorf("policyguard: parse %s: %w", planPath, err)
	}
	return CheckPlan(plan)
}

// CheckPlan is the pure in-memory variant of Check.
func CheckPlan(plan DispatchPlan) error {
	for i, op := range plan.Operations {
		if err := checkOp(i, op); err != nil {
			return err
		}
	}
	return nil
}

func checkOp(idx int, op Operation) error {
	// Invariant 1: safety=0 → REFUSE (hard floor, overrides everything)
	if op.Safety == 0 {
		return fmt.Errorf("invariant-1 violated: operation %d (%s %s) has safety=0, must REFUSE",
			idx, op.Skill, op.Command)
	}

	// Invariant 2: destructive → never AUTO
	if op.SafetyClass == "destructive" {
		return fmt.Errorf("invariant-2 violated: operation %d (%s %s) has safety_class=destructive, must ASK or REFUSE",
			idx, op.Skill, op.Command)
	}

	// Invariant 3: missing metadata → never AUTO (fail-safe)
	if !op.MetadataOK {
		return fmt.Errorf("invariant-3 violated: operation %d (%s %s) has missing metadata, must ASK",
			idx, op.Skill, op.Command)
	}

	return nil
}

// ComputeDecision mirrors the T06 scoreDecision scorer.
// It applies the execution-risk policy: Safety=0 → REFUSE,
// destructive → ASK, missing metadata → ASK, otherwise AUTO.
func ComputeDecision(op Operation) Decision {
	if op.Safety == 0 {
		return DECISION_REFUSE
	}
	if op.SafetyClass == "destructive" {
		return DECISION_ASK
	}
	if !op.MetadataOK {
		return DECISION_ASK
	}
	return DECISION_AUTO
}

// ViolationReport describes a single invariant violation for human review.
type ViolationReport struct {
	OpIndex      int    `json:"op_index"`
	Skill        string `json:"skill"`
	Command      string `json:"command"`
	Invariant    int    `json:"invariant"` // 1, 2, or 3
	Actual       string `json:"actual_decision"`
	Expected     string `json:"expected_decision"`
	Description  string `json:"description"`
}

// CheckPlanWithReport is like CheckPlan but also returns detailed violation info.
func CheckPlanWithReport(plan DispatchPlan) ([]ViolationReport, error) {
	var reports []ViolationReport
	for i, op := range plan.Operations {
		if op.Safety == 0 {
			reports = append(reports, ViolationReport{
				OpIndex:   i,
				Skill:     op.Skill,
				Command:   op.Command,
				Invariant: 1,
				Actual:    "REFUSE",
				Expected:  "REFUSE",
				Description: "safety=0 is a hard floor — must REFUSE regardless of other signals",
			})
		}
		if op.SafetyClass == "destructive" {
			reports = append(reports, ViolationReport{
				OpIndex:   i,
				Skill:     op.Skill,
				Command:   op.Command,
				Invariant: 2,
				Actual:    "AUTO or ASK",
				Expected:  "ASK or REFUSE",
				Description: "destructive ops require human confirmation — AUTO is never appropriate",
			})
		}
		if !op.MetadataOK {
			reports = append(reports, ViolationReport{
				OpIndex:   i,
				Skill:     op.Skill,
				Command:   op.Command,
				Invariant: 3,
				Actual:    "AUTO",
				Expected:  "ASK",
				Description: "missing metadata triggers fail-safe — must ASK, never AUTO",
			})
		}
	}
	return reports, nil
}

// AllowedSkills returns the set of skills eligible for AUTO per domain-allowlist.
var AllowedSkills = map[string]bool{
	"ve-cms-ops":       true,
	"ve-ecs-ops":       true,
	"ve-rds-mysql-ops": true,
	"ve-redis-ops":     true,
	"ve-vpc-ops":       true,
	"ve-iam-ops":       true,
	"ve-kms-ops":       true,
	"ve-billing-ops":   true,
}

// IsAllowedSkill returns true if the skill is in the domain allow-list.
func IsAllowedSkill(skill string) bool {
	return AllowedSkills[skill]
}

// DestructiveVerbs matches commands that perform destructive actions.
var DestructiveVerbs = []string{
	"Delete", "Remove", "Terminate", "Destroy", "Stop", "Shutdown",
	"PowerOff", "Release", "Revoke", "Disable", "Deactivate", "Flush",
	"Purge", "Drop", "Truncate", "Detach", "Disassociate",
}

// IsDestructiveCommand returns true if the command verb suggests a destructive operation.
func IsDestructiveCommand(command string) bool {
	lower := strings.ToLower(command)
	for _, verb := range DestructiveVerbs {
		if strings.Contains(lower, strings.ToLower(verb)) {
			return true
		}
	}
	return false
}

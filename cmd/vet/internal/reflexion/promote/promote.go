package promote

import "fmt"

// Level represents the enforcement level of a failure pattern.
type Level int

const (
	LevelPruned     Level = iota // count < 3
	LevelHint                    // 3 ≤ count < 10
	LevelConstraint              // 10 ≤ count < 30
	LevelHard                    // count ≥ 30
)

// Pattern represents a failure pattern tracked across sessions.
type Pattern struct {
	Category string
	Skill    string
	Pattern  string
	Fix      string
	Count    int
}

// LevelOf returns the Level for a pattern based on its count.
func LevelOf(p Pattern) Level {
	switch {
	case p.Count >= 30:
		return LevelHard
	case p.Count >= 10:
		return LevelConstraint
	case p.Count >= 3:
		return LevelHint
	default:
		return LevelPruned
	}
}

// Enforce checks a pattern against a plan and returns the enforcement action.
// - Hint: returns LevelHint, no error (just inject context)
// - Constraint: returns LevelConstraint with error if plan conflicts
// - Hard: returns LevelHard with error (ABORT)
// - Pruned: returns LevelPruned (skip)
func Enforce(p Pattern, planSafety int) (Level, error) {
	lvl := LevelOf(p)
	switch lvl {
	case LevelHint, LevelPruned:
		return lvl, nil
	case LevelConstraint:
		return lvl, fmt.Errorf("constraint violated")
	case LevelHard:
		return lvl, fmt.Errorf("hard guard ABORT")
	default:
		return lvl, nil
	}
}

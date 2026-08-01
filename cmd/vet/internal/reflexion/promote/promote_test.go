package promote

import (
	"testing"
)

func TestLevelOf_Pruned(t *testing.T) {
	counts := []int{0, 1, 2}
	for _, c := range counts {
		p := Pattern{Count: c}
		if got := LevelOf(p); got != LevelPruned {
			t.Errorf("LevelOf(count=%d) = %v, want LevelPruned", c, got)
		}
	}
}

func TestLevelOf_Hint(t *testing.T) {
	counts := []int{3, 4}
	for _, c := range counts {
		p := Pattern{Count: c}
		if got := LevelOf(p); got != LevelHint {
			t.Errorf("LevelOf(count=%d) = %v, want LevelHint", c, got)
		}
	}
}

func TestLevelOf_Constraint(t *testing.T) {
	counts := []int{5, 10, 14}
	for _, c := range counts {
		p := Pattern{Count: c}
		if got := LevelOf(p); got != LevelConstraint {
			t.Errorf("LevelOf(count=%d) = %v, want LevelConstraint", c, got)
		}
	}
}

func TestLevelOf_Hard(t *testing.T) {
	counts := []int{15, 100}
	for _, c := range counts {
		p := Pattern{Count: c}
		if got := LevelOf(p); got != LevelHard {
			t.Errorf("LevelOf(count=%d) = %v, want LevelHard", c, got)
		}
	}
}

func TestLevelOf_Boundaries(t *testing.T) {
	tests := []struct {
		count int
		want  Level
	}{
		{2, LevelPruned},
		{3, LevelHint},
		{4, LevelHint},
		{5, LevelConstraint},
		{14, LevelConstraint},
		{15, LevelHard},
	}
	for _, tt := range tests {
		p := Pattern{Count: tt.count}
		if got := LevelOf(p); got != tt.want {
			t.Errorf("LevelOf(count=%d) = %v, want %v", tt.count, got, tt.want)
		}
	}
}

func TestEnforce_Hint(t *testing.T) {
	p := Pattern{Count: 4}
	lvl, err := Enforce(p, 1)
	if lvl != LevelHint {
		t.Errorf("Enforce(Hint) level = %v, want LevelHint", lvl)
	}
	if err != nil {
		t.Errorf("Enforce(Hint) error = %v, want nil", err)
	}
}

func TestEnforce_Constraint(t *testing.T) {
	p := Pattern{Count: 10}
	lvl, err := Enforce(p, 1)
	if lvl != LevelConstraint {
		t.Errorf("Enforce(Constraint) level = %v, want LevelConstraint", lvl)
	}
	if err == nil {
		t.Fatal("Enforce(Constraint) error = nil, want non-nil")
	}
	if err.Error() != "constraint violated" {
		t.Errorf("Enforce(Constraint) error = %q, want %q", err.Error(), "constraint violated")
	}
}

func TestEnforce_Hard(t *testing.T) {
	p := Pattern{Count: 20}
	lvl, err := Enforce(p, 1)
	if lvl != LevelHard {
		t.Errorf("Enforce(Hard) level = %v, want LevelHard", lvl)
	}
	if err == nil {
		t.Fatal("Enforce(Hard) error = nil, want non-nil")
	}
	if err.Error() != "hard guard ABORT" {
		t.Errorf("Enforce(Hard) error = %q, want %q", err.Error(), "hard guard ABORT")
	}
}

func TestEnforce_Pruned(t *testing.T) {
	p := Pattern{Count: 1}
	lvl, err := Enforce(p, 1)
	if lvl != LevelPruned {
		t.Errorf("Enforce(Pruned) level = %v, want LevelPruned", lvl)
	}
	if err != nil {
		t.Errorf("Enforce(Pruned) error = %v, want nil", err)
	}
}

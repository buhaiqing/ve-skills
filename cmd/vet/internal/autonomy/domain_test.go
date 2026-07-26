package autonomy

import (
	"testing"
)

func TestAutoExpand(t *testing.T) {
	d := NewAutonomousDomain("test-domain")
	d.Policies = append(d.Policies, Policy{Name: "p1", Action: "heal", Enabled: true})
	d.Actions = append(d.Actions, Action{Type: "heal", Target: "skill-a"})

	ok := d.AutoExpand("skill-a", 0.8)
	if !ok {
		t.Error("expected AutoExpand to return true when maturity >= 0.7 and can expand")
	}
	if d.Level != L2Autonomous {
		t.Errorf("expected level L2, got %s", d.Level)
	}
	if len(d.Skills) != 1 || d.Skills[0] != "skill-a" {
		t.Errorf("expected skill 'skill-a', got %v", d.Skills)
	}
}

func TestAutoExpandLowMaturity(t *testing.T) {
	d := NewAutonomousDomain("test-domain")
	d.Policies = append(d.Policies, Policy{Name: "p1", Action: "heal", Enabled: true})
	d.Actions = append(d.Actions, Action{Type: "heal", Target: "skill-a"})

	ok := d.AutoExpand("skill-a", 0.5)
	if ok {
		t.Error("expected AutoExpand to return false when maturity < 0.7")
	}
	if d.Level != L1Basic {
		t.Errorf("expected level L1 (unchanged), got %s", d.Level)
	}
	if len(d.Skills) != 1 {
		t.Errorf("expected 1 skill added, got %d", len(d.Skills))
	}
}

func TestAutoExpandDuplicateSkill(t *testing.T) {
	d := NewAutonomousDomain("test-domain")
	d.Skills = append(d.Skills, "skill-a")

	ok := d.AutoExpand("skill-a", 0.9)
	if ok {
		t.Error("expected AutoExpand to return false for duplicate skill")
	}
	if len(d.Skills) != 1 {
		t.Errorf("expected 1 skill (no duplicate), got %d", len(d.Skills))
	}
}

func TestAutoExpandCannotProgress(t *testing.T) {
	d := NewAutonomousDomain("test-domain")

	ok := d.AutoExpand("skill-a", 0.9)
	if ok {
		t.Error("expected AutoExpand to return false when cannot progress (no policies/actions)")
	}
	if d.Level != L1Basic {
		t.Errorf("expected level L1, got %s", d.Level)
	}
}

func TestCanExpandTo(t *testing.T) {
	d := NewAutonomousDomain("test-domain")

	d.Level = L1Basic
	d.Policies = append(d.Policies, Policy{Name: "p1", Action: "heal", Enabled: true})
	d.Actions = append(d.Actions, Action{Type: "heal"})

	if !d.canExpandTo(L2Autonomous) {
		t.Error("expected L1→L2 expansion to succeed with 1 policy and 1 action")
	}

	if d.canExpandTo(L3Semantic) {
		t.Error("expected L1→L3 expansion to fail without meeting L3 requirements")
	}

	d.Policies = append(d.Policies, Policy{Name: "p2"}, Policy{Name: "p3"})
	d.Actions = append(d.Actions, Action{Type: "scale"}, Action{Type: "notify"})
	d.Knowledge["k1"] = "v1"
	d.Knowledge["k2"] = "v2"
	d.Knowledge["k3"] = "v3"

	if !d.canExpandTo(L3Semantic) {
		t.Error("expected L2→L3 expansion to succeed with 3 policies, 3 actions, 3 knowledge")
	}

	if d.canExpandTo(L4SelfEvolving) {
		t.Error("expected L2→L4 expansion to fail without meeting L4 requirements")
	}

	d.Policies = append(d.Policies, Policy{Name: "p4"}, Policy{Name: "p5"})
	d.Actions = append(d.Actions, Action{Type: "expand"}, Action{Type: "heal"})
	d.Knowledge["k4"] = "v4"
	d.Knowledge["k5"] = "v5"

	if !d.canExpandTo(L4SelfEvolving) {
		t.Error("expected L3→L4 expansion to succeed with 5 policies, 5 actions, 5 knowledge")
	}
}

func TestEvaluateMaturity(t *testing.T) {
	d := NewAutonomousDomain("test-domain")

	score := d.EvaluateMaturity()
	if score != 0.0 {
		t.Errorf("expected score 0.0 for empty domain, got %f", score)
	}

	d.Policies = append(d.Policies, Policy{Name: "p1"}, Policy{Name: "p2"})
	d.Actions = append(d.Actions, Action{Type: "heal"}, Action{Type: "scale"})
	d.Knowledge["k1"] = "v1"

	score = d.EvaluateMaturity()
	if score <= 0.0 {
		t.Errorf("expected positive maturity score, got %f", score)
	}

	d.Policies = append(d.Policies,
		Policy{Name: "p3"}, Policy{Name: "p4"}, Policy{Name: "p5"},
	)
	d.Actions = append(d.Actions,
		Action{Type: "notify"}, Action{Type: "expand"}, Action{Type: "heal"},
	)
	d.Knowledge["k2"] = "v2"
	d.Knowledge["k3"] = "v3"
	d.Knowledge["k4"] = "v4"
	d.Knowledge["k5"] = "v5"

	score = d.EvaluateMaturity()
	if score < 1.0 {
		t.Errorf("expected high maturity score (>=1.0), got %f", score)
	}
}

func TestCrossCoordinate(t *testing.T) {
	d1 := NewAutonomousDomain("domain-1")
	d1.Skills = append(d1.Skills, "skill-a", "skill-b")

	d2 := NewAutonomousDomain("domain-2")
	d2.Skills = append(d2.Skills, "skill-b", "skill-c")

	actions := d1.CrossCoordinate(d2)
	if len(actions) == 0 {
		t.Fatal("expected cross-domain actions for overlapping skill 'skill-b'")
	}

	overlapCount := 0
	for _, a := range actions {
		if a.Target == "skill-b" {
			overlapCount++
		}
	}
	if overlapCount != 2 {
		t.Errorf("expected 2 actions for overlapping skill (heal+notify), got %d", overlapCount)
	}
}

func TestCrossCoordinateNoOverlap(t *testing.T) {
	d1 := NewAutonomousDomain("domain-1")
	d1.Skills = append(d1.Skills, "skill-a")

	d2 := NewAutonomousDomain("domain-2")
	d2.Skills = append(d2.Skills, "skill-b")

	actions := d1.CrossCoordinate(d2)
	if actions != nil {
		t.Error("expected nil actions for no overlapping skills")
	}
}

func TestReconcile(t *testing.T) {
	d := NewAutonomousDomain("test-domain")
	d.Skills = append(d.Skills, "orphan-skill", "covered-skill")
	d.Policies = append(d.Policies, Policy{Name: "p1", Action: "heal", Enabled: true})
	d.Actions = append(d.Actions, Action{Type: "heal", Target: "covered-skill"})

	anomalies := d.Reconcile()

	foundOrphan := false
	for _, a := range anomalies {
		if a == "orphan skill: orphan-skill" {
			foundOrphan = true
		}
	}
	if !foundOrphan {
		t.Error("expected to detect orphan skill 'orphan-skill'")
	}
}

func TestReconcilePolicyWithoutAction(t *testing.T) {
	d := NewAutonomousDomain("test-domain")
	d.Skills = append(d.Skills, "skill-a")
	d.Policies = append(d.Policies,
		Policy{Name: "p1", Action: "heal", Enabled: true},
		Policy{Name: "p2", Action: "scale", Enabled: true},
	)
	d.Actions = append(d.Actions, Action{Type: "heal", Target: "skill-a"})

	anomalies := d.Reconcile()

	foundPolicyAnomaly := false
	for _, a := range anomalies {
		if a == "policy without action: p2" {
			foundPolicyAnomaly = true
		}
	}
	if !foundPolicyAnomaly {
		t.Error("expected to detect policy 'p2' without matching action")
	}
}

func TestReconcileClean(t *testing.T) {
	d := NewAutonomousDomain("test-domain")
	d.Skills = append(d.Skills, "skill-a")
	d.Policies = append(d.Policies, Policy{Name: "p1", Action: "heal", Enabled: true})
	d.Actions = append(d.Actions, Action{Type: "heal", Target: "skill-a"})

	anomalies := d.Reconcile()
	if len(anomalies) != 0 {
		t.Errorf("expected no anomalies, got %v", anomalies)
	}
}
package policyguard

import (
	"encoding/json"
	"os"
	"testing"
)

func TestComputeDecision_SafetyZeroRefuse(t *testing.T) {
	op := Operation{Safety: 0, SafetyClass: "read_only", MetadataOK: true}
	if got := ComputeDecision(op); got != DECISION_REFUSE {
		t.Errorf("safety=0: got %v, want REFUSE", got)
	}
}

func TestComputeDecision_DestructiveAsk(t *testing.T) {
	op := Operation{Safety: 1.0, SafetyClass: "destructive", MetadataOK: true}
	if got := ComputeDecision(op); got != DECISION_ASK {
		t.Errorf("destructive: got %v, want ASK", got)
	}
}

func TestComputeDecision_MissingMetadataAsk(t *testing.T) {
	op := Operation{Safety: 1.0, SafetyClass: "read_only", MetadataOK: false}
	if got := ComputeDecision(op); got != DECISION_ASK {
		t.Errorf("missing metadata: got %v, want ASK", got)
	}
}

func TestComputeDecision_AutoHappyPath(t *testing.T) {
	op := Operation{Safety: 1.0, SafetyClass: "read_only", MetadataOK: true}
	if got := ComputeDecision(op); got != DECISION_AUTO {
		t.Errorf("happy path: got %v, want AUTO", got)
	}
}

func TestCheckPlan_Invariant1_SafetyZeroNotRefuse(t *testing.T) {
	// safety=0 but decision is AUTO → must be caught
	plan := DispatchPlan{Operations: []Operation{
		{Skill: "ve-ecs-ops", Command: "DeleteInstance", Safety: 0, SafetyClass: "destructive", MetadataOK: true},
	}}
	err := CheckPlan(plan)
	if err == nil {
		t.Fatal("expected invariant-1 violation, got nil")
	}
	if Invariant1 != 1 || !contains(err.Error(), "invariant-1") {
		t.Errorf("expected invariant-1 message, got: %v", err)
	}
}

func TestCheckPlan_Invariant2_DestructiveAuto(t *testing.T) {
	// destructive but AUTO → invariant-2 violated
	plan := DispatchPlan{Operations: []Operation{
		{Skill: "ve-ecs-ops", Command: "DeleteInstance", Safety: 1.0, SafetyClass: "destructive", MetadataOK: true},
	}}
	err := CheckPlan(plan)
	if err == nil {
		t.Fatal("expected invariant-2 violation, got nil")
	}
	if !contains(err.Error(), "invariant-2") {
		t.Errorf("expected invariant-2 message, got: %v", err)
	}
}

func TestCheckPlan_Invariant3_MissingMetaAuto(t *testing.T) {
	// missing metadata but AUTO → invariant-3 violated
	plan := DispatchPlan{Operations: []Operation{
		{Skill: "ve-ecs-ops", Command: "DescribeInstances", Safety: 1.0, SafetyClass: "read_only", MetadataOK: false},
	}}
	err := CheckPlan(plan)
	if err == nil {
		t.Fatal("expected invariant-3 violation, got nil")
	}
	if !contains(err.Error(), "invariant-3") {
		t.Errorf("expected invariant-3 message, got: %v", err)
	}
}

func TestCheckPlan_AllClean(t *testing.T) {
	plan := DispatchPlan{Operations: []Operation{
		{Skill: "ve-ecs-ops", Command: "DescribeInstances", Safety: 1.0, SafetyClass: "read_only", MetadataOK: true},
		{Skill: "ve-ecs-ops", Command: "StopInstances", Safety: 1.0, SafetyClass: "mutating", MetadataOK: true},
	}}
	if err := CheckPlan(plan); err != nil {
		t.Errorf("expected clean pass, got: %v", err)
	}
}

func TestCheckPlanWithReport_Mixed(t *testing.T) {
	plan := DispatchPlan{Operations: []Operation{
		{Skill: "ve-ecs-ops", Command: "DescribeInstances", Safety: 1.0, SafetyClass: "read_only", MetadataOK: true},   // clean → 0 violations
		{Skill: "ve-ecs-ops", Command: "StopInstances", Safety: 0, SafetyClass: "destructive", MetadataOK: true},        // invariant-1 only (safety=0 triggers hard floor first)
		{Skill: "ve-redis-ops", Command: "DeleteKeys", Safety: 1.0, SafetyClass: "destructive", MetadataOK: true},    // invariant-2 only
		{Skill: "ve-ecs-ops", Command: "DescribeInstances", Safety: 1.0, SafetyClass: "read_only", MetadataOK: false},   // invariant-3 only
	}}
	reports, err := CheckPlanWithReport(plan)
	if err != nil {
		t.Fatalf("CheckPlanWithReport returned error: %v", err)
	}
	if len(reports) != 4 {
		t.Errorf("expected 4 violations (op1 triggers invariants 1+2; op2 triggers inv2; op3 triggers inv3), got %d: %+v", len(reports), reports)
	}
}

func TestIsAllowedSkill(t *testing.T) {
	if !IsAllowedSkill("ve-ecs-ops") {
		t.Error("ve-ecs-ops should be allowed")
	}
	if !IsAllowedSkill("ve-cms-ops") {
		t.Error("ve-cms-ops should be allowed")
	}
	if IsAllowedSkill("ve-unknown-ops") {
		t.Error("ve-unknown-ops should NOT be allowed")
	}
}

func TestIsDestructiveCommand(t *testing.T) {
	cases := []struct {
		cmd  string
		want bool
	}{
		{"DeleteInstance", true},
		{"StopInstances", true},
		{"DescribeInstances", false},
		{"ModifyInstanceAttribute", false},
		{"CreateUser", false},
	}
	for _, c := range cases {
		if got := IsDestructiveCommand(c.cmd); got != c.want {
			t.Errorf("IsDestructiveCommand(%q) = %v, want %v", c.cmd, got, c.want)
		}
	}
}

func TestCheck_FileNotFound(t *testing.T) {
	err := Check("/nonexistent/path/plan.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestCheck_JSONInvalid(t *testing.T) {
	f, err := os.CreateTemp("", "policyguard_invalid_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString("not json"); err != nil {
		t.Fatal(err)
	}
	f.Close()
	if err := Check(f.Name()); err == nil {
		t.Error("expected parse error")
	}
}

func TestDecision_String(t *testing.T) {
	tests := []struct {
		d    Decision
		want string
	}{
		{DECISION_REFUSE, "REFUSE"},
		{DECISION_ASK, "ASK"},
		{DECISION_AUTO, "AUTO"},
		{Decision(99), "UNKNOWN"},
	}
	for _, tt := range tests {
		if got := tt.d.String(); got != tt.want {
			t.Errorf("Decision(%d).String() = %q, want %q", tt.d, got, tt.want)
		}
	}
}

// TestCheck_viaFile exercises Check through a real JSON file (testdata).
func TestCheck_viaFile(t *testing.T) {
	f, err := os.CreateTemp("", "policyguard_clean_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	plan := DispatchPlan{Operations: []Operation{
		{Skill: "ve-ecs-ops", Command: "DescribeInstances", Safety: 1.0, SafetyClass: "read_only", MetadataOK: true},
	}}
	if err := json.NewEncoder(f).Encode(plan); err != nil {
		t.Fatal(err)
	}
	f.Close()
	if err := Check(f.Name()); err != nil {
		t.Errorf("expected clean pass via file, got: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Invariant numbers for use in error messages.
const (
	Invariant1 = 1 // safety=0 → REFUSE
	Invariant2 = 2 // destructive → not AUTO
	Invariant3 = 3 // missing metadata → not AUTO
)

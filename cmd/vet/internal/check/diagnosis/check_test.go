package diagnosis

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRulesEngine_Match(t *testing.T) {
	e := NewRulesEngine()
	e.Rules = []DiagnosisRule{
		{
			ID:       "cpu-surge",
			Product:  "ECS",
			Severity: "high",
			Trigger: TriggerCondition{
				Metric:    "CPU",
				Operator:  ">",
				Threshold: 90,
				Duration:  "5m",
			},
			Steps: []RuleStep{
				{Action: "check", Target: "memory", Suggestion: "check memory"},
			},
		},
	}

	actions := e.Match("ECS", "CPU", 95, 10*time.Minute)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].RuleID != "cpu-surge" {
		t.Errorf("expected rule id cpu-surge, got %s", actions[0].RuleID)
	}

	// Should NOT match — value too low
	actions = e.Match("ECS", "CPU", 50, 10*time.Minute)
	if len(actions) != 0 {
		t.Fatalf("expected 0 actions for low value, got %d", len(actions))
	}

	// Should NOT match — wrong product
	actions = e.Match("RDS", "CPU", 95, 10*time.Minute)
	if len(actions) != 0 {
		t.Fatalf("expected 0 actions for wrong product, got %d", len(actions))
	}

	// Should NOT match — duration too short
	actions = e.Match("ECS", "CPU", 95, 2*time.Minute)
	if len(actions) != 0 {
		t.Fatalf("expected 0 actions for short duration, got %d", len(actions))
	}
}

func TestCheckDir_ValidRules(t *testing.T) {
	tmp := t.TempDir()
	skillDir := filepath.Join(tmp, "ve-test-ops", "references", "advanced")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	yamlContent := `rules:
  - id: test-rule
    product: ECS
    severity: high
    description: test
    trigger:
      metric: CPU
      operator: ">"
      threshold: 90
    steps:
      - action: check
        target: memory
        suggestion: check memory
`
	if err := os.WriteFile(filepath.Join(skillDir, "diagnosis-rules.yaml"), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	results, skills := CheckDir(tmp)
	if len(skills) == 0 {
		t.Fatal("expected at least one skill")
	}
	_, ok := results["ve-test-ops"]
	if !ok {
		t.Fatal("expected ve-test-ops in results")
	}
	errs := results["ve-test-ops"]
	if len(errs) > 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
}

func TestCheckDir_MissingFields(t *testing.T) {
	tmp := t.TempDir()
	skillDir := filepath.Join(tmp, "ve-test-ops", "references", "advanced")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	yamlContent := `rules:
  - id: ""
    product: ""
    severity: ""
    trigger:
      metric: ""
      operator: ""
`
	if err := os.WriteFile(filepath.Join(skillDir, "diagnosis-rules.yaml"), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	results, _ := CheckDir(tmp)
	errs := results["ve-test-ops"]
	if len(errs) == 0 {
		t.Fatal("expected validation errors for missing fields")
	}
}

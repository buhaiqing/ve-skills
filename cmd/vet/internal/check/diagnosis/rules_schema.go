package diagnosis

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// DiagnosisRule represents one rule from diagnosis-rules.yaml
type DiagnosisRule struct {
	ID            string             `yaml:"id"`
	Product       string             `yaml:"product"`
	Trigger      TriggerCondition   `yaml:"trigger"`
	Steps        []RuleStep         `yaml:"steps"`
	CorrelateWith []CorrelationSpec `yaml:"correlate_with,omitempty"`
	Severity     string             `yaml:"severity"`
	Description  string             `yaml:"description"`
}

// TriggerCondition specifies the metric that triggers this rule
type TriggerCondition struct {
	Metric    string  `yaml:"metric"`
	Operator  string  `yaml:"operator"`
	Threshold float64 `yaml:"threshold"`
	Duration  string  `yaml:"duration,omitempty"`
}

// RuleStep is one step in the diagnosis workflow for a matched rule
type RuleStep struct {
	Action     string `yaml:"action"`
	Target     string `yaml:"target,omitempty"`
	Condition  string `yaml:"condition,omitempty"`
	ThenDo     string `yaml:"then_do,omitempty"`
	Suggestion string `yaml:"suggestion,omitempty"`
}

// CorrelationSpec describes cross-product alarm correlation
type CorrelationSpec struct {
	Product   string  `yaml:"product"`
	Metric    string  `yaml:"metric"`
	Window    string  `yaml:"window"`
	Threshold float64 `yaml:"threshold"`
}

// DiagnosisAction is the output of RulesEngine.Match()
type DiagnosisAction struct {
	RuleID     string
	StepIndex  int
	Action    string
	Target    string
	Suggestion string
	DelegateTo string // populated when Action == "delegate"
}

// RulesEngine loads and matches diagnosis rules against incoming alarms
type RulesEngine struct {
	Rules []DiagnosisRule
}

func NewRulesEngine() *RulesEngine {
	return &RulesEngine{}
}

// LoadRules reads diagnosis-rules.yaml for a given skill
func (e *RulesEngine) LoadRules(root, skill string) error {
	p := filepath.Join(root, skill, "references", "advanced", "diagnosis-rules.yaml")
	data, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	var doc struct {
		Rules []DiagnosisRule `yaml:"rules"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return err
	}
	e.Rules = doc.Rules
	return nil
}

// Match finds all rules matching the given alarm parameters.
// Returns ordered list of DiagnosisAction to take.
func (e *RulesEngine) Match(product, metric string, value float64, dur time.Duration) []*DiagnosisAction {
	var result []*DiagnosisAction
	for _, rule := range e.Rules {
		if rule.Product != product {
			continue
		}
		if !e.matchTrigger(&rule.Trigger, metric, value, dur) {
			continue
		}
		for i, step := range rule.Steps {
			a := &DiagnosisAction{
				RuleID:     rule.ID,
				StepIndex:  i,
				Action:    step.Action,
				Target:    step.Target,
				Suggestion: step.Suggestion,
				DelegateTo: step.Target,
			}
			result = append(result, a)
		}
	}
	return result
}

func (e *RulesEngine) matchTrigger(tc *TriggerCondition, metric string, value float64, dur time.Duration) bool {
	if tc.Metric != metric {
		return false
	}
	var matched bool
	switch tc.Operator {
	case ">":
		matched = value > tc.Threshold
	case ">=":
		matched = value >= tc.Threshold
	case "<":
		matched = value < tc.Threshold
	case "<=":
		matched = value <= tc.Threshold
	case "==":
		matched = value == tc.Threshold
	default:
		return false
	}
	if !matched {
		return false
	}
	if tc.Duration != "" {
		expected, err := time.ParseDuration(tc.Duration)
		if err == nil && dur < expected {
			return false
		}
	}
	return true
}

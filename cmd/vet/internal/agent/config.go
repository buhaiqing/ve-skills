package agent

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	triage "github.com/buhaiqing/ve-skills/cmd/vet/internal/triage"
	"gopkg.in/yaml.v3"
)

type TriageConfig struct {
	ConfidenceThreshold float64 `yaml:"confidence_threshold"`
	StrategyEnable      bool    `yaml:"strategy_enable"`
	TopK                int     `yaml:"top_k"`
	FallbackSkill       string  `yaml:"fallback_skill"`
}

type AgentConfig struct {
	Triage        TriageConfig `yaml:"triage"`
	MaxStateRetry int          `yaml:"max_state_retry"`
	DryRun        bool         `yaml:"dry_run"`
}

func DefaultAgentConfig() AgentConfig {
	return AgentConfig{
		Triage: TriageConfig{
			ConfidenceThreshold: 0.4,
			StrategyEnable:      true,
			TopK:                3,
			FallbackSkill:       "ve-cms-ops",
		},
		MaxStateRetry: 1,
		DryRun:        false,
	}
}

func (a *AgentConfig) ApplyDefaults() {
	defaults := DefaultAgentConfig()
	if a.Triage.ConfidenceThreshold < 0 {
		a.Triage.ConfidenceThreshold = 0
	}
	if a.Triage.ConfidenceThreshold > 1.0 {
		a.Triage.ConfidenceThreshold = 1.0
	}
	if a.Triage.TopK < 1 {
		a.Triage.TopK = defaults.Triage.TopK
	}
	if a.Triage.FallbackSkill == "" {
		a.Triage.FallbackSkill = defaults.Triage.FallbackSkill
	}
	if a.MaxStateRetry < 1 {
		a.MaxStateRetry = defaults.MaxStateRetry
	}
}

func LoadConfigFromYAML(path string) (AgentConfig, error) {
	cfg := DefaultAgentConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return AgentConfig{}, fmt.Errorf("read config: %w", err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return cfg, nil
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return AgentConfig{}, fmt.Errorf("parse yaml: %w", err)
	}
	cfg.ApplyDefaults()
	return cfg, nil
}

func LoadConfigFromEnv(prefix string) AgentConfig {
	cfg := DefaultAgentConfig()
	p := prefix + "_"
	if v := os.Getenv(p + "TRIAGE_CONFIDENCE_THRESHOLD"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.Triage.ConfidenceThreshold = f
		}
	}
	if v := os.Getenv(p + "TRIAGE_STRATEGY_ENABLE"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.Triage.StrategyEnable = b
		}
	}
	if v := os.Getenv(p + "TRIAGE_TOP_K"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.Triage.TopK = i
		}
	}
	if v := os.Getenv(p + "TRIAGE_FALLBACK_SKILL"); v != "" {
		cfg.Triage.FallbackSkill = v
	}
	if v := os.Getenv(p + "MAX_STATE_RETRY"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.MaxStateRetry = i
		}
	}
	if v := os.Getenv(p + "DRY_RUN"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.DryRun = b
		}
	}
	cfg.ApplyDefaults()
	return cfg
}

func TriageWithConfig(payload *IncidentPayload, cfg TriageConfig) *TriageResult {
	classifier := triage.DefaultClassifier()
	input := payload.Symptom + " " + payload.ProductHint
	results := classifier.Classify(input, cfg.TopK)

	if len(results) > 0 && results[0].Confidence >= cfg.ConfidenceThreshold {
		primary := results[0].Skill
		secondary := make([]string, 0, len(results)-1)
		for i := 1; i < len(results); i++ {
			secondary = append(secondary, results[i].Skill)
		}
		return &TriageResult{
			PrimarySkill:    primary,
			SecondarySkills: secondary,
			Confidence:      fmt.Sprintf("%.0f%%", results[0].Confidence*100),
		}
	}
	skill, ok := skillMap[payload.ProductHint]
	if !ok {
		skill = cfg.FallbackSkill
	}
	return &TriageResult{
		PrimarySkill: skill,
		Confidence:   confidenceFromSkill(skill, ok),
	}
}

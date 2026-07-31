package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultAgentConfig(t *testing.T) {
	cfg := DefaultAgentConfig()

	if cfg.Triage.ConfidenceThreshold != 0.4 {
		t.Errorf("ConfidenceThreshold = %v, want 0.4", cfg.Triage.ConfidenceThreshold)
	}
	if cfg.Triage.StrategyEnable != true {
		t.Errorf("StrategyEnable = %v, want true", cfg.Triage.StrategyEnable)
	}
	if cfg.Triage.TopK != 3 {
		t.Errorf("TopK = %v, want 3", cfg.Triage.TopK)
	}
	if cfg.Triage.FallbackSkill != "ve-cms-ops" {
		t.Errorf("FallbackSkill = %q, want %q", cfg.Triage.FallbackSkill, "ve-cms-ops")
	}
	if cfg.MaxStateRetry != 1 {
		t.Errorf("MaxStateRetry = %v, want 1", cfg.MaxStateRetry)
	}
	if cfg.DryRun != false {
		t.Errorf("DryRun = %v, want false", cfg.DryRun)
	}
}

func TestApplyDefaultsBoundary(t *testing.T) {
	cfg := AgentConfig{
		Triage: TriageConfig{
			ConfidenceThreshold: 1.5,
			TopK:                0,
			FallbackSkill:       "",
		},
		MaxStateRetry: 0,
	}
	cfg.ApplyDefaults()

	if cfg.Triage.ConfidenceThreshold != 1.0 {
		t.Errorf("ConfidenceThreshold=1.5 clamped to %v, want 1.0", cfg.Triage.ConfidenceThreshold)
	}

	cfg2 := AgentConfig{
		Triage: TriageConfig{
			ConfidenceThreshold: -0.2,
		},
	}
	cfg2.ApplyDefaults()
	if cfg2.Triage.ConfidenceThreshold != 0.0 {
		t.Errorf("ConfidenceThreshold=-0.2 clamped to %v, want 0.0", cfg2.Triage.ConfidenceThreshold)
	}
	if cfg.Triage.TopK != 3 {
		t.Errorf("TopK=0 defaulted to %v, want 3", cfg.Triage.TopK)
	}
	if cfg.Triage.FallbackSkill != "ve-cms-ops" {
		t.Errorf("FallbackSkill=\"\" defaulted to %q, want %q", cfg.Triage.FallbackSkill, "ve-cms-ops")
	}
	if cfg.MaxStateRetry != 1 {
		t.Errorf("MaxStateRetry=0 defaulted to %v, want 1", cfg.MaxStateRetry)
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("VETR_AGENT_TRIAGE_CONFIDENCE_THRESHOLD", "0.7")
	os.Setenv("VETR_AGENT_TRIAGE_TOP_K", "5")
	t.Cleanup(func() {
		os.Unsetenv("VETR_AGENT_TRIAGE_CONFIDENCE_THRESHOLD")
		os.Unsetenv("VETR_AGENT_TRIAGE_TOP_K")
	})

	cfg := LoadConfigFromEnv("VETR_AGENT")

	if cfg.Triage.ConfidenceThreshold != 0.7 {
		t.Errorf("ConfidenceThreshold = %v, want 0.7", cfg.Triage.ConfidenceThreshold)
	}
	if cfg.Triage.TopK != 5 {
		t.Errorf("TopK = %v, want 5", cfg.Triage.TopK)
	}
}

func TestLoadFromYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `triage:
  confidence_threshold: 0.8
  fallback_skill: "ve-ecs-ops"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	cfg, err := LoadConfigFromYAML(path)
	if err != nil {
		t.Fatalf("LoadConfigFromYAML: %v", err)
	}

	if cfg.Triage.ConfidenceThreshold != 0.8 {
		t.Errorf("ConfidenceThreshold = %v, want 0.8", cfg.Triage.ConfidenceThreshold)
	}
	if cfg.Triage.FallbackSkill != "ve-ecs-ops" {
		t.Errorf("FallbackSkill = %q, want %q", cfg.Triage.FallbackSkill, "ve-ecs-ops")
	}
	defaults := DefaultAgentConfig()
	if cfg.Triage.StrategyEnable != defaults.Triage.StrategyEnable {
		t.Errorf("StrategyEnable = %v, want default %v", cfg.Triage.StrategyEnable, defaults.Triage.StrategyEnable)
	}
	if cfg.Triage.TopK != defaults.Triage.TopK {
		t.Errorf("TopK = %v, want default %v", cfg.Triage.TopK, defaults.Triage.TopK)
	}
	if cfg.MaxStateRetry != defaults.MaxStateRetry {
		t.Errorf("MaxStateRetry = %v, want default %v", cfg.MaxStateRetry, defaults.MaxStateRetry)
	}
	if cfg.DryRun != defaults.DryRun {
		t.Errorf("DryRun = %v, want default %v", cfg.DryRun, defaults.DryRun)
	}
}

func TestTriageWithConfigLowThreshold(t *testing.T) {
	payload := &IncidentPayload{
		ProductHint: "ecs",
		Symptom:     "ECS实例CPU高，内存占用大",
		RawInput:    "test",
		Source:      "test",
	}

	defaultCfg := DefaultAgentConfig().Triage
	lowCfg := TriageConfig{
		ConfidenceThreshold: 0.01,
		StrategyEnable:      true,
		TopK:                3,
		FallbackSkill:       "ve-cms-ops",
	}

	result1 := Triage(payload)
	result2 := TriageWithConfig(payload, defaultCfg)

	if result1.PrimarySkill != result2.PrimarySkill {
		t.Errorf("Triage() PrimarySkill=%q != TriageWithConfig(default) PrimarySkill=%q",
			result1.PrimarySkill, result2.PrimarySkill)
	}
	if result1.Confidence != result2.Confidence {
		t.Errorf("Triage() Confidence=%q != TriageWithConfig(default) Confidence=%q",
			result1.Confidence, result2.Confidence)
	}

	result3 := TriageWithConfig(payload, lowCfg)
	if result3.PrimarySkill == "" {
		t.Error("TriageWithConfig(low threshold) returned empty PrimarySkill")
	}
}

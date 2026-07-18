package autonomy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvelope(t *testing.T) {
	// Create temporary envelope file
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, "test-envelope.yaml")
	content := `domains:
  - name: redis-slow-commands
    skills: [ve-redis-ops]
    symptoms: [slow-commands, oom-prevention]
    blast_radius: single
    slo_ref: redis-p99-latency
  - name: ecs-idle-cleanup
    skills: [ve-ecs-ops]
    symptoms: [idle-resource-cleanup]
    blast_radius: single
    slo_ref: ecs-idle-cost
`
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test envelope: %v", err)
	}

	env, err := LoadEnvelope(envPath)
	if err != nil {
		t.Fatalf("LoadEnvelope failed: %v", err)
	}
	if len(env.Domains) != 2 {
		t.Errorf("expected 2 domains, got %d", len(env.Domains))
	}
}

func TestLoadEnvelope_NotFound(t *testing.T) {
	_, err := LoadEnvelope("/nonexistent/path.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestInEnvelope_Match(t *testing.T) {
	env := &Envelope{
		Domains: []Domain{
			{
				Name:        "redis-slow",
				Skills:      []string{"ve-redis-ops"},
				Symptoms:    []string{"slow-commands"},
				BlastRadius: "single",
			},
		},
	}

	if !env.InEnvelope("ve-redis-ops", "slow-commands", "single") {
		t.Error("expected InEnvelope to return true for matching domain")
	}
}

func TestInEnvelope_NoMatch_Skill(t *testing.T) {
	env := &Envelope{
		Domains: []Domain{
			{
				Name:        "redis-slow",
				Skills:      []string{"ve-redis-ops"},
				Symptoms:    []string{"slow-commands"},
				BlastRadius: "single",
			},
		},
	}

	if env.InEnvelope("ve-rds-ops", "slow-queries", "single") {
		t.Error("expected InEnvelope to return false for non-matching skill")
	}
}

func TestInEnvelope_NoMatch_BlastRadius(t *testing.T) {
	env := &Envelope{
		Domains: []Domain{
			{
				Name:        "redis-slow",
				Skills:      []string{"ve-redis-ops"},
				Symptoms:    []string{"slow-commands"},
				BlastRadius: "single",
			},
		},
	}

	if env.InEnvelope("ve-redis-ops", "slow-commands", "multi") {
		t.Error("expected InEnvelope to return false for blast_radius > single")
	}
}

func TestListSkills(t *testing.T) {
	env := &Envelope{
		Domains: []Domain{
			{Skills: []string{"ve-redis-ops"}},
			{Skills: []string{"ve-ecs-ops"}},
			{Skills: []string{"ve-redis-ops"}}, // duplicate
		},
	}

	skills := env.ListSkills()
	if len(skills) != 2 {
		t.Errorf("expected 2 unique skills, got %d", len(skills))
	}
}

func TestGetDomain(t *testing.T) {
	env := &Envelope{
		Domains: []Domain{
			{Name: "redis-slow", Skills: []string{"ve-redis-ops"}, Symptoms: []string{"slow-commands"}},
			{Name: "ecs-idle", Skills: []string{"ve-ecs-ops"}, Symptoms: []string{"idle"}},
		},
	}

	d := env.GetDomain("ve-redis-ops", "slow-commands")
	if d == nil {
		t.Fatal("expected to find domain")
	}
	if d.Name != "redis-slow" {
		t.Errorf("expected domain name 'redis-slow', got %q", d.Name)
	}
}

func TestGetDomain_NotFound(t *testing.T) {
	env := &Envelope{
		Domains: []Domain{
			{Name: "redis-slow", Skills: []string{"ve-redis-ops"}, Symptoms: []string{"slow-commands"}},
		},
	}

	d := env.GetDomain("ve-rds-ops", "slow-queries")
	if d != nil {
		t.Error("expected nil for non-matching domain")
	}
}

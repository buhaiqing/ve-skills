package autonomy

import (
	"path/filepath"
	"testing"
)

// TestLoadEnvelope_FromShippedMarkdown verifies the shipped policy doc
// (autonomy-envelope.md, prose + fenced YAML block) is loadable by LoadEnvelope.
// This guards T16 DoD #10: `vet autonomy test --envelope autonomy-envelope.md` must run.
func TestLoadEnvelope_FromShippedMarkdown(t *testing.T) {
	path := filepath.Join("..", "..", "..", "..", "incident-loop-agent", "references", "policies", "autonomy-envelope.md")
	env, err := LoadEnvelope(path)
	if err != nil {
		t.Fatalf("LoadEnvelope failed on shipped autonomy-envelope.md: %v", err)
	}
	if len(env.Domains) != 2 {
		t.Fatalf("expected 2 domains in shipped envelope, got %d", len(env.Domains))
	}

	// Anchor targets referenced by goals.yaml must resolve.
	if env.GetDomain("ve-redis-ops", "slow-commands") == nil {
		t.Error("expected domain for ve-redis-ops/slow-commands (goals.yaml anchor #redis-slow-commands)")
	}
	if env.GetDomain("ve-ecs-ops", "idle-resource-cleanup") == nil {
		t.Error("expected domain for ve-ecs-ops/idle-resource-cleanup (goals.yaml anchor #ecs-idle-cleanup)")
	}
	if !env.InEnvelope("ve-redis-ops", "slow-commands", "single") {
		t.Error("expected ve-redis-ops/slow-commands/single to be in envelope")
	}
}

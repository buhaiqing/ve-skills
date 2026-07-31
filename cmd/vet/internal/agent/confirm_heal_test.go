package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeMinimalPolicy(t *testing.T, root string) {
	t.Helper()
	dir := filepath.Join(root, "incident-loop-agent", "references", "policies")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(dir, "execution-risk.md"), []byte(`# Execution Risk Policy

## 2. Decision matrix

| risk | blast_radius | decision |
|------|--------------|----------|
| read-only | single | AUTO |
`), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "domain-allowlist.md"), []byte("# Domain\n\n## 1. Eligible skills\n\n`ve-ecs-ops`\n`ve-redis-ops`\n"), 0o644)
}

func autoPlan(healType, skill, cmd string) *DispatchPlan {
	return &DispatchPlan{
		Operations: []DispatchOp{{
			Skill: skill, Command: cmd, SafetyClass: "read_only",
			BlastRadius: "single", Confidence: "high", Safety: 1.0,
		}},
		BlastRadius:      "single",
		HealIncidentType: healType,
	}
}

func TestConfirmPromotedHealAllowsAUTO(t *testing.T) {
	root := t.TempDir()
	writeMinimalPolicy(t, root)
	res := Confirm(root, autoPlan("cpu_high", "ve-ecs-ops", "ve ecs DescribeInstances"))
	if res.Decision != "AUTO" {
		t.Fatalf("got %s (%s), want AUTO", res.Decision, res.Reason)
	}
}

func TestConfirmStubHealStillASK(t *testing.T) {
	root := t.TempDir()
	writeMinimalPolicy(t, root)
	res := Confirm(root, autoPlan("mysql_connection_pool", "ve-ecs-ops", "ve ecs DescribeInstances"))
	if res.Decision != "ASK" || !strings.Contains(strings.ToLower(res.Reason), "stub") {
		t.Fatalf("got %s (%s), want ASK stub", res.Decision, res.Reason)
	}
}

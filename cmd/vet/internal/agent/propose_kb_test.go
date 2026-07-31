package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/strategy"
)

func TestProposeFixUsesPersistedPattern(t *testing.T) {
	root := t.TempDir()
	kb := strategy.NewKnowledgeBase()
	kb.Learn(strategy.FailurePattern{
		Pattern:    "cpu spike special",
		Skill:      "ve-ecs-ops",
		Solution:   "ve ecs DescribeInstances --InstanceIds i-from-kb",
		Confidence: 0.95,
	})
	path := filepath.Join(root, ".runtime", "strategy", "kb.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := kb.Save(path); err != nil {
		t.Fatal(err)
	}
	ev := &DiagnosisEvidence{Skill: "ve-ecs-ops"}
	plan := ProposeFixWithRoot(root, ev, &IncidentPayload{Symptom: "cpu spike special", ProductHint: "ecs"})
	if len(plan.Operations) != 1 || !strings.Contains(plan.Operations[0].Command, "i-from-kb") {
		t.Fatalf("expected KB solution, got %+v", plan.Operations)
	}
}

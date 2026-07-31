package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildEvalReportRates(t *testing.T) {
	r := BuildEvalReport(EvalReportInput{Samples: []EvalSample{
		{PredictedSkill: "ve-ecs-ops", LabeledSkill: "ve-ecs-ops", GCLFirstPass: true, PolicyDecision: "AUTO", ShouldRefuse: false},
		{PredictedSkill: "ve-redis-ops", LabeledSkill: "ve-ecs-ops", GCLFirstPass: false, PolicyDecision: "REFUSE", ShouldRefuse: false},
		{PredictedSkill: "ve-ecs-ops", LabeledSkill: "ve-ecs-ops", GCLFirstPass: true, PolicyDecision: "REFUSE", ShouldRefuse: true},
	}})
	if r.TriageTop1Accuracy < 0.66 || r.TriageTop1Accuracy > 0.67 {
		t.Fatalf("triage=%v", r.TriageTop1Accuracy)
	}
	if r.GCLFirstPassRate < 0.66 || r.GCLFirstPassRate > 0.67 {
		t.Fatalf("gcl=%v", r.GCLFirstPassRate)
	}
	if r.FalseRefuseRate != 0.5 {
		t.Fatalf("falseRefuse=%v", r.FalseRefuseRate)
	}
	if r.Samples != 3 {
		t.Fatalf("samples=%d", r.Samples)
	}
}

func TestLoadEvalSamplesAndWriteEvalReport(t *testing.T) {
	dir := t.TempDir()
	inPath := filepath.Join(dir, "samples.json")
	data := []byte(`{"samples":[{"predicted_skill":"ve-ecs-ops","labeled_skill":"ve-ecs-ops","gcl_first_pass":true,"policy_decision":"AUTO","should_refuse":false}]}`)
	if err := os.WriteFile(inPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadEvalSamples(inPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Samples) != 1 {
		t.Fatalf("samples=%d", len(loaded.Samples))
	}
	outPath := filepath.Join(dir, "out", "eval-report.json")
	rep := BuildEvalReport(loaded)
	if err := WriteEvalReport(outPath, rep); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatal(err)
	}
}

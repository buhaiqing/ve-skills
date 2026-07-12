package aiops

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, p, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCheckSkill(t *testing.T) {
	dir := t.TempDir()
	skill := "ve-ecs-ops"
	base := filepath.Join(dir, skill, "references", "advanced")
	writeFile(t, filepath.Join(base, "aiops.md"), "# AIOps\n")
	writeFile(t, filepath.Join(base, "finops.md"), "# FinOps\n")
	writeFile(t, filepath.Join(dir, skill, "assets", "eval_queries.json"),
		`[{"should_trigger":true},{"should_trigger":true},{"should_trigger":true},{"should_trigger":true},{"should_trigger":true},{"should_trigger":false},{"should_trigger":false}]`)

	r := checkSkill(dir, skill)
	if !r.AIOpsMD {
		t.Error("expected AIOpsMD true")
	}
	if !r.FinOpsMD {
		t.Error("expected FinOpsMD true")
	}
	if !r.EvalQueries {
		t.Fatal("expected EvalQueries true")
	}
	if r.EvalTrigger != 5 {
		t.Errorf("expected 5 triggers, got %d", r.EvalTrigger)
	}
	if r.EvalNonTrigger != 2 {
		t.Errorf("expected 2 non-triggers, got %d", r.EvalNonTrigger)
	}
}

func TestCheckDirOK(t *testing.T) {
	dir := t.TempDir()
	// Every R+R skill gets aiops.md+finops.md; every skill gets a quality eval.
	for _, s := range AllSkills {
		base := filepath.Join(dir, s, "references", "advanced")
		if RequiredRecommended[s] {
			writeFile(t, filepath.Join(base, "aiops.md"), "# AIOps\n")
			writeFile(t, filepath.Join(base, "finops.md"), "# FinOps\n")
		}
		writeFile(t, filepath.Join(dir, s, "assets", "eval_queries.json"),
			`[{"should_trigger":true},{"should_trigger":true},{"should_trigger":true},{"should_trigger":true},{"should_trigger":true},{"should_trigger":false},{"should_trigger":false}]`)
	}
	rep := CheckDir(dir)
	if !rep.OK {
		t.Fatalf("expected OK report, got %+v", rep)
	}
	if rep.TotalSkills != len(AllSkills) {
		t.Errorf("expected %d skills, got %d", len(AllSkills), rep.TotalSkills)
	}
}

func TestCheckDirMissing(t *testing.T) {
	dir := t.TempDir()
	// ve-ecs-ops is R+R but missing both advanced md and eval.
	if err := os.MkdirAll(filepath.Join(dir, "ve-ecs-ops"), 0o755); err != nil {
		t.Fatal(err)
	}
	rep := CheckDir(dir)
	if rep.OK {
		t.Fatal("expected non-OK report when coverage is missing")
	}
	found := false
	for _, s := range rep.SkillsMissingAIOps {
		if s == "ve-ecs-ops" {
			found = true
		}
	}
	if !found {
		t.Error("expected ve-ecs-ops in SkillsMissingAIOps")
	}
}

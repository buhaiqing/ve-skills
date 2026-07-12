package eval

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestTokenize(t *testing.T) {
	// English words (3+ chars) lower-cased.
	tok := Tokenize("Create an ECS Instance")
	if !tok["create"] {
		t.Error("expected 'create' token")
	}
	if !tok["ecs"] {
		t.Error("expected product acronym 'ecs' token")
	}
	if !tok["instance"] {
		t.Error("expected 'instance' token")
	}
	// Short words are dropped.
	if tok["an"] {
		t.Error("short word 'an' should be dropped")
	}
	// Chinese keyword detection.
	tok2 := Tokenize("创建 ECS 实例")
	if !tok2["创建"] {
		t.Error("expected Chinese keyword '创建'")
	}
	if !tok2["实例"] {
		t.Error("expected Chinese keyword '实例'")
	}
	if !tok2["ecs"] {
		t.Error("expected acronym 'ecs' from Chinese text")
	}
}

func TestScoreMatches(t *testing.T) {
	query := Tokenize("create ECS instance")
	scope := Tokenize("Create an ECS instance via console")
	got := scoreMatches(query, scope)
	if got != 1.0 {
		t.Errorf("expected 1.0 coverage, got %v", got)
	}

	// Stop-words excluded from significant set.
	qStop := Tokenize("the how what when")
	if scoreMatches(qStop, scope) != 0.0 {
		t.Error("all-stopword query must score 0.0")
	}

	// Partial overlap.
	qPartial := Tokenize("redis cluster scaling")
	scPartial := Tokenize("redis instance backup")
	got2 := scoreMatches(qPartial, scPartial)
	// redis matches; cluster, scaling do not -> 1/3
	if got2 < 0.3 || got2 > 0.34 {
		t.Errorf("expected ~0.33, got %v", got2)
	}
}

func TestValidateEvalSchema(t *testing.T) {
	// Valid: 8 entries with required keys.
	valid := make([]map[string]any, 0, 8)
	for i := 0; i < 8; i++ {
		valid = append(valid, map[string]any{
			"query":          "q",
			"should_trigger": i%2 == 0,
			"skill":          "ve-x-ops",
			"confidence":     "high",
		})
	}
	if errs := validateEvalSchema(valid, "ve-x-ops"); len(errs) != 0 {
		t.Fatalf("expected no errors for valid schema, got %v", errs)
	}

	// Empty list.
	if errs := validateEvalSchema(nil, "ve-x-ops"); len(errs) != 1 {
		t.Fatalf("expected empty error, got %v", errs)
	}

	// Missing keys.
	bad := []map[string]any{{"query": "q"}}
	errs := validateEvalSchema(bad, "ve-x-ops")
	if len(errs) == 0 {
		t.Fatal("expected missing-key errors")
	}

	// Invalid bool and confidence.
	bad2 := []map[string]any{{
		"query":          "q",
		"should_trigger": "yes",
		"skill":          "ve-x-ops",
		"confidence":     "high",
	}}
	if errs := validateEvalSchema(bad2, "ve-x-ops"); len(errs) == 0 {
		t.Fatal("expected bool error")
	}

	bad3 := []map[string]any{{
		"query":          "q",
		"should_trigger": true,
		"skill":          "ve-x-ops",
		"confidence":     "unknown",
	}}
	if errs := validateEvalSchema(bad3, "ve-x-ops"); len(errs) == 0 {
		t.Fatal("expected confidence error")
	}
}

func TestCheckSkillAbsolute(t *testing.T) {
	dir := t.TempDir()
	skill := filepath.Join(dir, "ve-x-ops")
	if err := os.MkdirAll(filepath.Join(skill, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	// SKILL.md with a SHOULD section mentioning ECS.
	skillMd := `## Trigger & Scope

### SHOULD Use This Skill When
- Create an ECS instance
- Resize the ECS cluster

### SHOULD NOT Use This Skill When
- Manage RDS databases
`
	if err := os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte(skillMd), 0o644); err != nil {
		t.Fatal(err)
	}
	// Eval with a trigger query that matches and a non-trigger that does not.
	eval := `[
  {"query":"create an ECS instance for my project","should_trigger":true,"skill":"ve-x-ops","confidence":"high"},
  {"query":"manage RDS database backups","should_trigger":false,"skill":"ve-x-ops","confidence":"medium"},
  {"query":"create ECS cluster scaling","should_trigger":true,"skill":"ve-x-ops","confidence":"low"},
  {"query":"delete ECS instance now","should_trigger":true,"skill":"ve-x-ops","confidence":"high"},
  {"query":"resize ECS cluster size","should_trigger":true,"skill":"ve-x-ops","confidence":"medium"},
  {"query":"list RDS instances","should_trigger":false,"skill":"ve-x-ops","confidence":"low"},
  {"query":"stop ECS instance safely","should_trigger":true,"skill":"ve-x-ops","confidence":"high"},
  {"query":"configure RDS parameter","should_trigger":false,"skill":"ve-x-ops","confidence":"medium"}
]`
	if err := os.WriteFile(filepath.Join(skill, "assets", "eval_queries.json"), []byte(eval), 0o644); err != nil {
		t.Fatal(err)
	}
	name, errs := checkSkill(skill)
	if len(errs) != 0 {
		t.Fatalf("%s: expected no errors, got %v", name, errs)
	}
}

func TestCheckSkillLowMatch(t *testing.T) {
	dir := t.TempDir()
	skill := filepath.Join(dir, "ve-x-ops")
	if err := os.MkdirAll(filepath.Join(skill, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	skillMd := `## Trigger & Scope

### SHOULD Use This Skill When
- Create an ECS instance

### SHOULD NOT Use This Skill When
- Manage RDS databases
`
	if err := os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte(skillMd), 0o644); err != nil {
		t.Fatal(err)
	}
	eval := `[
  {"query":"create an ECS instance","should_trigger":true,"skill":"ve-x-ops","confidence":"high"},
  {"query":"create ECS instance now","should_trigger":true,"skill":"ve-x-ops","confidence":"high"},
  {"query":"create ECS cluster","should_trigger":true,"skill":"ve-x-ops","confidence":"high"},
  {"query":"create ECS instance fast","should_trigger":true,"skill":"ve-x-ops","confidence":"high"},
  {"query":"manage RDS database","should_trigger":false,"skill":"ve-x-ops","confidence":"medium"},
  {"query":"list RDS instances","should_trigger":false,"skill":"ve-x-ops","confidence":"low"},
  {"query":"backup RDS now","should_trigger":false,"skill":"ve-x-ops","confidence":"medium"},
  {"query":"tweak kafka topic config unrelated","should_trigger":true,"skill":"ve-x-ops","confidence":"high"}
]`
	if err := os.WriteFile(filepath.Join(skill, "assets", "eval_queries.json"), []byte(eval), 0o644); err != nil {
		t.Fatal(err)
	}
	name, errs := checkSkill(skill)
	if len(errs) == 0 {
		t.Fatalf("%s: expected a low-match error for the kafka query", name)
	}
}

// buildGitRepo creates a throwaway git repo with one skill and returns root.
func buildGitRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	run("config", "user.email", "t@t")
	run("config", "user.name", "t")
	run("config", "commit.gpgsign", "false")
	run("checkout", "-b", "main")
	return root
}

func TestCheckDeltaGitDrift(t *testing.T) {
	root := buildGitRepo(t)
	skill := filepath.Join(root, "ve-x-ops")
	if err := os.MkdirAll(filepath.Join(skill, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	oldScope := `## Trigger & Scope

### SHOULD Use This Skill When
- Create an ECS instance
- Resize the ECS cluster
- Scale the ECS node group

### SHOULD NOT Use This Skill When
- Manage RDS databases
`
	skillMd := "---\nname: ve-x-ops\n---\n" + oldScope + "\n# ve-x-ops\n"
	if err := os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte(skillMd), 0o644); err != nil {
		t.Fatal(err)
	}
	eval := `[
  {"query":"create ECS instance for project","should_trigger":true,"skill":"ve-x-ops","confidence":"high"},
  {"query":"resize ECS cluster size","should_trigger":true,"skill":"ve-x-ops","confidence":"high"},
  {"query":"scale ECS node group count","should_trigger":true,"skill":"ve-x-ops","confidence":"high"},
  {"query":"stop ECS instance","should_trigger":true,"skill":"ve-x-ops","confidence":"medium"},
  {"query":"manage RDS database","should_trigger":false,"skill":"ve-x-ops","confidence":"low"},
  {"query":"list RDS instances","should_trigger":false,"skill":"ve-x-ops","confidence":"low"},
  {"query":"backup RDS now","should_trigger":false,"skill":"ve-x-ops","confidence":"medium"},
  {"query":"delete ECS instance","should_trigger":true,"skill":"ve-x-ops","confidence":"high"}
]`
	if err := os.WriteFile(filepath.Join(skill, "assets", "eval_queries.json"), []byte(eval), 0o644); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	run("add", "-A")
	run("commit", "-q", "-m", "base")

	// Now drift the scope: drop the scale node group bullet.
	newScope := `## Trigger & Scope

### SHOULD Use This Skill When
- Create an ECS instance
- Resize the ECS cluster

### SHOULD NOT Use This Skill When
- Manage RDS databases
`
	skillMd = "---\nname: ve-x-ops\n---\n" + newScope + "\n# ve-x-ops\n"
	if err := os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte(skillMd), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "-A")
	run("commit", "-q", "-m", "drift")

	// getOldScope must recover the pre-drift scope from git history.
	old := getOldScope(root, skill, "HEAD~1")
	if old == "" {
		t.Fatal("getOldScope returned empty")
	}
	if !containsStr(old, "Scale the ECS node group") {
		t.Fatalf("getOldScope missing dropped bullet: %q", old)
	}

	// checkDelta must flag the dropped 'scale ECS node group' coverage.
	var data []map[string]any
	if err := jsonUnmarshalFile(filepath.Join(skill, "assets", "eval_queries.json"), &data); err != nil {
		t.Fatal(err)
	}
	delta := checkDelta("ve-x-ops", data, old, newScope)
	found := false
	for _, e := range delta {
		if containsStr(e, "[GIT-DIFF]") && containsStr(e, "scale ECS node group") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected GIT-DIFF regression for dropped 'scale ECS node group', got %v", delta)
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func jsonUnmarshalFile(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

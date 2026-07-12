package frontmatter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validSkill = `---
name: ve-ecs-ops
description: Manages ECS instances on Volcengine.
compatibility: Volcengine ECS
metadata:
  type: ve-product-skill
  cli_applicability: dual-path
  version: 1.2.1
  last_updated: 2026-07-11
---

# ve-ecs-ops
`

const missingName = `---
description: x
compatibility: y
metadata:
  version: 1.0.0
  last_updated: 2026-07-11
---
`

const badCLI = `---
name: ve-foo-ops
description: x
compatibility: y
metadata:
  cli_applicability: wrong
  version: 1.0.0
  last_updated: 2026-07-11
---
`

func TestExtract(t *testing.T) {
	block, errs := Extract(validSkill)
	if errs != nil {
		t.Fatalf("unexpected errs: %v", errs)
	}
	if !strings.Contains(block, "name: ve-ecs-ops") {
		t.Fatalf("block missing name line: %q", block)
	}
	if _, errs := Extract("no frontmatter here"); errs == nil {
		t.Fatal("expected missing-frontmatter error")
	}
}

func TestValidateSkill(t *testing.T) {
	if e := ValidateSkill("ve-ecs-ops", mustBlock(t, validSkill)); e != nil {
		t.Fatalf("valid skill should have no errors, got %v", e)
	}
	if e := ValidateSkill("ve-foo-ops", mustBlock(t, missingName)); len(e) == 0 {
		t.Fatal("expected missing-name error")
	} else if !contains(e, "missing 'name'") {
		t.Fatalf("expected missing name error, got %v", e)
	}
	if e := ValidateSkill("ve-foo-ops", mustBlock(t, badCLI)); !contains(e, "invalid cli_applicability 'wrong'") {
		t.Fatalf("expected invalid cli_applicability, got %v", e)
	}
}

func TestCheckDir(t *testing.T) {
	dir := t.TempDir()
	mkSkill(t, dir, "ve-good-ops", validSkill)
	mkSkill(t, dir, "ve-bad-ops", missingName)
	results, skills := CheckDir(dir)
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}
	if _, ok := results[filepath.Join(dir, "ve-bad-ops", "SKILL.md")]; !ok {
		t.Fatalf("expected ve-bad-ops errors, got %v", results)
	}
	if _, ok := results[filepath.Join(dir, "ve-good-ops", "SKILL.md")]; ok {
		t.Fatalf("ve-good-ops should have no errors")
	}
}

func mustBlock(t *testing.T, doc string) string {
	t.Helper()
	b, errs := Extract(doc)
	if errs != nil {
		t.Fatalf("fixture parse failed: %v", errs)
	}
	return b
}

func mkSkill(t *testing.T, root, name, doc string) {
	t.Helper()
	p := filepath.Join(root, name)
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p, "SKILL.md"), []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}
}

func contains(s []string, sub string) bool {
	for _, v := range s {
		if v == sub {
			return true
		}
	}
	return false
}

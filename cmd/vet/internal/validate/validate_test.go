package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeSkill writes a SKILL.md (and optional extra files) under a temp root and
// returns the root path.
func writeSkill(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

const goodSkill = `---
name: ve-example-ops
description: example
compatibility: ["ve-cli"]
metadata:
  type: product
  cli_applicability: dual-path
  version: "1.0"
  last_updated: "2026-01-01"
---

## Trigger & Scope

### SHOULD Use This Skill When
x

### SHOULD NOT Use This Skill When
y

## Quality Gate (GCL)

### What This Skill Does
does the thing

## Operational Best Practices
keep it safe

### Next Steps
done

## Error Taxonomy

| ` + "`ERR1`" + ` | desc | **HALT** |
| ` + "`ERR2`" + ` | desc | **RETRY** |
| ` + "`ERR3`" + ` | desc | **HALT** |
| ` + "`ERR4`" + ` | desc | **RETRY** |
| ` + "`ERR5`" + ` | desc | **HALT** |
| ` + "`ERR6`" + ` | desc | **RETRY** |
| ` + "`ERR7`" + ` | desc | **HALT** |
| ` + "`ERR8`" + ` | desc | **RETRY** |
| ` + "`ERR9`" + ` | desc | **HALT** |
| ` + "`ERR10`" + ` | desc | **RETRY** |
`

func TestCheckFileIntegrity(t *testing.T) {
	root := writeSkill(t, map[string]string{"ve-good/SKILL.md": goodSkill})
	if errs := checkFileIntegrity(root); len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}

	root2 := writeSkill(t, map[string]string{"ve-bad/SKILL.md": "hello\x00world"})
	errs := checkFileIntegrity(root2)
	if len(errs) != 1 || !containsStr(errs[0], "contains null bytes") {
		t.Fatalf("expected null byte error, got %v", errs)
	}
}

func TestCheckRequiredSections(t *testing.T) {
	root := writeSkill(t, map[string]string{"ve-good/SKILL.md": goodSkill})
	if errs := checkRequiredSections(root); len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}

	// meta-skill must be skipped even if missing sections
	rootMeta := writeSkill(t, map[string]string{"ve-skill-generator/SKILL.md": "no sections here"})
	if errs := checkRequiredSections(rootMeta); len(errs) != 0 {
		t.Fatalf("meta-skill should be skipped, got %v", errs)
	}

	bad := "name: ve-bad\n\n# No trigger section\n"
	rootBad := writeSkill(t, map[string]string{"ve-bad/SKILL.md": bad})
	errs := checkRequiredSections(rootBad)
	if len(errs) == 0 {
		t.Fatal("expected missing-section errors")
	}
	if !containsStr(errs[0], "missing ## Trigger & Scope") {
		t.Fatalf("expected trigger error, got %v", errs)
	}
}

func TestCheckErrorTaxonomy(t *testing.T) {
	root := writeSkill(t, map[string]string{"ve-good/SKILL.md": goodSkill})
	if w := checkErrorTaxonomy(root); len(w) != 0 {
		t.Fatalf("expected no taxonomy warnings, got %v", w)
	}

	// missing taxonomy → advisory warning
	rootMissing := writeSkill(t, map[string]string{"ve-bad/SKILL.md": goodSkillNoTaxonomy()})
	if w := checkErrorTaxonomy(rootMissing); len(w) == 0 || !containsStr(w[0], "missing ## Error Taxonomy") {
		t.Fatalf("expected missing-taxonomy warning, got %v", w)
	}
}

func TestCheckTE1Hardcodes(t *testing.T) {
	// Hardcoded EngineVersion inside a JSON body should be flagged.
	src := "do something\n\"EngineVersion\": \"5.0\"\nrest"
	root := writeSkill(t, map[string]string{"ve-bad/SKILL.md": src})
	w := checkTE1Hardcodes(root)
	if len(w) == 0 {
		t.Fatal("expected a TE-1 hardcoded version warning")
	}
	if !containsStr(w[0], "TE-1 hardcoded EngineVersion") {
		t.Fatalf("expected EngineVersion flag, got %v", w)
	}

	// No hardcodes → clean.
	rootClean := writeSkill(t, map[string]string{"ve-good/SKILL.md": goodSkill})
	if w := checkTE1Hardcodes(rootClean); len(w) != 0 {
		t.Fatalf("expected no TE-1 warnings, got %v", w)
	}
}

func TestBuildStepsPureGo(t *testing.T) {
	// The validation suite must not depend on python3 — every step runs
	// in-process Go. Assert the step list is non-empty and that none of
	// the step names reference an external python interpreter.
	steps := buildSteps()
	if len(steps) == 0 {
		t.Fatal("expected non-empty step list")
	}
	for _, s := range steps {
		if s.run == nil {
			t.Fatalf("step %q has nil run func", s.name)
		}
	}
	// Regression guard: the suite used to shell out to python3 scripts.
	// The unit-test step exercises `go test`, never `python3`.
	for _, s := range steps {
		if strings.Contains(s.name, "python3") {
			t.Fatalf("step %q still references python3", s.name)
		}
	}
}

func TestRunListOnlyDoesNotExecute(t *testing.T) {
	root := writeSkill(t, map[string]string{"ve-good/SKILL.md": goodSkill})
	_, _, total := Run(root, true)
	if total != len(buildSteps()) {
		t.Fatalf("listOnly total %d != buildSteps %d", total, len(buildSteps()))
	}
}

func goodSkillNoTaxonomy() string {
	return `---
name: ve-bad
description: x
compatibility: ["ve-cli"]
metadata:
  type: product
  cli_applicability: dual-path
  version: "1.0"
  last_updated: "2026-01-01"
---

## Trigger & Scope
### SHOULD Use This Skill When
x
### SHOULD NOT Use This Skill When
y
## Quality Gate (GCL)
### What This Skill Does
does
## Operational Best Practices
safe
### Next Steps
done
`
}

func containsStr(s, sub string) bool {
	return strings.Contains(s, sub)
}

package gcl

import (
	"os"
	"path/filepath"
	"testing"
)

// validRubric has exactly sections 0..7 (contiguous, gap-free).
const validRubric = `# Rubric

## 0. Operation Tier
x

## 1. Correctness
x

## 2. Safety
x

## 3. Idempotency
x

## 4. Traceability
x

## 5. Spec Compliance
x

## 6. Thresholds
x

## 7. Notes
x
`

// validPrompt has exactly sections 1..5 with the required generator/critic prompts.
const validPrompt = `# Prompt Templates

## 1. Generator Prompt
x

## 2. Critic Prompt
x

## 3. Orchestrator Prompt
x

## 4. Example
x

## 5. Notes
x
`

const validSkill = `# ve-ecs-ops

## Quality Gate (GCL)

x
`

// missingSectionRubric drops section 3 — countNumberedSections must return 0.
const missingSectionRubric = `# Rubric

## 0. Operation Tier
x

## 1. Correctness
x

## 2. Safety
x

## 4. Traceability
x

## 5. Spec Compliance
x

## 6. Thresholds
x

## 7. Notes
x
`

// promptMissingCritic has all 5 numbered sections but lacks the Critic Prompt.
const promptMissingCritic = `# Prompt Templates

## 1. Generator Prompt
x

## 2. Foo
x

## 3. Orchestrator Prompt
x

## 4. Example
x

## 5. Notes
x
`

func TestCountNumberedSections(t *testing.T) {
	if n := countNumberedSections(validRubric, 0, 7); n != 8 {
		t.Fatalf("valid rubric: want 8, got %d", n)
	}
	if n := countNumberedSections(missingSectionRubric, 0, 7); n != 0 {
		t.Fatalf("missing section rubric: want 0, got %d", n)
	}
	if n := countNumberedSections(validPrompt, 1, 5); n != 5 {
		t.Fatalf("valid prompt: want 5, got %d", n)
	}
}

func TestCheckSkill(t *testing.T) {
	tests := []struct {
		name       string
		rubric     string
		prompt     string
		skill      string
		wantOK     bool
		wantRubric int
		wantPrompt int
	}{
		{"valid", validRubric, validPrompt, validSkill, true, 8, 5},
		{"missing rubric section", missingSectionRubric, validPrompt, validSkill, false, 0, 5},
		{"prompt missing critic", validRubric, promptMissingCritic, validSkill, false, 8, 5},
		{"missing quality gate", validRubric, validPrompt, "# ve-ecs-ops\n", false, 8, 5},
		{"empty everything", "", "", "", false, 0, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			sk := filepath.Join(dir, "ve-test-ops")
			if err := os.MkdirAll(filepath.Join(sk, "references"), 0o755); err != nil {
				t.Fatal(err)
			}
			writeFile(t, filepath.Join(sk, "references", "rubric.md"), tc.rubric)
			writeFile(t, filepath.Join(sk, "references", "prompt-templates.md"), tc.prompt)
			writeFile(t, filepath.Join(sk, "SKILL.md"), tc.skill)

			r := checkSkill(dir, "ve-test-ops")
			if r.ok != tc.wantOK {
				t.Fatalf("ok: want %v, got %v (err reasons not asserted here)", tc.wantOK, r.ok)
			}
			if r.rubricSections != tc.wantRubric {
				t.Fatalf("rubricSections: want %d, got %d", tc.wantRubric, r.rubricSections)
			}
			if r.promptSections != tc.wantPrompt {
				t.Fatalf("promptSections: want %d, got %d", tc.wantPrompt, r.promptSections)
			}
		})
	}
}

func TestCheckDir(t *testing.T) {
	// Build a root containing only the fixed canonical skill set is large;
	// instead verify the function returns sorted skill names and an empty
	// error map when all artifacts are present (we stub a single skill by
	// temporarily overriding the set).
	orig := gclSkills
	gclSkills = map[string]bool{"ve-test-ops": true}
	defer func() { gclSkills = orig }()

	dir := t.TempDir()
	sk := filepath.Join(dir, "ve-test-ops")
	if err := os.MkdirAll(filepath.Join(sk, "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(sk, "references", "rubric.md"), validRubric)
	writeFile(t, filepath.Join(sk, "references", "prompt-templates.md"), validPrompt)
	writeFile(t, filepath.Join(sk, "SKILL.md"), validSkill)

	results, skills := CheckDir(dir)
	if len(skills) != 1 || skills[0] != "ve-test-ops" {
		t.Fatalf("skills: want [ve-test-ops], got %v", skills)
	}
	if len(results) != 0 {
		t.Fatalf("expected no errors, got %v", results)
	}
}

func writeFile(t *testing.T, p, content string) {
	t.Helper()
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

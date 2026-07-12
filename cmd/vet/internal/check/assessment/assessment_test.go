package assessment

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// jsonFence wraps a JSON body in a markdown ```json block. Used to avoid
// backticks inside raw string literals.
func jsonFence(body string) string {
	return "```json\n" + body + "\n```"
}

const headerValid = "# Worker Output Contract\n\nExample product_assessment:\n\n"

const validJSON = `{
  "skill_id": "ve-ecs-ops",
  "product": "ecs",
  "region": "cn-beijing",
  "scope": "x",
  "assessment_date": "2026-07-11",
  "status": "OK",
  "partial": false,
  "resource_count": 1,
  "pillars": {
    "reliability": {
      "status": "assessed",
      "findings": [
        {"id": "ecs-rel-001", "severity": "High", "confidence": "HIGH", "title": "t", "evidence": "e", "recommendation": "r", "effort": "quick"}
      ]
    }
  },
  "recommendations": [{"pillar": "reliability", "title": "t"}],
  "trace": {"commands": ["ve ecs DescribeInstances --Region cn-beijing"]},
  "errors": []
}`

const missingJSON = `{
  "product": "ecs",
  "pillars": {}
}`

const badJSON = `{
  "skill_id": "s",
  "product": "ecs",
  "region": "x",
  "scope": "x",
  "assessment_date": "2026-07-11",
  "status": "OK",
  "partial": false,
  "resource_count": 1,
  "pillars": {
    "security": {
      "status": "assessed",
      "findings": [
        {"id": "wrong", "severity": "Urgent", "confidence": "LOW", "title": "t", "evidence": "e", "recommendation": "r", "effort": "huge"}
      ]
    }
  },
  "recommendations": [{"pillar": "nope"}],
  "trace": {"commands": ["ve ecs X AKLTabcdef<masked>"]},
  "errors": []
}`

func validExample() string   { return headerValid + jsonFence(validJSON) + "\n" }
func missingExample() string { return headerValid + jsonFence(missingJSON) + "\n" }
func badExample() string     { return headerValid + jsonFence(badJSON) + "\n" }
func noContractExample() string {
	return "# Something else\n\n" + jsonFence(missingJSON) + "\n"
}
func noExampleDoc() string { return "# Worker Output Contract\n\nNo json block here.\n" }

func writeAssessment(t *testing.T, root, skill, doc string) string {
	t.Helper()
	p := filepath.Join(root, skill, "references")
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	fp := filepath.Join(p, "well-architected-assessment.md")
	if err := os.WriteFile(fp, []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}
	return fp
}

func TestValidateAssessment(t *testing.T) {
	if e := ValidateAssessment(mustObj(t, validExample()), "valid"); len(e) > 0 {
		t.Fatalf("valid example should pass, got %v", e)
	}
	if e := ValidateAssessment(mustObj(t, missingExample()), "missing"); len(e) == 0 {
		t.Fatal("expected missing top-level fields error")
	}
	if e := ValidateAssessment(mustObj(t, badExample()), "bad"); len(e) == 0 {
		t.Fatal("expected finding/pillar errors")
	}
}

func TestCheckDir(t *testing.T) {
	dir := t.TempDir()
	good := writeAssessment(t, dir, "ve-good-ops", validExample())
	bad := writeAssessment(t, dir, "ve-bad-ops", missingExample())
	noContract := writeAssessment(t, dir, "ve-nocontract-ops", noContractExample())
	none := writeAssessment(t, dir, "ve-none-ops", noExampleDoc())

	errs, files, examples := CheckDir(dir)
	if files != 4 {
		t.Fatalf("expected 4 files, got %d", files)
	}
	has := func(sub string) bool {
		for _, e := range errs {
			if strings.Contains(e, sub) {
				return true
			}
		}
		return false
	}
	if has(good) {
		t.Fatalf("ve-good-ops should have no errors, got %v", errs)
	}
	if !has(bad) {
		t.Fatal("ve-bad-ops should have errors")
	}
	if !has(noContract) {
		t.Fatal("ve-nocontract-ops should have missing-section error")
	}
	if !has(none) {
		t.Fatal("ve-none-ops should have no-example error")
	}
	_ = examples
}

func mustObj(t *testing.T, doc string) map[string]any {
	t.Helper()
	for _, raw := range extractExamples(doc) {
		var v any
		if err := json.Unmarshal([]byte(raw), &v); err == nil {
			if m, ok := v.(map[string]any); ok {
				return m
			}
		}
	}
	t.Fatal("fixture had no extractable product_assessment JSON")
	return nil
}

func contains(s []string, sub string) bool {
	for _, v := range s {
		if strings.Contains(v, sub) {
			return true
		}
	}
	return false
}

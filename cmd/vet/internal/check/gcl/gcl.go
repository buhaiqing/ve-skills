// Package gcl validates the Generator-Critic-Loop (GCL) artifact set that
// each ve-* skill must ship, per AGENTS.md §8 / §10.
//
// It is a faithful Go port of scripts/check_gcl_conformance.py.
package gcl

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

// gclSkills is the canonical 29-skill set declared in AGENTS.md §8.
// Order is irrelevant (we always sort on output); a set prevents accidental
// mutation and keeps the membership fixed.
var gclSkills = map[string]bool{
	"ve-ecs-ops":            true,
	"ve-redis-ops":          true,
	"ve-rds-mysql-ops":      true,
	"ve-rds-ops":            true,
	"ve-rds-pg-ops":         true,
	"ve-polar-mysql-ops":    true,
	"ve-mongodb-ops":        true,
	"ve-elasticsearch-ops":  true,
	"ve-tos-ops":            true,
	"ve-iam-ops":            true,
	"ve-kms-ops":            true,
	"ve-eip-ops":            true,
	"ve-security-group-ops": true,
	"ve-vpc-ops":            true,
	"ve-nat-ops":            true,
	"ve-vpn-ops":            true,
	"ve-clb-ops":            true,
	"ve-alb-ops":            true,
	"ve-vke-ops":            true,
	"ve-nas-ops":            true,
	"ve-cms-ops":            true,
	"ve-fg-ops":             true,
	"ve-ark-ops":            true,
	"ve-cdn-ops":            true,
	"ve-dns-ops":            true,
	"ve-kafka-ops":          true,
	"ve-sls-ops":            true,
	"ve-billing-ops":        true,
	"ve-skill-generator":    true,
	// Orchestration / loop agents (not product-bound).
	"incident-loop-agent": true,
}

// HumanFmt is the human-readable summary format produced by the Python source.
// The main agent handles printing; this constant keeps the format discoverable.
const HumanFmt = "GCL conformance: {passing}/{total} skills conform."

var (
	qualityGateRe = regexp.MustCompile(`(?m)^## Quality Gate \(GCL\)$`)
	opTierRe      = regexp.MustCompile(`(?m)^## 0\. Operation Tier`)
	safetyRe      = regexp.MustCompile(`(?m)^## 2\. Safety`)
	genPromptRe   = regexp.MustCompile(`(?m)^## 1\. Generator Prompt`)
	critPromptRe  = regexp.MustCompile(`(?m)^## 2\. Critic Prompt`)
)

// countNumberedSections returns the count of sections if all start..end are
// present, else 0. The `\. ` after the number guards against false matches
// like "## 1.0" or "## 10." — only single-digit ordinals count.
func countNumberedSections(text string, start, end int) int {
	if end < start {
		return 0
	}
	for n := start; n <= end; n++ {
		if !regexp.MustCompile(`(?m)^## ` + itoa(n) + `\. `).MatchString(text) {
			return 0
		}
	}
	return end - start + 1
}

func itoa(n int) string {
	return string(rune('0' + n))
}

// report is the per-skill conformance result (mirrors the Python dict).
type report struct {
	skill              string
	rubricSections     int
	promptSections     int
	hasQualityGate     bool
	hasOperationTier   bool
	hasSafetyRules     bool
	hasGeneratorPrompt bool
	hasCriticPrompt    bool
	rubricOK           bool
	promptOK           bool
	skillOK            bool
	ok                 bool
}

// checkSkill builds the per-skill conformance report.
func checkSkill(root, skill string) report {
	skillDir := filepath.Join(root, skill)
	rubricPath := filepath.Join(skillDir, "references", "rubric.md")
	promptPath := filepath.Join(skillDir, "references", "prompt-templates.md")
	skillPath := filepath.Join(skillDir, "SKILL.md")

	var r report
	r.skill = skill

	var rubricSections int
	hasOperationTier := false
	hasSafetyRules := false
	if b, err := os.ReadFile(rubricPath); err == nil {
		text := string(b)
		rubricSections = countNumberedSections(text, 0, 7)
		hasOperationTier = opTierRe.MatchString(text)
		hasSafetyRules = safetyRe.MatchString(text)
	}

	var promptSections int
	hasGeneratorPrompt := false
	hasCriticPrompt := false
	if b, err := os.ReadFile(promptPath); err == nil {
		text := string(b)
		promptSections = countNumberedSections(text, 1, 5)
		hasGeneratorPrompt = genPromptRe.MatchString(text)
		hasCriticPrompt = critPromptRe.MatchString(text)
	}

	hasQualityGate := false
	if b, err := os.ReadFile(skillPath); err == nil {
		hasQualityGate = qualityGateRe.MatchString(string(b))
	}

	rubricOK := rubricSections == 8 && hasOperationTier && hasSafetyRules
	promptOK := promptSections == 5 && hasGeneratorPrompt && hasCriticPrompt
	skillOK := hasQualityGate

	r.rubricSections = rubricSections
	r.promptSections = promptSections
	r.hasQualityGate = hasQualityGate
	r.hasOperationTier = hasOperationTier
	r.hasSafetyRules = hasSafetyRules
	r.hasGeneratorPrompt = hasGeneratorPrompt
	r.hasCriticPrompt = hasCriticPrompt
	r.rubricOK = rubricOK
	r.promptOK = promptOK
	r.skillOK = skillOK
	r.ok = rubricOK && promptOK && skillOK
	return r
}

// CheckDir validates every skill in the canonical set under root and returns a
// per-skill error map (only failing skills are present) plus the sorted skill
// list. The signature mirrors the frontmatter package's CheckDir.
func CheckDir(root string) (map[string][]string, []string) {
	skills := sortedSkillNames()
	results := make(map[string][]string)
	for _, sk := range skills {
		r := checkSkill(root, sk)
		if r.ok {
			continue
		}
		var errs []string
		if !r.rubricOK {
			errs = append(errs, rubricReason(r))
		}
		if !r.promptOK {
			errs = append(errs, promptReason(r))
		}
		if !r.skillOK {
			errs = append(errs, "missing `## Quality Gate (GCL)` in SKILL.md")
		}
		results[sk] = errs
	}
	return results, skills
}

// rubricReason mirrors the Python "rubric_sections=N/8 (0-7)" failure reason.
func rubricReason(r report) string {
	return "rubric_sections=" + itoa(r.rubricSections) + "/8 (0-7)"
}

// promptReason mirrors the Python "prompt_sections=N/5 (1-5)" failure reason.
func promptReason(r report) string {
	return "prompt_sections=" + itoa(r.promptSections) + "/5 (1-5)"
}

func sortedSkillNames() []string {
	out := make([]string, 0, len(gclSkills))
	for s := range gclSkills {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// PassingCount returns the number of skills in reports that conform.
func PassingCount(results map[string][]string, skills []string) int {
	return len(skills) - len(results)
}

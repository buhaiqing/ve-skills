// Package validate is the pure-Go implementation of the local validation
// suite formerly driven by scripts/validate_local.py.
//
// The original Python built a list of Steps, where some ran external scripts
// (validate_skills_frontmatter.py, check_markdown_links.py, ...) and several
// generated inline Python executed via `python3 -c`. This Go port runs every
// check in-process by calling the equivalent `vet check` / `vet gcl` packages
// directly — no python3 dependency. The four checks the Python original
// implemented as inline Python (file integrity, required sections, error
// taxonomy, TE-1 hardcodes) are implemented here as pure-Go helpers and wired
// in as steps (see engine.go buildSteps).
package validate

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// skillDirs returns sorted paths of every ve-*/SKILL.md under root.
func skillDirs(root string) []string {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "ve-") {
			out = append(out, filepath.Join(root, e.Name(), "SKILL.md"))
		}
	}
	sort.Strings(out)
	return out
}

// readText reads a file as UTF-8 text, returning "" on error.
func readText(p string) string {
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return string(b)
}

// checkFileIntegrity: every ve-*/SKILL.md must be null-byte free.
func checkFileIntegrity(root string) []string {
	var errs []string
	for _, f := range skillDirs(root) {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		if bytesContainNull(data) {
			errs = append(errs, f+": contains null bytes")
		}
	}
	return errs
}

func bytesContainNull(b []byte) bool {
	for _, c := range b {
		if c == 0 {
			return true
		}
	}
	return false
}

// checkRequiredSections. ve-skill-generator is a meta-skill and is skipped.
// Errors are returned; missing "### Next Steps" is advisory (warning) and not
// returned here.
func checkRequiredSections(root string) []string {
	var errs []string
	for _, f := range skillDirs(root) {
		skill := filepath.Base(filepath.Dir(f))
		if skill == "ve-skill-generator" {
			continue
		}
		text := readText(f)

		hasTS := strings.Contains(text, "## Trigger & Scope") ||
			strings.Contains(text, "## Trigger & Scope (Agent-Readable)")
		hasShall := strings.Contains(text, "### SHOULD Use This Skill When")
		hasShallNot := strings.Contains(text, "### SHOULD NOT Use This Skill When")
		hasGCL := strings.Contains(text, "## Quality Gate (GCL)")
		hasWhat := strings.Contains(text, "### What This Skill Does")
		hasOps := strings.Contains(text, "## Operational Best Practices")

		if !hasTS {
			errs = append(errs, skill+": missing ## Trigger & Scope")
		} else if !hasShall || !hasShallNot {
			errs = append(errs, skill+": ## Trigger & Scope lacks SHOULD/SHOULD NOT subsections")
		}
		if !hasGCL {
			errs = append(errs, skill+": missing ## Quality Gate (GCL)")
		}
		if !hasWhat {
			errs = append(errs, skill+": missing ### What This Skill Does (IMPORTANT — MUST exist)")
		}
		if !hasOps {
			errs = append(errs, skill+": missing ## Operational Best Practices (IMPORTANT — MUST exist)")
		}
	}
	return errs
}

// checkErrorTaxonomy is advisory: returns warnings for skills missing a
// sufficient ## Error Taxonomy (≥10 codes incl. HALT/RETRY).
func checkErrorTaxonomy(root string) []string {
	re := regexp.MustCompile(`(?m)^\s*\|[ ]*` + "`" + `[^` + "`" + `]+` + "`" + `[ ]*\|[ ]*[^|]+?\|[ ]*[^|]*?\*\*(HALT|RETRY)\*\*`)
	var warns []string
	for _, f := range skillDirs(root) {
		skill := filepath.Base(filepath.Dir(f))
		if skill == "ve-skill-generator" {
			continue
		}
		text := readText(f)
		if !strings.Contains(text, "## Error Taxonomy") {
			warns = append(warns, skill+": missing ## Error Taxonomy")
			continue
		}
		codes := re.FindAllStringSubmatch(text, -1)
		var classes []string
		for _, m := range codes {
			classes = append(classes, m[1])
		}
		if len(classes) < 10 {
			warns = append(warns, skill+": ## Error Taxonomy has only "+itoa(len(classes))+" codes, need ≥10")
		} else if !contains(classes, "HALT") {
			warns = append(warns, skill+": ## Error Taxonomy missing HALT classification")
		} else if !contains(classes, "RETRY") {
			warns = append(warns, skill+": ## Error Taxonomy missing RETRY classification")
		}
	}
	return warns
}

// te1Patterns mirrors the PATTERNS list in the original _check_te1_hardcodes.
var te1Patterns = []struct{ label, pattern string }{
	{"EngineVersion", `"EngineVersion":\s*"\d+\.\d+"`},
	{"MongoVersion", `"MongoVersion":\s*"\d+\.\d+"`},
	{"--MongoVersion", `--MongoVersion\s+\d+\.\d+`},
	{"--Version", `--Version\s+"\d+\.\d+"`},
	{"--TargetVersion", `--TargetVersion\s+"\d+\.\d+"`},
}

// checkTE1Hardcodes is advisory: returns candidate hardcoded-version warnings.
func checkTE1Hardcodes(root string) []string {
	var warns []string
	for _, glob := range []string{"ve-*/references/cli-usage.md", "ve-*/SKILL.md"} {
		matches, _ := filepath.Glob(filepath.Join(root, glob))
		sort.Strings(matches)
		for _, f := range matches {
			text := readText(f)
			rel, _ := filepath.Rel(root, f)
			for _, tp := range te1Patterns {
				re := regexp.MustCompile(tp.pattern)
				for _, m := range re.FindAllString(text, -1) {
					warns = append(warns, rel+": TE-1 hardcoded "+tp.label+" → "+m)
				}
			}
		}
	}
	return warns
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

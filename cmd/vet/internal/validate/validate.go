// Package validate is a faithful Go port of scripts/validate_local.py.
//
// The Python original builds a list of Steps, where some steps run external
// scripts (validate_skills_frontmatter.py, check_markdown_links.py, ...) and
// several steps dynamically generate inline Python source and execute it via
// `python3 -c`. This port preserves that "execute external command" behavior:
// the four checks that the original implemented as inline Python
// (_check_file_integrity, _check_required_sections, _check_error_taxonomy,
// _check_te1_hardcodes) are reimplemented as pure-Go helpers so their output
// matches the original, and the script-based steps are delegated to `python3`.
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

// CheckFileIntegrity mirrors _check_file_integrity: every ve-*/SKILL.md must
// be null-byte free.
func CheckFileIntegrity(root string) []string {
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

// CheckRequiredSections mirrors _check_required_sections. ve-skill-generator
// is a meta-skill and is skipped. Errors are returned; missing "### Next Steps"
// is advisory (warning) and not returned here.
func CheckRequiredSections(root string) []string {
	hard := []string{
		"## Trigger & Scope",
		"## Quality Gate (GCL)",
		"### What This Skill Does",
		"## Operational Best Practices",
		"### Next Steps",
	}
	_ = hard
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
		hasNext := strings.Contains(text, "### Next Steps")

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
		_ = hasNext
	}
	return errs
}

// CheckErrorTaxonomy mirrors _check_error_taxonomy. This is advisory in the
// original (returns 0 regardless); returns warnings here so callers may decide.
func CheckErrorTaxonomy(root string) []string {
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

// te1Patterns mirrors the PATTERNS list in _check_te1_hardcodes.
var te1Patterns = []struct{ label, pattern string }{
	{"EngineVersion", `"EngineVersion":\s*"\d+\.\d+"`},
	{"MongoVersion", `"MongoVersion":\s*"\d+\.\d+"`},
	{"--MongoVersion", `--MongoVersion\s+\d+\.\d+`},
	{"--Version", `--Version\s+"\d+\.\d+"`},
	{"--TargetVersion", `--TargetVersion\s+"\d+\.\d+"`},
}

// CheckTE1Hardcodes mirrors _check_te1_hardcodes. Advisory (returns 0 in the
// original); returns candidate warnings here.
func CheckTE1Hardcodes(root string) []string {
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

// Package frontmatter validates the YAML frontmatter of ve-* skill SKILL.md files.
//
// It is a faithful Go port of scripts/validate_skills_frontmatter.py.
package frontmatter

import (
	"regexp"
	"sort"
	"strings"
)

// CLIApplicability are the permitted values for metadata.cli_applicability.
var CLIApplicability = map[string]bool{
	"dual-path": true,
	"cli-first": true,
	"cli-only":  true,
	"sdk-only":  true,
}

// OrchestrationTypes are metadata.type values exempt from the product-skill
// conventions (ve- prefix, cli_applicability).
var OrchestrationTypes = map[string]bool{
	"orchestration-skill": true,
	"loop-agent":          true,
	"meta-skill":          true,
}

var frontmatterRe = regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---`)

// Extract returns the inner frontmatter block and any errors.
func Extract(text string) (string, []string) {
	m := frontmatterRe.FindStringSubmatch(text)
	if m == nil {
		return "", []string{"missing YAML frontmatter"}
	}
	return m[1], nil
}

func hasKey(block, key string) bool {
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + `:\s`)
	return re.MatchString(block)
}

func field(block, key string) string {
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + `:\s*["']?([^"'\n]+)`)
	m := re.FindStringSubmatch(block)
	if m == nil {
		return ""
	}
	return strings.Trim(m[1], `"'`)
}

// NestedMetadataField reads metadata.<field> (indented under metadata:).
func NestedMetadataField(block, field string) string {
	if !hasKey(block, "metadata") {
		return ""
	}
	re := regexp.MustCompile(`(?m)^\s+` + regexp.QuoteMeta(field) + `:\s*["']?([^"'\n]+)`)
	m := re.FindStringSubmatch(block)
	if m == nil {
		return ""
	}
	return strings.Trim(m[1], `"'`)
}

// ValidateSkill validates a single SKILL.md frontmatter block and returns errors.
func ValidateSkill(name string, block string) []string {
	var errs []string
	skillType := NestedMetadataField(block, "type")
	isOrchestration := OrchestrationTypes[skillType]

	n := field(block, "name")
	if n == "" {
		errs = append(errs, "missing 'name'")
	} else if !strings.HasPrefix(n, "ve-") && !isOrchestration {
		errs = append(errs, "missing or invalid 'name' (must start with ve-, unless metadata.type is orchestration-skill/loop-agent/meta-skill)")
	}

	if !hasKey(block, "description") {
		errs = append(errs, "missing 'description'")
	}
	if !hasKey(block, "compatibility") {
		errs = append(errs, "missing 'compatibility'")
	}

	cli := NestedMetadataField(block, "cli_applicability")
	if cli == "" {
		cli = field(block, "cli_applicability")
	}
	if cli != "" && !CLIApplicability[cli] {
		errs = append(errs, "invalid cli_applicability '"+cli+"'")
	} else if cli == "" && !isOrchestration {
		errs = append(errs, "missing cli_applicability")
	}

	// legacy skills without metadata are exempt from version/last_updated checks
	for legacy := range LegacyNoMetadata {
		if n == legacy {
			return errs
		}
	}
	version := NestedMetadataField(block, "version")
	updated := NestedMetadataField(block, "last_updated")
	if version == "" {
		errs = append(errs, "missing metadata.version")
	}
	if updated == "" {
		errs = append(errs, "missing metadata.last_updated")
	}
	return errs
}

// LegacyNoMetadata lists skill names exempt from metadata.version/last_updated.
var LegacyNoMetadata = map[string]bool{}

// CheckDir validates every ve-*/SKILL.md under root and returns a per-file
// error map (only files with errors are present) plus the sorted skill list.
func CheckDir(root string) (map[string][]string, []string) {
	skills := sortedSkills(root)
	results := make(map[string][]string)
	for _, sk := range skills {
		text := readFile(sk)
		block, errs := Extract(text)
		if errs != nil {
			results[sk] = errs
			continue
		}
		ve := skillName(sk)
		if e := ValidateSkill(ve, block); e != nil {
			results[sk] = e
		}
	}
	return results, skills
}

func sortedSkills(root string) []string {
	entries, err := readDir(root)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "ve-") {
			out = append(out, joinPath(root, e.Name(), "SKILL.md"))
		}
	}
	sort.Strings(out)
	return out
}

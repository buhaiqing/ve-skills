package transpile

import (
	"crypto/md5"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type FailurePattern struct {
	Category string
	Skill    string
	Pattern  string
	Fix      string
	Count    int
}

type Guardrail struct {
	ID          string    `yaml:"id"`
	Skill       string    `yaml:"skill"`
	Trigger     string    `yaml:"trigger"`
	Action      string    `yaml:"action"`
	Severity    string    `yaml:"severity"`
	SourceCount int       `yaml:"source_count"`
	CreatedAt   time.Time `yaml:"created_at"`
}

type guardrailDoc struct {
	Guardrails []Guardrail `yaml:"guardrails"`
}

// hashID returns the first 8 hex characters of MD5(skill+pattern).
func hashID(skill, pattern string) string {
	h := md5.Sum([]byte(skill + pattern))
	return fmt.Sprintf("gr-%x", h[:4])
}

// Transpile converts a single FailurePattern to a Guardrail.
// Returns (Guardrail, true) if count >= 10, (Guardrail{}, false) otherwise.
func Transpile(p FailurePattern) (Guardrail, bool) {
	if p.Count < 10 {
		return Guardrail{}, false
	}

	severity := "low"
	action := "auto-ASK"

	if p.Count >= 30 {
		severity = "high"
		action = "auto-REFUSE"
	} else if p.Count >= 15 {
		severity = "medium"
	}

	return Guardrail{
		ID:          hashID(p.Skill, p.Pattern),
		Skill:       p.Skill,
		Trigger:     p.Pattern,
		Action:      action,
		Severity:    severity,
		SourceCount: p.Count,
		CreatedAt:   time.Now().UTC(),
	}, true
}

// TranspileFile reads a markdown file, extracts failure pattern tables,
// and writes guardrails to a YAML file. Returns the number of guardrails written.
func TranspileFile(patternsPath, outPath string) (int, error) {
	data, err := os.ReadFile(patternsPath)
	if err != nil {
		return 0, err
	}

	patterns := ExtractPatterns(string(data))
	if len(patterns) == 0 {
		return 0, nil
	}

	var guardrails []Guardrail
	for _, p := range patterns {
		if g, ok := Transpile(p); ok {
			guardrails = append(guardrails, g)
		}
	}

	if len(guardrails) == 0 {
		return 0, nil
	}

	doc := guardrailDoc{Guardrails: guardrails}
	out, err := yaml.Marshal(&doc)
	if err != nil {
		return 0, fmt.Errorf("yaml marshal: %w", err)
	}

	if err := os.WriteFile(outPath, out, 0o644); err != nil {
		return 0, fmt.Errorf("write output: %w", err)
	}

	return len(guardrails), nil
}

// ExtractPatterns parses markdown table rows from content and returns
// FailurePattern entries for rows with count > 0.
func ExtractPatterns(content string) []FailurePattern {
	var patterns []FailurePattern
	lines := strings.Split(content, "\n")

	inTable := false
	var header []string
	// map column name (lowercased) → index
	var colIdx map[string]int

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect table rows: must start and end with |
		if !strings.HasPrefix(trimmed, "|") || !strings.HasSuffix(trimmed, "|") {
			inTable = false
			header = nil
			colIdx = nil
			continue
		}

		// Remove leading/trailing | and split
		inner := strings.Trim(trimmed, "|")
		cols := splitColumns(inner)

		// Detect separator row (e.g. |---|---|)
		if isSeparator(cols) {
			if header != nil {
				inTable = true
				colIdx = buildColIndex(header)
			}
			continue
		}

		// If we don't have a header yet, this might be the header row
		if header == nil {
			header = cols
			continue
		}

		if !inTable || colIdx == nil {
			continue
		}

		// Parse data row
		count := 0
		if ci, ok := colIdx["count"]; ok && ci < len(cols) {
			count, _ = strconv.Atoi(strings.TrimSpace(cols[ci]))
		}
		if ci, ok := colIdx["frequency"]; ok && ci < len(cols) {
			count, _ = strconv.Atoi(strings.TrimSpace(cols[ci]))
		}

		// Skip rows with count 0 (or non-numeric)
		if count == 0 {
			continue
		}

		skill := ""
		if ci, ok := colIdx["skill"]; ok && ci < len(cols) {
			skill = strings.TrimSpace(cols[ci])
		}
		if ci, ok := colIdx["source skill"]; ok && ci < len(cols) {
			skill = strings.TrimSpace(cols[ci])
		}
		if ci, ok := colIdx["scenario"]; ok && ci < len(cols) {
			skill = strings.TrimSpace(cols[ci])
		}
		if ci, ok := colIdx["issue type"]; ok && ci < len(cols) {
			skill = strings.TrimSpace(cols[ci])
		}
		if ci, ok := colIdx["te rule"]; ok && ci < len(cols) {
			skill = strings.TrimSpace(cols[ci])
		}

		pattern := ""
		if ci, ok := colIdx["pattern"]; ok && ci < len(cols) {
			pattern = strings.TrimSpace(cols[ci])
		}
		if ci, ok := colIdx["failure pattern"]; ok && ci < len(cols) {
			pattern = strings.TrimSpace(cols[ci])
		}
		if ci, ok := colIdx["error pattern"]; ok && ci < len(cols) {
			pattern = strings.TrimSpace(cols[ci])
		}
		if ci, ok := colIdx["common violation"]; ok && ci < len(cols) {
			pattern = strings.TrimSpace(cols[ci])
		}

		fix := ""
		if ci, ok := colIdx["fix"]; ok && ci < len(cols) {
			fix = strings.TrimSpace(cols[ci])
		}
		if ci, ok := colIdx["fix pattern"]; ok && ci < len(cols) {
			fix = strings.TrimSpace(cols[ci])
		}
		if ci, ok := colIdx["resolution"]; ok && ci < len(cols) {
			fix = strings.TrimSpace(cols[ci])
		}
		if ci, ok := colIdx["prevention"]; ok && ci < len(cols) {
			fix = strings.TrimSpace(cols[ci])
		}

		if skill == "" || pattern == "" {
			continue
		}

		patterns = append(patterns, FailurePattern{
			Skill:   skill,
			Pattern: pattern,
			Fix:     fix,
			Count:   count,
		})
	}

	return patterns
}

// splitColumns splits a pipe-delimited inner string into trimmed columns.
func splitColumns(inner string) []string {
	parts := strings.Split(inner, "|")
	cols := make([]string, len(parts))
	for i, p := range parts {
		cols[i] = strings.TrimSpace(p)
	}
	return cols
}

// isSeparator checks if a row is a markdown table separator (e.g. --- | ---).
func isSeparator(cols []string) bool {
	for _, c := range cols {
		t := strings.TrimSpace(c)
		t = strings.Trim(t, "-")
		t = strings.Trim(t, ":")
		if t != "" {
			return false
		}
	}
	return len(cols) > 0
}

// buildColIndex maps lowercased header names to column indices.
func buildColIndex(header []string) map[string]int {
	m := make(map[string]int, len(header))
	for i, h := range header {
		m[strings.ToLower(strings.TrimSpace(h))] = i
	}
	return m
}

// Package distillation validates Knowledge Distillation Principle compliance
// in knowledge documents (ADRs, specs, plans).
//
// It enforces rules R1-R7 from AGENTS.md §Knowledge Distillation Principle:
//   - R1: spec/plan files must have "Current State" / "当前状态" section (ADVISORY)
//   - R2: ADR files must contain "Decision", "Rationale", "Consequence" (ERROR)
//   - R6: No "我们"/"他们" in knowledge docs (ADVISORY)
//   - R7: No "可能"/"也许"/"大概" in knowledge docs (ADVISORY)
//
// Scope: only actual knowledge artifacts (docs/adr/, docs/superpowers/specs/,
// docs/superpowers/plans/). AGENTS.md is the rule book and contains
// counter-examples intentionally — it is not scanned for R6/R7.
package distillation

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// CheckResult holds validation findings for a single file.
type CheckResult struct {
	File     string
	Errors   []string // R2 violations (ADR completeness)
	Advisory []string // R1, R6, R7 violations
}

// CheckDir validates all knowledge documents under root and returns
// a list of CheckResult (only files with violations are present).
func CheckDir(root string) ([]CheckResult, error) {
	var results []CheckResult

	// Check ADR files (R2 mandatory, R6/R7 advisory)
	adrDir := filepath.Join(root, "docs", "adr")
	if info, err := os.Stat(adrDir); err == nil && info.IsDir() {
		entries, err := os.ReadDir(adrDir)
		if err != nil {
			return nil, fmt.Errorf("readdir %s: %w", adrDir, err)
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || e.Name() == "README.md" {
				continue
			}
			// Only check actual ADR files (pattern: NNNN-*.md like 0001-wave-a-first.md)
			// Skip glossary, README, and other non-ADR files
			if !isADRFile(e.Name()) {
				continue
			}
			path := filepath.Join(adrDir, e.Name())
			result := checkADR(path)
			vague := checkVagueLanguage(path)
			result.Advisory = append(result.Advisory, vague.Advisory...)
			if len(result.Errors) > 0 || len(result.Advisory) > 0 {
				results = append(results, result)
			}
		}
	}

	// Check spec/plan files (R1 advisory, R6/R7 advisory)
	for _, sub := range []string{"specs", "plans"} {
		dir := filepath.Join(root, "docs", "superpowers", sub)
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || !strings.HasSuffix(path, ".md") {
				return nil
			}
			result := checkSpecPlan(path)
			vague := checkVagueLanguage(path)
			result.Advisory = append(result.Advisory, vague.Advisory...)
			if len(result.Errors) > 0 || len(result.Advisory) > 0 {
				results = append(results, result)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walk %s: %w", dir, err)
		}
	}

	return results, nil
}

// isADRFile checks if a filename matches the ADR pattern (NNNN-*.md).
// Returns false for glossary, README, and other non-ADR files.
func isADRFile(name string) bool {
	// Pattern: starts with 4 digits followed by hyphen
	if len(name) < 5 {
		return false
	}
	for i := range 4 {
		if name[i] < '0' || name[i] > '9' {
			return false
		}
	}
	return name[4] == '-'
}

// adrRequiredSections are the three mandatory ADR elements (R2).
// Accepts standard ADR variants: "Context" as alias for "Rationale",
// and "Consequences" (plural) as alias for "Consequence".
var adrRequiredSections = []string{"Decision", "Rationale", "Consequence"}

// adrSectionAliases maps required section names to their accepted alternatives.
var adrSectionAliases = map[string][]string{
	"Rationale":   {"Context"},
	"Consequence": {"Consequences"},
}

// checkADR validates R2: ADR files must contain Decision, Rationale, Consequence.
// Accepts aliases: "Context" for "Rationale", "Consequences" for "Consequence".
func checkADR(path string) CheckResult {
	result := CheckResult{File: path}

	content, err := os.ReadFile(path)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("read error: %v", err))
		return result
	}

	text := string(content)
	rel := filepath.ToSlash(path)

	for _, section := range adrRequiredSections {
		// Check main section name
		pat := regexp.MustCompile(`(?i)\b` + section + `\b`)
		found := pat.MatchString(text)

		// Check aliases if main section not found
		if !found {
			if aliases, ok := adrSectionAliases[section]; ok {
				for _, alias := range aliases {
					aliasPat := regexp.MustCompile(`(?i)\b` + alias + `\b`)
					if aliasPat.MatchString(text) {
						found = true
						break
					}
				}
			}
		}

		if !found {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: missing required section '%s'", rel, section))
		}
	}

	return result
}

// currentStatePatterns match "Current State" or "当前状态" headings (R1).
var (
	currentStateEN = regexp.MustCompile(`(?i)\bCurrent\s+State\b`)
	currentStateCN = regexp.MustCompile(`当前状态`)
)

// checkSpecPlan validates R1: spec/plan files should have a current state section.
func checkSpecPlan(path string) CheckResult {
	result := CheckResult{File: path}

	content, err := os.ReadFile(path)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("read error: %v", err))
		return result
	}

	text := string(content)
	rel := filepath.ToSlash(path)

	if !currentStateEN.MatchString(text) && !currentStateCN.MatchString(text) {
		result.Advisory = append(result.Advisory, fmt.Sprintf("%s: missing 'Current State' / '当前状态' section (R1)", rel))
	}

	return result
}

// r6Pattern matches vague subjects (R6).
var r6Pattern = regexp.MustCompile(`我们|他们`)

// r7Pattern matches vague qualifiers (R7).
var r7Pattern = regexp.MustCompile(`可能|也许|大概`)

// checkVagueLanguage validates R6/R7: no vague Chinese subjects/qualifiers.
//
// Skips:
//   - Code fences (``` lines)
//   - Markdown headings (# lines)
//   - Blockquotes (> lines) — often contain counter-examples or quoted context
//   - Table rows containing rule definitions (lines with | ❌ or | ✅)
func checkVagueLanguage(path string) CheckResult {
	result := CheckResult{File: path}

	file, err := os.Open(path)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("open error: %v", err))
		return result
	}
	defer file.Close()

	rel := filepath.ToSlash(path)
	scanner := bufio.NewScanner(file)
	lineNum := 0
	inCodeFence := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Track code fence state
		if strings.HasPrefix(trimmed, "```") {
			inCodeFence = !inCodeFence
			continue
		}
		if inCodeFence {
			continue
		}

		// Skip headings, blockquotes, and rule-example table rows
		if strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, ">") ||
			strings.Contains(trimmed, "| ❌") ||
			strings.Contains(trimmed, "| ✅") {
			continue
		}

		if loc := r6Pattern.FindStringIndex(line); loc != nil {
			result.Advisory = append(result.Advisory,
				fmt.Sprintf("%s:%d: vague subject '%s' (R6)", rel, lineNum, line[loc[0]:loc[1]]))
		}

		if loc := r7Pattern.FindStringIndex(line); loc != nil {
			result.Advisory = append(result.Advisory,
				fmt.Sprintf("%s:%d: vague qualifier '%s' (R7)", rel, lineNum, line[loc[0]:loc[1]]))
		}
	}

	if err := scanner.Err(); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("scan error: %v", err))
	}

	return result
}

package distillation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckADR(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantErrors  int
	}{
		{
			name: "complete ADR",
			content: `# ADR-0001: Wave A First

## Decision
Implement heal probe before any AUTO execution.

## Rationale
No real probe → no safe AUTO.

## Consequence
Wave A blocked until heal probe complete.
`,
			wantErrors: 0,
		},
		{
			name: "missing Decision",
			content: `# ADR-0002: Incomplete ADR

## Rationale
Some rationale here.

## Consequence
Some consequence here.
`,
			wantErrors: 1,
		},
		{
			name: "missing Rationale and Consequence",
			content: `# ADR-0003: Incomplete

## Decision
Some decision here.
`,
			wantErrors: 2,
		},
		{
			name: "case insensitive",
			content: `# ADR-0004: Case Test

## decision
Lower case decision.

## RATIONALE
Upper case rationale.

## consequence
Lower case consequence.
`,
			wantErrors: 0,
		},
		{
			name: "alias Context for Rationale",
			content: `# ADR-0005: Alias Test

## Decision
Some decision here.

## Context
Context serves as rationale alias.

## Consequence
Some consequence here.
`,
			wantErrors: 0,
		},
		{
			name: "alias Consequences plural",
			content: `# ADR-0006: Plural Alias Test

## Decision
Some decision here.

## Rationale
Some rationale here.

## Consequences
Plural consequences as alias.
`,
			wantErrors: 0,
		},
		{
			name: "both aliases together",
			content: `# ADR-0007: Both Aliases

## Decision
Some decision here.

## Context
Context as rationale.

## Consequences
Plural consequences.
`,
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			adrFile := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(adrFile, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			result := checkADR(adrFile)
			if len(result.Errors) != tt.wantErrors {
				t.Errorf("checkADR() errors = %d, want %d\nErrors: %v", len(result.Errors), tt.wantErrors, result.Errors)
			}
		})
	}
}

func TestCheckSpecPlan(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		wantAdvisory int
	}{
		{
			name: "has Current State",
			content: `# Spec: Feature X

## Current State
Feature X is not implemented.

## Target State
Feature X will be implemented.
`,
			wantAdvisory: 0,
		},
		{
			name: "has 当前状态",
			content: `# 规格：功能 X

## 当前状态
功能 X 未实现。

## 目标状态
功能 X 将被实现。
`,
			wantAdvisory: 0,
		},
		{
			name: "missing current state",
			content: `# Spec: Feature Y

## Background
Some background.

## Goals
Some goals.
`,
			wantAdvisory: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			specFile := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(specFile, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			result := checkSpecPlan(specFile)
			if len(result.Advisory) != tt.wantAdvisory {
				t.Errorf("checkSpecPlan() advisory = %d, want %d\nAdvisory: %v", len(result.Advisory), tt.wantAdvisory, result.Advisory)
			}
		})
	}
}

func TestCheckVagueLanguage(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		wantAdvisory int
	}{
		{
			name: "no vague language",
			content: `# Clear Document

The system implements feature X.
Users can access the API.
`,
			wantAdvisory: 0,
		},
		{
			name: "R6 vague subject 我们",
			content: `# Document

我们认为应该实现这个功能。
`,
			wantAdvisory: 1,
		},
		{
			name: "R6 vague subject 他们",
			content: `# Document

他们决定使用这个方案。
`,
			wantAdvisory: 1,
		},
		{
			name: "R7 vague qualifier 可能",
			content: `# Document

这可能会导致性能问题。
`,
			wantAdvisory: 1,
		},
		{
			name: "R7 vague qualifier 也许",
			content: `# Document

也许这个方案更好。
`,
			wantAdvisory: 1,
		},
		{
			name: "R7 vague qualifier 大概",
			content: `# Document

大概需要三天时间。
`,
			wantAdvisory: 1,
		},
		{
			name: "skip code fence",
			content: "# Document\n\n```\n我们可能也许大概\n```\n\nClear text here.\n",
			wantAdvisory: 0,
		},
		{
			name: "skip heading",
			content: `# 我们认为

Some content.
`,
			wantAdvisory: 0,
		},
		{
			name: "skip blockquote",
			content: `# Document

> 我们认为这是正确的。

Clear text.
`,
			wantAdvisory: 0,
		},
		{
			name: "skip rule table",
			content: `# Document

| ❌ | 我们认为应该... |
| ✅ | The system implements... |

Clear text.
`,
			wantAdvisory: 0,
		},
		{
			name: "multiple violations",
			content: `# Document

我们认为可能需要三天。
他们也许会延迟。
`,
			wantAdvisory: 4, // 我们, 可能, 他们, 也许
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			result := checkVagueLanguage(testFile)
			if len(result.Advisory) != tt.wantAdvisory {
				t.Errorf("checkVagueLanguage() advisory = %d, want %d\nAdvisory: %v", len(result.Advisory), tt.wantAdvisory, result.Advisory)
			}
		})
	}
}

func TestCheckDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create docs/adr directory
	adrDir := filepath.Join(tmpDir, "docs", "adr")
	if err := os.MkdirAll(adrDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Valid ADR
	validADR := `# ADR-0001: Valid

## Decision
Valid decision.

## Rationale
Valid rationale.

## Consequence
Valid consequence.
`
	if err := os.WriteFile(filepath.Join(adrDir, "0001-valid.md"), []byte(validADR), 0644); err != nil {
		t.Fatal(err)
	}

	// Invalid ADR (missing Decision)
	invalidADR := `# ADR-0002: Invalid

## Rationale
Some rationale.

## Consequence
Some consequence.
`
	if err := os.WriteFile(filepath.Join(adrDir, "0002-invalid.md"), []byte(invalidADR), 0644); err != nil {
		t.Fatal(err)
	}

	// Create docs/superpowers/specs directory
	specsDir := filepath.Join(tmpDir, "docs", "superpowers", "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Valid spec
	validSpec := `# Spec: Feature X

## Current State
Not implemented.

## Target State
Implemented.
`
	if err := os.WriteFile(filepath.Join(specsDir, "feature-x.md"), []byte(validSpec), 0644); err != nil {
		t.Fatal(err)
	}

	// Invalid spec (missing Current State)
	invalidSpec := `# Spec: Feature Y

## Background
Some background.
`
	if err := os.WriteFile(filepath.Join(specsDir, "feature-y.md"), []byte(invalidSpec), 0644); err != nil {
		t.Fatal(err)
	}

	results, err := CheckDir(tmpDir)
	if err != nil {
		t.Fatalf("CheckDir() error = %v", err)
	}

	// Should have 2 results: invalid ADR and invalid spec
	if len(results) != 2 {
		t.Errorf("CheckDir() returned %d results, want 2", len(results))
		for _, r := range results {
			t.Logf("  %s: errors=%d, advisory=%d", r.File, len(r.Errors), len(r.Advisory))
		}
	}

	// Check that invalid ADR has errors
	foundInvalidADR := false
	foundInvalidSpec := false
	for _, r := range results {
		if filepath.Base(r.File) == "0002-invalid.md" {
			foundInvalidADR = true
			if len(r.Errors) != 1 {
				t.Errorf("invalid ADR has %d errors, want 1", len(r.Errors))
			}
		}
		if filepath.Base(r.File) == "feature-y.md" {
			foundInvalidSpec = true
			if len(r.Advisory) != 1 {
				t.Errorf("invalid spec has %d advisory, want 1", len(r.Advisory))
			}
		}
	}

	if !foundInvalidADR {
		t.Error("did not find invalid ADR in results")
	}
	if !foundInvalidSpec {
		t.Error("did not find invalid spec in results")
	}
}

package policy

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

func TestLoad_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	policyDir := filepath.Join(dir, "incident-loop-agent", "references", "policies")
	if err := os.MkdirAll(policyDir, 0755); err != nil {
		t.Fatalf("failed to create policy dir: %v", err)
	}

	_, err := Load(dir)
	if err == nil {
		t.Error("expected error for empty policy dir, got nil")
	}
}

func TestLoad_ExecutionRisk(t *testing.T) {
	dir := t.TempDir()
	policyDir := filepath.Join(dir, "incident-loop-agent", "references", "policies")
	if err := os.MkdirAll(policyDir, 0755); err != nil {
		t.Fatalf("failed to create policy dir: %v", err)
	}

	writeFile(t, policyDir, "execution-risk.md", `# Execution Risk Policy

## 2. Decision matrix

| risk | blast_radius | decision |
|------|--------------|----------|
| read-only | single | AUTO |
| read-only | multi | AUTO |
| read-only | account-or-region | AUTO |
| state-changing | single + high conf | AUTO |
| state-changing | single + medium/low conf | ASK |
| state-changing | multi | ASK |
| state-changing | account-or-region | ASK |
| destructive | single | ASK |
| destructive | multi | ASK |
| destructive | account-or-region | ASK |
`)

	writeFile(t, policyDir, "domain-allowlist.md", "# Domain Allow-list\n\n## 1. Eligible skills\n\n"+
		"`ve-ecs-ops`\n"+
		"`ve-redis-ops`\n")

	ps, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(ps.ExecutionRisk.AutoConditions) == 0 {
		t.Error("expected AUTO conditions, got none")
	}
	if len(ps.ExecutionRisk.AskConditions) == 0 {
		t.Error("expected ASK conditions, got none")
	}
	// Should have at least 1 REFUSE condition (hard safety floor)
	if len(ps.ExecutionRisk.RefuseConditions) == 0 {
		t.Error("expected REFUSE conditions (hard safety floor), got none")
	}

	// Verify specific condition
	found := false
	for _, c := range ps.ExecutionRisk.RefuseConditions {
		if c == "Safety=0 + any → REFUSE" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected hard safety floor condition 'Safety=0 + any → REFUSE', not found")
	}
}

func TestLoad_DomainAllowlist(t *testing.T) {
	dir := t.TempDir()
	policyDir := filepath.Join(dir, "incident-loop-agent", "references", "policies")
	if err := os.MkdirAll(policyDir, 0755); err != nil {
		t.Fatalf("failed to create policy dir: %v", err)
	}

	writeFile(t, policyDir, "execution-risk.md", `# Execution Risk Policy

## 2. Decision matrix

| risk | blast_radius | decision |
|------|--------------|----------|
| read-only | any | AUTO |
`)

	writeFile(t, policyDir, "domain-allowlist.md", "# Domain Allow-list\n\n## 1. Eligible skills\n\n"+
		"`ve-ecs-ops`\n"+
		"`ve-redis-ops`\n"+
		"`ve-vpc-ops`\n"+
		"`ve-iam-ops`\n"+
		"`ve-kms-ops`\n"+
		"`ve-billing-ops`\n"+
		"`ve-rds-mysql-ops`\n"+
		"`ve-cms-ops`\n")

	ps, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(ps.DomainAllowlist) < 8 {
		t.Errorf("expected at least 8 skills in DomainAllowlist, got %d: %v", len(ps.DomainAllowlist), ps.DomainAllowlist)
	}

	// Verify specific skills are present
	expected := []string{"ve-ecs-ops", "ve-redis-ops", "ve-vpc-ops", "ve-iam-ops"}
	skillSet := map[string]bool{}
	for _, s := range ps.DomainAllowlist {
		skillSet[s] = true
	}
	for _, s := range expected {
		if !skillSet[s] {
			t.Errorf("expected skill %q in DomainAllowlist, not found", s)
		}
	}
}

func TestLoad_FullPolicies(t *testing.T) {
	dir := t.TempDir()
	policyDir := filepath.Join(dir, "incident-loop-agent", "references", "policies")
	if err := os.MkdirAll(policyDir, 0755); err != nil {
		t.Fatalf("failed to create policy dir: %v", err)
	}

	writeFile(t, policyDir, "execution-risk.md", `# Execution Risk Policy

## 2. Decision matrix

| risk | blast_radius | decision |
|------|--------------|----------|
| read-only | any | AUTO |
| destructive | any | ASK |
`)

	writeFile(t, policyDir, "domain-allowlist.md", "# Domain Allow-list\n\n## 1. Eligible skills\n\n"+
		"`ve-ecs-ops`\n"+
		"`ve-redis-ops`\n")

	writeFile(t, policyDir, "guardrails.yaml", `guardrails:
  - id: gr-abc123
    skill: ve-ecs-ops
    trigger: evidence_overfetch
    action: auto-ASK
    severity: medium
    source_count: 15
  - id: gr-def456
    skill: ve-redis-ops
    trigger: safety_score_low
    action: auto-REFUSE
    severity: high
    source_count: 3
`)

	ps, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(ps.ExecutionRisk.AutoConditions) == 0 {
		t.Error("expected AUTO conditions")
	}
	if len(ps.ExecutionRisk.AskConditions) == 0 {
		t.Error("expected ASK conditions")
	}
	if len(ps.DomainAllowlist) != 2 {
		t.Errorf("expected 2 skills, got %d", len(ps.DomainAllowlist))
	}
	if len(ps.Guardrails) != 2 {
		t.Errorf("expected 2 guardrails, got %d", len(ps.Guardrails))
	}

	// Verify guardrail details
	gr0 := ps.Guardrails[0]
	if gr0.ID != "gr-abc123" {
		t.Errorf("expected first guardrail ID gr-abc123, got %s", gr0.ID)
	}
	if gr0.SourceCount != 15 {
		t.Errorf("expected source_count 15, got %d", gr0.SourceCount)
	}
}

func TestLoad_MissingGuardrails(t *testing.T) {
	dir := t.TempDir()
	policyDir := filepath.Join(dir, "incident-loop-agent", "references", "policies")
	if err := os.MkdirAll(policyDir, 0755); err != nil {
		t.Fatalf("failed to create policy dir: %v", err)
	}

	writeFile(t, policyDir, "execution-risk.md", `# Execution Risk Policy

## 2. Decision matrix

| risk | blast_radius | decision |
|------|--------------|----------|
| read-only | any | AUTO |
`)

	writeFile(t, policyDir, "domain-allowlist.md", "# Domain Allow-list\n\n## 1. Eligible skills\n\n`ve-ecs-ops`\n")

	ps, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(ps.Guardrails) != 0 {
		t.Errorf("expected empty Guardrails, got %d entries", len(ps.Guardrails))
	}
}

func TestDiff_Identical(t *testing.T) {
	a := &PolicySet{
		ExecutionRisk: ExecutionRiskPolicy{
			AutoConditions:   []string{"read-only + any → AUTO"},
			AskConditions:    []string{"destructive + any → ASK"},
			RefuseConditions: []string{"Safety=0 + any → REFUSE"},
		},
		DomainAllowlist: []string{"ve-ecs-ops", "ve-redis-ops"},
		Guardrails: []GuardrailEntry{
			{ID: "gr-abc123", Skill: "ve-ecs-ops", Trigger: "evidence_overfetch", Action: "auto-ASK", Severity: "medium", SourceCount: 15},
		},
	}
	b := &PolicySet{
		ExecutionRisk: ExecutionRiskPolicy{
			AutoConditions:   []string{"read-only + any → AUTO"},
			AskConditions:    []string{"destructive + any → ASK"},
			RefuseConditions: []string{"Safety=0 + any → REFUSE"},
		},
		DomainAllowlist: []string{"ve-ecs-ops", "ve-redis-ops"},
		Guardrails: []GuardrailEntry{
			{ID: "gr-abc123", Skill: "ve-ecs-ops", Trigger: "evidence_overfetch", Action: "auto-ASK", Severity: "medium", SourceCount: 15},
		},
	}

	changes := DiffPolicySets(a, b)
	if len(changes) != 0 {
		t.Errorf("expected 0 changes for identical PolicySets, got %d: %v", len(changes), changes)
	}
}

func TestDiff_DomainChanged(t *testing.T) {
	old := &PolicySet{
		ExecutionRisk: ExecutionRiskPolicy{
			AutoConditions:   []string{"read-only + any → AUTO"},
			AskConditions:    []string{"destructive + any → ASK"},
			RefuseConditions: []string{"Safety=0 + any → REFUSE"},
		},
		DomainAllowlist: []string{"ve-ecs-ops"},
	}
	new := &PolicySet{
		ExecutionRisk: ExecutionRiskPolicy{
			AutoConditions:   []string{"read-only + any → AUTO"},
			AskConditions:    []string{"destructive + any → ASK"},
			RefuseConditions: []string{"Safety=0 + any → REFUSE"},
		},
		DomainAllowlist: []string{"ve-ecs-ops", "ve-redis-ops"},
	}

	changes := DiffPolicySets(old, new)
	if len(changes) != 1 {
		t.Errorf("expected 1 change, got %d: %v", len(changes), changes)
	}

	found := false
	for _, c := range changes {
		if c.File == "domain-allowlist" && c.Type == "added" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected domain-allowlist added change, not found")
	}
}

func TestDiff_GuardrailChanged(t *testing.T) {
	old := &PolicySet{
		ExecutionRisk: ExecutionRiskPolicy{
			AutoConditions:   []string{"read-only + any → AUTO"},
			AskConditions:    []string{"destructive + any → ASK"},
			RefuseConditions: []string{"Safety=0 + any → REFUSE"},
		},
		DomainAllowlist: []string{"ve-ecs-ops"},
		Guardrails: []GuardrailEntry{
			{ID: "gr-abc123", Skill: "ve-ecs-ops", Trigger: "evidence_overfetch", Action: "auto-ASK", Severity: "medium", SourceCount: 15},
		},
	}
	new := &PolicySet{
		ExecutionRisk: ExecutionRiskPolicy{
			AutoConditions:   []string{"read-only + any → AUTO"},
			AskConditions:    []string{"destructive + any → ASK"},
			RefuseConditions: []string{"Safety=0 + any → REFUSE"},
		},
		DomainAllowlist: []string{"ve-ecs-ops"},
		Guardrails: []GuardrailEntry{
			{ID: "gr-abc123", Skill: "ve-ecs-ops", Trigger: "evidence_overfetch", Action: "auto-ASK", Severity: "medium", SourceCount: 42},
		},
	}

	changes := DiffPolicySets(old, new)
	if len(changes) != 1 {
		t.Errorf("expected 1 change, got %d: %v", len(changes), changes)
	}

	found := false
	for _, c := range changes {
		if c.File == "guardrails" && c.Type == "changed" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected guardrails changed change, not found")
	}
}

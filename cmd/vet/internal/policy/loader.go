package policy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// PolicySet is the unified view of all policy files.
type PolicySet struct {
	ExecutionRisk   ExecutionRiskPolicy `json:"execution_risk"`
	DomainAllowlist []string            `json:"domain_allowlist"`
	Guardrails      []GuardrailEntry    `json:"guardrails"`
}

type ExecutionRiskPolicy struct {
	AutoConditions   []string `json:"auto_conditions"`
	AskConditions    []string `json:"ask_conditions"`
	RefuseConditions []string `json:"refuse_conditions"`
}

type GuardrailEntry struct {
	ID          string `json:"id" yaml:"id"`
	Skill       string `json:"skill" yaml:"skill"`
	Trigger     string `json:"trigger" yaml:"trigger"`
	Action      string `json:"action" yaml:"action"`
	Severity    string `json:"severity" yaml:"severity"`
	SourceCount int    `json:"source_count" yaml:"source_count"`
}

type PolicyChange struct {
	File   string `json:"file"`
	Type   string `json:"type"`
	Detail string `json:"detail"`
}

// Load reads all policy files from rootPath/incident-loop-agent/references/policies/
// and merges them into a PolicySet.
func Load(rootPath string) (*PolicySet, error) {
	policyDir := filepath.Join(rootPath, "incident-loop-agent", "references", "policies")

	ps := &PolicySet{}

	execRisk, err := loadExecutionRisk(filepath.Join(policyDir, "execution-risk.md"))
	if err != nil {
		return nil, fmt.Errorf("loading execution-risk: %w", err)
	}
	ps.ExecutionRisk = *execRisk

	allowlist, err := loadDomainAllowlist(filepath.Join(policyDir, "domain-allowlist.md"))
	if err != nil {
		return nil, fmt.Errorf("loading domain-allowlist: %w", err)
	}
	ps.DomainAllowlist = allowlist

	guardrails, err := loadGuardrails(filepath.Join(policyDir, "guardrails.yaml"))
	if err != nil {
		return nil, fmt.Errorf("loading guardrails: %w", err)
	}
	ps.Guardrails = guardrails

	return ps, nil
}

// DiffPolicySets compares two PolicySets and returns the list of changes.
func DiffPolicySets(old, new *PolicySet) []PolicyChange {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		return []PolicyChange{{File: "policy", Type: "added", Detail: "entire PolicySet added"}}
	}
	if new == nil {
		return []PolicyChange{{File: "policy", Type: "removed", Detail: "entire PolicySet removed"}}
	}

	var changes []PolicyChange

	// Compare ExecutionRisk AutoConditions
	autoAdded, autoRemoved := diffStringSlices(old.ExecutionRisk.AutoConditions, new.ExecutionRisk.AutoConditions)
	for _, s := range autoAdded {
		changes = append(changes, PolicyChange{File: "execution-risk", Type: "added", Detail: "AUTO condition: " + s})
	}
	for _, s := range autoRemoved {
		changes = append(changes, PolicyChange{File: "execution-risk", Type: "removed", Detail: "AUTO condition: " + s})
	}

	// Compare ExecutionRisk AskConditions
	askAdded, askRemoved := diffStringSlices(old.ExecutionRisk.AskConditions, new.ExecutionRisk.AskConditions)
	for _, s := range askAdded {
		changes = append(changes, PolicyChange{File: "execution-risk", Type: "added", Detail: "ASK condition: " + s})
	}
	for _, s := range askRemoved {
		changes = append(changes, PolicyChange{File: "execution-risk", Type: "removed", Detail: "ASK condition: " + s})
	}

	// Compare ExecutionRisk RefuseConditions
	refuseAdded, refuseRemoved := diffStringSlices(old.ExecutionRisk.RefuseConditions, new.ExecutionRisk.RefuseConditions)
	for _, s := range refuseAdded {
		changes = append(changes, PolicyChange{File: "execution-risk", Type: "added", Detail: "REFUSE condition: " + s})
	}
	for _, s := range refuseRemoved {
		changes = append(changes, PolicyChange{File: "execution-risk", Type: "removed", Detail: "REFUSE condition: " + s})
	}

	// Compare DomainAllowlist
	domainAdded, domainRemoved := diffStringSlices(old.DomainAllowlist, new.DomainAllowlist)
	for _, s := range domainAdded {
		changes = append(changes, PolicyChange{File: "domain-allowlist", Type: "added", Detail: "skill: " + s})
	}
	for _, s := range domainRemoved {
		changes = append(changes, PolicyChange{File: "domain-allowlist", Type: "removed", Detail: "skill: " + s})
	}

	// Compare Guardrails by ID
	oldGR := buildGuardrailMap(old.Guardrails)
	newGR := buildGuardrailMap(new.Guardrails)

	for id := range newGR {
		if _, ok := oldGR[id]; !ok {
			changes = append(changes, PolicyChange{File: "guardrails", Type: "added", Detail: "guardrail: " + id})
		}
	}
	for id := range oldGR {
		if _, ok := newGR[id]; !ok {
			changes = append(changes, PolicyChange{File: "guardrails", Type: "removed", Detail: "guardrail: " + id})
		}
	}
	// Check for changes in existing guardrails
	for id := range oldGR {
		oldEntry := oldGR[id]
		newEntry, exists := newGR[id]
		if !exists {
			continue
		}
		oldJSON, _ := json.Marshal(oldEntry)
		newJSON, _ := json.Marshal(newEntry)
		if string(oldJSON) != string(newJSON) {
			changes = append(changes, PolicyChange{File: "guardrails", Type: "changed", Detail: "guardrail: " + id})
		}
	}

	return changes
}

func loadExecutionRisk(path string) (*ExecutionRiskPolicy, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	policy := &ExecutionRiskPolicy{}

	var inDecisionMatrix bool
	var headerSkipped bool
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		// Detect decision matrix section header
		if strings.Contains(line, "## 2. Decision matrix") {
			inDecisionMatrix = true
			continue
		}
		if !inDecisionMatrix {
			continue
		}
		if !headerSkipped {
			// Skip blank lines and prose between section header and table
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || !strings.HasPrefix(trimmed, "|") {
				continue
			}
			// Skip the header separator line and column headers
			if strings.Contains(line, "---") {
				continue
			}
			if strings.Contains(line, "risk") && strings.Contains(line, "blast_radius") {
				continue
			}
			headerSkipped = true
		}

		// End of table — stop parsing
		if !strings.HasPrefix(strings.TrimSpace(line), "|") {
			break
		}

		cols := parseTableRow(line)
		if len(cols) < 3 {
			continue
		}
		risk := strings.TrimSpace(cols[0])
		blastRadius := strings.TrimSpace(cols[1])
		decision := strings.TrimSpace(cols[2])

		condition := fmt.Sprintf("%s + %s → %s", risk, blastRadius, decision)

		if strings.Contains(decision, "AUTO") {
			policy.AutoConditions = append(policy.AutoConditions, condition)
		} else if strings.Contains(decision, "ASK") {
			policy.AskConditions = append(policy.AskConditions, condition)
		} else if strings.Contains(decision, "REFUSE") {
			policy.RefuseConditions = append(policy.RefuseConditions, condition)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Add hard safety floor condition from Section 3
	policy.RefuseConditions = append(policy.RefuseConditions, "Safety=0 + any → REFUSE")

	return policy, nil
}

var skillNameRe = regexp.MustCompile("`(ve-[a-z][a-z-]*ops)`")

func loadDomainAllowlist(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var skills []string
	seen := map[string]bool{}

	var inSkillsSection bool
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		// Detect Section 1: Eligible skills
		if strings.Contains(line, "Eligible skills") && strings.HasPrefix(line, "##") {
			inSkillsSection = true
			continue
		}
		if !inSkillsSection {
			continue
		}
		// Stop at next section header
		if strings.HasPrefix(line, "## ") {
			break
		}

		// Try backtick format: `ve-ecs-ops`
		matches := skillNameRe.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			name := m[1]
			if !seen[name] {
				seen[name] = true
				skills = append(skills, name)
			}
		}

		// Try table format: | ve-ecs-ops | ✅ |
		if strings.HasPrefix(line, "|") {
			cols := parseTableRow(line)
			for _, col := range cols {
				col = strings.TrimSpace(col)
				if strings.HasPrefix(col, "ve-") && strings.HasSuffix(col, "-ops") {
					if !seen[col] {
						seen[col] = true
						skills = append(skills, col)
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return skills, nil
}

type guardrailsFile struct {
	Guardrails []GuardrailEntry `yaml:"guardrails"`
}

func loadGuardrails(path string) ([]GuardrailEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var gf guardrailsFile
	if err := yaml.Unmarshal(data, &gf); err != nil {
		return nil, fmt.Errorf("parsing guardrails YAML: %w", err)
	}

	return gf.Guardrails, nil
}

func parseTableRow(line string) []string {
	// Split by |, trim each cell, remove leading/trailing empty strings
	parts := strings.Split(line, "|")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func diffStringSlices(old, new []string) (added, removed []string) {
	oldSet := map[string]bool{}
	for _, s := range old {
		oldSet[s] = true
	}
	newSet := map[string]bool{}
	for _, s := range new {
		newSet[s] = true
	}
	for _, s := range new {
		if !oldSet[s] {
			added = append(added, s)
		}
	}
	for _, s := range old {
		if !newSet[s] {
			removed = append(removed, s)
		}
	}
	return added, removed
}

func buildGuardrailMap(entries []GuardrailEntry) map[string]GuardrailEntry {
	m := map[string]GuardrailEntry{}
	for _, e := range entries {
		m[e.ID] = e
	}
	return m
}

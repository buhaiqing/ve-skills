package diagnosis

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

func CheckDir(root string) (map[string][]string, []string) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, nil
	}

	var skills []string
	results := make(map[string][]string)

	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "ve-") || !strings.HasSuffix(entry.Name(), "-ops") {
			continue
		}
		skill := entry.Name()
		skills = append(skills, skill)

		path := filepath.Join(root, skill, "references", "advanced", "diagnosis-rules.yaml")
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				results[skill] = nil
				continue
			}
			results[skill] = []string{fmt.Sprintf("read error: %v", err)}
			continue
		}

		var doc struct {
			Rules []DiagnosisRule `yaml:"rules"`
		}
		if err := yaml.Unmarshal(data, &doc); err != nil {
			results[skill] = []string{fmt.Sprintf("YAML parse error: %v", err)}
			continue
		}

		var errs []string
		seenIDs := make(map[string]bool)
		for i, rule := range doc.Rules {
			if rule.ID == "" {
				errs = append(errs, fmt.Sprintf("rule[%d]: missing id", i))
			}
			if rule.Product == "" {
				errs = append(errs, fmt.Sprintf("rule[%d] (%s): missing product", i, rule.ID))
			}
			if rule.Trigger.Metric == "" {
				errs = append(errs, fmt.Sprintf("rule[%d] (%s): missing trigger.metric", i, rule.ID))
			}
			if rule.Trigger.Operator == "" {
				errs = append(errs, fmt.Sprintf("rule[%d] (%s): missing trigger.operator", i, rule.ID))
			}
			if rule.Severity == "" {
				errs = append(errs, fmt.Sprintf("rule[%d] (%s): missing severity", i, rule.ID))
			}
			if seenIDs[rule.ID] {
				errs = append(errs, fmt.Sprintf("duplicate rule id: %s", rule.ID))
			}
			seenIDs[rule.ID] = true
		}
		if len(errs) > 0 {
			results[skill] = errs
		} else {
			results[skill] = nil
		}
	}

	sort.Strings(skills)
	return results, skills
}

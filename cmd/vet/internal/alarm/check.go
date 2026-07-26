package alarm

import (
	"os"
	"path/filepath"
	"strings"
)

func CheckAlarmMerge(root string) (map[string][]string, []string) {
	results := make(map[string][]string)
	ilaDir := filepath.Join(root, "incident-loop-agent")
	ilaPath := filepath.Join(ilaDir, "SKILL.md")

	// PerSkillCheck uses filepath.Dir(sk) then filepath.Base(that) for display name.
	// For SKILL.md path "/root/incident-loop-agent/SKILL.md":
	//   Dir = "/root/incident-loop-agent", Base = "incident-loop-agent" ✓
	ilaSKILL := ilaPath // SKILL.md path → Dir → Base = "incident-loop-agent"

	data, err := os.ReadFile(ilaPath)
	if err != nil {
		results[ilaSKILL] = []string{"incident-loop-agent/SKILL.md not found"}
		return results, []string{ilaSKILL}
	}

	content := string(data)
	hasAlarmMerge := strings.Contains(strings.ToLower(content), "alarm merge") ||
		strings.Contains(content, "告警归并")

	// Check that alarm package compiles (the merge.go testable units exist)
	mergeGoPath := filepath.Join(root, "cmd", "vet", "internal", "alarm", "merge.go")
	if _, err := os.Stat(mergeGoPath); os.IsNotExist(err) {
		results[ilaSKILL] = []string{"cmd/vet/internal/alarm/merge.go not found"}
		return results, []string{ilaSKILL}
	}

	// Check rubric.md for Alarm Merge rubric dimension (non-fatal)
	rubricPath := filepath.Join(ilaDir, "references", "rubric.md")
	rubricData, err := os.ReadFile(rubricPath)
	hasRubric := err == nil && strings.Contains(strings.ToLower(string(rubricData)), "alarm merge")

	_ = hasRubric // rubric check is informational only
	if hasAlarmMerge {
		results[ilaSKILL] = nil // pass
	} else {
		results[ilaSKILL] = []string{"incident-loop-agent/SKILL.md missing alarm merge section"}
	}

	return results, []string{ilaSKILL}
}

// CheckDir is the entry-point called from vet check.
func CheckDir(root string) (map[string][]string, []string) {
	return CheckAlarmMerge(root)
}

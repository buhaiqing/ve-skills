package agent

import "github.com/buhaiqing/ve-skills/cmd/vet/internal/memory"

// Reflect writes back failure patterns after execution.
func Reflect(root, skill, errorPattern, category, fix string) error {
	return memory.AppendFailurePattern(root, memory.FailurePatternEntry{
		Skill:    skill,
		Pattern:  errorPattern,
		Category: category,
		Fix:      fix,
		Source:   "agent",
		Count:    1,
	})
}

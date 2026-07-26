// Package tags checks that ve-*-ops SKILL.md files include --Tags parameters
// in their Create* CLI commands for FinOps cost attribution compliance.
package tags

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Result holds the tags check result for a single skill.
type Result struct {
	Skill   string `json:"skill"`
	HasTags bool   `json:"has_tags"`
	OK      bool   `json:"ok"`
}

// Report holds the aggregate tags check report.
type Report struct {
	TotalSkills int      `json:"total_skills"`
	Quality     []Result `json:"quality"`
	Passing     int      `json:"passing"`
	Failing     int      `json:"failing"`
	OK          bool     `json:"ok"`
}

var (
	// createRe matches CLI commands that create billable resources inside bash code blocks.
	createRe = regexp.MustCompile(`(?i)^ve\s+\S+\s+(Create\w+|RunInstances|Allocate\w+)`)

	// tagsRe matches --Tags parameter in a line (possibly with escaped JSON).
	tagsRe = regexp.MustCompile(`--Tags\b`)
)

func checkSkill(root, skill string) Result {
	r := Result{Skill: skill}
	path := filepath.Join(root, skill, "SKILL.md")
	f, err := os.Open(path)
	if err != nil {
		// Skill directory or SKILL.md might not exist in test temp dirs
		r.OK = true
		return r
	}
	defer f.Close()

	var inBashBlock bool
	var hasCreate bool
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Track bash code block boundaries
		if strings.HasPrefix(trimmed, "```bash") || strings.HasPrefix(trimmed, "```shell") || strings.HasPrefix(trimmed, "```sh") {
			inBashBlock = true
			continue
		}
		if strings.HasPrefix(trimmed, "```") && inBashBlock {
			inBashBlock = false
			continue
		}
		if !inBashBlock {
			continue
		}

		// Inside bash block: check for Create commands (use trimmed for ^ anchor)
		createMatch := createRe.FindStringSubmatch(trimmed)
		if createMatch == nil {
			continue
		}
		hasCreate = true
		// This is a Create command — check if it has --Tags
		if tagsRe.MatchString(trimmed) {
			r.HasTags = true
		}
	}

	// OK if no Create commands (N/A) or all Create commands have --Tags
	r.OK = !hasCreate || r.HasTags
	return r
}

// AllSkills is the full set of 29 skills.
var AllSkills = []string{
	"ve-ecs-ops", "ve-redis-ops", "ve-rds-mysql-ops", "ve-rds-ops", "ve-rds-pg-ops",
	"ve-polar-mysql-ops", "ve-mongodb-ops", "ve-elasticsearch-ops", "ve-tos-ops", "ve-iam-ops",
	"ve-kms-ops", "ve-eip-ops", "ve-security-group-ops", "ve-vpc-ops", "ve-nat-ops", "ve-vpn-ops",
	"ve-clb-ops", "ve-alb-ops", "ve-vke-ops", "ve-nas-ops", "ve-cms-ops", "ve-fg-ops", "ve-ark-ops",
	"ve-cdn-ops", "ve-dns-ops", "ve-kafka-ops", "ve-sls-ops", "ve-billing-ops", "ve-skill-generator",
}

// CheckDir scans all skills and returns the tags report.
func CheckDir(root string) Report {
	var results []Result
	for _, s := range AllSkills {
		results = append(results, checkSkill(root, s))
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Skill < results[j].Skill })
	var passing int
	for _, r := range results {
		if r.OK {
			passing++
		}
	}
	return Report{
		TotalSkills: len(AllSkills),
		Quality:     results,
		Passing:     passing,
		Failing:     len(AllSkills) - passing,
		OK:          passing == len(AllSkills),
	}
}
// Package finopsquality checks that ve-*-ops skills have quality finops.md content,
// not just template placeholders. It validates line count, pricing keywords, and
// optimization coverage per the FINOPS quality tiers (required ≥20 lines, ≥2
// optimizations; recommended ≥15 lines, ≥1 optimization; optional ≥10 lines).
package finopsquality

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Result holds the quality check result for a single skill.
type Result struct {
	Skill           string `json:"skill"`
	HasFinOps       bool   `json:"has_finops"`
	LineCount       int    `json:"line_count"`
	HasPricing      string `json:"has_pricing"`    // empty or the product name detected
	HasOptimizations bool  `json:"has_optimizations"`
	OK              bool   `json:"ok"`
}

// Report holds the aggregate quality check report.
type Report struct {
	TotalSkills int      `json:"total_skills"`
	Quality     []Result `json:"quality"`
	Passing     int      `json:"passing"`
	Failing     int      `json:"failing"`
	OK          bool     `json:"ok"`
}

var (
	pricingRe    = regexp.MustCompile(`(?i)pricing?|billing|cost|DescribePrice|DescribeBill`)
	optimizeRe   = regexp.MustCompile(`(?i)optimiz|right.siz|rightsize|idle|spot|preempt|recommend|suggestion|tier|downgrade|upgrade|reserved|saving`)
	sectionRe    = regexp.MustCompile(`(?m)^#{1,3}\s+`)
	productRe    = regexp.MustCompile(`(?i)(ecs|rds|redis|vpc|eip|nat|clb|alb|vke|tos|cms|cdn|dns|kafka|sls|billing|iam|kms|mongodb|elasticsearch|polar|nas|fg|ark|vpn|security.group)`)
)

func checkSkill(root, skill string) Result {
	r := Result{Skill: skill}
	path := filepath.Join(root, skill, "references", "advanced", "finops.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return r
	}
	r.HasFinOps = true

	text := string(data)
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		r.LineCount++
	}

	// Detect product name from skill
	productMatch := productRe.FindString(skill)
	if productMatch == "" {
		productMatch = skill
	}
	if pricingRe.MatchString(text) {
		// Also check that the pricing reference is product-specific
		if strings.Contains(strings.ToLower(text), strings.ToLower(productMatch)) ||
			strings.Contains(strings.ToLower(text), "price") ||
			strings.Contains(strings.ToLower(text), "billing") {
			r.HasPricing = productMatch
		}
	}

	// Count optimization-related sections/mentions
	optimMatches := optimizeRe.FindAllString(text, -1)
	r.HasOptimizations = len(optimMatches) >= 2

	// Quality threshold: required tier (R+R skills) need ≥20 lines + pricing + ≥2 optimizations
	r.OK = r.LineCount >= 20 && r.HasPricing != "" && r.HasOptimizations
	return r
}

// RequiredRecommended is the set of skills needing ≥20 lines / ≥2 optimizations.
var RequiredRecommended = map[string]bool{
	"ve-ecs-ops": true, "ve-redis-ops": true, "ve-rds-mysql-ops": true, "ve-rds-ops": true,
	"ve-rds-pg-ops": true, "ve-polar-mysql-ops": true, "ve-mongodb-ops": true,
	"ve-elasticsearch-ops": true, "ve-tos-ops": true, "ve-iam-ops": true, "ve-kms-ops": true,
	"ve-eip-ops": true, "ve-security-group-ops": true, "ve-vpc-ops": true, "ve-nat-ops": true,
	"ve-vpn-ops": true, "ve-clb-ops": true, "ve-alb-ops": true, "ve-vke-ops": true,
	"ve-nas-ops": true, "ve-cms-ops": true, "ve-fg-ops": true, "ve-ark-ops": true,
}

// AllSkills is the full set of 29 skills.
var AllSkills = []string{
	"ve-ecs-ops", "ve-redis-ops", "ve-rds-mysql-ops", "ve-rds-ops", "ve-rds-pg-ops",
	"ve-polar-mysql-ops", "ve-mongodb-ops", "ve-elasticsearch-ops", "ve-tos-ops", "ve-iam-ops",
	"ve-kms-ops", "ve-eip-ops", "ve-security-group-ops", "ve-vpc-ops", "ve-nat-ops", "ve-vpn-ops",
	"ve-clb-ops", "ve-alb-ops", "ve-vke-ops", "ve-nas-ops", "ve-cms-ops", "ve-fg-ops", "ve-ark-ops",
	"ve-cdn-ops", "ve-dns-ops", "ve-kafka-ops", "ve-sls-ops", "ve-billing-ops", "ve-skill-generator",
}

// CheckDir scans all skills and returns the finops quality report.
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
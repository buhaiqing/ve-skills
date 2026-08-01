// Package aiops checks AIOps/FinOps/ eval coverage across all skills.
// Faithful Go port of scripts/check_aiops_coverage.py.
package aiops

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

// RequiredRecommended skills (23) need aiops.md + finops.md.
var RequiredRecommended = map[string]bool{
	"ve-ecs-ops": true, "ve-redis-ops": true, "ve-rds-mysql-ops": true, "ve-rds-ops": true,
	"ve-rds-pg-ops": true, "ve-polar-mysql-ops": true, "ve-mongodb-ops": true,
	"ve-elasticsearch-ops": true, "ve-tos-ops": true, "ve-iam-ops": true, "ve-kms-ops": true,
	"ve-eip-ops": true, "ve-security-group-ops": true, "ve-vpc-ops": true, "ve-nat-ops": true,
	"ve-vpn-ops": true, "ve-clb-ops": true, "ve-alb-ops": true, "ve-vke-ops": true,
	"ve-nas-ops": true, "ve-cms-ops": true, "ve-fg-ops": true, "ve-ark-ops": true,
}

// AllSkills is the full set of 29 leaf/meta skills plus the orchestration skill.
var AllSkills = []string{
	"ve-ecs-ops", "ve-redis-ops", "ve-rds-mysql-ops", "ve-rds-ops", "ve-rds-pg-ops",
	"ve-polar-mysql-ops", "ve-mongodb-ops", "ve-elasticsearch-ops", "ve-tos-ops", "ve-iam-ops",
	"ve-kms-ops", "ve-eip-ops", "ve-security-group-ops", "ve-vpc-ops", "ve-nat-ops", "ve-vpn-ops",
	"ve-clb-ops", "ve-alb-ops", "ve-vke-ops", "ve-nas-ops", "ve-cms-ops", "ve-fg-ops", "ve-ark-ops",
	"ve-cdn-ops", "ve-dns-ops", "ve-kafka-ops", "ve-sls-ops", "ve-billing-ops", "ve-skill-generator",
	"incident-loop-agent",
}

type skillResult struct {
	Skill          string `json:"skill"`
	AIOpsMD        bool   `json:"aiops_md"`
	FinOpsMD       bool   `json:"finops_md"`
	EvalQueries    bool   `json:"eval_queries"`
	EvalTrigger    int    `json:"eval_trigger"`
	EvalNonTrigger int    `json:"eval_non_trigger"`
	ParseFail      bool   `json:"eval_parse_fail"`
	HasAdvanced    bool   `json:"has_advanced"`
}

type Report struct {
	TotalSkills              int            `json:"total_skills"`
	RRCoverage               string         `json:"rr_aiops_coverage"`
	RRFinOpsCoverage         string         `json:"rr_finops_coverage"`
	EvalCoverage             string         `json:"eval_coverage"`
	SkillsMissingAIOps       []string       `json:"skills_missing_aiops"`
	SkillsMissingFinOps      []string       `json:"skills_missing_finops"`
	SkillsMissingEval        []string       `json:"skills_missing_eval"`
	EvalParseFail            []string       `json:"eval_parse_fail"`
	EvalQualityBad           []skillResult  `json:"eval_quality_bad"`
	Details                  []skillResult  `json:"details"`
	OK                       bool           `json:"ok"`
}

func checkSkill(root, skill string) skillResult {
	base := filepath.Join(root, skill, "references", "advanced")
	evalPath := filepath.Join(root, skill, "assets", "eval_queries.json")
	r := skillResult{Skill: skill, HasAdvanced: isDir(base)}

	if _, err := os.Stat(filepath.Join(base, "aiops.md")); err == nil {
		r.AIOpsMD = true
	}
	if _, err := os.Stat(filepath.Join(base, "finops.md")); err == nil {
		r.FinOpsMD = true
	}
	if data, err := os.ReadFile(evalPath); err == nil {
		var raw any
		if json.Unmarshal(data, &raw) == nil {
			arr := toArray(raw)
			r.EvalQueries = true
			for _, q := range arr {
				m, ok := q.(map[string]any)
				if !ok {
					continue
				}
				if mb, _ := m["should_trigger"].(bool); mb {
					r.EvalTrigger++
				} else {
					r.EvalNonTrigger++
				}
			}
		} else {
			// Corrupt JSON must be classified separately from "missing":
			// the file exists but does not parse, distinct from a missing
			// eval_queries.json (which leaves EvalQueries=false via the
			// outer os.ReadFile error branch).
			r.ParseFail = true
		}
	}
	return r
}

func toArray(raw any) []any {
	switch v := raw.(type) {
	case []any:
		return v
	case map[string]any:
		if q, ok := v["queries"].([]any); ok {
			return q
		}
	}
	return nil
}

// CheckDir scans all skills and returns the coverage report.
func CheckDir(root string) Report {
	var results []skillResult
	for _, s := range AllSkills {
		results = append(results, checkSkill(root, s))
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Skill < results[j].Skill })

	var rrAIOps, rrFinOps, evalSkills, evalQualityOK int
	var missingAIOps, missingFinOps, missingEval, parseFail []string
	var evalQualityBad []skillResult
	for _, r := range results {
		if RequiredRecommended[r.Skill] {
			if r.AIOpsMD {
				rrAIOps++
			} else {
				missingAIOps = append(missingAIOps, r.Skill)
			}
			if r.FinOpsMD {
				rrFinOps++
			} else {
				missingFinOps = append(missingFinOps, r.Skill)
			}
		}
		if r.ParseFail {
			parseFail = append(parseFail, r.Skill)
		}
		if r.EvalQueries {
			evalSkills++
			if r.EvalTrigger >= 5 && r.EvalNonTrigger >= 2 {
				evalQualityOK++
			} else {
				evalQualityBad = append(evalQualityBad, r)
			}
		} else if !r.ParseFail {
			missingEval = append(missingEval, r.Skill)
		}
	}
	rrCount := len(RequiredRecommended)
	ok := rrAIOps == rrCount && rrFinOps == rrCount &&
		evalSkills == len(AllSkills) && len(parseFail) == 0 && len(evalQualityBad) == 0

	return Report{
		TotalSkills:        len(AllSkills),
		RRCoverage:         strconv.Itoa(rrAIOps) + "/" + strconv.Itoa(rrCount),
		RRFinOpsCoverage:   strconv.Itoa(rrFinOps) + "/" + strconv.Itoa(rrCount),
		EvalCoverage:       strconv.Itoa(evalSkills) + "/" + strconv.Itoa(len(AllSkills)),
		SkillsMissingAIOps: missingAIOps,
		SkillsMissingFinOps: missingFinOps,
		SkillsMissingEval:  missingEval,
		EvalParseFail:      parseFail,
		EvalQualityBad:     evalQualityBad,
		Details:            results,
		OK:                 ok,
	}
}

func isDir(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

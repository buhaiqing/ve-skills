// Package gate implements `vet gcl gate` — a structural CI smoke test across
// all GCL-equipped skills.
//
// Faithful Go port of scripts/gcl_ci_gate.py: runs `vet gcl run` in
// structural-only mode with a no-op echo command for each skill, then reports
// how many skills pass the structural smoke.
package gate

import (
	"fmt"
	"sort"
	"strings"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/run"
)

// GCLSkills mirrors gcl_ci_gate.GCL_SKILLS (29 canonical + incident-loop-agent).
var GCLSkills = []string{
	"ve-ecs-ops", "ve-redis-ops", "ve-rds-mysql-ops", "ve-rds-ops", "ve-rds-pg-ops",
	"ve-polar-mysql-ops", "ve-mongodb-ops", "ve-elasticsearch-ops", "ve-tos-ops", "ve-iam-ops",
	"ve-kms-ops", "ve-eip-ops", "ve-security-group-ops", "ve-vpc-ops", "ve-nat-ops", "ve-vpn-ops",
	"ve-clb-ops", "ve-alb-ops", "ve-vke-ops", "ve-nas-ops", "ve-cms-ops", "ve-fg-ops", "ve-ark-ops",
	"ve-cdn-ops", "ve-dns-ops", "ve-kafka-ops", "ve-sls-ops", "ve-billing-ops", "ve-skill-generator",
	"incident-loop-agent",
}

// SmokeCommand mirrors gcl_ci_gate.SMOKE_COMMAND.
const SmokeCommand = `echo '{"Response":{"RequestId":"ci-gate-smoke"}}'`

// Report is the per-skill smoke result.
type Report struct {
	Skill      string `json:"skill"`
	OK         bool   `json:"ok"`
	ExitCode   int    `json:"exit_code"`
	TimedOut   bool   `json:"timed_out,omitempty"`
	TraceLine  string `json:"trace_line,omitempty"`
	StderrLine string `json:"stderr_first_line,omitempty"`
}

// SmokeSkill runs the structural smoke for a single skill.
func SmokeSkill(root, skill string) Report {
	r := Report{Skill: skill, ExitCode: -1}
	res := run.Run(run.Options{
		Root:           root,
		Skill:          skill,
		Request:        "CI gate smoke: " + skill,
		Command:        SmokeCommand,
		MaxIter:        1,
		Timeout:        30,
		StructuralOnly: true,
	})
	r.ExitCode = res.ExitCode
	r.TimedOut = res.TimedOut
	r.TraceLine = res.TraceLine
	r.StderrLine = res.StderrLine
	// runner returns 0 (PASS) or 1 (MAX_ITER) — both structurally acceptable
	r.OK = res.ExitCode == 0 || res.ExitCode == 1
	return r
}

// SmokeAll runs the smoke for the given skills (or all if nil), sorted.
func SmokeAll(root string, skills []string) []Report {
	target := skills
	if target == nil {
		target = append([]string{}, GCLSkills...)
	}
	sort.Strings(target)
	var reports []Report
	for _, s := range target {
		reports = append(reports, SmokeSkill(root, s))
	}
	return reports
}

// Run executes the gate and returns the process exit code (0 if all pass).
func Run(root string, skills []string, skipIncidentLoop bool, jsonOut bool) int {
	target := skills
	if target == nil && skipIncidentLoop {
		for _, s := range GCLSkills {
			if s != "incident-loop-agent" {
				target = append(target, s)
			}
		}
	}
	reports := SmokeAll(root, target)
	passing := 0
	for _, r := range reports {
		if r.OK {
			passing++
		}
	}
	if jsonOut {
		fmt.Printf("{\"summary\":{\"total\":%d,\"passing\":%d,\"failing\":%d},\"reports\":[",
			len(reports), passing, len(reports)-passing)
		for i, r := range reports {
			if i > 0 {
				fmt.Print(",")
			}
			if r.TimedOut || r.TraceLine != "" || r.StderrLine != "" {
				fmt.Printf("{\"skill\":%q,\"ok\":%t,\"exit_code\":%d,\"timed_out\":%t,\"trace_line\":%q,\"stderr_first_line\":%q}",
					r.Skill, r.OK, r.ExitCode, r.TimedOut, r.TraceLine, r.StderrLine)
			} else {
				fmt.Printf("{\"skill\":%q,\"ok\":%t,\"exit_code\":%d}", r.Skill, r.OK, r.ExitCode)
			}
		}
		fmt.Println("]}")
	} else {
		fmt.Printf("GCL CI gate: %d/%d skills pass structural smoke.\n", passing, len(reports))
		if passing != len(reports) {
			fmt.Println()
			for _, r := range reports {
				if !r.OK {
					reason := "exit_code=" + itoa(r.ExitCode)
					if r.StderrLine != "" {
						reason += " stderr=" + r.StderrLine
					} else if r.TraceLine != "" {
						reason += " " + r.TraceLine
					}
					if r.TimedOut {
						reason = "TIMEOUT"
					}
					fmt.Printf("  FAIL %s: %s\n", r.Skill, reason)
				}
			}
		}
	}
	if passing == len(reports) {
		return 0
	}
	return 1
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

var _ = strings.TrimSpace

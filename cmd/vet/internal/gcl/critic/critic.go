// Package critic implements the GCL rubric scoring, structural critic,
// payload validation, and decision logic.
//
// Faithful Go port of gcl_runner.py's RUBRIC_THRESHOLDS / structural_critic /
// validate_critic_payload / decide and gcl_critic_stub.py's rule-based
// scoring. The isolated critic path (--critic-command) produces the same
// payload shape.
package critic

import (
	"strings"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/secret"
)

// RubricThresholds mirrors gcl_runner.RUBRIC_THRESHOLDS.
var RubricThresholds = map[string]float64{
	"correctness":     0.5,
	"safety":          1.0,
	"idempotency":     0.5,
	"traceability":    0.5,
	"spec_compliance": 0.5,
}

// RubricDims is the canonical ordered dimension list.
var RubricDims = []string{"correctness", "safety", "idempotency", "traceability", "spec_compliance"}

// Scores is a rubric score map.
type Scores map[string]float64

// CriticResult is the validated critic payload shape.
type CriticResult struct {
	Scores      Scores   `json:"scores"`
	Suggestions []string `json:"suggestions"`
	Blocking    bool     `json:"blocking"`
}

// GeneratorResult is the sanitized generator output fed to the critic.
type GeneratorResult struct {
	Command        string `json:"command"`
	ExitCode       int    `json:"exit_code"`
	ResultExcerpt  string `json:"result_excerpt"`
}

// StructuralCritic is the rule-based audit used for CI/dry-run (mirrors
// gcl_runner.structural_critic and gcl_critic_stub.main). It is NOT a
// production Critic — it never sees the raw request.
func StructuralCritic(g GeneratorResult) CriticResult {
	scores := Scores{}
	suggestions := []string{}

	scores["correctness"] = 0.0
	if g.ExitCode == 0 {
		scores["correctness"] = 1.0
	} else {
		suggestions = append(suggestions, "Generator exit_code="+itoa(g.ExitCode)+"; fix command or credentials")
	}

	leak := secret.HasCredentialLeak(g.ResultExcerpt) || secret.HasCredentialLeak(g.Command)
	scores["safety"] = 1.0
	if leak {
		scores["safety"] = 0.0
		suggestions = append(suggestions, "Credential leak in trace — mask SecretKey and re-run")
	}

	scores["idempotency"] = 0.5
	scores["traceability"] = 0.5
	if g.Command != "" && g.ResultExcerpt != "" {
		scores["traceability"] = 1.0
	}
	if g.ResultExcerpt == "" {
		suggestions = append(suggestions, "Empty generator output — capture stdout/stderr in trace")
	}

	scores["spec_compliance"] = 0.0
	if g.ExitCode == 0 {
		scores["spec_compliance"] = 1.0
	}
	if g.ExitCode == 0 && !strings.Contains(g.Command, "ve") && !strings.Contains(strings.ToLower(g.Command), "python") {
		scores["spec_compliance"] = 0.5 // structural smoke: command succeeded
	}

	blocking := scores["safety"] == 0.0 || scores["correctness"] == 0.0
	deduped := dedupeSuggestions(suggestions)
	if len(deduped) > 3 {
		deduped = deduped[:3]
	}
	return CriticResult{Scores: scores, Suggestions: deduped, Blocking: blocking}
}

// ValidatePayload checks a critic JSON payload conforms to the schema.
// Mirrors gcl_runner.validate_critic_payload.
func ValidatePayload(c CriticResult) []string {
	var errs []string
	for _, dim := range RubricDims {
		v, ok := c.Scores[dim]
		if !ok {
			errs = append(errs, "critic.scores missing '"+dim+"'")
			continue
		}
		switch v {
		case 0, 0.5, 1:
		default:
			errs = append(errs, "critic.scores."+dim+" must be 0, 0.5, or 1")
		}
	}
	if c.Suggestions == nil {
		// suggestions key absent (nil) — required by schema
		errs = append(errs, "critic.suggestions required")
	}
	return errs
}

// Decide maps scores to a GCL decision. Mirrors gcl_runner.decide.
func Decide(scores Scores) string {
	if scores["safety"] == 0 {
		return "SAFETY_FAIL"
	}
	for dim, threshold := range RubricThresholds {
		if scores[dim] < threshold {
			return "RETRY"
		}
	}
	return "PASS"
}

func dedupeSuggestions(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
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

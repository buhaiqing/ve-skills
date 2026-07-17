// Package predict implements predictive (pre-alert) triggering for the GCL
// harness. A Predictor evaluates a single time-series Metric and returns a
// RiskLevel; RiskHigh means the loop should be triggered before an alarm fires.
//
// Predictors are pure: they only consume the Metric fed in by the caller and
// never invoke any `ve` CLI to fetch metrics. Real metric collection is wired
// by T16 / incident-loop-agent.
package predict

import (
	"context"
	"math"
)

// RiskLevel is the risk grade a predictor emits for a Metric.
type RiskLevel string

const (
	// RiskLow means monitor only; do not trigger.
	RiskLow RiskLevel = "low"
	// RiskMedium means record a HINT (consumed by Reflexion, T14); do not trigger.
	RiskMedium RiskLevel = "medium"
	// RiskHigh means trigger the loop.
	RiskHigh RiskLevel = "high"
)

// Metric is a single time-series sample fed in by the caller.
type Metric struct {
	Skill   string    // source skill, e.g. "ve-redis-ops"
	Name    string    // metric name, e.g. "slow_commands_per_sec"
	Current float64   // current value
	History []float64 // most recent N samples, ascending by time
}

// Evaluation is the result of one predictor run.
type Evaluation struct {
	Predictor string    // predictor name that produced this evaluation
	Risk      RiskLevel // risk grade
	Trigger   bool      // whether the loop should be triggered (RiskHigh)
	Detail    string    // human-readable reasoning
}

// Predictor evaluates a Metric. The second return value is false when the
// predictor does not apply to the given Metric (e.g. unknown Name); the caller
// then tries the next predictor in the Registry.
type Predictor interface {
	Name() string
	Evaluate(ctx context.Context, m Metric) (Evaluation, bool)
}

// Registry holds the built-in predictors and evaluates a Metric against the
// first applicable one.
type Registry struct {
	predictors []Predictor
}

// NewRegistry builds the default registry with the three built-in predictors.
func NewRegistry() *Registry {
	return &Registry{
		predictors: []Predictor{
			&redisSlowQueryDegrade{},
			&rdsCapacityWaterline{},
			&ecsCPUTrend{},
		},
	}
}

// Evaluate runs the first applicable predictor. If none applies, it returns a
// low-risk evaluation with a descriptive Detail.
func (r *Registry) Evaluate(ctx context.Context, m Metric) Evaluation {
	for _, p := range r.predictors {
		if eval, ok := p.Evaluate(ctx, m); ok {
			return eval
		}
	}
	return Evaluation{
		Predictor: "none",
		Risk:      RiskLow,
		Trigger:   false,
		Detail:    "no predictor matches metric name " + m.Name,
	}
}

// finite reports whether v is a finite (non-NaN, non-inf) number.
func finite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && !math.IsInf(v, 1)
}

// slope returns the least-squares slope of ys over indices 0..n-1.
// Returns 0 when fewer than 2 points.
func slope(ys []float64) float64 {
	n := len(ys)
	if n < 2 {
		return 0
	}
	var sx, sy, sxx, sxy float64
	for i, y := range ys {
		x := float64(i)
		sx += x
		sy += y
		sxx += x * x
		sxy += x * y
	}
	denom := sxx - sx*sx/float64(n)
	if denom == 0 {
		return 0
	}
	return (sxy - sx*sy/float64(n)) / denom
}

// pctRise reports how much the series rose over the last window points,
// relative to the oldest of those points. Returns 0 when window < 2 or the
// oldest point is non-positive (avoids divide-by-zero blowups).
func pctRise(ys []float64, window int) float64 {
	if window < 2 || len(ys) < window {
		return 0
	}
	tail := ys[len(ys)-window:]
	base := tail[0]
	if base <= 0 {
		return 0
	}
	return (tail[len(tail)-1] - base) / base
}

// mean returns the arithmetic mean; 0 for empty input.
func mean(ys []float64) float64 {
	if len(ys) == 0 {
		return 0
	}
	var s float64
	for _, y := range ys {
		s += y
	}
	return s / float64(len(ys))
}

// minWindow is the minimum History length a predictor requires to make a
// high-confidence call. Shorter history degrades to RiskLow to avoid jitter.
const minWindow = 5

// ---------------------------------------------------------------------------
// redis-slow-query-degrade: slow_commands_per_sec rising dangerously.
// ---------------------------------------------------------------------------

type redisSlowQueryDegrade struct{}

func (redisSlowQueryDegrade) Name() string { return "redis-slow-query-degrade" }

func (redisSlowQueryDegrade) Evaluate(ctx context.Context, m Metric) (Evaluation, bool) {
	if m.Name != "slow_commands_per_sec" {
		return Evaluation{}, false
	}
	if !finite(m.Current) {
		return low("redis-slow-query-degrade", "current is non-finite"), true
	}
	for _, h := range m.History {
		if !finite(h) {
			return low("redis-slow-query-degrade", "history contains non-finite value"), true
		}
	}
	const threshold = 100.0
	rise := pctRise(m.History, minWindow)
	switch {
	case len(m.History) >= minWindow && rise >= 0.5 && m.Current > threshold:
		return Evaluation{
			Predictor: "redis-slow-query-degrade",
			Risk:      RiskHigh,
			Trigger:   true,
			Detail:    "slow_commands_per_sec up >=50% over last 5 samples and exceeds threshold 100",
		}, true
	case len(m.History) >= minWindow && rise >= 0.5:
		return Evaluation{
			Predictor: "redis-slow-query-degrade",
			Risk:      RiskMedium,
			Trigger:   false,
			Detail:    "slow_commands_per_sec up >=50% but below threshold 100 (record HINT)",
		}, true
	default:
		return low("redis-slow-query-degrade", "history < 5 samples or rise < 50%"), true
	}
}

// ---------------------------------------------------------------------------
// rds-capacity-waterline: disk_usage_percent trending to full within 24h.
// ---------------------------------------------------------------------------

type rdsCapacityWaterline struct{}

func (rdsCapacityWaterline) Name() string { return "rds-capacity-waterline" }

func (rdsCapacityWaterline) Evaluate(ctx context.Context, m Metric) (Evaluation, bool) {
	if m.Name != "disk_usage_percent" {
		return Evaluation{}, false
	}
	if !finite(m.Current) {
		return low("rds-capacity-waterline", "current is non-finite"), true
	}
	for _, h := range m.History {
		if !finite(h) {
			return low("rds-capacity-waterline", "history contains non-finite value"), true
		}
	}
	// Linear extrapolation: Current + slope * (samples in 24h).
	// Assume one sample per hour (24 samples/day) — conservative default; the
	// caller controls sampling density via History.
	const samplesPerDay = 24.0
	s := slope(m.History)
	proj := m.Current + s*samplesPerDay
	switch {
	case proj >= 90:
		return Evaluation{
			Predictor: "rds-capacity-waterline",
			Risk:      RiskHigh,
			Trigger:   true,
			Detail:    "disk_usage_percent projected >=90% within 24h (linear extrapolation)",
		}, true
	case proj >= 80:
		return Evaluation{
			Predictor: "rds-capacity-waterline",
			Risk:      RiskMedium,
			Trigger:   false,
			Detail:    "disk_usage_percent projected >=80% within 24h (record HINT)",
		}, true
	default:
		return low("rds-capacity-waterline", "disk_usage_percent projection < 80% within 24h"), true
	}
}

// ---------------------------------------------------------------------------
// ecs-cpu-trend: cpu_usage_percent sustained high with positive slope.
// ---------------------------------------------------------------------------

type ecsCPUTrend struct{}

func (ecsCPUTrend) Name() string { return "ecs-cpu-trend" }

func (ecsCPUTrend) Evaluate(ctx context.Context, m Metric) (Evaluation, bool) {
	if m.Name != "cpu_usage_percent" {
		return Evaluation{}, false
	}
	if !finite(m.Current) {
		return low("ecs-cpu-trend", "current is non-finite"), true
	}
	for _, h := range m.History {
		if !finite(h) {
			return low("ecs-cpu-trend", "history contains non-finite value"), true
		}
	}
	avg := mean(m.History)
	s := slope(m.History)
	switch {
	case len(m.History) >= minWindow && avg > 70 && s > 0:
		return Evaluation{
			Predictor: "ecs-cpu-trend",
			Risk:      RiskHigh,
			Trigger:   true,
			Detail:    "cpu_usage_percent 1h mean > 70% with positive slope",
		}, true
	case len(m.History) >= minWindow && avg > 70:
		return Evaluation{
			Predictor: "ecs-cpu-trend",
			Risk:      RiskMedium,
			Trigger:   false,
			Detail:    "cpu_usage_percent 1h mean > 70% (record HINT)",
		}, true
	default:
		return low("ecs-cpu-trend", "history < 5 samples or 1h mean <= 70%"), true
	}
}

func low(name, detail string) Evaluation {
	return Evaluation{Predictor: name, Risk: RiskLow, Trigger: false, Detail: detail}
}

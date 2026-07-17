// Package heal implements error-classification-driven retry and its
// observability for `vet gcl run` (L4 self-healing). See retry.go for the
// retry engine and this file for the metrics accumulator.
package heal

import (
	"os"
	"path/filepath"
)

// Target thresholds for the L4 self-healing SLOs (framework §6.1).
const (
	TargetSuccessRate      = 0.80
	TargetAvgDurationMs    = 30000.0
	TargetUserIntervention = 0.20
	TargetFallback         = 0.10
)

// HealEvent is one persisted self-healing attempt, written as a JSONL line so
// `heal-stats` can re-aggregate without keeping the process alive.
type HealEvent struct {
	ISO        string `json:"iso"`
	EventType  string `json:"event_type"`
	ErrorCode  string `json:"error_code"`
	Action     string `json:"action"`
	Result     string `json:"result"` // "ok" | "fail"
	DurationMs int64  `json:"duration_ms"`
}

// Metrics accumulates self-healing outcomes across a run (or process lifetime).
// UserInterventions / FallbackUsed are populated from signals available today;
// FallbackUsed counts multi-path fallbacks (T10): when the best path fails and
// a secondary path is tried.
type Metrics struct {
	SuccessCount      int64 `json:"success_count"`
	TotalCount        int64 `json:"total_count"`
	DurationSumMs     int64 `json:"duration_sum_ms"`
	UserInterventions int64 `json:"user_interventions"`
	FallbackUsed      int64 `json:"fallback_used"`
	// PerPath holds per-(class/name) success counters so SelectBest can prefer
	// the historically reliable path. Key format: "<class>/<path>".
	PerPath map[string]*PathStat `json:"per_path,omitempty"`
}

// PathStat is the per-path success tally backing the History interface.
type PathStat struct {
	Success int64 `json:"success"`
	Total   int64 `json:"total"`
}

// Record folds one HealEvent into the aggregate. ok increments SuccessCount.
// It also accumulates the per-path counter keyed by ErrorCode(class)/Action(path).
func (m *Metrics) Record(e HealEvent) {
	m.TotalCount++
	if e.Result == "ok" {
		m.SuccessCount++
	}
	m.DurationSumMs += e.DurationMs
	key := e.ErrorCode + "/" + e.Action
	if m.PerPath == nil {
		m.PerPath = map[string]*PathStat{}
	}
	ps := m.PerPath[key]
	if ps == nil {
		ps = &PathStat{}
		m.PerPath[key] = ps
	}
	ps.Total++
	if e.Result == "ok" {
		ps.Success++
	}
}

// PathSuccessRate implements History: returns 0.0..1.0 for class+name, or 0.0
// when no data (no division by zero). Distinct from the no-arg SuccessRate()
// aggregate used by heal-stats (T11).
func (m *Metrics) PathSuccessRate(class, name string) float64 {
	if m.PerPath == nil {
		return 0.0
	}
	ps := m.PerPath[class+"/"+name]
	if ps == nil || ps.Total == 0 {
		return 0.0
	}
	return float64(ps.Success) / float64(ps.Total)
}

// SuccessRate returns 0.0..1.0; 0.0 when no events recorded (no division by zero).
func (m *Metrics) SuccessRate() float64 {
	if m.TotalCount == 0 {
		return 0.0
	}
	return float64(m.SuccessCount) / float64(m.TotalCount)
}

// AvgDurationMs returns the mean healing time in ms; 0.0 when no events.
func (m *Metrics) AvgDurationMs() float64 {
	if m.TotalCount == 0 {
		return 0.0
	}
	return float64(m.DurationSumMs) / float64(m.TotalCount)
}

// UserInterventionRate returns 0.0..1.0; 0.0 when no events.
func (m *Metrics) UserInterventionRate() float64 {
	if m.TotalCount == 0 {
		return 0.0
	}
	return float64(m.UserInterventions) / float64(m.TotalCount)
}

// FallbackRate returns 0.0..1.0; 0.0 when no events.
func (m *Metrics) FallbackRate() float64 {
	if m.TotalCount == 0 {
		return 0.0
	}
	return float64(m.FallbackUsed) / float64(m.TotalCount)
}

// Persist appends one event to the §6.2 log file as a pipe-delimited row, so
// `heal-stats` can re-aggregate it later via ParseFile. Append-only so
// concurrent runs accumulate rather than overwrite.
func (m *Metrics) Persist(path string, e HealEvent) error {
	// OpenFile won't create the parent dir; do it here so a fresh repo root
	// (audit-results/ does not yet exist) does not make every heal fail.
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	return AppendEvent(f, Event{
		ISO:        e.ISO,
		EventType:  e.EventType,
		ErrorCode:  e.ErrorCode,
		Action:     e.Action,
		Result:     e.Result,
		DurationMs: e.DurationMs,
	})
}

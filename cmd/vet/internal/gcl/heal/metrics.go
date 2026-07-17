// Package heal implements error-classification-driven retry and its
// observability for `vet gcl run` (L4 self-healing). See retry.go for the
// retry engine and this file for the metrics accumulator.
package heal

import (
	"os"
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
// FallbackUsed stays 0 until T10 (multi-path healing) lands — see spec §2.2.
type Metrics struct {
	SuccessCount      int64 `json:"success_count"`
	TotalCount        int64 `json:"total_count"`
	DurationSumMs     int64 `json:"duration_sum_ms"`
	UserInterventions int64 `json:"user_interventions"`
	FallbackUsed      int64 `json:"fallback_used"`
}

// Record folds one HealEvent into the aggregate. ok increments SuccessCount.
func (m *Metrics) Record(e HealEvent) {
	m.TotalCount++
	if e.Result == "ok" {
		m.SuccessCount++
	}
	m.DurationSumMs += e.DurationMs
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

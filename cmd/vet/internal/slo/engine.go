package slo

import (
	"fmt"
	"sync"
	"time"
)

// SLOStatus represents the current state of an SLO.
type SLOStatus int

const (
	// StatusHealthy indicates the SLO is being met.
	StatusHealthy SLOStatus = iota
	// StatusWarning indicates the metric is approaching the target (80-100%).
	StatusWarning
	// StatusCritical indicates the metric has exceeded the target (100-120%).
	StatusCritical
	// StatusViolated indicates the SLO has been violated for the full window.
	StatusViolated
)

// String returns the string representation of SLOStatus.
func (s SLOStatus) String() string {
	switch s {
	case StatusHealthy:
		return "healthy"
	case StatusWarning:
		return "warning"
	case StatusCritical:
		return "critical"
	case StatusViolated:
		return "violated"
	default:
		return "unknown"
	}
}

// SLO defines a Service Level Objective.
type SLO struct {
	Name     string        // "redis-p99-latency"
	Skill    string        // "ve-redis-ops"
	Metric   string        // "p99_latency_ms"
	Target   float64       // 100 (ms)
	Window   time.Duration // 5m
	BurnRate float64       // 2.0 (alert threshold)
}

// Metric represents an observed metric value.
type Metric struct {
	Name  string
	Value float64
	Time  time.Time
	Tags  map[string]string
}

// Action represents a recommended action based on SLO state.
type Action struct {
	Type    string // "none", "predictive_trigger", "escalate", "rollback"
	Reason  string
	Skill   string
	Urgency string // "low", "medium", "high"
}

// SLOStatusEntry contains the status of an SLO.
type SLOStatusEntry struct {
	Name   string
	Status SLOStatus
}

// sloState tracks the internal state of a single SLO.
type sloState struct {
	status      SLOStatus
	lastObserve time.Time
	burnRate    float64
	violatedAt  time.Time // when the SLO first entered Violated state
}

// Engine manages multiple SLO states.
type Engine struct {
	slos   []SLO
	states map[string]*sloState
	mu     sync.RWMutex
}

// NewEngine creates a new SLO Engine with the given SLOs.
func NewEngine(slos []SLO) *Engine {
	states := make(map[string]*sloState, len(slos))
	for _, s := range slos {
		states[s.Name] = &sloState{
			status: StatusHealthy,
		}
	}
	return &Engine{
		slos:   slos,
		states: states,
	}
}

// Observe feeds a metric value into the engine and returns the current SLO status.
func (e *Engine) Observe(metric Metric) (SLOStatus, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Find the SLO for this metric
	for _, slo := range e.slos {
		if slo.Metric == metric.Name {
			state := e.states[slo.Name]
			state.lastObserve = metric.Time

			// Calculate status based on metric value vs target
			ratio := metric.Value / slo.Target
			var newStatus SLOStatus

			switch {
			case ratio <= 0.8:
				newStatus = StatusHealthy
			case ratio <= 1.0:
				newStatus = StatusWarning
			case ratio <= 1.2:
				newStatus = StatusCritical
			default:
				newStatus = StatusViolated
			}

			// Track when we first entered Violated state
			if newStatus == StatusViolated {
				if state.violatedAt.IsZero() {
					state.violatedAt = metric.Time
				}
				// Check if we've been violated for the full window
				if slo.Window > 0 && metric.Time.Sub(state.violatedAt) >= slo.Window {
					// Already violated for full window, stay violated
				} else if slo.Window > 0 && metric.Time.Sub(state.violatedAt) < slo.Window {
					// Not yet violated for full window, treat as Critical
					newStatus = StatusCritical
				}
			} else {
				state.violatedAt = time.Time{}
			}

			// Calculate burn rate
			state.burnRate = ratio
			state.status = newStatus

			return newStatus, nil
		}
	}

	return StatusHealthy, fmt.Errorf("unknown metric: %s", metric.Name)
}

// RecommendAction recommends an action based on the SLO status for a skill.
func (e *Engine) RecommendAction(skill string) Action {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, slo := range e.slos {
		if slo.Skill == skill {
			state := e.states[slo.Name]
			switch state.status {
			case StatusHealthy:
				return Action{Type: "none", Skill: skill, Urgency: "low"}
			case StatusWarning:
				return Action{
					Type:    "predictive_trigger",
					Reason:  fmt.Sprintf("SLO %s in warning state (burn rate: %.1f)", slo.Name, state.burnRate),
					Skill:   skill,
					Urgency: "medium",
				}
			case StatusCritical:
				return Action{
					Type:    "escalate",
					Reason:  fmt.Sprintf("SLO %s in critical state (burn rate: %.1f)", slo.Name, state.burnRate),
					Skill:   skill,
					Urgency: "high",
				}
			case StatusViolated:
				return Action{
					Type:    "rollback",
					Reason:  fmt.Sprintf("SLO %s violated (burn rate: %.1f)", slo.Name, state.burnRate),
					Skill:   skill,
					Urgency: "high",
				}
			}
		}
	}

	return Action{Type: "none", Skill: skill, Urgency: "low"}
}

// GetStatus returns the current status of the named SLO.
func (e *Engine) GetStatus(sloName string) SLOStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if state, ok := e.states[sloName]; ok {
		return state.status
	}
	return StatusHealthy
}

// ListStatuses returns the status of all SLOs.
func (e *Engine) ListStatuses() []SLOStatusEntry {
	e.mu.RLock()
	defer e.mu.RUnlock()

	entries := make([]SLOStatusEntry, 0, len(e.slos))
	for _, slo := range e.slos {
		entries = append(entries, SLOStatusEntry{
			Name:   slo.Name,
			Status: e.states[slo.Name].status,
		})
	}
	return entries
}

package slo

import (
	"testing"
	"time"
)

func TestNewEngine(t *testing.T) {
	slos := []SLO{
		{Name: "redis-p99", Skill: "ve-redis-ops", Metric: "p99_latency_ms", Target: 100, Window: 5 * time.Minute},
	}
	engine := NewEngine(slos)
	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}
	if len(engine.slos) != 1 {
		t.Fatalf("expected 1 SLO, got %d", len(engine.slos))
	}
}

func TestObserve_Healthy(t *testing.T) {
	slos := []SLO{
		{Name: "test-slo", Skill: "test-skill", Metric: "test_metric", Target: 100, Window: 5 * time.Minute},
	}
	engine := NewEngine(slos)

	status, err := engine.Observe(Metric{
		Name:  "test_metric",
		Value: 50,
		Time:  time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != StatusHealthy {
		t.Errorf("expected StatusHealthy, got %v", status)
	}
}

func TestObserve_Warning(t *testing.T) {
	slos := []SLO{
		{Name: "test-slo", Skill: "test-skill", Metric: "test_metric", Target: 100, Window: 5 * time.Minute},
	}
	engine := NewEngine(slos)

	status, err := engine.Observe(Metric{
		Name:  "test_metric",
		Value: 85,
		Time:  time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != StatusWarning {
		t.Errorf("expected StatusWarning, got %v", status)
	}
}

func TestObserve_Critical(t *testing.T) {
	slos := []SLO{
		{Name: "test-slo", Skill: "test-skill", Metric: "test_metric", Target: 100, Window: 5 * time.Minute},
	}
	engine := NewEngine(slos)

	status, err := engine.Observe(Metric{
		Name:  "test_metric",
		Value: 105,
		Time:  time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != StatusCritical {
		t.Errorf("expected StatusCritical, got %v", status)
	}
}

func TestObserve_Violated(t *testing.T) {
	slos := []SLO{
		{Name: "test-slo", Skill: "test-skill", Metric: "test_metric", Target: 100, Window: 5 * time.Minute},
	}
	engine := NewEngine(slos)

	now := time.Now()

	// First observation - Critical (not yet violated for full window)
	status, err := engine.Observe(Metric{
		Name:  "test_metric",
		Value: 150,
		Time:  now,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != StatusCritical {
		t.Errorf("expected StatusCritical (not yet violated for window), got %v", status)
	}

	// Second observation after window - Violated
	status, err = engine.Observe(Metric{
		Name:  "test_metric",
		Value: 150,
		Time:  now.Add(6 * time.Minute),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != StatusViolated {
		t.Errorf("expected StatusViolated after window, got %v", status)
	}
}

func TestObserve_UnknownMetric(t *testing.T) {
	slos := []SLO{
		{Name: "test-slo", Skill: "test-skill", Metric: "test_metric", Target: 100, Window: 5 * time.Minute},
	}
	engine := NewEngine(slos)

	_, err := engine.Observe(Metric{
		Name:  "unknown_metric",
		Value: 50,
		Time:  time.Now(),
	})
	if err == nil {
		t.Error("expected error for unknown metric, got nil")
	}
}

func TestRecommendAction_Healthy(t *testing.T) {
	slos := []SLO{
		{Name: "test-slo", Skill: "test-skill", Metric: "test_metric", Target: 100, Window: 5 * time.Minute},
	}
	engine := NewEngine(slos)

	action := engine.RecommendAction("test-skill")
	if action.Type != "none" {
		t.Errorf("expected action type 'none', got %q", action.Type)
	}
}

func TestRecommendAction_Warning(t *testing.T) {
	slos := []SLO{
		{Name: "test-slo", Skill: "test-skill", Metric: "test_metric", Target: 100, Window: 5 * time.Minute},
	}
	engine := NewEngine(slos)

	engine.Observe(Metric{
		Name:  "test_metric",
		Value: 85,
		Time:  time.Now(),
	})

	action := engine.RecommendAction("test-skill")
	if action.Type != "predictive_trigger" {
		t.Errorf("expected action type 'predictive_trigger', got %q", action.Type)
	}
	if action.Urgency != "medium" {
		t.Errorf("expected urgency 'medium', got %q", action.Urgency)
	}
}

func TestRecommendAction_Critical(t *testing.T) {
	slos := []SLO{
		{Name: "test-slo", Skill: "test-skill", Metric: "test_metric", Target: 100, Window: 5 * time.Minute},
	}
	engine := NewEngine(slos)

	engine.Observe(Metric{
		Name:  "test_metric",
		Value: 105,
		Time:  time.Now(),
	})

	action := engine.RecommendAction("test-skill")
	if action.Type != "escalate" {
		t.Errorf("expected action type 'escalate', got %q", action.Type)
	}
	if action.Urgency != "high" {
		t.Errorf("expected urgency 'high', got %q", action.Urgency)
	}
}

func TestGetStatus(t *testing.T) {
	slos := []SLO{
		{Name: "slo-1", Skill: "skill-1", Metric: "metric-1", Target: 100, Window: 5 * time.Minute},
		{Name: "slo-2", Skill: "skill-2", Metric: "metric-2", Target: 200, Window: 5 * time.Minute},
	}
	engine := NewEngine(slos)

	// Initially healthy
	if status := engine.GetStatus("slo-1"); status != StatusHealthy {
		t.Errorf("expected StatusHealthy, got %v", status)
	}

	// After observe
	engine.Observe(Metric{Name: "metric-1", Value: 85, Time: time.Now()})
	if status := engine.GetStatus("slo-1"); status != StatusWarning {
		t.Errorf("expected StatusWarning, got %v", status)
	}

	// Unknown SLO
	if status := engine.GetStatus("unknown"); status != StatusHealthy {
		t.Errorf("expected StatusHealthy for unknown SLO, got %v", status)
	}
}

func TestListStatuses(t *testing.T) {
	slos := []SLO{
		{Name: "slo-1", Skill: "skill-1", Metric: "metric-1", Target: 100, Window: 5 * time.Minute},
		{Name: "slo-2", Skill: "skill-2", Metric: "metric-2", Target: 200, Window: 5 * time.Minute},
	}
	engine := NewEngine(slos)

	entries := engine.ListStatuses()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// Both should be healthy initially
	for _, entry := range entries {
		if entry.Status != StatusHealthy {
			t.Errorf("expected StatusHealthy for %s, got %v", entry.Name, entry.Status)
		}
	}
}

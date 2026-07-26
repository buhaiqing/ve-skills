package slo

import (
	"math"
	"testing"
	"time"
)

const epsilon = 1e-9

func floatEqual(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

func TestCalculateBudget(t *testing.T) {
	slo := SLO{
		Name:       "test-slo",
		Target:     100,
		Comparator: CompareGreaterThan,
	}
	now := time.Now()
	window := 30 * time.Minute

	budget := CalculateBudget(slo, 100.2, now, window)

	if !floatEqual(budget.Consumed, 0.2) {
		t.Errorf("expected Consumed≈0.2, got %f", budget.Consumed)
	}
	if budget.Grade != "warn" {
		t.Errorf("expected Grade=warn, got %s", budget.Grade)
	}
	if !floatEqual(budget.TotalBudget, 1.0) {
		t.Errorf("expected TotalBudget≈1.0, got %f", budget.TotalBudget)
	}
	if !floatEqual(budget.Remaining, 0.8) {
		t.Errorf("expected Remaining≈0.8, got %f", budget.Remaining)
	}
	if budget.BurnRate <= 0.3 || budget.BurnRate > 0.6 {
		t.Errorf("expected BurnRate in warn range (0.3-0.6], got %f", budget.BurnRate)
	}
	if budget.ExhaustedAt.IsZero() {
		t.Error("expected ExhaustedAt to be set")
	}
}

func TestCalculateBudgetEmpty(t *testing.T) {
	slo := SLO{
		Name:       "test-slo",
		Target:     100,
		Comparator: CompareGreaterThan,
	}
	now := time.Now()
	window := 5 * time.Minute

	budget := CalculateBudget(slo, 50, now, window)

	if !floatEqual(budget.Consumed, 0) {
		t.Errorf("expected Consumed≈0, got %f", budget.Consumed)
	}
	if budget.Grade != "healthy" {
		t.Errorf("expected Grade=healthy, got %s", budget.Grade)
	}
	if !floatEqual(budget.BurnRate, 0) {
		t.Errorf("expected BurnRate≈0, got %f", budget.BurnRate)
	}
	if !floatEqual(budget.Remaining, 1.0) {
		t.Errorf("expected Remaining≈1.0, got %f", budget.Remaining)
	}
	if !budget.ExhaustedAt.IsZero() {
		t.Error("expected ExhaustedAt to be zero when no violation")
	}
}

func TestBudgetGrade(t *testing.T) {
	tests := []struct {
		name     string
		target   float64
		actual   float64
		window   time.Duration
		expected string
	}{
		{"healthy low", 100, 100.1, 1 * time.Hour, "healthy"},
		{"healthy boundary", 100, 100.3, 1 * time.Hour, "healthy"},
		{"warn", 100, 100.5, 1 * time.Hour, "warn"},
		{"burn", 100, 100.75, 1 * time.Hour, "burn"},
		{"fried", 100, 102.0, 30 * time.Minute, "fried"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slo := SLO{
				Name:       "test-slo",
				Target:     tt.target,
				Comparator: CompareGreaterThan,
			}
			budget := CalculateBudget(slo, tt.actual, time.Now(), tt.window)
			if budget.Grade != tt.expected {
				t.Errorf("expected Grade=%s, got %s (BurnRate=%f)", tt.expected, budget.Grade, budget.BurnRate)
			}
		})
	}
}

func TestAutoScaleByBudget(t *testing.T) {
	tests := []struct {
		name     string
		budget   ErrorBudget
		expected string
	}{
		{
			name:     "fried -> scale_emergency",
			budget:   ErrorBudget{Consumed: 0.5, Grade: "fried"},
			expected: "scale_emergency",
		},
		{
			name:     "burn -> scale_up",
			budget:   ErrorBudget{Consumed: 0.5, Grade: "burn"},
			expected: "scale_up",
		},
		{
			name:     "warn -> prepare_scale",
			budget:   ErrorBudget{Consumed: 0.5, Grade: "warn"},
			expected: "prepare_scale",
		},
		{
			name:     "healthy -> maintain",
			budget:   ErrorBudget{Consumed: 0.2, Grade: "healthy"},
			expected: "maintain",
		},
		{
			name:     "no consumption -> scale_down",
			budget:   ErrorBudget{Consumed: 0, Grade: "healthy"},
			expected: "scale_down",
		},
	}

	for i := range tests {
		name := tests[i].name
		expected := tests[i].expected
		budget := &tests[i].budget
		t.Run(name, func(t *testing.T) {
			action := AutoScaleByBudget(budget)
			if action != expected {
				t.Errorf("expected %s, got %s", expected, action)
			}
		})
	}
}

func TestCalculateBudgetZeroTime(t *testing.T) {
	slo := SLO{
		Name:       "test-slo",
		Target:     100,
		Comparator: CompareGreaterThan,
	}

	t.Run("zero window", func(t *testing.T) {
		budget := CalculateBudget(slo, 100.2, time.Now(), 0)
		if !floatEqual(budget.Consumed, 0.2) {
			t.Errorf("expected Consumed≈0.2, got %f", budget.Consumed)
		}
		if budget.Grade == "" {
			t.Error("expected Grade to be set")
		}
	})

	t.Run("negative window", func(t *testing.T) {
		budget := CalculateBudget(slo, 100.2, time.Now(), -1*time.Minute)
		if !floatEqual(budget.Consumed, 0.2) {
			t.Errorf("expected Consumed≈0.2, got %f", budget.Consumed)
		}
	})

	t.Run("zero now time", func(t *testing.T) {
		budget := CalculateBudget(slo, 100.2, time.Time{}, 30*time.Minute)
		if !floatEqual(budget.Consumed, 0.2) {
			t.Errorf("expected Consumed≈0.2, got %f", budget.Consumed)
		}
		if budget.BurnRate <= 0.3 || budget.BurnRate > 0.6 {
			t.Errorf("expected BurnRate in warn range, got %f", budget.BurnRate)
		}
		if budget.ExhaustedAt.IsZero() {
			t.Error("expected ExhaustedAt to be computed even with zero now time")
		}
	})
}
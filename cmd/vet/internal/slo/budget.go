package slo

import (
	"sync"
	"time"
)

const totalBudget = 1.0

type ErrorBudget struct {
	SLO         SLO
	TotalBudget float64
	Consumed    float64
	Remaining   float64
	BurnRate    float64
	Grade       string
	ExhaustedAt time.Time
	mu          sync.Mutex
}

func CalculateBudget(slo SLO, actualValue float64, now time.Time, window time.Duration) *ErrorBudget {
	budget := &ErrorBudget{
		SLO:         slo,
		TotalBudget: totalBudget,
		Remaining:   totalBudget,
		Grade:       "healthy",
	}

	elapsedHours := window.Hours()
	if elapsedHours <= 0 {
		elapsedHours = 1
	}

	violated := false
	var consumed float64

	switch slo.Comparator {
	case CompareLessThan:
		if actualValue < slo.Target {
			violated = true
			diff := slo.Target - actualValue
			consumed = diff
		}
	default:
		if actualValue > slo.Target {
			violated = true
			diff := actualValue - slo.Target
			consumed = diff
		}
	}

	if violated {
		if consumed > totalBudget {
			consumed = totalBudget
		}
		budget.Consumed = consumed
		budget.Remaining = totalBudget - consumed
		budget.BurnRate = consumed / elapsedHours

		switch {
		case budget.BurnRate <= 0.3:
			budget.Grade = "healthy"
		case budget.BurnRate <= 0.6:
			budget.Grade = "warn"
		case budget.BurnRate <= 0.9:
			budget.Grade = "burn"
		default:
			budget.Grade = "fried"
		}

		if budget.BurnRate > 0 && budget.Remaining > 0 {
			hoursRemaining := budget.Remaining / budget.BurnRate
			budget.ExhaustedAt = now.Add(time.Duration(hoursRemaining * float64(time.Hour)))
		}
	}

	return budget
}

func AutoScaleByBudget(budget *ErrorBudget) string {
	if budget.Consumed == 0 {
		return "scale_down"
	}

	switch budget.Grade {
	case "fried":
		return "scale_emergency"
	case "burn":
		return "scale_up"
	case "warn":
		return "prepare_scale"
	default:
		return "maintain"
	}
}
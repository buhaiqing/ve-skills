package predict

import (
	"context"
	"math"
	"testing"
)

func TestRedisSlowQueryDegrade(t *testing.T) {
	ctx := context.Background()
	p := NewRegistry()

	// Non-applicable metric name → registry falls through to "none".
	if e := p.Evaluate(ctx, Metric{Name: "other", Current: 200, History: rising(60, 130)}); e.Predictor != "none" {
		t.Fatalf("expected none for non-matching name, got %q", e.Predictor)
	}

	// History < 5 samples → degrade to low (no trigger).
	if e := p.Evaluate(ctx, Metric{Name: "slow_commands_per_sec", Current: 500, History: rising(150, 400)[:4]}); e.Risk != RiskLow || e.Trigger {
		t.Fatalf("expected low/no-trigger for short history, got %s trigger=%v", e.Risk, e.Trigger)
	}

	// History >= 5, rise >= 50% but below threshold 100 → Medium (HINT), no trigger.
	if e := p.Evaluate(ctx, Metric{Name: "slow_commands_per_sec", Current: 90, History: rising(40, 70)}); e.Risk != RiskMedium || e.Trigger {
		t.Fatalf("expected medium/no-trigger below threshold, got %s trigger=%v", e.Risk, e.Trigger)
	}

	// Threshold boundary: Current exactly 100 (not > 100) with rise >= 50% → Medium.
	if e := p.Evaluate(ctx, Metric{Name: "slow_commands_per_sec", Current: 100, History: rising(40, 80)}); e.Risk != RiskMedium || e.Trigger {
		t.Fatalf("expected medium at threshold boundary (Current=100 not >100), got %s trigger=%v", e.Risk, e.Trigger)
	}

	// Rise boundary: Current > 100 but rise < 50% over last 5 → Low.
	if e := p.Evaluate(ctx, Metric{Name: "slow_commands_per_sec", Current: 200, History: steady(120)}); e.Risk != RiskLow || e.Trigger {
		t.Fatalf("expected low when rise < 50%%, got %s trigger=%v", e.Risk, e.Trigger)
	}

	// High: Current > 100 AND rise >= 50% over last 5 → High + trigger.
	if e := p.Evaluate(ctx, Metric{Name: "slow_commands_per_sec", Current: 200, History: rising(40, 130)}); e.Risk != RiskHigh || !e.Trigger {
		t.Fatalf("expected high+trigger, got %s trigger=%v", e.Risk, e.Trigger)
	}

	// NaN current → low, no panic.
	if e := p.Evaluate(ctx, Metric{Name: "slow_commands_per_sec", Current: math.NaN(), History: rising(40, 130)}); e.Risk != RiskLow || e.Trigger {
		t.Fatalf("expected low for NaN current, got %s trigger=%v", e.Risk, e.Trigger)
	}

	// Non-finite value inside history → low.
	bad := rising(40, 130)
	bad[2] = math.Inf(1)
	if e := p.Evaluate(ctx, Metric{Name: "slow_commands_per_sec", Current: 200, History: bad}); e.Risk != RiskLow || e.Trigger {
		t.Fatalf("expected low for non-finite history, got %s trigger=%v", e.Risk, e.Trigger)
	}
}

func TestRdsCapacityWaterline(t *testing.T) {
	ctx := context.Background()
	p := NewRegistry()

	// Non-applicable name.
	if e := p.Evaluate(ctx, Metric{Name: "cpu_usage_percent", Current: 95, History: rising(70, 95)}); e.Predictor == "rds-capacity-waterline" {
		t.Fatalf("rds predictor should not fire for cpu metric")
	}

	// NaN current → low.
	if e := p.Evaluate(ctx, Metric{Name: "disk_usage_percent", Current: math.NaN(), History: rising(70, 95)}); e.Risk != RiskLow || e.Trigger {
		t.Fatalf("expected low for NaN current, got %s trigger=%v", e.Risk, e.Trigger)
	}

	// Projection >= 90 → High + trigger.
	// Current 70 with slope +1/hr → +24 in 24h → proj 94.
	h := linear(70, 1)
	if e := p.Evaluate(ctx, Metric{Name: "disk_usage_percent", Current: 70, History: h}); e.Risk != RiskHigh || !e.Trigger {
		t.Fatalf("expected high+trigger for proj>=90, got %s trigger=%v", e.Risk, e.Trigger)
	}

	// Projection boundary exactly 90 → High + trigger (>=).
	if e := p.Evaluate(ctx, Metric{Name: "disk_usage_percent", Current: 66, History: linear(66, 1)}); e.Risk != RiskHigh || !e.Trigger {
		t.Fatalf("expected high+trigger at proj=90 boundary, got %s trigger=%v", e.Risk, e.Trigger)
	}

	// Projection in [80, 90) → Medium (HINT), no trigger.
	if e := p.Evaluate(ctx, Metric{Name: "disk_usage_percent", Current: 70, History: linear(70, 0.5)}); e.Risk != RiskMedium || e.Trigger {
		t.Fatalf("expected medium for proj in [80,90), got %s trigger=%v", e.Risk, e.Trigger)
	}

	// Projection < 80 → Low.
	if e := p.Evaluate(ctx, Metric{Name: "disk_usage_percent", Current: 70, History: linear(70, 0.1)}); e.Risk != RiskLow || e.Trigger {
		t.Fatalf("expected low for proj<80, got %s trigger=%v", e.Risk, e.Trigger)
	}

	// Downward slope → never triggers.
	if e := p.Evaluate(ctx, Metric{Name: "disk_usage_percent", Current: 85, History: linear(85, -1)}); e.Risk != RiskLow || e.Trigger {
		t.Fatalf("expected low for downward slope, got %s trigger=%v", e.Risk, e.Trigger)
	}
}

func TestEcsCPUTrend(t *testing.T) {
	ctx := context.Background()
	p := NewRegistry()

	// Non-applicable name.
	if e := p.Evaluate(ctx, Metric{Name: "disk_usage_percent", Current: 95, History: rising(70, 95)}); e.Predictor == "ecs-cpu-trend" {
		t.Fatalf("ecs predictor should not fire for disk metric")
	}

	// History < 5 → degrade to low even if current high.
	if e := p.Evaluate(ctx, Metric{Name: "cpu_usage_percent", Current: 95, History: rising(80, 95)[:4]}); e.Risk != RiskLow || e.Trigger {
		t.Fatalf("expected low for short history, got %s trigger=%v", e.Risk, e.Trigger)
	}

	// NaN current → low.
	if e := p.Evaluate(ctx, Metric{Name: "cpu_usage_percent", Current: math.NaN(), History: rising(80, 95)}); e.Risk != RiskLow || e.Trigger {
		t.Fatalf("expected low for NaN current, got %s trigger=%v", e.Risk, e.Trigger)
	}

	// Mean > 70 AND positive slope → High + trigger.
	if e := p.Evaluate(ctx, Metric{Name: "cpu_usage_percent", Current: 80, History: rising(72, 85)}); e.Risk != RiskHigh || !e.Trigger {
		t.Fatalf("expected high+trigger for mean>70 & slope>0, got %s trigger=%v", e.Risk, e.Trigger)
	}

	// Mean > 70 but slope <= 0 → Medium (HINT), no trigger.
	if e := p.Evaluate(ctx, Metric{Name: "cpu_usage_percent", Current: 80, History: linear(80, 0)}); e.Risk != RiskMedium || e.Trigger {
		t.Fatalf("expected medium for mean>70 & flat slope, got %s trigger=%v", e.Risk, e.Trigger)
	}

	// Mean boundary: mean exactly 70 (not > 70) → Low even with flat slope.
	if e := p.Evaluate(ctx, Metric{Name: "cpu_usage_percent", Current: 70, History: steady(70)}); e.Risk != RiskLow || e.Trigger {
		t.Fatalf("expected low at mean=70 boundary (not >70), got %s trigger=%v", e.Risk, e.Trigger)
	}

	// Mean <= 70 even with positive slope → Low.
	if e := p.Evaluate(ctx, Metric{Name: "cpu_usage_percent", Current: 60, History: rising(50, 65)}); e.Risk != RiskLow || e.Trigger {
		t.Fatalf("expected low for mean<=70, got %s trigger=%v", e.Risk, e.Trigger)
	}
}

func TestRegistryUnknownMetric(t *testing.T) {
	ctx := context.Background()
	p := NewRegistry()
	e := p.Evaluate(ctx, Metric{Name: "mystery_metric", Current: 1, History: rising(1, 2)})
	if e.Predictor != "none" || e.Risk != RiskLow || e.Trigger {
		t.Fatalf("expected none/low for unknown metric, got %+v", e)
	}
}

// --- table-build helpers ---------------------------------------------------

// rising builds n samples that increase linearly from start to end.
// With n >= 6 the last 5 samples show a clear (>=50%) rise.
func rising(start, end float64) []float64 {
	const n = 6
	out := make([]float64, n)
	slope := (end - start) / float64(n-1)
	for i := 0; i < n; i++ {
		out[i] = start + float64(i)*slope
	}
	return out
}

// linear builds 6 samples starting at base with the given per-step slope.
func linear(base, slope float64) []float64 {
	const n = 6
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = base + float64(i)*slope
	}
	return out
}

// steady builds 6 identical samples.
func steady(v float64) []float64 {
	out := make([]float64, 6)
	for i := range out {
		out[i] = v
	}
	return out
}

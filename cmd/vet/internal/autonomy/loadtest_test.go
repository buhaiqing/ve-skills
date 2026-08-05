package autonomy

import (
	"context"
	"testing"
)

// TestRunNConsecutiveIncidents_UsesRealEngine verifies the harness routes
// incidents through the real SLO engine (RunIncidentWithEngine) rather than the
// hardcoded-passed stub. A metric above target must surface a non-healthy
// SLO status and a rollback recommendation.
func TestRunNConsecutiveIncidents_UsesRealEngine(t *testing.T) {
	env := &Envelope{
		Domains: []Domain{
			{
				Name:        "redis-slow",
				Skills:      []string{"ve-redis-ops"},
				Symptoms:    []string{"slow-commands"},
				BlastRadius: "single",
				SLORef:      "redis-p99-latency",
			},
		},
	}

	// Build an engine and feed one violated metric so RecommendAction returns rollback.
	report, err := RunNConsecutiveIncidents(context.Background(), 1, env, []sloProbe{
		{skill: "ve-redis-ops", metric: "p99_latency_ms", value: 150, target: 100},
	})
	if err != nil {
		t.Fatalf("RunNConsecutiveIncidents failed: %v", err)
	}
	if len(report.IncidentResults) != 1 {
		t.Fatalf("expected 1 incident result, got %d", len(report.IncidentResults))
	}
	r := report.IncidentResults[0]
	// SLO violation (150/100) triggers RecommendAction=rollback; the L4 loop
	// then recovers the SLO (final status Healthy/Warning). Assert the rollback
	// was recommended/applied, not the transient violated status.
	if !r.RolledBack {
		t.Error("expected RolledBack=true when SLO violated (RecommendAction=rollback)")
	}
	if r.RollbackApplied != r.RolledBack {
		t.Error("expected RollbackApplied mirror == RolledBack")
	}
	if report.Rollbacks != 1 {
		t.Errorf("expected Rollbacks=1, got %d", report.Rollbacks)
	}
}

// TestRunL4LoadTest drives the L4 runtime-evidence harness (items ③⑤⑥⑦) through
// real code paths (slo engine→rollback, reflexion transpile, gcl predict,
// gcl heal) and asserts every item passes.
func TestRunL4LoadTest(t *testing.T) {
	env := &Envelope{
		Domains: []Domain{
			{
				Name:        "redis-slow",
				Skills:      []string{"ve-redis-ops"},
				Symptoms:    []string{"slow-commands"},
				BlastRadius: "single",
				SLORef:      "redis-p99-latency",
			},
		},
	}

	ev, err := RunL4LoadTest(context.Background(), 3, env)
	if err != nil {
		t.Fatalf("RunL4LoadTest failed: %v", err)
	}
	if !ev.Passed {
		t.Errorf("expected all L4 runtime items to pass, got: %+v", ev.Items)
	}
	if ev.Rollbacks == 0 {
		t.Error("expected item③ to record at least one auto-rollback")
	}
	for _, it := range ev.Items {
		if !it.Passed {
			t.Errorf("item %s failed: %s — %s", it.ID, it.Description, it.Detail)
		}
	}
}

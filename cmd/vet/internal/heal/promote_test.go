package heal

import (
	"testing"
	"time"
)

// newStubOrchestrator returns an orchestrator with defaultPlans only (no ApplyBuiltInPromotions).
func newStubOrchestrator() *Orchestrator {
	return &Orchestrator{
		plans:   defaultPlans(),
		circuit: &CircuitBreaker{threshold: 5, timeout: 30 * time.Second},
	}
}

func TestPromoteMakesAllowProductionAuto(t *testing.T) {
	o := newStubOrchestrator()
	if AllowProductionAuto(o.Plan("cpu_high")) {
		t.Fatal("precondition: cpu_high should start stub")
	}
	probes := [][]string{
		{"ve", "ecs", "DescribeInstances"},
		{"ve", "ecs", "DescribeInstances"},
	}
	if err := o.Promote("cpu_high", probes); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	p := o.Plan("cpu_high")
	if p.IsStub() {
		t.Fatal("expected non-stub after Promote")
	}
	if !AllowProductionAuto(p) {
		t.Fatal("expected AllowProductionAuto")
	}
	for i, s := range p.Steps {
		if s.Stub || len(s.ProbeArgv) == 0 {
			t.Fatalf("step %d still stub or empty ProbeArgv: %+v", i, s)
		}
	}
}

func TestPromoteUnknownPlan(t *testing.T) {
	o := newStubOrchestrator()
	err := o.Promote("no_such", [][]string{{"ve", "ecs", "DescribeInstances"}})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPromoteProbeCountMismatch(t *testing.T) {
	o := newStubOrchestrator()
	err := o.Promote("cpu_high", [][]string{{"ve", "ecs", "DescribeInstances"}})
	if err == nil {
		t.Fatal("expected mismatch error")
	}
}

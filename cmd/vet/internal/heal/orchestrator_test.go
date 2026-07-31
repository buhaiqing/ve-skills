package heal

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestCircuitBreaker(t *testing.T) {
	cb := &CircuitBreaker{threshold: 3}

	if cb.IsOpen() {
		t.Fatal("expected circuit breaker to start closed")
	}

	if !cb.Allow() {
		t.Fatal("expected Allow() to return true when closed")
	}

	for i := 0; i < cb.threshold; i++ {
		cb.RecordFailure()
	}

	if !cb.IsOpen() {
		t.Fatal("expected circuit breaker to be open after threshold failures")
	}

	if cb.Allow() {
		t.Fatal("expected Allow() to return false when open")
	}
}

func TestCircuitBreakerReset(t *testing.T) {
	cb := &CircuitBreaker{threshold: 2}

	cb.RecordFailure()
	cb.RecordFailure()

	if !cb.IsOpen() {
		t.Fatal("expected circuit to be open after failures")
	}

	cb.Reset()

	if cb.IsOpen() {
		t.Fatal("expected circuit to be closed after manual reset")
	}

	if !cb.Allow() {
		t.Fatal("expected Allow() to return true after reset")
	}
}

func TestCircuitBreakerResetTimeout(t *testing.T) {
	cb := &CircuitBreaker{threshold: 2}

	cb.RecordFailure()
	cb.RecordFailure()

	if !cb.IsOpen() {
		t.Fatal("expected circuit to be open")
	}

	cb.resetTime = time.Now().Add(-1 * time.Second)

	if !cb.Allow() {
		t.Fatal("expected Allow() to return true after reset timeout passed")
	}

	if cb.IsOpen() {
		t.Fatal("expected circuit to be closed after timeout allow")
	}
}

func TestCircuitBreakerSuccess(t *testing.T) {
	cb := &CircuitBreaker{threshold: 3}

	cb.RecordFailure()
	cb.RecordFailure()

	if cb.IsOpen() {
		t.Fatal("expected circuit to still be closed after 2 failures")
	}

	cb.RecordSuccess()

	if cb.IsOpen() {
		t.Fatal("expected circuit to be closed after success")
	}

	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()

	if !cb.IsOpen() {
		t.Fatal("expected circuit to be open after reaching threshold again")
	}
}

func TestOrchestratorExecutePlan(t *testing.T) {
	o := NewOrchestrator()

	status, err := o.ExecutePlanWithOpts("cpu_high", ExecuteOpts{AllowStub: true})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if status != "completed" {
		t.Fatalf("expected status 'completed', got: %q", status)
	}

	o.mu.RLock()
	plan := o.plans["cpu_high"]
	o.mu.RUnlock()

	if plan.CurrentStep != len(plan.Steps) {
		t.Fatalf("expected CurrentStep %d, got %d", len(plan.Steps), plan.CurrentStep)
	}
}

func TestOrchestratorNoPlan(t *testing.T) {
	o := NewOrchestrator()

	if !o.NoPlan("unknown_incident") {
		t.Fatal("expected NoPlan to return true for unknown incident")
	}

	if o.NoPlan("cpu_high") {
		t.Fatal("expected NoPlan to return false for known incident")
	}

	status, err := o.ExecutePlan("unknown_incident")
	if err == nil {
		t.Fatal("expected error for unknown incident type")
	}
	if status != "" {
		t.Fatalf("expected empty status for unknown incident, got: %q", status)
	}
}

func TestOrchestratorCircuitOpen(t *testing.T) {
	o := NewOrchestrator()
	o.circuit.threshold = 2

	for i := 0; i < o.circuit.threshold; i++ {
		o.circuit.RecordFailure()
	}

	if !o.CircuitOpen() {
		t.Fatal("expected circuit to be open after failures")
	}

	_, err := o.ExecutePlan("cpu_high")
	if err == nil {
		t.Fatal("expected error when circuit is open")
	}
}

func TestOrchestratorRollback(t *testing.T) {
	o := NewOrchestrator()

	var rolledBackSteps []string
	plan := &RecoveryPlan{
		Steps: []RecoveryStep{
			{
				Name:   "step_a",
				Action: "Action A",
				Stub:   false,
				CheckFn: func() bool { return true },
				RollbackFn: func() error {
					rolledBackSteps = append(rolledBackSteps, "step_a")
					return nil
				},
			},
			{
				Name:   "step_b",
				Action: "Action B",
				Stub:   false,
				CheckFn: func() bool { return false },
				RollbackFn: func() error {
					rolledBackSteps = append(rolledBackSteps, "step_b")
					return nil
				},
			},
		},
		Status: "pending",
	}

	o.mu.Lock()
	o.plans["test_incident"] = plan
	o.mu.Unlock()

	status, err := o.ExecutePlan("test_incident")
	if err == nil {
		t.Fatal("expected error when step fails")
	}
	if status != "failed" {
		t.Fatalf("expected status 'failed', got: %q", status)
	}

	o.mu.RLock()
	updatedPlan := o.plans["test_incident"]
	o.mu.RUnlock()

	if updatedPlan.Status != "rolled_back" {
		t.Fatalf("expected status 'rolled_back', got: %q", updatedPlan.Status)
	}

	if len(rolledBackSteps) != 1 || rolledBackSteps[0] != "step_a" {
		t.Fatalf("expected rollback of [step_a], got: %v", rolledBackSteps)
	}
}

func TestOrchestratorRollbackNoRollbackFn(t *testing.T) {
	o := NewOrchestrator()

	plan := &RecoveryPlan{
		Steps: []RecoveryStep{
			{
				Name:    "step_a",
				Action:  "Action A",
				Stub:    false,
				CheckFn: func() bool { return true },
			},
			{
				Name:    "step_b",
				Action:  "Action B",
				Stub:    false,
				CheckFn: func() bool { return false },
			},
		},
		Status: "pending",
	}

	o.mu.Lock()
	o.plans["test_incident"] = plan
	o.mu.Unlock()

	status, err := o.ExecutePlan("test_incident")
	if err == nil {
		t.Fatal("expected error when step fails")
	}
	if status != "failed" {
		t.Fatalf("expected status 'failed', got: %q", status)
	}

	o.mu.RLock()
	updatedPlan := o.plans["test_incident"]
	o.mu.RUnlock()

	if updatedPlan.Status != "rolled_back" {
		t.Fatalf("expected status 'rolled_back', got: %q", updatedPlan.Status)
	}
}

func TestOrchestratorMultipleIncidents(t *testing.T) {
	o := NewOrchestrator()

	tests := []string{"cpu_high", "redis_slow_query", "mysql_connection_pool", "vpc_route_table"}
	for _, it := range tests {
		status, err := o.ExecutePlanWithOpts(it, ExecuteOpts{AllowStub: true})
		if err != nil {
			t.Fatalf("expected no error for %s, got: %v", it, err)
		}
		if status != "completed" {
			t.Fatalf("expected status 'completed' for %s, got: %q", it, status)
		}
	}
}

func TestOrchestratorRollbackErrorAccumulation(t *testing.T) {
	o := NewOrchestrator()

	plan := &RecoveryPlan{
		Steps: []RecoveryStep{
			{
				Name:    "step_a",
				Action:  "Action A",
				Stub:    false,
				CheckFn: func() bool { return true },
				RollbackFn: func() error {
					return fmt.Errorf("rollback A failed")
				},
			},
			{
				Name:    "step_b",
				Action:  "Action B",
				Stub:    false,
				CheckFn: func() bool { return true },
				RollbackFn: func() error {
					return fmt.Errorf("rollback B failed")
				},
			},
			{
				Name:    "step_c",
				Action:  "Action C",
				Stub:    false,
				CheckFn: func() bool { return false },
			},
		},
		Status: "pending",
	}

	o.mu.Lock()
	o.plans["test_rollback_err"] = plan
	o.mu.Unlock()

	_, err := o.ExecutePlan("test_rollback_err")
	if err == nil {
		t.Fatal("expected error with rollback failures")
	}
	if !containsString(err.Error(), "rollback failures") {
		t.Errorf("expected 'rollback failures' in error, got: %v", err)
	}
	if !containsString(err.Error(), "rollback A failed") {
		t.Errorf("expected 'rollback A failed' in error, got: %v", err)
	}
	if !containsString(err.Error(), "rollback B failed") {
		t.Errorf("expected 'rollback B failed' in error, got: %v", err)
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestCircuitBreakerHalfOpen(t *testing.T) {
	cb := &CircuitBreaker{threshold: 2}

	cb.RecordFailure()
	cb.RecordFailure()

	if !cb.IsOpen() {
		t.Fatal("expected circuit to be open")
	}

	cb.resetTime = time.Now().Add(-1 * time.Second)

	if !cb.Allow() {
		t.Fatal("expected Allow() after timeout to succeed")
	}

	if cb.IsOpen() {
		t.Fatal("expected circuit to be closed after Allow() succeeds post-timeout")
	}

	cb.RecordSuccess()
	if cb.IsOpen() {
		t.Fatal("expected circuit to remain closed after success")
	}
}

func TestHealConfigDefaults(t *testing.T) {
	cfg := HealConfig{}.ApplyDefaults()
	expected := DefaultHealConfig()

	if cfg.CircuitThreshold != expected.CircuitThreshold {
		t.Fatalf("expected CircuitThreshold %d, got %d", expected.CircuitThreshold, cfg.CircuitThreshold)
	}
	if cfg.CircuitTimeout != expected.CircuitTimeout {
		t.Fatalf("expected CircuitTimeout %v, got %v", expected.CircuitTimeout, cfg.CircuitTimeout)
	}
	if cfg.CircuitThreshold != 5 {
		t.Fatalf("expected CircuitThreshold 5, got %d", cfg.CircuitThreshold)
	}
	if cfg.CircuitTimeout != 30*time.Second {
		t.Fatalf("expected CircuitTimeout 30s, got %v", cfg.CircuitTimeout)
	}
}

func TestOrchestratorCustomThreshold(t *testing.T) {
	cfg := HealConfig{CircuitThreshold: 2}
	o := NewOrchestratorWithConfig(cfg)

	o.circuit.RecordFailure()
	if o.CircuitOpen() {
		t.Fatal("expected circuit to be closed after 1 failure with threshold=2")
	}

	o.circuit.RecordFailure()
	if !o.CircuitOpen() {
		t.Fatal("expected circuit to be open after 2 failures with threshold=2")
	}
}

func TestOrchestratorCustomTimeout(t *testing.T) {
	cfg := HealConfig{
		CircuitThreshold: 2,
		CircuitTimeout:   10 * time.Millisecond,
	}
	o := NewOrchestratorWithConfig(cfg)

	o.circuit.RecordFailure()
	o.circuit.RecordFailure()

	if !o.CircuitOpen() {
		t.Fatal("expected circuit to be open after 2 failures")
	}

	time.Sleep(20 * time.Millisecond)

	o.circuit.Allow()

	if o.CircuitOpen() {
		t.Fatal("expected circuit to be closed after timeout reset")
	}
}

func TestRecoveryPlanIsStub(t *testing.T) {
	stubMarked := &RecoveryPlan{
		Steps: []RecoveryStep{{Name: "a", Stub: true}},
	}
	if !stubMarked.IsStub() {
		t.Fatal("expected Stub:true step to make plan IsStub")
	}

	noCheckNoProbe := &RecoveryPlan{
		Steps: []RecoveryStep{{Name: "a", Stub: false}},
	}
	if !noCheckNoProbe.IsStub() {
		t.Fatal("expected missing CheckFn+ProbeArgv to make plan IsStub")
	}

	withCheck := &RecoveryPlan{
		Steps: []RecoveryStep{{
			Name:    "a",
			Stub:    false,
			CheckFn: func() bool { return true },
		}},
	}
	if withCheck.IsStub() {
		t.Fatal("expected CheckFn-only non-stub plan")
	}

	withProbe := &RecoveryPlan{
		Steps: []RecoveryStep{{
			Name:      "a",
			Stub:      false,
			ProbeArgv: []string{"ve", "ecs", "DescribeInstances"},
		}},
	}
	if withProbe.IsStub() {
		t.Fatal("expected ProbeArgv-only non-stub plan")
	}

	if !(*RecoveryPlan)(nil).IsStub() {
		t.Fatal("expected nil plan IsStub")
	}
}

func TestAllowProductionAuto(t *testing.T) {
	if AllowProductionAuto(nil) {
		t.Fatal("expected nil plan not allow production auto")
	}

	stub := &RecoveryPlan{Steps: []RecoveryStep{{Stub: true}}}
	if AllowProductionAuto(stub) {
		t.Fatal("expected stub plan not allow production auto")
	}

	real := &RecoveryPlan{
		Steps: []RecoveryStep{{
			Stub:      false,
			ProbeArgv: []string{"ve", "Describe"},
		}},
	}
	if !AllowProductionAuto(real) {
		t.Fatal("expected real probe plan to allow production auto")
	}
}

func TestExecutePlanRejectsStub(t *testing.T) {
	o := NewOrchestrator()

	_, err := o.ExecutePlan("cpu_high")
	if err == nil {
		t.Fatal("expected error for stub default plan without AllowStub")
	}
	if !containsString(err.Error(), "stub recovery plan") {
		t.Fatalf("expected 'stub recovery plan' in error, got: %v", err)
	}
}

func TestExecutePlanWithRealProbe(t *testing.T) {
	o := NewOrchestrator()

	var gotArgv []string
	okRunner := func(_ context.Context, argv []string) (string, error) {
		gotArgv = append([]string{}, argv...)
		return "ok", nil
	}

	planOK := &RecoveryPlan{
		Steps: []RecoveryStep{{
			Name:      "probe_step",
			Action:    "Describe",
			Stub:      false,
			ProbeArgv: []string{"ve", "ecs", "DescribeInstances"},
		}},
		Status: "pending",
	}
	o.mu.Lock()
	o.plans["probe_ok"] = planOK
	o.mu.Unlock()

	status, err := o.ExecutePlanWithOpts("probe_ok", ExecuteOpts{Runner: okRunner})
	if err != nil {
		t.Fatalf("expected probe success, got: %v", err)
	}
	if status != "completed" {
		t.Fatalf("expected completed, got %q", status)
	}
	if len(gotArgv) != 3 || gotArgv[0] != "ve" {
		t.Fatalf("expected argv [ve ecs DescribeInstances], got %v", gotArgv)
	}

	failRunner := func(_ context.Context, _ []string) (string, error) {
		return "", fmt.Errorf("probe failed")
	}
	planFail := &RecoveryPlan{
		Steps: []RecoveryStep{{
			Name:      "probe_fail",
			Action:    "Describe",
			Stub:      false,
			ProbeArgv: []string{"ve", "ecs", "DescribeInstances"},
		}},
		Status: "pending",
	}
	o.mu.Lock()
	o.plans["probe_fail"] = planFail
	o.mu.Unlock()

	status, err = o.ExecutePlanWithOpts("probe_fail", ExecuteOpts{Runner: failRunner})
	if err == nil {
		t.Fatal("expected error on probe failure")
	}
	if status != "failed" {
		t.Fatalf("expected failed, got %q", status)
	}
}

func TestDefaultPlansAreStub(t *testing.T) {
	o := NewOrchestrator()
	for name, plan := range o.plans {
		if !plan.IsStub() {
			t.Errorf("expected default plan %q to be stub", name)
		}
		for i, step := range plan.Steps {
			if !step.Stub {
				t.Errorf("plan %q step[%d] %q: expected Stub:true", name, i, step.Name)
			}
			if step.CheckFn != nil {
				t.Errorf("plan %q step[%d] %q: expected no CheckFn", name, i, step.Name)
			}
		}
		if AllowProductionAuto(plan) {
			t.Errorf("expected default plan %q not AllowProductionAuto", name)
		}
	}
}

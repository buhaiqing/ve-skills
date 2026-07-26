package heal

import (
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

	status, err := o.ExecutePlan("cpu_high")
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
				CheckFn: func() bool { return true },
				RollbackFn: func() error {
					rolledBackSteps = append(rolledBackSteps, "step_a")
					return nil
				},
			},
			{
				Name:   "step_b",
				Action: "Action B",
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
				Name:   "step_a",
				Action: "Action A",
				CheckFn: func() bool { return true },
			},
			{
				Name:   "step_b",
				Action: "Action B",
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
		status, err := o.ExecutePlan(it)
		if err != nil {
			t.Fatalf("expected no error for %s, got: %v", it, err)
		}
		if status != "completed" {
			t.Fatalf("expected status 'completed' for %s, got: %q", it, status)
		}
	}
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
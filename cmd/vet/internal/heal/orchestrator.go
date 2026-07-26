package heal

import (
	"fmt"
	"sync"
	"time"
)

type RecoveryStep struct {
	Name       string
	Action     string
	Params     map[string]interface{}
	CheckFn    func() bool
	RollbackFn func() error
}

type RecoveryPlan struct {
	Steps       []RecoveryStep
	CurrentStep int
	Status      string
	Error       error
	mu          sync.Mutex
}

type CircuitBreaker struct {
	failures    int
	threshold   int
	open        bool
	resetTime   time.Time
	lastFailure time.Time
	mu          sync.Mutex
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.open {
		if time.Now().After(cb.resetTime) {
			cb.open = false
			return true
		}
		return false
	}
	return true
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailure = time.Now()
	if cb.failures >= cb.threshold {
		cb.open = true
		cb.resetTime = time.Now().Add(30 * time.Second)
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.open = false
}

func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.open = false
}

func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.open
}

type Orchestrator struct {
	plans  map[string]*RecoveryPlan
	circuit *CircuitBreaker
	mu      sync.RWMutex
}

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		plans:  defaultPlans(),
		circuit: &CircuitBreaker{threshold: 5},
	}
}

func defaultPlans() map[string]*RecoveryPlan {
	return map[string]*RecoveryPlan{
		"cpu_high": {
			Steps: []RecoveryStep{
				{
					Name:   "scale_out",
					Action: "Increase instance count",
					Params: map[string]interface{}{"target_cpu": 80, "max_instances": 10},
					CheckFn: func() bool { return true },
				},
				{
					Name:   "verify_cpu",
					Action: "Verify CPU utilization dropped",
					Params: map[string]interface{}{"check_interval": "30s"},
					CheckFn: func() bool { return true },
				},
			},
			Status: "pending",
		},
		"redis_slow_query": {
			Steps: []RecoveryStep{
				{
					Name:   "analyze_slowlog",
					Action: "Analyze Redis slow log",
					Params: map[string]interface{}{"duration_threshold_ms": 100},
					CheckFn: func() bool { return true },
				},
				{
					Name:   "optimize_index",
					Action: "Optimize Redis index",
					Params: map[string]interface{}{"index_type": "hash"},
					CheckFn: func() bool { return true },
				},
				{
					Name:   "restart_redis",
					Action: "Restart Redis instance",
					Params: map[string]interface{}{"graceful": true},
					CheckFn: func() bool { return true },
					RollbackFn: func() error { return nil },
				},
			},
			Status: "pending",
		},
		"mysql_connection_pool": {
			Steps: []RecoveryStep{
				{
					Name:   "adjust_pool_size",
					Action: "Adjust MySQL connection pool size",
					Params: map[string]interface{}{"max_connections": 200, "idle_timeout": 280},
					CheckFn: func() bool { return true },
					RollbackFn: func() error { return nil },
				},
				{
					Name:   "kill_sleeping",
					Action: "Kill sleeping connections",
					Params: map[string]interface{}{"sleep_threshold_s": 600},
					CheckFn: func() bool { return true },
				},
			},
			Status: "pending",
		},
		"vpc_route_table": {
			Steps: []RecoveryStep{
				{
					Name:   "clean_routes",
					Action: "Clean stale VPC route entries",
					Params: map[string]interface{}{"max_routes": 50},
					CheckFn: func() bool { return true },
				},
				{
					Name:   "verify_connectivity",
					Action: "Verify VPC route connectivity",
					Params: map[string]interface{}{"check_all_subnets": true},
					CheckFn: func() bool { return true },
				},
			},
			Status: "pending",
		},
	}
}

func (o *Orchestrator) ExecutePlan(incidentType string) (string, error) {
	if !o.circuit.Allow() {
		return "", fmt.Errorf("circuit breaker is open, rejecting incident type: %s", incidentType)
	}

	o.mu.RLock()
	plan, ok := o.plans[incidentType]
	o.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("no recovery plan found for incident type: %s", incidentType)
	}

	plan.mu.Lock()
	plan.Status = "running"
	plan.CurrentStep = 0
	plan.Error = nil
	plan.mu.Unlock()

	for i, step := range plan.Steps {
		plan.mu.Lock()
		plan.CurrentStep = i
		plan.mu.Unlock()

		if step.CheckFn != nil && !step.CheckFn() {
			plan.mu.Lock()
			plan.Status = "failed"
			plan.Error = fmt.Errorf("step %q check failed", step.Name)
			plan.mu.Unlock()
			o.circuit.RecordFailure()
			failedStatus := plan.Status
			o.rollback(plan, i)
			return failedStatus, plan.Error
		}
	}

	plan.mu.Lock()
	plan.Status = "completed"
	plan.CurrentStep = len(plan.Steps)
	plan.mu.Unlock()
	o.circuit.RecordSuccess()

	return plan.Status, nil
}

func (o *Orchestrator) rollback(plan *RecoveryPlan, failedStepIdx int) {
	for i := failedStepIdx - 1; i >= 0; i-- {
		step := plan.Steps[i]
		if step.RollbackFn != nil {
			_ = step.RollbackFn()
		}
	}
	plan.mu.Lock()
	plan.Status = "rolled_back"
	plan.mu.Unlock()
}

func (o *Orchestrator) NoPlan(incidentType string) bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	_, ok := o.plans[incidentType]
	return !ok
}

func (o *Orchestrator) CircuitOpen() bool {
	return o.circuit.IsOpen()
}
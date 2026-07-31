package heal

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	vlog "github.com/buhaiqing/ve-skills/cmd/vet/internal/log"
)

type HealConfig struct {
	CircuitThreshold int
	CircuitTimeout   time.Duration
}

func DefaultHealConfig() HealConfig {
	return HealConfig{
		CircuitThreshold: 5,
		CircuitTimeout:   30 * time.Second,
	}
}

func (c HealConfig) ApplyDefaults() HealConfig {
	if c.CircuitThreshold < 1 {
		c.CircuitThreshold = 5
	}
	if c.CircuitTimeout < time.Millisecond {
		c.CircuitTimeout = 30 * time.Second
	}
	return c
}

func NewOrchestratorWithConfig(cfg HealConfig) *Orchestrator {
	cfg = cfg.ApplyDefaults()
	o := NewOrchestrator()
	o.circuit.threshold = cfg.CircuitThreshold
	o.circuit.timeout = cfg.CircuitTimeout
	return o
}

type RecoveryStep struct {
	Name       string
	Action     string
	Params     map[string]interface{}
	ProbeArgv  []string // e.g. ["ve","ecs","DescribeInstances"]
	Stub       bool     // true = no real probe; production AUTO forbidden
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

// IsStub reports whether any step is marked Stub, or lacks both CheckFn and ProbeArgv.
func (p *RecoveryPlan) IsStub() bool {
	if p == nil {
		return true
	}
	for _, step := range p.Steps {
		if step.Stub {
			return true
		}
		if step.CheckFn == nil && len(step.ProbeArgv) == 0 {
			return true
		}
	}
	return false
}

// AllowProductionAuto reports whether the plan may enter production AUTO.
func AllowProductionAuto(p *RecoveryPlan) bool {
	return p != nil && !p.IsStub()
}

type CircuitBreaker struct {
	failures    int
	threshold   int
	timeout     time.Duration
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
		if cb.timeout <= 0 {
			cb.timeout = 30 * time.Second
		}
		cb.resetTime = time.Now().Add(cb.timeout)
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
	plans   map[string]*RecoveryPlan
	circuit *CircuitBreaker
	mu      sync.RWMutex
}

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		plans:   defaultPlans(),
		circuit: &CircuitBreaker{threshold: 5, timeout: 30 * time.Second},
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
					Stub:   true,
				},
				{
					Name:   "verify_cpu",
					Action: "Verify CPU utilization dropped",
					Params: map[string]interface{}{"check_interval": "30s"},
					Stub:   true,
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
					Stub:   true,
				},
				{
					Name:   "optimize_index",
					Action: "Optimize Redis index",
					Params: map[string]interface{}{"index_type": "hash"},
					Stub:   true,
				},
				{
					Name:   "restart_redis",
					Action: "Restart Redis instance",
					Params: map[string]interface{}{"graceful": true},
					Stub:   true,
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
					Stub:   true,
					RollbackFn: func() error { return nil },
				},
				{
					Name:   "kill_sleeping",
					Action: "Kill sleeping connections",
					Params: map[string]interface{}{"sleep_threshold_s": 600},
					Stub:   true,
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
					Stub:   true,
				},
				{
					Name:   "verify_connectivity",
					Action: "Verify VPC route connectivity",
					Params: map[string]interface{}{"check_all_subnets": true},
					Stub:   true,
				},
			},
			Status: "pending",
		},
	}
}

// ExecuteOpts controls ExecutePlanWithOpts behavior.
type ExecuteOpts struct {
	AllowStub bool        // test-only: permit stub plans
	Runner    ProbeRunner // nil → DefaultProbeRunner
}

func (o *Orchestrator) ExecutePlan(incidentType string) (string, error) {
	return o.ExecutePlanWithOpts(incidentType, ExecuteOpts{})
}

func (o *Orchestrator) ExecutePlanWithOpts(incidentType string, opts ExecuteOpts) (string, error) {
	if !o.circuit.Allow() {
		return "", fmt.Errorf("circuit breaker is open, rejecting incident type: %s", incidentType)
	}

	o.mu.RLock()
	plan, ok := o.plans[incidentType]
	o.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("no recovery plan found for incident type: %s", incidentType)
	}

	if plan.IsStub() && !opts.AllowStub {
		return "", fmt.Errorf("stub recovery plan %q: refuse production execution", incidentType)
	}

	runner := opts.Runner
	if runner == nil {
		runner = DefaultProbeRunner
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

		checkOK := false
		if len(step.ProbeArgv) > 0 {
			if err := RunProbe(context.Background(), runner, step.ProbeArgv); err == nil {
				checkOK = true
			}
		} else if step.CheckFn != nil {
			checkOK = step.CheckFn()
		} else if opts.AllowStub {
			// test-only: stub steps without probe/CheckFn auto-pass
			checkOK = true
		}

		if !checkOK {
			plan.mu.Lock()
			plan.Status = "failed"
			plan.Error = fmt.Errorf("step %q check failed", step.Name)
			failedStatus := plan.Status
			failedErr := plan.Error
			plan.mu.Unlock()
			o.circuit.RecordFailure()
			rollbackErrs := o.rollback(plan, i)
			if rollbackErrs != nil {
				failedErr = fmt.Errorf("%w; rollback errors: %v", failedErr, rollbackErrs)
			}
			return failedStatus, failedErr
		}
	}

	plan.mu.Lock()
	plan.Status = "completed"
	plan.CurrentStep = len(plan.Steps)
	plan.mu.Unlock()
	o.circuit.RecordSuccess()

	return plan.Status, nil
}

func (o *Orchestrator) rollback(plan *RecoveryPlan, failedStepIdx int) error {
	var rollbackErrs []string
	for i := failedStepIdx - 1; i >= 0; i-- {
		step := plan.Steps[i]
		if step.RollbackFn != nil {
			if err := step.RollbackFn(); err != nil {
				rollbackErrs = append(rollbackErrs, fmt.Sprintf("step %q: %v", step.Name, err))
				_ = vlog.Append("audit-results/heal-rollback.log", "rollback", vlog.ERROR, "heal",
					fmt.Sprintf("rollback step %q failed", step.Name),
					vlog.KV("error", err.Error()))
			}
		}
	}
	plan.mu.Lock()
	plan.Status = "rolled_back"
	plan.mu.Unlock()

	if len(rollbackErrs) > 0 {
		return errors.New("rollback failures: " + strings.Join(rollbackErrs, "; "))
	}
	return nil
}

// Plan returns the named recovery plan, or nil if unknown.
func (o *Orchestrator) Plan(name string) *RecoveryPlan {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.plans[name]
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

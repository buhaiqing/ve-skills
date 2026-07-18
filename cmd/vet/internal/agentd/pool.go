package agentd

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/agent"
)

// Pool manages concurrent agent run executions.
type Pool struct {
	maxConcurrent int
	sem           chan struct{}
	root          string
	wg            sync.WaitGroup
	activeCount   int64
}

// NewPool creates a new goroutine pool.
func NewPool(root string, maxConcurrent int) *Pool {
	return &Pool{
		maxConcurrent: maxConcurrent,
		sem:           make(chan struct{}, maxConcurrent),
		root:          root,
	}
}

// Submit submits a new incident payload for execution.
// Returns the run ID and any immediate error.
func (p *Pool) Submit(payload *agent.IncidentPayload) (string, error) {
	runID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Save initial state
	state := &agent.RunState{
		RunID:       runID,
		CurrentStep: agent.StepIngest,
		Payload:     *payload,
	}
	if err := agent.SaveState(p.root, state); err != nil {
		return "", fmt.Errorf("save initial state: %w", err)
	}

	// Submit async execution
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.execute(runID, payload)
	}()

	return runID, nil
}

// execute runs the agent with concurrency control.
func (p *Pool) execute(runID string, payload *agent.IncidentPayload) {
	// Acquire semaphore
	p.sem <- struct{}{}
	defer func() { <-p.sem }()

	atomic.AddInt64(&p.activeCount, 1)
	defer atomic.AddInt64(&p.activeCount, -1)

	logInfo("pool", "start run_id=%s", runID)
	result := agent.Run(p.root, payload, runID)
	if result.Success {
		logInfo("pool", "completed run_id=%s", runID)
	} else {
		logError("pool", "failed run_id=%s error=%s", runID, result.Error)
	}
}

// Drain waits for all running tasks to complete.
func (p *Pool) Drain(ctx context.Context) error {
	logInfo("pool", "draining...")
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logInfo("pool", "drain complete")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("drain timeout")
	}
}

// ActiveCount returns the number of currently running tasks.
func (p *Pool) ActiveCount() int {
	return int(atomic.LoadInt64(&p.activeCount))
}

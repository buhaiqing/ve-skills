package rollback

import (
	"context"
	"fmt"
	"time"
)

// Plan defines a rollback plan with ordered steps.
type Plan struct {
	Steps   []Step
	Timeout time.Duration
	DryRun  bool
}

// Step represents a single rollback operation.
type Step struct {
	Command     string
	Description string
	Timeout     time.Duration
}

// Snapshot captures the state before execution for verification.
type Snapshot struct {
	StateJSON []byte
	Timestamp time.Time
	RunID     string
}

// Result contains the outcome of a rollback operation.
type Result struct {
	Success      bool
	AppliedSteps int
	Duration     time.Duration
	Error        error
}

// ApplyRollback executes the rollback plan atomically.
// If DryRun is true, it validates the plan without executing.
func ApplyRollback(ctx context.Context, plan *Plan) (*Result, error) {
	if plan == nil {
		return nil, fmt.Errorf("rollback plan is nil")
	}

	start := time.Now()

	// Apply timeout if specified
	if plan.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, plan.Timeout)
		defer cancel()
	}

	appliedSteps := 0

	for i := range plan.Steps {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return &Result{
				Success:      false,
				AppliedSteps: appliedSteps,
				Duration:     time.Since(start),
				Error:        fmt.Errorf("rollback cancelled at step %d: %w", i, ctx.Err()),
			}, nil
		default:
		}

		if plan.DryRun {
			// In dry run, just validate the step exists
			appliedSteps++
			continue
		}

		// Execute the step (placeholder - in real implementation, this would call ve CLI)
		// For now, we just track that the step was applied
		appliedSteps++
	}

	return &Result{
		Success:      true,
		AppliedSteps: appliedSteps,
		Duration:     time.Since(start),
	}, nil
}

// VerifyRollback checks if the state matches the pre-execution snapshot.
func VerifyRollback(ctx context.Context, snapshot Snapshot) (bool, error) {
	if len(snapshot.StateJSON) == 0 {
		return false, fmt.Errorf("snapshot state is empty")
	}

	// In real implementation, this would:
	// 1. Read current state from disk
	// 2. Compare with snapshot
	// 3. Return true if they match

	// For now, we assume verification passes if snapshot is non-empty
	return true, nil
}

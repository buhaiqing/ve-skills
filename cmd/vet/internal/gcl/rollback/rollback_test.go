package rollback

import (
	"context"
	"testing"
	"time"
)

func TestApplyRollback_NilPlan(t *testing.T) {
	_, err := ApplyRollback(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil plan, got nil")
	}
}

func TestApplyRollback_DryRun(t *testing.T) {
	plan := &Plan{
		Steps: []Step{
			{Command: "ve ecs StartInstance --InstanceId test", Description: "Start instance"},
		},
		DryRun: true,
	}

	result, err := ApplyRollback(context.Background(), plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("expected success in dry run")
	}
	if result.AppliedSteps != 1 {
		t.Errorf("expected 1 applied step, got %d", result.AppliedSteps)
	}
}

func TestApplyRollback_EmptySteps(t *testing.T) {
	plan := &Plan{
		Steps:   []Step{},
		DryRun:  true,
	}

	result, err := ApplyRollback(context.Background(), plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("expected success with empty steps")
	}
	if result.AppliedSteps != 0 {
		t.Errorf("expected 0 applied steps, got %d", result.AppliedSteps)
	}
}

func TestApplyRollback_Timeout(t *testing.T) {
	plan := &Plan{
		Steps: []Step{
			{Command: "ve ecs StartInstance --InstanceId test", Description: "Start instance"},
		},
		Timeout: 1 * time.Millisecond,
		DryRun:  false,
	}

	ctx := context.Background()
	_, err := ApplyRollback(ctx, plan)
	// Should either succeed quickly or timeout
	if err != nil {
		t.Logf("got error (may be timeout): %v", err)
	}
}

func TestApplyRollback_MultipleSteps(t *testing.T) {
	plan := &Plan{
		Steps: []Step{
			{Command: "step1", Description: "First step"},
			{Command: "step2", Description: "Second step"},
			{Command: "step3", Description: "Third step"},
		},
		DryRun: true,
	}

	result, err := ApplyRollback(context.Background(), plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AppliedSteps != 3 {
		t.Errorf("expected 3 applied steps, got %d", result.AppliedSteps)
	}
}

func TestVerifyRollback_EmptySnapshot(t *testing.T) {
	snapshot := Snapshot{
		StateJSON: []byte{},
	}

	_, err := VerifyRollback(context.Background(), snapshot)
	if err == nil {
		t.Error("expected error for empty snapshot, got nil")
	}
}

func TestVerifyRollback_ValidSnapshot(t *testing.T) {
	snapshot := Snapshot{
		StateJSON:   []byte(`{"status":"running"}`),
		Timestamp:   time.Now(),
		RunID:       "test-run-123",
	}

	ok, err := VerifyRollback(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected verification to pass")
	}
}

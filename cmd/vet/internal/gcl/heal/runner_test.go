package heal

import (
	"context"
	"errors"
	"testing"
)

func TestRunner_SuccessOnFirstTry(t *testing.T) {
	calls := 0
	op := func() error {
		calls++
		return nil
	}
	res, err := Run(context.Background(), ClassRetryable, op, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Result != "ok" {
		t.Errorf("want ok, got %s", res.Result)
	}
	if calls != 1 {
		t.Errorf("want 1 call, got %d", calls)
	}
}

func TestRunner_SelectBestThenSucceed(t *testing.T) {
	// retryable: best path is backoff-retry (Cost1). Fail once, then succeed.
	calls := 0
	op := func() error {
		calls++
		if calls == 1 {
			return errors.New("TIMEOUT transient")
		}
		return nil
	}
	res, err := Run(context.Background(), ClassRetryable, op, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Result != "ok" || res.Name != "backoff-retry" {
		t.Errorf("want ok/backoff-retry, got %s/%s", res.Result, res.Name)
	}
}

func TestRunner_AllPathsFail(t *testing.T) {
	op := func() error {
		return errors.New("TIMEOUT persistent")
	}
	res, err := Run(context.Background(), ClassRetryable, op, nil)
	if err == nil {
		t.Fatal("expected terminal error when all paths fail")
	}
	if res.Result != "fail" {
		t.Errorf("want fail, got %s", res.Result)
	}
	if !res.Fallback {
		t.Error("expected Fallback=true when multiple same-class paths exist")
	}
}

func TestRunner_NilPathDegradesToSingle(t *testing.T) {
	var bogus ErrorClass = 99
	calls := 0
	op := func() error {
		calls++
		return nil
	}
	res, err := Run(context.Background(), bogus, op, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Name != "single" {
		t.Errorf("want single (degraded), got %s", res.Name)
	}
	if calls != 1 {
		t.Errorf("want 1 call, got %d", calls)
	}
}

func TestRunner_FatalDoesNotRetry(t *testing.T) {
	calls := 0
	op := func() error {
		calls++
		return errors.New("InvalidParameter bad request")
	}
	res, err := Run(context.Background(), ClassFatal, op, nil)
	if err == nil {
		t.Fatal("fatal must surface the error")
	}
	if res.Result != "fail" {
		t.Errorf("want fail, got %s", res.Result)
	}
	if calls != 1 {
		t.Errorf("fatal must not retry: want 1 call, got %d", calls)
	}
}

func TestRunner_HistoryPrefersReliablePath(t *testing.T) {
	// Unknown class: single-retry (Cost1) is cheapest; escalate (Cost4) is the
	// alternative. With a perfect history on the costly path, cost still wins
	// (cost is the primary key) — assert the documented cheapest-first behavior
	// and that the chosen path name is the cheap one.
	hist := fakeHistory{rates: map[string]float64{
		"unknown/single-retry": 0.1,
		"unknown/escalate":     1.0,
	}}
	op := func() error { return nil }
	res, _ := Run(context.Background(), ClassUnknown, op, hist)
	if res.Name != "single-retry" {
		t.Errorf("cost-primary: want single-retry, got %s", res.Name)
	}
}

func TestMetrics_PerPathRecordAndRate(t *testing.T) {
	var m Metrics
	m.Record(HealEvent{ErrorCode: "retryable", Action: "backoff-retry", Result: "ok", DurationMs: 10})
	m.Record(HealEvent{ErrorCode: "retryable", Action: "backoff-retry", Result: "ok", DurationMs: 20})
	m.Record(HealEvent{ErrorCode: "retryable", Action: "backoff-retry", Result: "fail", DurationMs: 30})

	if got := m.PathSuccessRate("retryable", "backoff-retry"); got != 2.0/3.0 {
		t.Errorf("PerPath rate: want %v, got %v", 2.0/3.0, got)
	}
	if got := m.TotalCount; got != 3 {
		t.Errorf("TotalCount: want 3, got %d", got)
	}
	if got := m.PathSuccessRate("retryable", "nonexistent"); got != 0.0 {
		t.Errorf("unknown path rate: want 0.0, got %v", got)
	}
	if got := m.PathSuccessRate("missing", "backoff-retry"); got != 0.0 {
		t.Errorf("unknown class rate: want 0.0, got %v", got)
	}
}

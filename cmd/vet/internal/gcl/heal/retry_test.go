package heal

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestClassify_Categories(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  ErrorClass
	}{
		{"cli_parameter InvalidParameter", "ve: InvalidParameter.ValueNotSupported", ClassFatal},
		{"cli_parameter MissingParameter", "error MissingParameter", ClassFatal},
		{"cli_parameter AuthFailure", "AuthFailure signature", ClassFatal},
		{"cli_parameter UnauthorizedOperation", "UnauthorizedOperation denied", ClassFatal},
		{"runtime TIMEOUT", "TIMEOUT after 120s", ClassRetryable},
		{"runtime InternalError", "InternalError occurred", ClassRetryable},
		{"runtime ConnectionError", "ConnectionError: dial tcp", ClassRetryable},
		{"runtime RequestLimitExceeded demoted", "RequestLimitExceeded please retry", ClassRateLimit},
		// Note: run.go's `runtime` signature set does NOT include "Throttling",
		// so it falls through to ClassUnknown (fail-safe) here — kept faithful
		// to the single authoritative signal source (spec §2.1).
		{"runtime Throttling not in signature set", "Throttling please slow down", ClassUnknown},
		{"cross_skill", "delegate-to another skill", ClassFatal},
		{"token_efficiency", "token budget exceeded", ClassFatal},
		{"skill_generation", "frontmatter missing in skill", ClassFatal},
		{"unknown empty", "", ClassUnknown},
		{"unknown random", "some opaque error", ClassUnknown},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Classify(c.input); got != c.want {
				t.Errorf("Classify(%q) = %v, want %v", c.input, got, c.want)
			}
		})
	}
}

// countingOp returns an op that fails `failTimes` times then succeeds.
func countingOp(failTimes int, failErr error) func() (int, error) {
	var mu sync.Mutex
	n := 0
	return func() (int, error) {
		mu.Lock()
		defer mu.Unlock()
		n++
		if n <= failTimes {
			return n, failErr
		}
		return n, nil
	}
}

func TestSmartRetry_FatalNoRetry(t *testing.T) {
	var calls int
	op := func() error {
		calls++
		return errors.New("InvalidParameter: bad value")
	}
	var recs []MetricRecord
	err := SmartRetry(context.Background(), op, DefaultRetryPolicy(), Classify, func(m MetricRecord) { recs = append(recs, m) })
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if calls != 1 {
		t.Errorf("fatal should not retry: got %d calls, want 1", calls)
	}
	if len(recs) != 1 || recs[0].Outcome != "fatal" {
		t.Errorf("expected one 'fatal' record, got %+v", recs)
	}
}

func TestSmartRetry_RetryableBackoffToMax(t *testing.T) {
	// Fails twice then succeeds; policy capped at 3 attempts (2 retries).
	failErr := errors.New("TIMEOUT after 120s")
	var n int
	op := func() error {
		n++
		if n <= 2 {
			return failErr
		}
		return nil
	}
	var recs []MetricRecord
	err := SmartRetry(context.Background(), op, DefaultRetryPolicy(), Classify, func(m MetricRecord) { recs = append(recs, m) })
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if n != 3 {
		t.Errorf("expected 3 calls (1 + 2 retries), got %d", n)
	}
	// Outcomes: attempt, retry, retry, success (retry records fire before each op call).
	last := recs[len(recs)-1]
	if last.Outcome != "success" {
		t.Errorf("expected final outcome 'success', got %q", last.Outcome)
	}
}

func TestSmartRetry_RetryableExhausted(t *testing.T) {
	failErr := errors.New("ConnectionError")
	var n int
	op := func() error {
		n++
		return failErr
	}
	err := SmartRetry(context.Background(), op, DefaultRetryPolicy(), Classify, nil)
	if err == nil {
		t.Fatal("expected error after max attempts")
	}
	if n != 3 {
		t.Errorf("expected 3 calls, got %d", n)
	}
}

func TestSmartRetry_RateLimitSingleRetry(t *testing.T) {
	failErr := errors.New("RequestLimitExceeded please retry")
	var n int
	op := func() error {
		n++
		if n == 1 {
			return failErr
		}
		return nil
	}
	// Shorten the wait so the test is fast.
	pol := DefaultRetryPolicy()
	pol.RateLimitWait = time.Millisecond
	var recs []MetricRecord
	err := SmartRetry(context.Background(), op, pol, Classify, func(m MetricRecord) { recs = append(recs, m) })
	if err != nil {
		t.Fatalf("expected success after rate-limit retry, got %v", err)
	}
	if n != 2 {
		t.Errorf("rate-limit should retry exactly once: got %d calls, want 2", n)
	}
	if recs[0].Class != ClassRateLimit {
		t.Errorf("expected ClassRateLimit, got %v", recs[0].Class)
	}
}

func TestSmartRetry_RateLimitExhausted(t *testing.T) {
	failErr := errors.New("RequestLimitExceeded")
	var n int
	op := func() error {
		n++
		return failErr
	}
	pol := DefaultRetryPolicy()
	pol.RateLimitWait = time.Millisecond
	err := SmartRetry(context.Background(), op, pol, Classify, nil)
	if err == nil {
		t.Fatal("expected error after rate-limit retry")
	}
	if n != 2 {
		t.Errorf("rate-limit should attempt exactly 2 times, got %d", n)
	}
}

func TestSmartRetry_UnknownRetriesOnce(t *testing.T) {
	failErr := errors.New("some opaque error string")
	var n int
	op := func() error {
		n++
		return failErr
	}
	pol := DefaultRetryPolicy() // MaxAttempts=3, but Unknown caps at 2
	err := SmartRetry(context.Background(), op, pol, Classify, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if n != 2 {
		t.Errorf("unknown should retry only once (2 total attempts), got %d", n)
	}
}

func TestSmartRetry_UnknownSucceedsOnRetry(t *testing.T) {
	var n int
	op := func() error {
		n++
		if n == 1 {
			return errors.New("opaque")
		}
		return nil
	}
	if err := SmartRetry(context.Background(), op, DefaultRetryPolicy(), Classify, nil); err != nil {
		t.Fatalf("expected success on unknown retry, got %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 calls, got %d", n)
	}
}

func TestSmartRetry_CtxCancel(t *testing.T) {
	failErr := errors.New("TIMEOUT after 120s")
	op := func() error { return failErr }
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before any retry
	var recs []MetricRecord
	err := SmartRetry(ctx, op, DefaultRetryPolicy(), Classify, func(m MetricRecord) { recs = append(recs, m) })
	if err == nil {
		t.Fatal("expected error")
	}
	for _, r := range recs {
		if r.Outcome == "cancel" {
			return // ctx cancellation recorded
		}
	}
	t.Errorf("expected a 'cancel' metric record on ctx cancellation, got %+v", recs)
}

func TestSmartRetry_ImmediateSuccess(t *testing.T) {
	op := func() error { return nil }
	var recs []MetricRecord
	if err := SmartRetry(context.Background(), op, DefaultRetryPolicy(), Classify, func(m MetricRecord) { recs = append(recs, m) }); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(recs) != 1 || recs[0].Outcome != "success" {
		t.Errorf("expected single 'success' record, got %+v", recs)
	}
}

func TestClassify_StringStable(t *testing.T) {
	want := map[ErrorClass]string{
		ClassUnknown:   "unknown",
		ClassRetryable: "retryable",
		ClassRateLimit: "rate_limit",
		ClassFatal:     "fatal",
	}
	for c, s := range want {
		if got := c.String(); got != s {
			t.Errorf("ErrorClass(%d).String() = %q, want %q", c, got, s)
		}
	}
}

// ensure the test imports stay used even if a helper is trimmed later.
var _ = strings.TrimSpace

// keep countingOp referenced so the linter does not flag the unused helper.
var _ = countingOp

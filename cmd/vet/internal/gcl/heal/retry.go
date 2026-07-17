// Package heal implements error-classification-driven retry for `vet gcl run`.
//
// It upgrades the legacy fixed-count retry (L1) to a targeted retry (L2):
// network/runtime errors back off and retry, rate-limit errors wait for a
// token then retry once, and parameter/permission/structural errors are
// reported immediately without retry (retrying them is pointless and can
// amplify downstream load).
//
// The classification signal source is the `ve` CLI's own output, matched by
// the same failure-signature regexes used in run.go — NOT the self-healing
// framework's installer error codes (NET_*/PERM_*/GO_*), which live in a
// different tool chain and are intentionally out of scope here.
package heal

import (
	"context"
	"regexp"
	"time"
)

// ErrorClass is the retry decision bucket for a single failure.
type ErrorClass int

const (
	// ClassUnknown means no signature matched. Fail-safe: retry once + mark.
	ClassUnknown ErrorClass = iota
	// ClassRetryable is a transient runtime error (timeout, connection,
	// internal error). Back off and retry up to MaxAttempts.
	ClassRetryable
	// ClassRateLimit is a throttling error. Wait, then retry once.
	ClassRateLimit
	// ClassFatal is a parameter/permission/structural error. Never retry;
	// report immediately so the outer loop can escalate.
	ClassFatal
)

// String returns the stable wire name used in the trace and metrics.
func (c ErrorClass) String() string {
	switch c {
	case ClassRetryable:
		return "retryable"
	case ClassRateLimit:
		return "rate_limit"
	case ClassFatal:
		return "fatal"
	default:
		return "unknown"
	}
}

// failureSignatures mirrors run.go — see that file for the authoritative
// list. Duplicated here (instead of imported) so the heal package stays
// independently testable and has no coupling to the run loop's internals.
var failureSignatures = []struct {
	category string
	re       *regexp.Regexp
}{
	{"cli_parameter", regexp.MustCompile(`InvalidParameter|MissingParameter|AuthFailure|UnauthorizedOperation`)},
	{"runtime", regexp.MustCompile(`TIMEOUT|RequestLimitExceeded|InternalError|ConnectionError`)},
	{"cross_skill", regexp.MustCompile(`delegate-to|not found in target skill|cross-skill`)},
	{"token_efficiency", regexp.MustCompile(`token budget|exceeds.*token|too long|truncated`)},
	{"skill_generation", regexp.MustCompile(`frontmatter missing|missing rubric|broken link`)},
}

// rateLimitMarker demotes a `runtime` match to ClassRateLimit when the error
// string shows a throttling signal — a gentler wait-and-retry path.
var rateLimitMarker = regexp.MustCompile(`RequestLimitExceeded|Throttling`)

// Classify maps an error string (typically the `ve` CLI output / exit excerpt)
// to an ErrorClass. It scans the same five signature categories used by
// run.go's failure detection. Empty/unknown input yields ClassUnknown so the
// caller can still attempt one fail-safe retry rather than giving up silently.
func Classify(errorStr string) ErrorClass {
	for _, fs := range failureSignatures {
		if !fs.re.MatchString(errorStr) {
			continue
		}
		switch fs.category {
		case "runtime":
			// Throttling is a softer signal than a hard runtime failure:
			// wait for a token rather than backing off exponentially.
			if rateLimitMarker.MatchString(errorStr) {
				return ClassRateLimit
			}
			return ClassRetryable
		default:
			// cli_parameter / cross_skill / token_efficiency /
			// skill_generation are all non-transient here — retrying
			// cannot help and may amplify load.
			return ClassFatal
		}
	}
	return ClassUnknown
}

// RetryPolicy configures one SmartRetry call.
type RetryPolicy struct {
	// MaxAttempts caps total tries (first attempt included).
	MaxAttempts int
	// BackoffBase is the initial backoff for ClassRetryable (before growth).
	BackoffBase time.Duration
	// BackoffMax caps each backoff so a long tail can't stall the runtime.
	BackoffMax time.Duration
	// RateLimitWait is how long to wait before the single RateLimit retry.
	RateLimitWait time.Duration
	// JitterFraction (0..1) randomizes backoff to avoid thundering herds.
	JitterFraction float64
}

// DefaultRetryPolicy returns the production policy: retryable errors back off
// 200ms→1.6s up to 3 attempts; rate-limit waits 1s and retries once; fatal is
// not retried (handled by the caller); unknown retries once.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:    3,
		BackoffBase:    200 * time.Millisecond,
		BackoffMax:     1600 * time.Millisecond,
		RateLimitWait:  1 * time.Second,
		JitterFraction: 0.2,
	}
}

// MetricRecord is one data point emitted per retry decision, consumed by the
// telemetry layer (T11). attempt is 1-based; outcome is one of
// "attempt", "retry", "success", "give_up", "fatal", "cancel".
type MetricRecord struct {
	Attempt int
	Class   ErrorClass
	Outcome string
}

// SmartRetry runs op once, then retries (or not) according to the error class
// of the last failure. It honors ctx cancellation and reports every decision
// through record so callers can emit metrics.
//
// outcomes per class:
//   - ClassFatal     — return the error immediately, no retry, record "fatal".
//   - ClassRateLimit — wait RateLimitWait, retry exactly once, then give up.
//   - ClassRetryable — exponential backoff (capped) up to MaxAttempts, then give up.
//   - ClassUnknown   — fail-safe: one retry only, record "give_up" if still failing.
func SmartRetry(ctx context.Context, op func() error, policy RetryPolicy, classify func(string) ErrorClass, record func(MetricRecord)) error {
	if classify == nil {
		classify = Classify
	}
	err := op()
	if err == nil {
		if record != nil {
			record(MetricRecord{Attempt: 1, Class: ClassUnknown, Outcome: "success"})
		}
		return nil
	}

	class := classify(err.Error())
	maxAttempts := policy.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	// Fatal: never retry — report immediately so the outer loop escalates.
	if class == ClassFatal {
		if record != nil {
			record(MetricRecord{Attempt: 1, Class: class, Outcome: "fatal"})
		}
		return err
	}

	// RateLimit: single wait + single retry, independent of MaxAttempts.
	if class == ClassRateLimit {
		if record != nil {
			record(MetricRecord{Attempt: 1, Class: class, Outcome: "attempt"})
		}
		if !sleepCtx(ctx, policy.RateLimitWait) {
			if record != nil {
				record(MetricRecord{Attempt: 1, Class: class, Outcome: "cancel"})
			}
			return err
		}
		if record != nil {
			record(MetricRecord{Attempt: 2, Class: class, Outcome: "retry"})
		}
		if err2 := op(); err2 == nil {
			if record != nil {
				record(MetricRecord{Attempt: 2, Class: class, Outcome: "success"})
			}
			return nil
		}
		if record != nil {
			record(MetricRecord{Attempt: 2, Class: class, Outcome: "give_up"})
		}
		return err
	}

	// ClassRetryable / ClassUnknown: bounded backoff retry up to maxAttempts.
	// Unknown additionally caps at 1 retry regardless of policy.MaxAttempts.
	cap := maxAttempts
	if class == ClassUnknown {
		cap = 2 // first attempt + one retry
	}
	backoff := policy.BackoffBase
	if backoff <= 0 {
		backoff = 200 * time.Millisecond
	}
	for attempt := 2; attempt <= cap; attempt++ {
		if record != nil {
			record(MetricRecord{Attempt: attempt - 1, Class: class, Outcome: "attempt"})
		}
		if !sleepCtx(ctx, backoff) {
			if record != nil {
				record(MetricRecord{Attempt: attempt - 1, Class: class, Outcome: "cancel"})
			}
			return err
		}
		if record != nil {
			record(MetricRecord{Attempt: attempt, Class: class, Outcome: "retry"})
		}
		if err2 := op(); err2 == nil {
			if record != nil {
				record(MetricRecord{Attempt: attempt, Class: class, Outcome: "success"})
			}
			return nil
		}
		if attempt < cap {
			backoff *= 2
			if policy.BackoffMax > 0 && backoff > policy.BackoffMax {
				backoff = policy.BackoffMax
			}
		}
	}
	if record != nil {
		record(MetricRecord{Attempt: cap, Class: class, Outcome: "give_up"})
	}
	return err
}

// sleepCtx sleeps for d, but returns false if ctx is cancelled first so the
// caller can abort the retry chain promptly.
func sleepCtx(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return true
	}
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

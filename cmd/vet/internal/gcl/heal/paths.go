package heal

import (
	"context"
	"time"
)

// Path is one self-healing strategy for a single error class. Its Execute is
// a *wrapper* around the caller's op (the generator command) — it decides how
// op is retried (backoff, wait, escalate) but performs NO infrastructure
// action itself. `vet gcl run` is a GCL orchestrator; concrete remediation
// (switching an endpoint env, prompting for sudo) is injected by the caller
// before op runs, not by this package.
type Path struct {
	Name    string
	Class   ErrorClass
	Cost    int // 0 = cheapest (backoff-retry); ~5 = most expensive (escalate human)
	Execute func(ctx context.Context, op func() error) error
}

// History supplies per-(class, path) success rates so SelectBest can prefer
// the historically reliable path. Metrics implements it via PathSuccessRate
// (T11 aggregate SuccessRate() is a separate no-arg method kept for heal-stats).
type History interface {
	// PathSuccessRate returns 0.0..1.0 for the given class+path; 0.0 when no data.
	PathSuccessRate(class, name string) float64
}

// Paths is the registry of candidate self-healing paths, one entry per
// (class, strategy). Built once at package init; SelectBest reads it.
//
// 4 real error classes (NOT framework installer codes) × ≥2 strategies:
//
//	retryable  → backoff-retry(Cost1) / endpoint-switch-retry(Cost2)
//	rate_limit → wait-retry(Cost1) / wait-retry-long(Cost2)
//	fatal      → escalate(Cost4) / degrade-manual(Cost3)
//	unknown    → single-retry(Cost1) / escalate(Cost4)
var Paths = []Path{
	{Name: "backoff-retry", Class: ClassRetryable, Cost: 1, Execute: retryWith(3, 200*time.Millisecond, 1600*time.Millisecond)},
	{Name: "endpoint-switch-retry", Class: ClassRetryable, Cost: 2, Execute: retryWith(2, 400*time.Millisecond, 1600*time.Millisecond)},

	{Name: "wait-retry", Class: ClassRateLimit, Cost: 1, Execute: waitRetry(1 * time.Second)},
	{Name: "wait-retry-long", Class: ClassRateLimit, Cost: 2, Execute: waitRetry(5 * time.Second)},

	// fatal paths never retry — they immediately escalate / degrade. Cost is
	// high because a human is now in the loop. Execute returns the last error
	// (or a sentinel) so the runner records the escalation, not a retry.
	{Name: "escalate", Class: ClassFatal, Cost: 4, Execute: escalate},
	{Name: "degrade-manual", Class: ClassFatal, Cost: 3, Execute: escalate},

	{Name: "single-retry", Class: ClassUnknown, Cost: 1, Execute: retryWith(2, 200*time.Millisecond, 1600*time.Millisecond)},
	{Name: "escalate", Class: ClassUnknown, Cost: 4, Execute: escalate},
}

// SelectBest returns the cheapest path for class with the best historical
// success rate. Ties on cost break toward the higher success rate. When the
// history is empty for a path, it is treated as 0.0 and cost alone decides
// (cheapest first) — so we never get stuck on an unproven expensive path.
// Returns nil if the class has no registered path (caller degrades to a plain
// single op()).
func SelectBest(class ErrorClass, hist History) *Path {
	var best *Path
	bestScore := -1.0 // lower cost wins; use -1 sentinel before first pick
	for i := range Paths {
		p := &Paths[i]
		if p.Class != class {
			continue
		}
		rate := 0.0
		if hist != nil {
			rate = hist.PathSuccessRate(class.String(), p.Name)
		}
		// Primary key: cost ascending. Secondary key: rate descending.
		// Map to a single comparable score: prefer lower cost; among equal
		// cost prefer higher rate.
		score := float64(-p.Cost) + rate*0.001 // rate is a tiebreaker only
		if best == nil || score > bestScore {
			best = p
			bestScore = score
		}
	}
	return best
}

// PathsForClass returns all registered paths for class, cheapest first. Used by
// the runner to try alternatives after the best path fails.
func PathsForClass(class ErrorClass) []Path {
	var out []Path
	for _, p := range Paths {
		if p.Class == class {
			out = append(out, p)
		}
	}
	// cheapest first (stable; cost is already ascending in Paths order)
	return out
}

// retryWith returns a Path.Execute that runs op once, then retries on error
// with exponential backoff capped at maxBackoff, up to maxAttempts tries.
// ctx cancellation aborts immediately.
func retryWith(maxAttempts int, base, maxBackoff time.Duration) func(ctx context.Context, op func() error) error {
	return func(ctx context.Context, op func() error) error {
		if err := op(); err == nil {
			return nil
		}
		backoff := base
		for attempt := 2; attempt <= maxAttempts; attempt++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
			if err := op(); err == nil {
				return nil
			}
			backoff *= 2
			if maxBackoff > 0 && backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
		return op() // final attempt; returns its error (or nil)
	}
}

// waitRetry returns a Path.Execute that waits wait then runs op once (no
// further retry). Models "wait for a token / throttle to clear, then try".
func waitRetry(wait time.Duration) func(ctx context.Context, op func() error) error {
	return func(ctx context.Context, op func() error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
		return op()
	}
}

// escalate is the fatal-class path: it does NOT retry (retrying a parameter /
// permission / structural error is pointless and may amplify load). It returns
// the first op error so the runner can record a terminal failure and escalate.
func escalate(ctx context.Context, op func() error) error {
	return op()
}

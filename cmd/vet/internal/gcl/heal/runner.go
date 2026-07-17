package heal

import (
	"context"
	"time"
)

// pathTimeout caps a single path's execution so a stuck strategy (e.g. a long
// wait-retry) cannot stall the runtime. Each attempt of op is bounded by this
// via a derived context.
const pathTimeout = 30 * time.Second

// PathResult is the outcome of one Run: which path was chosen, its cost, and
// whether the op eventually succeeded.
type PathResult struct {
	Name       string `json:"path_name"`
	Class      string `json:"class"`
	Cost       int    `json:"cost"`
	Result     string `json:"result"` // "ok" | "fail"
	DurationMs int64  `json:"duration_ms"`
	Fallback   bool   `json:"fallback"` // true if a secondary path was tried
}

// Run applies the multi-path self-healing policy: classify → select best path
// → execute → fall back to cheaper alternatives on failure. It does NOT itself
// know how to fix the underlying error; op is the caller's generator command
// and each Path.Execute only decides how op is retried.
//
// Control flow:
//   - SelectBest(nil history) → cheapest path (cost ascending).
//   - nil path (class has no registry entry) → degrade to a single op() call
//     (T09-equivalent behavior), reported as path "single".
//   - fatal class → escalate path runs op once, no retry loop.
//   - on best-path failure, try remaining same-class paths cheapest-first; the
//     first that succeeds wins (Fallback=true); if all fail, return the last
//     error with Result="fail" (never silently swallowed).
func Run(ctx context.Context, class ErrorClass, op func() error, hist History) (PathResult, error) {
	start := time.Now()
	best := SelectBest(class, hist)
	if best == nil {
		err := runOp(ctx, op)
		return PathResult{
			Name:       "single",
			Class:      class.String(),
			Cost:       0,
			Result:     resultOf(err),
			DurationMs: time.Since(start).Milliseconds(),
		}, err
	}

	// Try the best path first (alts is cost-ascending, so best is first).
	// Fatal errors never retry, so we return the single escalation outcome
	// immediately. For transient classes we fall back to the remaining
	// same-class paths cheapest-first until one succeeds.
	alts := PathsForClass(class)
	var lastErr error
	for _, p := range alts {
		pctx, cancel := context.WithTimeout(ctx, pathTimeout)
		err := p.Execute(pctx, func() error { return runOp(ctx, op) })
		cancel()
		if err == nil {
			return PathResult{
				Name:       p.Name,
				Class:      class.String(),
				Cost:       p.Cost,
				Result:     "ok",
				DurationMs: time.Since(start).Milliseconds(),
				Fallback:   p.Name != best.Name,
			}, nil
		}
		lastErr = err
		if class == ClassFatal {
			// No fallback for fatal: report the single escalation attempt.
			break
		}
	}
	return PathResult{
		Name:       best.Name,
		Class:      class.String(),
		Cost:       best.Cost,
		Result:     "fail",
		DurationMs: time.Since(start).Milliseconds(),
		Fallback:   len(alts) > 1 && lastErr != nil && class != ClassFatal,
	}, lastErr
}

// runOp runs op once. ctx cancellation propagates as the returned error.
func runOp(ctx context.Context, op func() error) error {
	done := make(chan error, 1)
	go func() { done <- op() }()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

func resultOf(err error) string {
	if err == nil {
		return "ok"
	}
	return "fail"
}

# Design: L3 Exit Gate — CI Enforcement of Trace + Safety Invariants (P5/P7)

**Date:** 2026-07-17
**Author:** ve-skills (loop engineering)
**Status:** Approved
**Companion plan:** `docs/superpowers/plans/2026-07-17-l3-exit-gate.md`
**Companion task cards:** `docs/l2-to-l3-tasks` P5 (T07-derived), P7 (T08-derived)
**Prerequisite:** L3 runtime gate shipped (`5c467a1`); audit-chain hardening shipped (`9d9e03c`)

---

## Problem Statement

The L3 execution-risk gate (P8/P9) is implemented and unit-tested, but two L3 DoD items remain **partial** per `docs/l2-to-l3-plan.md`:

- **P5 (full-chain observability):** `vet gcl trace` validates `request_id` non-empty, but **CI does not run it**, and the runtime persists traces as `gcl-trace-*.json` while `vet check trace` only scans `incident-trace-*.json` — so real runtime traces are **never actually checked**.
- **P7 (safety-regression guard):** `policyguard` unit-tests pass, and `vet check policyguard` exists, but **CI does not run it**; the end-to-end guarantee "Safety=0 → never AUTO/executed" is not enforced in the pipeline.

Without these, `l2-to-l3-plan.md` §6 L3 DoD cannot be marked complete, and a future regression (e.g. a trace missing `request_id`, or `scoreDecision` accidentally returning AUTO for safety=0) would pass CI silently.

---

## Design Overview

**Approach:** Two changes, both small and low-risk.

1. **Fix the trace-coverage gap (real bug):** `vet check trace` must scan the actual runtime trace file `gcl-trace-*.json` (produced by `trace.PersistTrace`), not just `incident-trace-*.json`. Then wire `vet check trace` + `vet check policyguard` into `.github/workflows/validate.yml`.
2. **Wire existing checks into CI:** add two CI steps so P5 and P7 are enforced on every push/PR.

No new scorer logic. The invariants already exist (`policyguard` invariants 1/2/3; `trace.Check` request_id). This is enforcement + one glob fix.

---

## Component 1 — Trace glob fix (`cmd/vet/check.go` `traceCheck`)

**Current behavior (bug):** `traceCheck` skips any file not prefixed `incident-trace-`. `Run()` writes `gcl-trace-<ts>.json` (see `trace.PersistTrace`), so the files CI should validate are excluded.

**Fix:** scan `audit-results/` for `gcl-trace-*.json` (the runtime output) **and** `incident-trace-*.json` (the agent-produced wrapper). Both carry the same schema (`request_id` required per ve_call). Drop the `incident-trace-` only filter; instead match `*-trace-*.json` or explicitly both prefixes.

**Verification:** a `gcl-trace-*.json` with a missing `request_id` must cause `vet check trace` to exit 1.

---

## Component 2 — CI wiring (` .github/workflows/validate.yml`)

Add two steps after the existing `vet gcl gate` step:

```yaml
      - name: Check GCL trace request_id enforcement (P5)
        run: cmd/bin/vet check trace --root .

      - name: Check policy safety invariant (P7)
        run: cmd/bin/vet check policyguard --root .
```

Both already implemented in `cmd/vet/check.go` (`traceCheck`, `policyguard` case). This makes P5/P7 red on regression.

---

## Files Affected

| File | Change |
|------|--------|
| `cmd/vet/check.go` | `traceCheck`: widen glob to include `gcl-trace-*.json` |
| `.github/workflows/validate.yml` | add `vet check trace` + `vet check policyguard` steps |
| `cmd/vet/internal/check/trace/trace_test.go` | add a test: a `gcl-trace-*.json` with missing `request_id` fails `Check` (covers the runtime-naming path) |

---

## Success Criteria

- `vet check trace --root .` fails (exit 1) if any `gcl-trace-*.json` in `audit-results/` has a ve_call without `request_id`.
- `vet check policyguard --root .` fails if any fixture/plan violates invariants (safety=0→REFUSE, destructive→never AUTO, missing metadata→never AUTO).
- `.github/workflows/validate.yml` runs both checks; a forced regression makes the CI job red.
- `go build` + `go vet` + `go test ./...` clean in `cmd/vet`.
- `docs/l2-to-l3-plan.md` §6 L3 DoD P5/P7 can be marked ✅.

---

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| Widening glob picks up stale/foreign JSON | Only match `gcl-trace-` + `incident-trace-` prefixes; `trace.Check` validates schema, non-trace JSON fails fast and is excluded by prefix anyway. |
| CI red on pre-existing bad traces | None exist yet (runtime traces are produced per-run, not committed); `audit-results/` is gitignored. First CI run is clean. |
| `policyguard` fixture path missing in CI | `check.go` already handles missing fixture gracefully ("OK: no plan fixture found") — safe. |

---

## References

- `docs/l2-to-l3-plan.md` §3 P5/P7, §6 L3 DoD
- `cmd/vet/internal/check/trace/trace.go` — `Check`/`CheckDir`, request_id enforcement
- `cmd/vet/internal/check/policyguard/policyguard.go` — invariants 1/2/3
- `cmd/vet/internal/gcl/trace/trace.go` — `PersistTrace` writes `gcl-trace-*.json`
- `incident-loop-agent/SKILL.md` — trace contract (`request_id` required)

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

- **P5 (full-chain observability):** `vet check trace` validates `request_id`, but only against `incident-trace-*.json`. The runtime `vet gcl run` persists traces as `gcl-trace-*.json` with a **different schema** (`trace_schema_version`/`skill`/`operation_intent`; per-iteration `generator`/`decision`/`policy_decision`) that has **no `request_id` field at all** — the runtime never parses the `ve` CLI's `{"Response":{"RequestId":"..."}}` return. So (a) real runtime traces are never checked, and (b) P5's intent ("every `ve` call's RequestId recorded + validated") is not yet met because the runtime emits no RequestId.
- **P7 (safety-regression guard):** `policyguard` unit-tests pass, and `vet check policyguard` exists, but **CI does not run it**; the end-to-end guarantee "Safety=0 → never AUTO/executed" is not enforced in the pipeline.

Without these, `l2-to-l3-plan.md` §6 L3 DoD cannot be marked complete, and a future regression (e.g. a trace missing `request_id`, or `scoreDecision` accidentally returning AUTO for safety=0) would pass CI silently.

---

## Design Overview

**Approach:** Two changes, both small and low-risk.

1. **Fix the trace-coverage gap (real bug):** `vet check trace` must scan the actual runtime trace file `gcl-trace-*.json` (produced by `trace.PersistTrace`), not just `incident-trace-*.json`. Then wire `vet check trace` + `vet check policyguard` into `.github/workflows/validate.yml`.
2. **Wire existing checks into CI:** add two CI steps so P5 and P7 are enforced on every push/PR.

No new scorer logic. The invariants already exist (`policyguard` invariants 1/2/3; `trace.Check` request_id). This is enforcement + one glob fix.

---

## Component 1 — Runtime RequestId collection + dual-schema trace validation

**Two sub-changes (both required for P5):**

### 1a — Collect RequestId at runtime (`cmd/vet/internal/gcl/run/run.go`)

`ve` CLI returns JSON `{"Response":{"RequestId":"<id>"}}`. `Run()` currently discards it. Add `RequestID string` to `trace.Iteration` and populate it from each `runCommand` result: parse `Response.RequestId` out of the (masked) generator output. On the first iteration the generator has not run yet (gate may block), so `RequestID` is empty there — that is expected and valid.

### 1b — Runtime-trace validation (`cmd/vet/internal/gcl/trace/trace.go`)

> **Package boundary (important):** the existing `cmd/vet/internal/check/trace/trace.go`
> validates the **incident-trace** schema (`ticket_id`/`ve_calls[].request_id`), written by
> `incident-loop-agent`. The runtime `gcl-trace-*.json` is a **different** schema, written by
> `gcl/trace.PersistTrace`. The two are intentionally separate packages. We add a **new**
> `gcl/trace.Check` that validates the runtime shape — we do NOT mutate `check/trace.Check`.

`gcl/trace.Check(path)` validates the runtime trace:

- Detect runtime shape by presence of `trace_schema_version` (always `"v1"` for `PersistTrace` output).
- Validate: every `iterations[].request_id` non-empty **where the iteration actually ran a `ve` call** (i.e. `generator.command` is non-empty AND `decision != "POLICY_BLOCK"`); `redaction_pass == true`.
- Top-level `ticket_id`/`started_at`/`finished_at`/`policy_decision` are **not** required (they belong to the incident-trace schema).
- A file lacking `trace_schema_version` is treated as an incident/unknown trace and skipped by this checker (so the two checkers stay orthogonal).

The existing `check/trace.Check` keeps validating `incident-trace-*.json` exactly as before — no change to its schema or error messages.

### 1c — `traceCheck` glob (`cmd/vet/check.go`)

Scan both `gcl-trace-*.json` (runtime) and `incident-trace-*.json` (agent). Drop the `incident-trace-` only filter.

**Verification:** a `gcl-trace-*.json` with an empty `iterations[].request_id` must cause `vet check trace` to exit 1; a valid runtime trace (with RequestId) passes.

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
| `cmd/vet/internal/gcl/trace/trace.go` | `Iteration.RequestID` field (JSON `request_id`); **new** `Check(path)` validates runtime trace (dual-shape aware: runtime vs incident) |
| `cmd/vet/internal/gcl/run/run.go` | `runCommand`/`Run` parse `Response.RequestId` from generator output → store in `Iteration.RequestID` |
| `cmd/vet/check.go` | `traceCheck`: scan `gcl-trace-*` (runtime, via `gcl/trace.Check`) + `incident-trace-*` (via `check/trace.Check`) |
| `cmd/vet/internal/gcl/trace/trace_test.go` | add tests: runtime `gcl-trace-*.json` with missing/empty `request_id` fails; valid runtime trace passes |
| `.github/workflows/validate.yml` | add `vet check trace` + `vet check policyguard` steps |

---

## Success Criteria

- `vet gcl run` records the `ve` call `RequestId` into each `Iteration.request_id` of the persisted `gcl-trace-*.json`.
- `vet check trace --root .` fails (exit 1) if any `gcl-trace-*.json` in `audit-results/` has an empty `iterations[].request_id` (where an iteration actually ran a command), or if `redaction_pass` is false.
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

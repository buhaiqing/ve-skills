# Design: incident-loop-agent L3 Execution-Risk Gate (Conditional Autonomy)

**Date:** 2026-07-17
**Author:** ve-skills (loop engineering)
**Status:** Approved
**Superset plan:** `docs/superpowers/plans/2026-07-17-incident-loop-agent-l3.md`
**Companion task cards:** `docs/l2-to-l3-tasks/T01`–`T08`

---

## Problem Statement

The `incident-loop-agent` orchestration skill (7-step loop: `alert → triage → diagnose → propose → confirm → execute → validate → reflexion`) reached L2 (partial autonomy) with a blanket human gate: every destructive operation required `{{user.confirm}}`. L3 (conditional autonomy) requires a **graded execution-risk policy** so that:

- Bounded low-risk operations (read-only, single-resource state-changing at high confidence) **auto-execute with zero human prompts** (AUTO).
- Ambiguous / destructive / low-confidence operations **escalate to a human only when needed** (ASK).
- Any operation with `Safety = 0` is **never auto-executed** (REFUSE).

On-disk inspection (2026-07-16) showed the policy *spec* (prose + schema + allow-list + leaf metadata) and the *policy engine* (`scoreDecision`, `policyguard`) were already implemented and unit-tested, but `scoreDecision` was **never invoked inside `Run()`** — so the loop still executed every command unconditionally. This spec defines the L3 execution-risk gate that wires the policy into the runtime.

---

## Design Overview

**Approach:** Insert an execution-risk gate at the top of each GCL iteration in `vet gcl run`, *before* the generator command runs. The gate scores the operation from its derived intent (and optional Critic evidence) and blocks non-AUTO operations unless an external human confirmation is supplied.

**Scope:** One function group in `cmd/vet/internal/gcl/run/run.go` plus the `--confirmed` flag on `vet gcl run`, plus trace-record schema additions. No new subcommand. No change to the 7-step loop shape in `incident-loop-agent/SKILL.md` (already describes the policy gate at Step 5 / Step 5a).

---

## Component 1 — Policy Decision Function (`scoreDecision`)

Pure, deterministic scorer. Mirrors `docs/l2-to-l3-plan.md §3.2` 9-cell matrix and `incident-loop-agent/references/policies/execution-risk.md`.

**Inputs:** `skill, safetyClass, blastRadius, confidence, safety, metadataOK`
**Output:** `OpDecision` ∈ `{REFUSE, ASK, AUTO}`

### Decision logic (single source of truth)

```
if safety == 0:                       → REFUSE      # hard floor, overrides all
if safetyClass == "destructive":      → ASK         # never AUTO (per spec §2/§5)
if !metadataOK:                       → ASK         # fail-safe, never AUTO
if !allowedSkills[skill]:             → ASK         # outside domain allow-list
if safetyClass == "read_only"
   && confidence == "high":           → AUTO        # L3 happy path
if safetyClass == "mutating"
   && blastRadius == "single"
   && confidence == "high":           → AUTO
else:                                 → ASK
```

**Key invariants (must hold, see `policyguard`):**
1. `safety == 0` → REFUSE (never AUTO/ASK).
2. `safetyClass == "destructive"` → never AUTO (ASK or REFUSE only).
3. `metadataOK == false` → never AUTO (fail-safe ASK).

> **Note on destructive → ASK (not REFUSE):** This is intentional and **spec-compliant** (`execution-risk.md §2/§5`, T06 DoD line 71). A destructive op is not silent-refused; it is escalated to ASK so an upstream human gate (`--confirmed`) can authorize it. A blanket REFUSE would be the L2 behavior and is not the L3 design. The non-interactive runtime degrades an unconfirmed ASK to REFUSE (no human to ask). To keep this safe without downgrading to REFUSE, every `--confirmed` authorization is required to carry provenance (`--confirmed-by`) recorded in the trace (`Iteration.ConfirmedBy`), so the audit trail always answers "who authorized this op" — see §Audit-chain hardening.

---

## Component 2 — Policy Inputs Derivation (`policyInputs`)

Maps the operation intent (from `deriveOperationIntent`) and optional Critic scores into the `scoreDecision` arguments.

| Field | Derivation |
|-------|-----------|
| `safetyClass` | from `intent["safety_class"]` (default `read_only`) |
| `blastRadius` | `single` (a single `ve <svc> <Action>` maps to one resource) |
| `metadataOK` | `safetyClass != "" && allowedSkills[skill]` |
| `confidence` / `safety` | see below |

**Fail-safe when no Critic scores exist yet** (first iteration, gate runs before generator):
- `read_only` → confidence `high`, safety `1.0` → **AUTO** (read-only is inherently safe; L3 happy path).
- `mutating` / `destructive` → confidence `low` → falls to ASK/REFUSE (conservative).

**With Critic scores:** map the lowest non-passing rubric dimension (`correctness`/`idempotency`/`traceability`/`spec_compliance`) to confidence `medium`; a sub-1.0 `safety` → confidence `low`.

---

## Component 3 — Runtime Gate inside `Run()`

At the top of each iteration, before `runCommand`:

```go
sClass, bRadius, conf, safety, metaOK := policyInputs(opts.Skill, operationIntent, nil)
policy := scoreDecision(opts.Skill, sClass, bRadius, conf, safety, metaOK)
if policy != OpAuto && !(policy == OpAsk && opts.Confirmed) {
    blocked := policy
    if policy == OpAsk {            // non-interactive: no human to ask
        blocked = OpRefuse
    }
    // record Iteration{Decision: "POLICY_BLOCK", PolicyDecision: blocked}
    // record Final{Status: "POLICY_BLOCK", FailurePattern{...}}
    // PersistTrace + writebackFailurePattern
    return Result{ExitCode: 4, ...}  // POLICY_BLOCK
}
```

### Exit code contract (added to `docs/gcl-spec.md §5`)

| Code | Meaning |
|------|---------|
| `0` | PASS |
| `1` | MAX_ITER |
| `2` | invalid critic payload |
| `3` | SAFETY_FAIL (credential leak / safety=0) |
| `4` | POLICY_BLOCK (execution-risk gate blocked the op) |

### `--confirmed` flag

`vet gcl run --confirmed` sets `Options.Confirmed`. It is the runtime's external "human vouches for this ASK-class op" signal. Honors ASK → execute. Does **not** override REFUSE (Safety=0 or destructive-with-no-confidence still refuses unless separately authorized upstream). This is spec-compliant: ASK honors confirmation; REFUSE is the floor.

### Audit-chain hardening (confirmation provenance)

A `--confirmed` that authorizes an ASK-class op is only meaningful when a human actually confirmed it at Step 5. To make the audit trail answer "who authorized this op":

- `vet gcl run --confirmed-by <ticket_id|human_handle>` supplies provenance; stored in `Options.ConfirmedBy`.
- When an ASK op is authorized to execute, `Run` stamps `ConfirmedBy` into the trace `Iteration.ConfirmedBy` (and `gen.Args["confirmed_by"]`).
- `incident-loop-agent/SKILL.md` Step 5 requires `{{user.confirm}}` to be collected **before** `--confirmed` is passed, and `--confirmed` MUST be paired with `--confirmed-by`. Bare `--confirmed` with no provenance is treated as an audit violation.
- `Safety=0` REFUSE is never reachable via `--confirmed` (the gate checks `safety==0 → REFUSE` before the confirmation branch).

---

## Component 4 — Trace schema addition

`trace.Iteration` gains two fields, persisted on every iteration so the gate verdict and its authorization provenance are auditable:

- `PolicyDecision string` (JSON `policy_decision`) — the execution-risk verdict (AUTO/ASK/REFUSE).
- `ConfirmedBy string` (JSON `confirmed_by`, omitempty) — provenance of an external confirmation that authorized an ASK-class op to execute (ticket id / human handle). Empty when no confirmation was supplied.

---

## Files Affected

| File | Change |
|------|--------|
| `cmd/vet/internal/gcl/run/run.go` | `scoreDecision` (exists), `policyInputs` (exists), gate logic in `Run()` (added), `Options.Confirmed` + `Options.ConfirmedBy` fields; authorized ASK stamps `ConfirmedBy` into trace |
| `cmd/vet/internal/gcl/trace/trace.go` | `Iteration.PolicyDecision` + `Iteration.ConfirmedBy` fields |
| `cmd/vet/gcl.go` | `--confirmed` flag → `Options.Confirmed`; `--confirmed-by` flag → `Options.ConfirmedBy` |
| `docs/gcl-spec.md` | §4 Loop Flow step `[0.5] Execution-Risk Gate`; §5 Termination + exit-code table; §6 trace example |
| `incident-loop-agent/SKILL.md` | Step 5: non-interactive ASK w/o `--confirmed` → REFUSE (exit 4); `--confirmed` only after explicit `{{user.confirm}}`, paired with `--confirmed-by`; bare `--confirmed` = audit violation |

---

## Success Criteria

- `vet gcl run` blocks destructive / low-confidence ops unless `--confirmed` (ASK) or AUTO.
- `read_only` + high confidence auto-executes with zero prompts (L3 happy path).
- `Safety = 0` always REFUSE, never bypassed by `--confirmed`.
- `policy_decision` recorded in trace for every iteration.
- ASK-class op authorized via `--confirmed` persists `confirmed_by` provenance in trace.
- `--confirmed` requires explicit upstream `{{user.confirm}}`; bare `--confirmed` (no provenance) treated as audit violation.
- `go build ./...` + `go vet ./...` + `go test ./...` clean in `cmd/vet`.
- `policyguard` invariants (1/2/3) hold; `destructive → never AUTO` enforced.
- `docs/gcl-spec.md` documents the gate + exit code 4.

---

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| `--confirmed` silently authorizes destructive ops | By design destructive→ASK (not REFUSE); `--confirmed` honors ASK. If stricter zero-auto-destructive is later required, change `scoreDecision` to return REFUSE for destructive — single-line change, spec+code update together. |
| `--confirmed` abused with no provenance (who authorized?) | `--confirmed` MUST be paired with `--confirmed-by <ticket|handle>`; the value is persisted in `Iteration.ConfirmedBy`. `incident-loop-agent/SKILL.md` Step 5 forbids bare `--confirmed` without an upstream `{{user.confirm}}`; treated as an audit violation. |
| Gate runs before generator → no Critic evidence on iter 1 | `policyInputs` fail-safe gives read-only→AUTO, others→low→ASK/REFUSE. Conservative by default. |
| Trace drift | `PolicyDecision` + `ConfirmedBy` fields added to `Iteration`; `vet gcl trace` already validates schema. |

---

## References

- `docs/l2-to-l3-plan.md` §2–§4 — runtime wiring gap + 9-cell matrix
- `incident-loop-agent/references/policies/execution-risk.md` — L3 graded policy (prose)
- `incident-loop-agent/references/policies/execution-risk.schema.json` — machine-readable twin
- `incident-loop-agent/references/policies/domain-allowlist.md` — AUTO eligibility per skill
- `cmd/vet/internal/check/policyguard/policyguard.go` — invariant checker (mirrors `scoreDecision`)
- `docs/gcl-spec.md` — runtime GCL spec (gate + termination)

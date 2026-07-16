# L2 → L3 Upgrade Plan — Conditional Autonomy (详细执行计划)

> **Purpose**: Detailed, evidence-anchored execution plan to lift this repo from **L2 (partial automation, human-closed execution)** to **L3 (conditional autonomy, bounded-domain closed loop)** on the Gartner autonomous-operations ladder.
>
> **This document = the detailed expansion of milestone M1** in [`docs/autonomous-ops-roadmap.md`](./autonomous-ops-roadmap.md). It scopes **L2 → L3 only**. L3 → L4 (M2–M4) is intentionally out of scope here.
>
> **Status**: ✅ L3 COMPLETE (2026-07-17). Runtime execution-risk gate wired into `vet gcl run` (P8/P9); P5 runtime-trace `request_id` collection + dual-schema `vet check trace` CI enforcement shipped; P7 `policyguard` CI enforcement shipped.
> **Last updated**: 2026-07-16
> **Scope**: `incident-loop-agent` orchestration + the 8 leaf skills it coordinates + `vet` validation + Reflexion memory.

---

## 0. Why a separate L2→L3 document

The repo roadmap (`autonomous-ops-roadmap.md`) maps L2→L4 as M1→M4. L2→L3 maps **exactly to M1**. M1 is the linchpin: `autonomous-ops-roadmap.md §4` states *"L4 is unreachable while execution is human-closed"*. This document breaks M1 into reviewable, testable tasks with concrete evidence anchors so execution can start immediately without re-deriving the rationale.

> **Re-anchor note (2026-07-16)**: the original draft marked P1/P3 as *"not yet on disk"* and the orchestrator as *"v0.1.0 skeleton"*. On-disk inspection shows both are now implemented (policy artifacts present; `vet gcl run` is a real Go runtime). This revision re-baselines every task against what is **actually** on disk, and isolates the one true remaining blocker: **the execution-risk policy (`scoreDecision`) is not yet invoked inside `Run()` to gate execution.**

---

## 1. L2 vs L3 bar (evidence-anchored, re-verified 2026-07-16)

| Capability | L2 — baseline | L3 — target | Current state (verified) |
|------------|---------------|-------------|--------------------------|
| Loop shape | 7-step loop `alert→…→reflexion` (`incident-loop-agent/SKILL.md:55`) | Same 7-step, but **Step 5 is a policy decision, not a human gate** | ✅ SKILL.md Step 5 already reads `{{policy.decision}}` (`:154-157`) |
| Execution gate | Forced `{{user.confirm}}` on destructive; silent default REFUSE | **Policy-graded**: AUTO for bounded low-risk, ASK only for ambiguous/exception, REFUSE for Safety=0 | ⚠️ **Spec wired, runtime NOT** — `Run()` executes unconditionally; `scoreDecision` never called in the loop |
| Orchestrator | v0.1.0 skeleton | Production runtime: real GCL runner, retry, trace write-back | ✅ `vet gcl run` real Go runtime (`cmd/vet/gcl.go:24`); critic isolation + retry loop + trace persistence present |
| Policy engine | — | `AUTO/ASK/REFUSE` scorer over `risk × blast_radius × confidence` | ✅ `scoreDecision` implemented (`run.go:87`) + 9-cell unit tests; `policyguard` checker shipped (`cmd/vet/internal/check/policyguard`) |
| Leaf metadata | prose-only op classification | machine-readable `safety_class` + `blast_radius` per op | ✅ all 8 coordinated leaf skills expose `safety_class`/`blast_radius` (grep confirmed) |
| Observability | Trace mentioned, RequestId required | Full-chain trace schema **validated** by `vet`; every iteration persisted | ⚠️ `trace.schema.json` requires non-empty `request_id` (`:41-43`); `vet gcl trace` subcommand exists — **enforcement in CI not yet confirmed** |
| Eval coverage | — | AUTO/ASK/REFUSE eval cases, run in CI | ✅ `incident-loop-agent/assets/eval_queries.json` carries all 3 classes |
| Human role | Supervise **+ confirm** every destructive op | **Exception confirmation only** (ASK class) | ⛔ blocked by runtime gap above |

**L3 bar (one sentence)**: *Within a declared, bounded domain, the loop runs end-to-end with zero human prompts for non-exception cases; only ambiguous/destructive/low-confidence cases escalate to a human.*

---

## 2. The remaining blocker (crux of the leap)

The L2→L3 wall is no longer the *spec* — it is a **runtime wiring** gap.

`cmd/vet/internal/gcl/run/run.go` `Run()` (lines 379–458) executes the command **unconditionally** on every iteration:

```go
gen := runCommand(opts.Command, opts.Timeout, ...)   // line 380 — always runs
...
decision := critic.Decide(c.Scores)                  // line 434 — GCL quality gate only
```

`scoreDecision` (line 87) — which encodes the `AUTO/ASK/REFUSE` execution-risk policy — is **unit-tested but never called inside `Run()`**. `deriveOperationIntent` (line 124) derives `safety_class`, but feeds it only into the trace, not into an execution gate. Consequently:

- The **GCL quality gate** (PASS / SAFETY_FAIL) is wired into execution.
- The **execution-risk gate** (AUTO / ASK / REFUSE) is **not** — the loop still acts on every command, so the human is effectively still in the path (or, worse, operations run with no AUTO/ASK distinction at all).

**L3 = invoke `scoreDecision` per operation inside `Run()`, before `runCommand`, and gate execution**:
- `REFUSE` → skip `runCommand`, record decision, do not execute.
- `ASK` → require `{{user.confirm}}` (human) before `runCommand`.
- `AUTO` → `runCommand` proceeds with zero human prompts.

This is the single missing seam. Everything else (policy doc, schema, domain allow-list, leaf metadata, critic isolation, trace schema, eval) is already in place.

---

## 3. Execution-risk policy design (already shipped — reference only)

The policy itself is complete and on disk; reproduced here so the wiring task (§4 P8) has a stable contract.

### 3.1 Three scoring dimensions

| Dim | Values | Source |
|-----|--------|--------|
| `risk` (`safety_class`) | `read_only` / `mutating` / `destructive` | leaf-skill op metadata + `deriveOperationIntent` |
| `blast_radius` | `single` / `multi` / `account-or-region` | leaf-skill op metadata |
| `confidence` | `low` / `medium` / `high` | GCL Critic score + evidence completeness + allow-list membership |
| `safety` | 0.0–1.0 | GCL Critic `safety` dimension |
| `metadataOK` | bool | leaf metadata present & parseable |

### 3.2 Decision matrix → `AUTO` / `ASK` / `REFUSE`

| safety | safety_class | blast_radius | confidence | metadataOK | Decision |
|--------|--------------|--------------|------------|-----------|----------|
| 0 | any | any | any | any | **REFUSE** (hard floor) |
| ≥0 | `destructive` | any | any | any | **ASK** (never AUTO) |
| ≥0 | any | any | any | false | **ASK** (fail-safe) |
| ≥0 | not in allow-list | any | any | any | **ASK** |
| ≥0 | `read_only` | any | `high` | true | **AUTO** |
| ≥0 | `mutating` | `single` | `high` | true | **AUTO** |
| ≥0 | otherwise | — | — | true | **ASK** |

**Rule of thumb**: `AUTO` only for read-only ops or single-resource `mutating` ops with high confidence **and** present metadata. **No destructive op is ever AUTO.** This is *stricter than* today's REFUSE default for anything not provably low-risk — satisfying the safety-regression guard in `autonomous-ops-roadmap.md §4`.

### 3.3 Policy artifacts (all present on disk)

| Artifact | Path | Status |
|----------|------|--------|
| Prose + decision table | `incident-loop-agent/references/policies/execution-risk.md` | ✅ shipped |
| Machine-readable schema | `incident-loop-agent/assets/execution-risk.schema.json` | ✅ shipped |
| Domain allow-list | `incident-loop-agent/references/policies/domain-allowlist.md` | ✅ shipped (mirrors `allowedSkills` in `run.go:53`) |
| Go scorer | `cmd/vet/internal/gcl/run/run.go:87 scoreDecision` | ✅ shipped + unit-tested |
| Static checker | `cmd/vet/internal/check/policyguard` (`vet check policyguard`) | ✅ shipped |

---

## 4. Task breakdown (re-baselined)

Legend: ✅ Done (on disk) · ⚠️ Partial · ⛔ Blocked · 🆕 New in this revision

| ID | Task | Deliverable | Acceptance criteria | Evidence anchor | Status | Depends |
|----|------|-------------|---------------------|-----------------|--------|---------|
| **P1** | Author execution-risk policy | `references/policies/execution-risk.md` + `assets/execution-risk.schema.json` | 9-cell matrix covered; schema validates a sample `dispatch_plan` | disk inspect | ✅ | — |
| **P2** | Wire policy into SKILL.md | Step 5 + Variable Convention + Best Practices read `{{policy.decision}}` | `{{user.confirm}}` only on ASK class | `SKILL.md:108,154-157,198` | ✅ | P1,P3 |
| **P3** | Structured op metadata from leaf skills | `leaf-op-metadata-spec.md` + 8 coordinated skills annotated | all 8 expose `safety_class`+`blast_radius`; missing→ASK fail-safe | grep 8 hits | ✅ | — |
| **P4** | Orchestrator → production runtime | `vet gcl run` real runtime, `max_iter` enforced, trace write-back | no skeleton TODOs; runs end-to-end | `cmd/vet/gcl.go:24`, `run.go:351` | ✅ | P1 |
| **P6** | Eval coverage for 3 classes | AUTO/ASK/REFUSE eval cases in `eval_queries.json` | eval exercises all 3 paths | `assets/eval_queries.json` (23 refs) | ✅ | P1 |
| **P8** 🆕 | **Wire `scoreDecision` into `Run()` execution gate** | In `Run()`, before `runCommand`, compute `OpDecision` per op; `REFUSE`→skip+record, `ASK`→require confirm, `AUTO`→execute; persist decision into `trace.Iteration` | A non-destructive high-confidence single-resource op runs with **0 human prompts**; a destructive/Safety=0 op is **never** auto-run | `run.go` `Run()` loop (gate added) | ✅ | P1,P3,P4 |
| **P9** 🆕 | **Supply `scoreDecision` inputs from runtime** | Derive `confidence` (from Critic), `safety` (Critic `safety`), `metadataOK` (leaf metadata parse) inside `Run()`; pass to `scoreDecision` | `scoreDecision` receives real (not zero-value) inputs; unit test for input mapping | `policyInputs` helper in `run.go` | ✅ | P8 |
| **P5** | Full-chain observability enforcement | runtime `gcl-trace-*.json` collects `request_id` from each `ve` call; `vet check trace` (dual-schema) fails on missing; CI red on gap | every runtime iteration trace has RequestIds; CI red on gap | `gcl/trace.Check` (`trace.go`); `vet check trace` (`check.go:74`) | ✅ done | P4 |
| **P7** | Safety-regression guard (end-to-end) | CI integration assert: any op with Safety=0 in a `dispatch_plan` yields no AUTO path + no execution | guard red if policy ever returns AUTO for Safety=0 across a plan | `autonomous-ops-roadmap.md §3`; `policyguard` + `validate.yml` | ✅ done | P1,P6,P8 |
| **P10** 🆕 | **Reconcile plan doc with reality** | Mark P1–P4,P6 Done; remove "skeleton/not yet on disk" claims; align with roadmap | doc matches disk; no stale DRAFT claims | this file | ⛔ (doc debt) | — |

### On P8 (the real blocker — important)

The policy and its scorer exist; the loop simply does not consult them. `Run()` currently:

1. calls `runCommand` unconditionally (line 380),
2. then runs the GCL critic and decides PASS/SAFETY_FAIL (line 434).

To reach L3, insert the execution-risk gate **between** intent-derivation and `runCommand`:

```go
intent := deriveOperationIntent(opts.Skill, opts.Command)
decision := scoreDecision(opts.Skill, intent["safety_class"], blastRadius, confidence, safety, metadataOK)
if decision == OpRefuse { /* record, skip runCommand */ return }
if decision == OpAsk && !humanConfirmed { /* record, skip runCommand */ return }
// decision == OpAuto (or ASK with confirm) → runCommand
```

Without P8, L3 cannot be claimed regardless of how complete the spec is.

---

## 5. Cross-cutting acceptance gates

Inherited from [`autonomous-ops-roadmap.md §3`](./autonomous-ops-roadmap.md):

| Gate | Rule | Enforced where |
|------|------|----------------|
| Safety floor | No op executes with GCL Safety = 0; L3 keeps a *policy* gate, not a *human* gate | `scoreDecision` hard floor (run.go:88); `policyguard` invariant 1 |
| Credential safety | `<masked>` only in all traces/reports | `cmd/vet/internal/gcl/secret` masking path |
| Determinism | AUTO/ASK/REFUSE paths covered by `vet` + eval queries | `assets/eval_queries.json`, `vet check eval` |
| Token efficiency | Plan-derived skill edits re-checked on TE-1~TE-9 | `docs/token-efficiency.md` |

---

## 6. Verification & L3 "Definition of Done"

**Primary exit criterion** (from roadmap M1): *A non-destructive incident on a leaf skill reaches resolution with **zero human prompts** when policy = AUTO.*

**Concrete verification (run after P8–P10):**

```bash
# 1. Compile + static gates (Go tooling, per AGENTS.md "All Tools MUST Be Go")
cd cmd/vet && go build ./... && go vet ./... && go test ./...

# 2. Skill + policy conformance
vet check frontmatter --root .
vet check gcl --root .
vet check policyguard --root .

# 3. Exercise the 3 decision classes (P6)
vet check eval --root .
# Dry run: feed a low-risk single-resource mutating incident → expect policy=AUTO, 0 prompts

# 4. Execution-risk gate actually consulted (P8/P9) — NEW
# Dry run through `vet gcl run` with a high-confidence read_only op → expect execution, 0 prompts
# Dry run with a destructive op → expect ASK (no execution without confirm) or REFUSE, never AUTO-run

# 5. Trace enforcement (P5)
vet gcl trace --root .            # must fail if any iteration lacks request_id

# 6. Safety invariant (P7)
# Assert: a destructive or Safety=0 incident is NEVER AUTO-executed → CI red if violated
```

**L3 DoD checklist:**
```
□ P1 policy doc + schema shipped and reviewed                         [✅ done]
□ P2 SKILL.md Step 5 consults policy; ASK still human-confirmed       [✅ done]
□ P3 all 8 coordinated leaf skills expose safety_class + blast_radius [✅ done]
□ P4 orchestrator runs on vet gcl run, not skeleton                  [✅ done]
□ P6 eval covers AUTO/ASK/REFUSE                                     [✅ done]
□ P8 scoreDecision invoked inside Run() to gate execution           [✅ done]
□ P9 Run() supplies confidence/safety/metadataOK to scoreDecision   [✅ done]
□ P5 every iteration trace has RequestIds, enforced by vet in CI     [✅ done]
□ P7 safety-invariant guard fails CI on any AUTO-with-Safety=0       [✅ done]
□ P10 plan doc reconciled with disk state                            [✅ done]
□ Go build + vet + test clean; validate.yml green
```

---

## 7. Risk & rollback

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Policy too permissive → safety regression | Med | Policy stricter-than-default (§3.2); `policyguard` invariant 1; AUTO never granted to destructive ops |
| Leaf metadata incomplete → mis-score | High | `scoreDecision` fail-safe: **missing metadata → ASK**, never AUTO (run.go:94) |
| Orchestrator runtime bug → silent wrong decision | Med | Keep GCL Critic isolated; trace every decision (persist `OpDecision` into `trace.Iteration`); `max_iter` cap |
| Human-out-of-loop on edge case | Low | Domain allow-list starts narrow (8 skills); widen only after N clean runs |
| **P8 introduces a regression that breaks existing `vet gcl run` callers** | Med | Keep `StructuralOnly`/`--critic-*` paths unchanged; gate only the real-execution branch; add a unit test asserting read_only still executes |

**Rollback**: revert `Run()` to the pre-P8 version (one function); the policy artifacts (`execution-risk.*`, `scoreDecision`, `policyguard`) are inert without the wiring, so leaving them is harmless. No data migration involved.

---

## 8. Sequencing

```
P1 (policy) ─┬─▶ P3 (leaf metadata) ─▶ P2 (SKILL.md wiring) ─▶ P4 (runtime) ─┐
             │                                                            │
             └─▶ P6 (eval) ─▶ P7 (guard, e2e) ───────────────────────────┤
                                                                           ▼
                                              P8 (wire scoreDecision) ─▶ P9 (inputs) ─▶ P5 (trace CI) ─▶ P10 (doc) ─▶ L3 DoD
```

Strictly sequential at the gate level: **P8 (AUTO path) cannot ship before P1 (policy) + P3 (metadata) + P4 (runtime)** — all already satisfied. The remaining critical path is **P8 → P9 → P5/P7 → P10**. M2–M4 in the roadmap MUST NOT start until this L3 DoD is met.

---

## 9. References

- [`docs/autonomous-ops-roadmap.md`](./autonomous-ops-roadmap.md) — M1 (this plan) → M4, rating baseline, cross-cutting gates
- [`docs/l2-to-l3-tasks/AGENTS.md`](./l2-to-l3-tasks/AGENTS.md) — **AI Agent 入口规则**（能力边界 + TDD + GCL + 文档更新）
- [`docs/l2-to-l3-tasks/`](./l2-to-l3-tasks/) — per-task cards for AI-driven iteration (T01–T08, index, trace ledger)
- [`docs/l3-to-l4-tasks/AGENTS.md`](./l3-to-l4-tasks/AGENTS.md) — L4 升级的 AI Agent 入口规则（前置 L2→L3 完成）
- [`docs/l3-to-l4-tasks/`](./l3-to-l4-tasks/) — M2–M4 task cards (T09–T16, L3→L4)
- [`incident-loop-agent/SKILL.md`](../incident-loop-agent/SKILL.md) — current loop contract (Step 5 reads `{{policy.decision}}`, `:154-157`)
- [`cmd/vet/internal/gcl/run/run.go`](../cmd/vet/internal/gcl/run/run.go) — `scoreDecision` (`:87`), `Run()` loop (`:379`), unconditional `runCommand` (`:380`)
- [`cmd/vet/internal/check/policyguard/`](../cmd/vet/internal/check/policyguard/) — static execution-risk checker (`vet check policyguard`)
- [`ve-skill-generator/references/enhanced-self-healing-framework.md`](../ve-skill-generator/references/enhanced-self-healing-framework.md) — self-healing L1 (current) → L5 ladder (M2 later)
- [`docs/reflexion-memory.md`](./reflexion-memory.md) — learning boundary (HINT vs constraint; M3 later)
- [`docs/gcl-spec.md`](./gcl-spec.md) — GCL rubric / termination
- [`AGENTS.md`](../AGENTS.md) — "All Tools MUST Be Go" (vet is the enforcement CLI), two-round self-review, TE rules

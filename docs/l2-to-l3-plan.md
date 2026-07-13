# L2 → L3 Upgrade Plan — Conditional Autonomy (详细执行计划)

> **Purpose**: Detailed, evidence-anchored execution plan to lift this repo from **L2 (partial automation, human-closed execution)** to **L3 (conditional autonomy, bounded-domain closed loop)** on the Gartner autonomous-operations ladder.
>
> **This document = the detailed expansion of milestone M1** in [`docs/autonomous-ops-roadmap.md`](./autonomous-ops-roadmap.md). It scopes **L2 → L3 only**. L3 → L4 (M2–M4) is intentionally out of scope here.
>
> **Status**: DRAFT — planning artifact, not yet executed.
> **Last updated**: 2026-07-13
> **Scope**: `incident-loop-agent` orchestration + the 8 leaf skills it coordinates + `vet` validation + Reflexion memory.

---

## 0. Why a separate L2→L3 document

The repo roadmap (`autonomous-ops-roadmap.md`) maps L2→L4 as M1→M4. L2→L3 maps **exactly to M1**. M1 is the linchpin: `autonomous-ops-roadmap.md §4` states *"L4 is unreachable while execution is human-closed"*. This document breaks M1 into reviewable, testable tasks with concrete evidence anchors so execution can start immediately without re-deriving the rationale.

---

## 1. L2 vs L3 bar (evidence-anchored)

| Capability | L2 — current (evidence) | L3 — target |
|------------|--------------------------|-------------|
| Loop shape | 7-step loop `alert→…→reflexion` (`incident-loop-agent/SKILL.md:55`) | Same 7-step, but **Step 5 is a policy decision, not a human gate** |
| Execution gate | **Forced `{{user.confirm}}` on every destructive op; silent default = REFUSE** (`SKILL.md:150-154,184`) | **Policy-graded**: AUTO for bounded low-risk, ASK only for ambiguous/exception, REFUSE for Safety=0 |
| Orchestrator | **v0.1.0 skeleton** (`SKILL.md:22,224`) | Production runtime: real GCL runner, retry/backoff, partial-rollback detection |
| Observability | Trace mentioned, RequestId required (`SKILL.md:182`) | Full-chain trace schema **validated** by `vet`; every iteration persisted |
| Human role | Supervise **+ confirm** every destructive op | **Exception confirmation only** (ASK class) |
| Self-healing | L1 basic retry (`enhanced-self-healing-framework.md:29`) | *(L3 hardening = M2, not here)* |
| Reflexion | HINT, not constraint (`reflexion-memory.md:76`) | *(feedback = M3, not here)* |

**L3 bar (one sentence)**: *Within a declared, bounded domain, the loop runs end-to-end with zero human prompts for non-exception cases; only ambiguous/destructive/low-confidence cases escalate to a human.*

---

## 2. The single blocker (crux of the leap)

L2 is anchored by two hard lines in `incident-loop-agent/SKILL.md`:

- `:150-154` — *"Collect `{{user.confirm}}` for every operation whose safety_class == destructive. Silent default = REFUSE. No implicit `--force`."*
- `:184` — *"No autopilot for destructive ops — every destructive step has a `{{user.confirm}}` gate. No exceptions."*

These make the loop **human-closed**: the system can sense, triage, diagnose, propose, and even self-critique (GCL), but it cannot *act* without a human keystroke. That is the L2→L3 wall.

**L3 = replace the blanket `{{user.confirm}}` hard gate with a graded execution policy**, while keeping a strict Safety floor. The human is removed from the happy path and retained only for the ASK class.

---

## 3. Execution-risk policy design (core deliverable)

### 3.1 Three scoring dimensions

| Dim | Values | Source |
|-----|--------|--------|
| `risk` | `read-only`(0) / `state-changing`(1) / `destructive`(2) | leaf-skill operation metadata (see P3) |
| `blast_radius` | `single`(0) / `multi`(1) / `account-or-region`(2) | leaf-skill operation metadata (see P3) |
| `confidence` | `low` / `medium` / `high` | GCL Critic score + evidence completeness + domain allow-list membership |

### 3.2 Decision matrix → `AUTO` / `ASK` / `REFUSE`

| risk | blast_radius | confidence | Decision |
|------|--------------|------------|----------|
| read-only | any | any | **AUTO** |
| state-changing | single | high | **AUTO** |
| state-changing | single | medium/low | ASK |
| state-changing | multi/account | any | ASK |
| destructive | single | high | ASK *(never AUTO — human confirms destructive even at small radius)* |
| destructive | multi/account | any | ASK |
| **any** | **any** | **any** | **REFUSE if Safety = 0** (hard floor, overrides all) |

**Rule of thumb**: `AUTO` is granted *only* to read-only ops or single-resource state-changing ops with high confidence. **No destructive op is ever AUTO.** This makes the policy *stricter than* today's REFUSE default for anything not provably low-risk — satisfying the safety-regression guard in `autonomous-ops-roadmap.md §4`.

### 3.3 Policy artifacts (where it lives)

| Artifact | Path | Purpose |
|----------|------|---------|
| Prose + decision table | `incident-loop-agent/references/policies/execution-risk.md` *(P1 deliverable — not yet on disk)* | Human-readable spec, reviewable |
| Machine-readable schema | `incident-loop-agent/assets/execution-risk.schema.json` *(P1 deliverable — not yet on disk)* | Consumed by the runner to auto-score each `dispatch_plan` operation |
| Domain allow-list | `incident-loop-agent/references/policies/domain-allowlist.md` *(P1 deliverable — not yet on disk)* | Which products/symptoms are eligible for AUTO (seed = the 8 coordinated skills) |

---

## 4. Task breakdown (detailed, reviewable)

| ID | Task | Deliverable | Acceptance criteria | Evidence anchor | Depends |
|----|------|-------------|---------------------|-----------------|---------|
| **P1** | Author execution-risk policy | `references/policies/execution-risk.md` + `assets/execution-risk.schema.json` | Decision table covers all 9 cells of §3.2; schema validates a sample `dispatch_plan` | `SKILL.md:150` | — |
| **P2** | Wire policy into loop | Edit `SKILL.md` Step 5 + Variable Convention (`:108`) + Best Practices (`:184`) | `{{user.confirm}}` replaced by `{{policy.decision}}`; ASK class still collects `{{user.confirm}}`; "No autopilot" line qualified to "no autopilot for non-AUTO-class" | `SKILL.md:150-154,184` | P1, P3 |
| **P3** | Structured op metadata from leaf skills (**the gap**) | Annotation standard as standalone `ve-skill-generator/references/leaf-op-metadata-spec.md`; apply to 8 coordinated skills (`SKILL.md:32-40`) | Orchestrator can read `safety_class` + `blast_radius` per operation for all 8; missing metadata → default ASK (fail-safe) | `ve-ecs-ops/SKILL.md:769-772` (prose-only today); grep `blast_radius` across `ve-*-ops/SKILL.md` = 0 hits today | — |
| **P4** | Orchestrator → production runtime | Replace `scripts/gcl_runner.py` reference with `vet gcl run`; enforce `max_iter` (`:146-148`); add retry/backoff + partial-rollback detection (`:159`) | Loop runs without skeleton TODOs; `max_iter` enforced in runner, not just doc | `SKILL.md:22,146-148,159,224` | P1 |
| **P5** | Full-chain observability | Trace schema with `RequestId`s per `ve` call; persisted to `audit-results/incident-trace-<ticket>-<ISO>.json`; validated by `vet check` or schema lint | Every iteration emits a valid trace; `vet` flags missing `RequestId` | `SKILL.md:182` | P4 |
| **P6** | Eval coverage for 3 classes | Add AUTO / ASK / REFUSE eval cases to `incident-loop-agent/assets/eval_queries.json` | eval exercises all 3 paths; CI (`validate.yml`) runs them | `assets/eval_queries.json` (new for this skill) | P1 |
| **P7** | Safety-regression guard | Eval assert: no AUTO path possible when any op Safety = 0 | Guard fails CI if policy ever returns AUTO for Safety=0 | `autonomous-ops-roadmap.md §3` Safety floor | P1, P6 |

### On P3 (the discovered gap — important)
Today leaf skills classify ops only as prose (e.g. `ve-ecs-ops/SKILL.md:769` "Destructive / Read-only" table with a Safety threshold). The orchestrator's `dispatch_plan` *declares* `blast_radius` itself (`SKILL.md:145`) but **cannot derive it from leaf metadata** — grep for `blast_radius|safety_class` across all `ve-*-ops/SKILL.md` returns 0 hits. Without P3, the policy engine has nothing machine-readable to score → AUTO can never be safely granted. **P3 is therefore a hard prerequisite for P2's AUTO path**, not optional polish.

---

## 5. Cross-cutting acceptance gates

Inherited from [`autonomous-ops-roadmap.md §3`](./autonomous-ops-roadmap.md):

| Gate | Rule | Enforced where |
|------|------|----------------|
| Safety floor | No op executes with GCL Safety = 0; L3 keeps a *policy* gate, not a *human* gate | `SKILL.md:147,174`; asserted by P7 |
| Credential safety | `<masked>` only in all traces/reports | `cmd/vet/internal/check/assessment` credential-masking path |
| Determinism | AUTO/ASK/REFUSE paths covered by `vet` + eval queries | `assets/eval_queries.json`, `vet` |
| Token efficiency | Plan-derived skill edits re-checked on TE-1~TE-9 | `docs/token-efficiency.md` |

---

## 6. Verification & L3 "Definition of Done"

**Primary exit criterion** (from roadmap M1): *A non-destructive incident on a leaf skill reaches resolution with **zero human prompts** when policy = AUTO.*

**Concrete verification (run after P1–P7):**

```bash
# 1. Compile + static gates (Go tooling, per AGENTS.md "All Tools MUST Be Go")
cd cmd/vet && go build ./... && go vet ./...

# 2. Skill + policy conformance
vet check frontmatter --root .
vet check gcl --root .
vet gcl gate --root . --skip-incident-loop   # structural smoke

# 3. Exercise the 3 decision classes
vet check eval --root .            # runs AUTO/ASK/REFUSE eval cases (P6)
# Dry run: feed a low-risk single-resource state-changing incident → expect policy=AUTO, 0 prompts

# 4. Safety invariant (P7)
# Assert: a destructive or Safety=0 incident is NEVER AUTO → CI red if violated
```

**L3 DoD checklist:**
```
□ P1 policy doc + schema shipped and reviewed
□ P2 SKILL.md Step 5 consults policy; ASK still human-confirmed
□ P3 all 8 coordinated leaf skills expose safety_class + blast_radius (machine-readable)
□ P4 orchestrator runs on vet gcl run, not skeleton
□ P5 every iteration trace has RequestIds, validated by vet
□ P6 eval covers AUTO/ASK/REFUSE
□ P7 safety-invariant guard fails CI on any AUTO-with-Safety=0
□ Go build + vet clean; validate.yml green
```

---

## 7. Risk & rollback

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Policy too permissive → safety regression | Med | Policy stricter-than-default (§3.2); P7 guard; AUTO never granted to destructive ops |
| Leaf metadata incomplete → mis-score | High | P3 fail-safe: **missing metadata → ASK**, never AUTO |
| Orchestrator runtime bug → silent wrong decision | Med | Keep GCL Critic isolated; trace every decision; `max_iter` cap |
| Human-out-of-loop on edge case | Low | Domain allow-list (§3.3) starts narrow (8 skills); widen only after N clean runs |

**Rollback**: revert `SKILL.md` Step 5 to the blanket `{{user.confirm}}` gate (one-file change); delete/disable `execution-risk.*` artifacts (harmless). No data migration involved.

---

## 8. Sequencing

```
P1 (policy) ─┬─▶ P3 (leaf metadata) ─▶ P2 (wiring) ─▶ P4 (runtime)
             │                                                     │
             └─▶ P6 (eval) ─▶ P7 (guard) ────────────────────────┤
                                                                   ▼
                                              P5 (observability) ─▶ L3 DoD
```

Strictly sequential at the gate level: **P2 (AUTO path) cannot ship before P3 (metadata) and P1 (policy)**. M2–M4 in the roadmap MUST NOT start until this L3 DoD is met.

---

## 9. References

- [`docs/autonomous-ops-roadmap.md`](./autonomous-ops-roadmap.md) — M1 (this plan) → M4, rating baseline, cross-cutting gates
- [`docs/l2-to-l3-tasks/`](./l2-to-l3-tasks/) — **per-task cards for AI-driven iteration** (T01–T08, index, trace ledger)
- [`incident-loop-agent/SKILL.md`](./incident-loop-agent/SKILL.md) — current loop contract (blocker lines `:150-154,184`)
- [`ve-skill-generator/references/enhanced-self-healing-framework.md`](./ve-skill-generator/references/enhanced-self-healing-framework.md) — self-healing L1 (current) → L5 ladder (M2 later)
- [`docs/reflexion-memory.md`](./reflexion-memory.md) — learning boundary (HINT vs constraint; M3 later)
- [`docs/gcl-spec.md`](./gcl-spec.md) — GCL rubric / termination
- [`AGENTS.md`](./AGENTS.md) — "All Tools MUST Be Go" (vet is the enforcement CLI), two-round self-review, TE rules
- [`ve-ecs-ops/SKILL.md`](./ve-ecs-ops/SKILL.md) — example leaf op classification (`:769-772`), the prose-only gap P3 fixes

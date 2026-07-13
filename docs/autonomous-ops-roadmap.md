# Autonomous Operations Roadmap — L2 → L3 → L4

> **Purpose**: Translate the Gartner autonomous-operations rating of this repo (currently **L2**, with L3 skeleton present) into a milestone-based, testable upgrade plan toward **L4 (highly autonomous within bounded domains)**.
>
> **Status**: DRAFT — planning artifact. Not yet executed.
> **Last updated**: 2026-07-13
> **Scope**: `incident-loop-agent` orchestration + 29 `ve-*-ops` leaf skills + `vet` validation + Reflexion memory.

---

## 0. Rating Baseline (evidence-anchored)

| Capability | Current state | Evidence |
|------------|---------------|----------|
| Incident loop shape | 7-step closed loop (alert→…→reflexion) | `incident-loop-agent/SKILL.md:55` |
| Quality gate | GCL Critic loop + `vet` CLI | `cmd/vet/internal/check/`, `docs/gcl-spec.md` |
| Cross-session learning | Reflexion → `docs/failure-patterns.md` | `docs/reflexion-memory.md` |
| Destructive execution | **Forced human confirm; silent default = REFUSE** | `incident-loop-agent/SKILL.md:150-154,184` |
| Self-healing maturity | **L1 (basic retry) = current** | `ve-skill-generator/references/enhanced-self-healing-framework.md:29` |
| Orchestrator maturity | **v0.1.0 skeleton** | `incident-loop-agent/SKILL.md:22,224` |
| Trigger mode | Reactive (alarm/ticket) only | `incident-loop-agent/SKILL.md:76` |
| Reflexion role | HINT, not a constraint | `docs/reflexion-memory.md:76` |

**Verdict: L2** — sense/diagnose/propose reach L3 shape, but execution is human-closed and learning does not re-enter decisions.

---

## 1. Target Definitions

| Level | Bar to clear | Human role |
|-------|--------------|-----------|
| **L3 (conditional autonomy)** | Bounded-domain closed loop runs end-to-end; only *exception/ambiguous* cases escalate to human | Exception confirmation |
| **L4 (high autonomy)** | End-to-end autonomous within a declared domain; human sets policy/goals/guardrails only | Set strategy + guardrails |

---

## 2. Milestone Map

```
L2 ──M1──▶ L3 ──M2──▶ L3+ ──M3──▶ L4
 │          │          │           │
 │  Exec分级 │ 自愈L2→L3 │ 预测式触发 │ 策略自治+回滚
 │  +可观测  │ +遥测闭环 │ +学习回灌 │ +SLO目标驱动
```

### M1 — Execution Tiers + Observability (L2 → L3 floor)

**Goal**: Replace the blanket `{{user.confirm}}` hard gate with a *risk-graded* execution policy, and make the loop fully observable.

> **Detailed, task-level plan for M1 (L2 → L3 only)**: see [`docs/l2-to-l3-plan.md`](./l2-to-l3-plan.md) — execution-risk policy design, P1–P7 task breakdown, L3 Definition of Done.

| ID | Task | Acceptance | Evidence target |
|----|------|------------|-----------------|
| M1-1 | Define execution-risk policy: `risk × blast_radius × confidence` → AUTO / ASK / REFUSE | Policy table ships in `incident-loop-agent/references/` | `incident-loop-agent/SKILL.md:150` |
| M1-2 | Auto-execute low-risk + small-blast-radius + high-conf ops; ASK otherwise | ≥ 1 eval run with AUTO path exercised | `assets/eval_queries.json` |
| M1-3 | Full-chain trace to `audit-results/incident-trace-<ticket>-<ISO>.json` with `RequestId`s | Trace schema validated by `vet` | `incident-loop-agent/SKILL.md:182` |
| M1-4 | Promote `incident-loop-agent` v0.1.0 → production runtime (real GCL runner, retry/backoff, partial-rollback detection) | `max_iter` policies enforced in runner | `SKILL.md:146-148` |

**Exit criterion**: A non-destructive incident on a leaf skill reaches resolution with **zero human prompts** when policy = AUTO.

---

### M2 — Self-Healing L2→L3 + Closed Telemetry (L3 hardening)

**Goal**: Move self-healing from L1 (basic retry) to L3 (multi-path), with measured telemetry.

| ID | Task | Acceptance | Evidence target |
|----|------|------------|-----------------|
| M2-1 | Implement L2 intelligent retry (error classification → targeted retry) | Error taxonomy ≥ 10 codes reused from leaf skills | `enhanced-self-healing-framework.md:65-97` |
| M2-2 | Implement L3 multi-path self-healing (auto-select best path) | ≥ 2 distinct recovery paths per top error category | `enhanced-self-healing-framework.md:31` |
| M2-3 | Self-healing metrics: success rate > 80%, avg time < 30s | Metrics emitted and persisted per run | `enhanced-self-healing-framework.md:499-504` |
| M2-4 | Self-healing log to `/tmp/ve-self-healing.log` + aggregated | Log schema matches framework §6.2 | `enhanced-self-healing-framework.md:508-518` |

**Exit criterion**: Common transient faults (network blip, rate-limit, transient 5xx) self-heal without human or orchestrator escalation.

---

### M3 — Predictive Trigger + Learning Feedback (L3 → L3+)

**Goal**: Shift from reactive to predictive, and make Reflexion learning re-enter decisions.

| ID | Task | Acceptance | Evidence target |
|----|------|------------|-----------------|
| M3-1 | Add predictive trigger source (CMS metric trend / capacity / slow-query degradation) | ≥ 1 predictive scenario pre-empts an alarm | `docs/skill-routing-graph.md` |
| M3-2 | Pattern→policy transpiler: `count ≥ 10` patterns auto-promote to guardrail/threshold rules | Promotion path unit-testable | `docs/reflexion-memory.md:63` |
| M3-3 | Promote Reflexion from HINT to decision constraint for high-frequency patterns | High-freq pattern changes auto-exec threshold | `docs/reflexion-memory.md:76` |
| M3-4 | Versioned policy library (guardrails as code, reviewable) | Policy diffs tracked in repo | new `incident-loop-agent/references/policies/` |

**Exit criterion**: Repeat incidents on same symptom converge faster (2nd incident < 50% of 1st loop steps), provably via trace comparison.

---

### M4 — L4 Autonomous Domain (L3+ → L4)

**Goal**: Bounded-domain end-to-end autonomy; human sets goals only.

| ID | Task | Acceptance | Evidence target |
|----|------|------------|-----------------|
| M4-1 | SLO as objective function: loop optimizes to maintain SLO | SLO defined per domain; loop acts to preserve it | `references/advanced/aiops.md` (per skill) |
| M4-2 | Validation-failure auto-rollback (not just retry) | Rollback plan auto-applied on validation fail | `incident-loop-agent/SKILL.md:145,159` |
| M4-3 | Declare autonomous domain envelope (which products/symptoms are L4) | Explicit allow-list + blast-radius cap | `incident-loop-agent/references/policies/` |
| M4-4 | Human interface = policy/goals dashboard, not per-op confirm | Zero per-op prompts inside envelope | M1-2 / M4-3 |

**Exit criterion**: A declared L4 domain runs N consecutive incidents end-to-end (triage→validate) with only policy/goals input from humans; escalations only outside envelope.

---

## 3. Cross-Cutting Acceptance Gates

| Gate | Rule | Where enforced |
|------|------|----------------|
| Safety floor | No destructive op executes with GCL Safety = 0; L4 keeps a *policy* gate, not a *human* gate | `incident-loop-agent/SKILL.md:147,174` |
| Credential safety | `<masked>` only in all traces/reports | `cmd/vet/internal/check/assessment/assessment.go:143-151` |
| Determinism | Auto-exec paths covered by `vet` + eval queries | `assets/eval_queries.json`, `vet` |
| Token efficiency | Roadmap-derived skill edits re-checked on TE-1~TE-9 | `docs/token-efficiency.md` |

---

## 4. Risk & Sequencing

- **M1 is the linchpin**: L4 is unreachable while execution is human-closed. Do not start M4 before M1+auto-rollback (M4-2) exists.
- **Safety regression risk**: replacing human confirm with policy requires the policy to *stricter-than* the current REFUSE default for anything not provably low-risk. Gate M1-2 behind eval coverage.
- **Learning feedback (M3-2/3) must not become a hard gate prematurely** — keep rollback to HINT for `count < 10`, promote to constraint only at `count ≥ 10`.

---

## 5. Next Steps

- Read `incident-loop-agent/SKILL.md` — current loop contract.
- Read `ve-skill-generator/references/enhanced-self-healing-framework.md` — self-healing L1→L5 ladder.
- Read `docs/reflexion-memory.md` — learning boundary (HINT vs constraint).
- Read `docs/skill-routing-graph.md` — trigger/routing table to extend for predictive (M3-1).
- Execute M1 first; milestones are strictly sequential at the gate level.

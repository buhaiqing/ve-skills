---
name: incident-loop-agent
description: >-
  Use when the agent needs to handle customer-reported cloud incidents on
  Volcengine (火山引擎) — alert triage → cross-product diagnosis → fix
  proposal → safe execution → outcome validation → reflexion. Coordinate
  one or more `ve-*-ops` skills as the loop engine. Triggers: incoming
  CMS alarm, JIRA DOPS ticket, customer chat report with concrete symptom,
  or scheduled SRE patrol. Not for live single-CLI commands (delegate to
  the relevant `ve-*-ops` skill instead). Not for skill authoring (use
  `ve-skill-generator`).
license: MIT
compatibility: >-
  ve-skills repository checkout with all `ve-*-ops` skill directories
  available, `ve` CLI installed and authenticated (`VOLCENGINE_ACCESS_KEY`,
  `VOLCENGINE_SECRET_KEY`), `docs/skill-routing-graph.md` reachable,
  `docs/failure-patterns.md` reachable (for Reflexion step), GCL runner
  (`vet gcl run`) available, and a Harness AI Agent runtime
  compatible with the existing `ve-skill-generator` flow.
metadata:
  author: ve-skills
  version: "0.3.2"
  last_updated: "2026-08-05"
  runtime: Harness AI Agent, Claude Code, Cursor, or compatible Agent runtimes
  type: orchestration-skill
  guidance_freedom_level: medium
  cli_applicability: cli-only
  cli_support_evidence: >-
    Coordinates existing `ve-*-ops` skills via their official `ve` CLI
    path; no new CLI surface introduced in v0.1.0.
  parent_skill: null
  coordinates:
    - ve-cms-ops
    - ve-ecs-ops
    - ve-rds-mysql-ops
    - ve-redis-ops
    - ve-vpc-ops
    - ve-iam-ops
    - ve-kms-ops
    - ve-billing-ops
  references:
    - docs/skill-routing-graph.md
    - docs/failure-patterns.md
    - docs/reflexion-memory.md
    - docs/gcl-spec.md
---

# Incident Loop Agent (Orchestration Skill)

## Overview

`incident-loop-agent` is the **orchestration skill** for Volcengine cloud incidents. It does **not** operate a single product; it **coordinates** the existing 28 `ve-*-ops` skills to run an end-to-end loop:

```
alert → triage → diagnose → propose → confirm → execute → validate → reflexion
```

The loop is bounded: every destructive proposal must pass the GCL Safety = 1 floor before execution; every iteration produces a trace; every failure-pattern persists into `docs/failure-patterns.md` so the **next** incident on the same symptom is faster.

> **Boundary**: this skill is read-and-decide plus light-write only. Bulk data writes (e.g. `ve rds CreateDBInstance`) still flow through the target `ve-*-ops` skill, not through `incident-loop-agent` itself.

## Five Core Standards

| # | Standard | How this skill fulfills it |
|---|----------|----------------------------|
| 1 | ✅ Clear Boundaries | `Trigger & Scope` separates "I am the right loop for this" from "defer to leaf skill". |
| 2 | ✅ Structured I/O | `{{output.*}}` captures `decision`, `dispatch_plan`, `critic_verdict`. `{{user.*}}` mandatory for confirmations. `{{env.*}}` for credentials. |
| 3 | ✅ Explicit Actionable Steps | Loop flow is a 7-step sequence with verified pre/post conditions. |
| 4 | ✅ Complete Failure Strategies | `references/rubric.md` ships 5 dimensions + 2 extras; Safety = 0 → ABORT. `docs/failure-patterns.md` `## 6. Incident Response Failures` + auto-generated `## Extracted from GCL Traces` cover orchestration-layer failures. |
| 5 | ✅ Absolute Single Responsibility | One loop. Cross-product calls route through the matched `ve-*-ops` skill, never duplicated here. |

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- An incoming entity matches one of: CMS alarm (any product), JIRA DOPS ticket with severity `Sev-1` / `Sev-2`, customer-facing chat report with a concrete symptom (latency, error, resource outage)
- The triggering entity names or implies a Volcengine product (ECS / RDS / Redis / VPC / CLB / etc.)
- A scheduled SRE patrol reaches a milestone and the user wants the loop to consolidate findings

### SHOULD NOT Use This Skill When

- Task is a single, well-defined `ve <service> <Action>` query → delegate to the relevant `ve-*-ops` skill
- Task is skill authoring → delegate to `ve-skill-generator`
- Task is pure documentation lookup → delegate to `ve-billing-ops` or knowledge-base lookup
- Symptom is unrelated to Volcengine or there is no concrete resource scope → refuse and ask for scope

### Delegation Rules

> See `docs/skill-routing-graph.md` for the canonical alarm pattern table. This skill **uses** that table; it does not redefine it.

| Symptom cluster | Primary skill | Secondary (on demand) |
|----------------|---------------|------------------------|
| CPU > 90% (any product) | per routing graph | per routing graph |
| Network (latency / packet loss / ACL change) | `ve-vpc-ops` | `ve-security-group-ops` |
| AuthFailure / UnauthorizedOperation | `ve-iam-ops` | `ve-kms-ops` |
| Cost spike | `ve-billing-ops` | primary skill |
| Unknown pattern | `ve-cms-ops` (correlation first) | match any skill |

This skill MUST refuse to run a leaf `ve <service> <Action>` directly; it constructs a `dispatch_plan` and hands to the target skill.

## Variable Convention (Agent-Readable)

| Placeholder | Meaning | Source |
|-------------|---------|--------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | From runtime | NEVER ask; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | From runtime | NEVER ask; fail if unset |
| `{{env.VOLCENGINE_REGION}}` | From runtime | Use documented default if skill allows |
| `{{user.confirm}}` | Yes/No to ASK-class op | Collected only when `{{policy.decision}} == ASK` |
| `{{user.scope_choice}}` | When dispatch_plan lists multiple resource scopes | Collect before dispatch |
| `{{policy.decision}}` | Per-operation policy verdict | Computed by T06 `scoreDecision`; one of `AUTO`, `ASK`, or `REFUSE` |
| `{{policy.reason}}` | Human-readable decision rationale | Always populated; trace-only field |
| `{{user.ticket_id}}` | JIRA / DOPS / CMS alarm ID | Echo into trace |
| `{{input.incident_payload}}` | From triggering channel (alarm JSON, ticket, chat) | Required entrypoint |
| `{{output.triage_class}}` | routing-graph primary skill | Internal only |
| `{{output.dispatch_plan}}` | JSON {primary, secondary, blast_radius} | Internal only |
| `{{output.critic_verdict}}` | GCL Critic JSON | Persist into trace |
| `{{output.failure_pattern}}` | One-line token if any iteration fails | Auto-aggregated by GCL write-back into `docs/failure-patterns.md` `## Extracted from GCL Traces` (dedup by skill+pattern) |
| `{{output.alarm_groups}}` | Alarm groups from merge pre-processing | P0-2: `vet check alarm-merge` dedup by ResourceID+Metric+5min window |
| `{{output.is_storm}}` | True if alarm count exceeds storm threshold | P0-2: IsStorm = Count > 10 in `cmd/vet/internal/alarm/merge.go` |

> **Security**: this skill never sees `VOLCENGINE_SECRET_KEY`. It only authenticates via env and delegates. No credential is ever written to a trace or report.

## Loop Flow

The loop runs in **7 steps** with mandatory pre/post conditions. Each step emits `{{output.*}}` for the next.

### Step 1 — `{{input.incident_payload}}` ingestion + alarm pre-processing

- Source: CMS alarm webhook / JIRA DOPS create / chat trigger / patrol milestone.
- Normalize into `{ticket_id, severity, product_hint, raw_event, observed_at}`.
- **P0-2 Alarm Merge (告警归并引擎)**: when `{{input.incident_payload}}` contains ≥2 alarms:
  1. Group by `(Product, ResourceID, Metric, 5-min time-slot)` using `cmd/vet/internal/alarm/merge.go` `Merge()` → `AlarmGroup`.
  2. `IsStorm = true` when any group's `Count > 10`; route to `ve-cms-ops` Rule 5 first for storm classification.
  3. Cross-product same `ResourceID` across ECS/RDS/Redis → merge into single group (single root cause).
  4. `RootCause =` earliest alarm in each group (by `observed_at`).
  5. Emit `{{output.alarm_groups}}` and `{{output.is_storm}}` for downstream steps.
- Fail-fast on missing `product_hint` → ask user once.

### Step 2 — Triage (load routing graph)

- Read `docs/skill-routing-graph.md` (lazy, ~100 lines).
- If `{{output.is_storm}} == true`: prefer `ve-cms-ops` Rule 5 (alarm storm correlation) as primary.
- Match `product_hint + symptom` against the alarm pattern table.
- Emit `{{output.triage_class}}` = `{primary, secondary[], confidence}`.
- Unknown pattern → route to `ve-cms-ops` first (Rule 5).

### Step 3 — Diagnose (parallel read-only `Describe*` calls)

- Dispatch to `triage_class.primary` with `read-only` mode.
- Cap at `min(15, alarm_count × 3)` read calls; log skipped calls.
- Emit `{{output.diagnosis_evidence}}` = raw observation snippets.
- Hit `cross_skill` boundary → delegate (do not absorb).
- **Billing data integration** — query affected resource's monthly cost via `ve billing DescribeBillDetail --BillingCycle $(date +%Y-%m) --InstanceId $resource_id`. Flag cost anomalies (spike > 50% vs prior month). If `ve-billing-ops` unavailable, omit cost data but flag in trace.

### Step 4 — Propose fix + GCL loop

- Build `dispatch_plan` = `{operations[], blast_radius, pre_state_snapshot, rollback_plan}`.
- Run GCL via `vet gcl run` with this skill's `references/rubric.md`.
- Safety must equal 1.0 on every destructive operation.
- `max_iter = 3` for repair loop; `max_iter = 2` for any destructive.
- **Cost assessment** — for each proposed fix, generate cost estimate via `ve billing DescribeBillDetail`. Compare alternative strategies (e.g., scale up vs add replica). Surface cost difference: "方案 A = ¥500/mo | 方案 B = ¥300/mo". Include cost in `dispatch_plan`.

### Step 5 — Confirm (user gate, ASK-class only)

- Read `{{policy.decision}}` for each operation (computed by T06 scorer before this step).
- If `{{policy.decision}} == ASK`: collect `{{user.confirm}}` verbatim; trace it. **Note**: in the non-interactive `vet gcl run` runtime an ASK without an external `--confirmed` signal degrades to `REFUSE` (no human to ask) and the operation is blocked (exit 4, `POLICY_BLOCK`). **Authorization rule**: a destructive or otherwise ASK-class op is authorized to execute **only after** an explicit `{{user.confirm}}` is collected at this step; the runtime flag `--confirmed` carries that authorization and MUST be paired with `--confirmed-by <ticket_id|human_handle>` so the trace records *who* authorized the op. Never pass `--confirmed` without a human confirmation captured here — bare `--confirmed` with no provenance is treated as an audit violation.
- If `{{policy.decision}} == AUTO`: proceed directly to Step 6 (no prompt).
- If `{{policy.decision}} == REFUSE`: skip to Step 7 with `failure_pattern = "policy_refused"`; do NOT execute the operation.
- **No silent default** — every REFUSE is explicit from the policy scorer.

### Step 5a — Policy evaluation

- Before Step 5, the T06 scorer computes `{{policy.decision}}` per operation using:
  - Inputs: `safety_class`, `blast_radius`, `confidence`, `safety` from leaf-op metadata
  - Policy: `references/policies/execution-risk.md` (3×3 decision matrix, Safety=0→REFUSE hard floor)
  - Schema: `assets/execution-risk.schema.json` (if/then hard rules)
  - Allowlist: `references/policies/domain-allowlist.md` (AUTO eligibility per skill)
- Output: one `{{policy.decision}}` (`AUTO`/`ASK`/`REFUSE`) and `{{policy.reason}}` per operation
- `{{policy.decision}}` is then read in Step 5 to route each operation appropriately

### Step 6 — Execute (delegated, not direct)

- Hand `dispatch_plan` to the matched `ve-*-ops` skill; it runs the actual `ve <svc> <Action>` calls.
- This skill monitors, on validation failure auto-applies `rollback_plan` (T16), verifies pre-state restoration; only escalates if rollback itself fails.
- Re-emit trace on every retry.

### Step 7 — Reflexion (automatic write-back)

- Failure patterns are **not** hand-written by this skill. When the GCL loop ends (MAX_ITER / SAFETY_FAIL), `vet gcl run` calls `_writeback_failure_pattern` → `gcl_trace_aggregate.update_failure_patterns_file`, which appends/merges the pattern into `docs/failure-patterns.md` `## Extracted from GCL Traces (auto-generated)` block (fields `skill` / `pattern` / `category` / `source`), **dedup by `(skill, pattern)`**.
- The separate `## 6. Incident Response Failures` table is a manually-seeded, design-placeholder section (dedup by `(scenario, failure_pattern)`); it is not produced by the write-back.
- At session end, also update this skill's own working memory (`memory/working.json`) with `last_run`, `triage_class`, `success`.

> **Termination**: same as GCL spec §5 — PASS / MAX_ITER / SAFETY_FAIL. SAFETY_FAIL aborts regardless of progress.

## Quality Gate (GCL)

This skill runs the standard 7-dimension rubric (defined in `references/rubric.md`) plus 2 orchestration-specific dimensions, totaling 9 dimensions (Reflexion integration + Cross-skill delegation). `max_iter = 3` for read-only paths, `max_iter = 2` for any iteration that touches a destructive leaf skill.

- Safety = 0 → ABORT immediately; never return partial.
- Spec Compliance requires the loop to actually read `docs/skill-routing-graph.md` (not skipped).
- Reflexion dimension tracks whether `failure_pattern` was persisted when one exists.

See `references/rubric.md` for the per-dimension rules.

## Operational Best Practices

- **Trace every iteration** — even when the loop exits on PASS, persist the trace to `audit-results/incident-trace-<ticket_id>-<ISO>.json`. Trace MUST include `RequestId`s from every `ve` call.
- **Bounded Reflexion** — `docs/failure-patterns.md` ≤ 200 lines. Prune patterns with `count < 3`. Promote patterns with `count ≥ 10` to Anti-Patterns in `ve-*-ops` skills.
- **No autopilot for non-AUTO class** — only operations with `{{policy.decision}} == AUTO` auto-execute; ASK requires `{{user.confirm}}`; REFUSE is blocked outright.
- **No credential in trace** — `<masked>` only, `redaction_pass: true`.
- **First incident is the slowest** — expect the GCL loop to fail once before the rubric converges. The second incident on the same symptom should be faster if Reflexion worked.

### Periodic Cost Patrol Mode

A scheduled (weekly) variant of the loop that proactively finds cost waste:

1. **Scan** — iterate over all 28 product skills; for each, call read-only `Describe*` / `List*` to enumerate resources.
2. **Classify** — flag candidates for: unattached EIP (`Status=Available`), idle ECS (check CPU/network metrics via `ve cms`), oversized RDS (low connection count), expired PrePaid instances.
3. **Report** — aggregate findings into `audit-results/cost-patrol-<ISO>.json` with estimated monthly savings.
4. **Act** — `AUTO` for safe reclamation (release unattached EIP); `ASK` for destructive (terminate instance). Policy classification reuses the same `execution-risk.md` matrix.
5. **Learn** — recurring idle patterns are persisted via Reflexion as cost-related failure patterns. The next patrol on the same skill is faster.

## Cross-Skill Delegation (extends `docs/skill-routing-graph.md`)

This skill never reimplements routing. It loads the routing table, matches, and hands off. The five critical rules in `docs/skill-routing-graph.md` apply unchanged:

1. AuthFailure → `ve-iam-ops`
2. Alarm storm → `ve-cms-ops` correlation first
3. Network issue → `ve-vpc-ops` first
4. Cost spike → `ve-billing-ops`
5. Unknown pattern → `ve-cms-ops`

## What This Skill Does

- Listens for incoming incident payloads
- Triages via `docs/skill-routing-graph.md`
- Runs a bounded GCL loop with Safety = 1 floor
- Persists every failure pattern via GCL Reflexion write-back into `docs/failure-patterns.md` `## Extracted from GCL Traces` for next-time speedup
- Hands off all leaf `ve` calls to the matched `ve-*-ops` skill
- Auto-executes policy-AUTO operations without human prompt; prompts only for ASK; blocks REFUSE outright

## What This Skill Does NOT Do

- Run a `ve <service> <Action>` call directly (always delegate)
- Author or modify other skills (use `ve-skill-generator`)
- Bypass destructive-op confirmation even if user insists in chat (state limitation)
- Operate outside Volcengine

## Next Steps

- Read `references/rubric.md` — the 9-dimension GCL rubric for this skill
- Read `references/prompt-templates.md` — Generator / Critic / Safety / Trace / Cross-skill prompt skeletons
- Read `docs/skill-routing-graph.md` — the routing table this skill consumes
- Read `docs/failure-patterns.md` `## Extracted from GCL Traces` (auto-generated) and `## 6. Incident Response Failures` (seeded) — where this skill's learning lives

## References

- `references/policies/execution-risk.md` — L3 graded execution-risk policy (replaces the L2 `{{user.confirm}}` hard gate); `risk × blast_radius × confidence → AUTO/ASK/REFUSE` with a Safety = 0 → REFUSE floor.
- `references/policies/execution-risk.schema.json` — machine-readable twin of the policy; draft 2020-12, enum-validated `Operation` shape, and `if/then` hard rules: Safety = 0 → REFUSE (overrides all cells), missing metadata → ASK (fail-safe). Consumed by the T06 scorer.
- `references/policies/domain-allowlist.md` — narrow L3 AUTO eligibility: the 8 coordinated skills (matches `coordinates` 1:1) + per-skill symptom whitelist + explicit destructive-op exclusions + expansion policy.
- `ve-skill-generator/references/leaf-op-metadata-spec.md` — contract for per-operation `safety_class` + `blast_radius` columns (the L3 policy inputs); placement rule, fail-safe (missing → ASK), and C18 update rule.

## Changelog

| Version | Date | Change |
|---------|------|--------|
| 0.3.2   | 2026-08-05 | T16 L4 envelope: `references/policies/autonomy-envelope.md` 嵌入 fenced YAML `domains:` 块，使 `vet autonomy test --envelope` 可解析（L4 出口 DoD #10）；bump `version` → 0.3.2. |
| 0.3.1   | 2026-07-13 | T05 wiring: replace L2 `{{user.confirm}}` hard gate with policy decision gate; add `{{policy.decision}}`/`{{policy.reason}}` variables; Step 5 now reads policy verdict (AUTO→auto-exec, ASK→prompt, REFUSE→block); add Step 5a policy-evaluation description; bump `version` → 0.3.1. |
| 0.3.0   | 2026-07-13 | L3 policy foundation complete (T01–T04): add `execution-risk.schema.json` (machine-readable policy twin, Safety=0→REFUSE + missing→ASK hard rules), `domain-allowlist.md` (8 coordinated skills + per-skill symptom whitelist + destructive exclusions + expansion policy), and `ve-skill-generator/references/leaf-op-metadata-spec.md` (per-op `safety_class`/`blast_radius` contract — 8 leaf SKILL.md operation tables now annotated). |
| 0.2.0   | 2026-07-13 | Add L3 execution-risk policy (`references/policies/execution-risk.md`); `## References` now links the graded AUTO/ASK/REFUSE policy replacing the L2 blanket `{{user.confirm}}` gate. |
| 0.1.0   | 2026-07-10 | Initial skeleton: 7-step loop, 7-dim rubric, 5 prompt templates, Skill-Routing-Graph integration, Reflexion write-back into `docs/failure-patterns.md` (`## Extracted from GCL Traces` block). |

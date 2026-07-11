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
  (`scripts/gcl_runner.py`) available, and a Harness AI Agent runtime
  compatible with the existing `ve-skill-generator` flow.
metadata:
  author: ve-skills
  version: "0.1.0"
  last_updated: "2026-07-10"
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
| `{{user.confirm}}` | Yes/No to destructive op | MUST be collected before Safety=1 floor |
| `{{user.scope_choice}}` | When dispatch_plan lists multiple resource scopes | Collect before dispatch |
| `{{user.ticket_id}}` | JIRA / DOPS / CMS alarm ID | Echo into trace |
| `{{input.incident_payload}}` | From triggering channel (alarm JSON, ticket, chat) | Required entrypoint |
| `{{output.triage_class}}` | routing-graph primary skill | Internal only |
| `{{output.dispatch_plan}}` | JSON {primary, secondary, blast_radius} | Internal only |
| `{{output.critic_verdict}}` | GCL Critic JSON | Persist into trace |
| `{{output.failure_pattern}}` | One-line token if any iteration fails | Auto-aggregated by GCL write-back into `docs/failure-patterns.md` `## Extracted from GCL Traces` (dedup by skill+pattern) |

> **Security**: this skill never sees `VOLCENGINE_SECRET_KEY`. It only authenticates via env and delegates. No credential is ever written to a trace or report.

## Loop Flow

The loop runs in **7 steps** with mandatory pre/post conditions. Each step emits `{{output.*}}` for the next.

### Step 1 — `{{input.incident_payload}}` ingestion

- Source: CMS alarm webhook / JIRA DOPS create / chat trigger / patrol milestone.
- Normalize into `{ticket_id, severity, product_hint, raw_event, observed_at}`.
- Fail-fast on missing `product_hint` → ask user once.

### Step 2 — Triage (load routing graph)

- Read `docs/skill-routing-graph.md` (lazy, ~100 lines).
- Match `product_hint + symptom` against the alarm pattern table.
- Emit `{{output.triage_class}}` = `{primary, secondary[], confidence}`.
- Unknown pattern → route to `ve-cms-ops` first (Rule 5).

### Step 3 — Diagnose (parallel read-only `Describe*` calls)

- Dispatch to `triage_class.primary` with `read-only` mode.
- Cap at `min(15, alarm_count × 3)` read calls; log skipped calls.
- Emit `{{output.diagnosis_evidence}}` = raw observation snippets.
- Hit `cross_skill` boundary → delegate (do not absorb).

### Step 4 — Propose fix + GCL loop

- Build `dispatch_plan` = `{operations[], blast_radius, pre_state_snapshot, rollback_plan}`.
- Run GCL via `scripts/gcl_runner.py` with this skill's `references/rubric.md`.
- Safety must equal 1.0 on every destructive operation.
- `max_iter = 3` for repair loop; `max_iter = 2` for any destructive.

### Step 5 — Confirm (user gate, mandatory on destructive)

- Collect `{{user.confirm}}` for every operation whose safety_class == destructive.
- Silent default = **REFUSE**. No implicit `--force` allowed.
- Trace records the user response verbatim.

### Step 6 — Execute (delegated, not direct)

- Hand `dispatch_plan` to the matched `ve-*-ops` skill; it runs the actual `ve <svc> <Action>` calls.
- This skill monitors, retries with backoff (max 2), and detects partial rollback scenarios.
- Re-emit trace on every retry.

### Step 7 — Reflexion (automatic write-back)

- Failure patterns are **not** hand-written by this skill. When the GCL loop ends (MAX_ITER / SAFETY_FAIL), `scripts/gcl_runner.py` calls `_writeback_failure_pattern` → `gcl_trace_aggregate.update_failure_patterns_file`, which appends/merges the pattern into `docs/failure-patterns.md` `## Extracted from GCL Traces (auto-generated)` block (fields `skill` / `pattern` / `category` / `source`), **dedup by `(skill, pattern)`**.
- The separate `## 6. Incident Response Failures` table is a manually-seeded, design-placeholder section (dedup by `(scenario, failure_pattern)`); it is not produced by the write-back.
- At session end, also update this skill's own working memory (`memory/working.json`) with `last_run`, `triage_class`, `success`.

> **Termination**: same as GCL spec §5 — PASS / MAX_ITER / SAFETY_FAIL. SAFETY_FAIL aborts regardless of progress.

## Quality Gate (GCL)

This skill runs the standard 5-dimension rubric (defined in `references/rubric.md`) plus 2 orchestration-specific dimensions (Reflexion integration + Cross-skill delegation). `max_iter = 3` for read-only paths, `max_iter = 2` for any iteration that touches a destructive leaf skill.

- Safety = 0 → ABORT immediately; never return partial.
- Spec Compliance requires the loop to actually read `docs/skill-routing-graph.md` (not skipped).
- Reflexion dimension tracks whether `failure_pattern` was persisted when one exists.

See `references/rubric.md` for the per-dimension rules.

## Operational Best Practices

- **Trace every iteration** — even when the loop exits on PASS, persist the trace to `audit-results/incident-trace-<ticket_id>-<ISO>.json`. Trace MUST include `RequestId`s from every `ve` call.
- **Bounded Reflexion** — `docs/failure-patterns.md` ≤ 200 lines. Prune patterns with `count < 3`. Promote patterns with `count ≥ 10` to Anti-Patterns in `ve-*-ops` skills.
- **No autopilot for destructive ops** — every destructive step has a `{{user.confirm}}` gate. No exceptions.
- **No credential in trace** — `<masked>` only, `redaction_pass: true`.
- **First incident is the slowest** — expect the GCL loop to fail once before the rubric converges. The second incident on the same symptom should be faster if Reflexion worked.

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

## What This Skill Does NOT Do

- Run a `ve <service> <Action>` call directly (always delegate)
- Author or modify other skills (use `ve-skill-generator`)
- Bypass destructive-op confirmation even if user insists in chat (state limitation)
- Operate outside Volcengine

## Next Steps

- Read `references/rubric.md` — the 7-dimension GCL rubric for this skill
- Read `references/prompt-templates.md` — Generator / Critic / Safety / Trace / Cross-skill prompt skeletons
- Read `docs/skill-routing-graph.md` — the routing table this skill consumes
- Read `docs/failure-patterns.md` `## Extracted from GCL Traces` (auto-generated) and `## 6. Incident Response Failures` (seeded) — where this skill's learning lives

## Changelog

| Version | Date | Change |
|---------|------|--------|
| 0.1.0   | 2026-07-10 | Initial skeleton: 7-step loop, 7-dim rubric, 5 prompt templates, Skill-Routing-Graph integration, Reflexion write-back into `docs/failure-patterns.md` (`## Extracted from GCL Traces` block). |

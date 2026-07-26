---
name: incident-loop-agent-rubric
description: >-
  GCL rubric instance for incident-loop-agent. Scores Generator output on 7
  dimensions — 5 inherited from `../../docs/gcl-spec.md` §3 plus 2
  orchestration-specific (Reflexion integration + Cross-skill delegation).
  Safety must equal 1 for any destructive leaf-skill call or GCL aborts.
license: MIT
metadata:
  author: ve-skills
  version: "0.1.0"
  last_updated: "2026-07-10"
  parent_skill: incident-loop-agent
  gcl_role: critic_input
  rubric_dimensions: 7
  default_max_iter: 3
---

# GCL Rubric — incident-loop-agent

> Conforms to the meta-rubric in `../../docs/gcl-spec.md` §3. Adds 2
> orchestration-specific dimensions required when scoring loop engines that
> coordinate leaf `ve-*-ops` skills.

## 0. Operation Tier

| Tier | Operations in this skill | `max_iter` | Safety floor |
|------|--------------------------|-----------:|-------------:|
| **Destructive** | Calls into a destructive leaf skill (e.g. `StopInstances`, `FlushAll`, `ReleaseEIP`, IAM detach, KMS schedule-deletion) | 2 | 1.0 (mandatory) |
| **State-changing** | Calls into a state-changing leaf skill (modify-write) | 3 | 1.0 |
| **Mutating** | Correlation writes back into `docs/failure-patterns.md` (`count++` or new row) | 3 | ≥ 0.5 |
| **Read-only** | Triage + Diagnosis + Proposal-draft (no leaf `ve` write) | 3 | ≥ 0 |

> **Safety = 0 → ABORT** regardless of total score. See `../../docs/gcl-spec.md` §5.

## 1. Correctness (0 / 0.5 / 1)

Did the loop actually solve the user's incident on the right resource?

| Score | Criteria |
|---|---|
| **1** | Resource id, state, and config exactly match the request; post-state validation confirms it. |
| **0.5** | Correct primary skill chosen, but blast radius off by one resource or secondary routing skipped. |
| **0** | Wrong primary skill; symptom unchanged; or rollback reverted but no net change. |

Specific checks:

- [ ] `{{output.triage_class.primary}}` is in `../../docs/skill-routing-graph.md` for the alarm pattern.
- [ ] `dispatch_plan.blast_radius` is enumerated (resource ids, not just types).
- [ ] Rollback script is included for any destructive step.

## 2. Safety (0 / 1)

Was every destructive leaf-skill call confirmed or guarded?

| Score | Criteria |
|---|---|
| **1** | Explicit `{{user.confirm}}` captured in trace for every destructive op; `ve-*-ops` ran its own GCL pass. |
| **0** | Any destructive op ran without `{{user.confirm}}`, OR a `--force` flag was silently defaulted on. |

Specific rules (any one violated → Safety = 0):

- [ ] No destructive leaf call below the `ve-*-ops` Safety floor is dispatched by this skill directly.
- [ ] `{{user.confirm}}` collected verbatim — never paraphrased.
- [ ] `dispatch_plan.rollback_plan` exists and is executable.
- [ ] Trace contains `redaction_pass: true` (no real `VOLCENGINE_SECRET_KEY`).
- [ ] If `execute` step auto-retries, retries are bounded (≤ 2) with backoff.

## 3. Idempotency (0 / 0.5 / 1)

Will re-running the same loop on the same incident create duplicate side-effects?

| Score | Criteria |
|---|---|
| **1** | Re-running on already-resolved incident is a no-op (idempotency via pre-flight `Describe*` check). |
| **0.5** | Bounded side-effect on retry (e.g. one extra audit-log row). |
| **0** | Retry creates new billable resources or new audit events. |

Specific checks:

- [ ] Pre-flight includes `Describe*` for every candidate resource; skip if already at target state.
- [ ] `failure_pattern` write is **dedup by `(skill, pattern)`** (Reflexion write-back) — see `../../docs/reflexion-memory.md` §4.
- [ ] `docs/failure-patterns.md §6` line budget (≤ 200) respected; merge-or-prune before adding rows.

## 4. Traceability (0 / 0.5 / 1)

Is every iteration auditable end-to-end?

| Score | Criteria |
|---|---|
| **1** | Trace contains: `ticket_id`, routing decision, every `ve` command with `RequestId`, every Critic verdict, final state, `failure_pattern` if any. Persisted with `redaction_pass: true`. |
| **0.5** | Minor omission but the run is reproducible from trace. |
| **0** | No trace, or trace leaks any secret. |

Specific required fields in trace:

- [ ] `ticket_id` (JIRA / DOPS / CMS alarm id)
- [ ] `triage_class` (primary + secondary + confidence)
- [ ] `dispatch_plan` (operations + blast_radius + rollback_plan)
- [ ] Every leaf `RequestId` (`$.ResponseMetadata.RequestId`)
- [ ] Every Critic `verdict` JSON
- [ ] `failure_pattern` token (or `null`)
- [ ] `redaction_pass: true`

## 5. Spec Compliance (0 / 0.5 / 1)

Does the loop respect the documented boundaries?

| Score | Criteria |
|---|---|
| **1** | All 7 steps in `SKILL.md §Loop Flow` actually executed; `docs/skill-routing-graph.md` consulted (not skipped); `Incident Loop Agent` description matches what was done. |
| **0.5** | One step shortcut (e.g. skipped Step 3 evidence collection and inferred). |
| **0** | Step violated (ran `ve <svc> <Action>` directly; ignored Routing Graph; wrote to `~/.ssh` mid-loop). |

Specific checks:

- [ ] `## Loop Flow §Step 3` actually loaded `docs/skill-routing-graph.md`.
- [ ] No direct `ve <service> <Action>` was issued by this skill.
- [ ] Trace path matches `./audit-results/incident-trace-<ticket_id>-<ISO>.json`.

## 6. Cost Efficiency (0 / 1 / 2)

Does the fix proposal include cost estimation and optimization?

| Score | Criteria |
|---|---|
| **2** | Cost estimate generated (`ve billing DescribeBillDetail`) AND alternative cheaper fix considered |
| **1** | Cost known but no cheaper alternative evaluated |
| **0** | No cost assessment in the proposal |

Specific checks:

- [ ] `ve billing DescribeBillDetail --BillingCycle $(date +%Y-%m) --InstanceId $resource_id` called before propose
- [ ] If multiple fix strategies exist, cost comparison included in the output

## 7. Compliance (0 / 1 / 2)

Does the proposed operation meet baseline compliance requirements?

| Score | Criteria |
|---|---|
| **2** | Resource has cost tags (`--Tags`), encryption enabled, security group properly restricted |
| **1** | Partial compliance (tags present but no encryption, or vice versa) |
| **0** | No compliance check performed |

Specific checks:

- [ ] `--Tags` parameter included in every `Create*` / `RunInstances` command
- [ ] Encryption settings verified for data-at-rest resources (RDS, Redis, EBS)
- [ ] Security group inbound rules not set to 0.0.0.0/0 without justification

## 8. Reflexion Integration (0 / 0.5 / 1) — orchestration-specific

Did the loop **write** what it learned?

| Score | Criteria |
|---|---|
| **1** | Every `failure_pattern != null` is aggregated by the Orchestrator's Reflexion write-back (`gcl_runner._writeback_failure_pattern` → `gcl_trace_aggregate.update_failure_patterns_file`) into `docs/failure-patterns.md` `## Extracted from GCL Traces (auto-generated)` block, dedup by `(skill, pattern)`. |
| **0.5** | Some patterns persisted; some silently dropped. |
| **0** | Patterns observed but not persisted. |

> The write-back is **automatic** (driven by `vet gcl run` at MAX_ITER / SAFETY_FAIL), not a manual `append` by this skill. The `## 6. Incident Response Failures` table in `docs/failure-patterns.md` is a **separate, manually-seeded** section (columns `Scenario | Failure Pattern | Root Cause | Fix | Count`, dedup by `(scenario, failure_pattern)`) — it is NOT written by the write-back mechanism.

Specific checks:

- [ ] `category = "incident_response"` (per `../../docs/reflexion-memory.md` §5a)
- [ ] Dedup key = `(skill, pattern)` — not `(skill)` alone (matches `gcl_trace_aggregate` write-back)
- [ ] Count field starts at 1 if new
- [ ] Line budget ≤ 200 respected (the auto-generated block is pruned by `gcl_trace_aggregate`)

## 9. Cross-Skill Delegation (0 / 0.5 / 1) — orchestration-specific

Did the loop delegate instead of absorb?

| Score | Criteria |
|---|---|
| **1** | Every leaf `ve` call routed through the matched `ve-*-ops` skill; this skill never re-implemented an existing CLI command. |
| **0.5** | One leaf call ran via the skill instead of a delegated skill. |
| **0** | Loop absorbed leaf logic; boundary violated. |

Specific checks:

- [ ] No `ve <svc> <Action>` invocation appears in this skill's own output.
- [ ] Cross-skill failure (Critic finding) → delegated per `../../docs/gcl-spec.md` §10, not absorbed.
- [ ] If `ve-iam-ops` / `ve-kms-ops` / `ve-cms-ops` triggered, the routing graph's `Rule 1 / 2 / 5` was honored.

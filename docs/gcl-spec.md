# Generator-Critic-Loop (GCL) — Implementation Spec

> 本文档是 `AGENTS.md` §GCL 的完整实现规范。AGENTS.md 只保留摘要和入口。

---

## 1. Purpose

Apply an adversarial **Generator ↔ Critic** loop with a quantitative rubric to every skill execution.
Most valuable in **high-side-effect cloud operations** (delete, stop, restore, IAM/KMS/DDL) where a single
mistake is unrecoverable.

| GAN (real) | GCL (this spec) |
|---|---|
| Discriminator learns sample distribution | Critic scores an **explicit rubric** |
| No termination condition | Must terminate: **PASS / MAX_ITER / SAFETY_FAIL** |
| G and D train in parallel | G and C run **sequentially** |
| Goal: "fool the D" | Goal: "pass the rubric threshold" |

## 2. Roles

| Role | Job | Input | Output | Forbidden |
|---|---|---|---|---|
| **Generator (G)** | Execute the cloud operation | user request + previous Critic feedback | result + execution trace | modifying the rubric; self-scoring |
| **Critic (C)** | Independently audit G's output | G's result + trace + rubric | scores + suggestions | calling `ve` / SDK / mutating anything |
| **Orchestrator (O)** | Loop control, termination, final return | context + C scores + budget | continue / final result | executing or scoring on its own |

**Hard constraint:** G and C MUST live in **isolated prompt contexts** (preferably isolated sessions
or sub-agents). A shared context is a "pseudo-GCL" and is explicitly banned — see §9.

## 3. Rubric (mandatory per skill)

Each `SKILL.md` MUST declare its skill-specific rubric. Minimum 5 dimensions:

| Dimension | Meaning | Scale | Default threshold |
|---|---|---|---|
| **Correctness** | Resource id / state / config actually matches the request | 0 / 0.5 / 1 | ≥ 0.5 (1.0 required for `delete` / `stop` / IAM / KMS / DDL) |
| **Safety** | Destructive op (`delete` / `stop` / `restore` / IAM / KMS / DDL) was confirmed or guarded | 0 / 1 | = 1 |
| **Idempotency** | Retrying the same call will not cause duplicate side-effects | 0 / 0.5 / 1 | ≥ 0.5 |
| **Traceability** | Output is auditable: command, params, raw response, errors all captured | 0 / 0.5 / 1 | ≥ 0.5 |
| **Spec Compliance** | Conforms to the skill's `references/core-concepts.md` and the **Five Core Standards** (boundaries, I/O, steps, failures, single responsibility) | 0 / 0.5 / 1 | ≥ 0.5 |

**Safety = 0 → ABORT immediately, regardless of total score.**

## 4. Loop Flow

```
User Request
     │
     ▼
[0] Pre-flight (Orchestrator)
    - resolve {{env.*}} and {{user.*}} variables
    - pick skill, load its rubric
    - derive sanitized operation_intent (operation, expected_state, resource_scope, safety_class; no raw user wording or credentials)
     │
     ▼
[1] Generate (G) ───────────────────────┐
    - run ve <svc> <Action> --<Param>   │  (CLI primary, see CLAUDE.md)
    - OR JIT Go SDK script              │  (fallback)
    - capture trace                     │
     │                                  │
     ▼                                  │
[2] Critique (C)                       │
    - isolated prompt context           │
    - score every rubric dimension      │
    - emit actionable suggestions       │
     │                                  │
     ▼                                  │
[3] Decide (Orchestrator)              │
    - Safety=0  → ABORT (no partial)   │
    - all pass  → RETURN                │
    - else & iter<max → inject         │
       suggestions into G               │
    - else → RETURN best + unresolved   │
       rubric items                     │
     └──────────────────────────────────┘
```

The Orchestrator owns `operation_intent` generation during Pre-flight. It MUST derive this sanitized object before Critic scoring; the object may include `operation`, `expected_state`, `resource_scope`, and `safety_class`, but MUST NOT include raw user wording, credentials, or unmasked sensitive identifiers. The Critic receives only this sanitized `operation_intent` (never the raw user request) to prevent "answer-aligned" rubber-stamping.

## 5. Termination (first match wins)

| Condition | Behavior |
|---|---|
| **PASS** | Every rubric dimension meets its threshold → return G's result |
| **MAX_ITER** | Reached `max_iterations` (default 3, 2 for destructive skills) → return **best-so-far** + unresolved rubric items |
| **SAFETY_FAIL** | Safety = 0 → **ABORT**; never return partial or "best-effort" output |

`max_iterations` defaults per skill class — see §8.

## 6. Trace & Audit (mandatory)

Every GCL run MUST persist a JSON trace:

```json
{
  "trace_schema_version": "v1",
  "skill": "ve-ecs-ops",
  "request": "<sanitized user request>",
  "operation_intent": {
    "operation": "delete_instance",
    "resource_scope": ["i-***"],
    "expected_state": "DELETED",
    "safety_class": "destructive"
  },
  "rubric_version": "v1",
  "masked_fields": ["request", "operation_intent.resource_scope"],
  "redaction_pass": true,
  "iterations": [
    {
      "iter": 1,
      "generator": { "command": "ve ecs DeleteInstance --InstanceId i-xxx", "args": {"InstanceId": "i-xxx"}, "exit_code": 0, "result_excerpt": "..." },
      "critic": {
        "scores": {
          "correctness": 1, "safety": 1, "idempotency": 0.5,
          "traceability": 1, "spec_compliance": 1
        },
        "suggestions": ["..."],
        "blocking": false
      },
      "decision": "RETRY"
    }
  ],
  "final": {
    "status": "PASS",
    "iter": 2,
    "output": "...",
    "failure_pattern": null
  }
}
```

The `final.failure_pattern` field holds a short string (e.g. `"missing_iam_policy"`, `"timeout_on_delete"`) when the run completed with issues, or `null` when the run succeeded without incident. This field feeds into the Reflexion Integration mechanism (see SS12).

Path: `./audit-results/gcl-trace-YYYYMMDD-HHMMSS.json` — create the `audit-results/`
directory if absent. **Never** commit trace content; add `audit-results/` to
`.gitignore`.

**Credential handling in traces:** the trace must NEVER contain a real
`VOLCENGINE_ACCESS_KEY` / `VOLCENGINE_SECRET_KEY`. Only `<masked>` or `sha256:<hash-prefix>`
for existence checks. Add a `redaction_pass: true` boolean to the trace root.

## 7. Prompt Templates (mandatory per skill)

Each skill's `references/prompt-templates.md` MUST contain:

1. **Generator Prompt Template** — placeholders: `{{user.request}}`, `{{output.critic_feedback}}`, `{{output.rubric}}`
2. **Critic Prompt Template** — placeholders: `{{output.operation_intent}}`, `{{output.generator_output}}`, `{{output.trace}}`, `{{output.rubric}}`

> **Placeholder syntax** MUST follow the repository-wide convention
> (see `CLAUDE.md` **Placeholder System**): `{{env.*}}` / `{{user.*}}` / `{{output.*}}`.
> Bare `{...}` placeholders are NOT allowed in skill prompt templates.

**Critic prompt must hide the raw user request** to prevent "answer-aligned" rubber-stamping. It may use the sanitized `{{output.operation_intent}}` derived by the Orchestrator.
Recommended skeleton:

```text
You are an independent Volcengine cloud-operation auditor.
You will see one execution result and its trace. Score it STRICTLY against the rubric below.
Do NOT consider the original user request — judge only what was actually done.

rubric: {{output.rubric}}
operation_intent: {{output.operation_intent}}
generator_output: {{output.generator_output}}
trace: {{output.trace}}

Return strict JSON:
{
  "scores": { "correctness": 0|0.5|1, "safety": 0|0.5|1, "idempotency": 0|0.5|1,
              "traceability": 0|0.5|1, "spec_compliance": 0|0.5|1 },
  "suggestions": ["≤ 3 concrete, executable improvements"],
  "blocking": true|false
}
```

## 8. Per-Skill Defaults

| Skill | GCL | Default max_iter | Notes |
|---|---|---|---|
| `ve-ecs-ops` | **required** | 2 | instance `Delete` / `Stop` are destructive |
| `ve-redis-ops` | **required** | 2 | `FlushAll` / instance delete |
| `ve-rds-mysql-ops` | **required** | 2 | DDL / DELETE / TRUNCATE |
| `ve-rds-ops` | **required** | 2 | DDL / DELETE / TRUNCATE (RDS MySQL product-skill variant) |
| `ve-rds-pg-ops` | **required** | 2 | DDL / DELETE / TRUNCATE |
| `ve-polar-mysql-ops` | **required** | 2 | DDL / DELETE / TRUNCATE |
| `ve-mongodb-ops` | **required** | 2 | `dropDatabase` / delete |
| `ve-elasticsearch-ops` | **required** | 2 | delete index / cluster |
| `ve-tos-ops` | **required** | 2 | bucket delete + object lifecycle |
| `ve-iam-ops` | **required** | 2 | detach policy / delete role / rotate keys |
| `ve-kms-ops` | **required** | 2 | schedule key deletion is irreversible |
| `ve-eip-ops` | **required** | 2 | release EIP can break production |
| `ve-security-group-ops` | **required** | 2 | revoke-rule can lock out production |
| `ve-vpc-ops` | recommended | 3 | vpc / subnet delete |
| `ve-nat-ops` | recommended | 3 | SNAT / DNAT rule delete |
| `ve-vpn-ops` | recommended | 3 | tunnel / customer-gateway delete |
| `ve-clb-ops` | recommended | 3 | listener / backend delete |
| `ve-alb-ops` | recommended | 3 | listener / server group delete |
| `ve-vke-ops` | recommended | 3 | node / cluster delete |
| `ve-nas-ops` | recommended | 3 | filesystem / mount delete |
| `ve-cms-ops` | recommended | 3 | alarm rule delete |
| `ve-fg-ops` | recommended | 3 | function delete |
| `ve-ark-ops` | recommended | 3 | instance / template delete |
| `ve-cdn-ops` | optional | 5 | domain config / refresh |
| `ve-dns-ops` | optional | 5 | record delete |
| `ve-kafka-ops` | optional | 5 | topic delete |
| `ve-sls-ops` | optional | 5 | read-mostly |
| `ve-billing-ops` | optional | 5 | read-only |
| `ve-skill-generator` | optional | 3 | meta operation (governs other skills) |

Each skill may override `max_iter` in its own `SKILL.md` (under `## Quality Gate (GCL)`).

## 9. Anti-Patterns (banned)

- ❌ **Shared context G+C** — defeats independence → banned
- ❌ **Subjective scoring** — Critic must use the rubric, not "vibes" → banned
- ❌ **Unbounded loop** — always hard-cap iterations → banned
- ❌ **Critic sees the user request** — encourages rubber-stamping → banned
- ❌ **Silently downgrade on Safety fail** — must ABORT visibly → banned
- ❌ **Trace not persisted** — no post-mortem possible → banned
- ❌ **Critic mutates resources** — Critic is read-only by definition → banned
- ❌ **Real `VOLCENGINE_SECRET_KEY` in trace** — credential leakage → banned (use `<masked>` only)
- ❌ **GCL bypass for "obviously safe" ops** — even reads go through GCL; only `max_iter` and "required" tier change → banned

## 10. Cross-Skill Delegation

When GCL identifies cross-product gaps, the Orchestrator MUST delegate, not absorb:

| Critic finding | Delegate to |
|---|---|
| IAM policy gap (no permission for `ve <svc> <Action>`) | `ve-iam-ops` |
| KMS key / secret needed for the operation | `ve-kms-ops` |
| EIP / VPC network concern raised in a non-network skill | `ve-eip-ops` / `ve-vpc-ops` |
| Monitoring / alarm rule change needed after a destructive op | `ve-cms-ops` |
| Billing quota exceeded during the operation | `ve-billing-ops` |
| Audit / trace review needed for the run | (future) ve-audit-ops |

The Critic itself MUST NOT call any of the above — it only emits suggestions. The
Generator is the executor on the next iteration.

## 11. Rollout Roadmap

| Phase | Status | Primary artifact | Gate |
|---|---|---|---|
| 1 | Done | `ve-ecs-ops/references/{rubric.md,prompt-templates.md}` | ECS pilot for destructive operations |
| 2 | Done | `scripts/gcl_runner.py` | External Critic via `--critic-json`/stdin |
| 3 | Done | `scripts/gcl_trace_aggregate.py` + quality summary | Quality summary feeds monitoring / inspection |
| 4 | Done | Full rollout: all 29 skills GCL-equipped | Complete coverage across all product skills |
| 4.1 | In Progress | `scripts/check_gcl_conformance.py` | CI gate for Tier-A conformance |

Detailed phase changes live in the changelog below.

## 12. Reflexion Integration

GCL traces include a `failure_pattern` field in the `final` object (see SS6). This enables a lightweight cross-session failure-pattern learning mechanism:

- **Storage**: After each GCL run, extract `failure_pattern` (when non-null) and append it to `docs/failure-patterns.md` as a bounded memory store.
- **Retrieval**: During Pre-flight (SS4), the Orchestrator MAY query `docs/failure-patterns.md` for patterns matching the current `operation_intent.safety_class` and `operation_intent.operation`. Known failure patterns are injected into the Generator prompt as additional guardrails.
- **Scope**: This is a **bounded memory** mechanism -- only failure patterns (not full traces) are persisted across sessions. Full traces remain in `audit-results/` and MUST NOT be committed to the repository.
- **Trace schema impact**: The `failure_pattern` field at `final.failure_pattern` holds a short string (e.g. `"missing_iam_policy"`, `"timeout_on_delete"`) or `null` when the run succeeded without incident.

## 13. Changelog

| Version | Date | Change |
|---|---|---|
| 1.18.0 | 2026-06-19 | **Phase 4.1 (in progress):** `scripts/check_gcl_conformance.py` -- CI gate for Tier-A conformance across all 29 skills; spec enhanced with `operation_intent`, enhanced trace schema, Reflexion Integration, Rollout Roadmap, and See also sections |
| 1.17.0 | 2026-06-19 | ve-rds-ops (RDS MySQL variant) GCL rollout; full coverage of all 29 skills |
| 1.16.0 | 2026-06-04 | ve-skill-generator meta-skill GCL rollout |
| 1.15.0 | 2026-06-04 | optional tier rollout (max_iter=5): cdn, dns, kafka, sls, billing |
| 1.14.0 | 2026-06-04 | recommended tier mass rollout (max_iter=3): vpc, nat, vpn, clb, alb, vke, nas, cms, fg, ark |
| 1.13.0 | 2026-06-04 | ve-security-group-ops rollout |
| 1.12.0 | 2026-06-04 | ve-elasticsearch-ops rollout |
| 1.11.0 | 2026-06-04 | ve-mongodb-ops rollout |
| 1.10.0 | 2026-06-04 | ve-polar-mysql-ops rollout |
| 1.9.0 | 2026-06-04 | ve-rds-pg-ops rollout |
| 1.8.0 | 2026-06-04 | ve-tos-ops rollout |
| 1.7.0 | 2026-06-04 | ve-rds-mysql-ops rollout |
| 1.6.0 | 2026-06-04 | ve-eip-ops rollout |
| 1.5.0 | 2026-06-04 | ve-redis-ops rollout |
| 1.4.0 | 2026-06-04 | ve-skill-template.md updated with GCL block |
| 1.3.0 | 2026-06-04 | ve-kms-ops rollout |
| 1.2.0 | 2026-06-04 | ve-iam-ops rollout |
| 1.1.0 | 2026-06-04 | Phase 1 rollout: audit-results/.gitignore, ve-ecs-ops pilot, 4-tier operation classification |
| 1.0.0 | 2026-06-04 | Initial GCL specification added to Volcengine `AGENTS.md` (adapted from JD Cloud GCL spec) |

## 14. See also

- Each skill's `references/rubric.md` -- the rubric instance
- Each skill's `references/prompt-templates.md` -- the G/C prompt skeletons
- `ve-skill-generator/references/governance-and-adversarial-review.md` -- build-time R1-R4 review (sister gate)
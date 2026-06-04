@CLAUDE.md

# Agent-Specific Rules

These supplement `CLAUDE.md` (imported above). They encode lessons that have already cost agent sessions in this repo.

## Repository Type — Read This First

This repo contains **only Markdown skill specifications**. There is:

- **No build system, no tests, no lint, no package manager.** Do not run `go build`, `go test`, `npm`, `pip`, `make`, etc. — they will fail or do nothing.
- **No runtime code.** The `ve` CLI and Go SDK referenced in skills are *consumed by* generated skills at execution time, not built here.
- **No CI workflows yet.** Verification is human review against the P0/P1 checklist in `ve-skill-generator/SKILL.md`.

Verification for skill edits = re-reading the file plus walking the P0/P1 checklist. Do not invent build/test commands.

## MANDATORY: Two-Round Self-Review After Every Skill Update

After ANY edit to a `ve-*-ops/SKILL.md`, its `references/*.md`, `assets/*`, or to `ve-skill-generator/**`, you MUST run **two rounds of self-review** before declaring done. Fix every issue surfaced — do not just list them.

### Round 1 — Structural & Specification Compliance

Verify the file against `ve-skill-generator/SKILL.md` and `ve-skill-generator/references/ve-skill-template.md`:

- [ ] Frontmatter matches agentskills.io OpenSpec (name, description, license, compatibility, metadata)
- [ ] `cli_applicability` is set; if `dual-path`, BOTH `ve` CLI step and SDK step exist for every operation
- [ ] Five Core Standards section present and accurate (boundaries, I/O, steps, failures, single responsibility)
- [ ] Placeholders use the right channel: `{{env.*}}` for secrets/region, `{{user.*}}` for interactive, `{{output.*}}` for captured
- [ ] No credential value is logged, printed, or echoed anywhere in examples — only `<masked>` or existence checks
- [ ] Every destructive operation (delete, stop, release, restore) has an explicit safety gate / confirmation
- [ ] Error taxonomy has ≥ 10 product-specific codes with HALT vs retry classification
- [ ] All 6 standard reference files exist when applicable: `core-concepts`, `api-sdk-usage`, `cli-usage`, `troubleshooting`, `monitoring`, `integration`
- [ ] Cross-product operations declare delegation targets (e.g. ECS skill delegates IAM work to IAM skill)
- [ ] `last_updated` and `version` bumped if behavior changed

### Round 2 — Accuracy & Anti-Pattern Sweep

Verify against reality, not memory:

- [ ] Every API field name and CLI flag is grounded in official OpenAPI or verified via `ve <service> --help` — no guessed names
- [ ] Command prefix is `ve` (not `volcengine-cli`, which was renamed at v1.0.20)
- [ ] CLI command shape matches verified pattern: `ve <service> <Action> --<Param> value` (PascalCase actions and params)
- [ ] JSON bodies use `--body '{"Key":"Value"}'` form, not invented flags
- [ ] No generic prose, no speculative claims, no copy-pasted boilerplate from other products that doesn't apply
- [ ] Links to sibling references use correct relative paths
- [ ] Examples are minimal and runnable, not pseudocode
- [ ] Markdown renders cleanly (tables aligned, code fences closed, headings hierarchical)

**Fix every issue found in both rounds before responding "done".** If a finding requires information you don't have (e.g. an unverified API field), mark it `[blocked: needs OpenAPI verification]` rather than guessing.

## Skill Authoring Guardrails

- **Never create a new `ve-*-ops/` directory by hand.** Use the `ve-skill-generator` meta-skill flow so layout, frontmatter, and checklists stay consistent.
- **Never invent CLI flags or API parameters.** If a field isn't in the official OpenAPI doc or `ve <service> <action> --help`, do not write it. Mark it as needing verification instead.
- **Single-product rule.** One `ve-*-ops` skill = one product = one primary resource model. Cross-product work is delegated, not absorbed.
- **Prefer editing existing skills over creating new files.** Do not add tutorial-style `.md` files at the repo root.

## File Layout Anchors (do not relocate without reason)

```
ve-skill-generator/                  meta-skill: how to author new skills
  SKILL.md                           generator workflow + P0/P1 checklist
  references/
    ve-skill-template.md             canonical skill template
    cli-behavior.md                  verified `ve` CLI conventions
    execution-environment.md         CLI + Go SDK setup
    user-experience-spec.md          UX requirements every generated skill must follow
    governance-and-adversarial-review.md
ve-[product]-ops/                    one per Volcengine product
  SKILL.md                           main runbook
  references/                        6 standard reference files
  assets/example-config.yaml
```

`.omc/` and `.omo/` are tool-local state and are gitignored — do not commit anything inside them.

---

## Generator-Critic-Loop (GCL) — Adversarial Quality Gate

> Inspired by GAN's Generator/Discriminator idea, but deliberately **not** a real GAN.
> Naming: **GCL (Generator-Critic-Loop)** to avoid misleading reviewers and LLM trainees.
> Adapted to the Volcengine stack: `ve` CLI (primary) + Go SDK (fallback) on
> `{{env.VOLCENGINE_ACCESS_KEY}}` / `{{env.VOLCENGINE_SECRET_KEY}}` / `{{env.VOLCENGINE_REGION}}`.

### 1. Purpose

Apply an adversarial **Generator ↔ Critic** loop with a quantitative rubric to every skill execution.
Most valuable in **high-side-effect cloud operations** (delete, stop, restore, IAM/KMS/DDL) where a single
mistake is unrecoverable.

| GAN (real) | GCL (this spec) |
|---|---|
| Discriminator learns sample distribution | Critic scores an **explicit rubric** |
| No termination condition | Must terminate: **PASS / MAX_ITER / SAFETY_FAIL** |
| G and D train in parallel | G and C run **sequentially** |
| Goal: "fool the D" | Goal: "pass the rubric threshold" |

### 2. Roles

| Role | Job | Input | Output | Forbidden |
|---|---|---|---|---|
| **Generator (G)** | Execute the cloud operation | user request + previous Critic feedback | result + execution trace | modifying the rubric; self-scoring |
| **Critic (C)** | Independently audit G's output | G's result + trace + rubric | scores + suggestions | calling `ve` / SDK / mutating anything |
| **Orchestrator (O)** | Loop control, termination, final return | context + C scores + budget | continue / final result | executing or scoring on its own |

**Hard constraint:** G and C MUST live in **isolated prompt contexts** (preferably isolated sessions
or sub-agents). A shared context is a "pseudo-GCL" and is explicitly banned — see §9.

### 3. Rubric (mandatory per skill)

Each `SKILL.md` MUST declare its skill-specific rubric. Minimum 5 dimensions:

| Dimension | Meaning | Scale | Default threshold |
|---|---|---|---|
| **Correctness** | Resource id / state / config actually matches the request | 0 / 0.5 / 1 | ≥ 0.5 (1.0 required for `delete` / `stop` / IAM / KMS / DDL) |
| **Safety** | Destructive op (`delete` / `stop` / `restore` / IAM / KMS / DDL) was confirmed or guarded | 0 / 1 | = 1 |
| **Idempotency** | Retrying the same call will not cause duplicate side-effects | 0 / 0.5 / 1 | ≥ 0.5 |
| **Traceability** | Output is auditable: command, params, raw response, errors all captured | 0 / 0.5 / 1 | ≥ 0.5 |
| **Spec Compliance** | Conforms to the skill's `references/core-concepts.md` and the **Five Core Standards** (boundaries, I/O, steps, failures, single responsibility) | 0 / 0.5 / 1 | ≥ 0.5 |

**Safety = 0 → ABORT immediately, regardless of total score.**

### 4. Loop Flow

```
User Request
     │
     ▼
[0] Pre-flight (Orchestrator)
    - resolve {{env.*}} and {{user.*}} variables
    - pick skill, load its rubric
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

### 5. Termination (first match wins)

| Condition | Behavior |
|---|---|
| **PASS** | Every rubric dimension meets its threshold → return G's result |
| **MAX_ITER** | Reached `max_iterations` (default 3, 2 for destructive skills) → return **best-so-far** + unresolved rubric items |
| **SAFETY_FAIL** | Safety = 0 → **ABORT**; never return partial or "best-effort" output |

`max_iterations` defaults per skill class — see §8.

### 6. Trace & Audit (mandatory)

Every GCL run MUST persist a JSON trace:

```json
{
  "skill": "ve-ecs-ops",
  "request": "<sanitized user request>",
  "rubric_version": "v1",
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
  "final": { "status": "PASS", "iter": 2, "output": "..." }
}
```

Path: `./audit-results/gcl-trace-YYYYMMDD-HHMMSS.json` — create the `audit-results/`
directory if absent. **Never** commit trace content; add `audit-results/` to
`.gitignore` (current `.gitignore` covers `.omc/`, `.omo/`, `.env` — extend it
to include `audit-results/` as part of Phase 1 rollout on `ve-ecs-ops`).

**Credential handling in traces:** the trace must NEVER contain a real
`VOLCENGINE_ACCESS_KEY` / `VOLCENGINE_SECRET_KEY`. Only `<masked>` or `sha256:<hash-prefix>`
for existence checks. Add a `redaction_pass: true` boolean to the trace root.

### 7. Prompt Templates (mandatory per skill)

Each skill's `references/prompt-templates.md` MUST contain:

1. **Generator Prompt Template** — placeholders: `{{user.request}}`, `{{output.critic_feedback}}`, `{{output.rubric}}`
2. **Critic Prompt Template** — placeholders: `{{output.generator_output}}`, `{{output.trace}}`, `{{output.rubric}}`

> **Placeholder syntax** MUST follow the repository-wide convention
> (see `CLAUDE.md` **Placeholder System**): `{{env.*}}` / `{{user.*}}` / `{{output.*}}`.
> Bare `{...}` placeholders are NOT allowed in skill prompt templates.

**Critic prompt must hide the raw user request** to prevent "answer-aligned" rubber-stamping.
Recommended skeleton:

```text
You are an independent Volcengine cloud-operation auditor.
You will see one execution result and its trace. Score it STRICTLY against the rubric below.
Do NOT consider the original user request — judge only what was actually done.

rubric: {{output.rubric}}
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

### 8. Per-Skill Defaults

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

### 9. Anti-Patterns (banned)

- ❌ **Shared context G+C** — defeats independence → banned
- ❌ **Subjective scoring** — Critic must use the rubric, not "vibes" → banned
- ❌ **Unbounded loop** — always hard-cap iterations → banned
- ❌ **Critic sees the user request** — encourages rubber-stamping → banned
- ❌ **Silently downgrade on Safety fail** — must ABORT visibly → banned
- ❌ **Trace not persisted** — no post-mortem possible → banned
- ❌ **Critic mutates resources** — Critic is read-only by definition → banned
- ❌ **Real `VOLCENGINE_SECRET_KEY` in trace** — credential leakage → banned (use `<masked>` only)
- ❌ **GCL bypass for "obviously safe" ops** — even reads go through GCL; only `max_iter` and "required" tier change → banned

### 10. Rollout Roadmap

- **Phase 1 (this commit)** — add this section to `AGENTS.md`; pilot on **`ve-ecs-ops` only** (most representative
  destructive workload) with its `references/prompt-templates.md` and `references/rubric.md`. `ve-iam-ops` and `ve-kms-ops` follow in the next PRs.
- **Phase 2** — add `scripts/gcl_runner.py` (or `scripts/gcl_runner.go`) as a reusable Orchestrator
- **Phase 3** — feed `gcl-trace-*.json` into a future `ve-audit-ops` skill for quality dashboards
- **Phase 4** — wire rubric pass-rate to Cloud Monitor (CMS) alarms so real incidents refine thresholds

**Phase 1 deliverable status (2026-06-04):**

- [x] `AGENTS.md` §GCL added (this section)
- [x] `audit-results/` added to `.gitignore`
- [x] `ve-skill-generator/references/ve-skill-template.md` updated with GCL block + appendix
- [x] `ve-ecs-ops` GCL rollout (pilot)
- [x] `ve-iam-ops` GCL rollout
- [x] `ve-kms-ops` GCL rollout
- [x] `ve-redis-ops` GCL rollout
- [x] `ve-eip-ops` GCL rollout
- [x] `ve-rds-mysql-ops` GCL rollout
- [x] `ve-tos-ops` GCL rollout
- [x] `ve-rds-pg-ops` GCL rollout
- [x] `ve-polar-mysql-ops` GCL rollout
- [x] `ve-mongodb-ops` GCL rollout
- [x] `ve-elasticsearch-ops` GCL rollout
- [x] `ve-security-group-ops` GCL rollout
- [ ] First real GCL run trace captured (pending a destructive workload in a test environment)

**All 13 `required`-tier skills (per AGENTS.md §8) are now GCL-equipped.**

### 11. Cross-Skill Delegation (extends `CLAUDE.md`)

When GCL identifies cross-product gaps, the Orchestrator MUST delegate, not absorb:

| Critic finding | Delegate to |
|---|---|
| IAM policy gap (no permission for `ve <svc> <Action>`) | `ve-iam-ops` |
| KMS key / secret needed for the operation | `ve-kms-ops` |
| EIP / VPC network concern raised in a non-network skill | `ve-eip-ops` / `ve-vpc-ops` |
| Monitoring / alarm rule change needed after a destructive op | `ve-cms-ops` |
| Billing quota exceeded during the operation | `ve-billing-ops` |
| Audit / trace review needed for the run | (future) `ve-audit-ops` |

The Critic itself MUST NOT call any of the above — it only emits suggestions. The
Generator is the executor on the next iteration.

### 12. Changelog

| Version | Date | Change |
|---|---|---|
| 1.0.0 | 2026-06-04 | Initial GCL specification added to Volcengine `AGENTS.md` (adapted from JD Cloud GCL spec; Correctness threshold relaxed to ≥0.5; pilot scoped to `ve-ecs-ops`; ve-CLI / Go-SDK execution path; `VOLCENGINE_*` credential channels) |
| 1.1.0 | 2026-06-04 | Phase 1 rollout completed: `audit-results/` added to `.gitignore`; `ve-ecs-ops/SKILL.md` gained `## Quality Gate (GCL)` chapter; `ve-ecs-ops/references/rubric.md` and `ve-ecs-ops/references/prompt-templates.md` created; 4-tier operation classification (destructive / state_changing / mutating / read-only); 8 verbatim safety prompts; missing `ve-rds-ops` row added to per-skill defaults table |
| 1.2.0 | 2026-06-04 | `ve-iam-ops` rollout: rubric + prompts cover user/policy/role/group/access-key/STS lifecycle; IAM-specific rules for `AttachPolicy with Action=*:*`, `CreateRole with open trust`, `DeleteUser` dependency check, `CreateAccessKey` secret masking; 10 verbatim safety prompts |
| 1.3.0 | 2026-06-04 | `ve-kms-ops` rollout: rubric + prompts cover key lifecycle + encrypt/decrypt/data-key/policy/grant; KMS-specific rules for `ScheduleKeyDeletion PendingWindowDays ≥ 7`, plaintext/secret masking in trace, `PutKeyPolicy` broad-permission warning; 8 verbatim safety prompts |
| 1.4.0 | 2026-06-04 | `ve-skill-generator/references/ve-skill-template.md` updated with `## Quality Gate (GCL)` block + GCL appendix templates (rubric + prompt-templates); Four Core Standards row 4 updated to mention GCL rubric; newly generated skills inherit GCL by default |
| 1.5.0 | 2026-06-04 | `ve-redis-ops` rollout: rubric + prompts cover Redis instance lifecycle + allowlist/account/backup; Redis-specific rules for `DeleteDBInstance` deletion-protection, `ModifyDBInstanceSpec` downtime, `RestartDBInstance` connection cutoff; 5 verbatim safety prompts |
| 1.6.0 | 2026-06-04 | `ve-eip-ops` rollout: rubric + prompts cover EIP allocate/associate/disassociate/release + bandwidth; EIP-specific rules for `ReleaseEipAddress` irreversibility + status pre-check, `DisassociateEipAddress` production cutoff, `AssociateEipAddress` force-rebind; 4 verbatim safety prompts |
| 1.7.0 | 2026-06-04 | `ve-rds-mysql-ops` rollout: rubric + prompts cover RDS MySQL instance lifecycle + parameter/account/backup/restore; RDS-specific rules for `DeleteDBInstance` deletion-protection, `RebuildDBInstance` data-loss warning, `ModifyDBNodeSpec` downtime, `ModifyDBInstanceParameter` ForceRestart; 6 verbatim safety prompts |
| 1.8.0 | 2026-06-04 | `ve-tos-ops` rollout: rubric + prompts cover TOS bucket/object lifecycle + ACL/lifecycle/versioning/FinOps; TOS-specific rules for `DeleteBucket` emptiness+versioning guard, `DeleteObject` prefix-pattern review, `PutBucketACL` public-access warning, `OptimizeStorageClass` Archive-cost warning; 6 verbatim safety prompts; honors TOS_ACCESS_KEY/TOS_SECRET_KEY env-var convention |
| 1.9.0 | 2026-06-04 | `ve-rds-pg-ops` rollout: rubric + prompts for PG instance lifecycle; DeleteDBInstance deletion-protection guard; ModifyDBInstanceSpec downtime warning |
| 1.10.0 | 2026-06-04 | `ve-polar-mysql-ops` rollout: rubric + prompts for PolarDB cluster lifecycle; DeleteDBCluster cluster+data-loss guard; Failover production-write interruption warning; ScaleStorage irreversible guard |
| 1.11.0 | 2026-06-04 | `ve-mongodb-ops` rollout: rubric + prompts for MongoDB instance lifecycle; DeleteDBInstance data-loss guard; ModifyDBInstanceSpec replica-set election downtime |
| 1.12.0 | 2026-06-04 | `ve-elasticsearch-ops` rollout: rubric + prompts for ES instance/index/snapshot/plugin/Kibana lifecycle; DeleteInstance ALL-data-lost guard; UpgradeVersion no-downgrade warning; rolling-restart per-node interruption |
| 1.13.0 | 2026-06-04 | `ve-security-group-ops` rollout: rubric + prompts for SG rule lifecycle; RevokeSecurityGroupIngress on 0.0.0.0/0 sensitive-port double-confirm; AuthorizeSecurityGroupIngress internet-exposure warning; AuthorizeSecurityGroupEgress data-exfiltration warning |
| 1.14.0 | 2026-06-04 | `recommended` tier mass rollout (max_iter=3): ve-vpc-ops, ve-nat-ops, ve-vpn-ops (SSL cert key masking), ve-clb-ops, ve-alb-ops, ve-vke-ops, ve-nas-ops, ve-cms-ops, ve-fg-ops, ve-ark-ops — all with 4-tier rubrics, G/C/O prompt skeletons, and operation-specific safety prompts |
| 1.15.0 | 2026-06-04 | `optional` tier final rollout (max_iter=5): ve-cdn-ops, ve-dns-ops, ve-kafka-ops, ve-sls-ops, ve-billing-ops — compact rubrics and prompt-templates for low-risk operations |
| 1.16.0 | 2026-06-04 | `ve-skill-generator` meta-skill GCL rollout: rubric + prompt-templates for generation verification (checks generated skill has GCL section, 5-dimension rubric, secret-free examples); version bump to 1.1.0 |
| 1.17.0 | 2026-06-04 | `ve-rds-ops` (RDS MySQL variant) GCL rollout: rubric + prompt-templates for instance lifecycle; completes full coverage of all 29 skills in the repository |

## GCL Rollout Complete — All 29 Skills Equipped

### 13. See also

- `docs/GCL_RETROSPECTIVE.md` — post-rollout retrospective and Phase 3 dashboard design contract (2026-06-04, to be created)
- `ve-skill-generator/SKILL.md` — the meta-skill that scaffolds new skills; new skills MUST include `## Quality Gate (GCL)` section + `references/rubric.md` + `references/prompt-templates.md`
- `ve-skill-generator/references/ve-skill-template.md` — template now has a `## Quality Gate (GCL)` block (added 2026-06-04); newly generated skills inherit GCL by default
- `ve-skill-generator/references/governance-and-adversarial-review.md` — pre-GCL adversarial review
- Each skill's `references/rubric.md` — the rubric instance
- Each skill's `references/prompt-templates.md` — the G/C/O prompt skeletons

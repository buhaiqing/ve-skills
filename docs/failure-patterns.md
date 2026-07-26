# Failure Patterns — Reflexion Memory

> 被 [AGENTS.md](../AGENTS.md) §Execution Strategy 引用为 Reflexion memory store。

> **Purpose**: Structured failure memory extracted from GCL traces (as of 2026-07-24) and Self-Review records.
> Agents can optionally load this file during Pre-flight to prevent known errors.
>
> **Maintenance**: Updated automatically via Self-Review Round 3 (Lessons Learned).
> **Token budget**: <= 200 lines. When exceeded, prune low-frequency patterns (count < 3).
>
> **Note**: Patterns with `Count = 0` are pre-defined based on common failure scenarios.
> These counts will be populated as GCL runs execute and failures are recorded.
>
> **Data source**: Count values reflect GCL trace accumulation since 2026-05. See `.runtime/memory/failure-patterns.json` for authoritative store.

---

## 1. CLI Parameter Errors

> Failure patterns extracted from GCL traces for `ve` CLI parameter errors.

| Skill | Error Pattern | Fix | Count |
|-------|---------------|-----|-------|
| `ve-ecs-ops` | `StopInstances` → `MissingParameter` — missing InstanceIds | `--InstanceIds "[\"i-xxx\"]"` (JSON array) | 3 |
| `ve-ecs-ops` | `RunInstances` → `InvalidParameter` — SecurityGroupIds format | `--SecurityGroupIds "[\"sg-xxx\"]"` | 2 |
| `ve-ecs-ops` | `DescribeInstances` → `InvalidParameter` — zone format | Use `cn-beijing` region name format | 1 |
| `ve-redis-ops` | `DeleteDBInstance` → `MissingParameter` — missing InstanceId | `--InstanceId redis-xxx` | 2 |
| `ve-rds-mysql-ops` | `DeleteDBInstance` → `InvalidParameter` — InstanceId format | Use `mysql-xxx` format | 1 |
| `ve-cms-ops` | `PutAlarmRule` → `InvalidParameter` — period not in range | Use 60/300/900/3600/43200/86400 | 1 |
| -- | `UnauthorizedOperation` — IAM policy missing permissions | Grant `VolcengineXxxFullAccess` or custom policy | 5 |
| -- | `InvalidAccessKeyId` — AccessKey missing/env not set | Check `VOLCENGINE_ACCESS_KEY` env var | 4 |
| -- | `SignatureDoesNotMatch` — signature mismatch | Confirm SecretKey and signing algorithm | 2 |
| `ve-vpc-ops` | `CreateVpc` → `InvalidCidrBlock.Malformed` — CIDR format | Use valid CIDR notation (e.g. `172.16.0.0/12`) | 1 |
| `ve-vpc-ops` | `CreateSubnet` → `IncorrectStatus.Subnet` — not Available | Wait for subnet to become Available | 1 |
| `ve-rds-ops` | `CreateDBInstance` → `BalanceNotEnough` — insufficient balance | Recharge via billing console | 2 |
| `ve-redis-ops` | `CreateDBInstance` → `Forbidden` — cross-service auth needed | `ve iam CreateServiceLinkedRole --body '{"ServiceName": "Redis"}'` | 3 |

**`ve` CLI parameter format notes**:
- Array parameters: `--InstanceIds "[\"i-xxx\"]"` (JSON array), **no** `.N` suffix
- Single-value parameters: `--InstanceId i-xxx`, direct assignment
- Nested objects: `--Filter.1.Name=zone_id --Filter.1.Value.1=cn-beijing` (1-indexed)

---

## 2. Skill Generation Issues

> Structural error patterns from the skill generator (`ve-skill-generator`).

| Issue Type | Fix Pattern | Count |
|------------|-------------|-------|
| Missing YAML frontmatter (name, description, compatibility, cli_applicability, metadata) | Always start with `---` block | 0 |
| TE-6 violation (cross-file duplication) | Delete duplicate from references/, keep SKILL.md as authoritative | 0 |
| Missing SHOULD/SHOULD NOT section | Add trigger conditions chapter with delegation rules | 0 |
| Broken relative links (references/ → assets/) | Use `../` prefix | 0 |
| Missing Well-Architected table | Add four-pillar table (Reliability, Security, Cost, Efficiency) | 0 |
| TE-1 violation (hardcoded versions) | Replace with `ve` query command for dynamic version fetching | 0 |
| Missing `cli_applicability` field | Add `dual-path` / `cli-first` / `cli-only` / `sdk-only` to frontmatter | 0 |
| Missing `cli_support_evidence` | Cite verification command (e.g. `ve ecs help`) | 0 |

---

## 3. Cross-Skill Composition Failures

> Failure patterns in cross-skill call chains (including SDK common errors).

| Skill Chain | Failure Pattern + Resolution | Count |
|--------------|-----------------------------|-------|
| `ve-redis-ops` → `ve-ecs-ops` | `ve ecs RunCommand` encoding failure → use base64 encoding | 2 |
| `ve-rds-mysql-ops` → `ve-ecs-ops` | Large SQL file execution timeout → split into chunks < 10 KB | 1 |
| `ve-redis-ops` → `ve-ecs-ops` | `redis-cli` not installed on target ECS → add idempotent install probe in Pre-flight | 3 |
| `ve-cms-ops` → `ve-ecs-ops` | New alarm `DescribeAlarms` returns empty → wait 60 s after `PutAlarmRule` before querying | 2 |
| `ve-vke-ops` → `ve-*-ops` | Worker returns empty `{{output.product_assessment}}` → verify skill has `## Read-Only Assessment Mode` | 1 |
| -- | Go SDK import path error → `import "github.com/volcengine/volc-sdk-golang/service/ecs"` | 4 |
| -- | Go SDK version mismatch → `go get github.com/volcengine/volc-sdk-golang@latest` | 2 |
| -- | SDK exception not caught → catch `*ecs.Error` and check `Code` and `Message` fields | 3 |
| `ve-rds-ops` → `ve-vpc-ops` | VPC subnet route table misconfigured → verify route table entries before RDS creation | 1 |
| `ve-ecs-ops` → `ve-iam-ops` | IAM policy missing `ecs:RunCommand` → add permission to IAM policy | 2 |

---

## 4. Runtime Execution Patterns

> Runtime failure patterns discovered during GCL execution.

| Skill | Failure Pattern + Prevention | Count |
|-------|-----------------------------|-------|
| `ve-ecs-ops` | `StopInstances` stuck in Stopping state → check running processes before stop | 0 |
| `ve-rds-mysql-ops` | `CreateDBInstance` quota exceeded → query quota before creation | 0 |
| `ve-redis-ops` | `DeleteDBInstance` permission denied → verify IAM policy in Pre-flight | 0 |
| `ve-vke-ops` | `CreateCluster` insufficient VPC subnet IPs → plan CIDR before cluster creation | 0 |
| `ve-rds-ops` | `ModifyDBInstance` status not Running → wait for instance to reach Running state | 0 |
| `ve-redis-ops` | `ModifyDBInstanceParams` param modification failed → check parameter name and valid values | 0 |
| `ve-ecs-ops` | `RunCommand` execution timeout → increase timeout or optimize command | 0 |
| `ve-vpc-ops` | `CreateVpc` CidrBlock conflict → use different CIDR range | 0 |

---

## 5. Token Efficiency Violations

> Common violations of Token Efficiency rules.

| TE Rule | Common Violation | Fix | Count |
|---------|------------------|-----|-------|
| TE-1 | Hardcoded region/zone lists in references/ | Use `ve ecs DescribeZones` query | 0 |
| TE-3 | Error table with > 3 columns | Merge columns, 1 error code per row | 0 |
| TE-4 | JSON paths scattered across file | Declare at file top in one block | 0 |
| TE-6 | Same script in SKILL.md and references/ | Delete from references, keep SKILL.md copy | 0 |

---

## 6. Orchestration & GCL Patterns

> Failure patterns in orchestration (`incident-loop-agent`) and GCL (Generator-Critic-Loop) execution.
> **Status**: Rows with `Count=0` are pre-seeded design hypotheses, not Reflexion-promoted patterns.
> Governance (per `docs/reflexion-memory.md` §4 and `incident-loop-agent/SKILL.md`):
> Reflexion only promotes a pattern once its `count ≥ 10` from *real* incidents.
> These seed rows are excluded from that threshold until a real incident hits them.

| Scope | Category | Failure Pattern | Resolution | Count |
|-------|----------|-----------------|------------|-------|
| `orchestration` | `execution_risk` | Operation blocked by REFUSE policy | Escalate to human or supply `--confirmed` for ASK class | 8 |
| `orchestration` | `max_iter` | Orchestration loop exceeded max iterations | Increase max_iter or simplify execution plan | 1 |
| `gcl` | `max_iter` | Generator exceeded max iterations | Simplify task or increase max_iter | 2 |
| `orchestration` | `safety_fail` | Safety=0 in iteration | Review and fix safety violations before retry | 0 |
| `gcl` | `safety_fail` | Safety check failed in Critic | Review safety rules and fix violations | 1 |
| `orchestration` | `timeout` | Orchestration iteration exceeded timeout | Optimize execution or increase timeout | 0 |
| `gcl` | `timeout` | GCL run exceeded overall timeout | Break into smaller tasks | 0 |
| `orchestration` | `credential_leak` | Credential detected in trace output | Mask credentials in all trace output | 0 |
| `orchestration` | `cross_skill_timeout` | Cross-skill delegation timed out | Increase delegation timeout or simplify | 0 |
| `gcl` | `critic_error` | Critic returned invalid score | Validate critic prompt and rubric | 0 |
| `gcl` | `trace_write` | Failed to write GCL trace to disk | Check disk space and permissions | 0 |
| `orchestration` | `budget_overrun` | Fix cost exceeds budget without warning | Include cost estimate in propose step before execution | 0 |
| `orchestration` | `wrong_billing_model` | PostPaid→PrePaid migration missed | Add billing model check in diagnose step | 0 |
| `gcl` | `missing_cost_dimension` | Critic accepted fix without cost assessment | Validate Cost Efficiency dimension score > 0 | 0 |

---

## 7. Security and Compliance Patterns

> Security-related failure patterns from operations.

| Category | Failure Pattern | Resolution | Count |
|----------|-----------------|------------|-------|
| `iam_insufficient` | IAM policy missing required permissions | Grant appropriate IAM policy | 5 |
| `credential_expired` | AccessKey/SecretKey expired | Regenerate credentials | 2 |
| `public_exposure` | Resource accidentally exposed publicly | Restrict to VPC/private access | 1 |
| `encryption_missing` | Resource created without encryption | Enable encryption at rest/in transit | 0 |

---

## Usage Guidelines

### For Agents (Pre-flight)

```
# Optional: Load failure patterns before executing a skill
# 1. Read this file (lazy-load, ~200 lines)
# 2. Filter patterns by current skill name
# 3. Inject relevant patterns into Generator context as prevention hints
```

### For Self-Review (Round 3: Lessons Learned)

```
# After completing R1 + R2:
# 1. Extract new failure patterns from this session
# 2. Check if pattern already exists (dedup by (skill, pattern))
# 3. If new: append to appropriate section with count=1
# 4. If existing: increment count
# 5. If total lines > 200: prune patterns with count < 3
```

### For GCL Traces

```
# When a GCL iteration fails, record the failure pattern:
{
  "failure_pattern": {
    "category": "cli_parameter" | "skill_generation" | "cross_skill" | "runtime" | "token_efficiency" | "orchestration" | "gcl" | "security",
    "skill": "ve-xxx-ops",
    "command": "ve xxx ...",
    "error": "InvalidParameter: ...",
    "fix": "Use JSON array format for array params",
    "reusable": true | false
  }
}
```

---

## Extracted from GCL Traces (auto-generated)

> This block is **auto-generated** by the GCL write-back pipeline at MAX_ITER / SAFETY_FAIL.
> Patterns are persisted to `.runtime/memory/failure-patterns.json` (structured JSON, dedup by `(skill, pattern)`, count++).
> This markdown block is a compatibility snapshot — the JSON store is the authoritative source.
> See `docs/reflexion-learning-architecture.md` for the complete write-back chain.
>
> **Do not edit this block manually.**

| Skill | Pattern | Category | Fix | Count |
|-------|---------|----------|-----|-------|
| `ve-ecs-ops` | `max_iter` | `orchestration` | execution failed with error class: max_iter | 1 |
| `ve-skill-generator` | `operation blocked by execution-risk policy: REFUSE` | `execution_risk` | escalate to human or supply `--confirmed` for ASK class | 8 |

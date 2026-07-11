# Failure Patterns — Reflexion Memory

> 被 [AGENTS.md](../AGENTS.md) §Execution Strategy 引用为 Reflexion memory store。

> **Purpose**: Structured failure memory extracted from GCL traces and Self-Review records.
> Agents can optionally load this file during Pre-flight to prevent (prevent) known errors.
>
> **Maintenance**: Updated automatically via Self-Review Round 3 (Lessons Learned).
> **Token budget**: <= 200 lines. When exceeded, prune low-frequency patterns (count < 3).
>
> **Note**: Patterns with `Count = 0` are pre-defined based on common failure scenarios.
> These counts will be populated as GCL runs execute and failures are recorded.

---

## 1. CLI Parameter Errors

> Failure patterns extracted from GCL traces for `ve` CLI parameter errors.

| Skill | Command | Error Pattern | Root Cause | Fix | Count |
|-------|---------|---------------|------------|-----|-------|
| `ve-ecs-ops` | `StopInstances` | `MissingParameter` | Missing required parameter InstanceIds | `--InstanceIds "[\"i-xxx\"]"` (JSON array) | 0 |
| `ve-ecs-ops` | `RunInstances` | `InvalidParameter` | SecurityGroupIds format error | `--SecurityGroupIds "[\"sg-xxx\"]"` | 0 |
| `ve-ecs-ops` | `DescribeInstances` | `InvalidParameter` | Zone format error | Use `cn-beijing` region name format | 0 |
| `ve-redis-ops` | `DeleteDBInstance` | `MissingParameter` | Missing InstanceId | `--InstanceId redis-xxx` | 0 |
| `ve-rds-mysql-ops` | `DeleteDBInstance` | `InvalidParameter` | InstanceId format error | Use `mysql-xxx` format | 0 |
| `ve-cms-ops` | `PutAlarmRule` | `InvalidParameter` | Period not in valid range | Use 60/300/900/3600/43200/86400 | 0 |
| -- | `UnauthorizedOperation` | `UnauthorizedOperation` | IAM policy missing permissions | Grant `VolcengineXxxFullAccess` or custom policy | 0 |
| -- | `InvalidAccessKeyId` | `InvalidAccessKeyId` | AccessKey does not exist or env not set | Check `VOLCENGINE_ACCESS_KEY` env var | 0 |
| -- | `SignatureDoesNotMatch` | `SignatureDoesNotMatch` | Signature computation error | Confirm SecretKey and signing algorithm | 0 |

**`ve` CLI parameter format notes**:
- Array parameters: `--InstanceIds "[\"i-xxx\"]"` (JSON array), **no** `.N` suffix
- Single-value parameters: `--InstanceId i-xxx`, direct assignment
- Nested objects: `--Filter.1.Name=zone_id --Filter.1.Value.1=cn-beijing` (1-indexed)

---

## 2. Skill Generation Issues

> Structural error patterns from the skill generator (`ve-skill-generator`).

| Issue Type | Frequency | Fix Pattern | First Seen |
|------------|-----------|-------------|------------|
| Missing YAML frontmatter | 0x | Always start with `---` block (name, description, compatibility, cli_applicability, metadata) | 2026-06 |
| TE-6 violation (cross-file duplication) | 0x | Delete duplicate from references/, keep SKILL.md as authoritative | 2026-06 |
| Missing SHOULD/SHOULD NOT section | 0x | Add trigger conditions chapter with delegation rules | 2026-06 |
| Broken relative links | 0x | Use `../` prefix for references/ to assets/ links | 2026-06 |
| Missing Well-Architected table | 0x | Add four-pillar table (Reliability, Security, Cost, Efficiency) | 2026-06 |
| TE-1 violation (hardcoded versions) | 0x | Replace with `ve` query command for dynamic version fetching | 2026-06 |
| Missing `cli_applicability` field | 0x | Add `dual-path` / `cli-first` / `cli-only` / `sdk-only` to frontmatter | 2026-06 |
| Missing `cli_support_evidence` | 0x | Cite verification command (e.g. `ve ecs help`) | 2026-06 |

---

## 3. Cross-Skill Composition Failures

> Failure patterns in cross-skill call chains (including SDK common errors).

| Source Skill | Target Skill | Failure Pattern | Resolution | Count |
|--------------|--------------|-----------------|------------|-------|
| `ve-redis-ops` | `ve-ecs-ops` | `ve ecs RunCommand` encoding failure | Use base64 encoding for command content | 0 |
| `ve-rds-mysql-ops` | `ve-ecs-ops` | Large SQL file execution timeout | Split into chunks < 10 KB | 0 |
| `ve-redis-ops` | `ve-ecs-ops` | `redis-cli` not installed on target ECS | Add idempotent install probe in Pre-flight | 0 |
| `ve-cms-ops` | `ve-ecs-ops` | New alarm `DescribeAlarms` returns empty | Wait 60 s after `PutAlarmRule` before querying | 0 |
| `ve-vke-ops` | `ve-*-ops` | Worker returns empty `{{output.product_assessment}}` | Verify skill has `## Read-Only Assessment Mode` section | 0 |
| -- | -- | Go SDK import path error | `import "github.com/volcengine/volc-sdk-golang/service/ecs"` | 0 |
| -- | -- | Go SDK version mismatch | `go get github.com/volcengine/volc-sdk-golang@latest` | 0 |
| -- | -- | SDK exception not caught | Catch `*ecs.Error` and check `Code` and `Message` fields | 0 |

---

## 4. Runtime Execution Patterns

> Runtime failure patterns discovered during GCL execution.

| Skill | Operation | Failure Pattern | Root Cause | Prevention |
|-------|-----------|-----------------|------------|------------|
| `ve-ecs-ops` | `StopInstances` | Instance stuck in Stopping state | Dependent services not stopped | Check running processes before stop |
| `ve-rds-mysql-ops` | `CreateDBInstance` | Quota exceeded error | Account-level instance limit | Query quota before creation |
| `ve-redis-ops` | `DeleteDBInstance` | Permission denied | IAM policy missing `redis:*Action` | Verify IAM policy in Pre-flight |
| `ve-vke-ops` | `CreateCluster` | Insufficient VPC subnet IPs | Subnet CIDR too small | Plan CIDR before cluster creation |

---

## 5. Token Efficiency Violations

> Common violations of Token Efficiency rules.

| TE Rule | Common Violation | Fix | Frequency |
|---------|------------------|-----|-----------|
| TE-1 | Hardcoded region/zone lists in references/ | Use `ve ecs DescribeZones` query | 0x |
| TE-3 | Error table with > 3 columns | Merge columns, 1 error code per row | 0x |
| TE-4 | JSON paths scattered across file | Declare at file top in one block | 0x |
| TE-6 | Same script in SKILL.md and references/ | Delete from references, keep SKILL.md copy | 0x |

---

## 6. Incident Response Failures

> Failure patterns specific to the `incident-loop-agent` scenario — automating customer
> alert → triage → diagnose → fix loops. Distinct from §3 cross-skill because the
> failure mode is **the orchestration layer**, not the leaf skill invocation.

> **Status**: this table is **pre-seeded**, not empty. The 14 rows below are **`seed` / `hypothesis`** patterns — they are design placeholders authored up-front, **not** Reflexion memory harvested from real GCL traces.
>
> Governance (per `docs/reflexion-memory.md` §4): Reflexion only promotes a pattern once its `count ≥ 3` from *real* incidents. These seed rows start at `Count = 0` and are **excluded** from that threshold until a real incident hits them; only then does the GCL write-back (via `gcl_trace_aggregate`) fill in `Count` with observed data. Do not delete them — they are intentional design placeholders, but they must not be mistaken for learned Reflexion entries.

| Scenario | Failure Pattern | Root Cause | Fix | Count |
|----------|-----------------|------------|-----|-------|
| Alarm triage | `routing_mismatch` | `incident-loop-agent` matches alarm to wrong primary skill (e.g. Redis CPU → `ve-ecs-ops`) | Verify against `docs/skill-routing-graph.md` rule before delegation; default to `ve-cms-ops` on unknown | 0 |
| Alarm triage | `correlated_alarm_missed` | Multiple alarms fire within 5min but only the loudest is investigated | Always run correlation pass via `ve-cms-ops` first (Rule 2) before dispatching primary | 0 |
| Diagnosis | `data_collection_timeout` | Cross-domain data collection (>60s) blocks the loop past Critic's patience | Set per-call `timeout=30s`; emit `{{output.partial_data}}` and let Critic judge on what's available | 0 |
| Diagnosis | `evidence_overfetch` | AI runs >20 read-only `Describe*` calls when 5 would suffice | Cap data calls at `min(15, alarm_count × 3)`; log skipped calls | 0 |
| Hypothesis | `causal_hallucination` | AI infers root cause from a single data point (e.g. CPU spike → "OOM killer") | Critic MUST flag any conclusion based on < 3 independent data points; require ≥ 1 verifying query | 0 |
| Hypothesis | `recent_change_blame` | AI attributes anomaly to last-deploy without proof | Pull change-events from last 24h via `ve-cms-ops`; require timestamp-overlap with anomaly onset | 0 |
| Fix proposal | `unverified_fix` | AI proposes fix without testing impact scope (e.g. "restart Redis" without checking active connections) | Pre-flight MUST enumerate blast radius (active clients, downstream callers) before any destructive proposal | 0 |
| Fix proposal | `unbounded_scope` | Fix proposal mutates resources beyond the symptom scope (e.g. "restart all 6 Redis nodes" for 1 slow node) | Require `resource_scope` list with each proposal; Critic rejects if mutated scope > symptom scope | 0 |
| Safety gate | `safety_gate_bypass` | Destructive op (Redis FlushAll / instance delete) skipped user confirmation | `{{user.confirm}}` gate MUST be honored verbatim; Safety=0 → ABORT per `docs/gcl-spec.md` §5 | 0 |
| Safety gate | `silent_destructive_default` | Fix proposal defaults to `--force` or `--skip-confirm` | Force flag MUST be explicit; reject any proposal where destructive flag is on by default | 0 |
| Execute | `retry_storm` | Failed op auto-retried without backoff, hits rate limit | Exponential backoff mandatory; max 2 retries; surface original error after 2nd fail | 0 |
| Execute | `partial_rollback` | Multi-step fix aborted midway; some steps already applied | Snapshot resource state via `ve <svc> Describe*` before each destructive step; emit rollback script on partial completion | 0 |
| Reflexion | `pattern_undercount` | Pattern added to §6 with count=1 then forgotten; threshold for promotion never reached | Self-Review Round 3 MUST scan §6 every run; promote when count ≥ 3 | 0 |
| Reflexion | `category_mismatch` | Incident pattern placed in §3 `cross_skill` losing orchestration-specific context | §6 owns orchestration layer (§3 = leaf skill internal failures) | 0 |

**Anti-pattern candidates to seed (preventive)**:
- ❌ AI proposed "restart" without checking active client count
- ❌ AI picked `ve-ecs-ops` for Redis CPU alarm (wrong routing table match)
- ❌ AI issued `FlushAll` without `{{user.confirm}}` confirmation
- ❌ AI made >10 read-only calls for a simple diagnosis (evidence-overfetch)
- ❌ AI auto-retry without backoff (→ rate limit / API throttle)

---

## Usage Guidelines

### For Agents (Pre-flight)

```
# Optional: Load failure patterns before executing a skill
# 1. Read this file (lazy-load, ~130 lines)
# 2. Filter patterns by current skill name
# 3. Inject relevant patterns into Generator context as prevention hints
```

### For Self-Review (Round 3: Lessons Learned)

```
# After completing R1 + R2:
# 1. Extract new failure patterns from this session
# 2. Check if pattern already exists (dedup by skill + command + error)
# 3. If new: append to appropriate section with count=1
# 4. If existing: increment count
# 5. If total lines > 200: prune patterns with count < 3
```

### For GCL Traces

```
# When a GCL iteration fails, record the failure pattern:
{
  "failure_pattern": {
    "category": "cli_parameter" | "skill_generation" | "cross_skill" | "runtime" | "token_efficiency",
    "skill": "ve-xxx-ops",
    "command": "ve xxx ...",
    "error": "InvalidParameter: ...",
    "fix": "Use JSON array format for array params",
    "reusable": true | false
  }
}
```

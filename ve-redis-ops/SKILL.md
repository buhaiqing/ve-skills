---
name: ve-redis-ops
description: >-
  Use when the user needs to deploy, configure, troubleshoot, or monitor Volcengine
  Cache for Redis (缓存数据库 Redis 版) — Instance lifecycle, configuration, parameter
  management, allowlists, backups, and diagnostics. User mentions Redis, 缓存数据
 库, cache instance, or describes product-specific scenarios (e.g., connection drops,
  performance degradation, instance creation failures) even without naming the
  product directly. Not for billing, IAM, or VPC networking.
license: MIT
compatibility: >-
  Official Volcengine CLI (`ve`, Go binary, no runtime), Go 1.14+ runtime
  (for JIT SDK fallback), valid API credentials, network access to Volcengine
  endpoints.
metadata:
  author: volcengine
  version: "1.1.0"
  last_updated: "2026-07-04"
  runtime: Harness AI Agent, Claude Code, Cursor, or compatible Agent runtimes
  go_version_minimum: "1.14"
  go_version_jit: "1.21+"
  api_profile: "Redis OpenAPI 2020-12-07 — https://www.volcengine.com/docs/6293"
  cli_applicability: dual-path
  cli_support_evidence: >-
    Redis is a core Volcengine service; `ve redis --help` lists instance operations.
    OpenAPI service ID: Redis, version: 2020-12-07.
    Go SDK: github.com/volcengine/volc-sdk-golang/service/redis.
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

# Volcengine Cache for Redis Operations Skill

### What This Skill Does

This skill manages Volcengine Cache for Redis instance lifecycle, configuration, and monitoring using the `ve` CLI (primary) or JIT Go SDK (fallback). It covers instance creation, deletion, scaling, parameter tuning, allowlist management, and backup operations. Security gates guard destructive actions such as instance deletion.

## Overview

Volcengine Cache for Redis (缓存数据库 Redis 版) provides managed Redis-compatible in-memory cache instances. This skill is an **operational runbook** for agents: explicit scope, credential rules, pre-flight checks, **dual-path execution** (official **SDK/API** and official **`ve` CLI**), response validation, and failure recovery.

### CLI applicability

- **`cli_applicability: dual-path`:** `ve` CLI supports Redis operations. Document **both** SDK and CLI steps for every operation.

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use with Redis triggers; VPC delegated to ve-vpc-ops |
| 2 | **Structured I/O** | `{{env.*}}` for credentials, `{{user.*}}` for instance params, `{{output.*}}` for responses |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | 14 Redis-specific error codes with HALT/retry per type |
| 5 | **Absolute Single Responsibility** | Redis instances only; VPC/ECS delegated |

## Trigger & Scope

### SHOULD Use This Skill When

- User mentions "Redis", "缓存数据库", "Cache for Redis", "Volcengine Redis", "redis-cli"
- Task involves Redis instance lifecycle (create, describe, modify, delete, list)
- Task keywords: redis, cache, instance, allowlist, backup, 缓存, 实例, whitelist
- User asks about Redis connection issues, performance, or configuration

### SHOULD NOT Use This Skill When

- Task is billing / account → delegate to `ve-billing-ops`
- Task is IAM / permissions → delegate to `ve-iam-ops`
- Task is VPC / subnet → delegate to `ve-vpc-ops`
- Task is ECS instance management → delegate to `ve-ecs-ops`
- User insists on console-only flows → state limitation

### Delegation Rules

- Redis instances require VPC and subnets: verify via `ve-vpc-ops`
- Redis runs on ECS: underlying host issues delegate to `ve-ecs-ops`

## Variable Convention

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | Runtime credential | NEVER ask user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | Runtime credential | NEVER ask user; fail if unset |
| `{{env.VOLCENGINE_REGION}}` | Runtime region | Use from env |
| `{{user.instance_name}}` | Instance name | Ask once; reuse |
| `{{user.instance_id}}` | Instance ID | Ask if not known from context |
| `{{user.engine_version}}` | Redis version (4.0/5.0/6.0) | Ask with options |
| `{{output.instance_id}}` | From CreateDBInstance response | Parse `$.Result.InstanceId` |

> **Security Warning:** NEVER log, print, or expose `VOLCENGINE_SECRET_KEY`. Mask all credentials with `***`.

## API and Response Conventions

- **OpenAPI 2020-12-07** is canonical for all field names and response shapes
- **API Version:** `2020-12-07`, **Service:** `Redis`
- **Endpoint:** `redis.{region}.volcengineapi.com`

### Key Response Field Table

完整字段定义见 [references/api-sdk-usage.md §2](references/api-sdk-usage.md#2-instance-management)。

### Instance States

| State | Description |
|-------|-------------|
| `Creating` | Being created |
| `Running` | Operational |
| `Stopped` | Stopped |
| `Deleting` | Being deleted |
| `Error` | Error state |
| `Changing` | Config change in progress |

### State Transitions

`Create` — → `Running` (⏱ 5s poll, max 300s)  
`Delete` — any → gone (⏱ 5s poll, max 300s)  
`Restart` — Running → `Running` (⏱ 5s poll, max 180s)  
`Modify Spec` — Running → `Running` (⏱ 10s poll, max 600s)

## Quick Start

### Prerequisites

| Check | Status |
|-------|--------|
| `ve` CLI installed | ❌ |
| `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`, `VOLCENGINE_REGION` set | ❌ |

### Verify Setup
```bash
ve redis DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}"
```

### Your First Command
```bash
ve redis DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}" --PageNumber 1 --PageSize 10
```
### Next Steps
- [Core Concepts](references/core-concepts.md) — Understand Redis architecture
- [Common Operations](#execution-flows) — Create, manage, and manage Redis databases
- [Troubleshooting](references/troubleshooting.md) — Fix common issues

## Capabilities at a Glance

| Operation | Description | ⚡ Complexity | 🛡️ Risk | | safety_class | blast_radius |
|-----------|-------------|------------|------||---|---|
| CreateDBInstance | Create Redis instance | Medium | Low | | state-changing | single |
| DescribeDBInstanceDetail | Get instance details | Low | ✅ None | | read-only | single |
| DescribeDBInstances | List all instances | Low | ✅ None | | read-only | single |
| 🔴 DeleteDBInstance | Delete instance | Low | 🔴 **High** | | destructive | single |
| 🔴 ModifyDBInstanceSpec | Change instance spec | Medium | Medium | | state-changing | single |
| 🔴 RestartDBInstance | Restart instance | Low | Medium | | state-changing | single |
| DescribeDBInstanceParameters | Query parameters | Low | ✅ None | | read-only | single |
| 🔴 ModifyDBInstanceParameters | Modify parameters | Medium | Medium | | state-changing | single |
| CreateAllowList | Create IP whitelist | Low | Low | | state-changing | multi |
| DescribeAllowLists | List allowlists | Low | ✅ None | | read-only | single |
| 🔴 ModifyAllowList | Update allowlist | Low | Low | | state-changing | multi |
| 🔴 DeleteAllowList | Delete allowlist | Low | Medium | | destructive | single |
| DescribeAccounts | List accounts | Low | ✅ None | | read-only | single |
| 🔴 CreateAccount | Create DB account | Low | Medium | | state-changing | account-or-region |
| DescribeBackups | List backups | Low | ✅ None | | read-only | single |
| CreateBackup | Create manual backup | Low | Low | | state-changing | single |

## Changelog
| Version | Date | Changes |
|---------|------|---------|
| 1.0.1 | 2026-07-13 | T04: annotate operation table with safety_class + blast_radius leaf-op metadata columns (L3 policy inputs); see ve-skill-generator/references/leaf-op-metadata-spec.md |
| 1.0.0 | 2026-05-16 | Initial release with Redis instance lifecycle |
| 1.1.0 | 2026-06-04 | Phase 1 GCL rollout: added `## Quality Gate (GCL)` chapter, `references/rubric.md`, `references/prompt-templates.md`; `max_iter=2` for destructive / state_changing ops, `max_iter=3` for read-only ops |

## Quality Gate (GCL)

> This chapter is **mandatory** for every execution of `ve-redis-ops`. It implements
> the Generator-Critic-Loop defined in `../../AGENTS.md` §3-§9. Read
> [`references/rubric.md`](references/rubric.md) for scoring and
> [`references/prompt-templates.md`](references/prompt-templates.md) for safety prompts.

### Operation Tiers

> See [`references/rubric.md` §0](references/rubric.md#0-operation-tier) for the full operation tier table.

### Loop

1. **Pre-flight** — resolve `{{env.*}}` / `{{user.*}}`; classify tier; load rubric.
2. **Generate** — execute per `## Execution Flows`. Trace to `./audit-results/gcl-trace-*.json`.
3. **Critique** — isolated prompt; score 5 dimensions; MUST NOT see raw request.
4. **Decide** — Safety=0 on Destructive/State-changing → **ABORT**; all pass → return; `iter<max_iter` → loop.

### Redis-specific safety rules

- **DeleteDBInstance**: check deletion protection first; explicit confirmation.
- **ModifyDBInstanceSpec**: warn about 60-180s downtime.
- **RestartDBInstance**: warn about connection cutoff; production confirm.
- **ModifyAllowList**: production change risks locking out clients.

### Trace

`./audit-results/gcl-trace-*.json` — password masked as `<masked>`.

### Cross-skill delegation

| Critic finding | Delegate to |
|---|---|
| VPC/subnet not found | `ve-vpc-ops` |
| ECS host issue | `ve-ecs-ops` |
| Billing quota | `ve-billing-ops` |

## Execution Flows

### Operation: Create Redis Instance (CreateDBInstance)

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| CLI available | `ve version` | Exit code 0 | Go to JIT Go SDK fallback |
| Credentials | Verify env vars set | Non-empty | HALT |
| Region | Verify supported | Valid region | HALT; suggest regions |
| VPC/Subnet | Verify network exists | Valid VPC/subnet | HALT; use ve-vpc-ops |
| Quota | Check instance quota | Sufficient | HALT; raise quota |

#### Execution — CLI (Primary)

```bash
ve redis CreateDBInstance \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceName "{{user.instance_name}}" \
  --EngineVersion "{{user.engine_version}}" \
  --NodeNumber 2 \
  --ShardCapacity "{{user.shard_capacity_gb}}" \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --ChargeType "PostPaid" \
  --Password "{{user.password}}"
```

#### Execution — JIT Go SDK (Fallback)

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/redis"
)

func main() {
    instance := redis.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "Region":       os.Getenv("VOLCENGINE_REGION"),
        "InstanceName": os.Getenv("INSTANCE_NAME"),
        "EngineVersion": os.Getenv("ENGINE_VERSION"),
        "NodeNumber":   2,
        "ShardCapacity": os.Getenv("SHARD_CAPACITY"),
        "VpcId":        os.Getenv("VPC_ID"),
        "SubnetId":     os.Getenv("SUBNET_ID"),
        "ChargeType":   "PostPaid",
        "Password":     os.Getenv("PASSWORD"),
    }

    resp, err := instance.Client.Request("redis", "CreateDBInstance", params)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(resp))
}
```

#### Post-execution Validation

```bash
for i in $(seq 1 60); do
  STATUS=$(ve redis DescribeDBInstanceDetail --InstanceId "{{output.instance_id}}" | jq -r '.Result.Status // ""')
  [ "$STATUS" = "Running" ] && break
  [ "$STATUS" = "Error" ] && echo "Instance creation failed" && exit 1
  sleep 5
done
```

#### Failure Recovery

| Error Pattern | Agent Action | Recovery |
|---------------|-------------|----------|
| `OperationDenied.DeletionProtection` | **HALT** | Disable deletion protection first then retry |
| `ResourceAlreadyExists` | Ask reuse | Use unique instance name or reuse existing |
| `InvalidParameter.Password` | **HALT** | Use 8-32 chars with letters, digits, and special chars |
| `Throttling` | Retry 3x, exponential backoff | Rate limit reached; retry with backoff |
| `InternalError` | Retry 3x with backoff (2s,4s,8s); **HALT** after 3 | Capture RequestId; retry or escalate |

> 其他通用错误见 [references/troubleshooting.md §1](references/troubleshooting.md#1-error-taxonomy)。

### Operation: Describe/ List Instances

```bash
# Details
ve redis DescribeDBInstanceDetail --InstanceId "{{user.instance_id}}" --Region "{{env.VOLCENGINE_REGION}}"

# List all
ve redis DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}" --PageNumber 1 --PageSize 100
```

Present to user with the centralized **Key Response Field Table** above.

### Operation: Delete Instance

#### Safety Gate

- **MUST** obtain explicit confirmation for delete of `{{user.instance_name}}` (`{{user.instance_id}}`)
- **MUST** check deletion protection status first
- **MUST NOT** proceed without clear user assent

#### Execution

```bash
ve redis DeleteDBInstance --InstanceId "{{user.instance_id}}"
```

#### Validation
Poll `DescribeDBInstanceDetail` until 404 or instance gone.

### Operation: Modify Instance Spec

```bash
ve redis ModifyDBInstanceSpec --InstanceId "{{user.instance_id}}" --body '{"ShardCapacity": 4096}'
```

Poll until status returns to `Running`.

### Operation: Restart Instance

```bash
ve redis RestartDBInstance --InstanceId "{{user.instance_id}}"
```

### Operations: Allowlist (CRUD)

```bash
# Create
ve redis CreateAllowList --Region "{{env.VOLCENGINE_REGION}}" --Name "{{user.allowlist_name}}" --Desc "Auto-created"

# Describe
ve redis DescribeAllowLists --Region "{{env.VOLCENGINE_REGION}}"

# Modify (add IPs)
ve redis ModifyAllowList --AllowListId "{{user.allowlist_id}}" --body '{"IPList": ["10.0.0.0/8", "192.168.1.0/24"]}'

# Delete
ve redis DeleteAllowList --AllowListId "{{user.allowlist_id}}"
```

### Operations: Account Management

```bash
ve redis DescribeAccounts --InstanceId "{{user.instance_id}}"
ve redis CreateAccount --InstanceId "{{user.instance_id}}" --AccountName "{{user.account_name}}" --AccountPassword "{{user.password}}" --AccountRole "Standard"
```

### Operations: Backup Management

```bash
ve redis DescribeBackups --InstanceId "{{user.instance_id}}"
ve redis CreateBackup --InstanceId "{{user.instance_id}}" --BackupName "{{user.backup_name}}"
```

### Operations: Parameter Management

```bash
ve redis DescribeDBInstanceParameters --InstanceId "{{user.instance_id}}"
ve redis ModifyDBInstanceParameters --InstanceId "{{user.instance_id}}" --body '{"Parameters": [{"Name": "maxmemory-policy", "Value": "allkeys-lru"}]}'
```

## Error Taxonomy

| 错误码 | 含义 | 分辨率 |
|--------|------|--------|
| `InvalidInstanceId.NotFound` | 实例 ID 不存在 | 0 retries; **HALT** — 检查 InstanceId 是否正确 |
| `IncorrectStatus.Instance` | 实例状态不允许操作 | 0 retries; **HALT** — 等待实例进入 Running 态后重试 |
| `QuotaExceeded.Instance` | 实例数量已达配额上限 | 0 retries; **HALT** — 删除无用实例或申请提升配额 |
| `OperationDenied.DeletionProtection` | 实例启用了删除保护 | 0 retries; **HALT** — 先关闭删除保护再操作 |
| `InvalidParameter.Password` | 密码不符合 Redis 复杂度要求 | 0 retries; **HALT** — 使用 8-32 位含字母/数字/特殊字符的密码 |
| `ResourceAlreadyExists` | 实例名称已存在 | 0 retries; **HALT** — 使用唯一的实例名称 |
| `InsufficientBalance` | 账户余额不足 | 0 retries; **HALT** — 充值后重试 |
| `InvalidVpcId` | VPC ID 无效 | 0 retries; **HALT** — 通过 ve-vpc-ops 验证 VPC 是否存在 |
| `InvalidSecurityGroupId` | 安全组 ID 无效 | 0 retries; **HALT** — 通过 ve-ecs-ops 验证安全组 |
| `BackupInProgress` | 实例正在备份中 | 0 retries; **HALT** — 等待备份完成后再操作 |
| `Throttling` | 请求频率超限 | 3 retries/exponential/1s/2s/4s; **RETRY** |
| `InternalError` | 服务端内部错误 | 3 retries/exponential/2s/4s/8s; **RETRY** — 持续失败则记录 RequestId 并反馈 |

## Prerequisites

1. **`ve` CLI** installed per execution environment docs
2. **Go runtime** for JIT fallback (see references/integration.md)
3. **Credentials:** `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`, `VOLCENGINE_REGION`
4. **Verify:** `ve redis DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}"`

## Operational Best Practices

- **Authentication and network isolation:** Enable IP allowlists via `ve redis CreateAllowList` to restrict access. Deploy instances in private VPC subnets; avoid public endpoint exposure.
- **Persistence and backup:** Enable AOF persistence for durability-critical data. Schedule automatic backups via `ve redis CreateBackup` and test point-in-time recovery regularly.
- **High availability:** Deploy with primary-replica replication for failover. Use multi-zone deployment for region-level resilience.
- **Memory and eviction policy:** Set `maxmemory-policy` to `allkeys-lru` for cache workloads or `volatile-lru` for mixed use. Monitor `used_memory` and eviction rates.
- **Big key and hot key management:** Use `redis-cli --bigkeys` and `redis-cli --hotkeys` to identify oversized keys. Split large hashes or sets, and distribute hot keys with read replicas.
- **Cost-effective instance sizing:** Choose instance specs based on memory-to-CPU ratio. Start with small instances and scale vertically as needed. Delete idle instances to avoid unnecessary cost.

## Reference Directory

- [Core Concepts](references/core-concepts.md) — Redis architecture, instance types, engine versions
- [API & SDK Usage](references/api-sdk-usage.md) — Full operation map, JSON paths
- [CLI Usage](references/cli-usage.md) — `ve redis` command reference
- [Knowledge Base](references/advanced/knowledge-base.md) — fault pattern library (AIOps diagnosis)
- [SecurityOps (Advanced)](references/advanced/securityops.md) — Cache security baseline, access control, data protection, incident response
- [Troubleshooting Guide](references/troubleshooting.md) — Error codes, diagnostics
- [Monitoring & Alerts](references/monitoring.md) — Redis monitoring metrics
- [Integration](references/integration.md) — Go SDK setup, JIT workflow
- [GCL Rubric](references/rubric.md) — Scoring dimensions for the Generator-Critic-Loop
- [GCL Prompt Templates](references/prompt-templates.md) — G/C/O prompt skeletons + Redis-specific safety prompts

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
  version: "1.0.0"
  last_updated: "2026-05-16"
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

| Operation | JSON Path | Type | Description |
|-----------|-----------|------|-------------|
| CreateDBInstance | `$.Result.InstanceId` | string | New instance ID |
| DescribeDBInstances | `$.Result.Instances[].InstanceId` | string | Instance ID |
| DescribeDBInstances | `$.Result.Instances[].InstanceName` | string | Instance name |
| DescribeDBInstances | `$.Result.Instances[].Status` | string | Instance state |
| DescribeDBInstances | `$.Result.Instances[].EngineVersion` | string | Redis version |
| DescribeDBInstances | `$.Result.Instances[].Capacity.Total` | integer | Memory capacity (MB) |
| DescribeDBInstances | `$.Result.Instances[].ChargeType` | string | PostPaid/PrePaid |
| DescribeDBInstances | `$.Result.Instances[].PrivateAddress` | string | Connection address |
| DescribeDBInstances | `$.Result.Instances[].VpcId` | string | VPC ID |
| DescribeDBInstances | `$.Result.TotalInstancesNum` | integer | Total count |
| DeleteDBInstance | `$.ResponseMetadata.RequestId` | string | Request ID |

### Instance States

| State | Description |
|-------|-------------|
| `Creating` | Being created |
| `Running` | Operational |
| `Stopped` | Stopped |
| `Deleting` | Being deleted |
| `Error` | Error state |
| `Changing` | Configuration change in progress |

### State Transitions

| Operation | Initial | Target | Poll | Max Wait |
|-----------|---------|--------|------|----------|
| Create | — | `Running` | 5s | 300s |
| Delete | any | gone | 5s | 300s |
| Restart | Running | `Running` | 5s | 180s |
| Modify Spec | Running | `Running` | 10s | 600s |

## Quick Start

### Prerequisites
- [ ] `ve` CLI installed
- [ ] `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`, `VOLCENGINE_REGION` set

### Verify Setup
```bash
ve redis DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}"
```

### Your First Command
```bash
ve redis DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}" --PageNumber 1 --PageSize 10
```

## Capabilities at a Glance

| Operation | Description | Complexity | Risk |
|-----------|-------------|------------|------|
| CreateDBInstance | Create Redis instance | Medium | Low |
| DescribeDBInstanceDetail | Get instance details | Low | None |
| DescribeDBInstances | List all instances | Low | None |
| DeleteDBInstance | Delete instance | Low | **High** |
| ModifyDBInstanceSpec | Change instance spec | Medium | Medium |
| RestartDBInstance | Restart instance | Low | Medium |
| DescribeDBInstanceParameters | Query parameters | Low | None |
| ModifyDBInstanceParameters | Modify parameters | Medium | Medium |
| CreateAllowList | Create IP whitelist | Low | Low |
| DescribeAllowLists | List allowlists | Low | None |
| ModifyAllowList | Update allowlist | Low | Low |
| DeleteAllowList | Delete allowlist | Low | Medium |
| DescribeAccounts | List accounts | Low | None |
| CreateAccount | Create DB account | Low | Medium |
| DescribeBackups | List backups | Low | None |
| CreateBackup | Create manual backup | Low | Low |

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-16 | Initial release with Redis instance lifecycle |

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

| Error Pattern | Retries | Backoff | Agent Action | UX Feedback |
|---------------|---------|---------|--------------|-------------|
| `InvalidParameter.InstanceName` | 0 | — | HALT | `[ERROR] InvalidParameter.InstanceName. How to fix: Use valid name (alphanumeric, hyphens, 1-128 chars).` |
| `InvalidParameter.NetworkConfig` | 0 | — | HALT; verify VPC | `[ERROR] Invalid VPC/subnet config. How to fix: Verify network exists in target region.` |
| `QuotaExceeded.InstanceCount` | 0 | — | HALT | `[ERROR] Max instance count reached. How to fix: Delete unused instances or request quota increase.` |
| `OperationDenied.InstanceStatus` | 0 | — | HALT | `[ERROR] Operation not allowed in current instance state. How to fix: Wait for current operation to complete.` |
| `ResourceNotFound.Vpc` | 0 | — | HALT | `[ERROR] VPC not found. How to fix: Verify VPC ID exists in region.` |
| `InsufficientBalance` | 0 | — | HALT | `[ERROR] InsufficientBalance. How to fix: Recharge account.` |
| `Throttling` | 3 | exponential | Back off | `⚠️ Rate limit reached. Retrying...` |
| `InternalError` | 3 | 2s,4s,8s | Retry; HALT | `[ERROR] InternalError with RequestId: {RequestId}. How to fix: Retry or escalate.` |
| `ResourceAlreadyExists` | 0 | — | Ask reuse | `[ERROR] Instance name already exists. Use unique name or reuse.` |
| `InvalidParameter.Password` | 0 | — | HALT | `[ERROR] Invalid password format. How to fix: 8-32 chars, letters+digits+special chars.` |
| `Forbidden.RAM` | 0 | — | HALT | `[ERROR] Insufficient permissions. How to fix: Add RAM policy for Redis.` |
| `OperationDenied.DeletionProtection` | 0 | — | HALT | `[ERROR] Delete protection enabled. Disable first then retry.` |

### Operation: Describe/ List Instances

```bash
# Details
ve redis DescribeDBInstanceDetail --InstanceId "{{user.instance_id}}" --Region "{{env.VOLCENGINE_REGION}}"

# List all
ve redis DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}" --PageNumber 1 --PageSize 100
```

Present to user:

| Field | Path | Notes |
|-------|------|-------|
| Instance ID | `$.Result.InstanceId` | Primary identifier |
| Name | `$.Result.InstanceName` | Human-readable |
| Status | `$.Result.Status` | Running/Creating/Error |
| Engine Version | `$.Result.EngineVersion` | 4.0/5.0/6.0 |
| Capacity | `$.Result.Capacity.Total` | Memory in MB |
| Private Address | `$.Result.PrivateAddress` | Connection endpoint |
| VPC ID | `$.Result.VpcId` | Network context |

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

## Prerequisites

1. **`ve` CLI** installed per execution environment docs
2. **Go runtime** for JIT fallback (see references/integration.md)
3. **Credentials:** `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`, `VOLCENGINE_REGION`
4. **Verify:** `ve redis DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}"`

## Reference Directory

- [Core Concepts](references/core-concepts.md) — Redis architecture, instance types, engine versions
- [API & SDK Usage](references/api-sdk-usage.md) — Full operation map, JSON paths
- [CLI Usage](references/cli-usage.md) — `ve redis` command reference
- [Knowledge Base](references/knowledge-base.md) — fault pattern library (AIOps diagnosis)
- [Troubleshooting Guide](references/troubleshooting.md) — Error codes, diagnostics
- [Monitoring & Alerts](references/monitoring.md) — Redis monitoring metrics
- [Integration](references/integration.md) — Go SDK setup, JIT workflow

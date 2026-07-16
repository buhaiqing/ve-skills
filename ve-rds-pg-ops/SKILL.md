---
name: ve-rds-pg-ops
description: >-
  Use when the user needs to deploy, configure, troubleshoot, or monitor Volcengine
  RDS for PostgreSQL (云数据库 PostgreSQL 版) — DBInstance lifecycle, parameter
  management, account management, backup/restore, read-only nodes, and diagnostics.
  User mentions RDS PostgreSQL, PostgreSQL instance, Postgres database, 云数据库
  PostgreSQL, or describes product-specific scenarios (e.g., connection failures,
  WAL issues, instance creation, slow queries) even without naming the product
  directly. Not for billing, IAM, VPC networking, or application-level SQL schema.
license: MIT
compatibility: >-
  Official Volcengine CLI (`ve`, Go binary, no runtime), Go 1.14+ runtime
  (for JIT SDK fallback), valid API credentials, network access to Volcengine
  endpoints.
metadata:
  author: volcengine
  version: "1.1.0"
  last_updated: "2026-06-04"
  runtime: Harness AI Agent, Claude Code, Cursor, or compatible Agent runtimes
  go_version_minimum: "1.14"
  go_version_jit: "1.21+"
  api_profile: "RDS PostgreSQL OpenAPI 2022-01-01 — https://www.volcengine.com/docs/6438"
  cli_applicability: dual-path
  cli_support_evidence: >-
    RDS PostgreSQL is a core Volcengine service; `ve rds_postgresql --help` lists instance operations.
    OpenAPI service ID: rds_postgresql, version: 2022-01-01.
    Go SDK: github.com/volcengine/volc-sdk-golang/service/rds_postgresql.
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

# Volcengine RDS for PostgreSQL Operations Skill

### What This Skill Does

Manages Volcengine RDS for PostgreSQL instances including lifecycle, read-only nodes, backup/restore, parameter tuning, and account administration using the `ve` CLI (primary) or Go SDK (fallback).

## Overview

Volcengine RDS for PostgreSQL (云数据库 PostgreSQL 版) provides managed PostgreSQL database instances with support for versions 11-17. This skill is an **operational runbook** for agents: explicit scope, credential rules, pre-flight checks, **dual-path execution** (official **SDK/API** and official **`ve` CLI**), response validation, and failure recovery.

### CLI applicability

- **`cli_applicability: dual-path`:** `ve` CLI supports RDS PostgreSQL operations. Both SDK and CLI documented.

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | PostgreSQL-specific triggers; VPC/ECS delegated |
| 2 | **Structured I/O** | `{{env.*}}` for credentials, `{{user.*}}` for params, `{{output.*}}` for API responses |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | 15 PostgreSQL-specific error codes with HALT/retry per type |
| 5 | **Absolute Single Responsibility** | RDS PostgreSQL only; MySQL/VPC/ECS delegated |

## Trigger & Scope

### SHOULD Use This Skill When

- User mentions "RDS PostgreSQL", "PostgreSQL instance", "云数据库 PostgreSQL", "Postgres", "PG database"
- Task involves RDS PostgreSQL lifecycle (create, describe, modify, delete, list, restore)
- Task keywords: postgresql, postgres, pg, rds_pg, wal, vacuum, extension, pg_dump, 数据库实例
- User asks about PostgreSQL connection, configuration, spec changes, parameters, or read-only nodes

### SHOULD NOT Use This Skill When

- Billing / account → delegate to `ve-billing-ops`
- IAM / permissions → delegate to `ve-iam-ops`
- VPC / subnet → delegate to `ve-vpc-ops`
- RDS MySQL operations → use `ve-rds-mysql-ops` (different API service, different params)
- SQL schema design / DDL → application-level
- Console-only → state limitation

### Delegation Rules

- RDS requires VPC/subnets: verify via `ve-vpc-ops`
- Underlying compute: ECS issues delegate to `ve-ecs-ops`
- PostgreSQL-specific SQL extensions/application logic → not this skill

## Variable Convention

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | Runtime credential | NEVER ask user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | Runtime credential | NEVER ask user; fail if unset |
| `{{env.VOLCENGINE_REGION}}` | Runtime region | Use from env |
| `{{user.instance_name}}` | Instance name | Ask once |
| `{{user.instance_id}}` | Instance ID | Ask if not in context |
| `{{user.pg_version}}` | PG version (PostgreSQL_11 to PostgreSQL_17) | Ask with options |
| `{{user.primary_zone_id}}` | Primary node AZ | Ask with available zones |
| `{{user.secondary_zone_id}}` | Secondary node AZ | Ask with available zones |
| `{{user.storage_space}}` | Storage in GB (20-3000) | Ask with range |
| `{{output.instance_id}}` | From API response | Parse `$.Result.InstanceId` |

## API and Response Conventions

- **OpenAPI 2022-01-01** is canonical
- **API Version:** `2022-01-01`, **Service:** `rds_postgresql`
- **Endpoint:** `rds-postgresql.{region}.volcengineapi.com`

### Key Response Field Table

| Operation | JSON Path | Type | Description |
|-----------|-----------|------|-------------|
| CreateDBInstance | `$.Result.InstanceId` | string | New instance ID |
| DescribeDBInstances | `$.Result.Instances[].InstanceId` | array | Instance IDs |
| DescribeDBInstances | `$.Result.Instances[].InstanceName` | array | Instance names |
| DescribeDBInstances | `$.Result.Instances[].InstanceStatus` | array | RUNNING/CREATING/ERROR |
| DescribeDBInstances | `$.Result.Instances[].DbEngineVersion` | array | PostgreSQL_12, etc. |
| DescribeDBInstances | `$.Result.Instances[].StorageType` | array | LocalSSD |
| DescribeDBInstances | `$.Result.Instances[].StorageSpace` | array | Storage GB |
| DescribeDBInstances | `$.Result.Instances[].VpcId` | array | VPC ID |
| DescribeDBInstances | `$.Result.Instances[].NodeSpec` | array | e.g., rds.postgres.2c4g |
| DescribeDBInstances | `$.Result.TotalCount` | int | Total instances |
| DescribeDBInstanceDetail | `$.Result.Endpoints` | array | Connection strings |
| DescribeDBInstanceDetail | `$.Result.Nodes` | array | Primary/Secondary/ReadOnly nodes |
| DeleteDBInstance | `$.Metadata.RequestId` | string | Request correlation ID |

### Instance States

| State | Description |
|-------|-------------|
| `RUNNING` | Operational |
| `CREATING` | Being created |
| `DELETING` | Being deleted |
| `ERROR` | Error state |
| `RESTARTING` | Restarting |
| `MODIFYING` | Spec change in progress |
| `BACKING_UP` | Backup in progress |
| `RESTORING` | Restore in progress |
| `UPGRADING` | Engine version upgrade |

### State Transitions

| Operation | Initial | Target | Poll | Max Wait |
|-----------|---------|--------|------|----------|
| Create | -- | `RUNNING` | 10s | 600s |
| Delete | any | gone | 10s | 600s |
| Modify Spec | `RUNNING` | `RUNNING` | 10s | 900s |
| Restore | -- | `RUNNING` | 10s | 1800s |
| Rebuild | any | `RUNNING` | 10s | 900s |
| Restart | `RUNNING` | `RUNNING` | 5s | 300s |
| Add RO Node | `RUNNING` | `RUNNING` | 10s | 600s |

## Quick Start

### Prerequisites
- [ ] `ve` CLI installed
- [ ] `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`, `VOLCENGINE_REGION` set

### Verify Setup
```bash
ve rds_postgresql DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}"
```

### Your First Command
```bash
ve rds_postgresql DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}" --PageNumber 1 --PageSize 10
```
### Next Steps
- [Core Concepts](references/core-concepts.md) — Understand RDS for PostgreSQL architecture
- [Common Operations](#execution-flows) — Create, manage, and manage RDS PostgreSQL databases
- [Troubleshooting](references/troubleshooting.md) — Fix common issues

## Capabilities at a Glance

| Operation | Description | Level |
|-----------|-------------|-------|
| CreateDBInstance | Create PostgreSQL instance | Med/Low |
| DescribeDBInstanceDetail | Get instance details | Low/None |
| DescribeDBInstances | List all instances | Low/None |
| DeleteDBInstance | Delete instance | Low/**High** |
| ModifyDBInstanceSpec | Modify node spec/storage | Med/Med |
| DescribeDBInstanceParameters | Query parameters | Low/None |
| ModifyDBInstanceParameter | Modify parameters | Med/Med |
| DescribeRegions | List regions | Low/None |
| CreateReadonlyInstance | Create read-only instance | Med/Low |
| DescribeAccounts | List accounts | Low/None |
| CreateAccount | Create DB account | Low/Med |
| DescribeBackups | List backups | Low/None |
| CreateBackup | Create backup | Low/Low |
| RestoreToNewInstance | Restore from backup | High/Med |
| RebuildDBInstance | Rebuild instance | High/**High** |

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-16 | Initial release with RDS PostgreSQL lifecycle |
| 1.1.0 | 2026-06-04 | GCL rollout: added `## Quality Gate (GCL)`, references/rubric.md, references/prompt-templates.md |

## Quality Gate (GCL)

> Mandatory for every execution of `ve-rds-pg-ops`. Implements GCL per `../../AGENTS.md` §3-§9.

### Operation Tiers

> See [`references/rubric.md` §0](references/rubric.md#0-operation-tier) for the full operation tier table.

### Loop & Safety
- **DeleteDBInstance**: check deletion protection; warn irreversible data loss.
- **ModifyDBInstanceSpec**: warn 60-900s downtime.
- DB password masked in trace.

### Cross-skill delegation
| Finding | Delegate |
|---|---|
| VPC/subnet | `ve-vpc-ops` |
| Host-level | `ve-ecs-ops` |
| Billing | `ve-billing-ops` |

## Execution Flows

### Operation: Create PostgreSQL Instance (CreateDBInstance)

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| CLI | `ve version` | Exit 0 | JIT Go SDK |
| Credentials | Verify env vars | Set | HALT |
| Region | Supported PG region | Valid | HALT; DescribeRegions |
| VPC/Subnet | Network exists | Valid in region | HALT; ve-vpc-ops |
| Zones | Primary + secondary AZ available | Valid | HALT; DescribeAZs |
| Quota | Instance quota | Sufficient | HALT |

#### Execution — CLI (Primary)

```bash
ve rds_postgresql CreateDBInstance \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --DbEngineVersion "{{user.pg_version}}" \
  --NodeSpec "{{user.node_spec}}" \
  --PrimaryZoneId "{{user.primary_zone_id}}" \
  --SecondaryZoneId "{{user.secondary_zone_id}}" \
  --StorageSpace "{{user.storage_space}}" \
  --SubnetId "{{user.subnet_id}}" \
  --InstanceName "{{user.instance_name}}" \
  --ChargeInfo.ChargeType "PostPaid"
```

#### Execution — JIT Go SDK (Fallback)

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/rds_postgresql"
)

func main() {
    instance := rds_postgresql.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "DbEngineVersion": os.Getenv("PG_VERSION"),
        "NodeSpec":        os.Getenv("NODE_SPEC"),
        "PrimaryZoneId":   os.Getenv("PRIMARY_ZONE"),
        "SecondaryZoneId": os.Getenv("SECONDARY_ZONE"),
        "StorageSpace":    os.Getenv("STORAGE_SPACE"),
        "SubnetId":        os.Getenv("SUBNET_ID"),
        "InstanceName":    os.Getenv("INSTANCE_NAME"),
        "ChargeInfo":      map[string]interface{}{"ChargeType": "PostPaid"},
    }

    resp, err := instance.Client.Request("rds_postgresql", "CreateDBInstance", params)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(resp))
}
```

#### Post-execution Validation

```bash
for i in $(seq 1 60); do
  STATUS=$(ve rds_postgresql DescribeDBInstanceDetail --InstanceId "{{output.instance_id}}" | jq -r '.Result.InstanceStatus // ""')
  [ "$STATUS" = "RUNNING" ] && break
  [ "$STATUS" = "ERROR" ] && echo "Instance creation failed" && exit 1
  sleep 10
done
```

#### Failure Recovery

| Error Code | Action | Recovery |
|------------|--------|----------|
| `InvalidParameter.InstanceName` | HALT | 1-128 chars, no leading number/dash |
| `InvalidParameter.NodeSpec` | HALT | Check via DescribeInstanceSpecs |
| `InvalidParameter.StorageSpace` | HALT | [20-3000]GB, step 10GB |
| `InvalidParameter.ZoneConfig` | HALT | Verify zones exist |
| `InvalidParameter.NetworkConfig` | HALT | Verify VPC/subnet in region |
| `ResourceNotFound.Vpc` | HALT | Verify VPC ID |
| `QuotaExceeded.InstanceCount` | HALT | Delete unused or request increase |
| `OperationDenied.InstanceStatus` | HALT | Wait for current op to finish |
| `InsufficientBalance` | HALT | Recharge account |
| `Throttling` | Retry ×3, exponential | Back off |
| `InternalError` | Retry ×3, 2s/4s/8s | Report RequestId |
| `ResourceAlreadyExists` | HALT | Use unique name |
| `Forbidden.RAM` | HALT | Add RAM policy |
| `ResourceNotFound.Instance` | HALT | Verify InstanceId |
| `ResourceInUse` | HALT | Wait for concurrent op |

### Operation: Describe/ List Instances

```bash
# Details
ve rds_postgresql DescribeDBInstanceDetail --InstanceId "{{user.instance_id}}"

# List
ve rds_postgresql DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}" --PageNumber 1 --PageSize 100
```

| Field | Path | Notes |
|-------|------|-------|
| Instance ID | `$.Result.InstanceId` | Primary identifier |
| Name | `$.Result.InstanceName` | Human-readable |
| Status | `$.Result.InstanceStatus` | RUNNING/CREATING/ERROR |
| PG Version | `$.Result.DbEngineVersion` | PostgreSQL_12, etc. |
| Node Spec | `$.Result.NodeSpec` | rds.postgres.2c4g |
| Storage | `$.Result.StorageSpace` | GB |
| Endpoints | `$.Result.Endpoints[]` | Connection strings |
| Nodes | `$.Result.Nodes[]` | Primary/Secondary/ReadOnly |

### Operation: Delete Instance

#### Safety Gate
- Explicit confirmation required for `{{user.instance_name}}` (`{{user.instance_id}}`)
- MUST NOT proceed without user assent

```bash
ve rds_postgresql DeleteDBInstance --InstanceId "{{user.instance_id}}"
```

Poll until gone.

### Operation: Modify Instance Spec

```bash
ve rds_postgresql ModifyDBInstanceSpec --InstanceId "{{user.instance_id}}" --NodeSpec "{{user.new_spec}}" --StorageSpace "{{user.new_storage_gb}}"
```

Poll until `RUNNING` (up to 900s).

### Operation: Add Read-Only Instance

```bash
ve rds_postgresql CreateReadonlyInstance --SrcInstanceInstanceId "{{user.source_instance_id}}" --NodeSpec "{{user.ro_node_spec}}" --ZoneId "{{user.ro_zone_id}}"
```

### Operation: Parameter Management

```bash
# Query
ve rds_postgresql DescribeDBInstanceParameters --InstanceId "{{user.instance_id}}"

# Modify
ve rds_postgresql ModifyDBInstanceParameter --InstanceId "{{user.instance_id}}" --body '{"Parameters": [{"Name":"shared_buffers","Value":"2GB"}]}'
```

### Operation: Account Management

```bash
ve rds_postgresql DescribeAccounts --InstanceId "{{user.instance_id}}"
ve rds_postgresql CreateAccount --InstanceId "{{user.instance_id}}" --AccountName "{{user.account_name}}" --AccountPassword "{{user.password}}" --AccountType "Normal"
```

### Operation: Backup/Restore

```bash
ve rds_postgresql DescribeBackups --InstanceId "{{user.instance_id}}"
ve rds_postgresql CreateBackup --InstanceId "{{user.instance_id}}" --BackupName "{{user.backup_name}}"
ve rds_postgresql RestoreToNewInstance --body '{"BackupId":"{{user.backup_id}}","InstanceName":"{{user.new_name}}","NodeSpec":"{{user.spec}}","PrimaryZoneId":"{{user.zone}}"...}'
```

### Operation: Rebuild Instance

```bash
ve rds_postgresql RebuildDBInstance --InstanceId "{{user.instance_id}}"
```

## Error Taxonomy

| 错误码 | 含义 | 分辨率 |
|--------|------|--------|
| `InvalidDBInstanceId` | 实例 ID 不存在或格式错误 | 0 retries; **HALT** — 检查 InstanceId 格式 `(postgres-xxx)` |
| `IncorrectDBInstanceState` | 实例状态不匹配当前操作 | 0 retries; **HALT** — 等待实例回到 `RUNNING` 后重试 |
| `QuotaExceeded.InstanceCount` | 实例数量已达配额上限 | 0 retries; **HALT** — 删除闲置实例或提额 |
| `QuotaExceeded.StorageSize` | 存储空间已达配额上限 | 0 retries; **HALT** — 清理数据或提额 |
| `InvalidParameter.ConnectionMode` | 连接模式参数无效 | 0 retries; **HALT** — 修正为 `Standard`/`Session`/`Transaction` |
| `InvalidAccountName` | 账号名称不符合命名规范 | 0 retries; **HALT** — 小写字母开头，2-32 字符 |
| `InvalidDatabaseName` | 数据库名称不符合命名规范 | 0 retries; **HALT** — 字母开头，1-64 字符，不含特殊字符 |
| `InvalidVpcId` | VPC ID 不存在或不在当前区域 | 0 retries; **HALT** — 通过 `ve-vpc-ops` 确认 VPC |
| `InvalidSecurityGroupId` | 安全组 ID 不存在或格式错误 | 0 retries; **HALT** — 通过 `ve-security-group-ops` 确认安全组 |
| `BackupInProgress` | 备份操作正在进行中 | 0 retries; **HALT** — 等待当前备份完成 |
| `OperationDenied.InstanceStatus` | 实例当前状态不允许该操作 | 0 retries; **HALT** — 等待异步操作完成 |
| `Throttling` | 请求限流 | 3 retries/exponential/1s/2s/4s; **RETRY** |
| `InternalError` | 服务端内部错误 | 3 retries/exponential/2s/4s/8s; **RETRY** — 记录 RequestId |

## Prerequisites

1. **`ve` CLI** installed
2. **Go runtime** for JIT fallback (see references/integration.md)
3. **Credentials:** `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`, `VOLCENGINE_REGION`
4. **Verify:** `ve rds_postgresql DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}"`

## Operational Best Practices

- **Least privilege:** Grant IAM policies scoped to specific RDS PostgreSQL operations and instance IDs; avoid blanket `rds_postgresql:*` permissions.
- **Encrypted connections:** Enforce TLS for all client connections; disable SSL/TLS enforcement only during migration.
- **Backup strategy:** Configure automatic backups with a retention period of at least 7 days; perform manual backups before any spec modification or major version upgrade.
- **Multi-AZ deployment:** Deploy primary and secondary nodes in different availability zones for production workloads to ensure high availability.
- **Connection pooling:** Use PgBouncer or application-level pooling to manage connection counts and reduce overhead from frequent new connections.
- **Slow query monitoring:** Enable logging of slow queries via `log_min_duration_statement` (set to 1000ms or lower); correlate with parameter tuning via `DescribeDBInstanceParameters`.
- **Instance sizing:** Right-size based on observed CPU, memory, and IOPS metrics; use `ModifyDBInstanceSpec` for incremental scaling rather than upfront over-provisioning.

## Reference Directory

- [Core Concepts](references/core-concepts.md) — Architecture, engine versions, node specs
- [API & SDK Usage](references/api-sdk-usage.md) — Full operation map, JSON paths
- [CLI Usage](references/cli-usage.md) — `ve rds_postgresql` command reference
- [Knowledge Base](references/knowledge-base.md) — fault pattern library (AIOps diagnosis)
- [Troubleshooting Guide](references/troubleshooting.md) — Error codes, diagnostics
- [Monitoring & Alerts](references/monitoring.md) — RDS PG monitoring
- [Integration](references/integration.md) — Go SDK setup, JIT workflow
- [GCL Rubric](references/rubric.md)
- [GCL Prompt Templates](references/prompt-templates.md)

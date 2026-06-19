---
name: ve-rds-mysql-ops
description: >-
  Use when the user needs to deploy, configure, troubleshoot, or monitor Volcengine
  RDS for MySQL (云数据库 MySQL 版) — DBInstance lifecycle, parameter management,
  account management, backup/restore, and diagnostics. User mentions RDS MySQL,
  云数据库, MySQL instance, database, or describes product-specific scenarios
  (e.g., connection failures, slow queries, instance creation issues) even without
  naming the product directly. Not for billing, IAM, VPC networking, or application
  database schema management.
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
  api_profile: "RDS MySQL OpenAPI 2022-01-01 — https://www.volcengine.com/docs/6313"
  cli_applicability: dual-path
  cli_support_evidence: >-
    RDS MySQL is a core Volcengine service; `ve rds_mysql --help` lists instance operations.
    OpenAPI service ID: rds_mysql, version: 2022-01-01.
    Go SDK: github.com/volcengine/volc-sdk-golang/service/rds_mysql.
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

# Volcengine RDS for MySQL Operations Skill

## Overview

Volcengine RDS for MySQL (云数据库 MySQL 版) provides managed MySQL database instances. This skill is an **operational runbook** for agents: explicit scope, credential rules, pre-flight checks, **dual-path execution** (official **SDK/API** and official **`ve` CLI**), response validation, and failure recovery.

### CLI applicability

- **`cli_applicability: dual-path`:** `ve` CLI supports RDS MySQL operations. Document **both** SDK and CLI steps.

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | RDS MySQL triggers; VPC/ECS delegated to other skills |
| 2 | **Structured I/O** | `{{env.*}}` for credentials, `{{user.*}}` for params, `{{output.*}}` for API responses |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | 15 RDS-specific error codes with HALT/retry per type |
| 5 | **Absolute Single Responsibility** | RDS MySQL only; VPC/ECS/SQL schema delegated |

## Trigger & Scope

### SHOULD Use This Skill When

- User mentions "RDS MySQL", "云数据库", "MySQL instance", "RDS for MySQL", "数据库实例"
- Task involves RDS instance lifecycle (create, describe, modify, delete, list, restore)
- Task keywords: rds, mysql, database, instance, backup, restore, account, parameter, slow query
- User asks about RDS connection, configuration changes, spec modifications, or parameter tuning

### SHOULD NOT Use This Skill When

- Billing / account → delegate to `ve-billing-ops`
- IAM / permissions → delegate to `ve-iam-ops`
- VPC / subnet → delegate to `ve-vpc-ops`
- SQL schema design / DDL operations → application-level, not infrastructure
- Console-only flows → state limitation

### Delegation Rules

- RDS requires VPC/subnets: verify via `ve-vpc-ops`
- Underlying compute is ECS: host issues delegate to `ve-ecs-ops`

## Variable Convention

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| 🔴 `{{env.VOLCENGINE_ACCESS_KEY}}` | Runtime credential | NEVER ask user; fail if unset |
| 🔴 `{{env.VOLCENGINE_SECRET_KEY}}` | Runtime credential | NEVER ask user; fail if unset |
| `{{env.VOLCENGINE_REGION}}` | Runtime region | Use from env |
| `{{user.instance_name}}` | Instance name | Ask once |
| `{{user.instance_id}}` | Instance ID | Ask if not in context |
| `{{user.db_version}}` | MySQL version (5.7/8.0) | Ask with options |
| `{{user.storage_space}}` | Storage in GB (20-3000) | Ask with range |
| `{{user.storage_type}}` | LocalSSD or ESSD | Ask with options |
| `{{output.instance_id}}` | From API response | Parse `$.Result.InstanceId` |
| `{{output.account_name}}` | From account API | Parse `$.Result.AccountName` |

> **Security:** NEVER log, print, or expose `VOLCENGINE_SECRET_KEY` or any DB password.

## API and Response Conventions

- **OpenAPI 2022-01-01** is canonical for all fields and response shapes
- **API Version:** `2022-01-01`, **Service:** `rds_mysql`
- **Endpoint:** `rds-mysql.{region}.volcengineapi.com`

### Key Response Field Table

| Operation | JSON Path | Type | Description |
|-----------|-----------|------|-------------|
| CreateDBInstance | `$.Result.InstanceId` | string | New instance ID |
| DescribeDBInstances | `$.Result.Instances[].InstanceId` | array | Instance IDs |
| DescribeDBInstances | `$.Result.Instances[].InstanceName` | array | Instance names |
| DescribeDBInstances | `$.Result.Instances[].InstanceStatus` | array | Instance states |
| DescribeDBInstances | `$.Result.Instances[].DBEngineVersion` | array | MySQL versions |
| DescribeDBInstances | `$.Result.Instances[].StorageType` | array | LocalSSD/ESSD |
| DescribeDBInstances | `$.Result.Instances[].StorageSpace` | array | Storage GB |
| DescribeDBInstances | `$.Result.Instances[].VpcId` | array | VPC ID |
| DescribeDBInstances | `$.Result.Instances[].ZoneId` | array | Availability zone |
| DescribeDBInstances | `$.Result.Instances[].ChargeType` | array | PostPaid/PrePaid |
| DescribeDBInstanceParameters | `$.Result.Parameters[].ParameterName` | array | Parameter names |
| DescribeDBInstanceParameters | `$.Result.Parameters[].ParameterValue` | array | Current values |
| DescribeDBInstanceParameters | `$.Result.Parameters[].ForceRestart` | array | Requires restart |
| DescribeDBInstanceParameters | `$.Result.Parameters[].CheckingCode` | array | Valid value range |

### Instance States

| State | Description |
|-------|-------------|
| `RUNNING` | Operational |
| `CREATING` | Being created |
| `DELETING` | Being deleted |
| `DELETED` | Deleted |
| `ERROR` | Error state |
| `RESTARTING` | Restarting |
| `MODIFYING` | Spec change in progress |
| `BACKING_UP` | Backup in progress |
| `RESTORING` | Restore in progress |

### State Transitions

| Operation | Initial | Target | ⏱ Poll | ⏱ Max Wait |
|-----------|---------|--------|------|----------|
| Create | — | `RUNNING` | 10s | 600s |
| Delete | any | gone | 10s | 600s |
| Modify Spec | RUNNING | `RUNNING` | 10s | 900s |
| Restore | — | `RUNNING` | 10s | 1800s |
| Rebuild | any | `RUNNING` | 10s | 900s |
| Restart | RUNNING | `RUNNING` | 5s | 300s |

## Quick Start

### Prerequisites
- [ ] `ve` CLI installed
- [ ] `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`, `VOLCENGINE_REGION` set

### Verify Setup
```bash
ve rds_mysql DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}"
```

### Your First Command
```bash
ve rds_mysql DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}" --PageNumber 1 --PageSize 10
```

## Capabilities at a Glance

| Operation | Description | ⚡ Complexity | 🛡️ Risk |
|-----------|-------------|------------|------|
| CreateDBInstance | Create MySQL instance | Medium | Low |
| DescribeDBInstanceDetail | Get instance details | Low | ✅ None |
| DescribeDBInstances | List all instances | Low | ✅ None |
| DeleteDBInstance | Delete instance | Low | 🔴 **High** |
| ModifyDBNodeSpec | Modify node spec/storage | Medium | Medium |
| DescribeDBInstanceParameters | Query parameters | Low | ✅ None |
| ModifyDBInstanceParameter | Modify parameters | Medium | Medium |
| DescribeRegions | List regions | Low | ✅ None |
| DescribeAvailabilityZones | List AZs | Low | ✅ None |
| ListDBInstanceIPLists | List IP whitelist | Low | ✅ None |
| ModifyDBInstanceIPList | Modify IP whitelist | Low | Low |
| DescribeDBAccounts | List accounts | Low | ✅ None |
| CreateDBAccount | Create DB account | Low | Medium |
| DeleteDBAccount | Delete account | Low | Medium |
| DescribeBackups | List backups | Low | ✅ None |
| CreateBackup | Create backup | Low | Low |
| RestoreToNewInstance | Restore from backup | High | Medium |
| RebuildDBInstance | Rebuild instance | High | 🔴 **High** |

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-16 | Initial release with RDS MySQL lifecycle |
| 1.1.0 | 2026-06-04 | Phase 1 GCL rollout: added `## Quality Gate (GCL)` chapter, `references/rubric.md`, `references/prompt-templates.md` |

## Quality Gate (GCL)

> This chapter is **mandatory** for every execution of `ve-rds-mysql-ops`. It implements
> the Generator-Critic-Loop defined in `../../AGENTS.md` §3-§9. Read
> [`references/rubric.md`](references/rubric.md) for scoring and
> [`references/prompt-templates.md`](references/prompt-templates.md) for safety prompts.

### Operation Tiers

| Tier | Operations | `max_iter` | Safety floor |
|---|---|---|---|
| 🔴 **Destructive** | `DeleteDBInstance`, `DeleteDBAccount`, `RebuildDBInstance` | 2 | 1.0 (mandatory) |
| 🟡 **State-changing** | `ModifyDBNodeSpec`, `ModifyDBInstanceParameter`, `ModifyDBInstanceIPList` | 2 | 1.0 (mandatory) |
| **Mutating** | `CreateDBInstance`, `CreateDBAccount`, `CreateBackup`, `RestoreToNewInstance` | 2 | ≥ 0.5 |
| ✅ **Read-only** | `DescribeDBInstanceDetail`, `DescribeDBInstances`, `DescribeDBInstanceParameters`, `DescribeRegions`, `DescribeAvailabilityZones`, `ListDBInstanceIPLists`, `DescribeDBAccounts`, `DescribeBackups` | 3 | ≥ 0 |

### Loop

1. **Pre-flight** — resolve `{{env.*}}` / `{{user.*}}`; classify tier; load rubric.
2. **Generate** — execute per `## Execution Flows`. Trace to `./audit-results/gcl-trace-*.json`.
3. **Critique** — isolated prompt; score 5 dimensions; MUST NOT see raw request.
4. **Decide** — Safety=0 on Destructive/State-changing → **ABORT**; all pass → return; `iter<max_iter` → loop.

### RDS-specific safety rules

- 🔴 **DeleteDBInstance**: check deletion protection; warn about irreversible data loss.
- 🔴 **RebuildDBInstance**: warn that the instance will be rebuilt from its initial snapshot — any data changes since creation are lost.
- 🟡 **ModifyDBNodeSpec**: warn about 60-900s downtime.
- 🟡 **DeleteDBAccount**: warn that applications using this account lose access.
- 🟡 **ModifyDBInstanceParameter**: if `ForceRestart=true`, warn about restart.
- 🟡 **ModifyDBInstanceIPList** on production: warn about locking out clients.
- DB password masked as `<masked>` in trace.

### Trace

`./audit-results/gcl-trace-*.json` — password masked.

### Cross-skill delegation

| Critic finding | Delegate to |
|---|---|
| VPC/subnet not found | `ve-vpc-ops` |
| Host-level issue | `ve-ecs-ops` |
| Billing/quota | `ve-billing-ops` |

## Execution Flows

### Operation: Create MySQL Instance (CreateDBInstance)

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| CLI | `ve version` | Exit 0 | Go to JIT Go SDK |
| 🔴 Credentials | Verify env vars | Set and valid | HALT |
| 🔴 Region | Supported RDS region | Valid | HALT; list regions via DescribeRegions |
| 🔴 VPC/Subnet | Network exists | Valid in region | HALT; use ve-vpc-ops |
| 🔴 Quota | Instance quota | Sufficient | HALT; raise quota |

#### Execution — CLI (Primary)

```bash
ve rds_mysql CreateDBInstance \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_name}}" \
  --DBEngine "Mysql" \
  --DBEngineVersion "{{user.db_version}}" \
  --NodeInfo '[{"NodeType":"Primary","NodeSpec":"{{user.node_spec}}","ZoneId":"{{user.zone_id}}"},{"NodeType":"Secondary","NodeSpec":"{{user.node_spec}}","ZoneId":"{{user.zone_id2}}"}]' \
  --StorageType "{{user.storage_type}}" \
  --StorageSpace "{{user.storage_space}}" \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --ChargeType "PostPaid" \
  --InstanceName "{{user.instance_name}}"
```

#### Execution — JIT Go SDK (Fallback)

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/rds_mysql"
)

func main() {
    instance := rds_mysql.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "Region":         os.Getenv("VOLCENGINE_REGION"),
        "DBEngine":       "Mysql",
        "DBEngineVersion": os.Getenv("DB_VERSION"),
        "StorageType":    os.Getenv("STORAGE_TYPE"),
        "StorageSpace":   os.Getenv("STORAGE_SPACE"),
        "VpcId":          os.Getenv("VPC_ID"),
        "SubnetId":       os.Getenv("SUBNET_ID"),
        "ChargeType":     "PostPaid",
        "InstanceName":   os.Getenv("INSTANCE_NAME"),
    }

    resp, err := instance.Client.Request("rds_mysql", "CreateDBInstance", params)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(resp))
}
```

#### Post-execution Validation

```bash
for i in $(seq 1 60); do
  STATUS=$(ve rds_mysql DescribeDBInstanceDetail --InstanceId "{{output.instance_id}}" | jq -r '.Result.InstanceStatus // ""')
  [ "$STATUS" = "RUNNING" ] && break
  [ "$STATUS" = "ERROR" ] && echo "Instance creation failed" && exit 1
  sleep 10
done
```

#### Failure Recovery

| Error Pattern | Agent Action | Recovery |
|---------------|-------------|----------|
| `InvalidParameter.InstanceName` | **HALT** | Use valid name format |
| `InvalidParameter.NodeSpec` | **HALT** | Check available specs via DescribeDBInstanceSpecs |
| `InvalidParameter.StorageSpace` | **HALT** | Provide valid storage size [20-3000]GB |
| `InvalidParameter.NetworkConfig` | **HALT** | Verify network exists in region |
| `ResourceNotFound.Vpc` | **HALT** | VPC not found; verify VPC ID |
| `QuotaExceeded.InstanceCount` | **HALT** | Delete unused instances or request quota increase |
| `OperationDenied.InstanceStatus` | **HALT** | Wait for current operation to complete |
| `InsufficientBalance` | **HALT** | Recharge account |
| `Throttling` | Retry 3x, exponential backoff | Rate limit reached; retry with backoff |
| `InternalError` | Retry 3x with backoff (2s,4s,8s); **HALT** after 3 | Capture RequestId; retry or escalate |
| `ResourceAlreadyExists` | Ask reuse | Use unique instance name |
| `Forbidden.RAM` | **HALT** | Add RAM policy for RDS MySQL |
| `InvalidParameter.Parameter` | **HALT** | Fix value to match CheckingCode range |
| `ResourceNotFound.Instance` | **HALT** | Verify instance ID |
| `ResourceInUse` | **HALT** | Wait for current operation to complete |

### Operation: Describe/ List Instances

```bash
# Details
ve rds_mysql DescribeDBInstanceDetail --InstanceId "{{user.instance_id}}"

# List
ve rds_mysql DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}" --PageNumber 1 --PageSize 100
```

| Field | 🔍 Path | Notes |
|-------|------|-------|
| Instance ID | `$.Result.InstanceId` | Primary identifier |
| Name | `$.Result.InstanceName` | Human-readable |
| Status | `$.Result.InstanceStatus` | RUNNING/CREATING/ERROR |
| Engine Version | `$.Result.DBEngineVersion` | MySQL_5_7 or MySQL_8_0 |
| Storage Type | `$.Result.StorageType` | LocalSSD/ESSD |
| Storage Space | `$.Result.StorageSpace` | GB |
| VPC | `$.Result.VpcId` | Network |
| Connection | `$.Result.ConnectionStrings` | Endpoint |

### Operation: Delete Instance

#### Safety Gate
- Explicit confirmation required for delete of `{{user.instance_name}}` (`{{user.instance_id}}`)
- **MUST NOT** proceed without user assent

#### Execution
```bash
ve rds_mysql DeleteDBInstance --InstanceId "{{user.instance_id}}"
```

#### Validation
Poll until gone or status = `DELETED`.

### Operation: Modify Node Spec (ModifyDBNodeSpec)

```bash
ve rds_mysql ModifyDBNodeSpec --InstanceId "{{user.instance_id}}" --StorageSpace "{{user.new_storage_gb}}" --body '{"NodeInfo": [{"NodeType":"Primary","NodeSpec":"{{user.new_spec}}","NodeOperateType":"Modify"},{"NodeType":"Secondary","NodeSpec":"{{user.new_spec}}","NodeOperateType":"Modify"}]}'
```

Poll until `RUNNING` (may take up to 900s).

### Operation: Parameter Management

```bash
# Query
ve rds_mysql DescribeDBInstanceParameters --InstanceId "{{user.instance_id}}"

# Modify (note: some require restart)
ve rds_mysql ModifyDBInstanceParameter --InstanceId "{{user.instance_id}}" --body '{"Parameters": [{"ParameterName":"max_connections","ParameterValue":"2000"}]}'
```

### Operation: Account Management

```bash
ve rds_mysql DescribeDBAccounts --InstanceId "{{user.instance_id}}"
ve rds_mysql CreateDBAccount --InstanceId "{{user.instance_id}}" --AccountName "{{user.account_name}}" --AccountPassword "{{user.password}}" --AccountPrivilege "ReadWrite"
ve rds_mysql DeleteDBAccount --InstanceId "{{user.instance_id}}" --AccountName "{{user.account_name}}"
```

### Operation: Backup Management

```bash
ve rds_mysql DescribeBackups --InstanceId "{{user.instance_id}}"
ve rds_mysql CreateBackup --InstanceId "{{user.instance_id}}" --BackupName "{{user.backup_name}}"
```

### Operation: Restore Instance

```bash
ve rds_mysql RestoreToNewInstance --body '{"BackupId":"{{user.backup_id}}","InstanceName":"{{user.new_instance_name}}","NodeInfo": [...]}'
```

Poll until `RUNNING` (may take up to 1800s).

### Operation: Rebuild Instance

```bash
ve rds_mysql RebuildDBInstance --InstanceId "{{user.instance_id}}"
```

### Operation: IP Whitelist Management

```bash
ve rds_mysql ListDBInstanceIPLists --InstanceId "{{user.instance_id}}"
ve rds_mysql ModifyDBInstanceIPList --InstanceId "{{user.instance_id}}" --body '{"IPList": ["10.0.0.0/8"], "ModifyMode": "Cover"}'
```

## Prerequisites

1. **`ve` CLI** installed per execution environment
2. **Go runtime** for JIT fallback (see references/integration.md)
3. **Credentials:** `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`, `VOLCENGINE_REGION`
4. **Verify:** `ve rds_mysql DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}"`

## Reference Directory

- [Core Concepts](references/core-concepts.md) — RDS architecture, engine versions, node types
- [API & SDK Usage](references/api-sdk-usage.md) — Full operation map, JSON paths
- [CLI Usage](references/cli-usage.md) — `ve rds_mysql` command reference
- [Knowledge Base](references/knowledge-base.md) — fault pattern library (AIOps diagnosis)
- [Troubleshooting Guide](references/troubleshooting.md) — Error codes, diagnostics
- [Monitoring & Alerts](references/monitoring.md) — RDS monitoring metrics
- [Integration](references/integration.md) — Go SDK setup, JIT workflow
- [GCL Rubric](references/rubric.md) — Scoring dimensions for the Generator-Critic-Loop
- [GCL Prompt Templates](references/prompt-templates.md) — G/C/O prompt skeletons + RDS-specific safety prompts

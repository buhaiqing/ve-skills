---
name: ve-mongodb-ops
description: >-
  Use when the user needs to deploy, configure, troubleshoot, or monitor Volcengine
  MongoDB (云数据库 MongoDB 版) — instance lifecycle, database/collection management,
  user management, backup/restore, network configuration, monitoring, and scaling.
  User mentions MongoDB, 云数据库 MongoDB, document database, or describes
  MongoDB-specific scenarios (e.g., replica set, sharding, connection issues,
  slow queries) even without naming the product directly. Not for billing, IAM,
  VPC networking, or application-level schema design.
license: MIT
compatibility: >-
  Official Volcengine CLI (`ve`, Go binary, no runtime), Go 1.14+ runtime
  (for JIT SDK fallback), valid API credentials, network access to Volcengine
  endpoints.
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-05-27"
  runtime: Harness AI Agent, Claude Code, Cursor, or compatible Agent runtimes
  go_version_minimum: "1.14"
  go_version_jit: "1.21+"
  api_profile: "MongoDB OpenAPI 2022-01-01 — https://www.volcengine.com/docs/6314"
  cli_applicability: dual-path
  cli_support_evidence: >-
    MongoDB is a core Volcengine service; `ve mongodb --help` lists instance operations.
    OpenAPI service ID: mongodb, version: 2022-01-01.
    Go SDK: github.com/volcengine/volc-sdk-golang/service/mongodb.
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

# Volcengine MongoDB Operations Skill

## Overview

Volcengine MongoDB (云数据库 MongoDB 版) provides managed MongoDB instances with replica sets and sharded cluster architectures. This skill is an **operational runbook** for agents: explicit scope, credential rules, pre-flight checks, **dual-path execution** (official **SDK/API** and official **`ve` CLI**), response validation, and failure recovery.

### CLI applicability

- **`cli_applicability: dual-path`:** `ve` CLI supports MongoDB operations. Document **both** SDK and CLI steps.

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | MongoDB triggers; VPC/ECS delegated to other skills |
| 2 | **Structured I/O** | `{{env.*}}` for credentials, `{{user.*}}` for params, `{{output.*}}` for API responses |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | 15 MongoDB-specific error codes with HALT/retry per type |
| 5 | **Absolute Single Responsibility** | MongoDB only; VPC/ECS/SQL schema delegated |

## Trigger & Scope

### SHOULD Use This Skill When

- User mentions "MongoDB", "云数据库 MongoDB", "document database", "Mongo instance"
- Task involves MongoDB instance lifecycle (create, describe, modify, delete, list, restore)
- Task keywords: mongodb, mongo, replica set, sharding, instance, backup, restore, user, database, collection
- User asks about MongoDB connection, configuration changes, spec modifications, or parameter tuning
- Task involves database/collection operations (create, drop, list)
- Task involves MongoDB user management (create, delete, modify privileges)

### SHOULD NOT Use This Skill When

- Billing / account → delegate to `ve-billing-ops`
- IAM / permissions → delegate to `ve-iam-ops`
- VPC / subnet → delegate to `ve-vpc-ops`
- Application-level MongoDB schema design → application-level, not infrastructure
- Console-only flows → state limitation

### Delegation Rules

- MongoDB requires VPC/subnets: verify via `ve-vpc-ops`
- Underlying compute is ECS: host issues delegate to `ve-ecs-ops`

## Variable Convention

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | Runtime credential | NEVER ask user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | Runtime credential | NEVER ask user; fail if unset |
| `{{env.VOLCENGINE_REGION}}` | Runtime region | Use from env |
| `{{user.instance_name}}` | Instance name | Ask once |
| `{{user.instance_id}}` | Instance ID | Ask if not in context |
| `{{user.mongo_version}}` | MongoDB version (4.0/4.2/4.4/5.0/6.0) | Ask with options |
| `{{user.storage_gb}}` | Storage in GB (20-3000) | Ask with range |
| `{{user.node_spec}}` | Node specification | Ask with options |
| `{{user.vpc_id}}` | VPC ID | Ask if not in context |
| `{{user.subnet_id}}` | Subnet ID | Ask if not in context |
| `{{user.database_name}}` | Database name | Ask once |
| `{{user.collection_name}}` | Collection name | Ask once |
| `{{user.username}}` | MongoDB username | Ask once |
| `{{user.password}}` | MongoDB password | Ask once (masked input) |
| `{{output.instance_id}}` | From API response | Parse `$.Result.InstanceId` |
| `{{output.connection_string}}` | From API response | Parse `$.Result.ConnectionString` |

> **Security:** NEVER log, print, or expose `VOLCENGINE_SECRET_KEY` or any DB password.

## API and Response Conventions

- **OpenAPI 2022-01-01** is canonical for all fields and response shapes
- **API Version:** `2022-01-01`, **Service:** `mongodb`
- **Endpoint:** `mongodb.{region}.volcengineapi.com`

### Key Response Field Table

| Operation | JSON Path | Type | Description |
|-----------|-----------|------|-------------|
| CreateDBInstance | `$.Result.InstanceId` | string | New instance ID |
| DescribeDBInstances | `$.Result.Instances[].InstanceId` | array | Instance IDs |
| DescribeDBInstances | `$.Result.Instances[].InstanceName` | array | Instance names |
| DescribeDBInstances | `$.Result.Instances[].InstanceStatus` | array | Instance states |
| DescribeDBInstances | `$.Result.Instances[].MongoVersion` | array | MongoDB versions |
| DescribeDBInstances | `$.Result.Instances[].NodeSpec` | array | Node specifications |
| DescribeDBInstances | `$.Result.Instances[].StorageSpaceGB` | array | Storage GB |
| DescribeDBInstances | `$.Result.Instances[].VpcId` | array | VPC ID |
| DescribeDBInstances | `$.Result.Instances[].ZoneId` | array | Availability zone |
| DescribeDBInstanceDetail | `$.Result.ConnectionString` | string | Connection endpoint |
| DescribeDBInstanceDetail | `$.Result.Port` | integer | Connection port |
| DescribeDBAccounts | `$.Result.Accounts[].AccountName` | array | Account names |
| DescribeDBAccounts | `$.Result.Accounts[].AccountPrivilege` | array | Account privileges |
| DescribeDatabases | `$.Result.Databases[].DBName` | array | Database names |
| DescribeBackups | `$.Result.Backups[].BackupId` | array | Backup IDs |

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
| `CONFIG_CHANGING` | Configuration changing |

### State Transitions

| Operation | Initial | Target | Poll | Max Wait |
|-----------|---------|--------|------|----------|
| Create | — | `RUNNING` | 10s | 600s |
| Delete | any | gone | 10s | 600s |
| Modify Spec | RUNNING | `RUNNING` | 10s | 900s |
| Restart | RUNNING | `RUNNING` | 5s | 300s |
| Restore | — | `RUNNING` | 10s | 1800s |

## Quick Start

### Prerequisites
- [ ] `ve` CLI installed
- [ ] `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`, `VOLCENGINE_REGION` set

### Verify Setup
```bash
ve mongodb DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}"
```

### Your First Command
```bash
ve mongodb DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}" --PageNumber 1 --PageSize 10
```

## Capabilities at a Glance

| Operation | Description | Complexity | Risk |
|-----------|-------------|------------|------|
| CreateDBInstance | Create MongoDB instance | Medium | Low |
| DescribeDBInstanceDetail | Get instance details | Low | None |
| DescribeDBInstances | List all instances | Low | None |
| DeleteDBInstance | Delete instance | Low | **High** |
| ModifyDBInstanceSpec | Modify spec/storage | Medium | Medium |
| RestartDBInstance | Restart instance | Low | Medium |
| DescribeDBAccounts | List accounts | Low | None |
| CreateDBAccount | Create account | Low | Medium |
| DeleteDBAccount | Delete account | Low | Medium |
| ModifyDBAccountPrivilege | Modify privileges | Low | Medium |
| DescribeDatabases | List databases | Low | None |
| CreateDatabase | Create database | Low | Low |
| DeleteDatabase | Delete database | Low | **High** |
| DescribeCollections | List collections | Low | None |
| CreateCollection | Create collection | Low | Low |
| DeleteCollection | Delete collection | Low | **High** |
| DescribeBackups | List backups | Low | None |
| CreateBackup | Create backup | Low | Low |
| RestoreDBInstance | Restore from backup | High | Medium |
| ModifyDBInstanceIPList | Modify IP whitelist | Low | Low |
| DescribeDBInstanceParameters | Query parameters | Low | None |
| ModifyDBInstanceParameters | Modify parameters | Medium | Medium |

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-27 | Initial release with MongoDB lifecycle, users, databases, collections, backups |

## Execution Flows

### Operation: Create MongoDB Instance (CreateDBInstance)

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| CLI | `ve version` | Exit 0 | Go to JIT Go SDK |
| Credentials | Verify env vars | Set and valid | HALT |
| Region | Supported MongoDB region | Valid | HALT; list regions |
| VPC/Subnet | Network exists | Valid in region | HALT; use ve-vpc-ops |
| Quota | Instance quota | Sufficient | HALT; raise quota |

#### Execution — CLI (Primary)

```bash
ve mongodb CreateDBInstance \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceName "{{user.instance_name}}" \
  --MongoVersion "{{user.mongo_version}}" \
  --NodeSpec "{{user.node_spec}}" \
  --StorageSpaceGB "{{user.storage_gb}}" \
  --NodeNumber 3 \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --ChargeType "PostPaid"
```

#### Execution — JIT Go SDK (Fallback)

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/mongodb"
)

func main() {
    instance := mongodb.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "Region":         os.Getenv("VOLCENGINE_REGION"),
        "InstanceName":   os.Getenv("INSTANCE_NAME"),
        "MongoVersion":   os.Getenv("MONGO_VERSION"),
        "NodeSpec":       os.Getenv("NODE_SPEC"),
        "StorageSpaceGB": os.Getenv("STORAGE_GB"),
        "NodeNumber":     3,
        "VpcId":          os.Getenv("VPC_ID"),
        "SubnetId":       os.Getenv("SUBNET_ID"),
        "ChargeType":     "PostPaid",
    }

    resp, err := instance.Client.Request("mongodb", "CreateDBInstance", params)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(resp))
}
```

#### Post-execution Validation

```bash
for i in $(seq 1 60); do
  STATUS=$(ve mongodb DescribeDBInstanceDetail --InstanceId "{{output.instance_id}}" | jq -r '.Result.InstanceStatus // ""')
  [ "$STATUS" = "RUNNING" ] && break
  [ "$STATUS" = "ERROR" ] && echo "Instance creation failed" && exit 1
  sleep 10
done
```

#### Failure Recovery

| Error Pattern | Retries | Backoff | Agent Action | UX Feedback |
|---------------|---------|---------|--------------|-------------|
| `InvalidParameter.InstanceName` | 0 | — | HALT | `[ERROR] Invalid instance name. Use valid name format.` |
| `InvalidParameter.NodeSpec` | 0 | — | HALT | `[ERROR] Invalid node spec. Check available specs.` |
| `InvalidParameter.StorageSpace` | 0 | — | HALT | `[ERROR] Storage out of range [20-3000]GB.` |
| `InvalidParameter.MongoVersion` | 0 | — | HALT | `[ERROR] Invalid MongoDB version. Valid: 4.0, 4.2, 4.4, 5.0, 6.0` |
| `InvalidParameter.NetworkConfig` | 0 | — | HALT | `[ERROR] Invalid VPC/subnet. Verify network exists.` |
| `ResourceNotFound.Vpc` | 0 | — | HALT | `[ERROR] VPC not found.` |
| `QuotaExceeded.InstanceCount` | 0 | — | HALT | `[ERROR] Max instances reached. Delete unused or request quota.` |
| `OperationDenied.InstanceStatus` | 0 | — | HALT | `[ERROR] Cannot operate in current state. Wait.` |
| `InsufficientBalance` | 0 | — | HALT | `[ERROR] Insufficient balance. Recharge account.` |
| `Throttling` | 3 | exponential | Back off | `⚠️ Rate limit. Retrying...` |
| `InternalError` | 3 | 2s,4s,8s | Retry; HALT | `[ERROR] InternalError with RequestId: {RequestId}.` |
| `ResourceAlreadyExists` | 0 | — | Ask reuse | `[ERROR] Instance name exists. Use unique name.` |
| `Forbidden.RAM` | 0 | — | HALT | `[ERROR] Insufficient permissions. Add RAM policy.` |
| `InvalidParameter.Parameter` | 0 | — | HALT | `[ERROR] Parameter value invalid. Fix value.` |
| `ResourceNotFound.Instance` | 0 | — | HALT | `[ERROR] Instance not found. Verify ID.` |

### Operation: Describe/List Instances

```bash
# Details
ve mongodb DescribeDBInstanceDetail --InstanceId "{{user.instance_id}}"

# List
ve mongodb DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}" --PageNumber 1 --PageSize 100
```

| Field | Path | Notes |
|-------|------|-------|
| Instance ID | `$.Result.InstanceId` | Primary identifier |
| Name | `$.Result.InstanceName` | Human-readable |
| Status | `$.Result.InstanceStatus` | RUNNING/CREATING/ERROR |
| Mongo Version | `$.Result.MongoVersion` | 4.0/4.2/4.4/5.0/6.0 |
| Node Spec | `$.Result.NodeSpec` | Instance spec |
| Storage | `$.Result.StorageSpaceGB` | GB |
| VPC | `$.Result.VpcId` | Network |
| Connection | `$.Result.ConnectionString` | Endpoint |
| Port | `$.Result.Port` | Port (default 27017) |

### Operation: Delete Instance

#### Safety Gate
- Explicit confirmation required for delete of `{{user.instance_name}}` (`{{user.instance_id}}`)
- **MUST NOT** proceed without user assent

#### Execution
```bash
ve mongodb DeleteDBInstance --InstanceId "{{user.instance_id}}"
```

#### Validation
Poll until gone or status = `DELETED`.

### Operation: Modify Node Spec (ModifyDBInstanceSpec)

```bash
ve mongodb ModifyDBInstanceSpec \
  --InstanceId "{{user.instance_id}}" \
  --NodeSpec "{{user.new_node_spec}}" \
  --StorageSpaceGB "{{user.new_storage_gb}}"
```

Poll until `RUNNING` (may take up to 900s).

### Operation: Restart Instance

```bash
ve mongodb RestartDBInstance --InstanceId "{{user.instance_id}}"
```

Poll until `RUNNING` (may take up to 300s).

### Operation: User Management

```bash
# List accounts
ve mongodb DescribeDBAccounts --InstanceId "{{user.instance_id}}"

# Create account
ve mongodb CreateDBAccount \
  --InstanceId "{{user.instance_id}}" \
  --AccountName "{{user.username}}" \
  --AccountPassword "{{user.password}}" \
  --AccountPrivilege "ReadWrite"

# Delete account
ve mongodb DeleteDBAccount \
  --InstanceId "{{user.instance_id}}" \
  --AccountName "{{user.username}}"

# Modify privilege
ve mongodb ModifyDBAccountPrivilege \
  --InstanceId "{{user.instance_id}}" \
  --AccountName "{{user.username}}" \
  --AccountPrivilege "ReadOnly"
```

Account privileges: `ReadWrite`, `ReadOnly`, `dbAdmin`, `root`.

### Operation: Database Management

```bash
# List databases
ve mongodb DescribeDatabases --InstanceId "{{user.instance_id}}"

# Create database
ve mongodb CreateDatabase \
  --InstanceId "{{user.instance_id}}" \
  --DBName "{{user.database_name}}"

# Delete database
ve mongodb DeleteDatabase \
  --InstanceId "{{user.instance_id}}" \
  --DBName "{{user.database_name}}"
```

### Operation: Collection Management

```bash
# List collections
ve mongodb DescribeCollections \
  --InstanceId "{{user.instance_id}}" \
  --DBName "{{user.database_name}}"

# Create collection
ve mongodb CreateCollection \
  --InstanceId "{{user.instance_id}}" \
  --DBName "{{user.database_name}}" \
  --CollectionName "{{user.collection_name}}"

# Delete collection
ve mongodb DeleteCollection \
  --InstanceId "{{user.instance_id}}" \
  --DBName "{{user.database_name}}" \
  --CollectionName "{{user.collection_name}}"
```

### Operation: Backup Management

```bash
# List backups
ve mongodb DescribeBackups --InstanceId "{{user.instance_id}}"

# Create backup
ve mongodb CreateBackup \
  --InstanceId "{{user.instance_id}}" \
  --BackupName "{{user.backup_name}}"
```

### Operation: Restore Instance

```bash
ve mongodb RestoreDBInstance \
  --InstanceId "{{user.instance_id}}" \
  --BackupId "{{user.backup_id}}"
```

Poll until `RUNNING` (may take up to 1800s).

### Operation: IP Whitelist Management

```bash
# Get current IP list
ve mongodb DescribeDBInstanceIPList --InstanceId "{{user.instance_id}}"

# Modify IP list
ve mongodb ModifyDBInstanceIPList \
  --InstanceId "{{user.instance_id}}" \
  --IPList '["10.0.0.0/8", "192.168.0.0/16"]'
```

### Operation: Parameter Management

```bash
# Query parameters
ve mongodb DescribeDBInstanceParameters --InstanceId "{{user.instance_id}}"

# Modify parameters
ve mongodb ModifyDBInstanceParameters \
  --InstanceId "{{user.instance_id}}" \
  --Parameters '[{"ParameterName":"maxConnections","ParameterValue":"2000"}]'
```

## Prerequisites

1. **`ve` CLI** installed per execution environment
2. **Go runtime** for JIT fallback (see references/integration.md)
3. **Credentials:** `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`, `VOLCENGINE_REGION`
4. **Verify:** `ve mongodb DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}"`

## Reference Directory

- [Core Concepts](references/core-concepts.md) — MongoDB architecture, versions, node types
- [API & SDK Usage](references/api-sdk-usage.md) — Full operation map, JSON paths
- [CLI Usage](references/cli-usage.md) — `ve mongodb` command reference
- [Troubleshooting Guide](references/troubleshooting.md) — Error codes, diagnostics
- [Monitoring & Alerts](references/monitoring.md) — MongoDB monitoring metrics
- [Integration](references/integration.md) — Go SDK setup, JIT workflow
- [Knowledge Base](references/knowledge-base.md) — fault pattern library

---

**Security Note:** All credentials masked as `<masked>` in logs and output.

---
name: ve-polar-mysql-ops
description: >-
  Use when the user needs to deploy, configure, troubleshoot, or monitor Volcengine
  PolarDB for MySQL (云原生数据库 PolarDB MySQL 版) — Cluster lifecycle, compute node
  management, storage pool operations, primary/standby failover, read replicas,
  backup/restore, and parameter groups. User mentions PolarDB MySQL, PolarDB,
  云原生数据库, or describes product-specific scenarios (e.g., cluster creation,
  compute node scaling, storage expansion, failover configuration) even without
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
  api_profile: "PolarDB MySQL OpenAPI 2022-01-01 — https://www.volcengine.com/docs/6498"
  cli_applicability: dual-path
  cli_support_evidence: >-
    PolarDB MySQL is a core Volcengine service; `ve polardb_mysql --help` lists
    cluster operations. OpenAPI service ID: polardb_mysql, version: 2022-01-01.
    Go SDK: github.com/volcengine/volc-sdk-golang/service/polardb_mysql.
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

# Volcengine PolarDB for MySQL Operations Skill

### What This Skill Does

Operate Volcengine PolarDB MySQL clusters, compute nodes, storage pools, read replicas, and backups using the `ve` CLI (primary) or JIT Go SDK (fallback). Manages cloud-native MySQL workloads with compute-storage separation architecture.

## Overview

Volcengine PolarDB for MySQL (云原生数据库 PolarDB MySQL 版) is a cloud-native relational database service with compute-storage separation architecture. It provides high performance, elastic scalability, and high availability through shared storage, multi-node clusters, and automatic failover. This skill is an **operational runbook** for agents: explicit scope, credential rules, pre-flight checks, **dual-path execution** (official **SDK/API** and official **`ve` CLI**), response validation, and failure recovery.

### CLI applicability

- **`cli_applicability: dual-path`:** `ve` CLI supports PolarDB MySQL operations. Document **both** SDK and CLI steps.

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | PolarDB MySQL triggers; VPC/ECS delegated to other skills |
| 2 | **Structured I/O** | `{{env.*}}` for credentials, `{{user.*}}` for params, `{{output.*}}` for API responses |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | 15 PolarDB-specific error codes with HALT/retry per type |
| 5 | **Absolute Single Responsibility** | PolarDB MySQL only; VPC/ECS/SQL schema delegated |

## Trigger & Scope

### SHOULD Use This Skill When

- User mentions "PolarDB MySQL", "PolarDB", "云原生数据库", "PolarDB集群", "计算节点", "存储池"
- Task involves PolarDB cluster lifecycle (create, describe, modify, delete, list, failover)
- Task involves compute node operations (add, remove, scale)
- Task involves storage pool management (expand, monitor)
- Task involves read replica operations (create read-only nodes, endpoint management)
- Task involves backup/restore for PolarDB clusters
- Task involves parameter group management
- Task keywords: polardb, polar, mysql, cluster, node, storage pool, read replica, failover, endpoint

### SHOULD NOT Use This Skill When

- Billing / account → delegate to `ve-billing-ops`
- IAM / permissions → delegate to `ve-iam-ops`
- VPC / subnet → delegate to `ve-vpc-ops`
- SQL schema design / DDL operations → application-level, not infrastructure
- Console-only flows → state limitation
- RDS MySQL (traditional) operations → delegate to `ve-rds-mysql-ops`

### Delegation Rules

- PolarDB requires VPC/subnets: verify via `ve-vpc-ops`
- Underlying compute nodes run on ECS: host issues delegate to `ve-ecs-ops`
- Storage pool uses shared distributed storage: issues delegate to `ve-cms-ops` for metrics

## Variable Convention

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | Runtime credential | NEVER ask user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | Runtime credential | NEVER ask user; fail if unset |
| `{{env.VOLCENGINE_REGION}}` | Runtime region | Use from env |
| `{{user.cluster_name}}` | Cluster name | Ask once |
| `{{user.cluster_id}}` | Cluster ID | Ask if not in context |
| `{{user.node_class}}` | Compute node class | Ask with options |
| `{{user.node_count}}` | Number of nodes | Ask with range (2-16) |
| `{{user.storage_size}}` | Storage in GB (100-100000) | Ask with range |
| `{{user.db_version}}` | MySQL version (5.7/8.0) | Ask with options |
| `{{output.cluster_id}}` | From API response | Parse `$.Result.ClusterId` |
| `{{output.endpoint}}` | From API response | Parse `$.Result.Endpoints[].Address` |

> **Security:** NEVER log, print, or expose `VOLCENGINE_SECRET_KEY` or any DB password.

## API and Response Conventions

- **OpenAPI 2022-01-01** is canonical for all fields and response shapes
- **API Version:** `2022-01-01`, **Service:** `polardb_mysql`
- **Endpoint:** `polardb-mysql.{region}.volcengineapi.com`

### Key Response Field Table

| Operation | JSON Path | Type | Description |
|-----------|-----------|------|-------------|
| CreateDBCluster | `$.Result.ClusterId` | string | New cluster ID |
| DescribeDBClusterDetail | `$.Result.ClusterId` | string | Cluster identifier |
| DescribeDBClusterDetail | `$.Result.ClusterName` | string | Cluster name |
| DescribeDBClusterDetail | `$.Result.ClusterStatus` | string | Cluster state |
| DescribeDBClusterDetail | `$.Result.DBEngineVersion` | string | MySQL version |
| DescribeDBClusterDetail | `$.Result.Nodes[]` | array | Compute nodes |
| DescribeDBClusterDetail | `$.Result.StorageSpace` | number | Storage GB |
| DescribeDBClusterDetail | `$.Result.StorageUsed` | number | Used storage GB |
| DescribeDBClusterDetail | `$.Result.VpcId` | string | VPC ID |
| DescribeDBClusterDetail | `$.Result.Endpoints[]` | array | Connection endpoints |
| DescribeDBClusters | `$.Result.Clusters[].ClusterId` | array | Cluster IDs |
| DescribeDBClusters | `$.Result.Clusters[].ClusterStatus` | array | Cluster states |
| DescribeDBNodes | `$.Result.Nodes[].NodeId` | array | Node IDs |
| DescribeDBNodes | `$.Result.Nodes[].NodeClass` | array | Node specs |
| DescribeDBNodes | `$.Result.Nodes[].NodeRole` | array | Primary/Secondary/ReadOnly |
| DescribeDBNodes | `$.Result.Nodes[].NodeStatus` | array | Node states |

### Cluster States

| State | Description |
|-------|-------------|
| `RUNNING` | Operational |
| `CREATING` | Being created |
| `DELETING` | Being deleted |
| `DELETED` | Deleted |
| `ERROR` | Error state |
| `RESTARTING` | Restarting |
| `SCALING` | Scaling compute or storage |
| `MODIFYING` | Configuration change in progress |
| `BACKING_UP` | Backup in progress |
| `RESTORING` | Restore in progress |
| `FAILOVERING` | Failover in progress |

### State Transitions

| Operation | Initial | Target | Poll | Max Wait |
|-----------|---------|--------|------|----------|
| Create | — | `RUNNING` | 10s | 900s |
| Delete | any | gone | 10s | 600s |
| Scale Compute | RUNNING | `RUNNING` | 10s | 900s |
| Scale Storage | RUNNING | `RUNNING` | 10s | 600s |
| Failover | RUNNING | `RUNNING` | 5s | 300s |
| Restart | RUNNING | `RUNNING` | 5s | 300s |
| Restore | — | `RUNNING` | 10s | 1800s |

## Quick Start

### Prerequisites
- [ ] `ve` CLI installed
- [ ] `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`, `VOLCENGINE_REGION` set

### Verify Setup
```bash
ve polardb_mysql DescribeDBClusters --Region "{{env.VOLCENGINE_REGION}}"
```

### Your First Command
```bash
ve polardb_mysql DescribeDBClusters --Region "{{env.VOLCENGINE_REGION}}" --PageNumber 1 --PageSize 10
```
### Next Steps
- [Core Concepts](references/core-concepts.md) — Understand PolarDB for MySQL architecture
- [Common Operations](#execution-flows) — Create, manage, and manage PolarDB databases
- [Troubleshooting](references/troubleshooting.md) — Fix common issues

## Capabilities at a Glance

| Operation | Description | Complexity | Risk |
|-----------|-------------|------------|------|
| CreateDBCluster | Create PolarDB cluster | Medium | 🟢 Low |
| DescribeDBClusterDetail | Get cluster details | Low | ✅ None |
| DescribeDBClusters | List all clusters | Low | ✅ None |
| DeleteDBCluster | Delete cluster | Low | 🔴 **High** |
| ModifyDBNodeClass | Scale compute nodes | Medium | 🟡 Medium |
| ScaleStorage | Expand storage pool | Low | 🟢 Low |
| CreateDBNode | Add read-only node | Medium | 🟢 Low |
| DeleteDBNode | Remove read-only node | Low | 🟡 Medium |
| FailoverDBCluster | Manual failover | High | 🔴 **High** |
| RestartDBNode | Restart node | Medium | 🟡 Medium |
| DescribeDBNodes | List cluster nodes | Low | ✅ None |
| CreateBackup | Create backup | Low | 🟢 Low |
| DescribeBackups | List backups | Low | ✅ None |
| RestoreDBCluster | Restore from backup | High | 🟡 Medium |
| ModifyDBClusterEndpoint | Manage endpoints | Medium | 🟢 Low |
| CreateParameterGroup | Create parameter template | Low | 🟢 Low |
| ModifyDBClusterParameters | Modify cluster params | Medium | 🟡 Medium |

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-27 | Initial release with PolarDB MySQL lifecycle |
| 1.1.0 | 2026-06-04 | GCL rollout: added `## Quality Gate (GCL)`, references/rubric.md, references/prompt-templates.md |

## Quality Gate (GCL)

> Mandatory for every execution of `ve-polar-mysql-ops`. Implements GCL per `../../AGENTS.md` §3-§9.

### Operation Tiers

| Tier | Operations | `max_iter` | Safety |
|---|---|---|---|
| **Destructive** | `DeleteDBCluster` | 2 | 1.0 |
| **State-changing** | `ModifyDBNodeSpec`, `ScaleStorage`, `Failover`, `RestartDBInstance`, `ModifyDBInstanceParameter`, `ModifyDBInstanceIPList`, `ModifyDBEndpoint` | 2 | 1.0 |
| **Mutating** | `CreateDBCluster`, `CreateDBAccount`, `CreateBackup`, `RestoreToNewInstance`, `CreateReadOnlyNode` | 2 | ≥0.5 |
| **Read-only** | `DescribeDBClusters`, `DescribeDBClusterDetail`, `DescribeDBNodes`, `DescribeDBInstanceParameters`, `DescribeDBAccounts` | 3 | ≥0 |

### Loop & Safety
- **DeleteDBCluster**: ALL compute nodes + shared storage + data lost.
- **Failover**: 30-60s write interruption.
- **ScaleStorage**: irreversible (can only increase).
- DB password masked in trace.

### Cross-skill delegation
| Finding | Delegate |
|---|---|
| VPC/subnet | `ve-vpc-ops` |
| Host-level | `ve-ecs-ops` |
| Billing | `ve-billing-ops` |

## Execution Flows

### Operation: Create PolarDB Cluster (CreateDBCluster)

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| CLI | `ve version` | Exit 0 | Go to JIT Go SDK |
| Credentials | Verify env vars | Set and valid | HALT |
| Region | Supported PolarDB region | Valid | HALT; list regions via DescribeRegions |
| VPC/Subnet | Network exists | Valid in region | HALT; use ve-vpc-ops |
| Quota | Cluster quota | Sufficient | HALT; raise quota |
| Node Class | Valid node class | Supported | HALT; list classes via DescribeDBNodeClasses |

#### Execution — CLI (Primary)

```bash
ve polardb_mysql CreateDBCluster \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --ZoneId "{{user.zone_id}}" \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --ClusterName "{{user.cluster_name}}" \
  --DBEngineVersion "{{user.db_version}}" \
  --NodeClass "{{user.node_class}}" \
  --NodeNumber "{{user.node_count}}" \
  --StorageSpace "{{user.storage_size}}" \
  --ChargeType "PostPaid"
```

#### Execution — JIT Go SDK (Fallback)

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/polardb_mysql"
)

func main() {
    instance := polardb_mysql.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "Region":          os.Getenv("VOLCENGINE_REGION"),
        "ZoneId":          os.Getenv("ZONE_ID"),
        "VpcId":           os.Getenv("VPC_ID"),
        "SubnetId":        os.Getenv("SUBNET_ID"),
        "ClusterName":     os.Getenv("CLUSTER_NAME"),
        "DBEngineVersion": os.Getenv("DB_VERSION"),
        "NodeClass":       os.Getenv("NODE_CLASS"),
        "NodeNumber":      os.Getenv("NODE_COUNT"),
        "StorageSpace":    os.Getenv("STORAGE_SIZE"),
        "ChargeType":      "PostPaid",
    }

    resp, err := instance.Client.Request("polardb_mysql", "CreateDBCluster", params)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(resp))
}
```

#### Post-execution Validation

```bash
for i in $(seq 1 90); do
  STATUS=$(ve polardb_mysql DescribeDBClusterDetail --ClusterId "{{output.cluster_id}}" | jq -r '.Result.ClusterStatus // ""')
  [ "$STATUS" = "RUNNING" ] && break
  [ "$STATUS" = "ERROR" ] && echo "Cluster creation failed" && exit 1
  sleep 10
done
```

#### Failure Recovery

| Error Code | Agent Action | Recovery |
|-----------|--------------|----------|
| `InvalidParameter.ClusterName` | HALT | 1-64 chars, letters/numbers/hyphens/underscores |
| `InvalidParameter.NodeClass` | HALT | Check via DescribeDBNodeClasses |
| `InvalidParameter.StorageSpace` | HALT | Storage [100-100000]GB |
| `InvalidParameter.VpcId` | HALT | Verify VPC exists in region via ve-vpc-ops |
| `InvalidParameter.SubnetId` | HALT | Verify subnet exists in VPC |
| `InvalidParameter.ZoneId` | HALT | Check zones via DescribeAvailabilityZones |
| `ResourceNotFound.Vpc` | HALT | Verify VPC ID |
| `QuotaExceeded.ClusterCount` | HALT | Delete unused or request quota |
| `QuotaExceeded.NodeCount` | HALT | Remove unused nodes or request quota |
| `InsufficientBalance` | HALT | Recharge account |
| `OperationDenied.ClusterStatus` | HALT | Wait for current operation to complete |
| `Throttling` | Retry (3x, exponential) | Back off and retry |
| `InternalError` | Retry (3x, 2s/4s/8s) | Report RequestId if persists |
| `ResourceAlreadyExists` | HALT | Use unique name |
| `Forbidden.RAM` | HALT | Add RAM policy |

### Operation: Describe/List Clusters

```bash
# Cluster details
ve polardb_mysql DescribeDBClusterDetail --ClusterId "{{user.cluster_id}}"

# List clusters
ve polardb_mysql DescribeDBClusters --Region "{{env.VOLCENGINE_REGION}}" --PageNumber 1 --PageSize 100
```

| Field | Path | Notes |
|-------|------|-------|
| Cluster ID | `$.Result.ClusterId` | Primary identifier |
| Name | `$.Result.ClusterName` | Human-readable |
| Status | `$.Result.ClusterStatus` | RUNNING/CREATING/ERROR |
| Engine | `$.Result.DBEngineVersion` | MySQL_5_7 / MySQL_8_0 |
| Storage | `$.Result.StorageSpace` | Total storage GB |
| Used | `$.Result.StorageUsed` | Used storage GB |
| VPC | `$.Result.VpcId` | Network |
| Endpoint | `$.Result.Endpoints[].Address` | Connection string |

### Operation: Delete Cluster

#### Safety Gate
- Explicit confirmation required for delete of `{{user.cluster_name}}` (`{{user.cluster_id}}`)
- **MUST NOT** proceed without user assent
- **Warning:** Delete is irreversible; data cannot be recovered

#### Execution
```bash
ve polardb_mysql DeleteDBCluster --ClusterId "{{user.cluster_id}}"
```

#### Validation
Poll until gone or status = `DELETED`.

### Operation: Scale Compute (ModifyDBNodeClass)

```bash
ve polardb_mysql ModifyDBNodeClass \
  --ClusterId "{{user.cluster_id}}" \
  --NodeClass "{{user.new_node_class}}"
```

Poll until `RUNNING` (may take up to 900s).

### Operation: Scale Storage (ScaleStorage)

```bash
ve polardb_mysql ScaleStorage \
  --ClusterId "{{user.cluster_id}}" \
  --StorageSpace "{{user.new_storage_size}}"
```

Poll until `RUNNING` (may take up to 600s).

### Operation: Manage Read-Only Nodes

#### Add Read-Only Node
```bash
ve polardb_mysql CreateDBNode \
  --ClusterId "{{user.cluster_id}}" \
  --NodeClass "{{user.node_class}}" \
  --NodeNumber 1
```

#### Remove Read-Only Node
```bash
ve polardb_mysql DeleteDBNode \
  --ClusterId "{{user.cluster_id}}" \
  --NodeId "{{user.node_id}}"
```

**Safety Gate:** Removing a node is irreversible; confirm with user.

#### List Nodes
```bash
ve polardb_mysql DescribeDBNodes --ClusterId "{{user.cluster_id}}"
```

### Operation: Failover (FailoverDBCluster)

#### Safety Gate
- **Warning:** Failover causes brief connection interruption (typically < 30s)
- Confirm with user before proceeding
- Verify cluster has at least 2 nodes (primary + secondary)

#### Execution
```bash
ve polardb_mysql FailoverDBCluster --ClusterId "{{user.cluster_id}}"
```

Poll until `RUNNING`.

### Operation: Restart Node

```bash
ve polardb_mysql RestartDBNode --NodeId "{{user.node_id}}"
```

Poll until node status = `RUNNING`.

### Operation: Backup Management

```bash
# List backups
ve polardb_mysql DescribeBackups --ClusterId "{{user.cluster_id}}"

# Create backup
ve polardb_mysql CreateBackup --ClusterId "{{user.cluster_id}}" --BackupName "{{user.backup_name}}"
```

### Operation: Restore Cluster

```bash
ve polardb_mysql RestoreDBCluster \
  --BackupId "{{user.backup_id}}" \
  --ClusterName "{{user.new_cluster_name}}" \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}"
```

Poll until `RUNNING` (may take up to 1800s).

### Operation: Endpoint Management

```bash
# List endpoints
ve polardb_mysql DescribeDBClusterEndpoints --ClusterId "{{user.cluster_id}}"

# Modify endpoint
ve polardb_mysql ModifyDBClusterEndpoint \
  --ClusterId "{{user.cluster_id}}" \
  --EndpointId "{{user.endpoint_id}}" \
  --AutoAddNewNodes true
```

### Operation: Parameter Group Management

```bash
# List parameters
ve polardb_mysql DescribeDBClusterParameters --ClusterId "{{user.cluster_id}}"

# Modify parameters
ve polardb_mysql ModifyDBClusterParameters \
  --ClusterId "{{user.cluster_id}}" \
  --Parameters '[{"ParameterName":"max_connections","ParameterValue":"2000"}]'

# List parameter groups
ve polardb_mysql DescribeParameterGroups --Region "{{env.VOLCENGINE_REGION}}"

# Create parameter group
ve polardb_mysql CreateParameterGroup \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --ParameterGroupName "{{user.group_name}}" \
  --DBEngineVersion "{{user.db_version}}"
```

## Error Taxonomy

| 错误码 | 含义 | 分辨率 |
|--------|------|--------|
| `InvalidClusterId.NotFound` | 集群 ID 不存在 | 0 retries; **HALT** — 检查 ClusterId 是否正确 |
| `IncorrectClusterState` | 集群状态不允许操作 | 0 retries; **HALT** — 等待集群进入 RUNNING 态后重试 |
| `QuotaExceeded.Cluster` | 集群数量已达配额上限 | 0 retries; **HALT** — 删除无用集群或申请提升配额 |
| `InvalidParameter.NodeClass` | 计算节点规格无效 | 0 retries; **HALT** — 通过 DescribeDBNodeClasses 查询可用规格 |
| `InvalidParameter.StorageSpace` | 存储空间超出范围 [100-100000]GB | 0 retries; **HALT** — 提供合法的存储容量值 |
| `InvalidZoneId` | 可用区 ID 无效 | 0 retries; **HALT** — 通过 DescribeAvailabilityZones 查询可用区 |
| `InvalidSecurityGroupId` | 安全组 ID 无效 | 0 retries; **HALT** — 通过 ve-ecs-ops 验证安全组 |
| `InvalidVpcId` | VPC ID 无效 | 0 retries; **HALT** — 通过 ve-vpc-ops 验证 VPC 是否存在 |
| `OperationDenied.ClusterStatus` | 当前集群状态不允许操作 | 0 retries; **HALT** — 等待当前操作完成后再试 |
| `InsufficientBalance` | 账户余额不足 | 0 retries; **HALT** — 充值后重试 |
| `Throttling` | 请求频率超限 | 3 retries/exponential/1s/2s/4s; **RETRY** |
| `InternalError` | 服务端内部错误 | 3 retries/exponential/2s/4s/8s; **RETRY** — 持续失败则记录 RequestId 并反馈 |

## Prerequisites

1. **`ve` CLI** installed per execution environment
2. **Go runtime** for JIT fallback (see references/integration.md)
3. **Credentials:** `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`, `VOLCENGINE_REGION`
4. **Verify:** `ve polardb_mysql DescribeDBClusters --Region "{{env.VOLCENGINE_REGION}}"`

## Operational Best Practices

- **Security:** Deploy clusters in dedicated VPCs with strict subnet ACLs; use least-privilege database accounts; enable TLS for client connections.
- **Reliability:** Ensure multi-AZ deployments with primary and standby nodes; schedule regular `CreateBackup`; enable backup retention for point-in-time recovery.
- **Compute-storage separation:** Storage scaling is irreversible — plan `ScaleStorage` carefully; monitor `StorageUsed` vs `StorageSpace` and scale before reaching 90%.
- **Failover planning:** Test `FailoverDBCluster` regularly to verify RTO; failover causes <30s write interruption — schedule during maintenance windows.
- **Endpoint management:** Use dedicated endpoints for read-write and read-only traffic; enable `AutoAddNewNodes` to automatically attach new read replicas to the endpoint.
- **Naming conventions:** Use consistent prefixes for cluster names, parameter groups, and endpoints to simplify cross-team auditing.
- **Cost optimization:** Right-size `NodeClass` to workload; scale down idle compute nodes during off-peak hours; monitor storage utilization to avoid over-provisioning.

## Reference Directory

- [Core Concepts](references/core-concepts.md) — PolarDB architecture, compute-storage separation, node types
- [API & SDK Usage](references/api-sdk-usage.md) — Full operation map, JSON paths
- [CLI Usage](references/cli-usage.md) — `ve polardb_mysql` command reference
- [Troubleshooting Guide](references/troubleshooting.md) — Error codes, diagnostics
- [Monitoring & Alerts](references/monitoring.md) — PolarDB monitoring metrics
- [Integration](references/integration.md) — Go SDK setup, JIT workflow
- [GCL Rubric](references/rubric.md)
- [GCL Prompt Templates](references/prompt-templates.md)

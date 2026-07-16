---
name: ve-elasticsearch-ops
description: >-
  Use when the user needs to deploy, configure, troubleshoot, or monitor Volcengine
  (火山引擎) Elasticsearch Service (Elasticsearch 服务) — instance lifecycle,
  index management, Kibana configuration, plugin management, snapshot/restore,
  monitoring, and scaling operations. User mentions Elasticsearch, ES, 弹性搜索,
  Elasticsearch 服务, or describes search-related scenarios (e.g., indexing,
  query performance, cluster health, shard allocation) even without naming the
  product directly. Not for Logstash, Beats, Kibana standalone, or other Elastic
  Stack components outside the managed service scope.
license: MIT
compatibility: >-
  Official Volcengine CLI (`ve`, Go binary, no runtime), Go 1.21+ runtime
  (JIT SDK fallback; scripts compatible with Go 1.14+ syntax), valid API credentials,
  network access to Elasticsearch endpoints (es.{region}.volces.com).
metadata:
  author: volcengine
  version: "1.1.0"
  last_updated: "2026-06-04"
  runtime: Harness AI Agent, Claude Code, Cursor, or compatible Agent runtimes
  go_version_minimum: "1.14"
  go_jit_runtime_version: "1.21+"
  cli_applicability: dual-path
  cli_support_evidence: >-
    Elasticsearch API is accessible via `ve elasticsearch --help`. CLI supports instance,
    index, snapshot, and Kibana management operations.
    API docs: https://www.volcengine.com/docs/6337
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

> This skill follows the [Agent Skill OpenSpec](https://agentskills.io/specification).

# Volcengine Elasticsearch Operations Skill

## Overview

Elasticsearch Service (Elasticsearch 服务) on Volcengine (火山引擎) provides fully managed Elasticsearch clusters with high-performance search, analytics, and indexing capabilities. This skill is an **operational runbook** for agents with **dual-path execution**: `ve` CLI for API calls and JIT Go SDK fallback.

### CLI applicability

- **`cli_applicability: dual-path`:** Official `ve` CLI supports Elasticsearch operations.
  - **`ve elasticsearch`**: Instance, index, snapshot, Kibana, and plugin management

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use conditions with precise triggers and delegation rules |
| 2 | **Structured I/O** | `{{env.VOLCENGINE_*}}` (env), `{{user.*}}` (interactive), `{{output.*}}` (API response) |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | Error taxonomy with ≥ 10 ES-specific codes; HALT vs retry per type |
| 5 | **Absolute Single Responsibility** | One product (Elasticsearch), one primary resource model; cross-product delegation documented |

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- User mentions "Volcengine Elasticsearch", "火山引擎 Elasticsearch 服务", "ES", "Elasticsearch", "弹性搜索"
- Task involves instance operations: CreateInstance, DescribeInstances, ModifyInstanceSpec, DeleteInstance, ListInstances
- Task involves index operations: CreateIndex, DescribeIndices, DeleteIndex, ListIndices
- Task involves Kibana operations: EnableKibana, DescribeKibana, DisableKibana
- Task involves plugin management: InstallPlugin, DescribePlugins, UninstallPlugin
- Task involves snapshot operations: CreateSnapshot, DescribeSnapshots, DeleteSnapshot
- Task involves configuration: ModifyInstanceChargeType, ModifyNodeSpec
- Task involves monitoring: cluster health, shard allocation, disk usage, query performance
- Task involves scaling: vertical scaling (node specs), horizontal scaling (node count)
- Task involves restart or upgrade: RestartInstance, UpgradeVersion

### SHOULD NOT Use This Skill When

- Task is about Logstash → delegate to: `ve-logstash-ops` (when present)
- Task is about Beats → delegate to: `ve-beats-ops` (when present)
- Task is about Kibana standalone (not managed) → delegate to separate skill
- Task is purely billing → delegate to billing ops
- Task is application-level search query tuning (ES query DSL) → application-level, not infrastructure

### Delegation Rules

- ES instances depend on VPC networking → reference `ve-vpc-ops` for subnet/VPC setup
- ES authentication depends on IAM for access control → reference `ve-iam-ops` (when present)
- ES metrics are collected via CMS → reference `ve-cms-ops` for monitoring setup
- ES backup storage uses TOS → reference `ve-tos-ops` for bucket configuration

## Variable Convention (Agent-Readable)

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | Access key from runtime | NEVER ask user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | Secret key from runtime | NEVER ask user; fail if unset; **ALWAYS mask as `<masked>`** |
| `{{env.VOLCENGINE_REGION}}` | Region (e.g., cn-beijing) | Use documented default if skill allows |
| `{{user.instance_id}}` | ES instance ID | Ask once; starts with `es-` |
| `{{user.instance_name}}` | ES instance name | Ask once; globally unique per region |
| `{{user.instance_version}}` | ES version (7.10/7.16/8.5) | Ask with version options |
| `{{user.node_spec}}` | Node specification | Ask with available specs |
| `{{user.node_count}}` | Number of data nodes | Ask; default 3 |
| `{{user.storage_gb}}` | Storage per node in GB | Ask with range [20-2000] |
| `{{user.vpc_id}}` | VPC ID | Ask if not in context |
| `{{user.subnet_id}}` | Subnet ID | Ask if not in context |
| `{{user.index_name}}` | Index name | Ask once |
| `{{user.shard_count}}` | Number of primary shards | Ask; default 5 |
| `{{user.replica_count}}` | Number of replicas | Ask; default 1 |
| `{{user.snapshot_name}}` | Snapshot name | Ask once |
| `{{user.plugin_name}}` | Plugin name | Ask once |
| `{{user.kibana_user}}` | Kibana username | Ask once |
| `{{user.kibana_password}}` | Kibana password | Ask once; min 8 chars, mask in output |
| `{{output.instance_id}}` | Instance ID from create response | Parse from response |
| `{{output.request_id}}` | Request ID for tracing | Parse from response |

> **Security Warning (Credential Masking):** NEVER echo or log `VOLCENGINE_SECRET_KEY` or any credential values. Verify existence only with `test -n "$VOLCENGINE_SECRET_KEY"`.

> **Password Masking:** NEVER display Kibana passwords in output. Always show as `<masked>`.

## API and Response Conventions (Agent-Readable)

- **Elasticsearch uses RESTful API** with JSON responses
- **Endpoint:** `es.{region}.volces.com` for API calls
- **Go SDK:** `github.com/volcengine/volc-sdk-golang/service/elasticsearch`
- **Error responses:** JSON with `Error` object containing `Code` and `Message`

### Key Response Fields

| Operation | Response Field | Type | Description |
|-----------|---------------|------|-------------|
| CreateInstance | `$.Result.InstanceId` | string | New instance ID |
| DescribeInstances | `$.Result.Instances[].InstanceId` | array | Instance IDs |
| DescribeInstances | `$.Result.Instances[].Status` | array | Instance state |
| DescribeInstances | `$.Result.Instances[].Version` | array | ES version |
| ListIndices | `$.Result.Indices[].IndexName` | array | Index names |
| DescribeIndices | `$.Result.ShardCount` | integer | Shard count |
| DescribeIndices | `$.Result.ReplicaCount` | integer | Replica count |
| DescribeSnapshots | `$.Result.Snapshots[].SnapshotName` | array | Snapshot names |
| DescribePlugins | `$.Result.Plugins[].PluginName` | array | Plugin names |
| DescribeKibana | `$.Result.KibanaEndpoint` | string | Kibana URL |

### Expected State Transitions

| Operation | Initial State | Target State | Poll Interval | Max Wait |
|-----------|---------------|--------------|---------------|----------|
| Create | — | `Running` | 30s | 1800s |
| Restart | `Running` | `Restarting` → `Running` | 30s | 600s |
| Upgrade | `Running` | `Upgrading` → `Running` | 30s | 3600s |
| ModifySpec | `Running` | `Modifying` → `Running` | 30s | 1800s |
| Delete | any stable | `Deleting` → absent | 30s | 1800s |

## Quick Start

### What This Skill Does
This skill enables you to manage Volcengine Elasticsearch Service instances and resources — create instances, manage indices, configure Kibana, install plugins, take snapshots, and scale resources.

### Prerequisites
- [ ] `ve` CLI installed
- [ ] Credentials configured: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`
- [ ] Region set: `VOLCENGINE_REGION`

### Verify Setup
```bash
# List ES instances
ve elasticsearch DescribeInstances --Region {{env.VOLCENGINE_REGION}}
```

### Your First Command
```bash
# List all Elasticsearch instances
ve elasticsearch DescribeInstances --Region {{env.VOLCENGINE_REGION}}
```

### Next Steps
- [Core Concepts](references/core-concepts.md) — Understand Elasticsearch architecture
- [Common Operations](#execution-flows) — Create, manage, and search and analyze data
- [Troubleshooting](references/troubleshooting.md) — Fix common issues

## Capabilities at a Glance

| Operation | Description | Complexity | Risk Level |
|-----------|-------------|------------|------------|
| CreateInstance | Create a new ES instance | Medium | Low |
| DescribeInstances | View instance details | Low | None |
| ModifyInstanceSpec | Change instance configuration | Medium | Medium |
| DeleteInstance | Remove an instance | Low | **High** — irreversible |
| ListInstances | View all instances | Low | None |
| RestartInstance | Restart an instance | Low | Medium |
| UpgradeVersion | Upgrade ES version | High | **High** — potential incompatibility |
| ModifyNodeCount | Scale nodes horizontally | Medium | Medium |
| CreateIndex | Create an index | Low | Low |
| DescribeIndices | View index details | Low | None |
| DeleteIndex | Delete an index | Low | **High** — data loss |
| ListIndices | List all indices | Low | None |
| CreateSnapshot | Create a snapshot | Low | Low |
| DescribeSnapshots | List snapshots | Low | None |
| DeleteSnapshot | Delete a snapshot | Low | Medium |
| InstallPlugin | Install a plugin | Medium | Medium |
| DescribePlugins | List installed plugins | Low | None |
| UninstallPlugin | Remove a plugin | Medium | Medium |
| EnableKibana | Enable Kibana access | Low | Low |
| DescribeKibana | Get Kibana endpoint | Low | None |
| DisableKibana | Disable Kibana access | Low | Low |
| ModifyChargeType | Change billing method | Medium | Medium |

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-31 | Initial release with instance, index, snapshot, Kibana, and plugin management |
| 1.1.0 | 2026-06-04 | GCL rollout: added `## Quality Gate (GCL)`, references/rubric.md, references/prompt-templates.md |

## Quality Gate (GCL)

> Mandatory for every execution of `ve-elasticsearch-ops`. Implements GCL per `../../AGENTS.md` §3-§9.

### Operation Tiers

> See [`references/rubric.md` §0](references/rubric.md#0-operation-tier) for the full operation tier table.

### Loop & Safety
- **DeleteInstance**: ALL indices+data+snapshots lost.
- **DeleteIndex**: confirm index name; data loss.
- **ModifyInstanceSpec**: 60-1800s cluster rebalancing.
- **UpgradeVersion**: rolling upgrade per node; cannot downgrade.
- **VOLCENGINE_SECRET_KEY** never in trace. Kibana password masked.

### Cross-skill delegation
| Finding | Delegate |
|---|---|
| VPC/subnet | `ve-vpc-ops` |
| TOS backup storage | `ve-tos-ops` |
| IAM access control | `ve-iam-ops` |
| Billing | `ve-billing-ops` |

## Execution Flows (Agent-Readable)

Every operation: **Pre-flight → Execute → Validate → Recover**.

### Operation: CreateInstance — Create an Elasticsearch Instance

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | `test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY"` | Both set | HALT; configure credentials |
| Region | Verify `{{user.region}}` is valid | Region supported | Suggest valid region |
| VPC/Subnet | `ve vpc DescribeVpcs` | VPC exists | HALT; create VPC first |
| Name unique | Instance name not in use | No conflict | Use different name |
| Quota | `ve elasticsearch ListInstances` | Under quota | HALT; request quota increase |
| Version valid | Check supported versions | Version supported | Suggest valid version |

#### Execution — CLI (`ve`)

```bash
# Create an Elasticsearch instance (PostPaid)
ve elasticsearch CreateInstance \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceName "{{user.instance_name}}" \
  --Version "{{user.instance_version}}" \
  --NodeSpec "{{user.node_spec}}" \
  --NodeNumber {{user.node_count}} \
  --StorageSpaceGb {{user.storage_gb}} \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --ChargeType "PostPaid"

# Create with specific configuration
ve elasticsearch CreateInstance \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceName "{{user.instance_name}}" \
  --Version "{{user.es_version}}" \
  --NodeSpec "es.x2.medium" \
  --NodeNumber 3 \
  --StorageSpaceGb 100 \
  --StorageType "ESSD" \
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

    "github.com/volcengine/volc-sdk-golang/service/elasticsearch"
)

func main() {
    instance := elasticsearch.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "Region":       os.Getenv("VOLCENGINE_REGION"),
        "InstanceName": os.Getenv("INSTANCE_NAME"),
        "Version":      os.Getenv("ES_VERSION"),
        "NodeSpec":     os.Getenv("NODE_SPEC"),
        "NodeNumber":   3,
        "StorageSpaceGb": 100,
        "VpcId":        os.Getenv("VPC_ID"),
        "SubnetId":     os.Getenv("SUBNET_ID"),
        "ChargeType":   "PostPaid",
    }

    resp, err := instance.Client.Request("elasticsearch", "CreateInstance", params)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(resp))
}
```

#### Validation

```bash
# Poll until instance is Running
for i in $(seq 1 60); do
  STATUS=$(ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{output.instance_id}}" | jq -r '.Result.Instances[0].Status')
  echo "Status: $STATUS (attempt $i/60)"
  [ "$STATUS" = "Running" ] && break
  sleep 30
done

# Verify instance details
ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{output.instance_id}}" | jq '.Result.Instances[0] | {Name: .InstanceName, Version: .Version, Status: .Status, NodeSpec: .NodeSpec}'
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `InvalidVpc.NotFound` | HALT; VPC does not exist — create VPC first via `ve-vpc-ops` |
| `InvalidSubnet.NotFound` | HALT; Subnet does not exist — create subnet first |
| `InvalidParameter` | HALT; fix parameter values per error message |
| `QuotaExceeded` | HALT; instance quota exceeded — request quota increase |
| `InsufficientBalance` | HALT; account balance insufficient — recharge |
| `InvalidVersion` | HALT; ES version not supported — select from valid versions |
| `InvalidNodeSpec` | HALT; node spec not available — select valid spec |

---

### Operation: DeleteInstance — Delete an Elasticsearch Instance

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: irreversible delete of instance `{{user.instance_id}}` ({{user.instance_name}})
- **MUST NOT** proceed without clear user assent
- **MUST** warn about data loss: all indices, data, snapshots, and configurations will be permanently deleted
- **MUST** verify instance status is `Running` or `Stopped`

```bash
# Verify instance exists and get details
ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}"

# List indices to warn about data loss
echo "WARNING: The following indices and all their data will be permanently deleted:"
ve elasticsearch ListIndices --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" | jq -r '.Result.Indices[].IndexName'
```

#### Execution

```bash
# Delete the instance
ve elasticsearch DeleteInstance \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}"
```

#### Validation

```bash
# Poll until instance is deleted (DescribeInstances returns empty or error)
for i in $(seq 1 60); do
  RESULT=$(ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" 2>&1)
  if echo "$RESULT" | grep -q "InvalidInstance.NotFound" || echo "$RESULT" | grep -q '"Instances": \[\]'; then
    echo "Instance deleted successfully"
    break
  fi
  echo "Waiting for deletion... (attempt $i/60)"
  sleep 30
done
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `InvalidInstance.NotFound` | Instance already deleted; skip |
| `InvalidInstanceStatus` | HALT; instance is being modified — wait and retry |
| `DependencyViolation` | HALT; resources still attached — remove dependencies first |
| `Unauthorized` | HALT; check IAM permissions |

---

### Operation: CreateIndex — Create an Index

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | `ve elasticsearch DescribeInstances` | Instance Running | HALT |
| Index name unique | `ve elasticsearch ListIndices` | Name not in list | Use different name |
| Shard count valid | Check limits | Within range | Adjust shard count |
| Storage available | Instance storage check | Sufficient space | Scale instance storage |

#### Execution

```bash
# Create an index
ve elasticsearch CreateIndex \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --IndexName "{{user.index_name}}" \
  --ShardCount {{user.shard_count}} \
  --ReplicaCount {{user.replica_count}}

# Create index with settings
ve elasticsearch CreateIndex \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --IndexName "{{user.index_name}}" \
  --ShardCount 5 \
  --ReplicaCount 1 \
  --Settings '{"index.refresh_interval": "30s", "index.number_of_routing_shards": 10}'
```

#### Validation

```bash
# Verify index exists
ve elasticsearch DescribeIndices \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --IndexName "{{user.index_name}}"

# Check index health
ve elasticsearch DescribeIndices \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --IndexName "{{user.index_name}}" | jq '.Result.Health'
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `InvalidInstance.NotFound` | HALT; instance does not exist |
| `IndexAlreadyExists` | HALT; index exists — use different name or delete first |
| `InvalidParameter.ShardCount` | HALT; shard count exceeds limit — reduce shard count |
| `QuotaExceeded.Index` | HALT; index quota exceeded — delete unused indices |
| `InvalidIndexName` | HALT; index name contains invalid characters |

---

### Operation: DeleteIndex — Delete an Index

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: irreversible delete of index `{{user.index_name}}`
- **MUST NOT** proceed without clear user assent
- **MUST** warn about data loss: all documents in the index will be permanently deleted
- **MUST** check if any aliases or search applications depend on this index

```bash
# Warn user
echo "WARNING: Deleting index '{{user.index_name}}' will permanently delete all documents."
echo "Ensure no active search queries depend on this index."
```

#### Execution

```bash
# Delete the index
ve elasticsearch DeleteIndex \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --IndexName "{{user.index_name}}"
```

#### Validation

```bash
# Verify index is deleted
if ve elasticsearch DescribeIndices --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" --IndexName "{{user.index_name}}" 2>&1 | grep -q "IndexNotFound"; then
  echo "Index deleted successfully"
fi
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `IndexNotFound` | Index already deleted; skip |
| `InvalidInstance.NotFound` | HALT; instance does not exist |
| `DeleteIndexPartial` | Some shards failed deletion — investigate cluster health |

---

### Operation: CreateSnapshot — Create a Snapshot

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | `ve elasticsearch DescribeInstances` | Instance Running | HALT |
| Snapshot name unique | `ve elasticsearch DescribeSnapshots` | Name not in list | Use different name |
| Repository configured | Snapshot repository exists | Repository ready | HALT; configure repository |

#### Execution

```bash
# Create a snapshot
ve elasticsearch CreateSnapshot \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --SnapshotName "{{user.snapshot_name}}" \
  --Indices "*"

# Create snapshot of specific indices
ve elasticsearch CreateSnapshot \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --SnapshotName "{{user.snapshot_name}}" \
  --Indices "index1,index2,index3"
```

#### Validation

```bash
# Verify snapshot status
ve elasticsearch DescribeSnapshots \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --SnapshotName "{{user.snapshot_name}}" | jq '.Result.Snapshots[0].Status'

# Poll until SUCCESS
for i in $(seq 1 60); do
  STATUS=$(ve elasticsearch DescribeSnapshots --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" --SnapshotName "{{user.snapshot_name}}" | jq -r '.Result.Snapshots[0].Status')
  echo "Snapshot: $STATUS (attempt $i/60)"
  [ "$STATUS" = "SUCCESS" ] && break
  sleep 30
done
```

---

### Operation: InstallPlugin — Install a Plugin

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | `ve elasticsearch DescribeInstances` | Instance Running | HALT |
| Plugin not installed | `ve elasticsearch DescribePlugins` | Plugin not in list | HALT; plugin already installed |
| Plugin compatible | Check plugin ES version compatibility | Compatible | Select compatible plugin |

#### Execution

```bash
# Install a plugin
ve elasticsearch InstallPlugin \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --PluginName "{{user.plugin_name}}"

# Install with forced restart
ve elasticsearch InstallPlugin \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --PluginName "{{user.plugin_name}}" \
  --ForceRestart true
```

#### Validation

```bash
# Verify plugin is installed
ve elasticsearch DescribePlugins \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --PluginName "{{user.plugin_name}}"

# List all installed plugins
ve elasticsearch DescribePlugins \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" | jq -r '.Result.Plugins[].PluginName'
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `PluginAlreadyExists` | HALT; plugin already installed — skip or uninstall first |
| `PluginNotFound` | HALT; plugin not available for this ES version |
| `PluginIncompatible` | HALT; plugin incompatible with current ES version |
| `RestartRequired` | Plugin requires instance restart — proceed with restart |

---

### Operation: UninstallPlugin — Remove a Plugin

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Plugin installed | `ve elasticsearch DescribePlugins` | Plugin found | HALT; plugin not installed |
| No dependency | Check if other plugins depend | No dependency | HALT; remove dependencies first |

#### Execution

```bash
# Uninstall a plugin
ve elasticsearch UninstallPlugin \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --PluginName "{{user.plugin_name}}"
```

#### Validation

```bash
# Verify plugin is removed
ve elasticsearch DescribePlugins \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" | jq -r '.Result.Plugins[].PluginName' | grep -v "{{user.plugin_name}}" || echo "Plugin removed"
```

---

### Operation: EnableKibana — Enable Kibana Access

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | `ve elasticsearch DescribeInstances` | Instance Running | HALT |
| Kibana disabled | `ve elasticsearch DescribeKibana` | Not enabled | Already enabled |
| Password valid | Length and complexity | Min 8 chars | Require stronger password |

#### Execution

```bash
# Enable Kibana with admin credentials
ve elasticsearch EnableKibana \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --KibanaUser "{{user.kibana_user}}" \
  --KibanaPassword "{{user.kibana_password}}"
```

> **Security:** Kibana password is masked in all output as `<masked>`.

#### Validation

```bash
# Get Kibana endpoint
ve elasticsearch DescribeKibana \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}"

# Extract the Kibana URL
KIBANA_URL=$(ve elasticsearch DescribeKibana --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" | jq -r '.Result.KibanaEndpoint')
echo "Kibana is available at: $KIBANA_URL"
```

---

### Operation: UpgradeVersion — Upgrade ES Version

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: upgrade ES instance `{{user.instance_id}}` from current version to target version
- **MUST** warn about potential incompatibilities: deprecated features, breaking API changes
- **MUST** take a snapshot before upgrade for rollback capability
- **MUST** verify the upgrade path is supported (no skip versions)

```bash
# Take pre-upgrade snapshot
ve elasticsearch CreateSnapshot \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --SnapshotName "pre-upgrade-$(date +%Y%m%d%H%M%S)" \
  --Indices "*"
```

#### Execution

```bash
# Upgrade ES version
ve elasticsearch UpgradeVersion \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --TargetVersion "{{user.target_version}}"
```

#### Validation

```bash
# Poll until Running on new version
for i in $(seq 1 120); do
  VERSION=$(ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" | jq -r '.Result.Instances[0].Version')
  STATUS=$(ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" | jq -r '.Result.Instances[0].Status')
  echo "Version: $VERSION, Status: $STATUS (attempt $i/120)"
  [ "$STATUS" = "Running" ] && [ "$VERSION" = "{{user.target_version}}" ] && break
  sleep 30
done
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `IncompatibleVersion` | HALT; upgrade path not supported — check version compatibility matrix |
| `VersionNotFound` | HALT; target version not available — check supported versions |
| `UpgradeInProgress` | Another upgrade in progress — wait and retry |
| `SnapshotRequired` | HALT; take snapshot before upgrade |
| `ClusterHealthNotGreen` | HALT; cluster health must be at least Yellow |

---

### Operation: ModifyInstanceSpec — Modify Spec / Scale

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | `ve elasticsearch DescribeInstances` | Instance Running | HALT |
| Valid spec | Target spec is available | Spec found | Select valid spec |
| No ongoing operation | Instance status | Not modifying | Wait for completion |

#### Execution

```bash
# Modify node specification
ve elasticsearch ModifyInstanceSpec \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --NodeSpec "{{user.new_node_spec}}"

# Modify node count (horizontal scaling)
ve elasticsearch ModifyNodeCount \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --NodeNumber {{user.new_node_count}}

# Modify storage per node
ve elasticsearch ModifyInstanceSpec \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --StorageSpaceGb {{user.new_storage_gb}}
```

#### Validation

```bash
# Poll until spec change completes
for i in $(seq 1 60); do
  STATUS=$(ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" | jq -r '.Result.Instances[0].Status')
  echo "Status: $STATUS (attempt $i/60)"
  [ "$STATUS" = "Running" ] && break
  sleep 30
done

# Verify new spec
ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" | jq '.Result.Instances[0] | {Spec: .NodeSpec, Count: .NodeNumber, Storage: .StorageSpaceGb}'
```

---

### Operation: RestartInstance — Restart an Instance

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | `ve elasticsearch DescribeInstances` | Instance Running | HALT |
| No pending operations | Instance status | Stable | Wait for completion |

#### Execution

```bash
# Restart the instance
ve elasticsearch RestartInstance \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}"

# Restart with force (ignore cluster health)
ve elasticsearch RestartInstance \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --Force true
```

#### Validation

```bash
# Poll until Running again
for i in $(seq 1 20); do
  STATUS=$(ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" | jq -r '.Result.Instances[0].Status')
  echo "Status: $STATUS (attempt $i/20)"
  [ "$STATUS" = "Running" ] && break
  sleep 30
done
```

---

### Operation: DescribeSnapshots — List Snapshots

#### Execution

```bash
# List all snapshots
ve elasticsearch DescribeSnapshots \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}"

# Get specific snapshot details
ve elasticsearch DescribeSnapshots \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --SnapshotName "{{user.snapshot_name}}"
```

---

### Operation: ModifyChargeType — Change Billing Method

#### Pre-flight (Safety Gate)

- **MUST** confirm billing change: switching from PostPaid to PrePaid or vice versa
- **MUST** warn about potential cost impact

#### Execution

```bash
# Switch to PrePaid (subscription)
ve elasticsearch ModifyInstanceChargeType \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --ChargeType "PrePaid" \
  --PeriodUnit "Month" \
  --Period 1

# Switch to PostPaid (pay-as-you-go)
ve elasticsearch ModifyInstanceChargeType \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --ChargeType "PostPaid"
```

---

### Operation: ListInstances — List Elasticsearch Instances

#### Execution

```bash
# List all instances
ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}"

# Filter by name
ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceName "{{user.instance_name}}"

# Filter by status
ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --Status "Running"

# Get single instance details
ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}"
```

#### Present to User

| Field | JSON Path | Notes |
|-------|-----------|-------|
| Instance ID | `.Result.Instances[].InstanceId` | Format `es-xxxxxxxxx` |
| Name | `.Result.Instances[].InstanceName` | User-defined |
| Version | `.Result.Instances[].Version` | ES version |
| Status | `.Result.Instances[].Status` | Running, Creating, Modifying, Deleting |
| Node Spec | `.Result.Instances[].NodeSpec` | Instance type |
| Node Count | `.Result.Instances[].NodeNumber` | Data node count |
| Storage | `.Result.Instances[].StorageSpaceGb` | GB per node |
| ChargeType | `.Result.Instances[].ChargeType` | PostPaid or PrePaid |
| CreateTime | `.Result.Instances[].CreateTime` | ISO 8601 |

---

### Operation: ListIndices — List Indices

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | `ve elasticsearch DescribeInstances` | Instance Running | HALT |

#### Execution

```bash
# List all indices
ve elasticsearch ListIndices \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}"

# List indices with details
ve elasticsearch ListIndices \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --Verbose true
```

#### Present to User

| Field | JSON Path | Notes |
|-------|-----------|-------|
| Index Name | `.Result.Indices[].IndexName` | Index identifier |
| Health | `.Result.Indices[].Health` | green, yellow, red |
| Docs Count | `.Result.Indices[].DocsCount` | Document count |
| Storage Size | `.Result.Indices[].StoreSize` | Storage used |

---

### Operation: DeleteSnapshot — Delete a Snapshot

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: delete snapshot `{{user.snapshot_name}}`
- **MUST** warn that deleted snapshots cannot be recovered

#### Execution

```bash
# Delete a snapshot
ve elasticsearch DeleteSnapshot \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --SnapshotName "{{user.snapshot_name}}"
```

#### Validation

```bash
# Verify snapshot is deleted
if ve elasticsearch DescribeSnapshots --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" --SnapshotName "{{user.snapshot_name}}" 2>&1 | grep -q "SnapshotNotFound"; then
  echo "Snapshot deleted successfully"
fi
```

#### Failure Recovery

| Error pattern | Max retries | Recovery |
|--------------|-------------|---------|
| `SnapshotNotFound` | 0 | Already deleted; skip — `[INFO] Snapshot already deleted.` |
| `SnapshotInProgress` | 3 | 10s wait for completion — `[WARNING] Snapshot in progress. Retrying...` |

---

### Operation: DisableKibana — Disable Kibana Access

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: disable Kibana for instance `{{user.instance_id}}`
- **MUST** warn that Kibana URL will become inaccessible
- **MUST** verify Kibana is currently enabled

```bash
# Verify Kibana is enabled
ve elasticsearch DescribeKibana \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" | jq '.Result.KibanaEndpoint'
```

#### Execution

```bash
# Disable Kibana
ve elasticsearch DisableKibana \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}"
```

#### Validation

```bash
# Verify Kibana is disabled
ve elasticsearch DescribeKibana \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" 2>&1 | grep -q "KibanaDisabled" || echo "Kibana has been disabled"
```

#### Failure Recovery

| Error pattern | Max retries | Recovery |
|--------------|-------------|---------|
| `KibanaAlreadyDisabled` | 0 | Already disabled; skip — `[INFO] Kibana already disabled.` |
| `InvalidInstance.NotFound` | 0 | HALT; instance does not exist — `[ERROR] Instance not found.` |

---

### Operation: DescribePlugins — List Installed Plugins

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | `ve elasticsearch DescribeInstances` | Instance Running | HALT |

#### Execution

```bash
# List all installed plugins
ve elasticsearch DescribePlugins \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}"

# Check specific plugin
ve elasticsearch DescribePlugins \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --PluginName "{{user.plugin_name}}"
```

#### Present to User

| Field | JSON Path | Notes |
|-------|-----------|-------|
| Plugin Name | `.Result.Plugins[].PluginName` | Plugin identifier |
| Version | `.Result.Plugins[].Version` | Plugin version |
| Status | `.Result.Plugins[].Status` | Installed or installing |

---

### Operation: ModifyNodeCount — Scale Nodes Horizontally

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | `ve elasticsearch DescribeInstances` | Instance Running | HALT |
| No ongoing operation | Instance status | Not modifying | Wait for completion |
| Valid node count | Check min/max limits | Within range | Adjust count |

#### Execution

```bash
# Scale data nodes
ve elasticsearch ModifyNodeCount \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --NodeNumber {{user.new_node_count}}
```

#### Validation

```bash
# Poll until spec change completes
for i in $(seq 1 60); do
  STATUS=$(ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" | jq -r '.Result.Instances[0].Status')
  COUNT=$(ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" | jq -r '.Result.Instances[0].NodeNumber')
  echo "Status: $STATUS, NodeCount: $COUNT (attempt $i/60)"
  [ "$STATUS" = "Running" ] && [ "$COUNT" -eq "{{user.new_node_count}}" ] && break
  sleep 30
done
```

#### Failure Recovery

| Error pattern | Max retries | Recovery |
|--------------|-------------|---------|
| `InvalidInstance.NotFound` | 0 | HALT; instance does not exist — `[ERROR] Instance not found.` |
| `InvalidInstanceStatus` | 3 | 10s wait for stable status — `[WARNING] Instance busy. Retrying...` |
| `QuotaExceeded` | 0 | HALT; node quota exceeded — `[ERROR] Node quota exceeded. Request quota increase.` |

---

## Error Taxonomy (≥ 10 Codes)

| Error Code | Meaning | Resolution |
|------------|---------|-----------|
| `InvalidParameter` | Request parameter invalid | 0 retries; **RETRY** — Fix parameter and retry |
| `InvalidInstance.NotFound` | Instance does not exist | 0 retries; **HALT** — verify instance ID |
| `InvalidVpc.NotFound` | VPC does not exist | 0 retries; **HALT** — create VPC first |
| `InvalidSubnet.NotFound` | Subnet does not exist | 0 retries; **HALT** — create subnet first |
| `IndexAlreadyExists` | Index already exists | 0 retries; **RETRY** — Use different name or delete first |
| `IndexNotFound` | Index does not exist | 0 retries; **RETRY** — Verify index name or create index |
| `PluginAlreadyExists` | Plugin already installed | 0 retries; **RETRY** — Plugin already installed — skip |
| `PluginNotFound` | Plugin not available | 0 retries; **RETRY** — Verify plugin name and ES version |
| `PluginIncompatible` | Plugin incompatible with ES version | 0 retries; **HALT** — select compatible plugin |
| `QuotaExceeded` | Resource quota exceeded | 0 retries; **HALT** — request quota increase |
| `InsufficientBalance` | Account balance insufficient | 0 retries; **HALT** — recharge account |
| `Unauthorized` | IAM permission denied | 0 retries; **HALT** — check IAM policies |
| `InvalidInstanceStatus` | Instance status not valid for operation | 3 retries/10s; **RETRY** — Wait and retry |
| `IncompatibleVersion` | Upgrade version not supported | 0 retries; **HALT** — check version compatibility |
| `InternalError` | Server-side error | 3 retries/2s/4s/8s; **RETRY** — Retry with backoff |
| `Throttling` | Rate limit exceeded | 3 retries/1s/2s/4s; **RETRY** — Back off and retry |
| `ClusterHealthNotGreen` | Cluster health not Green/Yellow | 0 retries; **HALT** — fix cluster health first |
| `SnapshotInProgress` | Snapshot already in progress | 3 retries/10s; **RETRY** — Wait and retry |

## Prerequisites

1. **Install `ve` CLI**:

   ```bash
   curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/{{env.ve_version}}/ve-linux-amd64 -o /usr/local/bin/ve
   chmod +x /usr/local/bin/ve
   ve version
   ```

2. **Configure Credentials**:

   ```bash
   export VOLCENGINE_ACCESS_KEY="{{env.VOLCENGINE_ACCESS_KEY}}"
   export VOLCENGINE_SECRET_KEY="{{env.VOLCENGINE_SECRET_KEY}}"
   export VOLCENGINE_REGION="{{env.VOLCENGINE_REGION}}"
   ```

3. **Verify Configuration**:

   ```bash
   ve elasticsearch DescribeInstances --Region {{env.VOLCENGINE_REGION}}
   ```

## Reference Directory

- [Core Concepts](references/core-concepts.md)
- [API & SDK Usage](references/api-sdk-usage.md)
- [CLI Usage](references/cli-usage.md)
- [Troubleshooting Guide](references/troubleshooting.md)
- [Monitoring & Alerts](references/monitoring.md)
- [Integration](references/integration.md)
- [Knowledge Base](references/knowledge-base.md)
- [SecurityOps (Advanced)](references/advanced/securityops.md) — Cluster security baseline, access control, data protection, incident response

## Operational Best Practices

- **Instance naming:** use descriptive names with environment prefix (e.g., `prod-search`, `dev-logs`)
- **Index naming:** use lowercase with date suffix for time-series data (e.g., `logs-2026.05.31`)
- **Shard sizing:** aim for 10-50GB per shard; avoid over-sharding (>1000 shards per node)
- **Replica count:** use 1 for production (provides high availability)
- **Snapshot schedule:** configure automated daily snapshots for all production instances
- **Version upgrades:** always take a snapshot before upgrading; test in staging first
- **Storage thresholds:** set up alerts for disk usage > 75% and 85%
- **Kibana:** enable for production instances; restrict access via IP whitelist
- **Plugins:** only install verified, compatible plugins; test in non-production first
- **Scaling:** scale storage before scaling compute; allow rebalancing time after changes
- **Monitoring:** set up alerts for cluster health Yellow/Red, high thread pool queue, and circuit breaker trips
- **FinOps:** right-size instances based on actual usage; use PrePaid for stable workloads to save cost

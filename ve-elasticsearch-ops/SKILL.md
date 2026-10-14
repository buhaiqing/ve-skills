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

# See Execution Flows in references/execution-flows.md


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

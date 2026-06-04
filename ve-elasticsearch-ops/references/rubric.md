---
name: ve-elasticsearch-ops-rubric
description: GCL rubric for ve-elasticsearch-ops. Destructive: DeleteInstance, DeleteIndex. State-changing: ModifyInstanceSpec, RestartInstance, UpgradeVersion, Enable/DisableKibana, Install/UninstallPlugin, ModifyInstanceChargeType.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-elasticsearch-ops
  gcl_role: critic_input
  rubric_dimensions: 5
  default_max_iter: 2
---

# GCL Rubric — ve-elasticsearch-ops

## 0. Operation Tier

| Tier | Operations | `max_iter` | Safety floor |
|---|---|---|---|
| **Destructive** | `DeleteInstance`, `DeleteIndex` (with caution for cluster-wide index), `DeleteSnapshot` | 2 | 1.0 (mandatory) |
| **State-changing** | `ModifyInstanceSpec`, `RestartInstance`, `UpgradeVersion`, `EnableKibana`, `DisableKibana`, `InstallPlugin`, `UninstallPlugin`, `ModifyInstanceChargeType` | 2 | 1.0 (mandatory) |
| **Mutating** | `CreateInstance`, `CreateIndex`, `CreateSnapshot`, `KibanaCreateUser` | 2 | ≥ 0.5 |
| **Read-only** | `DescribeInstances`, `DescribeIndices`, `DescribeSnapshots`, `DescribePlugins`, `DescribeKibana` | 3 | ≥ 0 |

**Safety = 0 → ABORT.**

## 1. Correctness (0 / 0.5 / 1)

- **CreateInstance**: `$.Result.InstanceId` present; status `Running` within 1800s.
- **DeleteInstance**: gone within poll budget.
- **CreateIndex**: `DescribeIndices` shows the index with expected shard/replica count.
- **DeleteIndex**: index no longer listed.
- **ModifyInstanceSpec**: new spec reflected; cluster health `green`.
- **InstallPlugin**: `DescribePlugins` shows the plugin.

## 2. Safety (0 / 1)

- **DeleteInstance**: explicit confirmation; ALL data + indices + snapshots lost.
- **DeleteIndex**: explicit confirmation naming the index; warn about data loss.
- **ModifyInstanceSpec**: warn 60-1800s downtime (cluster rebalancing).
- **RestartInstance**: warn about search/index interruption during rolling restart.
- **UpgradeVersion**: warn that rolling upgrade causes 30-60s per node interruption.
- **UninstallPlugin**: warn that plugin functionality stops immediately.
- **DisableKibana**: warn that Kibana access is lost.
- **VOLCENGINE_SECRET_KEY** never in trace. Kibana password masked.

## 3. Idempotency

`CreateInstance` NOT idempotent. `CreateIndex` NOT idempotent. `UpgradeVersion` once per version (cannot downgrade). `RestartInstance` safe.

## 4. Traceability

Full command, RequestId, validation, retries. Kibana password masked.

## 5. Spec Compliance

Dual-path (ve elasticsearch + Go SDK). ≥ 10 ES error codes. Delegation: VPC→ve-vpc-ops, backup→ve-tos-ops, IAM→ve-iam-ops.

## Changelog

| Version | Date | Change |
|---|---|---|
| 1.0.0 | 2026-06-04 | Initial GCL rubric for ve-elasticsearch-ops |
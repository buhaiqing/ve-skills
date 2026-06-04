---
name: ve-mongodb-ops-rubric
description: GCL rubric for ve-mongodb-ops. Destructive: DeleteDBInstance (irreversible). State-changing: ModifyDBInstanceSpec, RestartDBInstance, ModifyDBInstanceIPList.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-mongodb-ops
  gcl_role: critic_input
  rubric_dimensions: 5
  default_max_iter: 2
---

# GCL Rubric — ve-mongodb-ops

## 0. Operation Tier

| Tier | Operations | `max_iter` | Safety floor |
|---|---|---|---|
| **Destructive** | `DeleteDBInstance` | 2 | 1.0 (mandatory) |
| **State-changing** | `ModifyDBInstanceSpec`, `RestartDBInstance`, `ModifyDBInstanceIPList`, `CreateDatabase`, `CreateDBAccount` | 2 | 1.0 (mandatory) |
| **Mutating** | `CreateDBInstance`, `CreateBackup`, `RestoreToNewInstance` | 2 | ≥ 0.5 |
| **Read-only** | `DescribeDBInstances`, `DescribeDBInstanceDetail`, `DescribeDBAccounts`, `DescribeDatabases`, `DescribeBackups` | 3 | ≥ 0 |

**Safety = 0 → ABORT.**

## 1. Correctness (0 / 0.5 / 1)

- **CreateDBInstance**: `$.Result.InstanceId` present; status `RUNNING` within 600s.
- **DeleteDBInstance**: gone within poll budget.
- **ModifyDBInstanceSpec**: new spec reflected; status returns `RUNNING`.
- **CreateDatabase / CreateDBAccount**: `DescribeDatabases` / `DescribeDBAccounts` confirms.

## 2. Safety (0 / 1)

- **DeleteDBInstance**: explicit confirmation; check deletion protection; ALL data lost.
- **ModifyDBInstanceSpec**: warn 60-900s downtime (replica set election).
- **RestartDBInstance**: warn about connection interruption.
- **VOLCENGINE_SECRET_KEY** never in trace. DB password masked.

## 3. Idempotency

`CreateDBInstance` NOT idempotent. `CreateDatabase` pre-check. `CreateDBAccount` pre-check.

## 4. Traceability

Full command, RequestId, validation, retries. Password masked.

## 5. Spec Compliance

Dual-path (ve mongodb + Go SDK). ≥ 15 MongoDB error codes. Delegation: VPC→ve-vpc-ops.

## Changelog

| Version | Date | Change |
|---|---|---|
| 1.0.0 | 2026-06-04 | Initial GCL rubric for ve-mongodb-ops |
---
name: ve-rds-pg-ops-rubric
description: GCL rubric for ve-rds-pg-ops. Destructive: DeleteDBInstance, DeleteDBAccount. State-changing: ModifyDBInstanceSpec, ModifyDBInstanceParameter, ModifyDBInstanceIPList.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-rds-pg-ops
  gcl_role: critic_input
  rubric_dimensions: 5
  default_max_iter: 2
---

# GCL Rubric — ve-rds-pg-ops

## 0. Operation Tier

| Tier | Operations | `max_iter` | Safety floor |
|---|---|---|---|
| **Destructive** | `DeleteDBInstance`, `DeleteDBAccount` | 2 | 1.0 (mandatory) |
| **State-changing** | `ModifyDBInstanceSpec`, `ModifyDBInstanceParameter`, `ModifyDBInstanceIPList`, `RestartDBInstance` | 2 | 1.0 (mandatory) |
| **Mutating** | `CreateDBInstance`, `CreateDBAccount`, `CreateBackup`, `RestoreToNewInstance`, `CreateReadOnlyNode` | 2 | ≥ 0.5 |
| **Read-only** | `DescribeDBInstances`, `DescribeDBInstanceDetail`, `DescribeDBInstanceParameters`, `DescribeAccounts`, `DescribeBackups` | 3 | ≥ 0 |

**Safety = 0 → ABORT.**

## 1. Correctness (0 / 0.5 / 1)

| Score | Criteria |
|---|---|
| **1** | Instance id/status/spec match request; `DescribeDBInstanceDetail` confirms. |
| **0.5** | Minor mismatch but core id and state correct. |
| **0** | Wrong instance; state unchanged; instance not found. |

## 2. Safety (0 / 1)

- **DeleteDBInstance**: explicit confirmation; check deletion protection; irreversible data loss.
- **ModifyDBInstanceSpec**: warn 60-900s downtime.
- **DeleteDBAccount**: warn apps lose DB access.
- **VOLCENGINE_SECRET_KEY** never in trace. Password masked.

## 3. Idempotency

`Create*` not idempotent — pre-check. `Delete*` on already-deleted = no-op.

## 4. Traceability

Full command, RequestId, validation, retries. Password masked.

## 5. Spec Compliance

Dual-path (ve rds_postgresql + Go SDK). ≥ 15 PG error codes. Delegation: VPC→ve-vpc-ops.

## Changelog

| Version | Date | Change |
|---|---|---|
| 1.0.0 | 2026-06-04 | Initial GCL rubric for ve-rds-pg-ops |
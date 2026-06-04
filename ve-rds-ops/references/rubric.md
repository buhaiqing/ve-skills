---
name: ve-rds-ops-rubric
description: GCL rubric for ve-rds-ops (RDS MySQL variant). Destructive: DeleteDBInstance.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-rds-ops
  gcl_role: critic_input
  rubric_dimensions: 5
  default_max_iter: 2
---
# GCL Rubric — ve-rds-ops
## 0. Operation Tiers (see AGENTS.md §GCL for meta-rubric)
| Tier | Operations | max_iter | Safety |
|---|---|---|---|
| **Destructive** | DeleteDBInstance, DeleteDBAccount | 2 | 1.0 |
| **State-changing** | ModifyDBNodeSpec, ModifyDBInstanceParameter | 2 | 1.0 |
| **Mutating** | CreateDBInstance, CreateDBAccount, CreateBackup | 2 | ≥0.5 |
| **Read-only** | DescribeDBInstances, DescribeDBInstanceDetail, DescribeDBInstanceParameters, DescribeBackups | 3 | ≥0 |
Safety: DeleteDBInstance check deletion protection; irreversible. ModifyDBNodeSpec downtime 60-900s. Password masked. VOLCENGINE_SECRET_KEY never.
## Changelog
| 1.0.0 | 2026-06-04 | Initial GCL rubric for ve-rds-ops |
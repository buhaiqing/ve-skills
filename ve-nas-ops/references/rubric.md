---
name: ve-nas-ops-rubric
description: GCL rubric for ve-nas-ops. Destructive: DeleteFileSystem (data lost), DeleteMountTarget.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-nas-ops
  gcl_role: critic_input
  rubric_dimensions: 5
  default_max_iter: 3
---
# GCL Rubric — ve-nas-ops
## 0. Operation Tiers
| Tier | Operations | max_iter | Safety |
|---|---|---|---|
| **Destructive** | DeleteFileSystem, DeleteMountTarget | 3 | 1.0 |
| **State-changing** | ModifyFileSystem, ModifyMountTarget | 3 | 1.0 |
| **Mutating** | CreateFileSystem, CreateMountTarget | 3 | ≥0.5 |
| **Read-only** | DescribeFileSystems, ListMountTargets | 3 | ≥0 |
## 1-5. Dimensions (standard)
Safety: DeleteFileSystem warn ALL data LOST (files, snapshots, mount targets). DeleteMountTarget warn instance disconnection. VOLCENGINE_SECRET_KEY never.
## Changelog
| 1.0.0 | 2026-06-04 | Initial GCL rubric for ve-nas-ops |
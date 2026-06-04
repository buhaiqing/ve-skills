---
name: ve-vke-ops-rubric
description: GCL rubric for ve-vke-ops. Destructive: DeleteCluster (ALL workloads lost), DeleteNodePool.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-vke-ops
  gcl_role: critic_input
  rubric_dimensions: 5
  default_max_iter: 3
---
# GCL Rubric — ve-vke-ops
## 0. Operation Tiers
| Tier | Operations | max_iter | Safety |
|---|---|---|---|
| **Destructive** | DeleteCluster, DeleteNodePool | 3 | 1.0 |
| **State-changing** | UpdateCluster, UpdateNodePool | 3 | 1.0 |
| **Mutating** | CreateCluster, CreateNodePool | 3 | ≥0.5 |
| **Read-only** | ListClusters, DescribeCluster, ListNodePools, DescribeNodePool | 3 | ≥0 |
## 1-5. Dimensions (standard)
Safety: DeleteCluster warn ALL k8s workloads + data + config LOST. DeleteNodePool warn pods on those nodes rescheduled. VOLCENGINE_SECRET_KEY never.
## Changelog
| 1.0.0 | 2026-06-04 | Initial GCL rubric for ve-vke-ops |
---
name: ve-sls-ops-rubric
description: GCL rubric for ve-sls-ops. Optional tier, max_iter=5.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-sls-ops
  gcl_role: critic_input
  rubric_dimensions: 5
  default_max_iter: 5
---
# GCL Rubric — ve-sls-ops
## 0. Operation Tiers
| Tier | Operations | max_iter | Safety |
|---|---|---|---|
| **Destructive** | DeleteProject, DeleteTopic, DeleteShipper | 5 | 1.0 |
| **State-changing** | CreateProject, CreateTopic, CreateShipper, ModifyProject, ModifyTopic, ModifyShipper, CreateIndex, ModifyIndex | 5 | 1.0 |
| **Mutating** | — (all state-changing) | 5 | ≥0.5 |
| **Read-only** | DescribeProjects, DescribeTopics, DescribeShippers, SearchLog, DescribeLogHistogram | 5 | ≥0 |
Safety: DeleteProject ALL topics + shippers + logs LOST. DeleteTopic — log data and index LOST. SearchLog query: NEVER include user credentials in query strings. VOLCENGINE_SECRET_KEY never.
## Changelog
| 1.0.0 | 2026-06-04 | Initial GCL rubric for ve-sls-ops |
---
name: ve-fg-ops-rubric
description: GCL rubric for ve-fg-ops. Destructive: DeleteFunction, DeleteTrigger.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-fg-ops
  gcl_role: critic_input
  rubric_dimensions: 5
  default_max_iter: 3
---
# GCL Rubric — ve-fg-ops
## 0. Operation Tiers
| Tier | Operations | max_iter | Safety |
|---|---|---|---|
| **Destructive** | DeleteFunction, DeleteTrigger | 3 | 1.0 |
| **State-changing** | UpdateFunction, PublishVersion, CreateTrigger | 3 | 1.0 |
| **Mutating** | CreateFunction | 3 | ≥0.5 |
| **Read-only** | GetFunction, ListFunctions, InvokeFunction | 3 | ≥0 |
## 1-5. Dimensions (standard)
Safety: DeleteFunction warn ALL versions + aliases + triggers LOST. DeleteTrigger warn function stops responding to events. UpdateFunction: warn about in-flight invocations. VOLCENGINE_SECRET_KEY never.
## Changelog
| 1.0.0 | 2026-06-04 | Initial GCL rubric for ve-fg-ops |
---
name: ve-cms-ops-rubric
description: GCL rubric for ve-cms-ops. Destructive: DeleteAlarmRule, DeleteAlarmStrategy (monitoring blackout).
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-cms-ops
  gcl_role: critic_input
  rubric_dimensions: 5
  default_max_iter: 3
---
# GCL Rubric — ve-cms-ops
## 0. Operation Tiers
| Tier | Operations | max_iter | Safety |
|---|---|---|---|
| **Destructive** | DeleteAlarmRule, DeleteAlarmStrategy | 3 | 1.0 |
| **State-changing** | CreateAlarmRule, ModifyAlarmRule, DisableAlarmRule, CreateAlarmStrategy, ModifyAlarmStrategy | 3 | 1.0 |
| **Mutating** | — (all alarm operations are state-changing) | 3 | ≥0.5 |
| **Read-only** | DescribeAlarmRules, DescribeAlarmStrategies, GetMetricData | 3 | ≥0 |
## 1-5. Dimensions (standard)
Safety: DeleteAlarmRule warn monitoring blackout for affected resources. DisableAlarmRule warn no alert notification. CreateAlarmRule with empty notification channel: warn no notification. VOLCENGINE_SECRET_KEY never.
## Changelog
| 1.0.0 | 2026-06-04 | Initial GCL rubric for ve-cms-ops |
---
name: ve-alb-ops-rubric
description: GCL rubric for ve-alb-ops. Destructive: DeleteLoadBalancer, DeleteListener, DeleteRule, DeleteServerGroup.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-alb-ops
  gcl_role: critic_input
  rubric_dimensions: 5
  default_max_iter: 3
---
# GCL Rubric — ve-alb-ops
## 0. Operation Tiers
| Tier | Operations | max_iter | Safety |
|---|---|---|---|
| **Destructive** | DeleteLoadBalancer, DeleteListener, DeleteRule, DeleteServerGroup | 3 | 1.0 |
| **State-changing** | AddBackendServers, RemoveBackendServers, ModifyListenerAttributes, ModifyRuleAttributes | 3 | 1.0 |
| **Mutating** | CreateLoadBalancer, CreateListener, CreateRule, CreateServerGroup | 3 | ≥0.5 |
| **Read-only** | DescribeLoadBalancers, DescribeListeners, DescribeRules, DescribeServerGroups | 3 | ≥0 |
## 1-5. Dimensions (standard)
Safety: DeleteLoadBalancer warn ALL listeners/rules/server-groups lost — traffic cut. DeleteServerGroup in use by listener: warn about routing disruption. RemoveBackendServers >50%: warn. VOLCENGINE_SECRET_KEY never.
## Changelog
| 1.0.0 | 2026-06-04 | Initial GCL rubric for ve-alb-ops |
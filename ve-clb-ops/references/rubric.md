---
name: ve-clb-ops-rubric
description: GCL rubric for ve-clb-ops. Destructive: DeleteLoadBalancer (traffic cut), DeleteListener.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-clb-ops
  gcl_role: critic_input
  rubric_dimensions: 5
  default_max_iter: 3
---
# GCL Rubric — ve-clb-ops
## 0. Operation Tiers
| Tier | Operations | max_iter | Safety |
|---|---|---|---|
| **Destructive** | DeleteLoadBalancer, DeleteListener | 3 | 1.0 |
| **State-changing** | RemoveBackendServers, AddBackendServers, SetHealthCheckConfig, ModifyLoadBalancerAttributes | 3 | 1.0 |
| **Mutating** | CreateLoadBalancer, CreateListener | 3 | ≥0.5 |
| **Read-only** | DescribeLoadBalancers, DescribeListeners, DescribeBackendServers | 3 | ≥0 |
## 1-5. Dimensions (standard)
Safety: DeleteLoadBalancer warn traffic to ALL backends cut. RemoveBackendServers >50%: warn about capacity drop. VOLCENGINE_SECRET_KEY never in trace.
## Changelog
| 1.0.0 | 2026-06-04 | Initial GCL rubric for ve-clb-ops |
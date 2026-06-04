---
name: ve-nat-ops-rubric
description: GCL rubric for ve-nat-ops. Destructive: DeleteNatGateway (removes SNAT/DNAT).
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-nat-ops
  gcl_role: critic_input
  rubric_dimensions: 5
  default_max_iter: 3
---
# GCL Rubric — ve-nat-ops
## 0. Operation Tiers
| Tier | Operations | max_iter | Safety |
|---|---|---|---|
| **Destructive** | DeleteNatGateway, DeleteSnatRule, DeleteDnatRule | 3 | 1.0 |
| **State-changing** | CreateSnatRule, CreateDnatRule | 3 | 1.0 |
| **Mutating** | CreateNatGateway, ModifyNatGatewayAttribute | 3 | ≥0.5 |
| **Read-only** | DescribeNatGateways, DescribeSnatRules, DescribeDnatRules | 3 | ≥0 |
## 1-5. Dimensions (standard)
Safety: DeleteNatGateway warn about ALL SNAT/DNAT rules removed — instances lose internet access. CreateDnatRule with 0.0.0.0/0 on sensitive port: warn. VOLCENGINE_SECRET_KEY never in trace.
## Changelog
| 1.0.0 | 2026-06-04 | Initial GCL rubric for ve-nat-ops |
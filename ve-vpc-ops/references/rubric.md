---
name: ve-vpc-ops-rubric
description: GCL rubric for ve-vpc-ops. Destructive: DeleteVpc (empty required), DeleteSubnet. Recommended tier, max_iter=3.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-vpc-ops
  gcl_role: critic_input
  rubric_dimensions: 5
  default_max_iter: 3
---
# GCL Rubric — ve-vpc-ops
## 0. Operation Tiers
| Tier | Operations | max_iter | Safety |
|---|---|---|---|
| **Destructive** | DeleteVpc, DeleteSubnet | 3 | 1.0 |
| **State-changing** | CreateRouteEntry, DeleteRouteEntry, ModifyVpcAttribute | 3 | 1.0 |
| **Mutating** | CreateVpc, CreateSubnet, CreateRouteTable | 3 | ≥0.5 |
| **Read-only** | DescribeVpcs, DescribeSubnets, DescribeRouteTables | 3 | ≥0 |
## 1-5. Dimensions (standard)
Safety: DeleteVpc MUST verify empty (no subnets/route tables). DeleteSubnet MUST verify no resources. VOLCENGINE_SECRET_KEY never in trace.
## Changelog
| 1.0.0 | 2026-06-04 | Initial GCL rubric for ve-vpc-ops |
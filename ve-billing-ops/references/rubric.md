---
name: ve-billing-ops-rubric
description: GCL rubric for ve-billing-ops. Optional tier, max_iter=5. Read-mostly — financial data visibility.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-billing-ops
  gcl_role: critic_input
  rubric_dimensions: 5
  default_max_iter: 5
---
# GCL Rubric — ve-billing-ops
## 0. Operation Tiers
| Tier | Operations | max_iter | Safety |
|---|---|---|---|
| **Destructive** | DeleteBudget | 5 | 1.0 |
| **State-changing** | CreateBudget, UpdateBudget | 5 | 1.0 |
| **Mutating** | — (all state-changing) | 5 | ≥0.5 |
| **Read-only** | DescribeBills, DescribeBillDetail, DescribeBalance, DescribeBudgets, DescribeReservedInstances | 5 | ≥0 |
Safety: DeleteBudget — budget tracking and alerts stop. CreateBudget with wrong amount — incorrect cost control. VOLCENGINE_SECRET_KEY never in trace. Financial data is sensitive — output only aggregated totals, not per-resource details without user request.
## Changelog
| 1.0.0 | 2026-06-04 | Initial GCL rubric for ve-billing-ops |
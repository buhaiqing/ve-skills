# FinOps — IAM Governance Efficiency

> IAM is **free** — FinOps focus = governance efficiency + cost avoidance.

## Cost Overview

IAM (users, roles, policies, API calls) has **zero direct cost**. Costs arise indirectly:
- **Management overhead** → too many roles/policies = audit drag, human error risk
- **Over-permissioned roles** → potential resource abuse (e.g., someone with `ecs:*` starts expensive instances)
- **Cross-account sprawl** → untracked trusted relationships increase blast radius

## Cost Optimization

| Pattern | Action | Savings |
|---------|--------|---------|
| 🧹 Orphaned policies | Delete un-attached IAM policies | Reduced audit scope |
| 🔗 Inline → managed | Convert inline policies to reusable managed policies | + Maintainability |
| 🔑 Stale access keys | Deactivate keys unused > 90d | + Security posture |
| 🏷️ Cross-account trust | Review & prune external accounts in trust policies | Reduced blast radius |
| 📋 Role consolidation | Merge roles with identical permission sets | Lower cognitive overhead |

## Cost Anomaly Detection

| Signal | What to Check | Response |
|--------|---------------|----------|
| ⚠️ Role count spike > 20% MoM | New roles created without review | Audit & consolidate |
| ⚠️ Policy attachment surge | Bulk permission grants | Verify least-privilege |
| ⚠️ Access key age > 180d | Keys not rotated | Force rotation |
| 🚨 Root user activity | Direct root usage | Enforce role-based access |

## Pricing

IAM is **free** → no API pricing query needed. Cost is in **governance debt** and **incident risk**.

## Related Resources

- [Billing FinOps → cost allocation & budgets](../../../ve-billing-ops/references/advanced/finops.md)
- [IAM SecurityOps → risk-based least privilege](../advanced/securityops.md)
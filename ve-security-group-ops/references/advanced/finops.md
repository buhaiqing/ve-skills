# FinOps — Security Group Governance

> Security Groups are **free** — FinOps focus = incident cost avoidance + audit efficiency.

## Cost Overview

SGs have **zero direct billing**. Costs are indirect:
- **Audit overhead** → 100+ orphaned SGs waste compliance review time
- **Incident cost** → over-permissive rules (0.0.0.0/0 on critical ports) → breach potential
- **Operational drag** → too many SGs per ENI → troubleshooting latency

## Cost Optimization

| Pattern | Action | Savings |
|---------|--------|---------|
| 🧹 Orphaned SGs | Delete SGs not associated with any ENI | Cleaner inventory |
| 🔗 Rule consolidation | Merge identical CIDR rules, use prefix lists | 10-30% rule count ↓ |
| 🔒 Least-privilege audit | Quarterly review of 0.0.0.0/0 → narrow to specific CIDRs | ↓ breach surface area |
| 🏷️ Tag for cost center | Tag every SG → `CostCenter`, `Owner` | + Cost attribution |
| 🗑️ Stale rules | Remove rules for decommissioned instance CIDRs | Cleaner, more secure |

## Cost Anomaly Detection

| Signal | What to Check | Response |
|--------|---------------|----------|
| ⚠️ SG count spike | New SGs created without cleanup plan | Audit & consolidate |
| ⚠️ 0.0.0.0/0 on ports 22/3389 | Public SSH/RDP exposed | Restrict immediately |
| ⚠️ Unused rule ratio > 30% | Rules matching no active ENI | Prune stale rules |
| 🚨 Default SG with all traffic | VPC default SG still wide open | Harden default rules |

## Pricing

SGs are **free** → no pricing query needed. Cost comes from **security incidents prevented** and **audit hours saved**.

## Related Resources

- [Billing FinOps → tagging & cost allocation](../../../ve-billing-ops/references/advanced/finops.md)
- [Security Group SecurityOps → least-privilege patterns](../advanced/securityops.md)
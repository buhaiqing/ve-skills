# FinOps — Billing Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7.

## Billing Model Comparison

| Model | Discount | Commitment | Best For |
|-------|----------|------------|----------|
| PostPaid | 0% | None | Variable workloads, testing |
| PrePaid (1 year) | ~30-35% | 12 months | Steady production workloads |
| PrePaid (3 years) | ~45-50% | 36 months | Long-term infrastructure |
| Spot/Preemptible | ~60-90% | Interruptible | Fault-tolerant batch |

## Cost Optimization Quick Reference

| Situation | Action | 💰 Savings |
|-----------|--------|---------|
| Unused instance > 30 days | Delete or stop | Up to 100% |
| Over-provisioned instance | Right-size down | 25-75% |
| Steady workload on PostPaid | Switch to PrePaid (1yr) | ~35% |
| Long-term workload | Switch to PrePaid (3yr) | ~50% |
| Fault-tolerant batch | Use Spot instances | 60-90% |
| Unattached storage | Snapshot + delete | 100% |
| Idle EIP/NAT | Release | 100% |
| Snapshot > 90 days old | Delete | 100% |
| Low-hit CDN domain | Consolidate or prune | Variable |

## Query Current Prices

```bash
# Query billing summary for current month
ve billing DescribeBillSummaryByMonth --body '{"BillingPeriod":"{{env.BILLING_MONTH}}"}'

# Always query the pricing API for current rates
ve ecs DescribeInstanceTypes --body '{"InstanceTypeIds":["ecs.g3i.large"]}'
```

> Prices change over time — always query the Price API for current rates rather than relying on hardcoded tables.

## Budget Alert Tiers

| Threshold | Severity | Action |
|-----------|----------|--------|
| 80% | Warning | Review cost trends, plan optimization |
| 90% | Critical | Immediate cost review, pause non-essential |
| 100% | Blocking | Stop non-production resources, escalate |

## Tag-Based Cost Allocation

```bash
# List cost tags
ve billing ListCostTags --body '{}'
# Tag untagged resources
ve ecs CreateTags --body '{"ResourceIds":["i-xxx"],"Tags":[{"Key":"CostCenter","Value":"Platform"}]}'
```

> Tagging is the foundation of cost allocation — ensure ≥ 80% resources tagged.
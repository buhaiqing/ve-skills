# FinOps — ECS Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7.

## Billing Model Comparison

| Model | Discount | Commitment | Best For |
|-------|----------|------------|----------|
| PostPaid | 0% | None | Variable, testing |
| PrePaid (1yr) | ~35% | 12 months | Steady production |
| PrePaid (3yr) | ~50% | 36 months | Long-term infra |
| Spot | 60-90% | Interruptible | Fault-tolerant batch |

## Product-Specific Cost Optimization

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Use Spot for batch workloads | Stateless/non-critical → Spot | 60-90% |
| Right-size underutilized | CPU avg < 15% for 7d → downsize spec | ~50% |
| Convert steady PostPaid → PrePaid | Running > 30d → 1yr PrePaid | ~35% |
| Delete orphaned volumes | Unattached disk > 7d → snapshot + delete | 100% |

## Cost Anomaly Detection

| Warning Sign | Check Command | Action |
|-------------|---------------|--------|
| 💰 Unexpected instance count spike | `ve ecs DescribeInstances` | Investigate auto-scaling or rogue deployment |
| ⚠️ Spot price surge | `ve ecs DescribeSpotPriceHistory` | Fallback to PostPaid or switch AZ |
| 📊 CPU credit exhaustion (burstable t5) | `ve ecs DescribeInstanceTypes` for t5 specs | Switch to standard instance type |

## Query Current Pricing

```bash
# Query instance type pricing for current region
ve ecs DescribeInstanceTypes --body '{"InstanceTypeIds":["ecs.g3i.large"]}'
```

> 💡 For billing-level queries, use `ve billing DescribeBillSummaryByMonth`.

## Related Resources

- [FinOps — Billing Cost Optimization](../../ve-billing-ops/references/advanced/finops.md)
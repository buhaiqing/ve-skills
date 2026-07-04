# FinOps — ECS Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Billing Model Comparison

| Model | Discount | Commitment | Best For |
|-------|----------|------------|----------|
| PostPaid | 0% | None | Variable workloads, testing |
| PrePaid (1 year) | ~35% | 12 months | Steady production workloads |
| PrePaid (3 years) | ~50% | 36 months | Long-term infrastructure |
| Spot | ~60-90% | Interruptible | Batch processing, fault-tolerant |

## Cost Per Instance Type

Pricing varies by region, billing model, and instance family. Query current prices:

```bash
# Describe price for a specific instance type (cn-beijing, PostPaid)
ve ecs DescribeInstanceTypes --InstanceTypeIds '["ecs.g3i.large"]' | jq '.Result.InstanceTypes[] | {InstanceType, Price}'
```

> Prices change over time — always query the Price API for current rates rather than relying on hardcoded tables.

## Cost Optimization Quick Reference

| Situation | Action | 💰 Savings |
|-----------|--------|---------|
| Instance running > 30 days | Convert PostPaid → PrePaid | ~35% |
| CPU avg < 15% for 7 days | Right-size down | 25-75% |
| Non-critical batch workload | Use Spot | 60-90% |
| Stopped instance > 7 days | Delete or snapshot + delete | 100% |
| Unattached disk | Snapshot + delete | 100% |
| Snapshot > 90 days old | Delete | 100% |
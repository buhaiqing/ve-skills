# FinOps — ARK Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7.

## Billing Model Comparison

| Model | Discount | Commitment | Best For |
|-------|----------|------------|----------|
| PostPaid | 0% | None | Variable workloads |
| PrePaid | ~35% | 12 months | Steady production |

## Cost Optimization Quick Reference

| Situation | Action | Savings |
|-----------|--------|---------|
| Over-provisioned ARK | Right-size instance type | 25-50% |
| Long-running cluster | PrePaid conversion | ~35% |
| Unused cluster | Stop or delete | Up to 100% |
| Idle worker nodes | Scale to zero | Up to 100% |

> Always query the pricing API for current rates — never hardcode prices.

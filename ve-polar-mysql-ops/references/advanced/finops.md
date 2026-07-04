# FinOps — PolarDB MySQL Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Billing Model Comparison

| Model | Discount | Commitment | Best For |
|-------|----------|------------|----------|
| PostPaid | 0% | None | Variable workloads, testing |
| PrePaid (1 year) | ~35% | 12 months | Steady production |
| PrePaid (3 years) | ~50% | 36 months | Long-term infrastructure |
| Spot | ~60-90% | Interruptible | Fault-tolerant batch |

## Cost Optimization Quick Reference

| Situation | Action | Savings |
|-----------|--------|---------|
| Instance running > 30 days | Convert PostPaid → PrePaid | ~35% |
| Idle resource | Stop or delete | Up to 100% |
| Over-provisioned | Right-size down | 25-75% |
| Batch workload | Use Spot | 60-90% |
| Unused attached resource | Detach and delete | 100% |
| Long-term storage | Use lifecycle policy | 50-80% |

## Query Current Prices

```bash
# Always query the pricing API for current rates
ve polar-mysql DescribePrice --<Param> value
```

> Prices change over time — never rely on hardcoded tables.

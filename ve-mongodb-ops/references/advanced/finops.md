# FinOps — MongoDB Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Cost Optimization Tips

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Right-size instance type | Sharded cluster nodes over-provisioned → downsize to smaller spec | ~50% |
| Convert steady PostPaid to PrePaid | Running > 30d → 1yr PrePaid | ~35% |
| Optimize storage | Reduce oplog size and backup retention period | 20-40% on storage cost |

## Query Current Prices

```bash
# Query current MongoDB instance pricing
ve mongodb DescribePrice
```

> ⚠️ Pricing data sourced from ve DescribePrice command — use `ve mongodb DescribePrice` for current quotes.
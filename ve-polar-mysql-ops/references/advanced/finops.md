# FinOps — PolarDB MySQL Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Cost Optimization Tips

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Right-size compute node | CPU avg < 15% → downsize PolarDB compute spec | ~50% |
| Use storage warm tier | Move infrequently accessed data to lower-cost storage tier | ~40-60% on storage cost |
| Convert steady PostPaid to PrePaid | Running > 30d → 1yr PrePaid | ~35% |

## Query Current Prices

```bash
# Query current PolarDB MySQL pricing
ve polar-mysql DescribePrice
```

> ⚠️ Pricing data sourced from ve DescribePrice command — use `ve polar-mysql DescribePrice` for current quotes.
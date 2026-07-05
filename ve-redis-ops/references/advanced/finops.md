# FinOps — Redis Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Cost Optimization Tips

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Right-size instance type | CPU avg < 20% → downsize to smaller spec (e.g., redis.4c8g → redis.2c4g) | ~50% |
| Convert steady PostPaid to PrePaid | Running > 30d → 1yr PrePaid | ~35% |
| Reduce backup retention | Keep only last 7 days of backups instead of 30 | 30-60% on storage |

## Query Current Prices

```bash
# Query current Redis instance pricing
ve redis DescribePrice
```

> ⚠️ Pricing data sourced from ve DescribePrice command — use `ve redis DescribePrice` for current quotes.
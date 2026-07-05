# FinOps — RDS MySQL Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Cost Optimization Tips

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Right-size instance type | CPU avg < 15% for 7d → downsize (e.g., rds.mysql.4c8g → rds.mysql.2c4g) | ~50% |
| Convert steady PostPaid to PrePaid | Running > 30d → 1yr PrePaid | ~35% |
| Reduce backup retention | Keep 7 days of backups instead of 30 for test/dev DBs | 30-60% on backup storage |

## Query Current Prices

```bash
# Query current RDS MySQL instance pricing
ve rds-mysql DescribePrice
```

> ⚠️ Pricing data sourced from ve DescribePrice command — use `ve rds-mysql DescribePrice` for current quotes.
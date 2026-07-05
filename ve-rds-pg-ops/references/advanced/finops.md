# FinOps — RDS PostgreSQL Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Cost Optimization Tips

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Right-size instance type | CPU avg < 15% → downsize (e.g., rds.pg.4c8g → rds.pg.2c4g) | ~50% |
| Convert steady PostPaid to PrePaid | Running > 30d → 1yr PrePaid | ~35% |
| Reduce backup retention | Keep 7 days instead of 30 for test/dev | 30-60% on backup storage cost |

## Query Current Prices

```bash
# Query current RDS PostgreSQL instance pricing
ve rds-pg DescribePrice
```

> ⚠️ Pricing data sourced from ve DescribePrice command — use `ve rds-pg DescribePrice` for current quotes.
# FinOps — NAS Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Cost Optimization Tips

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Use Infrequent Access tier | Move files not accessed in 30+ days to IA storage tier | ~60% on storage cost |
| Implement lifecycle policy | Automatically transition data between Performance, Capacity, and IA tiers | 40-80% |
| Delete orphaned snapshots | Old snapshots of deleted file systems — delete | 100% of snapshot cost |

## Query Current Prices

```bash
# Query current NAS pricing
ve nas DescribePrice
```

> ⚠️ Pricing data sourced from ve DescribePrice command — use `ve nas DescribePrice` for current quotes.
# FinOps — Cloud Monitor Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Cost Optimization Tips

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Reduce metric granularity | Change from 1-minute to 5-minute intervals for non-critical metrics | ~80% reduction in metric count |
| Remove unused alarms | Delete stale alarm rules with no active resources | 100% of idle alarm rule cost |
| Aggregate custom metrics | Batch multiple counter metrics into one gauge or use statistical metric | Variable |

## Query Current Prices

```bash
# Query current Cloud Monitor pricing
ve cms DescribePrice
```

> ⚠️ Pricing data sourced from ve DescribePrice command — use `ve cms DescribePrice` for current quotes.
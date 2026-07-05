# FinOps — CLB Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Cost Optimization Tips

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Delete idle CLB instances | No backend servers attached → delete immediately | 100% of hourly fee |
| Reduce data transfer | Co-locate clients and servers in same AZ | Variable (cross-AZ transfer fee) |
| Right-size specification | Use performance-guaranteed only when needed; use shared-performance for low traffic | Up to 60% |

## Query Current Prices

```bash
# Query current CLB pricing
ve clb DescribePrice
```

> ⚠️ Pricing data sourced from ve DescribePrice command — use `ve clb DescribePrice` for current quotes.
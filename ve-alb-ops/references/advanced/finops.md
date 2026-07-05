# FinOps — ALB Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Cost Optimization Tips

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Clean up idle listeners | Delete unused listener rules with no backend traffic | 100% of idle listener cost |
| Consolidate ALBs | Merge multiple ALBs serving same VPC via path-based routing | ~50% on ALB hourly fee |
| Reduce data transfer | Enable cross-zone load balancing only when needed | Variable (cross-zone transfer fee) |

## Query Current Prices

```bash
# Query current ALB pricing
ve alb DescribePrice
```

> ⚠️ Pricing data sourced from ve DescribePrice command — use `ve alb DescribePrice` for current quotes.
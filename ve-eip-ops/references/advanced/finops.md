# FinOps — EIP Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Cost Optimization Tips

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Release unused EIPs | Unassociated EIPs incur hourly fee — release if idle > 7 days | 100% of idle EIP cost |
| Right-size bandwidth | Monitor peak bandwidth and downgrade excess bandwidth tier | 20-50% |
| Use PrePaid bandwidth | Steady bandwidth needs → 1yr PrePaid for bandwidth package | ~35% |

## Query Current Prices

```bash
# Query current EIP pricing
ve eip DescribePrice
```

> ⚠️ Pricing data sourced from ve DescribePrice command — use `ve eip DescribePrice` for current quotes.
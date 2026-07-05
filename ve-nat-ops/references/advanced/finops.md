# FinOps — NAT Gateway Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Cost Optimization Tips

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Release idle NAT gateways | No active SNAT/DNAT rules or no traffic — delete | 100% of hourly fee |
| Right-size NAT spec | Monitor throughput — downgrade small specification if peak < 50% capacity | Up to 50% |
| Optimize SNAT entries | Consolidate SNAT entries to reduce number of required EIPs | Variable |

## Query Current Prices

```bash
# Query current NAT gateway pricing
ve nat DescribePrice
```

> ⚠️ Pricing data sourced from ve DescribePrice command — use `ve nat DescribePrice` for current quotes.
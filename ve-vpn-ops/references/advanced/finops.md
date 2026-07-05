# FinOps — VPN Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Cost Optimization Tips

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Delete unused VPN tunnels | No traffic for > 30 days → remove tunnel and customer gateway | 100% of idle tunnel fee |
| Consolidate connections | Multiple tunnels to same site → merge into one with BGP ECMP | ~50% on connection fees |
| Right-size bandwidth | Monitor peak usage and downgrade excess bandwidth spec | 20-40% |

## Query Current Prices

```bash
# Query current VPN gateway pricing
ve vpn DescribePrice
```

> ⚠️ Pricing data sourced from ve DescribePrice command — use `ve vpn DescribePrice` for current quotes.
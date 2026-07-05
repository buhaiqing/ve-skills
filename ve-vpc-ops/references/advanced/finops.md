# FinOps — VPC Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Cost Overview

VPC itself is free of charge. Costs come from associated resources: NAT Gateway, VPN Gateway, EIPs, and data transfer.

## Cost Optimization Tips

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Reduce cross-AZ traffic | Co-locate communicating resources in same AZ | Variable (data transfer fee) |
| Release unused resources | Delete idle NAT Gateways, VPN tunnels, and unassociated EIPs | 100% of idle resource cost |
| Use internal CLB instead of public | For intra-VPC traffic, avoid public-facing load balancers | EIP + data transfer cost |

## Query Current Prices

```bash
# Query current VPC-related pricing (NAT, VPN, EIP)
ve vpc DescribePrice
```

> ⚠️ Pricing data sourced from ve DescribePrice command — use `ve vpc DescribePrice` for current quotes.
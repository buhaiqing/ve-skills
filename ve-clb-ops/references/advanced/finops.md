# FinOps — CLB Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7.

## Cost Model Overview

CLB billing: (1) hourly fee by specification (shared-performance / performance-guaranteed), (2) data transfer for cross-AZ traffic within LB. Public CLB additionally requires EIP (hourly + bandwidth).

## Product-Specific Cost Optimization

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Delete idle CLB instances | No backend servers attached → delete | 100% of hourly fee |
| Use shared-performance spec | Low traffic/dev → shared-performance | Up to 60% vs guaranteed |
| Reduce cross-AZ data transfer | Co-locate clients + servers in same AZ | Variable (transfer fee) |
| Merge CLBs serving same VPC | Path-based routing → single CLB | ~50% on hourly fee |

## Cost Anomaly Detection

| Warning Sign | Check Command | Action |
|-------------|---------------|--------|
| 💰 Unexpected traffic spike | `ve clb DescribeLoadBalancers` — check metrics | Investigate DDoS or misconfigured DNS |
| ⚠️ Orphaned CLB with no backends | `ve clb DescribeLoadBalancers` — filter empty groups | Delete immediately |
| 📊 Cross-AZ transfer cost surge | `ve billing DescribeBillDetail` — filter `CLB` | Co-locate services in same AZ |

## Query Current Pricing

CLB pricing by spec/region. Use billing API for current rates:

```bash
ve billing DescribeBillDetail --body '{"BillingPeriod":"{{env.BILLING_MONTH}}","Product":"CLB"}'
```

## Related Resources

- [FinOps — Billing Cost Optimization](../../ve-billing-ops/references/advanced/finops.md)
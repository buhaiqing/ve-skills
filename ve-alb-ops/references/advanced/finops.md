# FinOps — ALB Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7.

## Cost Model Overview

ALB billing: (1) hourly fee by LCU (Loadbalancer Capacity Unit), (2) per-rule fee beyond free quota, (3) data transfer. ALB is typically higher-cost than CLB due to Layer 7 processing overhead.

## Product-Specific Cost Optimization

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Clean up unused listener rules | Delete unused path/host rules | 100% of idle rule cost |
| Consolidate ALBs per VPC | Merge via path-based routing | ~50% on hourly fee |
| Disable cross-zone LB if not needed | Only enable for multi-AZ workloads | Variable (data TX) |
| Reduce LCU consumption | Optimize new conn/s and active conn handling | Proportional to LCU |

## Cost Anomaly Detection

| Warning Sign | Check Command | Action |
|-------------|---------------|--------|
| 💰 LCU consumption surge | `ve alb DescribeLoadBalancers` — check LCU metrics | Investigate traffic pattern change |
| ⚠️ Orphaned ALB with no listeners | `ve alb DescribeListeners` — filter empty | Delete idle ALBs |
| 📊 Rule count cost spike | Count listener rules vs free quota | Consolidate rules, reduce wildcards |

## Query Current Pricing

ALB pricing by LCU/region. Use billing API for current rates:

```bash
ve billing DescribeBillDetail --body '{"BillingPeriod":"{{env.BILLING_MONTH}}","Product":"ALB"}'
```

## Related Resources

- [FinOps — Billing Cost Optimization](../../ve-billing-ops/references/advanced/finops.md)
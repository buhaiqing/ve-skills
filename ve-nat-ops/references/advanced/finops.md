# FinOps — NAT Gateway Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7.

## Cost Model Overview

NAT Gateway billing: (1) hourly fee by specification (Small/Medium/Large), (2) data transfer fee for outbound traffic. SNAT entries consume EIPs, adding separate hourly + bandwidth costs.

## Product-Specific Cost Optimization

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Delete idle NAT gateways | No active SNAT/DNAT → delete | 100% of hourly fee |
| Right-size NAT spec | Peak throughput < 50% capacity → downgrade spec | Up to 50% |
| Consolidate SNAT EIPs | Fewer SNAT entries → fewer EIPs needed | Variable (EIP fee) |
| Tune SNAT idle timeout | Shorter timeout → faster port reuse, fewer EIPs | Reduces EIP count |

## Cost Anomaly Detection

| Warning Sign | Check Command | Action |
|-------------|---------------|--------|
| 💰 Data transfer cost spike | `ve billing DescribeBillDetail` — filter `NAT` | Check for abnormal outbound traffic |
| ⚠️ SNAT port exhaustion errors | `ve nat DescribeSnatEntries` — check concurrency | Add EIPs or increase timeouts |
| 📊 Unexpected gateway count change | `ve nat DescribeNatGateways` | Verify authorization, delete if unauth'd |

## Query Current Pricing

NAT pricing by spec/region. Use billing API for current rates:

```bash
ve billing DescribeBillDetail --body '{"BillingPeriod":"{{env.BILLING_MONTH}}","Product":"NAT"}'
```

## Related Resources

- [FinOps — Billing Cost Optimization](../../ve-billing-ops/references/advanced/finops.md)
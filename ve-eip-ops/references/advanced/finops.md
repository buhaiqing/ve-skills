# FinOps — EIP Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7.

## Cost Model Overview

EIP billing includes: (1) hourly lease fee (varies by spec), (2) outbound data transfer (per GB), (3) optional bandwidth package. **Unassociated EIPs still incur hourly fee** → release if idle.

## Product-Specific Cost Optimization

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Release idle EIPs | Unassociated > 7d → release | 100% of idle fee |
| Right-size bandwidth tier | Monitor peak → downgrade excess tier | 20-50% |
| Use Shared Bandwidth Pool | Consolidate EIPs into shared plan | ~30% on data TX |
| Use PrePaid bandwidth | Steady needs → 1yr bandwidth package | ~35% |

## Cost Anomaly Detection

| Warning Sign | Check Command | Action |
|-------------|---------------|--------|
| 💰 Surge in unassociated EIPs | `ve eip DescribeEipAddresses` — filter `InstanceId:null` | Release orphaned EIPs |
| ⚠️ Bandwidth spike > 2x baseline | `ve billing DescribeBillDetail` — filter `EIP` | Investigate traffic source |
| 📊 Data transfer cost jump | Compare MoM `EIP` cost in billing | Trace by associated instance |

## Query Current Pricing

EIP pricing is zone/spec-dependent. Use billing API for current rates:

```bash
ve billing DescribeBillDetail --body '{"BillingPeriod":"{{env.BILLING_MONTH}}","Product":"EIP"}'
```

## Related Resources

- [FinOps — Billing Cost Optimization](../../ve-billing-ops/references/advanced/finops.md)
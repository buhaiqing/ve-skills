# FinOps — VPC Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7.

## Cost Overview

VPC itself is **free**. Costs come from associated resources: NAT Gateway (hourly + data TX), VPN Gateway (hourly), EIPs (hourly + bandwidth), and cross-AZ data transfer.

## Product-Specific Cost Optimization

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Reduce cross-AZ traffic | Co-locate communicating resources in same AZ | Variable (data TX fee) |
| Use internal CLB instead of public | Intra-VPC traffic → internal LB, no EIP | EIP + bandwidth cost |
| Delete idle NAT/VPN gateways | No active traffic > 7d → delete | 100% of hourly fee |
| Disable Flow Logs if unused | Logs to TOS incur storage + request cost | TOS storage fee |

## Cost Anomaly Detection

| Warning Sign | Check Command | Action |
|-------------|---------------|--------|
| 💰 Unexpected NAT Gateway creation | `ve nat DescribeNatGateways` | Verify with team, delete if unauth'd |
| ⚠️ Cross-AZ traffic surge | `ve cms GetMetricStatistics` — `Traffic:CrossAZ` | Co-locate services, add LB |
| 📊 Flow Logs cost spike | `ve billing DescribeBillDetail` — filter `TOS` | Reduce log sampling rate/retention |

## Query Current Pricing

VPC itself is free. Associated resource pricing via billing API:

```bash
ve billing DescribeBillSummaryByMonth --body '{"BillingPeriod":"{{env.BILLING_MONTH}}"}'
```

## Related Resources

- [FinOps — Billing Cost Optimization](../../ve-billing-ops/references/advanced/finops.md)
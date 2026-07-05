# FinOps — Function Compute Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7.

## Billing Model Comparison

| Dimension | Pay-per-use | Resource Plans | Savings |
|-----------|-------------|----------------|---------|
| Invocations | Per 1M calls | Pre-purchased | ~50-80% |
| Compute | Per GB-s | Pre-purchased | ~50-80% |
| Outbound | Per GB | Same | — |
| Provisioned | Per instance-hour | Reserved | ~60% |

## Cost Optimization Quick Reference

| Situation | Action | Savings |
|-----------|--------|---------|
| High invocation volume > 10M/month | Pre-purchase resource plans | 50-80% |
| High GB-s compute > 400K/month | Pre-purchase compute plan | 50-80% |
| Cold start on critical path | Add provisioned concurrency (2-5 instances) | Latency ↓ cost stable |
| Asynchronous workloads | Increase max retries to reduce recompute waste | 10-30% |
| Idle functions > 7d no invocation | Delete or reduce concurrency to 0 | Up to 100% |
| Large deployment packages > 50MB | Minimize dependencies + use layers | Deploy cost ↓ |

> ⚠️ Pricing data sourced from ve DescribePrice command — use `ve fg DescribePrice` for current quotes.

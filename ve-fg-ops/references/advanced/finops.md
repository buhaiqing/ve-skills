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
| High invocation volume | Use resource plans | 50-80% |
| Cold start latency | Provisioned concurrency | Latency ↓ cost ↑ |
| Long function runtime | Optimize code | Compute ↓ cost ↓ |
| Idle functions | Set concurrency limit | Up to 100% |
| Large deployment packages | Minimize dependencies | Deploy cost ↓ |

> Always query the FG pricing API — never hardcode rates.

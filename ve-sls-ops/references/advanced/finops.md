# FinOps — SLS Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7.

## Billing Model Comparison

| Dimension | Pay-per-use | Resource Plans |
|-----------|-------------|----------------|
| Write | Per GB ingested | Pre-purchased |
| Storage | Per GB/month | Tiered storage |
| Query | Per scan GB | Included quota |

## Cost Optimization Quick Reference

| Situation | Action | Savings |
|-----------|--------|---------|
| High ingestion | Enable compression | 30-50% |
| Long storage | Move to cold tier | 60-80% |
| Excessive query | Optimize SQL, use indexes | 50-90% |
| Unused projects | Delete stale projects | ~100% |
| Verbose logging | Filter low-value logs | 20-60% |

> Always query the SLS pricing API — never hardcode rates.

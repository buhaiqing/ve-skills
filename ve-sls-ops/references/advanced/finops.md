# FinOps — SLS Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7.

## Billing Model Comparison

| Dimension | Pay-per-use | Resource Plans |
|-----------|-------------|----------------|
| Write (ingestion) | Per GB ingested (tiered: lower rate >1TB/day) | Pre-purchased capacity |
| Storage | Per GB/month (tiered: hot vs cold) | Tiered storage pricing |
| Query (scan) | Per GB scanned per query | Included quota in plans |
| Shard | Per active shard per day | Included in plans |

## Cost Optimization

| Situation | Action | 💰 Savings |
|-----------|--------|---------|
| High daily ingestion >1TB | Enable client-side compression (LZ4/ZSTD) | 30-50% ingestion cost |
| Long retention >30d | Move to cold storage tier | 60-80% storage cost |
| Expensive SELECT * queries | Use indexed fields, limit projection | 50-90% scan cost |
| Unused log projects | Delete stale projects | ~100% reclaim |
| Verbose debug logging | Filter low-value DEBUG/TRACE logs at source | 20-60% |

## Cost Anomaly Detection

| Sign | Cause | Investigation |
|------|-------|-------------|
| ⚠️ Ingestion volume spike 3x+ | Log misconfiguration, loop logging | `ve sls DescribeProject --project {{output.ProjectName}}` |
| ⚠️ Storage cost rising with no retention change | Hot tier accumulating, no cold transition | Check logstore retention + storage tier config |
| ⚠️ Query cost > ingestion cost | Excessive ad-hoc full-scan queries | Review query patterns, add indexes |
| ⚠️ Shard cost growing | Too many shards for log volume | Consolidate shards to match throughput |

## Query Pricing

```bash
# SLS pricing is per-GB tiered by ingestion/storage — check project status
ve sls DescribeProject --project {{output.ProjectName}}
# Use Billing API for detailed cost breakdown
ve billing DescribeBillSummaryByMonth --body '{"BillingPeriod":"{{env.BILLING_MONTH}}"}'
```

## Related Resources

- [ve-billing-ops FinOps](../../ve-billing-ops/references/advanced/finops.md)
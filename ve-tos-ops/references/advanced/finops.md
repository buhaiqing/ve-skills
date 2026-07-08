# FinOps — TOS Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7.

## Billing Model

| Tier | Standard | Infrequent Access (IA) | Archive | Cold Archive |
|------|----------|------------------------|---------|--------------|
| Storage | Per GB/month | ~60% of Standard | ~20% of Standard | ~10% of Standard |
| Request | Per 10K requests | Per 10K + retrieval fee | Per 10K + retrieval fee | Per 10K + retrieval fee |
| Min bill | — | 30d min storage | 90d min storage | 180d min storage |

## Cost Optimization

| Situation | Action | 💰 Savings |
|-----------|--------|---------|
| Data accessed <1x/month | Lifecycle rule: Standard → IA | ~60% storage cost |
| Data not accessed >90d | Lifecycle rule: IA → Archive | ~80% on stored objects |
| Orphaned multipart uploads | Lifecycle rule: abort incomplete >7d | Recovers orphaned storage |
| Cross-region replication | Disable if not needed | 100% of replication cost |
| Public-read objects | Use pre-signed URL instead | Avoids bandwidth surcharge |

## Cost Anomaly Detection

| Sign | Cause | Investigation |
|------|-------|-------------|
| ⚠️ Storage spike >20% in 1d | Unplanned upload, data sync | `ve tos ListObjects --bucket {{output.BucketName}}` |
| ⚠️ Request count surge | Misconfigured app, DDoS | Check access logs, enable CDN |
| ⚠️ Cross-region replication cost | Replication rules enabled | `ve tos GetBucketReplication --bucket {{output.BucketName}}` |
| ⚠️ Unexpected storage tier cost | Lifecycle rule not applied | Verify lifecycle policy exists |

## Query Pricing

```bash
# TOS pricing is per-GB tiered by region — use Billing API for cost analysis
ve billing DescribeBillSummaryByMonth --body '{"BillingPeriod":"{{env.BILLING_MONTH}}"}'
# Check bucket location (pricing varies by region)
ve tos GetBucketLocation --bucket {{output.BucketName}}
```

## Related Resources

- [ve-billing-ops FinOps](../../ve-billing-ops/references/advanced/finops.md)
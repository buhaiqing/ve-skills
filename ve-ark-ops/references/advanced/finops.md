# FinOps — ARK Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7.

## Billing Model Comparison

| Component | Factor | Optimization |
|-----------|--------|-------------|
| Backup storage | Per GB/month (incremental) | Prune stale recovery points |
| Cross-region replication | Per GB transferred + stored | Disable if DR not required |
| Backup job | Per job execution | Consolidate backup windows |
| Snapshot (manual) | Per GB/month | Delete orphaned snapshots |

## Cost Optimization

| Situation | Action | 💰 Savings |
|-----------|--------|---------|
| Full backup every day | Switch to incremental + weekly full | 50-80% storage cost |
| Retention >90d | Reduce retention to 30d (compliance-permitting) | 60-70% |
| Cross-region backup | Disable DR copy for non-critical vaults | 100% of replication cost |
| Orphaned snapshots | List & delete snapshots of deleted instances | 100% reclaim |
| Overlapping backup windows | Stagger backup start time to avoid burst billing | Variable |

## Cost Anomaly Detection

| Sign | Cause | Investigation |
|------|-------|-------------|
| ⚠️ Backup storage doubling monthly | Full backup policy, no incremental | `ve ark DescribeVault --vault {{output.VaultName}}` |
| ⚠️ Cross-region cost > primary cost | Unintended replication rule | Check replication config on vault |
| ⚠️ Snapshot count >100 | No clean-up policy for manual snapshots | List snapshots, delete expired ones |
| ⚠️ Backup job failure cost | Retry loops generating extra job charges | Check job status in ARK console |

## Query Pricing

```bash
# ARK pricing is per-GB for backup storage — check vault details
ve ark DescribeVault --vault {{output.VaultName}}
# Use Billing API for cost analysis
ve billing DescribeBillSummaryByMonth --body '{"BillingPeriod":"{{env.BILLING_MONTH}}"}'
```

## Related Resources

- [ve-billing-ops FinOps](../../ve-billing-ops/references/advanced/finops.md)
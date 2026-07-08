# FinOps — NAS Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7.

## Billing Model Comparison

| Tier | Performance (SSD) | Capacity (HDD) | Infrequent Access (IA) |
|------|-------------------|----------------|------------------------|
| Storage | Per GB/month (high) | Per GB/month (medium) | Per GB/month (low) |
| Throughput | ~2× capacity tier | Baseline | Lower throughput |
| Min bill | — | — | 30d min storage |
| Best For | Database, high IOPS | Large files, backup | Cold data, archive |

## Cost Optimization

| Situation | Action | 💰 Savings |
|-----------|--------|---------|
| Low I/O files on Performance tier | Migrate to Capacity tier (lifecycle rule) | 40-60% storage cost |
| Files not accessed >30d | Lifecycle rule → IA tier | ~60% storage cost |
| Daily snapshot of stable data | Reduce snapshot frequency from hourly → daily | 80-90% snapshot cost |
| Orphaned file systems | Delete unused file systems | 100% reclaim |
| Over-provisioned throughput | Use auto-scaling or reserve lower throughput | Variable |

## Cost Anomaly Detection

| Sign | Cause | Investigation |
|------|-------|-------------|
| ⚠️ Storage usage spike | Unplanned large file write, data migration | `ve nas DescribeFileSystems --body '{}'` check capacity |
| ⚠️ Snapshot cost growing | Snapshot policy too aggressive | Review snapshot schedule |
| ⚠️ IA promotion cost | Files rapidly accessed after lifecycle to IA | Check access patterns before lifecycle rules |
| ⚠️ Throughput cost on idle FS | Reserved throughput not matching workload | Monitor actual throughput vs reserved |

## Query Pricing

```bash
# NAS pricing is per-GB tiered by capacity tier — check file system info
ve nas DescribeFileSystems --body '{"FileSystemId":"{{output.FileSystemId}}"}'
# Use Billing API for cost analysis
ve billing DescribeBillSummaryByMonth --body '{"BillingPeriod":"{{env.BILLING_MONTH}}"}'
```

## Related Resources

- [ve-billing-ops FinOps](../../ve-billing-ops/references/advanced/finops.md)
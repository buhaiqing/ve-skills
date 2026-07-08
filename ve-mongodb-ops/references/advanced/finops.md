# FinOps — MongoDB Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Billing Model Comparison

| Model | Discount | Best For |
|-------|----------|----------|
| PostPaid | 0% | Dev/test, variable workloads |
| PrePaid (1yr) | ~35% | Steady production |
| PrePaid (3yr) | ~50% | Long-term infra |

## Cost Optimization Tips

| Action | Condition | 💰 Savings |
|--------|-----------|---------|
| Optimal shard key design | Uneven data distribution → hot shard | 30-50% on cluster cost |
| Storage planning: pre-allocate vs. on-demand | Known data growth rate | ~20% via reserved storage |
| Backup window off-peak | Daily backup during high-write hours | Avoids IOPS throttling cost |
| Downsize config servers | Small-scale sharded cluster | $20-50/month per config replica |
| Convert PostPaid → PrePaid (1yr) | Running > 30d steady | ~35% |
| Reduce oplog size | Oplog retention > 24h sufficient | 10-20% on storage |

## Cost Anomaly Detection

| Warning Sign | Check | Action |
|-------------|-------|--------|
| 📈 Storage spike > 50% in 24h | `db.stats().dataSize` vs `storageSize` | Check for unbounded collections, missing TTL index |
| ⚡ IOPS burst on single shard | Shard key causing hot spot | Review shard key distribution, split chunk |
| 🔗 Connection count surge | `db.serverStatus().connections` > 80% | Add mongos router or connection pooling |
| 🔥 CPU > 80% on a shard | `currentOp()` slow queries on that shard | Add index or rebalance shards |
| 💣 Oplog window shrinking | `rs.printReplicationInfo()` shows < 2h | Increase oplog size or reduce write load |

## Query Current Prices

```bash
# Query MongoDB instance pricing
ve mongodb DescribeDBInstancePrice --body '{"InstanceId":"{{output.InstanceId}}"}'
# Query monthly billing summary
ve billing DescribeBillSummaryByMonth --body '{"BillingPeriod":"{{env.BILLING_MONTH}}"}'
```

## Related Resources

- [ve-billing-ops FinOps](../../../ve-billing-ops/references/advanced/finops.md) — General billing models, budget alerts & cost tags
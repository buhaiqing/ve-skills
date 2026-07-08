# FinOps — PolarDB MySQL Cost Optimization (Advanced)

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
| Compute-storage separation: scale compute only | Storage needs stable, QPS fluctuates | ~40% vs. bundled scaling |
| Warm storage tier for cold data | Data accessed < 1x/month | ~50% on archive storage |
| Read-only node for analytic queries | Heavy reporting on primary node | ~30% vs. primary spec upgrade |
| Auto-pause idle compute (Serverless) | Dev/test with intermittent usage | Up to 70% on compute |
| Convert PostPaid → PrePaid (1yr) | Running > 30d steady | ~35% |
| Reduce backup retention | Dev/test: keep 1d snapshots | 40-60% on backup |

## Cost Anomaly Detection

| Warning Sign | Check | Action |
|-------------|-------|--------|
| 📈 Storage spike > 50% in 24h | PolarDB storage consumption log | Check undo purge, temp tables, binlog |
| ⚡ IOPS burst on single node | Read-only node IOPS > primary 2× | Redistribute read workload to more ROs |
| 🔗 Connections hitting max | PolarDB cluster endpoints saturated | Add read-only endpoint, enable connection pooling |
| 🔥 Compute node CPU > 80% | PolarDB performance insight | Scale compute spec or add RO node |
| 💰 Storage auto-scaling cost | Storage > 80% of provisioned limit | Review data lifecycle, archive cold partitions |

## Query Current Prices

```bash
# Query PolarDB MySQL instance pricing
ve rds_mysql DescribeDBInstancePrice --body '{"DBInstanceId":"{{output.DBInstanceId}}"}'
# Query monthly billing summary
ve billing DescribeBillSummaryByMonth --body '{"BillingPeriod":"{{env.BILLING_MONTH}}"}'
```

## Related Resources

- [ve-billing-ops FinOps](../../../ve-billing-ops/references/advanced/finops.md) — General billing models, budget alerts & cost tags
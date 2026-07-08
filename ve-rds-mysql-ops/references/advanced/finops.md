# FinOps — RDS MySQL Cost Optimization (Advanced)

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
| Storage: ESSD PL0 for archive tables | Old data accessed < 1x/quarter | ~40% on cold storage |
| Read replica for read-heavy workload | Read/write ratio > 3:1, CPU < 70% | ~30% vs. spec upgrade |
| Slow query → index → right-size | avg CPU < 15% for 7d | ~50% instance cost |
| Backup: 7d retention for dev/test | Non-production | 30-60% backup cost |
| Convert PostPaid → PrePaid (1yr) | Running > 30d | ~35% |
| Drop unused Binlog files | Long retention with no PITR need | 100% of binlog storage |

## Cost Anomaly Detection

| Warning Sign | Check | Action |
|-------------|-------|--------|
| 📈 Storage spike > 50% in 24h | `show binary logs;` growth rate | Purge old binlogs, increase storage |
| ⚡ IOPS burst > baseline 3× | `show processlist;` full table scans | Add index, optimize SQL |
| 🔗 Threads_connected surge > 2× | App connection pool config | Set max_connections, use RDS Proxy |
| 🔥 CPU > 80% sustained | Slow query log, `performance_schema` | Right-size or add read replica |
| 💾 Undo tablespace growth | Long-running txns not committing | Tune `innodb_max_purge_lag` |

## Query Current Prices

```bash
# Query RDS MySQL instance pricing
ve rds_mysql DescribeDBInstancePrice --body '{"DBInstanceId":"{{output.DBInstanceId}}"}'
# Query monthly billing summary
ve billing DescribeBillSummaryByMonth --body '{"BillingPeriod":"{{env.BILLING_MONTH}}"}'
```

## Related Resources

- [ve-billing-ops FinOps](../../../ve-billing-ops/references/advanced/finops.md) — General billing models, budget alerts & cost tags
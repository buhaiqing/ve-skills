# FinOps — RDS PostgreSQL Cost Optimization (Advanced)

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
| Tune autovacuum to prevent bloat | Tables with > 20% dead tuples → `pg_bloat_check` | 30-50% on storage |
| Connection pooling (PgBouncer) | > 200 idle connections | ~20% on CPU/mem |
| Read replica for reporting workload | Analytical queries impacting write throughput | ~30% vs. scale-up |
| Archive old partitions | Partitioned tables with data > 90d old | 40-60% on cold storage |
| Convert PostPaid → PrePaid (1yr) | Running > 30d steady | ~35% |
| Reduce `wal_keep_segments` | Replica lag stable, no long replication delay | 10-20% WAL storage cost |

## Cost Anomaly Detection

| Warning Sign | Check | Action |
|-------------|-------|--------|
| 📈 Storage spike > 50% in 24h | `pg_stat_user_tables.n_dead_tup` growth | Tune autovacuum, `VACUUM` aggressive |
| ⚡ IOPS burst > baseline 3× | Check seq_page_read vs. idx_scan ratio | Add index, tune `shared_buffers` |
| 🔗 Idle-in-transaction connections > 100 | `pg_stat_activity.state = 'idle in transaction'` | Set `idle_in_transaction_session_timeout` |
| 🔥 CPU > 80% sustained | `pg_stat_statements` top queries | Optimize or add read replica |
| 🐌 WAL generation spike | `pg_wal_lsn_diff` rapid growth | Check bulk load patterns, tune `checkpoint_completion_target` |

## Query Current Prices

```bash
# Query RDS PostgreSQL instance pricing
ve rds_mysql DescribeDBInstancePrice --body '{"DBInstanceId":"{{output.DBInstanceId}}"}'
# Query monthly billing summary
ve billing DescribeBillSummaryByMonth --body '{"BillingPeriod":"{{env.BILLING_MONTH}}"}'
```

## Related Resources

- [ve-billing-ops FinOps](../../../ve-billing-ops/references/advanced/finops.md) — General billing models, budget alerts & cost tags
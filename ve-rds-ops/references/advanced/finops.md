# FinOps — RDS Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Billing Model Comparison

| Model | Discount | Best For |
|-------|----------|----------|
| PostPaid | 0% | Variable/testing workloads |
| PrePaid (1yr) | ~35% | Steady production |
| PrePaid (3yr) | ~50% | Long-term infra |

## Cost Optimization Tips

| Action | Condition | 💰 Savings |
|--------|-----------|---------|
| Storage tiering: ESSD PL0 for cold data | IOPS < 500, access < 1x/day | ~40% on storage |
| Read replica instead of vertical scale | Read-heavy, CPU < 60%, QPS spikes | ~30% vs. scale-up |
| Backup retention: 7d for dev/test | Non-production DBs | 30-60% on backup cost |
| Slow query → right-size down | avg CPU < 15% for 7d | ~50% |
| Convert PostPaid → PrePaid (1yr) | Running > 30d steady | ~35% |
| Release idle read replicas | Standby used < 10% capacity | 100% |

## Cost Anomaly Detection

| Warning Sign | Check | Action |
|-------------|-------|--------|
| 📈 Storage spike > 50% in 24h | Binlog/undo growth rate | Enable auto-purging, increase storage |
| ⚡ IOPS burst > baseline 3× | Slow query log, full scans | Optimize queries, add read replica |
| 🔗 Connection surge > 2× normal | App-side connection leak | Set max_connections, enable RDS Proxy |
| 🔥 CPU > 80% sustained | Slow query log analysis | Right-size or scale out read replicas |

## Query Current Prices

```bash
# Query RDS instance pricing
ve rds_mysql DescribeDBInstancePrice --body '{"DBInstanceId":"{{output.DBInstanceId}}"}'
# Query monthly billing summary
ve billing DescribeBillSummaryByMonth --body '{"BillingPeriod":"{{env.BILLING_MONTH}}"}'
```

## Related Resources

- [ve-billing-ops FinOps](../../../ve-billing-ops/references/advanced/finops.md) — General billing models, budget alerts & cost tags
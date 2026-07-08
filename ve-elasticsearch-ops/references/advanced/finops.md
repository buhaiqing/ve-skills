# FinOps — Elasticsearch Cost Optimization (Advanced)

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
| Shard count optimization | Shard > 50 GB or < 5 GB per node | Merge small indices → 20-30% on storage |
| Cold/warm node tiering | Data > 30d old queried infrequently | Move to warm nodes → ~40% vs. hot tier |
| Index Lifecycle Management (ILM) | Time-series data (logs, metrics) | Auto-rollover + cold → 50-70% total cost |
| Reduce replica count after indexing | 2+ replicas during hot → reduce to 1 | ~33% storage cost |
| Convert PostPaid → PrePaid (1yr) | Running > 30d steady | ~35% |
| Delete unused indices | `_cat/indices` shows no queries in 90d | 100% of that index storage |

## Cost Anomaly Detection

| Warning Sign | Check | Action |
|-------------|-------|--------|
| 📈 Storage spike > 50% in 24h | `_cat/allocation` → shard distribution | Check for log index flood, ILM not applied |
| ⚡ Search latency > 5× normal | `_cat/nodes` → search thread pool queue | Add warm nodes, optimize shard count |
| 🔗 Cluster health yellow/red | `_cluster/health` → unassigned shards | Add nodes or adjust `index.routing.allocation` |
| 🔥 CPU > 80% on data nodes | `_nodes/hot_threads` → merge/query threads | Merge segments, reduce shard count |
| 💣 Heap usage > 85% | `_nodes/stats/jvm` → young/old GC ratio | Increase nodes or reduce field data cache |

## Query Current Prices

```bash
# Query Elasticsearch instance pricing
ve elasticsearch DescribeInstancePrice --body '{"InstanceId":"{{output.InstanceId}}"}'
# Query monthly billing summary
ve billing DescribeBillSummaryByMonth --body '{"BillingPeriod":"{{env.BILLING_MONTH}}"}'
```

## Related Resources

- [ve-billing-ops FinOps](../../../ve-billing-ops/references/advanced/finops.md) — General billing models, budget alerts & cost tags
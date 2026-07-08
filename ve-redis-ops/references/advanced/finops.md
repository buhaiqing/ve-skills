# FinOps — Redis Cost Optimization (Advanced)

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
| Fix memory fragmentation | `used_memory_rss / used_memory` > 1.5 | Restart or `MEMORY PURGE` → 20-40% mem |
| Tune eviction policy | LRU (allkeys-lru) vs. noeviction → avoid OOM | Prevents forced scaling, saves 100% of scale cost |
| Read replicas for cache-heavy | Read/write > 5:1, cache hit ratio > 90% | 2 replicas → ~30% vs. 1 large instance |
| Downsize for CPU-low instances | CPU < 20% for 7d, memory usage < 60% | ~50% |
| Convert PostPaid → PrePaid (1yr) | Running > 30d steady | ~35% |
| Delete unused keys | `INFO keyspace` shows many expired keys | 10-30% on memory usage |

## Cost Anomaly Detection

| Warning Sign | Check | Action |
|-------------|-------|--------|
| 📈 Memory spike > 50% in 1h | `MEMORY STATS` → keys count + TTL | Evict stale keys, add `maxmemory-policy` |
| ⚡ CPU burst > baseline 3× | `INFO COMMANDSTATS` → slow commands (`KEYS *`, `SMEMBERS large`) | Replace with `SCAN`, `SSCAN` |
| 🔗 Connection explosion > 10000 | Client list check → idle connections | Set `timeout` config, pool management |
| 💥 Fragmentation ratio > 2.0 | `used_memory_rss / used_memory` | Schedule `MEMORY PURGE` or restart |
| 📉 Cache hit ratio drop < 80% | `keyspace_misses / (hits + misses)` | Review TTL, pre-warm cache strategy |

## Query Current Prices

```bash
# Query Redis instance pricing
ve redis DescribePrice --body '{"RegionId":"{{env.REGION}}","InstanceClass":"redis.shard.xlarge"}'
# Query monthly billing summary
ve billing DescribeBillSummaryByMonth --body '{"BillingPeriod":"{{env.BILLING_MONTH}}"}'
```

## Related Resources

- [ve-billing-ops FinOps](../../../ve-billing-ops/references/advanced/finops.md) — General billing models, budget alerts & cost tags
# AIOps — Elasticsearch Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[Elasticsearch Alarm Triggered]
    │
    ├── Is it cluster health-related?
    │   ├── Cluster status RED → Check unassigned shards
    │   │   ├── Disk usage > 95% → Expand storage or delete old indices
    │   │   └── Node down > 1 node → Restart or replace node
    │   ├── Cluster status YELLOW → Check replica allocation
    │   │   └── Insufficient nodes → Add data nodes
    │   ├── Node CPU > 80% for > 5 min → Check heavy queries
    │   │   └── Optimize query patterns or add nodes
    │   └── BulkRejection > 10/min → Check indexing throughput
    │       └── Increase bulk queue size or add nodes
    │
    ├── Is it performance-related?
    │   ├── SearchLatency p99 > 1s → Review query patterns
    │   │   ├── Unoptimized wildcard/regex → Rewrite queries
    │   │   └── Shards per node > 1000 → Rebalance shards
    │   ├── IndexingRate dropping > 30% → Check disk IOPS
    │   │   └── Delegate to ve-ecs-ops for disk performance
    │   ├── JVMHeapUsage > 85% → Check GC pressure
    │   │   └── Reduce field data cache or scale up
    │   └── CircuitBreakerTripped > 0 in 5 min → Reduce request load
    │       └── Check for expensive queries and cancel long-running tasks
    │
    ├── Is it capacity-related?
    │   ├── DiskUsage > 80% on any node → Set up ILM policy
    │   ├── ShardCount per node > 800 → Merge small indices
    │   └── SnapshotFailure > 2 consecutive → Check repo connectivity
    │       └── Delegate to ve-tos-ops for backup repository
    │
    └── Unknown pattern → Delegate to ve-cms-ops for correlation analysis
```

## Alarm Storm Handling

**Detection Criteria:**
- Cluster status change impacting > 3 nodes simultaneously
- > 10 shard-related alarms within 5 minutes

**Suppression Workflow:**
1. Correlate by cluster and time window
2. Identify root cause (node failure vs capacity vs query load)
3. Group related shard/node alarms
4. Address root cause → verify cluster health recovers

## Proactive Inspection Checklist

```markdown
## Elasticsearch Proactive Inspection — [Date]

### Cluster Health
- [ ] Cluster status GREEN across all nodes
- [ ] No unassigned shards
- [ ] All nodes responding to health checks

### Performance
- [ ] Search latency p99 < 500ms
- [ ] Indexing latency p99 < 200ms
- [ ] JVM heap usage < 75% on all nodes
- [ ] GC frequency < 5 collections per minute

### Capacity
- [ ] Disk usage < 75% on all nodes
- [ ] Shards per node < 800
- [ ] ILM policy configured for all time-series indices

### Data Safety
- [ ] Daily snapshots succeeding
- [ ] Replica count ≥ 1 for production indices
- [ ] Force merge configured for read-only indices
```

## Multi-Round Diagnosis Review

Before finalizing any Elasticsearch diagnosis:

1. **Fact Check:** Are the cluster metrics current? Is the node count correct?
2. **Causal Analysis:** Is the issue caused by query load, capacity, or software bug?
3. **Solution Validation:** Will the fix improve cluster health without data loss?
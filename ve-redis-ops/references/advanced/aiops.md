# AIOps — Redis Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[Redis Alarm Triggered]
    │
    ├── Is it memory-related?
    │   ├── Memory usage > 80% (used_memory > maxmemory * 0.8) → Check key eviction policy
    │   │   ├── Keys evicted > 1000/min → Eviction policy too aggressive
    │   │   ├── Keyspace hit rate < 90% → Large keyspace not fitting in memory
    │   │   └── Use OBJECT FREQ to find hot keys
    │   ├── Memory fragmentation > 1.5 (mem_fragmentation_ratio) → Check client connections
    │   │   ├── Large number of short-lived keys → Tune maxmemory-samples
    │   │   └── Run MEMORY PURGE or restart
    │   └── OOM (used_memory > maxmemory + maxmemory*0.1) → Check maxmemory setting
    │       └── Adjust maxmemory or add replicas
    │
    ├── Is it connection-related?
    │   ├── Max clients reached (connected_clients > maxclients * 0.9) → Check client list
    │   │   ├── Idle clients > 100 → Kill with CLIENT KILL TYPE idle
    │   │   └── Kill idle clients or increase maxclients
    │   ├── High latency > 10ms avg → Check slowlog
    │   │   ├── Slowlog > 100 entries in 5min → Blocking commands (KEYS, SORT, etc.)
    │   │   └── Analyze slow commands
    │   ├── Replication lag > 5s (master_repl_offset - slave_repl_offset) → Check replica status
    │   │   ├── Replication buffer > 256MB → RDB snapshot on master
    │   │   └── Increase network bandwidth or optimize commands
    │   └── Input/output buffer > 1GB → Large pipelining or big keys
    │       └── Split big keys or reduce batch size
    │
    ├── Is it availability-related?
    │   ├── Sentinel failover (sentinel failover count > 3/day) → Check sentinel log
    │   │   ├── Quorum not reached → Network partition suspected
    │   │   └── Verify network stability
    │   ├── Cluster rebalancing (slot migration > 1000 slots/h) → Check slot distribution
    │   │   ├── Hot slots causing rebalance → Redistribute hash slots
    │   │   └── Verify data integrity after resharding
    │   ├── Node down (cluster known_nodes < expected) → Check node health
    │   │   ├── PING timeout > 30s → Node unreachable
    │   │   └── Add new node or repair existing
    │   └── Persistence failure (RDB/AOF last save > 5min) → Check disk I/O
    │       └── AOF rewrite stuck → Disable AOF rewrite or increase disk
    │
    └── Unknown → Delegate to ve-cms-ops
```

## Alarm Storm Handling

**Detection Criteria:**
- > 5 memory/connection alarms within 5 minutes
- > 20% keys evicted within 1 hour
- Cascade: OOM → service unavailable → downstream failures

**Suppression Workflow:**
1. Correlate by instance and metric type
2. Identify root cause (memory vs connection vs availability)
3. Apply targeted fix (evict keys, kill connections, restart)
4. Verify service recovers
5. Monitor for recurrence

## Proactive Inspection Checklist

```markdown
## Redis Proactive Inspection — [Date]

### Performance
- [ ] Memory usage < 70%
- [ ] Memory fragmentation ratio < 1.5
- [ ] No OOM events in past 7 days
- [ ] Slowlog length < 100

### Connections
- [ ] Connected clients < 80% of maxclients
- [ ] No zombie connections
- [ ] Replication lag < 1 second

### Reliability
- [ ] Sentinel/Cluster health check passing
- [ ] All slots covered in cluster mode
- [ ] Backups configured and tested

### Security
- [ ] AUTH enabled
- [ ] No 0.0.0.0:6379 binding
- [ ] TLS enabled for prod instances
```

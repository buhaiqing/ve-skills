# AIOps — Redis Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[Redis Alarm Triggered]
    │
    ├── Is it memory-related?
    │   ├── Memory usage > 80% → Check key eviction policy
    │   │   └── Use OBJECT FREQ to find hot keys
    │   ├── Memory fragmentation > 1.5 → Check client connections
    │   │   └── Run MEMORY PURGE or restart
    │   └── OOM → Check maxmemory setting
    │       └── Adjust maxmemory or add replicas
    │
    ├── Is it connection-related?
    │   ├── Max clients reached → Check client list
    │   │   └── Kill idle clients or increase maxclients
    │   ├── High latency → Check slowlog
    │   │   └── Analyze slow commands
    │   └── Replication lag → Check replica status
    │       └── Increase network bandwidth or optimize commands
    │
    ├── Is it availability-related?
    │   ├── Sentinel failover → Check sentinel log
    │   │   └── Verify network stability
    │   ├── Cluster rebalancing → Check slot distribution
    │   │   └── Verify data integrity after resharding
    │   └── Node down → Check node health
    │       └── Add new node or repair existing
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

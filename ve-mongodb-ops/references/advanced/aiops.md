# AIOps — MongoDB Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[MongoDB Alarm Triggered]
    │
    ├── Is it connection-related?
    │   ├── Too many connections → Check connection pool
    │   │   └── Optimize application connection reuse
    │   ├── Connection timeout → Check network path
    │   │   └── Delegate to ve-vpc-ops if network issue
    │   ├── Auth failure → Check credentials
    │   │   └── Verify IAM policy or key rotation
    │   ├── CurrentActive connections > 70% of max → Scale connection pool
    │   │   └── Increase maxConns or optimize app pooling
    │   └── Connection storms > 200 connections/min → Check app reconnect behavior
    │       └── Implement exponential backoff in app
    │
    ├── Is it storage-related?
    │   ├── Disk usage > 85% → Check data size
    │   │   └── Run cleanup or expand storage
    │   ├── Slow queries (execTime > 100ms) → Check explain plan
    │   │   └── Add missing indexes
    │   ├── Lock contention → Check currentOp
    │   │   └── Kill long-running ops or optimize
    │   ├── WiredTiger cache dirty ratio > 20% → Check write pressure
    │   │   └── Increase cache size or optimize write pattern
    │   └── Oplog size > 10% of disk → Review oplog retention
    │       └── Adjust oplog size or speed up secondaries
    │
    ├── Is it replication-related?
    │   ├── Replication lag > 10s → Check secondary status
    │   │   └── Increase oplog size or optimize writes
    │   ├── Primary election → Check network stability
    │   │   └── Verify connectivity between replicas
    │   ├── Rollback detected → Check oplog
    │   │   └── Manually resolve rollback
    │   ├── Heartbeat latency > 2s → Network congestion between replicas
    │   │   └── Delegate to ve-vpc-ops for network diagnosis
    │   └── Oplog window < 2 hours → Secondary falling too far behind
    │       └── Increase oplog size or improve secondary performance
    │
    └── Unknown → Delegate to ve-cms-ops
```

## Alarm Storm Handling

**Detection Criteria:**
- > 5 alarms within 5 minutes for same instance
- > 3 concurrent performance issues
- Replication lag > 10 seconds

**Suppression Workflow:**
1. Correlate by instance and operation type
2. Identify resource bottleneck
3. Apply fix (index, cleanup, scale)
4. Verify performance recovers
5. Monitor for recurrence

## Proactive Inspection Checklist

```markdown
## MongoDB Proactive Inspection — [Date]

### Performance
- [ ] Disk usage < 80%
- [ ] Connection count < 80% of limit
- [ ] No slow queries (> 100ms)
- [ ] Index coverage > 90%

### Replication
- [ ] Replication lag < 10 seconds
- [ ] No rollback events
- [ ] All secondaries reachable

### Security
- [ ] Auth enabled
- [ ] TLS enabled
- [ ] No publicly accessible instances

### Backup
- [ ] Oplog configured
- [ ] Point-in-time recovery tested
- [ ] Backup retention policy set
```

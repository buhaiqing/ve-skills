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
    │   └── Auth failure → Check credentials
    │       └── Verify IAM policy or key rotation
    │
    ├── Is it storage-related?
    │   ├── Disk usage > 85% → Check data size
    │   │   └── Run cleanup or expand storage
    │   ├── Slow queries → Check explain plan
    │   │   └── Add missing indexes
    │   └── Lock contention → Check currentOp
    │       └── Kill long-running ops or optimize
    │
    ├── Is it replication-related?
    │   ├── Replication lag → Check secondary status
    │   │   └── Increase oplog size or optimize writes
    │   ├── Primary election → Check network stability
    │   │   └── Verify connectivity between replicas
    │   └── Rollback detected → Check oplog
    │       └── Manually resolve rollback
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

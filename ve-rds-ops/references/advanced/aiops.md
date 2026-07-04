# AIOps — RDS MySQL variant Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[RDS MySQL variant Alarm Triggered]
    │
    ├── Is it CPU/Memory-related?
    │   ├── Yes → Check connection pool exhaustion
    │   │   └── Delegate to application team for connection pool tuning
    │   └── No → Continue...
    │
    ├── Is it storage/disk-related?
    │   ├── Disk usage > 80% → Check data growth rate, cleanup old data
    │   │   └── Run cleanup queries → verify freespace
    │   └── Slow queries → Check query execution plan
    │       └── Use EXPLAIN to identify full table scans
    │
    ├── Is it availability-related?
    │   ├── Instance unreachable → Check network path
    │   │   └── Delegate to ve-vpc-ops for VPC/ENI diagnosis
    │   ├── Replication lag > threshold → Check replica status
    │   │   └── Verify network bandwidth between primary and replicas
    │   └── Failover triggered → Check instance health
    │       └── Review recent operations that may have caused failover
    │
    └── Unknown pattern → Delegate to ve-cms-ops for correlation analysis
```

## Alarm Storm Handling

**Detection Criteria:**
- > 5 alarms within 5 minutes from same RDS MySQL variant instance
- > 3 concurrent performance degradation alarms

**Suppression Workflow:**
1. Correlate alarms by instance ID and metric category
2. Identify root alarm (earliest or highest severity)
3. Group related alarms under root cause
4. Execute diagnosis on primary alarm
5. Verify all related alarms clear after resolution

## Proactive Inspection Checklist

```markdown
## RDS MySQL variant Proactive Inspection — [Date]

### Resource Health
- [ ] CPU usage < 70% across all instances
- [ ] Memory usage < 85% across all instances
- [ ] Disk usage < 80% with > 20% free space
- [ ] Connection count < 80% of max connections

### Performance
- [ ] Slow query count < 100
- [ ] No queries running > 30 seconds
- [ ] Replication lag < 60 seconds

### Security
- [ ] SSL/TLS enabled for all connections
- [ ] No overly permissive access rules
- [ ] Backup retention policy configured

### Reliability
- [ ] Multi-AZ deployment for production
- [ ] Automated backups configured
- [ ] Point-in-time recovery tested
```

## Multi-Round Diagnosis Review

Before finalizing any RDS MySQL variant diagnosis:
1. **Fact Check:** Are the metrics and instance status current?
2. **Causal Analysis:** Is the identified cause the true root cause?
3. **Solution Validation:** Will the fix actually resolve the issue?

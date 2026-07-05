# AIOps — RDS MySQL variant Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[RDS MySQL variant Alarm Triggered]
    │
    ├── Is it CPU/Memory-related?
    │   ├── CPU > 85% sustained 5min → Check connection pool exhaustion
    │   │   ├── Threads_running > max_connections * 0.8 → Connection storm
    │   │   ├── CPU > 85% + Memory > 80% → Buffer pool thrashing or full table scans
    │   │   └── Delegate to application team for connection pool tuning
    │   ├── Memory > 85% → Check InnoDB buffer pool hit rate < 95%
    │   │   └── Increase innodb_buffer_pool_size or optimize queries
    │   └── No → Continue...
    │
    ├── Is it storage/disk-related?
    │   ├── Disk usage > 85% → Check data growth rate, binlog retention
    │   │   ├── Binlog retention > 7 days → Reduce binlog expiry
    │   │   ├── Undo tablespace growth > 10GB/week → Check long-running transactions
    │   │   └── Run cleanup queries → verify freespace
    │   ├── Slow queries (> 100/min) → Check query execution plan
    │   │   ├── Full table scans detected (rows_examined >> rows_sent) → Missing index
    │   │   └── Use EXPLAIN to identify full table scans
    │   ├── IOPS > 90% of max throughput → Check checkpoint frequency
    │   │   └── Increase innodb_io_capacity or optimize write-heavy queries
    │   └── Temp table spill to disk > 10/min → Increase tmp_table_size
    │
    ├── Is it availability-related?
    │   ├── Instance unreachable → Check network path
    │   │   ├── Connection refused → Instance in restart or failover
    │   │   └── Delegate to ve-vpc-ops for VPC/ENI diagnosis
    │   ├── Replication lag > 30s → Check replica status
    │   │   ├── Seconds_Behind_Master > 300s → Heavy write load on primary
    │   │   ├── Slave_IO_Running = No → Network or binlog corruption
    │   │   └── Verify network bandwidth between primary and replicas
    │   ├── Failover triggered → Check instance health
    │   │   ├── Failovers > 2 in 24h → Instance instability or AZ issue
    │   │   └── Review recent operations that may have caused failover
    │   └── Connection count > max_connections * 0.85 → Connection pool exhausted
    │       └── Kill idle connections or increase max_connections
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

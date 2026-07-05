# AIOps — RDS PostgreSQL Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[RDS PostgreSQL Alarm Triggered]
    │
    ├── Is it CPU/Memory-related?
    │   ├── CPU > 85% sustained 5min → Check active connections
    │   │   ├── High context switch → Too many connections (check_pg_activity)
    │   │   ├── Autovacuum worker CPU > 30% → Tune autovacuum parameters
    │   │   └── Delegate to application team for connection pool tuning
    │   ├── Memory > 85% → Check shared_buffers vs effective_cache_size
    │   │   └── Increase shared_buffers or reduce work_mem per connection
    │   └── No → Continue...
    │
    ├── Is it storage/disk-related?
    │   ├── Disk usage > 85% → Check WAL retention, dead tuples
    │   │   ├── WAL archive lag > 5min → Archive destination full
    │   │   ├── Table bloat ratio > 20% → Run VACUUM FULL or pg_repack
    │   │   └── Run cleanup queries → verify freespace
    │   ├── Slow queries (> 100/min) → Check query execution plan
    │   │   ├── Seq scans on large tables > index scans → Missing index
    │   │   └── Use EXPLAIN to identify full table scans
    │   ├── IOPS > 90% of max throughput → Check checkpoint frequency
    │   │   └── Increase checkpoint_timeout or checkpoint_completion_target
    │   └── Temp file creation > 1GB/min → Increase work_mem or sort_mem
    │
    ├── Is it availability-related?
    │   ├── Instance unreachable → Check network path
    │   │   ├── Connection refused → Instance restart or streaming replication conflict
    │   │   └── Delegate to ve-vpc-ops for VPC/ENI diagnosis
    │   ├── Replication lag > 60s → Check replica status
    │   │   ├── pg_wal_archive > 5min behind → WAL shipping bottleneck
    │   │   ├── sync_priority = 0 and sync_state = 'sync' mismatch → Configuration drift
    │   │   └── Verify network bandwidth between primary and replicas
    │   ├── Failover triggered → Check instance health
    │   │   ├── Failovers > 2 in 24h → HA configuration issue or AZ instability
    │   │   └── Review recent operations that may have caused failover
    │   └── Connection count > max_connections * 0.85 → Connection pool exhausted
    │       └── Kill idle_in_transaction connections or increase max_connections
    │
    └── Unknown pattern → Delegate to ve-cms-ops for correlation analysis
```

## Alarm Storm Handling

**Detection Criteria:**
- > 5 alarms within 5 minutes from same RDS PostgreSQL instance
- > 3 concurrent performance degradation alarms

**Suppression Workflow:**
1. Correlate alarms by instance ID and metric category
2. Identify root alarm (earliest or highest severity)
3. Group related alarms under root cause
4. Execute diagnosis on primary alarm
5. Verify all related alarms clear after resolution

## Proactive Inspection Checklist

```markdown
## RDS PostgreSQL Proactive Inspection — [Date]

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

Before finalizing any RDS PostgreSQL diagnosis:
1. **Fact Check:** Are the metrics and instance status current?
2. **Causal Analysis:** Is the identified cause the true root cause?
3. **Solution Validation:** Will the fix actually resolve the issue?

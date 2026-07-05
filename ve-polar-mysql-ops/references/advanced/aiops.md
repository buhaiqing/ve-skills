# AIOps — PolarDB MySQL Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[PolarDB MySQL Alarm Triggered]
    │
    ├── Is it CPU/Memory-related?
    │   ├── CPU > 85% sustained 5min → Check connection pool exhaustion
    │   │   ├── Parallel query threads > 80% → Heavy OLAP workload on OLTP node
    │   │   ├── CPU > 85% + Memory > 80% → Buffer pool thrashing
    │   │   └── Delegate to application team for connection pool tuning
    │   ├── Memory > 85% → Check PolarDB buffer pool usage
    │   │   └── Increase loose_polar_buffer_pool_size or migrate to larger spec
    │   └── No → Continue...
    │
    ├── Is it storage/disk-related?
    │   ├── Disk usage > 85% → Check PolarStore capacity, log retention
    │   │   ├── Redo log lag > 10GB on PolarStore → Write-heavy workload
    │   │   ├── Binlog retention > 7 days → Reduce binlog_expire_logs_seconds
    │   │   └── Run cleanup queries → verify freespace
    │   ├── Slow queries (> 100/min) → Check query execution plan
    │   │   ├── Full table scans detected (rows_examined >> rows_sent) → Missing index
    │   │   ├── Using filesort on large result sets → Optimize ORDER BY with indexes
    │   │   └── Use EXPLAIN to identify full table scans
    │   ├── IOPS > 90% of max throughput → Check PolarDB IO throttle
    │   │   └── Increase polar_io_capacity or scale up instance class
    │   └── Temp table on disk > 50MB/query → Increase tmp_table_size or max_heap_table_size
    │
    ├── Is it availability-related?
    │   ├── Instance unreachable → Check network path
    │   │   ├── PolarProxy connection refused → Proxy node restart
    │   │   └── Delegate to ve-vpc-ops for VPC/ENI diagnosis
    │   ├── Replication lag > 30s → Check PolarDB read-only node status
    │   │   ├── PolarStore IO bottleneck → Parallel queries slowing redo apply
    │   │   └── Verify network bandwidth between compute nodes
    │   ├── Failover triggered → Check instance health
    │   │   ├── Failovers > 2 in 24h → PolarStore or compute node instability
    │   │   └── Review recent operations that may have caused failover
    │   └── Connection count > max_connections * 0.85 → Connection pool exhausted
    │       └── Kill idle connections or increase max_connections via PolarProxy
    │
    └── Unknown pattern → Delegate to ve-cms-ops for correlation analysis
```

## Alarm Storm Handling

**Detection Criteria:**
- > 5 alarms within 5 minutes from same PolarDB MySQL instance
- > 3 concurrent performance degradation alarms

**Suppression Workflow:**
1. Correlate alarms by instance ID and metric category
2. Identify root alarm (earliest or highest severity)
3. Group related alarms under root cause
4. Execute diagnosis on primary alarm
5. Verify all related alarms clear after resolution

## Proactive Inspection Checklist

```markdown
## PolarDB MySQL Proactive Inspection — [Date]

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

Before finalizing any PolarDB MySQL diagnosis:
1. **Fact Check:** Are the metrics and instance status current?
2. **Causal Analysis:** Is the identified cause the true root cause?
3. **Solution Validation:** Will the fix actually resolve the issue?

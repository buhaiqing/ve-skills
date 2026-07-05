# AIOps — RDS MySQL Intelligent Operations

> Deep AIOps content per TE-7. `monitoring.md` links here.

## Cross-Skill Diagnosis Decision Tree

```
[Alarm: RDS MySQL issue]
    │
    ├── Is it query-performance related?
    │   ├── SlowQueries > 100/min → Check execution plan → Application query issue
    │   │   └── Rows_examined >> rows_sent → Missing index
    │   ├── Full table scan detected → Index missing or stats stale
    │   │   └── Run ANALYZE TABLE or add appropriate index
    │   ├── Lock waits (Innodb_row_lock_current_waits > 100) → Check transaction list
    │   │   ├── Long-running transactions > 60s → Kill or rollback
    │   │   └── ve-rds-mysql-ops (SHOW ENGINE INNODB STATUS equivalent)
    │   └── Temp tables created on disk > 50/min → Increase tmp_table_size or optimize queries
    │
    ├── Is it resource-exhaustion?
    │   ├── CPU > 90% + Connections high (Threads_running > max_connections*0.8) → ve-ecs-ops (app server connection leak)
    │   ├── Disk > 85% → Binlog growth? → ve-cms-ops cleanup policy
    │   │   └── Binlog retention > 7 days → Reduce binlog_expire_logs_seconds
    │   ├── Memory > 85% → Buffer pool hit rate < 95%? → Rebalance innodb_buffer_pool_size
    │   ├── IOPS > 90% → Check for heavy read/write pattern
    │   │   └── Checkpoint age > 75% of total log → Increase innodb_io_capacity
    │   └── Aborted_connects > 10/min → Connection auth failures or network issues
    │
    ├── Is it replication-related?
    │   ├── RO lag > 30s → Check heavy writes on primary → Throttle writes
    │   │   └── Seconds_Behind_Master > 300s → Parallel replication bottleneck
    │   ├── Slave_IO_Running = No → Network or binlog corruption
    │   │   └── Check relay log corruption → Reset slave
    │   └── RO node unresponsive → ve-cms-ops (network or node failover)
    │
    └── Unknown/cluster-wide → ve-cms-ops for correlation + ve-vpc-ops for network
```

## Delegation Matrix

| Alarm Signal | Primary Skill | Secondary Skill | Fallback |
|-------------|--------------|-----------------|----------|
| Slow query surge | Application team | ve-rds-mysql-ops (explain plan) | ve-das-ops |
| Connection pool full | ve-rds-mysql-ops | ve-ecs-ops (app server) | ve-cms-ops |
| Disk usage > 90% | ve-rds-mysql-ops | ve-cms-ops (binlog management) | Manual |
| Replication lag > 30s | ve-rds-mysql-ops | ve-ecs-ops (primary CPU) | ve-cms-ops |
| Network bandwidth cap | ve-vpc-ops | ve-rds-mysql-ops | Manual |

## Cross-Skill Routing

| Symptom | Delegate To |
|---------|------------|
| Host-level resource exhaustion (CPU/memory) | ve-ecs-ops |
| IAM permission denied for backup/restore | ve-iam-ops |
| Network connectivity to application layer | ve-vpc-ops |
| Alarm rule suppression or threshold tuning | ve-cms-ops |

## Proactive Inspection Checklist

### Resource Health
- [ ] CPU avg < 70%, peak < 85% over 24h
- [ ] Memory usage < 80% (buffer pool efficient)
- [ ] Disk usage < 80% with > 50GB free
- [ ] Max connections < 80% of max_connections
- [ ] Slow queries < 10/min
- [ ] Replication lag < 5s on all RO nodes
- [ ] Buffer pool hit rate > 97%

### Security
- [ ] IP whitelist restricted (not 0.0.0.0/0 without need)
- [ ] No superuser accounts used by application
- [ ] Backup encryption enabled
- [ ] SSL/TLS for all connections where supported

### Cost
- [ ] No instances with CPU < 5% for 14 days
- [ ] No instances with storage < 20% used for 14 days
- [ ] RO nodes with connections < 10% → consider removing
- [ ] Backup retention period aligned with compliance (not over-retained)

## Alarm Storm Handling

**Detection:** > 10 table-level or connection-level alarms within 5 min, or slow query alarms cascading across tables.

**Suppression:**
1. Correlate by instance + time window
2. Root alarm = earliest slow query spike or connection limit reached
3. Group: all table-level slow query alarms → single "query performance degraded"
4. Throttle: 1 notification per root cause

## Multi-Round Diagnosis Review

1. **Fact Check**: Metrics from correct instance? Binlog growth verified?
2. **Causal Analysis**: Is slow query the root cause or symptom?
3. **Solution Validation**: Will index creation require table lock during prod hours?
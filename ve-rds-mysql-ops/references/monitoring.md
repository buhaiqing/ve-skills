# RDS MySQL Multi-Metric Correlation — AIOps

## Anomaly Patterns

| Pattern | Metrics | Detection Logic | Severity |
|---------|---------|-----------------|----------|
| Slow Query Cascade | SlowQueries > 50/min + AvgQueryTime > 5s + CPU > 85% | All AND over 10min | Critical |
| Connection Pool Exhaustion | Connections > 90% + ConnectionWait > 2s + ActiveConnections = max_connections | Sustained 5min | Critical |
| Disk Write Pressure | Disk usage > 85% + IOPS write > 90% limit + Binlog growth > 5GB/hr | AND sustained 15min | Critical |
| Replication Lag Spike | RO lag > 30s + Relay log growth > 1GB + Write QPS > 5000 | Sustained 5min | Warning |
| Buffer Pool Inefficiency | Hit rate < 95% + Pages read/s > threshold + Temp tables > 100/s | AND over 15min | Warning |
| Lock Contention | Lock wait > 30s + Lock waits/s > 100 + Deadlocks > 5/min | AND over 5min | Critical |

## Cross-Skill Diagnosis Decision Tree

```
[Alarm: RDS MySQL issue]
    │
    ├── Is it query-performance related?
    │   ├── SlowQueries spike → Check execution plan → Application query issue
    │   ├── Full table scan detected → Index missing or stats stale
    │   └── Lock waits → ve-rds-mysql-ops (SHOW ENGINE INNODB STATUS equivalent)
    │
    ├── Is it resource-exhaustion?
    │   ├── CPU > 90% + Connections high → ve-ecs-ops (app server connection leak)
    │   ├── Disk > 85% → Binlog growth? → ve-cms-ops cleanup policy
    │   ├── Memory > 85% → Buffer pool too large? → Rebalance
    │   └── IOPS > 90% → Check for heavy read/write pattern
    │
    ├── Is it replication-related?
    │   ├── RO lag > 30s → Check heavy writes on primary → Throttle writes
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

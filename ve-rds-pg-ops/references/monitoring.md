# RDS PostgreSQL Multi-Metric Correlation — AIOps

## Anomaly Patterns

| Pattern | Metrics | Detection Logic | Severity |
|---------|---------|-----------------|----------|
| Slow Query Cascade | SlowQueries > 50/min + AvgQueryTime > 5s + CPU > 85% | All AND over 10min | Critical |
| Connection Pool Exhaustion | Connections > 90% + ConnectionWait > 2s + ActiveConnections ~ max | Sustained 5min | Critical |
| Write-Ahead Log Pressure | WAL size > 80% disk + Checkpoints > 2/min + Disk IOPS > 90% | AND sustained 10min | Critical |
| Dead Tuple Accumulation | Dead tuples > 1M + Table bloat > 30% + Autovacuum running > 1hr | AND over 30min | Warning |
| Replication Lag Spike | RO lag > 30s + WAL send queue growth > 500MB + Write QPS > 5000 | Sustained 5min | Warning |
| Autovacuum Storm | Autovacuum processes > 20 + CPU from vacuum > 50% + Temp file creation | AND over 15min | Warning |

## Cross-Skill Diagnosis Decision Tree

```
[Alarm: RDS PostgreSQL issue]
    │
    ├── Is it query-performance related?
    │   ├── SlowQueries spike → Check EXPLAIN ANALYZE → Missing index or bad plan
    │   ├── Lock waits → Check pg_locks → ve-rds-pg-ops (blocking queries)
    │   └── Temp files created → work_mem too low or missing index → Sort to disk
    │
    ├── Is it resource-exhaustion?
    │   ├── CPU > 90% → ve-ecs-ops (app connection leak)
    │   ├── WAL disk pressure → wal_keep_size too large → Check replication lag
    │   ├── Disk > 85% → WAL or user data? → pg_wal analysis
    │   └── Memory > 85% → shared_buffers or per-connection work_mem?
    │
    ├── Is it vacuum-related?
    │   ├── Dead tuples growing → Autovacuum not keeping up → Increase autovacuum params
    │   └── Table bloat > 30% → Schedule VACUUML FULL or pg_repack
    │
    ├── Is it replication-related?
    │   ├── RO lag > 30s → Heavy writes or RO node undersized
    │   └── Replication slot lag → Slot inactive, WAL accumulating → Drop stale slot
    │
    └── Unknown → ve-cms-ops for cross-service correlation
```

## Delegation Matrix

| Alarm Signal | Primary Skill | Secondary Skill | Fallback |
|-------------|--------------|-----------------|----------|
| Slow query surge | Application team | ve-rds-pg-ops (explain plan) | ve-das-ops |
| Connection pool full | ve-rds-pg-ops | ve-ecs-ops (app server) | ve-cms-ops |
| WAL disk pressure | ve-rds-pg-ops | ve-ecs-ops (disk I/O) | Manual |
| Replication lag > 30s | ve-rds-pg-ops | ve-cms-ops | Manual |
| Table bloat > 30% | ve-rds-pg-ops | Application team | ve-das-ops |

## Proactive Inspection Checklist

### Resource Health
- [ ] CPU avg < 70%, peak < 85% over 24h
- [ ] Memory < 80% (shared_buffers + per-connection efficient)
- [ ] Disk < 80% with > 50GB free
- [ ] Max connections < 80% of limit
- [ ] Slow queries < 10/min
- [ ] Replication lag < 5s on all RO nodes
- [ ] Dead tuple ratio < 10% on active tables
- [ ] Table bloat < 20% on critical tables

### Security
- [ ] IP whitelist restricted
- [ ] No superuser used by application
- [ ] SSL/TLS enabled
- [ ] No default postgres password

### Cost
- [ ] No instances with CPU < 5% for 14 days
- [ ] No instances with storage < 20% used for 14 days
- [ ] RO instances with connections < 10% → consider removal
- [ ] Backup retention aligned with compliance needs

## Alarm Storm Handling

**Detection:** > 10 table-level or query alarms within 5 min from same instance.

**Suppression:** Correlate by instance → root alarm = earliest slow query or WAL pressure → group table alarms → single "query performance degraded" notification.

## Multi-Round Diagnosis Review

1. **Fact Check**: Metrics from correct instance? WAL size verified?
2. **Causal Analysis**: Is slow query root cause or symptom of bloat?
3. **Solution Validation**: Will index creation or VACUUM lock production tables?

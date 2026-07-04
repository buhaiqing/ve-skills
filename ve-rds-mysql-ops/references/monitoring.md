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

## AIOps — Intelligent Operations

> **TE-7:** Deep AIOps content → [`references/advanced/aiops.md`](advanced/aiops.md)

See [`references/advanced/aiops.md`](advanced/aiops.md) for:
- Cross-Skill Diagnosis Decision Tree
- Delegation Matrix
- Proactive Inspection Checklist
- Alarm Storm Handling
- Multi-Round Diagnosis Review

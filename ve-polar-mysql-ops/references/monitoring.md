# PolarDB MySQL Multi-Metric Correlation — AIOps

## Anomaly Patterns

| Pattern | Metrics | Detection Logic | Severity |
|---------|---------|-----------------|----------|
| Compute Saturation | CPU > 90% + ActiveConnections > 80% max + QueryTime > 5s | All AND over 10min | Critical |
| Storage Pressure | StorageUsed > 85% total + IOPS > 90% limit + QueueDepth > 10 | Sustained 15min | Critical |
| Connection Exhaustion | Connections > 90% max + ConnectionErrors spiking + NewConnections > threshold | Sustained 5min | Critical |
| Node Failover Detection | Primary node change + ConnectionDip + FailoverEvents > 0 | Immediate | Warning |
| Slow Query Cascade | SlowQueries > 50/min + AvgQueryTime > 5s + CPU > 85% | All AND over 10min | Critical |
| Buffer Pool Inefficiency | Hit rate < 95% + Pages read/s > threshold + Temp tables > 100/s | AND over 15min | Warning |
| Lock Contention | Lock wait > 30s + Lock waits/s > 100 + Deadlocks > 5/min | AND over 5min | Critical |
| Replication Anomaly | RO node latency > threshold (should be 0 for PolarDB) | Sustained 1min | Warning |
| Storage IO Throttling | IOWait > 50% + Throughput < expected + Latency > 20ms | AND over 10min | Critical |

## Cross-Skill Diagnosis Decision Tree

```
[Alarm: PolarDB MySQL issue]
    │
    ├── Is it query-performance related?
    │   ├── SlowQueries spike → Check execution plan → Application query issue
    │   ├── Full table scan detected → Index missing or stats stale
    │   └── Lock waits → ve-polar-mysql-ops (parameter tuning)
    │
    ├── Is it resource-exhaustion?
    │   ├── CPU > 90% + Connections high → ve-ecs-ops (if compute-bound)
    │   ├── Storage > 85% → Scale storage → ve-polar-mysql-ops
    │   ├── Memory > 85% → Buffer pool optimization → ve-polar-mysql-ops
    │   └── IOPS > 90% → Check for heavy read/write pattern → Scale compute or storage
    │
    ├── Is it failover/HA related?
    │   ├── Failover detected → Check node health → ve-polar-mysql-ops
    │   ├── Node status ERROR → Restart node or failover
    │   └── Connection drops → Endpoint configuration check
    │
    ├── Is it storage-layer related?
    │   ├── Storage scaling stuck → ve-polar-mysql-ops (storage operations)
    │   ├── IO latency high → Check storage pool metrics
    │   └── Storage quota exceeded → ve-billing-ops
    │
    └── Unknown/cluster-wide → ve-cms-ops for correlation + ve-vpc-ops for network
```

## Delegation Matrix

| Alarm Signal | Primary Skill | Secondary Skill | Fallback |
|-------------|--------------|-----------------|----------|
| Slow query surge | Application team | ve-polar-mysql-ops (explain plan) | ve-rds-mysql-ops |
| Connection pool full | ve-polar-mysql-ops | ve-ecs-ops (app server) | ve-cms-ops |
| Storage usage > 90% | ve-polar-mysql-ops | ve-billing-ops (quota) | Manual |
| Failover detected | ve-polar-mysql-ops | ve-cms-ops (monitoring) | Manual |
| Node status ERROR | ve-polar-mysql-ops | ve-ecs-ops (host issues) | ve-cms-ops |
| Network bandwidth cap | ve-vpc-ops | ve-polar-mysql-ops | Manual |
| Compute saturation | ve-polar-mysql-ops (scale) | ve-ecs-ops | ve-cms-ops |

## Key Metrics Reference

### Compute Metrics

| Metric | Namespace | Unit | Healthy Threshold | Critical Threshold |
|--------|-----------|------|-------------------|-------------------|
| CPUUtilization | Volcengine/PolarDB | % | < 70% | > 90% |
| MemoryUtilization | Volcengine/PolarDB | % | < 80% | > 90% |
| ActiveConnections | Volcengine/PolarDB | count | < 80% max | > 95% max |
| ConnectionsPerNode | Volcengine/PolarDB | count | < 500 | > 1000 |

### Storage Metrics

| Metric | Namespace | Unit | Healthy Threshold | Critical Threshold |
|--------|-----------|------|-------------------|-------------------|
| StorageSpace | Volcengine/PolarDB | GB | > 20% free | < 15% free |
| StorageUsed | Volcengine/PolarDB | GB | — | — |
| IOPS | Volcengine/PolarDB | ops/s | < 80% limit | > 95% limit |
| IOLatency | Volcengine/PolarDB | ms | < 10ms | > 20ms |
| Throughput | Volcengine/PolarDB | MB/s | < 80% limit | > 95% limit |

### Query Performance Metrics

| Metric | Namespace | Unit | Healthy Threshold | Critical Threshold |
|--------|-----------|------|-------------------|-------------------|
| SlowQueries | Volcengine/PolarDB | count/min | < 10 | > 50 |
| QueriesPerSecond | Volcengine/PolarDB | QPS | — | — |
| AvgQueryTime | Volcengine/PolarDB | ms | < 100ms | > 5000ms |
| TempTablesCreated | Volcengine/PolarDB | count/s | < 10 | > 100 |

### Node Health Metrics

| Metric | Namespace | Unit | Healthy Threshold | Critical Threshold |
|--------|-----------|------|-------------------|-------------------|
| NodeStatus | Volcengine/PolarDB | enum | RUNNING | ERROR |
| FailoverCount | Volcengine/PolarDB | count | 0 | > 1/hour |
| NodeRestartCount | Volcengine/PolarDB | count | 0 | > 1/day |

## Proactive Inspection Checklist

### Resource Health

- [ ] CPU avg < 70%, peak < 90% over 24h
- [ ] Memory usage < 80% (buffer pool efficient)
- [ ] Storage usage < 80% with > 100GB free
- [ ] Max connections < 80% of max_connections
- [ ] Slow queries < 10/min
- [ ] IOPS < 80% of provisioned limit
- [ ] IO latency < 10ms average
- [ ] Buffer pool hit rate > 97%

### High Availability

- [ ] Both primary and secondary nodes healthy
- [ ] Failover events < 1 per day (non-planned)
- [ ] Node restart count = 0 (unexpected)
- [ ] Read-only nodes (if any) all RUNNING

### Security

- [ ] IP whitelist restricted (not 0.0.0.0/0 without need)
- [ ] No superuser accounts used by application
- [ ] Backup encryption enabled
- [ ] SSL/TLS for all connections where supported

### Cost

- [ ] No clusters with CPU < 5% for 14 days (underutilized)
- [ ] No clusters with storage < 20% used for 14 days
- [ ] RO nodes with connections < 10% → consider removing
- [ ] Backup retention period aligned with compliance (not over-retained)
- [ ] Right-sized node classes (not over-provisioned)

## Alert Storm Handling

**Detection:** > 10 metric-level alarms within 5 min, or cascading alerts across nodes.

**Suppression:**
1. Correlate by cluster + time window
2. Root alarm = earliest compute/storage pressure or connection limit
3. Group: all node-level alarms → single "cluster performance degraded"
4. Throttle: 1 notification per root cause

## Multi-Round Diagnosis Review

1. **Fact Check**: Metrics from correct cluster? Storage metrics verified?
2. **Causal Analysis**: Is high CPU the root cause or symptom of slow queries?
3. **Solution Validation**: Will scaling compute affect running transactions?
4. **Risk Assessment**: What is impact of restart vs continued degraded performance?

## Monitoring Integration

### Cloud Monitor (CMS) Integration

```bash
# Describe metric data via CMS API
ve cms DescribeMetricData \
  --Region cn-beijing \
  --Namespace Volcengine/PolarDB \
  --MetricName CPUUtilization \
  --Dimensions '[{"ClusterId":"pc-xxx"}]' \
  --StartTime "2026-05-27T00:00:00+08:00" \
  --EndTime "2026-05-27T23:59:59+08:00"
```

### Key Alert Rules

| Alert Name | Condition | Severity | Action |
|------------|-----------|----------|--------|
| High CPU | CPU > 90% for 5min | Warning | Scale compute or investigate queries |
| Critical CPU | CPU > 95% for 5min | Critical | Scale compute immediately |
| Storage Critical | Storage > 90% | Critical | Scale storage immediately |
| Connection Limit | Connections > 90% max | Warning | Scale compute or optimize connections |
| Slow Query Spike | SlowQueries > 50/min | Warning | Investigate query performance |
| Node Down | NodeStatus != RUNNING | Critical | Restart node or failover |
| Failover Occurred | FailoverEvents > 0 | Warning | Investigate cause |

## Log Analysis

### Slow Query Log

Access via:
- Console download
- Direct query to `mysql.slow_log` table (if enabled)

### Error Log

Key patterns to watch:
```
[ERROR] InnoDB: Unable to allocate memory
[ERROR] Too many connections
[Warning] Aborted connection
```

### Audit Log

Enable for:
- DDL operations tracking
- Privilege changes
- Connection auditing

# Monitoring MongoDB

## Key Metrics

### Instance Health Metrics

| Metric | Namespace | Unit | Description |
|--------|-----------|------|-------------|
| `InstanceStatus` | MongoDB | Enum | RUNNING, CREATING, ERROR, etc. |
| `CPUUtilization` | Volcengine_MongoDB_CPU | Percent | CPU usage percentage |
| `MemoryUtilization` | Volcengine_MongoDB_Memory | Percent | Memory usage percentage |
| `StorageUtilization` | Volcengine_MongoDB_Storage | Percent | Storage usage percentage |
| `StorageFreeSpace` | Volcengine_MongoDB_StorageFree | GB | Free storage space |
| `ConnectionCount` | Volcengine_MongoDB_Connections | Count | Current connections |
| `ConnectionUtilization` | Volcengine_MongoDB_ConnUtil | Percent | Connection usage vs max |
| `QPS` | Volcengine_MongoDB_QPS | Count/Sec | Queries per second |
| `TPS` | Volcengine_MongoDB_TPS | Count/Sec | Transactions per second |

### Performance Metrics

| Metric | Namespace | Unit | Description |
|--------|-----------|------|-------------|
| `QueryTime` | Volcengine_MongoDB_QueryTime | ms | Average query latency |
| `SlowQueryCount` | Volcengine_MongoDB_SlowQueries | Count/Min | Slow queries per minute |
| `IndexHitRate` | Volcengine_MongoDB_IndexHit | Percent | Index hit rate |
| `CacheHitRate` | Volcengine_MongoDB_CacheHit | Percent | WiredTiger cache hit rate |
| `ScanAndOrderCount` | Volcengine_MongoDB_ScanOrder | Count/Min | Sort operations without index |
| `LockWaitTime` | Volcengine_MongoDB_LockWait | ms | Average lock wait time |

### Replication Metrics

| Metric | Namespace | Unit | Description |
|--------|-----------|------|-------------|
| `ReplicationLag` | Volcengine_MongoDB_ReplLag | Seconds | Secondary replication lag |
| `OplogWindow` | Volcengine_MongoDB_OplogWindow | Hours | Oplog time window |
| `HeartbeatDelay` | Volcengine_MongoDB_HBDelay | ms | Node heartbeat delay |

### Resource Metrics

| Metric | Namespace | Unit | Description |
|--------|-----------|------|-------------|
| `DiskIOPS` | Volcengine_MongoDB_DiskIOPS | Count/Sec | Disk IOPS |
| `DiskThroughput` | Volcengine_MongoDB_DiskTP | MB/Sec | Disk throughput |
| `NetworkIn` | Volcengine_MongoDB_NetIn | MB/Sec | Network inbound |
| `NetworkOut` | Volcengine_MongoDB_NetOut | MB/Sec | Network outbound |

## Alerting Rules

### Critical Alerts

| Alert | Condition | Severity | Action |
|-------|-----------|----------|--------|
| Instance Down | `InstanceStatus != RUNNING` for 5m | P0 | Page on-call |
| High CPU | `CPUUtilization > 90%` for 10m | P1 | Scale up or optimize |
| High Memory | `MemoryUtilization > 90%` for 10m | P1 | Scale up or optimize |
| Storage Full | `StorageUtilization > 85%` for 5m | P1 | Expand storage |
| Connection Exhaustion | `ConnectionUtilization > 90%` for 5m | P1 | Increase max connections |
| Replication Lag | `ReplicationLag > 60s` for 5m | P2 | Investigate replication |
| High Slow Queries | `SlowQueryCount > 100/min` for 10m | P2 | Optimize queries |

### Warning Alerts

| Alert | Condition | Severity | Action |
|-------|-----------|----------|--------|
| Moderate CPU | `CPUUtilization > 70%` for 15m | P3 | Monitor and plan |
| Moderate Memory | `MemoryUtilization > 70%` for 15m | P3 | Monitor and plan |
| Low Index Hit Rate | `IndexHitRate < 80%` for 30m | P3 | Review indexes |
| Low Cache Hit Rate | `CacheHitRate < 90%` for 30m | P3 | Consider memory upgrade |

## Monitoring Queries

### Using CloudMonitor (CMS)

```bash
# Query CPU utilization
ve cms DescribeMetricData \
  --Region cn-beijing \
  --Namespace Volcengine_MongoDB \
  --MetricName CPUUtilization \
  --Dimensions '[{"InstanceId":"mongo-xxx"}]' \
  --StartTime 2026-05-27T00:00:00+08:00 \
  --EndTime 2026-05-27T23:59:59+08:00 \
  --Period 300

# Query connection count
ve cms DescribeMetricData \
  --Region cn-beijing \
  --Namespace Volcengine_MongoDB \
  --MetricName ConnectionCount \
  --Dimensions '[{"InstanceId":"mongo-xxx"}]' \
  --StartTime 2026-05-27T00:00:00+08:00 \
  --EndTime 2026-05-27T23:59:59+08:00 \
  --Period 60
```

### Using MongoDB Shell

```javascript
// Server status
rs.status()

// Connection info
db.serverStatus().connections

// Memory usage
db.serverStatus().mem

// Replication status
rs.printSecondaryReplicationInfo()

// Current operations
db.currentOp()

// Slow queries (requires profiling enabled)
db.system.profile.find().sort({ ts: -1 }).limit(10)

// Collection stats
db.getCollection("mycollection").stats()

// Database stats
db.stats()
```

## Dashboards

### Essential Dashboard Panels

1. **Instance Health**
   - Instance status over time
   - Node availability

2. **Resource Utilization**
   - CPU usage (all nodes)
   - Memory usage (all nodes)
   - Storage usage trend

3. **Performance**
   - QPS/TPS trends
   - Query latency percentiles
   - Slow query rate

4. **Connections**
   - Active connections
   - Connection utilization
   - Connection pool metrics

5. **Replication**
   - Replication lag (all secondaries)
   - Oplog window
   - Heartbeat delays

## Log Analysis

### MongoDB Logs

```bash
# Download logs via console or API
# Analyze slow queries
grep -i "slow" mongodb.log | tail -n 100

# Analyze errors
grep -i "error" mongodb.log | tail -n 50

# Connection events
grep -i "connection" mongodb.log | tail -n 100
```

### Audit Logs

Enable audit logging for security monitoring:

```javascript
// Configure audit log (via parameter)
{
  "auditLog": {
    "destination": "file",
    "format": "JSON",
    "path": "/var/log/mongodb/audit.log"
  }
}
```

## Best Practices

### Monitoring Setup

1. Enable all default metrics
2. Set up alerts for critical thresholds
3. Create custom dashboards
4. Configure log collection
5. Set up audit logging for compliance

### Alert Tuning

1. Adjust thresholds based on workload patterns
2. Use percentiles for latency metrics (p95, p99)
3. Set different alert levels for different environments
4. Include runbook links in alert notifications

### Capacity Planning

1. Monitor storage growth trend
2. Track connection growth
3. Analyze query patterns
4. Plan upgrades before hitting limits

### Troubleshooting

1. Correlate metrics with application events
2. Use time-series analysis for trend detection
3. Compare primary and secondary metrics
4. Monitor replication lag during high write loads

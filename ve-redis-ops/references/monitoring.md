# Redis Multi-Metric Correlation — AIOps

## Anomaly Patterns

| Pattern | Metrics | Detection Logic | Severity |
|---------|---------|-----------------|----------|
| Memory Saturation | Memory > 85% + Evictions > 0 + Hit rate < 90% | All AND over 5min | Critical |
| Connection Storm | Connections > 90% + New connections > 100/s + Rejected > 0 | Concurrent thresholds | Critical |
| Slow Command Accumulation | Latency p99 > 100ms + Slow log count > 10/min + CPU > 70% | All AND over 10min | Warning |
| Network I/O Bottleneck | Bandwidth > 85% + Dropped packets > 0 + Latency > 50ms | AND sustained 10min | Critical |
| Replication Lag (cluster) | Sync offset growing + Failed sync commands > 0 | Sustained 5min | Warning |
| Failover Event | Primary node unreachable + Automatic failover triggered | State change | Critical |

## Cross-Skill Diagnosis Decision Tree

```
[Alarm: Redis instance issue]
    │
    ├── Is it memory-related?
    │   ├── Memory > 85% → What eviction policy? → If noeviction: urgent scale
    │   ├── Evictions > 0 → Is hit rate dropping? → Optimize key TTL
    │   └── Memory spike sudden → Check for key pattern: big keys → application issue
    │
    ├── Is it connection-related?
    │   ├── Connections > 90% → Connection leak in app → ve-ecs-ops (check app logs)
    │   └── Rejected connections → AllowList not configured → Configure whitelist
    │
    ├── Is it performance-related?
    │   ├── Slow commands → Check SLOWLOG → Identify KEYS/scan operations
    │   └── High latency → Network or bandwidth? → ve-vpc-ops
    │
    └── Cluster failover → ve-cms-ops for AZ health check
```

## Delegation Matrix

| Alarm Signal | Primary Skill | Secondary Skill | Fallback |
|-------------|--------------|-----------------|----------|
| Memory > 90% sustained | ve-redis-ops | Application team | ve-cms-ops |
| Connection exhaustion | ve-redis-ops | ve-ecs-ops (app server) | ve-cms-ops |
| VPC bandwidth limit | ve-vpc-ops | ve-redis-ops | Manual |
| Failover triggered | ve-cms-ops | ve-redis-ops | Manual |
| Slow log accumulation | Application team | ve-redis-ops (SLOWLOG) | Manual |

## Proactive Inspection Checklist

### Resource Health
- [ ] Memory usage < 80% across all instances
- [ ] EvictedKeys = 0 or stable
- [ ] Cache hit rate > 95%
- [ ] Connections < 70% of max
- [ ] Bandwidth < 70% of limit
- [ ] AllowList configured (not open to all)

### Security
- [ ] No password-less instances (v5.0+)
- [ ] AllowList restricted to app server IPs only
- [ ] No direct internet access (VPC-only)

### Cost
- [ ] No instances with memory < 10% used for 14 days
- [ ] No instances with connections < 5% for 7 days

## Alarm Storm Handling

**Detection:** > 10 key eviction or latency alarms within 2 min, or cascading across shard replicas.

**Suppression:**
1. Correlate by instance + time window
2. Root alarm = earliest memory spike or node failure
3. Group: all shard-level alarms → single "instance degraded"
4. Throttle: 1 notification per root cause

## Multi-Round Diagnosis Review

1. **Fact Check**: Metrics current? Instance status valid?
2. **Causal Analysis**: Is the application pattern truly the cause?
3. **Solution Validation**: Will scaling config or eviction change resolve the issue?

# Monitoring — ECS

## Key Metrics (Volcengine Cloud Monitor)

ECM namespace: `Volcengine_ECS_*`

### CPU Metrics

| Metric | Description | Threshold (Warning) | Threshold (Critical) |
|--------|-------------|---------------------|---------------------|
| `CpuUsage` | CPU utilization (%) | > 80% for 5min | > 95% for 5min |
| `CpuUser` | User-space CPU time | — | — |
| `CpuSystem` | Kernel-space CPU time | — | — |
| `CpuIdle` | Idle CPU time | < 10% | < 5% |

### Memory Metrics

| Metric | Description | Threshold (Warning) | Threshold (Critical) |
|--------|-------------|---------------------|---------------------|
| `MemoryUtilization` | Memory usage (%) | > 85% for 5min | > 95% for 5min |
| `MemoryUsed` | Memory used (bytes) | — | — |
| `MemoryFree` | Memory free (bytes) | < 500MB | < 200MB |

### Disk Metrics

| Metric | Description | Threshold (Warning) | Threshold (Critical) |
|--------|-------------|---------------------|---------------------|
| `DiskReadIOPS` | Disk read operations/sec | — | — |
| `DiskWriteIOPS` | Disk write operations/sec | — | — |
| `DiskReadBps` | Disk read throughput (bytes/s) | — | — |
| `DiskWriteBps` | Disk write throughput (bytes/s) | — | — |
| `DiskUsage` | Disk utilization (%) | > 80% | > 95% |

### Network Metrics

| Metric | Description | Threshold (Warning) | Threshold (Critical) |
|--------|-------------|---------------------|---------------------|
| `NetworkInBps` | Inbound bandwidth (bits/s) | > 80% of limit | > 95% of limit |
| `NetworkOutBps` | Outbound bandwidth (bits/s) | > 80% of limit | > 95% of limit |
| `NetworkInPps` | Inbound packets/sec | — | — |
| `NetworkOutPps` | Outbound packets/sec | — | — |
| `NetworkDropOut` | Outbound packet drops | > 0/sec sustained | > 100/sec |
| `TcpConnectionCount` | Active TCP connections | > 40,000 | > 60,000 |

## Anomaly Patterns

| Pattern | Metrics Involved | Detection Logic | Severity |
|---------|-----------------|-----------------|----------|
| CPU-Memory Pressure | CPU > 90% + Memory > 85% | AND over 5min window | Critical |
| Disk I/O Bottleneck | DiskReadIOPS > 90% limit + QueueDepth > 4 | Sustained for 10min | Critical |
| Network Congestion | NetworkDropOut > 0 + NetworkOutBps > limit | Concurrent threshold | Warning |
| Instance Stuck | CPU < 1% + NetworkInPps = 0 for 10min | Instance unresponsive | Critical |
| Runaway Process | CPU > 95% single core + Memory growth > 100MB/min | Rapid resource growth | Critical |

## Monitoring Query

```bash
# Query instance metrics via ve CLI (if supported) or monitoring API
ve cms DescribeMetricData --Region cn-beijing --Namespace Volcengine_ECS --MetricName CpuUsage --InstanceId i-xxx
```

> For comprehensive monitoring workflows, delegate to `ve-cms-ops` (when present).

## Alert Delegation Matrix

| Alarm Source | Primary Diagnosis Skill | Secondary Skill |
|-------------|------------------------|-----------------|
| ECS CPU > 90% | ve-ecs-ops | ve-cms-ops |
| ECS Disk > 85% | ve-ecs-ops | ve-cms-ops |
| ECS Network Drop | ve-ecs-ops | ve-vpc-ops |
| ECS Instance Unreachable | ve-ecs-ops | ve-vpc-ops |
| ECS Instance System Event | ve-ecs-ops | ve-cms-ops |

## AIOps — Intelligent Operations

### Cross-Skill Diagnosis Decision Tree

```
[ECS Alarm Triggered]
    │
    ├── Is it CPU-related?
    │   ├── Yes → Is memory also high?
    │   │   ├── Yes → Application-level issue (check logs via Cloud Assistant, GC, threads)
    │   │   │       └── Delegate to application skill if Java/JVM
    │   │   └── No → Check for runaway processes (Cloud Assistant: top, ps aux)
    │   │            └── If single process: kill or restart
    │   │            └── If system-wide: reboot
    │   └── No → Continue...
    │
    ├── Is it disk-related?
    │   ├── Disk usage > 90% → Check log files, temporary files (Cloud Assistant)
    │   │   └── If log-related: set up rotation → delegate to app team
    │   │   └── If data-related: expand disk or add data disk
    │   └── Disk IOPS > 90% → Check database queries, backup jobs
    │       └── If database: delegate to ve-rds-ops
    │       └── If backup: reschedule to off-peak
    │
    ├── Is it network-related?
    │   ├── High latency → Check upstream dependencies
    │   │   └── Delegate to ve-vpc-ops for network path analysis
    │   ├── Packet loss → Check security groups, ACLs
    │   │   └── Recent SG change? Rollback → delegate to security team
    │   └── Connection limit → Check application connection pool
    │       └── Delegate to application skill
    │
    └── Unknown pattern → Delegate to ve-cms-ops for correlation analysis
```

### Alarm Storm Handling

**Detection Criteria:**
- > 10 alarms within 5 minutes from same ECS resource group
- > 50% of alarms share the same root cause metric
- Alarm rate exceeds 3x the baseline rate

**Suppression Workflow:**
1. Correlate alarms by instance ID and time window
2. Identify root alarm (earliest or highest severity)
3. Group related alarms under root alarm
4. Notify once per root cause (not per alarm)
5. Execute root cause diagnosis on primary alarm
6. After resolution, verify all related alarms clear

### Proactive Inspection Checklist

```markdown
## ECS Proactive Inspection — [Date]

### Resource Health
- [ ] CPU usage < 70% across all instances (avg over 7 days)
- [ ] Memory usage < 80% across all instances (avg over 7 days)
- [ ] Disk usage < 80% with > 50GB free space
- [ ] Network error rate < 0.1%
- [ ] No instances in ERROR state

### Cost Optimization
- [ ] No idle instances (CPU < 5% for 7 days)
- [ ] No stopped instances > 7 days without planned restart
- [ ] No unattached cloud disks
- [ ] No snapshots older than 90 days without retention policy
- [ ] Reserved instance coverage > 60% for steady workloads

### Security Posture
- [ ] No instances with public IP unless explicitly required
- [ ] Security group rules follow least privilege (no 0.0.0.0/0 on non-HTTP ports)
- [ ] No instances without Cloud Assistant installed
- [ ] Deletion protection enabled for production instances

### Reliability
- [ ] Multi-AZ deployment for production workloads
- [ ] Automated backups configured for instances with data disks
- [ ] Health checks configured for load-balanced instances
```

### Multi-Round Diagnosis Review

Before finalizing any ECS diagnosis:

1. **Fact Check:** Are the ECS metrics current? Are thresholds correct?
2. **Causal Analysis:** Is the identified cause the true root cause? Could something else explain the symptoms?
3. **Solution Validation:** Will the fix actually resolve the issue? Could it cause side effects?

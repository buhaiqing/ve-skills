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

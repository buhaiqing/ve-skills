# Monitoring — TOS

## TOS does not expose built-in metrics by default unlike ECS. Metrics are available via Volcengine Cloud Monitor (CMS).

## Key Metrics

| Metric | Description | Threshold (Warning) | Threshold (Critical) |
|--------|-------------|---------------------|---------------------|
| `RequestCount` | Requests per minute | Monitor trend | — |
| `4xxErrorRate` | Client error rate (%) | > 1% | > 5% |
| `5xxErrorRate` | Server error rate (%) | > 0.1% | > 1% |
| `BandwidthIn` | Inbound bandwidth (bytes/s) | — | — |
| `BandwidthOut` | Outbound bandwidth (bytes/s) | — | — |
| `StorageUsed` | Total storage (bytes) | — | — |
| `ObjectCount` | Total object count | — | — |
| `FirstByteLatency` | Time to first byte (ms) | > 1000ms | > 5000ms |
| `ActiveConnections` | Active connections | — | — |

## Anomaly Patterns

| Pattern | Metrics | Detection Logic | Severity |
|---------|---------|-----------------|----------|
| Error Spike | 4xxErrorRate > 5% | Sustained for 5min | Warning |
| Access Denied Flood | AccessDenied rate spikes | Sudden increase > 10x | Critical |
| Bandwidth Saturation | BandwidthOut > 90% of limit | Sustained for 10min | Warning |
| High Latency | FirstByteLatency > 5s for GET | Sustained for 5min | Warning |
| Storage Surge | StorageUsed growth > 50% in 1hr | Rapid increase | Warning |

## Monitoring Query

```bash
# Query TOS metrics via CMS API
ve cms DescribeMetricData --Region cn-beijing --Namespace Volcengine_TOS --MetricName RequestCount
```

## Alert Delegation Matrix

| Alarm Source | Primary Diagnosis Skill | Secondary Skill |
|-------------|------------------------|-----------------|
| TOS 4xx/5xx error spike | ve-tos-ops | ve-cms-ops |
| TOS storage growth anomaly | ve-tos-ops | ve-cms-ops |
| TOS access denied flood | ve-tos-ops | ve-iam-ops |
| TOS bandwidth saturation | ve-tos-ops | ve-vpc-ops |

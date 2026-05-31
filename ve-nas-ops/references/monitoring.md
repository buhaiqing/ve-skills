# Monitoring NAS

## Key Metrics

| Metric | Description | Namespace |
|--------|-------------|-----------|
| `NasIOPSRead` | Read IOPS | `Volcengine_NAS` |
| `NasIOPSWrite` | Write IOPS | `Volcengine_NAS` |
| `NasThroughputRead` | Read throughput (MB/s) | `Volcengine_NAS` |
| `NasThroughputWrite` | Write throughput (MB/s) | `Volcengine_NAS` |
| `NasLatencyRead` | Read latency (ms) | `Volcengine_NAS` |
| `NasLatencyWrite` | Write latency (ms) | `Volcengine_NAS` |
| `NasUsedCapacityGiB` | Used storage capacity | `Volcengine_NAS` |

## Alert Recommendations

| Alert Name | Condition | Severity |
|------------|-----------|----------|
| High Capacity Usage | `NasUsedCapacityGiB > 80% of quota` | Warning |
| Critical Capacity | `NasUsedCapacityGiB > 95% of quota` | Critical |
| High Latency | `NasLatencyRead > 20ms` | Warning |
| High IOPS | `NasIOPSRead > 80% of tier limit` | Warning |

# Monitoring — CMS

CMS is the monitoring platform itself. For monitoring metrics **of other cloud services** via CMS, see the respective product ops skills.

## CMS Service Health

| Metric | Description | Threshold (Warning) |
|--------|-------------|---------------------|
| API Error Rate | % of CMS API calls returning errors | > 1% |
| API Latency | Average response time for CMS APIs | > 2000ms |
| Alarm Delivery Latency | Time from alarm trigger to notification | > 5 minutes |
| Alarm Success Rate | % of alarms successfully delivered | < 99% |

## Cross-Alarm Diagnosis Patterns

| Pattern | Correlated Metrics | Diagnosis |
|---------|-------------------|-----------|
| App Slow | DB connections ↑ + CPU ↑ + latency ↑ | Database bottleneck |
| Disk Full | DiskUtilization ↑ + IOPS ↑ + errors ↑ | Log buildup; cleanup needed |
| Network Drop | BandwidthOut → limit + DropRate > 0 | Scale up bandwidth or CDN |
| Memory Leak | MemoryUtilization ↑ steadily over hours | Restart or investigate application |

## Alarm Storm Handling

An alarm storm is detected when:
- > 10 alarms within 5 minutes from the same resource group
- > 50% of alarms share the same root cause metric

**Suppression workflow:**
1. Correlate alarms by resource and time window
2. Identify the root alarm (earliest or highest severity)
3. Group related alarms under the root alarm
4. Suppress duplicate notifications
5. Execute root cause diagnosis on the primary alarm

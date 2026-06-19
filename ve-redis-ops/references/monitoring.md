# Monitoring — Volcengine Redis

> **Purpose:** Monitoring guide for Redis resources on Volcengine.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [Cloud Monitor Metrics](#1-cloud-monitor-metrics)
2. [Memory Monitoring](#2-memory-monitoring)
3. [Connection Monitoring](#3-connection-monitoring)
4. [Performance Monitoring](#4-performance-monitoring)
5. [Alarm Configuration](#5-alarm-configuration)

---

## 1. Cloud Monitor Metrics

| Metric | Unit | Description |
|--------|------|-------------|
| `cpuusage` | % | CPU utilization |
| `memoryusedratio` | % | Memory utilization ratio |
| `intranet_in_bps` | bps | Inbound network traffic |
| `intranet_out_bps` | bps | Outbound network traffic |
| `newconnections` | Count/s | New connections per second |
| `totalconnections` | Count | Total active connections |
| `usedmemory` | Bytes | Memory currently used |
| `qps` | Count/s | Queries per second |

### Query via CMS CLI

```bash
ve cms DescribeMetricData \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --Namespace "redis" \
  --MetricName "cpuusage" \
  --Dimensions '{"InstanceId":"{{user.instance_id}}"}'
```

---

## 2. Memory Monitoring

### Capacity Utilization

```
MemoryUtilization = (Capacity.Used / Capacity.Total) × 100%
```

| Utilization | Concern Level | Action |
|-------------|--------------|--------|
| < 60% | ✅ Normal | — |
| 60–80% | ⚠️ Warning | Monitor growth rate |
| > 80% | 🔴 Critical | Increase capacity or tune eviction policy |
| > 90% | 🚨 Emergency | Risk of OOM errors |

### Key Analysis

- **Big Keys:** Use `DescribeBigKeys` to find keys consuming excessive memory
- **Hot Keys:** Use `DescribeHotKeys` to identify keys causing throughput bottlenecks
- **Recommendation:** Run scans during off-peak hours — they increase CPU load

---

## 3. Connection Monitoring

| Connection Utilization | Concern Level | Action |
|-----------------------|--------------|--------|
| < 50% | ✅ Normal | — |
| 50–75% | ⚠️ Warning | Review connection pooling |
| > 75% | 🔴 Critical | Increase max connections or add instances |

---

## 4. Performance Monitoring

### CPU High Alert

Redis is single-threaded per shard. CPU > 90% usually means:

- Slow commands blocking the event loop
- Too many requests per second for the shard capacity
- Key scanning or big key operations

**Resolution:**
1. Check slow logs: `ve redis DescribeSlowLogs`
2. Identify hot/big keys
3. Optimize application usage patterns

### High Latency

| Latency | Concern Level |
|---------|--------------|
| < 1ms | ✅ Normal |
| 1–10ms | ⚠️ Warning — network or load issue |
| > 10ms | 🔴 Critical — investigate immediately |

---

## 5. Alarm Configuration

### Recommended Alarms

| Alarm | Metric | Condition | Severity |
|-------|--------|-----------|----------|
| Memory High | `memoryusedratio` | > 80% for 5min | ⚠️ Warning |
| Memory Critical | `memoryusedratio` | > 90% for 5min | 🔴 Critical |
| CPU High | `cpuusage` | > 90% for 5min | 🔴 Critical |
| Connection High | `totalconnections` | > 75% of max | ⚠️ Warning |
| Instance Error | Instance status | Not `Running` | 🔴 Critical |

---

*This reference document is part of the ve-redis-ops skill.*

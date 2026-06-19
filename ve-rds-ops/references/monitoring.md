# Monitoring — Volcengine RDS MySQL

> **Purpose:** Monitoring guide for RDS MySQL resources on Volcengine.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [Cloud Monitor Metrics](#1-cloud-monitor-metrics)
2. [Connection Monitoring](#2-connection-monitoring)
3. [Query Performance](#3-query-performance)
4. [Storage Monitoring](#4-storage-monitoring)
5. [Replication Lag](#5-replication-lag)
6. [Alarm Configuration](#6-alarm-configuration)

---

## 1. Cloud Monitor Metrics

| Metric | Unit | Description |
|--------|------|-------------|
| `CpuUsage` | % | CPU utilization |
| `MemUsage` | % | Memory utilization |
| `IOPSUsage` | % | IOPS utilization |
| `DiskUsage` | % | Disk space utilization |
| `ConnectionUsage` | % | Connection usage (% of max) |
| `QPS` | Count/s | Queries per second |
| `TPS` | Count/s | Transactions per second |
| `NetworkIn` | Bytes/s | Inbound network traffic |
| `NetworkOut` | Bytes/s | Outbound network traffic |

### Query via CMS CLI

```bash
ve cms DescribeMetricData \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --Namespace "rds_mysql" \
  --MetricName "CpuUsage" \
  --Dimensions '{"InstanceId":"{{user.instance_id}}"}'
```

---

## 2. Connection Monitoring

Monitor connection count vs. maximum:

```
ConnectionUtilization = (ActiveConnections / MaxConnections) × 100%
```

| Utilization | Concern Level | Action |
|-------------|--------------|--------|
| < 50% | ✅ Normal | — |
| 50–75% | ⚠️ Warning | Monitor trend; consider max_connections increase |
| > 75% | 🚨 Critical | Increase max_connections or add read replicas |

---

## 3. Query Performance

### Slow Query Monitoring

```bash
ve rds_mysql DescribeSlowLogs --Region "$VOLCENGINE_REGION" --InstanceId "$INSTANCE_ID"
```

### Key Indicators

| Metric | Threshold | Action |
|--------|-----------|--------|
| QPS > 10,000 | ⚠️ Warning | Check index usage, optimize queries |
| Slow queries > 10/min | ⚠️ Warning | Analyze slow log, add indexes |
| TPS > 2,000 | ℹ️ Monitor for HA | Watch replication lag |

---

## 4. Storage Monitoring

| Disk Usage | Concern Level | Action |
|------------|--------------|--------|
| < 70% | ✅ Normal | — |
| 70–85% | ⚠️ Warning | Enable auto-scaling; clean old data |
| > 85% | 🚨 Critical | Expand storage or archive old data |
| > 95% | 🛡️ Emergency | Instance may go read-only |

---

## 5. Replication Lag

For HA and MultiNode instances, monitor replica lag:

| Lag Time | Concern Level | Action |
|----------|--------------|--------|
| < 1s | ✅ Normal | — |
| 1–10s | ⚠️ Warning | Check write load; optimize transactions |
| > 10s | 🚨 Critical | Reduce write throughput; investigate blocking |

---

## 6. Alarm Configuration

### Recommended Alarms

| Alarm | Metric | Condition | Severity |
|-------|--------|-----------|----------|
| CPU High | `CpuUsage` | > 80% for 5min | ⚠️ Warning |
| CPU Critical | `CpuUsage` | > 95% for 5min | 🚨 Critical |
| Memory High | `MemUsage` | > 85% for 5min | 🚨 Critical |
| Disk Space Low | `DiskUsage` | > 85% | ⚠️ Warning |
| Space Critical | `DiskUsage` | > 95% | 🚨 Critical |
| Connection Exhaustion | `ConnectionUsage` | > 75% | ⚠️ Warning |
| Replication Lag | `ReplicationLag` | > 10s | 🚨 Critical |

---

*This reference document is part of the ve-rds-ops skill.*

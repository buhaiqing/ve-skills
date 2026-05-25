# Monitoring — Volcengine EIP

> **Purpose:** Monitoring and observability guide for EIP resources on Volcengine.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [Cloud Monitor Metrics](#1-cloud-monitor-metrics)
2. [EIP Bandwidth Monitoring](#2-eip-bandwidth-monitoring)
3. [Traffic Monitoring](#3-traffic-monitoring)
4. [Alarm Configuration](#4-alarm-configuration)

---

## 1. Cloud Monitor Metrics

Volcengine Cloud Monitor (云监控) publishes EIP metrics automatically.

| Metric | Unit | Description |
|--------|------|-------------|
| `eip_in_bps` | bps | Inbound bandwidth rate |
| `eip_out_bps` | bps | Outbound bandwidth rate |
| `eip_in_pps` | pps | Inbound packets per second |
| `eip_out_pps` | pps | Outbound packets per second |
| `eip_in_bytes` | Bytes | Cumulative inbound bytes |
| `eip_out_bytes` | Bytes | Cumulative outbound bytes |

### Query via CMS CLI

```bash
ve cms DescribeMetricData \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --Namespace "VPC_EIP" \
  --MetricName "eip_out_bps" \
  --Dimensions '{"AllocationId":"{{user.eip_id}}"}' \
  --StartTime "{{user.start_time}}" \
  --EndTime "{{user.end_time}}"
```

---

## 2. EIP Bandwidth Monitoring

### Bandwidth Utilization Calculation

```
Utilization = (actual_traffic_bps / allocated_bandwidth_bps) × 100%
```

### Utilization Levels

| Utilization | Concern Level | Action |
|-------------|--------------|--------|
| < 60% | Normal | — |
| 60–80% | Warning | Monitor trend |
| 80–95% | High | Consider bandwidth increase |
| > 95% | Critical | Increase bandwidth immediately |

### Check Current EIP Bandwidth

```bash
ve eip DescribeEipBandwidth --Region "$VOLCENGINE_REGION" --AllocationId "$EIP_ID" \
  | jq '{
    bandwidth_mbps: .Result.Bandwidth,
    unit: .Result.BandwidthUnit
  }'
```

---

## 3. Traffic Monitoring

### Traffic Cost Estimation (PayByTraffic)

```
Daily Cost = Daily_GB_Transferred × Price_per_GB
```

Monitor cumulative byte counters to estimate monthly costs:

```bash
# Get inbound bytes
ve cms DescribeMetricData \
  --Namespace "VPC_EIP" \
  --MetricName "eip_in_bytes" \
  --Dimensions '{"AllocationId":"eipalloc-xxx"}'
```

### Bandwidth Spike Detection

Watch for sustained traffic > 80% of allocated bandwidth for > 5 minutes:

- Possible causes: DDoS, traffic surge, misconfigured application
- Actions: Increase bandwidth, enable traffic limiting, investigate source

---

## 4. Alarm Configuration

### Recommended Alarms for EIP

| Alarm | Metric | Condition | Severity |
|-------|--------|-----------|----------|
| Bandwidth Saturation | `eip_out_bps` | > 90% of allocated | Critical |
| Traffic Spike | `eip_in_bps` | > 3x 7-day avg for 5min | Warning |
| Packet Rate Spike | `eip_in_pps` | > 100,000 pps | Warning |
| EIP Binding Change | Resource lifecycle | Associate/Disassociate event | Info |

### Create Bandwidth Alarm

```bash
ve cms CreateAlarm \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --Namespace "VPC_EIP" \
  --MetricName "eip_out_bps" \
  --ComparisonOperator "GreaterThanThreshold" \
  --Threshold "{{user.threshold_bps}}" \
  --EvaluationCount 3 \
  --Period 60 \
  --ContactGroup "{{user.contact_group}}"
```

---

*This reference document is part of the ve-eip-ops skill.*

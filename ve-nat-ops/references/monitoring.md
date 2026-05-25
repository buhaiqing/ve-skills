# Monitoring — Volcengine NAT Gateway

> **Purpose:** Monitoring guide for NAT Gateway resources on Volcengine.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [Cloud Monitor Metrics](#1-cloud-monitor-metrics)
2. [SNAT Connection Monitoring](#2-snat-connection-monitoring)
3. [DNAT Connection Monitoring](#3-dnat-connection-monitoring)
4. [Alarm Configuration](#4-alarm-configuration)

---

## 1. Cloud Monitor Metrics

| Metric | Unit | Description |
|--------|------|-------------|
| `natgw_in_bps` | bps | Inbound bandwidth (DNAT traffic) |
| `natgw_out_bps` | bps | Outbound bandwidth (SNAT traffic) |
| `natgw_in_pps` | pps | Inbound packet rate |
| `natgw_out_pps` | pps | Outbound packet rate |
| `natgw_active_connections` | Count | Active concurrent connections |

### Query via CMS CLI

```bash
ve cms DescribeMetricData \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --Namespace "VPC_NATGW" \
  --MetricName "natgw_out_bps" \
  --Dimensions '{"NatGatewayId":"{{user.nat_id}}"}'
```

---

## 2. SNAT Connection Monitoring

### Connection Saturation

Each EIP in a SNAT rule supports ~200,000 concurrent connections. Monitor active connections:

| Connection Count per EIP | Concern Level |
|-------------------------|--------------|
| < 100,000 | Normal |
| 100,000–160,000 | Warning |
| > 160,000 | Critical — add more EIPs |

### Bandwidth Saturation

Monitor outbound traffic vs NAT Gateway spec bandwidth:

| Spec | Bandwidth Limit | Alarm Threshold |
|------|----------------|-----------------|
| Small | 1 Gbps | > 800 Mbps |
| Medium | 5 Gbps | > 4 Gbps |
| Large | 10 Gbps | > 8 Gbps |

---

## 3. DNAT Connection Monitoring

Watch DNAT rule connection counts per port:

```bash
ve cms DescribeMetricData \
  --Namespace "VPC_NATGW" \
  --MetricName "natgw_active_connections" \
  --Dimensions '{"NatGatewayId":"ngw-xxx"}'
```

---

## 4. Alarm Configuration

### Recommended Alarms

| Alarm | Metric | Condition | Severity |
|-------|--------|-----------|----------|
| SNAT Connection Exhaustion | `natgw_active_connections` | > 80% of limit | Critical |
| Outbound Bandwidth High | `natgw_out_bps` | > 80% of spec | Warning |
| NAT Gateway Unhealthy | Health check | Status check fail | Critical |
| DNAT Port Unreachable | Port health check | Connection fail | Warning |

---

*This reference document is part of the ve-nat-ops skill.*

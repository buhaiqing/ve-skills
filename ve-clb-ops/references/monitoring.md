# Monitoring — Volcengine CLB

> **Purpose:** Monitoring guide for CLB resources on Volcengine.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [Cloud Monitor Metrics](#1-cloud-monitor-metrics)
2. [Listener Metrics](#2-listener-metrics)
3. [Backend Server Health](#3-backend-server-health)
4. [Alarm Configuration](#4-alarm-configuration)

---

## 1. Cloud Monitor Metrics

| Metric | Unit | Description |
|--------|------|-------------|
| `load_balancer_max_conn` | Count | Max concurrent connections |
| `load_balancer_new_conn` | Count/s | New connections per second |
| `load_balancer_active_conn` | Count | Active connections |
| `load_balancer_inactive_conn` | Count | Inactive connections |
| `load_balancer_lost_conn` | Count/s | Lost connections per second |
| `load_balancer_in_bps` | bps | Inbound bandwidth |
| `load_balancer_out_bps` | bps | Outbound bandwidth |
| `load_balancer_in_pps` | pps | Inbound packet rate |
| `load_balancer_out_pps` | pps | Outbound packet rate |

---

## 2. Listener Metrics

| Metric | Unit | Description |
|--------|------|-------------|
| `listener_max_conn` | Count | Max concurrent connections per listener |
| `listener_new_conn` | Count/s | New connections per listener |
| `listener_active_conn` | Count | Active connections per listener |
| `listener_healthy_rs_count` | Count | Healthy backend servers |
| `listener_unhealthy_rs_count` | Count | Unhealthy backend servers |

---

## 3. Backend Server Health

### Health Check Monitoring

Monitor the ratio of healthy to unhealthy backends:

```
Healthy Ratio = healthy_rs_count / (healthy_rs_count + unhealthy_rs_count)
```

| Healthy Ratio | Concern Level |
|---------------|--------------|
| 100% | Normal |
| 50–99% | Warning — investigate failures |
| < 50% | Critical — CLB may not serve traffic |

---

## 4. Alarm Configuration

### Recommended Alarms

| Alarm | Metric | Condition | Severity |
|-------|--------|-----------|----------|
| All Backends Unhealthy | `listener_unhealthy_rs_count` = total backends | All backends failed | Critical |
| Partial Backend Failures | `listener_unhealthy_rs_count` > 0 AND < total | Partial failures | Warning |
| Connection Saturation | `load_balancer_max_conn` | > 80% of limit | Warning |
| Lost Connections | `load_balancer_lost_conn` | > 100/s for 5min | Warning |
| CLB Inactive | CLB status | Not `active` | Critical |

> **Note:** For single-backend CLBs, any unhealthy count = all backends down. Set `listener_unhealthy_rs_count > 0` to Critical for single-backend configurations.

---

*This reference document is part of the ve-clb-ops skill.*

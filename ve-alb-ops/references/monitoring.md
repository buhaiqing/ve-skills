# Monitoring — Volcengine ALB (应用型负载均衡)

> **Purpose:** Monitoring guide for ALB resources on Volcengine.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-31

---

## Table of Contents

1. [Cloud Monitor Metrics](#1-cloud-monitor-metrics)
2. [Listener Metrics](#2-listener-metrics)
3. [Server Group Metrics](#3-server-group-metrics)
4. [Alarm Configuration](#4-alarm-configuration)

---

## 1. Cloud Monitor Metrics

| Metric | Unit | Description |
|--------|------|-------------|
| `alb_active_conn` | Count | Active concurrent connections |
| `alb_new_conn` | Count/s | New connections per second |
| `alb_max_conn` | Count | Max concurrent connections |
| `alb_in_bps` | bps | Inbound bandwidth |
| `alb_out_bps` | bps | Outbound bandwidth |
| `alb_in_pps` | pps | Inbound packet rate |
| `alb_out_pps` | pps | Outbound packet rate |
| `alb_lost_conn` | Count/s | Lost connections per second |
| `alb_http_2xx` | Count/s | HTTP 2xx responses per second |
| `alb_http_3xx` | Count/s | HTTP 3xx responses per second |
| `alb_http_4xx` | Count/s | HTTP 4xx responses per second |
| `alb_http_5xx` | Count/s | HTTP 5xx responses per second |
| `alb_http_502` | Count/s | HTTP 502 errors per second (upstream issues) |
| `alb_http_503` | Count/s | HTTP 503 errors per second (service unavailable) |
| `alb_http_504` | Count/s | HTTP 504 errors per second (upstream timeout) |

---

## 2. Listener Metrics

| Metric | Unit | Description |
|--------|------|-------------|
| `listener_active_conn` | Count | Active connections per listener |
| `listener_new_conn` | Count/s | New connections per listener |
| `listener_http_2xx` | Count/s | HTTP 2xx per listener |
| `listener_http_4xx` | Count/s | HTTP 4xx per listener (client errors) |
| `listener_http_5xx` | Count/s | HTTP 5xx per listener (server errors) |
| `listener_request_time_avg` | ms | Average request processing time |
| `listener_request_time_p99` | ms | P99 request processing time |
| `listener_upstream_rt_avg` | ms | Average upstream response time |

---

## 3. Server Group Metrics

| Metric | Unit | Description |
|--------|------|-------------|
| `server_group_healthy_count` | Count | Healthy backend servers |
| `server_group_unhealthy_count` | Count | Unhealthy backend servers |
| `server_group_in_bps` | bps | Inbound bandwidth per server group |
| `server_group_out_bps` | bps | Outbound bandwidth per server group |
| `server_group_active_conn` | Count | Active connections per server group |

### Health Ratio Calculation

```
Health Ratio = healthy_count / (healthy_count + unhealthy_count)
```

| Health Ratio | Status | Action |
|-------------|--------|--------|
| 100% | Normal | None |
| 50–99% | Warning | Investigate failing backends |
| < 50% | Critical | Immediate investigation required |
| 0% | Down | Service may be unavailable |

---

## 4. Alarm Configuration

### Recommended Alarms

| Alarm Name | Metric | Condition | Severity |
|-----------|--------|-----------|----------|
| All Backends Unhealthy | `server_group_healthy_count` = 0 | Zero healthy backends | Critical |
| High 5xx Rate | `alb_http_5xx` | > 5/s for 5 minutes | Critical |
| High 4xx Rate | `alb_http_4xx` | > 100/s for 5 minutes | Warning |
| Connection Saturation | `alb_max_conn` | > 80% of limit | Warning |
| Upstream Latency High | `listener_upstream_rt_avg` | > 5000ms for 5 minutes | Warning |
| ALB Inactive | ALB status != `active` | Not active for 1 minute | Critical |
| Lost Connections | `alb_lost_conn` | > 100/s for 5 minutes | Warning |

### Alarm Setup via CLI

```bash
# Note: Alarm configuration is done through Cloud Monitor (CMS), not directly via ve CLI
# ve cms CreateAlarmRule ...
```

---

*This reference document is part of the ve-alb-ops skill.*

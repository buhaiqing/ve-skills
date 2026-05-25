# Monitoring — Volcengine VPC

> **Purpose:** Monitoring and observability guide for VPC resources on Volcengine. Covers Cloud Monitor metrics, alarm configuration, and logging for network health assessment.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [Cloud Monitor Overview](#1-cloud-monitor-overview)
2. [VPC Metrics](#2-vpc-metrics)
3. [Subnet Monitoring](#3-subnet-monitoring)
4. [Alarm Configuration](#4-alarm-configuration)
5. [Log Collection](#5-log-collection)
6. [Dashboard Recommendations](#6-dashboard-recommendations)

---

## 1. Cloud Monitor Overview

Volcengine Cloud Monitor (云监控) provides metrics and alarms for VPC resources. Metrics are published automatically at regular intervals.

### Accessing Metrics

- **Console:** https://console.volcengine.com/monitor
- **API:** Cloud Monitor API (`ve cms`)
- **Dashboard:** Built-in VPC monitoring dashboards available in the VPC console

### Data Retention

| Metric Granularity | Retention Period |
|-------------------|-----------------|
| 1-minute | 7 days |
| 5-minute | 31 days |
| 1-hour | 90 days |

---

## 2. VPC Metrics

### Key Network Metrics

| Metric Name | Unit | Description | Typical Threshold |
|-------------|------|-------------|-------------------|
| `NetworkIn` | Bytes/s | Inbound traffic rate | Alert on sustained > 80% of bandwidth |
| `NetworkOut` | Bytes/s | Outbound traffic rate | Alert on sustained > 80% of bandwidth |
| `NetworkInPackets` | Packets/s | Inbound packet rate | Alert on unusual spikes |
| `NetworkOutPackets` | Packets/s | Outbound packet rate | Alert on unusual spikes |

### Query VPC Metrics via CLI

```bash
# Query VPC metrics (requires CMS/Cloud Monitor CLI)
ve cms DescribeAlarmHistory \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --Namespace "vpc" \
  --MetricName "NetworkIn" \
  --ResourceId "{{user.vpc_id}}"
```

---

## 3. Subnet Monitoring

### IP Address Utilization

The most critical subnet metric is IP address exhaustion.

**Key Indicators:**

| Metric | Concern Level | Action |
|--------|--------------|--------|
| Available IPs > 50% | Normal | Monitor |
| Available IPs 20–50% | Warning | Plan expansion |
| Available IPs < 20% | Critical | Create additional subnet |

**Check IP Utilization:**

```bash
ve vpc DescribeSubnets --Region "$VOLCENGINE_REGION" --VpcId "$VPC_ID" \
  | jq -r '.Result.Subnets[] |
    "\(.SubnetName)\t\(.TotalIpAddressCount)\t\(.AvailableIpAddressCount)\t\((.AvailableIpAddressCount / .TotalIpAddressCount * 100) | floor)% available")'
```

### Subnet Status Monitoring

Watch for status changes from `Available` to `Pending`:

```bash
ve vpc DescribeSubnets --Region "$VOLCENGINE_REGION" --VpcId "$VPC_ID" \
  | jq -r '.Result.Subnets[] |
    select(.Status != "Available") |
    "\(.SubnetId) is in state \(.Status)"'
```

---

## 4. Alarm Configuration

### Recommended Alarms

| Alarm | Metric | Condition | Severity |
|-------|--------|-----------|----------|
| IP Exhaustion | `AvailableIpAddressCount` | < 10 IPs | Critical |
| Subnet Deletion | Resource lifecycle | DeleteSubnet attempt | Warning |
| VPC Quota Approaching | VPC count | > 8 of 10 | Warning |
| Network Traffic Spike | `NetworkIn` / `NetworkOut` | > 2x baseline for 5min | Warning |

### Create Alarm via CLI

```bash
# Create alarm for subnet IP availability
ve cms CreateAlarm \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --MetricName "AvailableIpAddressCount" \
  --Namespace "vpc_subnet" \
  --ComparisonOperator "LessThan" \
  --Threshold "10" \
  --EvaluationCount 3 \
  --Period 300 \
  --ContactGroup "{{user.contact_group}}"
```

---

## 5. Log Collection

### VPC Flow Logs

Volcengine supports VPC Flow Logs for network traffic analysis. Flow logs capture information about IP traffic going to and from network interfaces in a VPC.

**Use Cases:**
- Security analysis and threat detection
- Network troubleshooting
- Traffic pattern optimization
- Compliance and audit

**Enablement (via console or API):**
1. Navigate to VPC console → Flow Logs
2. Select the VPC or subnet
3. Choose a log store (in Log Service / TLS)
4. Set filter conditions (Accept/Reject/All traffic)

### Log Service Integration

```bash
# Example: Query flow logs via Log Service CLI
ve tls SearchLogs \
  --TopicId "{{user.flow_log_topic_id}}" \
  --Query "*" \
  --Limit 100
```

---

## 6. Dashboard Recommendations

### Essential VPC Dashboard Panels

1. **VPC Resource Inventory**
   - Total VPCs per region
   - Subnets per VPC
   - Route tables per VPC

2. **IP Address Utilization**
   - Available IPs per subnet (bar chart)
   - Subnets below 20% threshold (highlighted)

3. **Network Traffic**
   - Inbound/outbound traffic rate (time series)
   - Packet rate (time series)
   - Top talkers by subnet

4. **Resource Health**
   - VPC status overview
   - Route table configuration changes
   - Failed API operations (error rate)

### Dashboard Configuration

```bash
# List available dashboards
ve cms ListDashboards --Region "{{env.VOLCENGINE_REGION}}"

# Create custom VPC dashboard (via API or console)
# Panels should reference metric namespace: vpc
# Dimensions: VpcId, SubnetId, Region
```

---

*This reference document is part of the ve-vpc-ops skill. For integration with ECS and other services, see [integration.md](integration.md).*

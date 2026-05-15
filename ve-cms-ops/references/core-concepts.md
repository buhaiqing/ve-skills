# Core Concepts — Volcengine Cloud Monitor (CMS)

## Architecture

CMS provides centralized monitoring and alerting for all Volcengine cloud services:

```
┌────────────────────────────────────────────────────┐
│               Cloud Monitor (CMS)                   │
│  ┌─────────────┐  ┌──────────────┐  ┌───────────┐  │
│  │ Metric Data │  │ Alarm Strategy│  │  Events   │  │
│  │ (TimeSeries)│  │ (Rules)       │  │ (System)  │  │
│  └──────┬──────┘  └──────┬───────┘  └─────┬─────┘  │
└─────────┼───────────────┼────────────────┼────────┘
          │               │                │
   ┌──────┴──────┐  ┌────┴───────┐  ┌─────┴──────┐
   │ ECS / RDS / │  │ Alarms via │  │ System     │
   │ Redis / VKE /│  │ SMS/Email/ │  │ Events via │
   │ TOS / SLB   │  │ Webhook/IM │  │ Notification│
   └─────────────┘  └────────────┘  └────────────┘
```

## Metric Data Model

Metrics follow a hierarchical structure:

```
Namespace → MetricName → Dimensions → DataPoints
  (Service)   (Metric)    (Resource)   (Time Series)
```

| Level | Description | Example |
|-------|-------------|---------|
| `Namespace` | Service namespace | `Volcengine_ECS` |
| `MetricName` | Specific metric | `CpuUtilization` |
| `Dimensions` | Resource filters | `[{"InstanceId":"i-xxx"}]` |
| `DataPoints` | Time-series values | `[{Timestamp: ..., Value: 75.5}]` |

## Alarm Strategy Lifecycle

```
Created → Disabled → Enabled → ALARM State → Notify → Auto-escalate
                                    ↓
                              DATA_INSUFFICIENT
                                    ↓
                                 Disabled
```

### Alarm States

| State | Description |
|-------|-------------|
| `OK` | Metric is within normal threshold |
| `ALARM` | Metric exceeds threshold |
| `DATA_INSUFFICIENT` | No data received within evaluation window |

## Alarm Escalation Rules

Alarms support multi-level escalation:

| Level | Trigger | Typical Action |
|-------|---------|----------------|
| Info | Metric crosses low threshold | Log, channel notification |
| Warn | Metric crosses medium threshold | Email, team notification |
| Critical | Metric crosses high threshold | SMS, on-call, webhook |

## Notification Channels

| Channel | Description | Config |
|---------|-------------|--------|
| SMS | Text message | Phone number in contact group |
| Email | Email notification | Email address in contact group |
| Webhook | HTTP POST to URL | Custom URL in contact group |
| IM | Integrated messaging (Feishu, DingTalk) | Webhook URL |

## Supported Services (Namespaces)

| Service | Namespace | Key Metrics |
|---------|-----------|-------------|
| ECS | `Volcengine_ECS` | CpuUtilization, MemoryUtilization, DiskUtilization |
| RDS MySQL | `Volcengine_RDS_MySQL` | CpuUtilization, Connections, QPS |
| Redis | `Volcengine_Redis` | CpuUtilization, MemoryUsage, Connections |
| VKE | `Volcengine_VKE` | NodeCpuUtilization, PodCpuUtilization |
| TOS | `Volcengine_TOS` | RequestCount, 4xxErrorRate, BandwidthIn |
| SLB | `Volcengine_SLB` | ActiveConnection, QPS, 5xxErrorRate |
| VPC | `Volcengine_VPC` | BandwidthIn, BandwidthOut, DropRate |
| ELK/ES | `Volcengine_Elasticsearch` | CpuUtilization, JvmMemoryUsage |

## Resource Limits (Defaults)

| Resource | Default Limit |
|----------|---------------|
| Alarm rules per account | 200 per namespace |
| Alarm templates per account | 50 |
| Contact groups per account | 20 |
| Data points per API call | 5000 |

## AIOps Diagnosis Decision Tree

```
[Alarm Triggered]
    │
    ├── Single-service alarm?
    │   ├── ECS → ve-ecs-ops for detailed diagnosis
    │   ├── RDS → ve-rds-ops for query/slow log analysis
    │   ├── Redis → ve-redis-ops for memory/connection issues
    │   └── Other → query metrics via ve-cms-ops
    │
    ├── Multi-service correlation?
    │   ├── Network + Compute spikes → check upstream dependency
    │   ├── DB + App latency → correlation via timestamps
    │   └── Storage + App errors → check disk I/O patterns
    │
    └── Alarm storm (>10 alarms in 5 min)?
        ├── Correlate by resource group
        ├── Identify root alarm (earliest or highest severity)
        └── Suppress duplicates; focus on root cause
```

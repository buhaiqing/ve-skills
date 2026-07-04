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

See escalation levels in [SKILL.md](../SKILL.md#operation-createalarm--create-an-alarm-strategy).

Alarms support Info → Warn → Critical multi-level escalation.

## Notification Channels

| Channel | Config |
|---------|--------|
| SMS | Phone in contact group |
| Email | Address in contact group |
| Webhook | Custom URL |
| IM | Feishu/DingTalk webhook |

## Supported Services (Namespaces)

See namespace table in [SKILL.md](../SKILL.md#namespace-convention).

Additional services: `Volcengine_VPC`, `Volcengine_Elasticsearch`.

## Resource Limits

Quota can be queried via `ve`:
```bash
ve metrics DescribeMetricRuleList --PageSize 1   # check alarm rule count
```

Reference defaults (verify via API): alarm rules ~200/namespace, templates ~50, contact groups ~20, data points ~5000/call.

> **Specialist content** — see `references/advanced/aiops-diagnosis.md` for the full decision tree (alarm correlation, multi-service diagnosis, alarm storm suppression).

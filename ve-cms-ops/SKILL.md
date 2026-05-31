---
name: ve-cms-ops
description: >-
  Use when the user needs to query, configure, troubleshoot, or manage Volcengine
  (火山引擎) Cloud Monitor / CMS (云监控) — metric data queries, alarm strategies,
  alarm templates, resource groups, monitoring dashboards, event monitoring,
  and notification groups. User mentions CMS, 云监控, Cloud Monitor, alarm,
  告警, 监控, monitoring data, or describes monitoring/alerting scenarios
  (e.g., querying CPU metrics, creating alarm rules for ECS, setting up
  notification channels) even without naming the product directly. Not for
  product-specific monitoring that has its own ops skill (e.g., ECS metrics
  via ve-ecs-ops).
license: MIT
compatibility: >-
  Official Volcengine CLI (`ve`), Go SDK `github.com/volcengine/volc-sdk-golang`,
  valid API credentials, network access to Volcengine endpoints
  (open.volcengineapi.com).
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-05-15"
  runtime: Harness AI Agent, Claude Code, Cursor, or compatible Agent runtimes
  go_version_minimum: "1.14"
  go_version_jit: "1.21+"
  api_profile: "Cloud Monitor API 2018-03-14 (https://www.volcengine.com/docs/6408/78941)"
  cli_applicability: dual-path
  cli_support_evidence: >-
    Confirmed via `ve metrics --help` — Cloud Monitor APIs are accessible via ve CLI.
    SDK service: `github.com/volcengine/volc-sdk-golang`
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

> This skill follows the [Agent Skill OpenSpec](https://agentskills.io/specification).

# Volcengine Cloud Monitor Operations Skill

## Overview

Cloud Monitor (CMS, 云监控) on Volcengine (火山引擎) provides centralized monitoring and alerting for all cloud services. This skill is an **operational runbook** for agents: metric data queries, alarm strategy management (create/enable/disable/delete), alarm templates, event monitoring, and notification group configuration. **Dual-path execution**: `ve` CLI for API calls, JIT Go SDK fallback.

### CLI applicability

- **`cli_applicability: dual-path`:** Official `ve` CLI supports Cloud Monitor operations via `ve metrics` commands. You **MUST** document **both** the SDK step **and** the `ve` CLI step for every operation.

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use conditions with precise triggers and delegation rules |
| 2 | **Structured I/O** | `{{env.VOLCENGINE_*}}` (env), `{{user.*}}` (interactive), `{{output.*}}` (API response) |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | Error taxonomy with ≥ 10 CMS-specific codes; HALT vs retry per type |
| 5 | **Absolute Single Responsibility** | One product (CMS), one primary resource (Alarm/Metric); cross-product delegation documented |

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- User mentions "Volcengine CMS", "火山引擎 云监控", "Cloud Monitor", or "云监控"
- Task involves **metric queries**: GetMetricData (query time-series data for any cloud service)
- Task involves **alarm strategies**: create, enable, disable, edit, delete alarm rules
- Task involves **alarm templates**: create, apply, delete alarm templates
- Task involves **event monitoring**: query system events, configure event subscriptions
- Task involves **notification**: manage contact groups, webhooks, notification channels
- Task involves **monitoring dashboards**: create, configure custom dashboards

### SHOULD NOT Use This Skill When

- Task is about a specific product's monitoring (e.g., ECS CPU) → handle via the product ops skill (`ve-ecs-ops`), which delegates here for alarm configuration
- Task is about log analysis → delegate to Loki/logging ops skill (if not available, use Loki API directly)
- Task is purely billing → delegate to billing ops

### Delegation Rules

- CMS alarm strategies depend on IAM permissions for notifications → reference `ve-iam-ops` (if not available, use Volcengine IAM API directly)
- Cross-service monitoring queries may need context from product-specific ops skills

## Variable Convention (Agent-Readable)

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_REGION}}` | From runtime environment | Use documented default only if skill explicitly allows |
| `{{user.namespace}}` | Cloud service metric namespace (e.g., Volcengine_ECS) | Ask once; reuse |
| `{{user.metric_name}}` | Metric name (e.g., CpuUsage) | Ask once; reuse |
| `{{user.resource_id}}` | Resource dimension (e.g., instance ID) | Ask once; reuse |
| `{{user.alarm_name}}` | Alarm strategy name | Ask once; reuse |
| `{{user.contact_group}}` | Notification group name | Ask once; reuse |
| `{{output.request_id}}` | Request identifier for tracing | Parse from response |
| `{{output.datapoints}}` | Time-series metric data points | Parse from GetMetricData response |
| `{{user.threshold}}` | Alarm threshold value | e.g., `90` for CPU % |
| `{{user.period}}` | Data granularity in seconds | `60`, `300`, `3600` |
| `{{user.dimension_key}}` | Dimension key for resource filter | e.g., `InstanceId` |
| `{{user.start_time_ms}}` | Query start time (Unix ms) | e.g., `$(($(date +%s)-3600))000` |
| `{{user.end_time_ms}}` | Query end time (Unix ms) | e.g., `$(date +%s)000` |
| `{{user.rule_id}}` | Alarm rule ID | Format `rule-xxxxxxxxx` |
| `{{user.template_name}}` | Alarm template name | Ask once; reuse |

> **`{{env.*}}` MUST NOT** be collected from the user. **`{{user.*}}`** MUST be collected interactively when missing.

> **Security Warning (Credential Masking):** NEVER log, print, or expose `VOLCENGINE_SECRET_KEY`. Verify existence only with `test -n "$VOLCENGINE_SECRET_KEY"`.

## API and Response Conventions (Agent-Readable)

- **Cloud Monitor OpenAPI (2018-03-14)** is canonical for all API fields and response shapes
- **Endpoint:** `open.volcengineapi.com` (or `monitor.volcengineapi.com`)
- **Service code:** `metrics` (or use the general `ve` CLI endpoint)
- **Errors:** Responses with `Error` object containing `Code` and `Message` fields

### Key Response Fields

| Operation | JSON Path | Type | Description |
|-----------|-----------|------|-------------|
| GetMetricData | `$.MetricName` | string | Metric name |
| GetMetricData | `$.Datapoints` | array | Data point list |
| GetMetricData | `$.Datapoints[].Value` | float | Metric value |
| GetMetricData | `$.Datapoints[].Timestamp` | integer | Unix timestamp (ms) |
| GetMetricData | `$.Period` | integer | Data granularity (seconds) |
| CreateAlarm | `$.Id` | string | Alarm rule ID |
| ListAlarms | `$.Alarms.Alarm[]` | array | Alarm rule list |
| ListAlarms | `$.Total` | integer | Total alarm count |

### Namespace Convention

Cloud Monitor uses service namespaces to scope metrics:

| Service | Namespace |
|---------|-----------|
| ECS | `Volcengine_ECS` |
| RDS MySQL | `Volcengine_RDS_MySQL` |
| Redis | `Volcengine_Redis` |
| VKE | `Volcengine_VKE` |
| TOS | `Volcengine_TOS` |
| SLB | `Volcengine_SLB` |
| VPC | `Volcengine_VPC` |

> Query all available namespaces and metrics via the metrics meta API.

## Quick Start

### What This Skill Does
This skill enables you to query monitoring metrics, create and manage alarm strategies, and configure event monitoring for all Volcengine cloud services using the `ve` CLI (primary) or JIT Go SDK (fallback).

### Prerequisites
- [ ] `ve` CLI installed
- [ ] Credentials configured: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`
- [ ] Region set: `VOLCENGINE_REGION`

### Verify Setup
```bash
ve version
```

### Your First Command
```bash
# Query ECS CPU usage for the last hour
ve metrics GetMetricData --Namespace Volcengine_ECS --MetricName CpuUtilization --Dimensions '[{"InstanceId":"i-xxxxx"}]' --StartTime $(($(date +%s)-3600))000 --EndTime $(date +%s)000 --Period 60
```

## Capabilities at a Glance

| Operation | Description | Complexity | Risk Level |
|-----------|-------------|------------|------------|
| GetMetricData | Query time-series metric data | Medium | None |
| ListMetrics | List available metrics for a namespace | Low | None |
| CreateAlarm | Create an alarm strategy | Medium | Medium |
| EnableAlarm | Enable an alarm strategy | Low | Low |
| DisableAlarm | Disable an alarm strategy | Low | Low |
| DeleteAlarm | Delete an alarm strategy | Low | **High** |
| ListAlarms | List all alarm rules | Low | None |
| CreateAlarmTemplate | Create an alarm template | Medium | Medium |
| ApplyAlarmTemplate | Apply template to resources | Medium | Medium |
| ListEvents | Query system events | Low | None |
| DescribeContactGroups | List notification groups | Low | None |
| CreateContactGroup | Create a notification group | Low | Low |

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.1 | 2026-05-28 | Enhanced error handling taxonomy with 12 CMS-specific error codes; added retry strategy matrix |
| 1.0.0 | 2026-05-15 | Initial release with core capabilities: GetMetricData (time-series queries), ListMetrics (namespace metrics), CreateAlarm/EnableAlarm/DisableAlarm/DeleteAlarm (alarm lifecycle), CreateAlarmTemplate/ApplyAlarmTemplate (template management), ListEvents (system events), DescribeContactGroups/CreateContactGroup (notification groups); includes CLI dual-path execution (ve CLI + Go SDK fallback), Pre-flight→Execute→Validate→Recover workflow, namespace convention guide (ECS/RDS/Redis/VKE/TOS/SLB/VPC), security credential masking guidelines |

## Testing Guide

### Unit Testing Strategy

| Component | Test Approach | Coverage Target |
|-----------|---------------|-----------------|
| Credential validation | Mock environment vars | 100% |
| Namespace resolution | Test with valid/invalid namespaces | 100% |
| Parameter parsing | Test JSON format for Dimensions | 100% |
| Error code mapping | Test all 12 CMS error codes | 100% |

### Integration Testing

```bash
# Test 1: Verify credentials are configured
export VOLCENGINE_ACCESS_KEY="test_key"
export VOLCENGINE_SECRET_KEY="test_secret"
export VOLCENGINE_REGION="cn-north-1"

# Test 2: Query ECS CPU metrics (should return data or specific error)
ve metrics GetMetricData \
  --Namespace Volcengine_ECS \
  --MetricName CpuUtilization \
  --Dimensions '[{"InstanceId":"i-xxxxx"}]' \
  --StartTime $(($(date +%s)-3600)) P3 000 \
  --EndTime $(date +%s P4 )000 \
  --Period 60

# Test 3: List alarm rules
ve metrics DescribeMetricRuleList

# Test 4: Create and delete test alarm (with cleanup)
ve metrics PutResourceMetricRule \
  --RuleName "test-alarm-$(date +%s)" \
  --Namespace Volcengine_ECS \
  --MetricName CpuUtilization \
  --Resources '[{"Dimensions":[{"InstanceId":"i-xxxxx"}]}]' \
  --AlertState Critical \
  --ComparisonOperator GreaterThanThreshold \
  --Statistics Average \
  --Threshold 99 \
  --Times 1 \
  --Period 60
```

### Test Scenarios

| Scenario | Expected Result |
|----------|-----------------|
| Invalid credentials | `Unauthorized` error with clear message |
| Non-existent namespace | `NoSuchMetric` error |
| Invalid dimensions format | `InvalidParameter` error |
| Throttling (rate limit) | Retry with exponential backoff (2s, 4s, 8s) |
| Create alarm with duplicate name | `DuplicateRuleName` error |
| Delete non-existent alarm | Success (idempotent) |

### Test Environment Setup

```bash
# Create dedicated test namespace/project
export VOLCENGINE_TEST_PROJECT="cms-ops-test"

# Use separate notification group for tests
export CMS_TEST_CONTACT_GROUP="test-alerts"
```

### Smoke Tests

```bash
# Quick validation of setup
ve version  # Should return version info

# List metrics to verify API connectivity
ve metrics ListMetrics --Namespace Volcengine_ECS

# Verify notification groups
ve metrics DescribeContactGroups
```

---

## Execution Flows (Agent-Readable)

Every operation: **Pre-flight → Execute → Validate → Recover**.

### Operation: GetMetricData — Query Time-Series Metric Data

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | `test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY"` | Both set | HALT |
| Namespace valid | Verify namespace exists (e.g., `Volcengine_ECS`) | Valid | Suggest: `ve metrics ListMetrics --Namespace {{user.namespace}}` |
| Resource exists | Verify resource ID is active | Resource found | HALT; fix resource ID |

#### Execution

```bash
# Query a single metric for a specific resource
ve metrics GetMetricData \
  --Namespace "{{user.namespace}}" \
  --MetricName "{{user.metric_name}}" \
  --Dimensions '[{"{{user.dimension_key}}":"{{user.resource_id}}"}]' \
  --StartTime "{{user.start_time_ms}}" \
  --EndTime "{{user.end_time_ms}}" \
  --Period "{{user.period}}"
```

**Common parameters:**

| Parameter | Description | Example |
|-----------|-------------|---------|
| `Namespace` | Service namespace | `Volcengine_ECS` |
| `MetricName` | Metric name | `CpuUtilization` |
| `Dimensions` | Resource filter as JSON | `[{"InstanceId":"i-xxx"}]` |
| `StartTime` / `EndTime` | Unix timestamp (milliseconds) | `1715000000000` |
| `Period` | Data granularity (seconds) | `60`, `300`, `3600` |

#### Go SDK (Fallback)

```go
package main

import (
	"fmt"
	"os"

	"github.com/volcengine/volc-sdk-golang/base"
)

func main() {
	client := base.NewClient(
		os.Getenv("VOLCENGINE_ACCESS_KEY"),
		os.Getenv("VOLCENGINE_SECRET_KEY"),
	)
	client.SetHost("open.volcengineapi.com")

	params := map[string]string{
		"Action":    "GetMetricData",
		"Version":   "2018-03-14",
		"Namespace": os.Getenv("NAMESPACE"),
		"MetricName": os.Getenv("METRIC_NAME"),
	}

	resp, err := client.Get("metrics_v2", params)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(resp))
}
```

#### Validation

1. Parse `$.Datapoints` — verify non-empty array
2. Check `$.Period` matches expected granularity
3. Report min/avg/max values across the queried range

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `InvalidParameter` | Verify namespace, metric name, and dimensions format |
| `NoSuchMetric` | HALT; metric not found in this namespace |
| `Throttling` | Retry with exponential backoff (2s, 4s, 8s) |
| `InternalError` | Retry up to 3 times; then HALT with request ID |
| `Unauthorized` | HALT; ensure VMSReadOnlyAccess IAM policy is attached |

---

### Operation: CreateAlarm — Create an Alarm Strategy

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Metrics available | GetMetricData returns data for the target metric | Data exists | HALT; metric has no data yet |
| Contact group exists | `ve metrics DescribeContactGroups` | Group found | HALT; create contact group first |
| Threshold reasonable | Metric baseline is within expected range | Normal | Warn if threshold is unusual |

#### Execution

```bash
# Create a threshold-triggered alarm
ve metrics PutResourceMetricRule \
  --RuleName "{{user.alarm_name}}" \
  --Namespace "{{user.namespace}}" \
  --MetricName "{{user.metric_name}}" \
  --Resources '[{"Dimensions":[{"{{user.dimension_key}}":"{{user.resource_id}}"}]}]' \
  --AlertState Critical \
  --ComparisonOperator GreaterThanThreshold \
  --Statistics Average \
  --Threshold "{{user.threshold}}" \
  --Times 3 \
  --Period "{{user.period}}" \
  --NotifyType "{{user.contact_group}}"
```

**Alarm escalation levels:**

| Level | Statistics | ComparisonOperator | Threshold | Times (consecutive) |
|-------|-----------|-------------------|-----------|---------------------|
| Critical | Average | GreaterThanThreshold | 90 | 3 |
| Warn | Average | GreaterThanThreshold | 80 | 3 |
| Info | Average | GreaterThanThreshold | 70 | 3 |

#### Validation

```bash
# Verify alarm was created
ve metrics DescribeMetricRuleList --RuleName "{{user.alarm_name}}"
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `ContactGroupNotFound` | HALT; create contact group first via CreateContactGroup |
| `InvalidParameters` | Verify all required fields; check JSON format for Resources |
| `DuplicateRuleName` | HALT; rule name must be unique — use a different name |
| `QuotaExceeded` | HALT; alarm rule limit reached per namespace |

---

### Operation: ListAlarms — List Alarm Rules

#### Execution

```bash
# List all alarm rules
ve metrics DescribeMetricRuleList

# Filter by namespace
ve metrics DescribeMetricRuleList --Namespace Volcengine_ECS

# Filter by rule name pattern
ve metrics DescribeMetricRuleList --RuleName "cpu"

# Paginated
ve metrics DescribeMetricRuleList --PageNumber 1 --PageSize 50
```

#### Validation

Parse `$.Alarms.Alarm[]` for:
- `RuleId`: Alarm rule ID
- `RuleName`: Alarm name
- `MetricName`: Monitored metric
- `Namespace`: Service namespace
- `EnableState`: `true` = active, `false` = disabled
- `AlertState`: Current state (`OK`, `ALARM`, `DATA_INSUFFICIENT`)

---

### Operation: EnableAlarm / DisableAlarm — Toggle Alarm

#### Execution

```bash
# Enable an alarm
ve metrics EnableMetricRule --RuleIds '["{{user.rule_id}}"]'

# Disable an alarm
ve metrics DisableMetricRule --RuleIds '["{{user.rule_id}}"]'
```

#### Validation

Re-check `DescribeMetricRuleList --RuleIds '["{{user.rule_id}}"]'` and verify `EnableState` changed.

---

### Operation: DeleteAlarm — Delete Alarm Rule

#### Pre-flight (Safety Gate)

- **MUST** confirm: deleting alarm rule `{{user.alarm_name}}` (ID: `{{user.rule_id}}`)
- **MUST NOT** proceed without clear user assent
- **MUST** warn: this stops all monitoring and notifications for this rule

#### Execution

```bash
ve metrics DeleteMetricRules --RuleIds '["{{user.rule_id}}"]'
```

#### Validation

Verify rule no longer appears in `DescribeMetricRuleList`.

---

### Operation: CreateAlarmTemplate — Create and Apply Alarm Template

#### Execution

```bash
# Create alarm template
ve metrics PutMetricRuleTemplate \
  --TemplateName "{{user.template_name}}" \
  --Namespace "{{user.namespace}}" \
  --Rules '[{"MetricName":"CpuUtilization","Threshold":90,"ComparisonOperator":">=","Statistics":"Average","Times":3,"NotifyType":"DefaultGroup"}]'

# Apply template to resources
ve metrics ApplyMetricRuleTemplate \
  --TemplateName "{{user.template_name}}" \
  --Namespace "{{user.namespace}}" \
  --Resources '[{"Dimensions":[{"InstanceId":"i-xxx"}]}]'
```

#### Validation

```bash
ve metrics DescribeMetricRuleTemplate --TemplateName "{{user.template_name}}"
```

---

### Operation: DescribeContactGroups — List Notification Groups

#### Execution

```bash
ve metrics DescribeContactGroups
```

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `$.ContactGroups.ContactGroup[].Name` | string | Group name |
| `$.ContactGroups.ContactGroup[].Webhook` | string | Webhook URL |
| `$.ContactGroups.ContactGroup[].Type` | string | Group type |

---

## Reference Directory

- [Core Concepts](references/core-concepts.md)
- [API & SDK Usage](references/api-sdk-usage.md)
- [CLI Usage](references/cli-usage.md)
- [Troubleshooting Guide](references/troubleshooting.md)
- [Monitoring](references/monitoring.md)
- [Integration](references/integration.md)
- [User Experience Specification](../../ve-skill-generator/references/user-experience-spec.md)
- [Execution Environment Setup](../../ve-skill-generator/references/execution-environment.md)
- [CLI Behavioral Reference](../../ve-skill-generator/references/cli-behavior.md)
- [Enhanced Self-Healing Framework](../../ve-skill-generator/references/enhanced-self-healing-framework.md)

## Operational Best Practices

- **Threshold tuning:** Start with conservative thresholds, adjust based on baseline metrics
- **Silence period:** Configure at least 5 minutes to prevent alert storms
- **Multi-level alerts:** Use Info → Warn → Critical escalation levels
- **Resource coverage:** Apply alarm templates to all production resources
- **Notification routing:** Route critical alerts to on-call, info to channels
- **No-data policy:** Configure action when metric data stops (e.g., treat as alarm)

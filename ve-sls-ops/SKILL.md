---
name: ve-sls-ops
description: >-
  Use when the user needs to analyze, optimize, or manage Volcengine (火山引擎)
  SLS (Simple Log Service / 日志服务) — log collection, indexing, storage lifecycle,
  LogShipper, alert rules, and cost optimization. User mentions SLS, 日志服务,
  Log Service, log collection, log indexing, log storage, LogShipper, or describes
  scenarios like analyzing log volume spikes, optimizing indexes, managing log TTL,
  configuring log delivery, or troubleshooting missing logs. Not for application
  logging code changes or log analysis queries that use existing tools.
license: MIT
compatibility: >-
  Official Volcengine CLI (`ve`, Go binary, no runtime), Go 1.14+ runtime
  (for JIT SDK fallback), valid API credentials, network access to Volcengine
  endpoints (open.volcengineapi.com).
metadata:
  author: volcengine
  version: "1.1.0"
  last_updated: "2026-06-04"
  runtime: Harness AI Agent, Claude Code, Cursor, or compatible Agent runtimes
  go_version_minimum: "1.14"
  go_jit_runtime_version: "1.21+"
  api_profile: "TLS API 2020-09-01"
  cli_applicability: dual-path
  cli_support_evidence: >-
  Confirmed via `ve tls --help` — TLS/SLS is supported by the ve CLI.
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

> This skill follows the [Agent Skill OpenSpec](https://agentskills.io/specification).

# Volcengine SLS (Simple Log Service) Operations Skill

## Overview

SLS (Simple Log Service / 日志服务, also known as TLS - Total Log Service) on Volcengine (火山引擎) provides log collection, storage, indexing, analysis, and delivery capabilities. This skill is an **operational runbook** for agents: log collection volume analysis, index optimization, storage cost optimization, LogShipper configuration audit, alert rule optimization, and common troubleshooting patterns. **Do not use the web console as the primary agent execution path.**

> **UX Compliance:** This skill follows the [User Experience Specification](../ve-skill-generator/references/user-experience-spec.md). All operations include onboarding guidance, minimal prompts, smart defaults, clear feedback, and user-friendly error handling.

### CLI applicability

- **`cli_applicability: dual-path`:** Official `ve` CLI supports TLS/SLS. You **MUST** document **both** the SDK step **and** the `ve` CLI step for every operation.

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use conditions with precise triggers and delegation rules |
| 2 | **Structured I/O** | `{{env.VOLCENGINE_*}}` (env), `{{user.*}}` (interactive), `{{output.*}}` (API response) |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | Error taxonomy with SLS-specific codes; HALT vs retry per type |
| 5 | **Absolute Single Responsibility** | One product (SLS/TLS); cross-product delegation documented |
| 6 | **FinOps Integration** | Storage cost analysis, index optimization, TTL tuning, log volume anomaly detection |
| 7 | **AIOps Integration** | Log pattern analysis, parsing failure detection, alert rule tuning |

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- User mentions "SLS", "TLS", "日志服务", "Log Service", "Simple Log Service"
- Task involves **Log Projects/Topics**: CreateProject, DescribeProjects, CreateTopic, DescribeTopics, DeleteTopic
- Task involves **Index configuration**: CreateIndex, ModifyIndex, DescribeIndex
- Task involves **LogShipper**: CreateShipper, DescribeShippers, ModifyShipper
- Task involves **Log collection**: machine groups, collectors, logtail configuration
- Task involves **log storage lifecycle**: TTL settings, hot/cold tiering, archive policies
- Task involves **cost optimization**: identifying high-volume topics, unused indexes, storage reduction
- Task involves **alert rules**: log-based alert creation, tuning, false positive reduction
- Task keywords: log volume, log storage, index optimization, LogShipper, log delivery, log parsing

### SHOULD NOT Use This Skill When

- Task is about **writing application log code** → delegate to app development
- Task is about **querying log content** (already configured) → use SLS console/CLI directly
- Task is about **ECS instance management** → delegate to: `ve-ecs-ops`
- Task is about **VKE/Kubernetes logs** → delegate to: `ve-vke-ops` (for cluster config)
- Task is purely billing/account management → delegate to billing ops

### Delegation Rules

- Log collection agents run on ECS → verify ECS exists via `ve-ecs-ops`
- LogShipper delivers to TOS → reference `ve-tos-ops` for destination config
- Log alerts may trigger notifications → reference alert management skills

## Variable Convention (Agent-Readable)

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_REGION}}` | From runtime environment | Use documented default only if skill explicitly allows |
| `{{user.region}}` | User-supplied region | Ask once; reuse; default from env |
| `{{user.project_id}}` | User-supplied log project ID | Ask once; reuse; format UUID |
| `{{user.project_name}}` | User-supplied log project name | Ask once; reuse |
| `{{user.topic_id}}` | User-supplied log topic ID | Ask once; reuse; format UUID |
| `{{user.topic_name}}` | User-supplied log topic name | Ask once; reuse |
| `{{user.ttl_days}}` | User-supplied TTL in days | Ask once; format integer |
| `{{user.shipper_id}}` | User-supplied shipper ID | Format UUID |
| `{{output.project_id}}` | From CreateProject response | Parse from `$.ProjectId` |
| `{{output.topic_id}}` | From CreateTopic response | Parse from `$.TopicId` |

> **Security Warning (Credential Masking — MANDATORY):** **NEVER** log, print, or expose `VOLCENGINE_SECRET_KEY`.

## API and Response Conventions (Agent-Readable)

- **Volcengine TLS OpenAPI (2020-09-01)** is canonical for SLS APIs.
- **Endpoint:** `tls.volcengineapi.com` (default: `open.volcengineapi.com`)

### Key Response Fields

| Operation | JSON Path | Type | Description |
|-----------|-----------|------|-------------|
| CreateProject | `$.ProjectId` | string | Created project ID |
| DescribeProjects | `$.Projects` | array | Project list |
| DescribeProjects | `$.Projects[].ProjectId` | string | Project ID |
| DescribeProjects | `$.Projects[].ProjectName` | string | Project name |
| CreateTopic | `$.TopicId` | string | Created topic ID |
| DescribeTopics | `$.Topics` | array | Topic list |
| DescribeTopics | `$.Topics[].TopicId` | string | Topic ID |
| DescribeTopics | `$.Topics[].TopicName` | string | Topic name |
| DescribeTopics | `$.Topics[].Ttl` | integer | TTL in days |
| DescribeTopics | `$.Topics[].ShardCount` | integer | Shard count |
| DescribeTopics | `$.Topics[].StorageUsed` | integer | Storage used (bytes) |
| DescribeTopics | `$.Topics[].LogCount` | integer | Total log count |
| DescribeIndex | `$.IndexRules` | object | Index configuration |
| DescribeIndex | `$.IndexRules.FullText` | object | Full-text index config |
| DescribeIndex | `$.IndexRules.KeyValue` | object | Key-value index config |
| DescribeShippers | `$.Shippers` | array | LogShipper list |
| DescribeShippers | `$.Shippers[].ShipperName` | string | Shipper name |
| DescribeShippers | `$.Shippers[].DestinationType` | string | Destination type (TOS, etc.) |
| DescribeShippers | `$.Shippers[].Enable` | boolean | Whether shipper is enabled |

## Quick Start

### What This Skill Does
This skill enables you to analyze, optimize, and manage Volcengine SLS (日志服务) using the `ve` CLI (primary) or JIT Go SDK (fallback).

### Prerequisites
- [ ] `ve` CLI installed (or Go runtime for JIT fallback)
- [ ] Credentials configured: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`
- [ ] Region set: `VOLCENGINE_REGION`

### Verify Setup
```bash
ve tls DescribeProjects --Region {{env.VOLCENGINE_REGION}}
```

### Your First Command
```bash
# List all log projects
ve tls DescribeProjects --Region {{env.VOLCENGINE_REGION}}
```

### Next Steps
- [Core Concepts](references/core-concepts.md) — Understand SLS architecture
- [Common Operations](#execution-flows) — Create, manage, and collect and analyze logs
- [Troubleshooting](references/troubleshooting.md) — Fix common issues

## Capabilities at a Glance

| Operation | Description | Complexity | Risk Level |
|-----------|-------------|------------|------------|
| DescribeProjects | List log projects | Low | None |
| CreateProject | Create a log project | Low | Low |
| DeleteProject | Delete a log project | Low | **High** |
| DescribeTopics | List log topics | Low | None |
| CreateTopic | Create a log topic | Low | Low |
| DeleteTopic | Delete a log topic | Low | **High** |
| ModifyTopic | Modify topic TTL and settings | Low | Medium |
| DescribeIndex | Query index configuration | Low | None |
| CreateIndex | Enable full-text/key-value index | Medium | Low |
| ModifyIndex | Update index configuration | Medium | Low |
| DescribeShippers | List LogShippers | Low | None |
| CreateShipper | Create log delivery task | Medium | Low |
| ModifyShipper | Modify shipper configuration | Medium | Low |
| AnalyzeLogVolume | Analyze log volume and trends | Medium | None |
| OptimizeIndexes | Identify and remove unused indexes | High | Medium |
| AuditLogShipper | Check LogShipper health and config | Medium | None |
| DetectAnomalies | Detect log volume anomalies | Medium | None |
| TroubleshootMissingLogs | Diagnose missing log issues | High | None |

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-27 | Initial release with SLS lifecycle, cost optimization, troubleshooting |
| 1.1.0 | 2026-06-04 | GCL rollout: added ## Quality Gate (GCL), references/rubric.md, references/prompt-templates.md |

## Quality Gate (GCL)

> Optional tier. max_iter=5.

| Tier | Operations | Safety |
|---|---|---|
| **Destructive** | DeleteProject, DeleteTopic, DeleteShipper | 1.0 |
| **State-changing** | CreateProject, CreateTopic, CreateShipper, ModifyProject, ModifyTopic, ModifyShipper, CreateIndex, ModifyIndex | 1.0 |
| **Mutating** | — | ≥0.5 |
| **Read-only** | DescribeProjects, DescribeTopics, DescribeShippers, SearchLog, DescribeLogHistogram | ≥0 |

Safety: DeleteProject ALL topics+data LOST. SearchLog query MUST NOT contain credentials. VOLCENGINE_SECRET_KEY never.

### Cross-skill: Billing→ve-billing-ops

## Execution Flows (Agent-Readable)

### Operation: DescribeProjects — List Log Projects

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | `test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY"` | Both set | HALT |
| Region | Verify `{{user.region}}` or `{{env.VOLCENGINE_REGION}}` | Valid region | Suggest: `ve tls DescribeRegions` |
| CLI | `ve version` | Exit code 0 | Install ve CLI |

#### Execution

```bash
# List all log projects
ve tls DescribeProjects --Region "{{user.region}}"

# Filter by project name
ve tls DescribeProjects --Region "{{user.region}}" --ProjectName "{{user.project_name}}"
```

#### Validation

1. Check `$.Total` for total project count
2. Parse `$.Projects[]` for project details
3. Report project IDs, names, and descriptions

---

### Operation: DescribeTopics — List Log Topics

#### Execution

```bash
# List all topics in a project
ve tls DescribeTopics --Region "{{user.region}}" --ProjectId "{{user.project_id}}"

# Filter by topic name
ve tls DescribeTopics --Region "{{user.region}}" --ProjectId "{{user.project_id}}" --TopicName "{{user.topic_name}}"
```

---

### Operation: CreateTopic — Create Log Topic

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Project exists | DescribeProjects with project ID | Project found | HALT; create project first |
| Shard quota | Check shard quota per project | Sufficient | HALT; request quota increase |

#### Execution

```bash
ve tls CreateTopic \
  --Region "{{user.region}}" \
  --ProjectId "{{user.project_id}}" \
  --TopicName "{{user.topic_name}}" \
  --Ttl {{user.ttl_days}} \
  --ShardCount 1 \
  --Description "{{user.description}}"
```

---

### Operation: ModifyTopic — Modify Topic TTL

#### Execution

```bash
# Change TTL for a topic
ve tls ModifyTopic \
  --Region "{{user.region}}" \
  --ProjectId "{{user.project_id}}" \
  --TopicId "{{user.topic_id}}" \
  --Ttl {{user.ttl_days}}
```

#### TTL Recommendations

| Log Type | Recommended TTL | Reason |
|----------|----------------|--------|
| Application logs | 7-30 days | Short-term debugging |
| Audit logs | 90-180 days | Compliance requirements |
| Security logs | 180-365 days | Security investigation |
| Access logs | 7-14 days | Traffic analysis |
| Error logs | 30-90 days | Incident investigation |

---

### Operation: DeleteTopic — Delete Log Topic

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: irreversible delete
- **MUST** warn about data loss — all logs in topic will be deleted
- **MUST** check if LogShipper is configured (delivery may break)

```bash
# Check for active shippers
ve tls DescribeShippers --Region "{{user.region}}" --ProjectId "{{user.project_id}}" --TopicId "{{user.topic_id}}"
```

#### Execution

```bash
ve tls DeleteTopic --Region "{{user.region}}" --ProjectId "{{user.project_id}}" --TopicId "{{user.topic_id}}"
```

---

## Index Management

### Operation: DescribeIndex — Query Index Configuration

#### Execution

```bash
# Get index configuration for a topic
ve tls DescribeIndex --Region "{{user.region}}" --ProjectId "{{user.project_id}}" --TopicId "{{user.topic_id}}"
```

#### Output Analysis

```bash
# Show full-text index config
ve tls DescribeIndex --Region "{{user.region}}" --ProjectId "{{user.project_id}}" --TopicId "{{user.topic_id}}" | jq '.IndexRules.FullText'

# Show key-value index fields
ve tls DescribeIndex --Region "{{user.region}}" --ProjectId "{{user.project_id}}" --TopicId "{{user.topic_id}}" | jq '.IndexRules.KeyValue'
```

---

### Operation: CreateIndex — Enable Indexing

#### Execution

```bash
# Enable full-text index
ve tls CreateIndex \
  --Region "{{user.region}}" \
  --ProjectId "{{user.project_id}}" \
  --TopicId "{{user.topic_id}}" \
  --FullText '{"CaseSensitive": false, "Delimiter": " ,;\\n\\t"}'

# Enable key-value index for specific fields
ve tls CreateIndex \
  --Region "{{user.region}}" \
  --ProjectId "{{user.project_id}}" \
  --TopicId "{{user.topic_id}}" \
  --KeyValue '{"level": {"CaseSensitive": false, "Type": "text"}, "request_id": {"CaseSensitive": true, "Type": "text"}, "duration_ms": {"CaseSensitive": false, "Type": "float"}}'

# Enable both full-text and key-value
ve tls CreateIndex \
  --Region "{{user.region}}" \
  --ProjectId "{{user.project_id}}" \
  --TopicId "{{user.topic_id}}" \
  --FullText '{"CaseSensitive": false, "Delimiter": " ,;\\n\\t"}' \
  --KeyValue '{"level": {"Type": "text"}, "message": {"Type": "text"}}'
```

---

## FinOps Operations

### Operation: AnalyzeLogVolume — Analyze Log Volume and Cost Drivers

Identifies high-volume topics and cost drivers.

#### Execution

```bash
# Get all topics with storage usage
ve tls DescribeTopics --Region "{{user.region}}" --ProjectId "{{user.project_id}}" | jq '.Topics[] | {TopicId, TopicName, Ttl, ShardCount, StorageUsed, LogCount}'

# Calculate total storage per project
ve tls DescribeProjects --Region "{{user.region}}" | jq -r '.Projects[].ProjectId' | while read project_id; do
  total=$(ve tls DescribeTopics --Region "{{user.region}}" --ProjectId "$project_id" | jq '[.Topics[].StorageUsed] | add // 0')
  echo "Project $project_id: $total bytes ($(echo "$total / 1073741824" | bc) GB)"
done
```

#### Analysis Thresholds

| Metric | Warning | Critical | Action |
|--------|---------|----------|--------|
| Single topic storage > 100 GB | Yes | No | Review TTL, consider archiving |
| Single topic storage > 500 GB | No | Yes | Immediate optimization needed |
| Daily ingestion > 50 GB | Yes | No | Review collection config |
| Daily ingestion > 200 GB | No | Yes | Check for log loops/duplicates |
| Index storage > 3x raw storage | Yes | No | Review index configuration |

#### Output Format

```markdown
## Log Volume Analysis Report — [Date]

### Storage by Topic
| Project | Topic | TTL (days) | Storage (GB) | Log Count | Cost Estimate/mo |
|---------|-------|------------|--------------|-----------|-----------------|
| prod-app | nginx-access | 7 | 45.2 | 120M | ¥135 |
| prod-app | app-error | 30 | 12.8 | 8M | ¥38 |
| prod-app | audit-log | 180 | 256.0 | 500M | ¥1,536 |

### Top Cost Drivers
1. **audit-log** (256 GB, 180-day TTL) — Consider archiving to TOS after 30 days
2. **nginx-access** (45 GB, high ingestion) — Consider reducing TTL or sampling

### Optimization Opportunities
- Reduce audit-log TTL from 180 to 90 days: save ~¥768/mo
- Archive old logs to TOS cold storage: save ~¥1,200/mo
- Remove unused indexes on nginx-access: save ~¥45/mo
```

---

### Operation: OptimizeIndexes — Identify Unused/Redundant Indexes

Analyzes index usage and recommends optimizations.

#### Index Optimization Rules

| Issue | Detection Method | Recommendation |
|-------|-----------------|----------------|
| Full-text index on structured logs | Logs are JSON, full-text not needed | Disable full-text, use key-value only |
| Unused key-value fields | Field never queried in 30 days | Remove from index |
| Duplicate indexing | Same field indexed multiple ways | Keep only needed type |
| Over-indexing | > 20 key-value fields | Review necessity |
| Case-sensitive on low-cardinality | Field like `level` indexed case-sensitive | Set CaseSensitive: false |

#### Execution

```bash
# Get index configuration for all topics
ve tls DescribeProjects --Region "{{user.region}}" | jq -r '.Projects[].ProjectId' | while read project_id; do
  ve tls DescribeTopics --Region "{{user.region}}" --ProjectId "$project_id" | jq -r '.Topics[].TopicId' | while read topic_id; do
    topic_name=$(ve tls DescribeTopics --Region "{{user.region}}" --ProjectId "$project_id" --TopicId "$topic_id" | jq -r '.Topics[0].TopicName')
    index_config=$(ve tls DescribeIndex --Region "{{user.region}}" --ProjectId "$project_id" --TopicId "$topic_id" 2>/dev/null)
    if [ -n "$index_config" ]; then
      kv_count=$(echo "$index_config" | jq '.IndexRules.KeyValue | length')
      has_fulltext=$(echo "$index_config" | jq 'if .IndexRules.FullText then "yes" else "no" end')
      echo "Project: $project_id | Topic: $topic_name | FullText: $has_fulltext | KV Fields: $kv_count"
    fi
  done
done
```

#### Output Format

```markdown
## Index Optimization Report

| Project | Topic | FullText | KV Fields | Issue | Recommendation | Est. Savings |
|---------|-------|----------|-----------|-------|---------------|-------------|
| prod-app | nginx-access | yes | 15 | FullText on structured logs | Disable FullText | ¥45/mo |
| prod-app | app-log | no | 25 | Over-indexing | Remove 10 unused fields | ¥30/mo |
```

---

### Operation: DetectAnomalies — Detect Log Volume Anomalies

Identifies abnormal spikes or drops in log ingestion.

#### Execution

```bash
# Get topic ingestion stats (requires CMS metrics integration)
ve tls DescribeTopics --Region "{{user.region}}" --ProjectId "{{user.project_id}}" | jq '.Topics[] | {TopicName, LogCount, StorageUsed}'
```

#### Anomaly Detection Logic

| Pattern | Possible Cause | Investigation |
|---------|---------------|---------------|
| Sudden 10x+ increase | Log loop, debug logging enabled, attack | Check application logs, review recent deployments |
| Sudden drop to 0 | Agent crash, network issue, topic misconfigured | Check Logtail agent status, network connectivity |
| Gradual increase over days | Feature adding logs, traffic growth | Expected if correlated with traffic |
| Periodic spikes | Cron jobs, batch processing | Expected if scheduled |

---

## LogShipper Operations

### Operation: DescribeShippers — List LogShippers

#### Execution

```bash
# List all shippers for a topic
ve tls DescribeShippers --Region "{{user.region}}" --ProjectId "{{user.project_id}}" --TopicId "{{user.topic_id}}"

# List all shippers across all topics
ve tls DescribeProjects --Region "{{user.region}}" | jq -r '.Projects[].ProjectId' | while read project_id; do
  ve tls DescribeTopics --Region "{{user.region}}" --ProjectId "$project_id" | jq -r '.Topics[].TopicId' | while read topic_id; do
    ve tls DescribeShippers --Region "{{user.region}}" --ProjectId "$project_id" --TopicId "$topic_id" | jq -r '.Shippers[] | {ShipperId, ShipperName, DestinationType, Enable}'
  done
done
```

---

### Operation: AuditLogShipper — Check Shipper Health

#### Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Shipper enabled | `Enable` = true | Active | Enable or investigate |
| Destination accessible | TOS bucket exists and writable | Accessible | Check bucket permissions |
| No delivery errors | Check shipper status | No errors | Review error logs |
| Correct time window | Delivery interval matches requirements | Configured correctly | Adjust interval |

#### Execution

```bash
# Check shipper details
ve tls DescribeShipper --Region "{{user.region}}" --ProjectId "{{user.project_id}}" --TopicId "{{user.topic_id}}" --ShipperId "{{user.shipper_id}}"

# Check shipper task history
ve tls DescribeShipperTasks --Region "{{user.region}}" --ProjectId "{{user.project_id}}" --TopicId "{{user.topic_id}}" --ShipperId "{{user.shipper_id}}"
```

#### Output Format

```markdown
## LogShipper Audit Report

| Shipper | Topic | Destination | Status | Last Delivery | Issues |
|---------|-------|-------------|--------|---------------|--------|
| shipper-aaa | nginx-access | TOS: logs-bucket | Active | 2026-05-27 10:00 | None |
| shipper-bbb | app-error | TOS: logs-bucket | **Failed** | 2026-05-26 08:00 | Bucket not found |
```

---

## Troubleshooting

### Operation: TroubleshootMissingLogs — Diagnose Missing Logs

#### Diagnostic Checklist

| Check | Command | Expected |
|-------|---------|----------|
| Topic exists | `ve tls DescribeTopics --TopicId "{{user.topic_id}}"` | Topic found |
| Logtail agent running | Check ECS instance | Process running |
| Machine group configured | `ve tls DescribeMachineGroups` | Machine group active |
| Collector configured | `ve tls DescribeCollectors` | Collector active |
| Network connectivity | Test TLS endpoint from ECS | Reachable |
| Log file exists | Check file path on ECS | File exists and readable |
| Log format matches | Compare log format with parsing config | Format matches |

#### Execution

```bash
# Step 1: Verify topic exists
ve tls DescribeTopics --Region "{{user.region}}" --ProjectId "{{user.project_id}}" --TopicId "{{user.topic_id}}"

# Step 2: Check machine groups
ve tls DescribeMachineGroups --Region "{{user.region}}"

# Step 3: Check collectors
ve tls DescribeCollectors --Region "{{user.region}}" --ProjectId "{{user.project_id}}"

# Step 4: Search for recent logs
ve tls SearchLogs --Region "{{user.region}}" --ProjectId "{{user.project_id}}" --TopicId "{{user.topic_id}}" --Query "*" --StartTime "$(date -d '1 hour ago' +%s)" --EndTime "$(date +%s)"
```

#### Common Issues and Fixes

| Issue | Root Cause | Fix |
|-------|-----------|-----|
| No logs in topic | Logtail agent not installed | Install Logtail on ECS |
| Logs delayed > 5 min | Network issue, high load | Check network, increase shards |
| Partial logs collected | File path mismatch, regex wrong | Fix collector configuration |
| Logs not searchable | Index not enabled or misconfigured | Enable/correct index configuration |
| Parse failures | Log format doesn't match extractor | Update extractor configuration |

---

## Error Taxonomy

| `code` | 含义 | 分辨率 |
|--------|------|--------|
| `InvalidProjectName` | 日志项目名称格式无效或已存在 | 0 retries; **HALT** — 使用合法名称，避免重名 |
| `InvalidTopicName` | 日志主题名称格式无效或已存在 | 0 retries; **HALT** — 使用合法名称，避免重名 |
| `QuotaExceeded.Project` | 日志项目数量超出配额限制 | 0 retries; **HALT** — 删除未使用项目或提额 |
| `QuotaExceeded.Shard` | 分片数量超出配额限制 | 0 retries; **HALT** — 合并分片或提额 |
| `InvalidShardId` | 分片 ID 不存在或已合并 | 0 retries; **HALT** — 检查 ShardId 是否有效 |
| `InvalidLogEntry` | 日志内容格式无效或超过大小限制 | 2 retries/exponential/5s/10s/20s; **RETRY** — 检查日志格式和大小 (< 1MB/条) |
| `InvalidIndexConfig` | 索引配置格式无效或字段类型不支持 | 0 retries; **HALT** — 验证索引字段类型 (text/long/double/json) |
| `ProjectNotActive` | 日志项目状态异常，不可写入 | 0 retries; **HALT** — 检查项目状态，联系技术支持 |
| `InvalidAlarmId` | 告警规则 ID 不存在或格式错误 | 0 retries; **HALT** — 检查告警 ID 是否有效 |
| `Throttling` | 请求频率过高触发限流 | 3 retries/exponential/2s/4s/8s; **RETRY** — 背退等待后重试 |
| `InternalError` | 服务端内部错误 | 3 retries/exponential/2s/4s/8s; **RETRY** — 超过重试次数后 HALT 并记录 RequestId |
| `Unauthorized` | 鉴权失败，权限不足 | 0 retries; **HALT** — 检查 IAM 策略是否包含 TLS/SLS 权限 |

## Reference Directory

- [Core Concepts](references/core-concepts.md)
- [API & SDK Usage](references/api-sdk-usage.md)
- [CLI Usage](references/cli-usage.md)
- [Troubleshooting Guide](references/troubleshooting.md)
- [Monitoring](references/monitoring.md)
- [Integration](references/integration.md)
- [User Experience Specification](../ve-skill-generator/references/user-experience-spec.md)
- [Execution Environment Setup](../ve-skill-generator/references/execution-environment.md)
- [CLI Behavioral Reference](../ve-skill-generator/references/cli-behavior.md)
- [FinOps Best Practices](../ve-skill-generator/references/finops-best-practices.md)
- [GCL Rubric](references/rubric.md)
- [GCL Prompt Templates](references/prompt-templates.md)

## Operational Best Practices

- **TTL management:** Set appropriate TTL per topic type; don't use 365 days for all logs
- **Index wisely:** Only index fields you actually query; full-text index has storage overhead
- **LogShipper:** Configure delivery to TOS for long-term retention and cost savings
- **Shard sizing:** Start with 1-2 shards; increase only if ingestion rate requires it
- **Structured logging:** Use JSON format for application logs; enables efficient key-value indexing
- **Tagging:** Tag projects and topics with owner, environment, and purpose
- **Alert on anomalies:** Set up volume-based alerts to catch issues early
- **Regular review:** Monthly review of topics, indexes, and shippers for optimization

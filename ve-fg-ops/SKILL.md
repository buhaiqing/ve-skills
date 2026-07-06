---
name: ve-fg-ops
description: >-
  Use when the user needs to deploy, configure, troubleshoot, or manage Volcengine
  (火山引擎) FunctionGraph (函数计算) — function lifecycle, trigger management,
  version/alias management, concurrent quota, and diagnostics. User mentions
  函数计算, FunctionGraph, serverless, lambda, 函数, function invoke, or describes
  serverless scenarios (e.g., creating functions, configuring triggers, viewing logs,
  debugging invocation failures) even without naming the product directly. Not for
  container service (VKE), ECS, or traditional application deployment.
license: MIT
compatibility: >-
  Official Volcengine CLI (`ve`, Go binary, no runtime), Go 1.21+ runtime
  (JIT SDK fallback; scripts compatible with Go 1.14+ syntax), valid API credentials,
  network access to Volcengine endpoints (open.volcengineapi.com).
metadata:
  author: volcengine
  version: "1.1.0"
  last_updated: "2026-06-04"
  runtime: Harness AI Agent, Claude Code, Cursor, or compatible Agent runtimes
  go_script_syntax_minimum: "1.14"
  go_jit_runtime_version: "1.21+"
  api_profile: "FunctionGraph OpenAPI — https://www.volcengine.com/docs/6668"
  cli_applicability: dual-path
  cli_support_evidence: >-
    Confirmed via `ve functiongraph --help` — FunctionGraph is supported by the ve CLI.
    CLI prefix: `ve functiongraph`.
    Service ID: functiongraph.
    Go SDK: github.com/volcengine/volc-sdk-golang/service/functiongraph.
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

> This skill follows the [Agent Skill OpenSpec](https://agentskills.io/specification).

# Volcengine FunctionGraph Operations Skill

## Overview

Volcengine FunctionGraph (火山引擎函数计算) provides serverless compute for event-driven and on-demand applications. This skill is an **operational runbook** for agents: explicit scope, credential rules, pre-flight checks, **dual-path execution** (official **SDK/API** and official **`ve` CLI** flows), response validation, and failure recovery. **Do not use the web console as the primary agent execution path.**

### CLI applicability

- **`cli_applicability: dual-path`:** Official `ve` CLI supports FunctionGraph. Document **both** the SDK step **and** the `ve` CLI step for every operation the CLI exposes.

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use conditions with precise triggers; VPC/ECS delegated to other skills |
| 2 | **Structured I/O** | `{{env.*}}` for credentials, `{{user.*}}` for function params, `{{output.*}}` for API responses |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover, with numbered imperative steps |
| 4 | **Complete Failure Strategies** | Error taxonomy with 15+ FunctionGraph-specific codes; HALT vs retry per error type |
| 5 | **Absolute Single Responsibility** | FunctionGraph only; triggers managed inline; VPC/ECS/APIG delegated |

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- User mentions "函数计算", "FunctionGraph", "serverless", "lambda", "function invoke"
- Task involves function lifecycle (create, delete, update, get, list)
- Task involves trigger management (Timer, APIG, HTTP, CTS, RocketMQ)
- Task involves version or alias management (publish, create/delete alias, list versions)
- Task involves concurrent quota (query, adjust reserved concurrency)
- Task involves function monitoring (logs, metrics, invocation errors)
- Task keywords: function, 函数, trigger, 触发器, serverless, 无服务器, invoke, 调用, 版本, 别名

### SHOULD NOT Use This Skill When

- Task is purely billing / account management → delegate to: `ve-billing-ops` (when present)
- Task is IAM / permission model only → delegate to: `ve-iam-ops` (when present)
- Task is about **Container Service (VKE)** → delegate to: `ve-vke-ops`
- Task is about **ECS instances** → delegate to: `ve-ecs-ops`
- User insists on **console-only** flows with no API → state limitation; do not invent undocumented HTTP steps

### Delegation Rules

- FunctionGraph relies on VPC for private network access: verify VPC exists via `ve-vpc-ops` before configuring VPC for functions
- API Gateway triggers: delegate APIG configuration to `ve-apig-ops` (if not available, use APIG OpenAPI directly)
- Monitoring and alerts: use `ve-cms-ops` for alarm rule configuration

## Variable Convention (Agent-Readable)

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_REGION}}` | From runtime environment | Use documented default only if skill explicitly allows |
| `{{user.region}}` | User-supplied region | Ask once; reuse |
| `{{user.function_name}}` | User-supplied function name | Ask once; reuse; pattern `^[a-zA-Z][a-zA-Z0-9-_]{1,63}$` |
| `{{user.runtime}}` | Runtime type (e.g., Python3.9, Node.js16) | Ask once; reuse; validate against supported list |
| `{{user.handler}}` | Entry handler (e.g., index.handler) | Ask once; reuse |
| `{{user.memory_size}}` | Memory in MB (128-32768) | Ask once; reuse; default 128 |
| `{{user.timeout}}` | Timeout in seconds (1-900) | Ask once; reuse; default 60 |
| `{{user.trigger_type}}` | Trigger type (Timer/APIG/HTTP/CTS/RocketMQ) | Ask once; reuse |
| `{{output.function_id}}` | From CreateFunction/GetFunction response | Parse from response |
| `{{output.function_arn}}` | Function ARN | Parse from response |
| `{{output.trigger_id}}` | From CreateTrigger response | Parse from response |
| `{{output.version}}` | Published version number | Parse from PublishVersion response |

> **`{{env.*}}` MUST NOT** be collected from the user. **`{{user.*}}`** MUST be collected interactively when missing.

> **Security Warning (Credential Masking — MANDATORY):** **NEVER** log, print, or expose `VOLCENGINE_SECRET_KEY`, `secret_key`, `SecretKey`, or any credential field value in console output, debug messages, error messages, or logs.
>
> **Credential verification MUST check existence only**, never echo the value:
> - Bash: `test -n "$VOLCENGINE_SECRET_KEY"` ✅ | `echo $VOLCENGINE_SECRET_KEY` ❌
> - Go: `if os.Getenv("VOLCENGINE_SECRET_KEY") == ""` ✅ | `fmt.Println(os.Getenv("VOLCENGINE_SECRET_KEY"))` ❌

## API and Response Conventions (Agent-Readable)

- **Volcengine FunctionGraph OpenAPI** is canonical for path, query, body fields, enums, and response shapes.
- **Endpoint:** `functiongraph.volcengineapi.com` (default: `open.volcengineapi.com`)
- **Errors:** Responses with `Error` object containing `Code` and `Message` fields.
- **Timestamps:** ISO 8601 format (e.g. `2026-04-28T10:00:00Z`).

### Key Response Fields

| Operation | JSON Path | Type | Description |
|-----------|-----------|------|-------------|
| CreateFunction | `$.Result.FunctionId` | string | Function ID |
| CreateFunction | `$.Result.FunctionName` | string | Function name |
| CreateFunction | `$.Result.FunctionArn` | string | Function ARN |
| GetFunction | `$.Result.FunctionName` | string | Function name |
| GetFunction | `$.Result.Status` | string | Function status (`Active`, `Creating`, `Failed`) |
| ListFunctions | `$.Result.Functions` | array | Function list |
| CreateTrigger | `$.Result.TriggerId` | string | Trigger ID |
| PublishVersion | `$.Result.Version` | integer | Published version number |

### State Transitions

| Operation | Initial State | Target State | Poll Interval | Max Wait |
|-----------|---------------|--------------|---------------|----------|
| CreateFunction | — | `Active` | 5s | 120s |
| UpdateFunction | `Active` | `Active` | 5s | 120s |
| DeleteFunction | `Active` | absent | 5s | 60s |

## Quick Start

### What This Skill Does
This skill enables you to deploy, configure, troubleshoot, and monitor Volcengine (火山引擎) FunctionGraph functions using the `ve` CLI (primary) or JIT Go SDK (fallback).

### Prerequisites
- [ ] `ve` CLI installed (or Go runtime for JIT fallback)
- [ ] Credentials configured: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`
- [ ] Region set: `VOLCENGINE_REGION`

### Verify Setup
```bash
ve functiongraph ListFunctions --Region {{env.VOLCENGINE_REGION}}
```

### Your First Command
```bash
ve functiongraph ListFunctions --Region {{env.VOLCENGINE_REGION}}
```

### Next Steps
- [Core Concepts](references/core-concepts.md) — Understand FunctionGraph architecture
- [Common Operations](#execution-flows) — Create, invoke, and manage functions
- [Troubleshooting](references/troubleshooting.md) — Fix common issues

## Capabilities at a Glance

| Operation | Description | Complexity | Risk Level |
|-----------|-------------|------------|------------|
| CreateFunction | Create a new function | Medium | Low |
| GetFunction | View function details | Low | None |
| ListFunctions | List all functions | Low | None |
| UpdateFunction | Modify function code/config | Medium | Medium |
| InvokeFunction | Invoke function (sync/async) | Low | None |
| DeleteFunction | Remove a function | Medium | **High** — irreversible |
| CreateTrigger | Add trigger to function | Medium | Low |
| DeleteTrigger | Remove trigger from function | Low | Medium |
| PublishVersion | Publish a function version | Low | Low |

## Supported Runtimes

| Runtime | Description | Recommended Scenarios |
|---------|-------------|----------------------|
| Python 3.x | Python runtime | Data processing, script tasks |
| Node.js 16/18 | JavaScript runtime | Web backend, API services |
| Go 1.x | Go runtime | High-performance compute |
| Java 8/11/17 | Java runtime | Enterprise applications |
| PHP 7.4/8.0 | PHP runtime | Web applications |
| Custom Image | Custom container image | Complex dependency scenarios |

## Execution Flows (Agent-Readable)

Every operation: **Pre-flight → Execute (CLI + SDK) → Validate → Recover**. Do not skip phases.

**Preference hint:** When CLI does not support a specific operation, JIT build a Go SDK script. CLI is preferred for coverage; Go SDK is used for operations CLI does not expose.

### Operation: Create Function

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| CLI / deps | `ve version` | Exit code 0 | Document CLI install |
| Credentials | Verify env vars set | Non-empty keys | HALT; user configures env |
| Region | Verify `{{user.region}}` valid | Region supported | Suggest valid region |
| Function name | Validate pattern `^[a-zA-Z][a-zA-Z0-9-_]{1,63}$` | Valid name | HALT; provide naming rules |
| Code source | Check code URL or zip exists | Accessible | HALT; provide valid code |

#### Execution — CLI (`ve`) (Primary Path)

```bash
ve functiongraph CreateFunction \
  --FunctionName "{{user.function_name}}" \
  --Runtime "{{user.runtime}}" \
  --Handler "{{user.handler}}" \
  --MemorySize {{user.memory_size}} \
  --Timeout {{user.timeout}} \
  --CodeType "{{user.code_type}}" \
  --CodeUrl "{{user.code_url}}" \
  --Description "{{user.description}}" \
  --Region "{{env.VOLCENGINE_REGION}}"
```

#### Execution — JIT Go SDK (Fallback Path)

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/functiongraph"
)

func main() {
    instance := functiongraph.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "FunctionName": os.Args[1],
        "Runtime":      os.Args[2],
        "Handler":      os.Args[3],
        "MemorySize":   128,
        "Timeout":      60,
        "CodeType":     "URL",
        "CodeUrl":      os.Args[4],
        "Region":       os.Getenv("VOLCENGINE_REGION"),
    }

    resp, err := instance.Client.Request("functiongraph", "CreateFunction", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "API call failed: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(string(resp))
}
```

Execute:
```bash
cd /tmp/ve-sdk-workspace && go run main.go \
  "{{user.function_name}}" "{{user.runtime}}" \
  "{{user.handler}}" "{{user.code_url}}"
```

#### Post-execution Validation

1. Read `{{output.function_id}}` and `{{output.function_arn}}` from the response.
2. Poll until `Status == "Active"` or timeout:

```bash
for i in $(seq 1 24); do
  STATUS=$(ve functiongraph GetFunction \
    --FunctionName "{{user.function_name}}" \
    --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.Status')
  [ "$STATUS" = "Active" ] && break
  sleep 5
done
```

3. On success, report `{{output.function_name}}` and `{{output.function_arn}}`.
4. On failure, go to **Failure Recovery**.

#### Failure Recovery

| Error pattern | Max retries | Recovery |
|--------------|-------------|---------|
| `InvalidParameter` / 400 | 0–1 | Fix args from OpenAPI; retry once if safe — `[ERROR] InvalidParameter: Check parameters against OpenAPI docs` |
| `FunctionAlreadyExists` | 0 | HALT; ask user for different name or update existing — `[ERROR] Function "{{user.function_name}}" already exists.` |
| `QuotaExceeded` | 0 | HALT — `[ERROR] Function quota exceeded. Delete unused functions or request quota increase.` |
| `InsufficientBalance` | 0 | HALT — `[ERROR] Account balance insufficient. Recharge before proceeding.` |
| `InvalidRuntime` | 0 | HALT; suggest valid runtime — `[ERROR] Runtime "{{user.runtime}}" not supported.` |
| Throttling / 429 | 3 | Exponential backoff; respect `Retry-After` — `⚠️ Rate limit reached. Retrying in {backoff}s...` |
| `InternalError` / 5xx | 3 | 2s, 4s, 8s backoff; retry, then HALT with RequestId |

### Operation: Invoke Function

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Function exists | `ve functiongraph GetFunction` | Status `Active` | HALT; create function first |
| Payload | Validate JSON if synchronous | Valid JSON | HALT; fix payload format |

#### Execution — CLI

```bash
# Synchronous invocation
ve functiongraph InvokeFunction \
  --FunctionName "{{user.function_name}}" \
  --InvocationType RequestResponse \
  --Payload '{{user.payload}}' \
  --Region "{{env.VOLCENGINE_REGION}}"

# Asynchronous invocation
ve functiongraph InvokeFunction \
  --FunctionName "{{user.function_name}}" \
  --InvocationType Event \
  --Payload '{{user.payload}}' \
  --Region "{{env.VOLCENGINE_REGION}}"
```

#### Post-execution Validation

- Synchronous: Check exit code == 0; parse `{{output.invocation_result}}` from response JSON
- Asynchronous: Check exit code == 0; function processes in background

### Operation: Delete Function (Safety Gate Required)

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Function exists | `ve functiongraph GetFunction` | Exists | HALT; function not found |
| User confirmation | Prompt user | Yes | HALT |
| Check triggers | List triggers | Note warning if triggers exist | Warn user about cascading deletes |

#### Execution — CLI

```bash
# List and delete all triggers first
for trigger_id in $(ve functiongraph ListTriggers \
  --FunctionName "{{user.function_name}}" \
  --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.Triggers[].TriggerId'); do
  ve functiongraph DeleteTrigger \
    --FunctionName "{{user.function_name}}" \
    --TriggerId "$trigger_id" \
    --Region "{{env.VOLCENGINE_REGION}}"
done

# Delete function
ve functiongraph DeleteFunction \
  --FunctionName "{{user.function_name}}" \
  --Region "{{env.VOLCENGINE_REGION}}"
```

#### Post-execution Validation

```bash
ve functiongraph GetFunction \
  --FunctionName "{{user.function_name}}" \
  --Region "{{env.VOLCENGINE_REGION}}" 2>&1 | grep -q "FunctionNotFound"
```

Expected: `FunctionNotFound` error confirms deletion.

### Operation: Create Timer Trigger

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Function exists | GetFunction | Exists | HALT |
| Cron expression | Validate format | Valid | HALT; provide correct format |

#### Execution — CLI

```bash
ve functiongraph CreateTrigger \
  --FunctionName "{{user.function_name}}" \
  --TriggerType Timer \
  --TriggerName "{{user.trigger_name}}" \
  --TimerConfig '{"Schedule":"{{user.cron_expression}}","Enable":true}' \
  --Region "{{env.VOLCENGINE_REGION}}"
```

#### Post-execution Validation

```bash
ve functiongraph ListTriggers \
  --FunctionName "{{user.function_name}}" \
  --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.Triggers[] | select(.TriggerName == "{{user.trigger_name}}") | .TriggerName'
```

Expected: `{{user.trigger_name}}` appears in the list.

## Error Taxonomy

| Error Code | Meaning | Resolution |
|-----------|---------|-----------|
| `FunctionNotFound` | Function does not exist | 0 retries; **HALT** — Check name, suggest creation |
| `FunctionAlreadyExists` | Function name in use | 0 retries; **RETRY** — Suggest different name or UpdateFunction |
| `InvalidParameter` | Request validation failed | 1 retry; **RETRY** — Align with OpenAPI schema |
| `InvalidRuntime` | Unsupported runtime | 0 retries; **HALT** — Suggest supported runtime list |
| `InvalidCronExpression` | Cron expression invalid | 0 retries; **RETRY** — Provide correct cron format |
| `ResourceLimitExceeded` | Quota limit reached | 0 retries; **HALT** — request quota increase |
| `InsufficientBalance` | Account not funded | 0 retries; **HALT** — recharge required |
| `CodeStorageExceeded` | Code package too large | 0 retries; **RETRY** — Optimize code or use OSS URL |
| `InvocationError` | Function runtime error | 2 retries; **RETRY** — Check function logs for details |
| `TimeoutError` | Execution timed out | 2 retries; **RETRY** — Increase timeout or optimize code |
| `MemoryExceeded` | Memory limit reached | 2 retries; **RETRY** — Increase MemorySize configuration |
| `ConcurrentInvocationExceeded` | Concurrency limit hit | 3 retries; **RETRY** — Wait or request quota increase |
| `TriggerNotFound` | Trigger does not exist | 0 retries; **HALT** — Check trigger ID/name |
| `TriggerAlreadyExists` | Trigger name conflict | 0 retries; **RETRY** — Use different trigger name |
| `VersionNotFound` | Version does not exist | 0 retries; **HALT** — Check version number |
| `Throttling` | Rate limit exceeded | 3 retries/exponential; **RETRY** — Backoff with delay |
| `InternalError` | Server-side error | 3 retries; **RETRY** — Retry, escalate with RequestId |

## Prerequisites

### Install CLI

```bash
# macOS ARM64
curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/{{env.ve_version}}/ve-darwin-arm64 -o /usr/local/bin/ve && chmod +x /usr/local/bin/ve

# Linux x86_64
curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/{{env.ve_version}}/ve-linux-amd64 -o /usr/local/bin/ve && chmod +x /usr/local/bin/ve
```

### Configure Credentials

```bash
export VOLCENGINE_ACCESS_KEY="{{env.VOLCENGINE_ACCESS_KEY}}"
export VOLCENGINE_SECRET_KEY="{{env.VOLCENGINE_SECRET_KEY}}"
export VOLCENGINE_REGION="{{env.VOLCENGINE_REGION}}"
```

## Reference Directory

- [Core Concepts](references/core-concepts.md)
- [API & SDK Usage](references/api-sdk-usage.md)
- [CLI Usage](references/cli-usage.md)
- [Troubleshooting Guide](references/troubleshooting.md)
- [Monitoring & Alerts](references/monitoring.md)
- [Integration](references/integration.md)
- [Knowledge Base](references/knowledge-base.md)
- [GCL Rubric](references/rubric.md)
- [GCL Prompt Templates](references/prompt-templates.md)

## Diagnostic Commands

```bash
# View function details
ve functiongraph GetFunction --FunctionName "{{user.function_name}}" --Region "{{env.VOLCENGINE_REGION}}"

# List all functions
ve functiongraph ListFunctions --Region "{{env.VOLCENGINE_REGION}}"

# View function logs
ve functiongraph GetFunctionLogs \
  --FunctionName "{{user.function_name}}" \
  --StartTime "2024-01-01T00:00:00Z" \
  --EndTime "2024-01-01T23:59:59Z" \
  --Region "{{env.VOLCENGINE_REGION}}"

# View function metrics
ve functiongraph GetFunctionMetrics \
  --FunctionName "{{user.function_name}}" \
  --MetricName ConcurrentExecutions \
  --StartTime "2024-01-01T00:00:00Z" \
  --EndTime "2024-01-01T23:59:59Z" \
  --Region "{{env.VOLCENGINE_REGION}}"

# List function versions
ve functiongraph ListVersions \
  --FunctionName "{{user.function_name}}" \
  --Region "{{env.VOLCENGINE_REGION}}"
```

## Related Skills

- **ve-iam-ops** — Configure IAM permissions for functions
- **ve-vpc-ops** — Configure VPC network for functions
- **ve-cms-ops** — Configure function monitoring and alerts
- **ve-billing-ops** — View function billing and cost

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-31 | Initial frontmatter + Five Core Standards + dual-path + error taxonomy |
| 1.1.0 | 2026-06-04 | GCL rollout: added ## Quality Gate (GCL), references/rubric.md, references/prompt-templates.md |

## Operational Best Practices

- **Cold start mitigation:** Set reserved concurrency for latency-sensitive functions. Use warm-up triggers (Timer) during low-traffic periods to keep instances warm. Choose provisioned concurrency for predictable traffic patterns.
- **Concurrency limits:** Monitor `ConcurrentExecutions` metric. Set function-level reserved concurrency to prevent a single function from exhausting account-level quota. Default account quota is 100 concurrent executions; request increase via support ticket when needed.
- **Function timeout configuration:** Set `Timeout` realistically based on function workload. For API-facing functions use 10-30s; for data-processing functions use 120-300s. Timeout values above 300s require async invocation pattern.
- **Logging and monitoring:** Enable function logs via SLS (Simple Log Service). Configure `GetFunctionLogs` for real-time debugging. Set up `ve-cms-ops` alarm rules on error count, duration P99, and throttling rate.
- **Version and alias management:** Use `PublishVersion` to freeze stable code and `CreateAlias` to route traffic (e.g., 90% v1, 10% v2 for canary). Never reference `$LATEST` in production triggers; always pin to a version or alias.
- **Least privilege IAM:** Assign functions the minimum permissions needed via a dedicated service role. Do not reuse root-account credentials. Rotate function configuration secrets via `{{env.VOLCENGINE_SECRET_KEY}}` — never embed in code.

## Quality Gate (GCL)

> Mandatory. max_iter=3.

| Tier | Operations | Safety |
|---|---|---|
| **Destructive** | DeleteFunction, DeleteTrigger | 1.0 |
| **State-changing** | UpdateFunction, PublishVersion, CreateTrigger | 1.0 |
| **Mutating** | CreateFunction | ≥0.5 |
| **Read-only** | GetFunction, ListFunctions, InvokeFunction | ≥0 |

Safety: DeleteFunction ALL versions/triggers lost. UpdateFunction: in-flight invocations complete with old code.

### Cross-skill: Billing→ve-billing-ops

---
name: ve-alb-ops
description: >-
  Use when the user needs to deploy, configure, troubleshoot, or monitor Volcengine
  (火山引擎) ALB (应用型负载均衡 / Application Load Balancer) — instance lifecycle,
  listener/rule management, server group configuration, health checks, and content-based
  routing. User mentions ALB, 应用型负载均衡, Application Load Balancer, 应用型ALB,
  or describes L7 routing scenarios (e.g., path-based routing, host-based forwarding,
  HTTPS termination, WebSocket/gRPC load balancing) even without naming the product
  directly. Not for CLB (Classic Load Balancer), billing, IAM, or compute resources.
license: MIT
compatibility: >-
  Official Volcengine CLI (`ve`, Go binary, no runtime), Go 1.21+ runtime
  (JIT SDK fallback; scripts compatible with Go 1.14+ syntax), valid API credentials,
  network access to Volcengine ALB endpoints.
metadata:
  author: volcengine
  version: "1.1.0"
  last_updated: "2026-06-04"
  runtime: Harness AI Agent, Claude Code, Cursor, or compatible Agent runtimes
  go_version_minimum: "1.14"
  go_jit_runtime_version: "1.21+"
  api_profile: "ALB API 2022-03-01"
  cli_applicability: dual-path
  cli_support_evidence: >-
    Confirmed via `ve alb --help` — ALB is supported by the ve CLI.
    API docs: https://www.volcengine.com/docs/6398
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

> This skill follows the [Agent Skill OpenSpec](https://agentskills.io/specification).

# Volcengine ALB Operations Skill

## Overview

ALB (应用型负载均衡 / Application Load Balancer) on Volcengine (火山引擎) provides Layer 7 application-level load balancing with content-based routing, HTTPS termination, WebSocket/gRPC support, and fine-grained traffic management. This skill is an **operational runbook** for agents: explicit scope, credential rules, pre-flight checks, **dual-path execution** (official **SDK/API** and **`ve` CLI**), response validation, and failure recovery.

> **UX Compliance:** This skill follows the [User Experience Specification](../../ve-skill-generator/references/user-experience-spec.md).

### CLI applicability

- **`cli_applicability: dual-path`:** Official `ve` CLI supports ALB. You **MUST** document **both** the SDK step **and** the `ve` CLI step for every operation.

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use conditions with precise triggers and delegation rules |
| 2 | **Structured I/O** | `{{env.VOLCENGINE_*}}` (env), `{{user.*}}` (interactive), `{{output.*}}` (API response) |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | Error taxonomy with ≥ 10 ALB-specific codes; HALT vs retry per type |
| 5 | **Absolute Single Responsibility** | One product (ALB); cross-product delegation documented |

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- User mentions "Volcengine ALB", "火山引擎 ALB", "应用型负载均衡", "Application Load Balancer"
- Task involves lifecycle operations on **ALB**: CreateLoadBalancer, DescribeLoadBalancers, DeleteLoadBalancer, ModifyLoadBalancerAttributes
- Task involves **Listeners**: CreateListener, DescribeListeners, DeleteListener, ModifyListenerAttributes
- Task involves **Rules**: CreateRule, DescribeRules, DeleteRule, ModifyRuleAttributes
- Task involves **Server Groups**: CreateServerGroup, DescribeServerGroups, DeleteServerGroup, ModifyServerGroupAttributes
- Task involves **Health Checks**: health check configuration for server groups
- Task involves **content-based routing**: path-based routing, host-based routing, weighted forwarding
- Task involves **TLS/HTTPS termination**: certificate binding, SSL policy configuration
- Task keywords: L7路由, 路径转发, 域名转发, path-based routing, host-based routing, HTTPS termination, server group, 服务器组
- Cost optimization: idle ALB detection, cost summary, orphaned resource cleanup

### SHOULD NOT Use This Skill When

- Task is about **CLB (Classic Load Balancer)** → delegate to: `ve-clb-ops`
- Task is about **VPC creation** → delegate to: `ve-vpc-ops`
- Task is about **EIP allocation** → delegate to: `ve-eip-ops`
- Task is about **ECS instance creation** → delegate to: `ve-ecs-ops`
- Task is about **TLS certificate management** → delegate to: `ve-certificate-ops` (when present)
- Task is purely billing / account management → delegate to: `ve-billing-ops` (when present)
- User insists on **console-only** flows → state limitation

### Delegation Rules

- ALB requires VPC + Subnet → verify via `ve-vpc-ops`
- ALB may need EIP (public type) → allocate via `ve-eip-ops`
- ALB backend servers (ServerGroup members) are ECS → coordinate with `ve-ecs-ops`
- ALB HTTPS listeners require TLS certificates → reference `ve-certificate-ops`
- Multi-product requests: handle each product with its skill; do not merge unrelated APIs

## Variable Convention (Agent-Readable)

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | From runtime environment | NEVER ask user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | From runtime environment | NEVER ask; fail if unset; **ALWAYS mask** |
| `{{env.VOLCENGINE_REGION}}` | From runtime environment | Default from env |
| `{{user.region}}` | User-supplied region | Ask once; reuse |
| `{{user.alb_name}}` | ALB name | Ask once; reuse |
| `{{user.alb_id}}` | ALB instance ID | Format `alb-xxxxxxxxx` |
| `{{user.vpc_id}}` | Parent VPC ID | Format `vpc-xxxxxxxxx` |
| `{{user.subnet_id}}` | ALB subnet | Format `subnet-xxxxxxxxx` |
| `{{user.listener_name}}` | Listener name | Ask once |
| `{{user.listener_id}}` | Listener ID | Format `lsnr-xxxxxxxxx` |
| `{{user.listener_protocol}}` | Listener protocol | HTTPS, HTTP, gRPC |
| `{{user.listener_port}}` | Listener port | e.g., `80`, `443` |
| `{{user.certificate_id}}` | TLS Certificate ID | For HTTPS listeners |
| `{{user.server_group_id}}` | Server group ID | Format `rsp-xxxxxxxxx` |
| `{{user.server_group_name}}` | Server group name | Ask once |
| `{{user.rule_id}}` | Rule ID | Format `rule-xxxxxxxxx` |
| `{{user.domain}}` | Domain for host-based rule | e.g., `api.example.com` |
| `{{user.path_pattern}}` | URL path pattern | e.g., `/api/*` |
| `{{output.alb_id}}` | From CreateLoadBalancer response | Parse `$.Result.LoadBalancerId` |
| `{{output.listener_id}}` | From CreateListener response | Parse `$.Result.ListenerId` |
| `{{output.server_group_id}}` | From CreateServerGroup response | Parse `$.Result.ServerGroupId` |
| `{{output.rule_id}}` | From CreateRule response | Parse `$.Result.RuleId` |

> **Security Warning (Credential Masking — MANDATORY):** **NEVER** log, print, or expose `VOLCENGINE_SECRET_KEY`, `secret_key`, `SecretKey`, or any credential field value in console output, debug messages, error messages, or logs.
>
> **Security Warning:** NEVER echo or log `VOLCENGINE_SECRET_KEY` or any credential values. Verify existence only with `test -n "$VOLCENGINE_SECRET_KEY"`.
> **Password Masking:** NEVER display passwords in output. Always show as `<masked>`.

## API and Response Conventions (Agent-Readable)

- **Volcengine ALB OpenAPI (2022-03-01)** is canonical.
- **Endpoint:** By default, the Volcengine gateway (`open.volcengineapi.com`).
- **Errors:** Responses with `Error` object containing `Code` and `Message` fields.
- **Timestamps:** ISO 8601 format.
- **Idempotency:** ALB operations typically use ClientToken for idempotent creation.

### Key Response Fields

| Operation | JSON Path | Type | Description |
|-----------|-----------|------|-------------|
| CreateLoadBalancer | `$.Result.LoadBalancerId` | string | ALB ID, format `alb-xxx` |
| DescribeLoadBalancers | `$.Result.LoadBalancers` | array | ALB instance list |
| DescribeLoadBalancers | `$.Result.LoadBalancers[].LoadBalancerId` | string | ALB ID |
| DescribeLoadBalancers | `$.Result.LoadBalancers[].LoadBalancerName` | string | ALB name |
| DescribeLoadBalancers | `$.Result.LoadBalancers[].VpcId` | string | Parent VPC |
| DescribeLoadBalancers | `$.Result.LoadBalancers[].Status` | string | ALB status: `active`, `inactive`, `creating`, `deleting` |
| DescribeLoadBalancers | `$.Result.LoadBalancers[].Type` | string | `public` or `private` |
| DescribeLoadBalancers | `$.Result.LoadBalancers[].Address` | string | Internal IP address |
| DescribeLoadBalancers | `$.Result.LoadBalancers[].EipAddress` | string | Public IP (public type) |
| DescribeLoadBalancers | `$.Result.LoadBalancers[].CreateTime` | string | Creation timestamp |
| CreateListener | `$.Result.ListenerId` | string | Listener ID |
| DescribeListeners | `$.Result.Listeners[].ListenerId` | string | Listener ID |
| DescribeListeners | `$.Result.Listeners[].Protocol` | string | HTTPS, HTTP, gRPC |
| DescribeListeners | `$.Result.Listeners[].Port` | int | Listening port |
| CreateRule | `$.Result.RuleId` | string | Rule ID |
| DescribeRules | `$.Result.Rules[].RuleId` | string | Rule ID |
| DescribeRules | `$.Result.Rules[].Domain` | string | Host domain |
| DescribeRules | `$.Result.Rules[].Url` | string | URL path pattern |
| CreateServerGroup | `$.Result.ServerGroupId` | string | Server group ID |
| DescribeServerGroups | `$.Result.ServerGroups[].ServerGroupId` | string | Server group ID |
| DescribeServerGroups | `$.Result.ServerGroups[].ServerGroupName` | string | Server group name |
| DescribeServerGroups | `$.Result.ServerGroups[].Servers[].ServerId` | string | Backend ECS ID |
| DescribeServerGroups | `$.Result.ServerGroups[].Servers[].Port` | int | Backend port |
| DescribeServerGroups | `$.Result.ServerGroups[].Servers[].Weight` | int | Weight (0-100) |
| DescribeServerGroups | `$.Result.ServerGroups[].Servers[].Status` | string | `healthy`, `unhealthy`, `unavailable` |

### Expected State Transitions

| Operation | Initial State | Target State | Poll Interval | Max Wait |
|-----------|---------------|--------------|---------------|----------|
| CreateLoadBalancer | — | `active` | 5s | 300s |
| DeleteLoadBalancer | any stable | `deleting` → absent | 5s | 300s |
| CreateListener | — | `active` | 5s | 120s |
| DeleteListener | any stable | absent | 5s | 120s |
| CreateServerGroup | — | `active` | 5s | 60s |
| ModifyLoadBalancer | `active` | `active` | 5s | 120s |

## Quick Start

### What This Skill Does
This skill enables you to create, configure, and manage Volcengine (火山引擎) ALB Application Load Balancers, listeners, content-based rules, server groups, and health checks using the `ve` CLI (primary) or JIT Go SDK (fallback).

### Prerequisites
- [ ] `ve` CLI installed (or Go runtime for JIT fallback)
- [ ] Credentials configured
- [ ] Region set
- [ ] VPC + Subnet already created — see `ve-vpc-ops`
- [ ] Backend ECS instances ready — see `ve-ecs-ops`

### Verify Setup
```bash
ve alb DescribeLoadBalancers --Region {{env.VOLCENGINE_REGION}}
```

### Your First Command
```bash
ve alb DescribeLoadBalancers --Region {{env.VOLCENGINE_REGION}}
```

### Next Steps
- [Core Concepts](references/core-concepts.md) — Understand ALB architecture
- [Common Operations](#execution-flows) — Create, manage, and delete resources
- [Troubleshooting](references/troubleshooting.md) — Fix common issues

## Capabilities at a Glance

| Operation | Description | Complexity | Risk Level |
|-----------|-------------|------------|------------|
| CreateLoadBalancer | Create ALB instance | Medium | Low |
| DescribeLoadBalancers | Query ALB list | Low | None |
| DeleteLoadBalancer | Delete ALB | Low | **High** — irreversible |
| ModifyLoadBalancerAttributes | Change name/description | Low | Low |
| CreateListener | Create HTTP/HTTPS/gRPC listener | Medium | Medium |
| DescribeListeners | List listeners | Low | None |
| DeleteListener | Delete listener | Low | Medium |
| ModifyListenerAttributes | Modify listener config | Medium | Medium |
| CreateRule | Create routing rule | Medium | Medium |
| DescribeRules | List routing rules | Low | None |
| DeleteRule | Delete routing rule | Low | Medium |
| CreateServerGroup | Create server group | Medium | Low |
| DescribeServerGroups | List server groups | Low | None |
| DeleteServerGroup | Delete server group | Low | Medium |
| ModifyServerGroupAttributes | Modify server group | Medium | Medium |
| AddServersToGroup | Add backend servers | Low | Medium |
| RemoveServersFromGroup | Remove backend servers | Low | Medium |
| DescribeIdleLoadBalancers | Find idle ALBs (FinOps) | Low | None |
| DescribeCostSummary | Cost estimation (FinOps) | Low | None |
| CleanupOrphanedListeners | Remove unused listeners (FinOps) | Medium | Medium |

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-31 | Initial release with ALB, listeners, rules, server groups, health checks, FinOps |
| 1.1.0 | 2026-06-04 | GCL rollout: added ## Quality Gate (GCL), references/rubric.md, references/prompt-templates.md |

## Quality Gate (GCL)

> Mandatory. max_iter=3.

| Tier | Operations | Safety |
|---|---|---|
| **Destructive** | DeleteLoadBalancer, DeleteListener, DeleteRule, DeleteServerGroup | 1.0 |
| **State-changing** | AddBackendServers, RemoveBackendServers, ModifyListenerAttributes, ModifyRuleAttributes | 1.0 |
| **Mutating** | CreateLoadBalancer, CreateListener, CreateRule, CreateServerGroup | ≥0.5 |
| **Read-only** | DescribeLoadBalancers, DescribeListeners, DescribeRules, DescribeServerGroups | ≥0 |

Safety: DeleteLoadBalancer ALL listeners/rules/server-groups lost. ServerGroup in-use warning.

### Cross-skill: ECS→ve-ecs-ops, VPC→ve-vpc-ops, CLB→ve-clb-ops, Billing→ve-billing-ops

## Execution Flows (Agent-Readable)

Every operation: **Pre-flight → Execute → Validate → Recover**. Do not skip phases.

**Preference hint:** CLI is preferred for coverage and simplicity; Go SDK is used for operations CLI does not expose.

---

### Operation: CreateLoadBalancer — Create ALB Instance

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| VPC exists | `ve vpc DescribeVpcs` | VPC found | HALT; create VPC first via `ve-vpc-ops` |
| Subnet exists | `ve vpc DescribeSubnets` | Subnet found | HALT; create subnet first |
| Credentials | `test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY"` | Both set | HALT; configure credentials |
| Region | Verify `{{user.region}}` is valid ALB region | Region supported | Suggest valid region |
| Quota | `ve alb DescribeLoadBalancers` | Under quota | HALT; request quota increase |
| Name unique | ALB name not in use | No conflict | Use different name |

#### Execution — CLI (`ve`) (Primary Path)

```bash
# Create a private ALB (internal)
ve alb CreateLoadBalancer \
  --Region "{{user.region}}" \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --LoadBalancerName "{{user.alb_name}}" \
  --Type "private"

# Create a public ALB (internet-facing)
ve alb CreateLoadBalancer \
  --Region "{{user.region}}" \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --LoadBalancerName "{{user.alb_name}}" \
  --Type "public" \
  --EipBillingConfig '{"Bandwidth":10,"EipBillingType":"PayByTraffic","ISP":"BGP"}'
```

> **Type Values:** `public` (internet-facing with EIP) or `private` (internal VPC only)

#### Execution — JIT Go SDK (Fallback)

```go
package main

import (
    "fmt"
    "os"
    "github.com/volcengine/volc-sdk-golang/service/alb"
)

func main() {
    instance := alb.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    resp, err := instance.CreateLoadBalancer(&alb.CreateLoadBalancerInput{
        VpcId:            "{{user.vpc_id}}",
        SubnetId:         "{{user.subnet_id}}",
        LoadBalancerName: "{{user.alb_name}}",
        Type:             "private",
    })
    if err != nil {
        fmt.Fprintf(os.Stderr, "CreateLoadBalancer failed: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(resp.Result.LoadBalancerId)
}
```

#### Post-execution Validation

```bash
# Parse ALB ID
ALB_ID=$(echo "$RESPONSE" | jq -r '.Result.LoadBalancerId')

# Poll until active
for i in $(seq 1 60); do
  STATUS=$(ve alb DescribeLoadBalancers --Region "{{user.region}}" --LoadBalancerIds "[\"$ALB_ID\"]" | jq -r '.Result.LoadBalancers[0].Status')
  echo "Status: $STATUS (attempt $i/60)"
  [ "$STATUS" = "active" ] && break
  sleep 5
done

# Report details
echo "ALB created: $ALB_ID"
ve alb DescribeLoadBalancers --Region "{{user.region}}" --LoadBalancerIds "[\"$ALB_ID\"]" | jq '.Result.LoadBalancers[0]'
```

#### Failure Recovery

| Error pattern | Max retries | Backoff | Agent Action | UX Feedback |
|--------------|-------------|---------|--------------|-------------|
| `InvalidVpc.NotFound` | 0 | — | HALT; VPC does not exist | `[ERROR] VPC not found. Create VPC via ve-vpc-ops first.` |
| `InvalidSubnet.NotFound` | 0 | — | HALT; subnet does not exist | `[ERROR] Subnet not found. Create subnet first.` |
| `InvalidParameter` | 1 | — | Fix args from error message | `[ERROR] InvalidParameter: The request parameter is invalid.` |
| `QuotaExceeded.LoadBalancer` | 0 | — | HALT; request quota increase | `[ERROR] ALB quota exceeded. Request quota increase from support.` |
| `InsufficientBalance` | 0 | — | HALT; recharge account | `[ERROR] Insufficient balance. Recharge your account.` |
| `Throttling` / 429 | 3 | 1s, 2s, 4s | Back off and retry | `⚠️ Rate limit reached. Retrying in {backoff}s...` |
| `InternalError` | 3 | 2s, 4s, 8s | Retry; then HALT | `[ERROR] InternalError. Retrying...` |

---

### Operation: DescribeLoadBalancers — List/Describe ALB Instances

#### Execution — CLI (`ve`)

```bash
# List all ALBs
ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}"

# Filter by ID
ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerIds "[\"{{user.alb_id}}\"]"

# Filter by name
ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerName "{{user.alb_name}}"

# Filter by VPC
ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" --VpcId "{{user.vpc_id}}"

# Filter by type
ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" --Type "public"
```

#### Present to User

| Field | JSON Path | Notes |
|-------|-----------|-------|
| ID | `.Result.LoadBalancers[].LoadBalancerId` | Format `alb-xxxxxxxxx` |
| Name | `.Result.LoadBalancers[].LoadBalancerName` | User-defined |
| Type | `.Result.LoadBalancers[].Type` | `public` or `private` |
| Status | `.Result.LoadBalancers[].Status` | Human-readable state |
| Address | `.Result.LoadBalancers[].Address` | Internal VPC IP |
| EIP | `.Result.LoadBalancers[].EipAddress` | Public IP (public type) |
| VPC | `.Result.LoadBalancers[].VpcId` | Parent VPC |
| Created | `.Result.LoadBalancers[].CreateTime` | ISO 8601 |

---

### Operation: DeleteLoadBalancer — Delete ALB Instance

#### Pre-flight (Safety Gate)

1. **MUST** obtain explicit user confirmation: irreversible delete of `{{user.alb_name}}` (`{{user.alb_id}}`)
2. **MUST NOT** proceed without clear user assent
3. **WARN** about downstream impact: all listeners, rules, and server groups will be affected
4. Verify no active listeners:

```bash
# Check for listeners before deletion
ve alb DescribeListeners --Region "{{user.region}}" --LoadBalancerId "{{user.alb_id}}"

# If listeners exist, warn user and delete them
for listener_id in $(ve alb DescribeListeners --Region "{{user.region}}" --LoadBalancerId "{{user.alb_id}}" | jq -r '.Result.Listeners[].ListenerId'); do
  echo "WARNING: Listener $listener_id will be deleted"
  ve alb DeleteListener --Region "{{user.region}}" --ListenerId "$listener_id"
done

# For public ALB, check EIP association
ve alb DescribeLoadBalancers --Region "{{user.region}}" --LoadBalancerIds "[\"{{user.alb_id}}\"]" | jq '.Result.LoadBalancers[0].EipAddress'
```

#### Execution — CLI (`ve`)

```bash
ve alb DeleteLoadBalancer --Region "{{user.region}}" --LoadBalancerId "{{user.alb_id}}"
```

#### Post-execution Validation

```bash
# Poll until ALB is deleted (NotFound)
for i in $(seq 1 60); do
  if ve alb DescribeLoadBalancers --Region "{{user.region}}" --LoadBalancerIds "[\"{{user.alb_id}}\"]" 2>&1 | grep -q "NotFound"; then
    echo "ALB deleted successfully"
    break
  fi
  echo "Waiting for deletion... (attempt $i/60)"
  sleep 5
done
```

#### Failure Recovery

| Error pattern | Max retries | Backoff | Agent Action | UX Feedback |
|--------------|-------------|---------|--------------|-------------|
| `InvalidLoadBalancer.NotFound` | 0 | — | Already deleted; skip | `[INFO] ALB already deleted.` |
| `DependencyViolation.Listener` | 0 | — | HALT; delete listeners first | `[ERROR] Listeners still attached. Delete listeners first.` |
| `IncorrectStatus.LoadBalancer` | 3 | 10s | Wait for stable status | `[WARNING] ALB not in deletable state. Waiting...` |
| `Forbidden.RAM` | 0 | — | HALT; check IAM permissions | `[ERROR] IAM permission denied. Check Volcengine IAM policies.` |

---

### Operation: CreateListener — Create ALB Listener

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| ALB exists | `ve alb DescribeLoadBalancers` | Status `active` | HALT |
| Port available | `ve alb DescribeListeners` | Port not in use | Use different port |
| Certificate (HTTPS) | Verify cert exists | Certificate valid | HALT; create certificate first |

#### Execution — CLI (`ve`)

```bash
# Create HTTP listener
ve alb CreateListener \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.alb_id}}" \
  --Protocol "HTTP" \
  --Port "{{user.listener_port}}" \
  --ListenerName "{{user.listener_name}}"

# Create HTTPS listener (with certificate)
ve alb CreateListener \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.alb_id}}" \
  --Protocol "HTTPS" \
  --Port "443" \
  --ListenerName "{{user.listener_name}}" \
  --CertificateId "{{user.certificate_id}}" \
  --TLSPolicy "tls-1-2"

# Create HTTPS listener with server group association
ve alb CreateListener \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.alb_id}}" \
  --Protocol "HTTPS" \
  --Port "443" \
  --ListenerName "https-listener" \
  --CertificateId "{{user.certificate_id}}" \
  --TLSPolicy "tls-1-2" \
  --ServerGroupId "{{user.server_group_id}}"
```

**Protocol Values:** `HTTP`, `HTTPS`, `gRPC`

#### Execution — JIT Go SDK (Fallback)

```go
package main

import (
    "fmt"
    "os"
    "github.com/volcengine/volc-sdk-golang/service/alb"
)

func main() {
    instance := alb.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    resp, err := instance.CreateListener(&alb.CreateListenerInput{
        LoadBalancerId: "{{user.alb_id}}",
        Protocol:       "HTTP",
        Port:           80,
        ListenerName:   "{{user.listener_name}}",
    })
    if err != nil {
        fmt.Fprintf(os.Stderr, "CreateListener failed: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(resp.Result.ListenerId)
}
```

#### Post-execution Validation

```bash
# Verify listener creation
echo "Listener created:"
ve alb DescribeListeners --Region "{{user.region}}" --LoadBalancerId "{{user.alb_id}}" | jq -r '.Result.Listeners[] | "\(.ListenerId) \(.Protocol):\(.Port)"'
```

#### Failure Recovery

| Error pattern | Max retries | Backoff | Agent Action | UX Feedback |
|--------------|-------------|---------|--------------|-------------|
| `InvalidLoadBalancer.NotFound` | 0 | — | HALT; ALB does not exist | `[ERROR] ALB not found. Verify the ALB ID.` |
| `PortConflict.Listener` | 0 | — | HALT; port already in use | `[ERROR] Port conflict. Use a different port.` |
| `InvalidCertificate.NotFound` | 0 | — | HALT; cert does not exist | `[ERROR] Certificate not found. Create cert first.` |
| `QuotaExceeded.Listener` | 0 | — | HALT; listener quota exceeded | `[ERROR] Listener quota exceeded.` |

---

### Operation: CreateServerGroup — Create Server Group

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| ALB exists | `ve alb DescribeLoadBalancers` | Status `active` | HALT |
| Name unique | `ve alb DescribeServerGroups` | Name not in use | Use different name |

#### Execution — CLI (`ve`)

```bash
# Create server group with health check
ve alb CreateServerGroup \
  --Region "{{user.region}}" \
  --ServerGroupName "{{user.server_group_name}}" \
  --ServerGroupType "instance" \
  --Description "Production server group" \
  --HealthCheckConfig '{
    "Enabled": true,
    "Protocol": "HTTP",
    "Method": "GET",
    "Uri": "/health",
    "Timeout": 3,
    "Interval": 5,
    "HealthyThreshold": 3,
    "UnhealthyThreshold": 3,
    "HealthyHttpCode": "200"
  }'

# Create server group (TCP health check)
ve alb CreateServerGroup \
  --Region "{{user.region}}" \
  --ServerGroupName "{{user.server_group_name}}" \
  --ServerGroupType "instance" \
  --HealthCheckConfig '{
    "Enabled": true,
    "Protocol": "TCP",
    "Timeout": 3,
    "Interval": 5,
    "HealthyThreshold": 3,
    "UnhealthyThreshold": 3
  }'
```

**ServerGroupType Values:** `instance` (ECS instance), `ip` (IP address)

#### Post-execution Validation

```bash
SG_ID=$(echo "$RESPONSE" | jq -r '.Result.ServerGroupId')
echo "Server Group created: $SG_ID"

# Verify server group details
ve alb DescribeServerGroups --Region "{{user.region}}" --ServerGroupIds "[\"$SG_ID\"]" | jq '.Result.ServerGroups[0]'
```

---

### Operation: CreateRule — Create Routing Rule

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Listener exists | `ve alb DescribeListeners` | Listener found | HALT |
| Server group exists | `ve alb DescribeServerGroups` | Server group found | HALT; create server group first |

#### Execution — CLI (`ve`)

```bash
# Path-based routing rule
ve alb CreateRule \
  --Region "{{user.region}}" \
  --ListenerId "{{user.listener_id}}" \
  --ServerGroupId "{{user.server_group_id}}" \
  --Domain "{{user.domain}}" \
  --Url "{{user.path_pattern}}" \
  --RuleName "api-route"

# Host-based routing rule only
ve alb CreateRule \
  --Region "{{user.region}}" \
  --ListenerId "{{user.listener_id}}" \
  --ServerGroupId "{{user.server_group_id}}" \
  --Domain "api.example.com" \
  --RuleName "host-rule"

# Path-based rule with multiple server groups (weighted)
ve alb CreateRule \
  --Region "{{user.region}}" \
  --ListenerId "{{user.listener_id}}" \
  --ServerGroupId "{{user.server_group_id}}" \
  --Url "/api/*" \
  --RuleName "api-weighted"
```

**Pattern Examples:** `/api/*`, `/images/*.jpg`, `/*`

#### Post-execution Validation

```bash
# Verify rule creation
ve alb DescribeRules --Region "{{user.region}}" --ListenerId "{{user.listener_id}}" | jq -r '.Result.Rules[] | "\(.RuleId): \(.Domain) \(.Url)"'
```

#### Failure Recovery

| Error pattern | Max retries | Backoff | Agent Action | UX Feedback |
|--------------|-------------|---------|--------------|-------------|
| `InvalidListener.NotFound` | 0 | — | HALT; listener does not exist | `[ERROR] Listener not found. Verify listener ID.` |
| `InvalidServerGroup.NotFound` | 0 | — | HALT; server group does not exist | `[ERROR] Server group not found. Create server group first.` |
| `InvalidRule.Pattern` | 0 | — | HALT; URL pattern invalid | `[ERROR] Invalid URL pattern. Use wildcard patterns like /api/*.` |
| `QuotaExceeded.Rule` | 0 | — | HALT; rule quota exceeded | `[ERROR] Rule quota exceeded.` |

---

### Operation: AddServersToGroup — Add Backend Servers to Server Group

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Server group exists | `ve alb DescribeServerGroups` | Server group found | HALT |
| ECS instances exist | `ve ecs DescribeInstances` | Instances found | HALT |

#### Execution — CLI (`ve`)

```bash
# Add ECS instances as backend servers
ve alb AddServersToGroup \
  --Region "{{user.region}}" \
  --ServerGroupId "{{user.server_group_id}}" \
  --Servers '[{"ServerId":"i-xxx","Port":8080,"Weight":100,"ServerType":"ecs"},{"ServerId":"i-yyy","Port":8080,"Weight":100,"ServerType":"ecs"}]'
```

#### Post-execution Validation

```bash
ve alb DescribeServerGroups --Region "{{user.region}}" --ServerGroupIds "[\"{{user.server_group_id}}\"]" | jq -r '.Result.ServerGroups[0].Servers[] | "\(.ServerId):\(.Port) weight=\(.Weight) status=\(.Status)"'
```

---

### Operation: DescribeServerGroups — List Server Groups

#### Execution — CLI (`ve`)

```bash
# List all server groups
ve alb DescribeServerGroups --Region "{{env.VOLCENGINE_REGION}}"

# Filter by ID
ve alb DescribeServerGroups --Region "{{env.VOLCENGINE_REGION}}" --ServerGroupIds "[\"{{user.server_group_id}}\"]"

# Filter by name
ve alb DescribeServerGroups --Region "{{env.VOLCENGINE_REGION}}" --ServerGroupName "{{user.server_group_name}}"
```

---

### Operation: DescribeRules — List Routing Rules

#### Execution — CLI (`ve`)

```bash
# List all rules for a listener
ve alb DescribeRules --Region "{{env.VOLCENGINE_REGION}}" --ListenerId "{{user.listener_id}}"
```

#### Present to User

| Field | JSON Path | Notes |
|-------|-----------|-------|
| Rule ID | `.Result.Rules[].RuleId` | Format `rule-xxxxxxxxx` |
| Domain | `.Result.Rules[].Domain` | Host header match |
| URL | `.Result.Rules[].Url` | Path pattern match |
| ServerGroupId | `.Result.Rules[].ServerGroupId` | Target server group |
| RuleName | `.Result.Rules[].RuleName` | User-defined name |

---

### Operation: ModifyLoadBalancerAttributes — Modify ALB Attributes

#### Execution — CLI (`ve`)

```bash
# Rename ALB
ve alb ModifyLoadBalancerAttributes \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.alb_id}}" \
  --LoadBalancerName "{{user.alb_name}}"

# Modify description
ve alb ModifyLoadBalancerAttributes \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.alb_id}}" \
  --Description "Updated description"
```

---

### Operation: DeleteListener — Delete Listener

#### Pre-flight (Safety Gate)

- **MUST** obtain confirmation before deleting a listener
- **MUST** check for attached rules and warn user

```bash
# Check for attached rules
ve alb DescribeRules --Region "{{user.region}}" --ListenerId "{{user.listener_id}}"
```

#### Execution

```bash
ve alb DeleteListener --Region "{{user.region}}" --ListenerId "{{user.listener_id}}"
```

---

### Operation: DeleteRule — Delete Routing Rule

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: delete rule `{{user.rule_id}}`

#### Execution

```bash
ve alb DeleteRule --Region "{{user.region}}" --RuleId "{{user.rule_id}}"
```

---

### Operation: ModifyServerGroupAttributes — Modify Server Group

#### Execution — CLI (`ve`)

```bash
# Rename server group
ve alb ModifyServerGroupAttributes \
  --Region "{{user.region}}" \
  --ServerGroupId "{{user.server_group_id}}" \
  --ServerGroupName "new-name"

# Modify health check config
ve alb ModifyServerGroupAttributes \
  --Region "{{user.region}}" \
  --ServerGroupId "{{user.server_group_id}}" \
  --HealthCheckConfig '{"Enabled":true,"Uri":"/healthz","Interval":10}'
```

---

### Operation: RemoveServersFromGroup — Remove Backend Servers

#### Pre-flight (Safety Gate)

- **MUST** obtain confirmation before removing servers
- **MUST** warn about traffic interruption if this is the only server group member

#### Execution

```bash
ve alb RemoveServersFromGroup \
  --Region "{{user.region}}" \
  --ServerGroupId "{{user.server_group_id}}" \
  --ServerIds '[{"ServerId":"i-xxx","ServerType":"ecs"}]'
```

---

### Operation: DescribeListeners — List ALB Listeners

#### Execution — CLI (`ve`)

```bash
# List all listeners for an ALB
ve alb DescribeListeners --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerId "{{user.alb_id}}"

# Filter by listener ID
ve alb DescribeListeners --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerId "{{user.alb_id}}" --ListenerIds '["{{user.listener_id}}"]'
```

#### Present to User

| Field | JSON Path | Notes |
|-------|-----------|-------|
| Listener ID | `.Result.Listeners[].ListenerId` | Format `lsnr-xxxxxxxxx` |
| Protocol | `.Result.Listeners[].Protocol` | HTTP, HTTPS, gRPC |
| Port | `.Result.Listeners[].Port` | Listening port |
| Status | `.Result.Listeners[].Status` | `active` or `pending` |
| CertificateId | `.Result.Listeners[].CertificateId` | HTTPS only |
| ServerGroupId | `.Result.Listeners[].ServerGroupId` | Associated server group |

---

### Operation: ModifyListenerAttributes — Modify Listener Configuration

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Listener exists | `ve alb DescribeListeners` | Listener found | HALT |

#### Execution

```bash
# Rename listener
ve alb ModifyListenerAttributes \
  --Region "{{user.region}}" \
  --ListenerId "{{user.listener_id}}" \
  --ListenerName "updated-listener-name"

# Update certificate for HTTPS listener
ve alb ModifyListenerAttributes \
  --Region "{{user.region}}" \
  --ListenerId "{{user.listener_id}}" \
  --CertificateId "{{user.certificate_id}}"

# Set request timeout
ve alb ModifyListenerAttributes \
  --Region "{{user.region}}" \
  --ListenerId "{{user.listener_id}}" \
  --RequestTimeout 120
```

#### Validation

```bash
ve alb DescribeListeners --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerId "{{user.alb_id}}" --ListenerIds '["{{user.listener_id}}"]' | jq '.Result.Listeners[0]'
```

#### Failure Recovery

| Error pattern | Max retries | Backoff | Agent Action | UX Feedback |
|--------------|-------------|---------|--------------|-------------|
| `InvalidListener.NotFound` | 0 | — | HALT; listener does not exist | `[ERROR] Listener not found. Verify listener ID.` |
| `InvalidCertificate.NotFound` | 0 | — | HALT; cert does not exist | `[ERROR] Certificate not found.` |

---

### Operation: DeleteServerGroup — Delete Server Group

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation before deleting a server group
- **MUST** warn about traffic interruption if the server group is associated with active listeners
- **MUST** check if any listeners reference this server group

```bash
# Check which listeners reference this server group
ve alb DescribeListeners --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerId "{{user.alb_id}}" | jq -r '.Result.Listeners[] | select(.ServerGroupId == "{{user.server_group_id}}") | "Listener: \(.ListenerId) on \(.Protocol):\(.Port)"'
```

#### Execution

```bash
# ⚠️ Ensure no listeners reference this server group before deletion
ve alb DeleteServerGroup --Region "{{user.region}}" --ServerGroupId "{{user.server_group_id}}"
```

#### Post-execution Validation

```bash
# Verify server group is deleted
ve alb DescribeServerGroups --Region "{{env.VOLCENGINE_REGION}}" --ServerGroupIds '["{{user.server_group_id}}"]' 2>&1 | grep -q "InvalidServerGroup.NotFound" && echo "Server group deleted successfully" || echo "Server group still exists"
```

#### Failure Recovery

| Error pattern | Max retries | Backoff | Agent Action | UX Feedback |
|--------------|-------------|---------|--------------|-------------|
| `InvalidServerGroup.NotFound` | 0 | — | Already deleted; skip | `[INFO] Server group already deleted.` |
| `DependencyViolation` | 0 | — | HALT; listeners still reference this group | `[ERROR] Listeners still reference this server group. Remove associations first.` |

---

### FinOps Operation: DescribeIdleLoadBalancers — Find Idle ALBs

#### Execution

```bash
# List all ALBs and check for idle ones (no listeners, no traffic)
ALL_ALBS=$(ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.LoadBalancers[] | select(.Status == "active") | .LoadBalancerId')

for ALB_ID in $ALL_ALBS; do
  LISTENER_COUNT=$(ve alb DescribeListeners --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerId "$ALB_ID" | jq '.Result.Listeners | length')
  if [ "$LISTENER_COUNT" -eq 0 ]; then
    echo "IDLE ALB: $ALB_ID - no listeners configured"
  fi
done
```

---

### FinOps Operation: DescribeCostSummary — ALB Cost Estimation

#### Execution

```bash
# List ALBs with creation time and type for cost analysis
echo "=== ALB Cost Summary ==="
echo ""
echo "| ALB ID | Name | Type | Status | Created |"
echo "|--------|------|------|--------|---------|"
ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.LoadBalancers[] | "|\(.LoadBalancerId)|\(.LoadBalancerName)|\(.Type)|\(.Status)|\(.CreateTime)|"'
```

---

### FinOps Operation: CleanupOrphanedListeners — Remove Unused Listeners

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation before deleting any listeners
- **MUST** list all affected listeners with their IDs and ports
- **MUST** allow user to exclude specific listeners

```bash
# Find listeners with no rules or server groups attached
echo "=== Orphaned Listeners ==="
ALL_ALBS=$(ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.LoadBalancers[] | .LoadBalancerId')

for ALB_ID in $ALL_ALBS; do
  ve alb DescribeListeners --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerId "$ALB_ID" | jq -r '.Result.Listeners[] | "\(.ListenerId) \(.Protocol):\(.Port)"'
done
```

#### Execution

```bash
# Delete orphaned listener
ve alb DeleteListener --Region "{{user.region}}" --ListenerId "{{user.listener_id}}"
```

---

## Error Taxonomy (≥ 10 Codes)

| Error Code | HTTP | Meaning | Max Retries | Backoff | Agent Action | UX Template |
|------------|------|---------|-------------|---------|--------------|-------------|
| `InvalidParameter` | 400 | Request parameter invalid | 1 | — | Fix parameter and retry | `[ERROR] InvalidParameter: Request parameter is invalid. Check the parameter against API docs.` |
| `InvalidLoadBalancer.NotFound` | 404 | ALB instance not found | 0 | — | HALT; verify ALB ID | `[ERROR] ALB not found. Verify the ALB instance ID.` |
| `InvalidListener.NotFound` | 404 | Listener not found | 0 | — | HALT; verify listener ID | `[ERROR] Listener not found. Verify the listener ID.` |
| `InvalidServerGroup.NotFound` | 404 | Server group not found | 0 | — | HALT; verify server group ID | `[ERROR] Server group not found. Verify the server group ID.` |
| `PortConflict.Listener` | 400 | Port already in use | 0 | — | HALT; use different port | `[ERROR] Port conflict. Choose a different port.` |
| `QuotaExceeded.LoadBalancer` | 400 | ALB quota exceeded | 0 | — | HALT; request quota increase | `[ERROR] ALB quota exceeded. Request quota increase from support.` |
| `QuotaExceeded.Listener` | 400 | Listener quota exceeded | 0 | — | HALT; request quota increase | `[ERROR] Listener quota exceeded.` |
| `QuotaExceeded.ServerGroup` | 400 | Server group quota exceeded | 0 | — | HALT; request quota increase | `[ERROR] Server group quota exceeded.` |
| `IncorrectStatus.LoadBalancer` | 400 | ALB in wrong state for operation | 3 | 10s | Wait for stable status | `[WARNING] ALB not in correct state. Waiting...` |
| `DependencyViolation.Listener` | 400 | Listeners still attached | 0 | — | HALT; delete listeners first | `[ERROR] Listeners still attached. Delete them first.` |
| `InvalidCertificate.NotFound` | 404 | TLS certificate not found | 0 | — | HALT; create cert first | `[ERROR] Certificate not found. Create the certificate first.` |
| `InvalidRule.Pattern` | 400 | URL pattern invalid | 0 | — | HALT; fix pattern | `[ERROR] Invalid URL pattern. Use wildcard patterns like /api/*.` |
| `Forbidden.RAM` | 403 | IAM permission denied | 0 | — | HALT; check IAM policies | `[ERROR] IAM permission denied. Check Volcengine IAM policies.` |
| `Throttling` | 429 | Rate limit exceeded | 3 | 1s, 2s, 4s | Back off and retry | `⚠️ Rate limit reached. Retrying in {backoff}s...` |
| `InternalError` | 500 | Server-side error | 3 | 2s, 4s, 8s | Retry with backoff | `[ERROR] InternalError: Server-side error occurred. Retrying...` |
| `InsufficientBalance` | 400 | Account balance insufficient | 0 | — | HALT; recharge account | `[ERROR] Insufficient balance. Recharge your account.` |

## Prerequisites

1. **Install `ve` CLI** (primary execution path — static Go binary, no runtime dependencies):

   ```bash
   # Download from GitHub releases
   curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/v1.0.42/ve-darwin-arm64 -o /usr/local/bin/ve
   chmod +x /usr/local/bin/ve
   ve version
   ```

2. **Configure Credentials** — Environment variables (recommended for Agent execution):

   ```bash
   export VOLCENGINE_ACCESS_KEY="{{env.VOLCENGINE_ACCESS_KEY}}"
   export VOLCENGINE_SECRET_KEY="{{env.VOLCENGINE_SECRET_KEY}}"
   export VOLCENGINE_REGION="{{env.VOLCENGINE_REGION}}"
   ```

3. **Verify Configuration**:

   ```bash
   ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}"
   ```

## Reference Directory

- [Core Concepts](references/core-concepts.md) — ALB architecture, types, routing
- [API & SDK Usage](references/api-sdk-usage.md) — ALB API reference and Go SDK examples
- [CLI Usage](references/cli-usage.md) — `ve alb` CLI command reference
- [Troubleshooting Guide](references/troubleshooting.md) — Common issues and diagnostics
- [Monitoring & Alerts](references/monitoring.md) — Key metrics and alarm configuration
- [Integration](references/integration.md) — ALB integration with VPC, EIP, ECS
- [Knowledge Base](references/knowledge-base.md) — Fault pattern library
- [GCL Rubric](references/rubric.md)
- [GCL Prompt Templates](references/prompt-templates.md)

## Operational Best Practices

- **Listener Planning:** Use separate listeners for HTTP (80) and HTTPS (443); plan routing rules before creating listeners
- **Content-Based Routing:** Use host-based rules for multi-tenant apps, path-based rules for microservices
- **Server Groups:** Create separate server groups for different services (e.g., `web-servers`, `api-servers`)
- **Health Checks:** Always enable health checks with correct path and interval; `/health` or `/healthz` recommended
- **HTTPS Termination:** Use TLS 1.2+ only in production; store certificates in Volcengine Certificate Manager
- **Backend Weighting:** Use weight=100 for equal distribution; weight=0 for graceful removal
- **Naming Convention:** Use consistent naming: `{project}-{env}-{service}-alb`
- **Security Groups:** Ensure backend ECS security groups allow traffic from ALB subnets on backend ports
- **Idempotency:** Use ClientToken for create operations to prevent duplicate resources
- **FinOps:** Regularly audit idle ALBs (no listeners); delete orphaned listeners and unused server groups

---

*This skill is part of the [ve-skills](https://github.com/volcengine/ve-skills) collection.*
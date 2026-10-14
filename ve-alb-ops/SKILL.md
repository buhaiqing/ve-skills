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

> **UX Compliance:** This skill follows the [User Experience Specification](../ve-skill-generator/references/user-experience-spec.md).

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

| Operation | Description | Complexity | Risk |
|-----------|-------------|------------|------|
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

# See Execution Flows in references/execution-flows.md

## Error Taxonomy (>= 10 Codes)

| Error Code | Action | Recovery |
|------------|--------|----------|
| `InvalidParameter` | Fix param & retry (1x) | `[ERROR] InvalidParameter: Request parameter is invalid. Check against API docs.` |
| `InvalidLoadBalancer.NotFound` | HALT; verify ALB ID | `[ERROR] ALB not found. Verify the ALB instance ID.` |
| `InvalidListener.NotFound` | HALT; verify listener ID | `[ERROR] Listener not found. Verify the listener ID.` |
| `InvalidServerGroup.NotFound` | HALT; verify server group ID | `[ERROR] Server group not found. Verify the server group ID.` |
| `PortConflict.Listener` | HALT; use different port | `[ERROR] Port conflict. Choose a different port.` |
| `QuotaExceeded.LoadBalancer` | HALT; request quota increase | `[ERROR] ALB quota exceeded. Request quota increase from support.` |
| `QuotaExceeded.Listener` | HALT; request quota increase | `[ERROR] Listener quota exceeded.` |
| `QuotaExceeded.ServerGroup` | HALT; request quota increase | `[ERROR] Server group quota exceeded.` |
| `IncorrectStatus.LoadBalancer` | Wait 10s; retry (3x) | `[WARNING] ALB not in correct state. Waiting...` |
| `DependencyViolation.Listener` | HALT; delete listeners first | `[ERROR] Listeners still attached. Delete them first.` |
| `InvalidCertificate.NotFound` | HALT; create cert first | `[ERROR] Certificate not found. Create the certificate first.` |
| `InvalidRule.Pattern` | HALT; fix pattern | `[ERROR] Invalid URL pattern. Use wildcard patterns like /api/*.` |
| `Forbidden.RAM` | HALT; check IAM policies | `[ERROR] IAM permission denied. Check Volcengine IAM policies.` |
| `Throttling` | Backoff 1s/2s/4s; retry (3x) | `WARNING Rate limit reached. Retrying...` |
| `InternalError` | Backoff 2s/4s/8s; retry (3x) | `[ERROR] InternalError: Server-side error occurred. Retrying...` |
| `InsufficientBalance` | HALT; recharge account | `[ERROR] Insufficient balance. Recharge your account.` |
| `HealthCheckFailed` | Retry 3x/10s; check backend | `[ERROR] Backend health check failed. Verify backend process and security group rules.` |
| `BackendUnavailable` | Retry 3x/5s; verify backend | `[ERROR] Backend server not ready. Verify backend instance state.` |

## Prerequisites

1. **Install `ve` CLI** (primary execution path — static Go binary, no runtime dependencies):

   ```bash
   # Download from GitHub releases
   curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/{{env.ve_version}}/ve-darwin-arm64 -o /usr/local/bin/ve
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
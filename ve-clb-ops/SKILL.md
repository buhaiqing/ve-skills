---
name: ve-clb-ops
description: >-
  Use when the user needs to deploy, configure, or troubleshoot Volcengine (火山引擎)
  CLB (负载均衡) — Load Balancer instances, listeners, health checks, and backend
  servers. User mentions CLB, 负载均衡, LoadBalancer, listener, 监听器, backend
  server, 后端服务器, or describes traffic distribution, health check, port
  forwarding scenarios. Not for billing, IAM, or compute resources.
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
  api_profile: "CLB API 2020-04-01"
  cli_applicability: dual-path
  cli_support_evidence: >-
    Confirmed via `ve clb --help` — CLB is supported by the ve CLI.
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

> This skill follows the [Agent Skill OpenSpec](https://agentskills.io/specification).

# Volcengine CLB Operations Skill

## Overview

CLB (负载均衡 / Classic Load Balancer) on Volcengine (火山引擎) distributes incoming traffic across multiple backend servers for high availability and scalability. This skill is an **operational runbook** for agents: explicit scope, credential rules, pre-flight checks, **dual-path execution** (official **SDK/API** and **`ve` CLI**), response validation, and failure recovery.

> **UX Compliance:** This skill follows the [User Experience Specification](../ve-skill-generator/references/user-experience-spec.md).

### CLI applicability

- **`cli_applicability: dual-path`:** Official `ve` CLI supports CLB. You **MUST** document **both** the SDK step **and** the `ve` CLI step for every operation.

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use conditions with precise triggers and delegation rules |
| 2 | **Structured I/O** | `{{env.VOLCENGINE_*}}` (env), `{{user.*}}` (interactive), `{{output.*}}` (API response) |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | Error taxonomy with ≥ 10 CLB-specific codes; HALT vs retry per type |
| 5 | **Absolute Single Responsibility** | One product (CLB); cross-product delegation documented |

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- User mentions "Volcengine CLB", "火山引擎 CLB", "负载均衡", "LoadBalancer"
- Task involves lifecycle operations on **CLB**: CreateLoadBalancer, DescribeLoadBalancers, DeleteLoadBalancer, ModifyLoadBalancerAttributes
- Task involves **Listeners**: CreateListener, DescribeListeners, DeleteListener, ModifyListenerAttributes
- Task involves **Backend Servers**: AddBackendServers, DescribeBackendServers, RemoveBackendServers
- Task involves **Health Checks**: SetHealthCheckConfig
- Task keywords: 流量分发, 负载均衡, traffic distribution, health check, listener, 后端服务器

### SHOULD NOT Use This Skill When

- Task is about **VPC creation** → delegate to: `ve-vpc-ops`
- Task is about **EIP allocation** → delegate to: `ve-eip-ops`
- Task is about **ECS instance creation** → delegate to: `ve-ecs-ops`
- Task is about **ALB/Advanced Load Balancer** → delegate to: `ve-alb-ops` (if not available, use Volcengine ALB API directly via `ve alb` commands)
- User insists on **console-only** flows → state limitation

### Delegation Rules

- CLB requires VPC + Subnet → verify via `ve-vpc-ops`
- CLB may need EIP (public type) → allocate via `ve-eip-ops`
- CLB backend servers are ECS → coordinate with `ve-ecs-ops`

## Variable Convention (Agent-Readable)

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | From runtime environment | NEVER ask; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | From runtime environment | NEVER ask; fail if unset |
| `{{env.VOLCENGINE_REGION}}` | From runtime environment | Default from env |
| `{{user.region}}` | User-supplied region | Ask once; reuse |
| `{{user.clb_name}}` | CLB name | Ask once; reuse |
| `{{user.clb_id}}` | CLB ID | Format `clb-xxxxxxxxx` |
| `{{user.vpc_id}}` | Parent VPC ID | Format `vpc-xxxxxxxxx` |
| `{{user.subnet_id}}` | CLB subnet | Format `subnet-xxxxxxxxx` |
| `{{user.listener_protocol}}` | Listener protocol | TCP, UDP, HTTP, HTTPS |
| `{{user.listener_port}}` | Listener port | e.g., `80`, `443`, `8080` |
| `{{user.certificate_id}}` | TLS Certificate ID | For HTTPS listeners |
| `{{user.backend_server_id}}` | Backend ECS ID | Format `i-xxxxxxxxx` |
| `{{user.backend_port}}` | Backend port | e.g., `8080` |
| `{{output.clb_id}}` | From CreateLoadBalancer response | Parse from `$.Result.LoadBalancerId` |
| `{{output.listener_id}}` | From CreateListener response | Parse from `$.Result.ListenerId` |

> **Security Warning (Credential Masking — MANDATORY):** **NEVER** log, print, or expose `VOLCENGINE_SECRET_KEY`.

## API and Response Conventions (Agent-Readable)

- **Volcengine CLB OpenAPI (2020-04-01)** is canonical.
- **Endpoint:** `clb.volcengineapi.com` (default: `open.volcengineapi.com`)
- **Errors:** Responses with `Error` object containing `Code` and `Message` fields.
- **Timestamps:** ISO 8601 format.

### Key Response Fields

| Operation | JSON Path | Type | Description |
|-----------|-----------|------|-------------|
| CreateLoadBalancer | `$.Result.LoadBalancerId` | string | CLB ID |
| DescribeLoadBalancers | `$.Result.LoadBalancers` | array | CLB list |
| DescribeLoadBalancers | `$.Result.LoadBalancers[].LoadBalancerId` | string | CLB ID |
| DescribeLoadBalancers | `$.Result.LoadBalancers[].LoadBalancerName` | string | CLB name |
| DescribeLoadBalancers | `$.Result.LoadBalancers[].Description` | string | CLB description |
| DescribeLoadBalancers | `$.Result.LoadBalancers[].VpcId` | string | Parent VPC |
| DescribeLoadBalancers | `$.Result.LoadBalancers[].Status` | string | CLB status |
| DescribeLoadBalancers | `$.Result.LoadBalancers[].Type` | string | CLB type (public/private) |
| DescribeLoadBalancers | `$.Result.LoadBalancers[].Address` | string | Internal IP address |
| CreateListener | `$.Result.ListenerId` | string | Listener ID |
| DescribeListeners | `$.Result.Listeners` | array | Listener list |
| AddBackendServers | `$.Result.BackendServers` | array | Added backend servers |
| DescribeBackendServers | `$.Result.BackendServers` | array | Backend server list |

## Quick Start

### What This Skill Does
This skill enables you to create, configure, and manage Volcengine (火山引擎) CLB Load Balancers, listeners, health checks, and backend servers using the `ve` CLI (primary) or JIT Go SDK (fallback).

### Prerequisites
- [ ] `ve` CLI installed (or Go runtime for JIT fallback)
- [ ] Credentials configured
- [ ] Region set
- [ ] VPC + Subnet already created — see `ve-vpc-ops`
- [ ] Backend ECS instances ready — see `ve-ecs-ops`

### Verify Setup
```bash
ve clb DescribeLoadBalancers --Region {{env.VOLCENGINE_REGION}}
```

### Your First Command
```bash
ve clb DescribeLoadBalancers --Region {{env.VOLCENGINE_REGION}}
```

## Capabilities at a Glance

| Operation | Description | Complexity | Risk Level |
|-----------|-------------|------------|------------|
| CreateLoadBalancer | Create CLB instance | Medium | Low |
| DescribeLoadBalancers | Query CLB list | Low | None |
| DeleteLoadBalancer | Delete CLB | Low | **High** — removes all listeners/backends |
| ModifyLoadBalancerAttributes | Change name/description | Low | Low |
| CreateListener | Create TCP/UDP/HTTP/HTTPS listener | Medium | Medium |
| DescribeListeners | List listeners | Low | None |
| DeleteListener | Delete listener | Low | Medium |
| AddBackendServers | Add backend servers | Low | Medium |
| DescribeBackendServers | List backend servers | Low | None |
| RemoveBackendServers | Remove backend servers | Low | Medium |
| SetHealthCheckConfig | Configure health check | Medium | Medium |

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-25 | Initial release with CLB, listeners, backend servers, health checks |
| 1.1.0 | 2026-06-04 | GCL rollout: added ## Quality Gate (GCL), references/rubric.md, references/prompt-templates.md |

## Quality Gate (GCL)

> Mandatory. max_iter=3.

| Tier | Operations | Safety |
|---|---|---|
| **Destructive** | DeleteLoadBalancer, DeleteListener | 1.0 |
| **State-changing** | RemoveBackendServers, AddBackendServers, SetHealthCheckConfig, ModifyLoadBalancerAttributes | 1.0 |
| **Mutating** | CreateLoadBalancer, CreateListener | ≥0.5 |
| **Read-only** | DescribeLoadBalancers, DescribeListeners, DescribeBackendServers | ≥0 |

Safety: DeleteLoadBalancer traffic cut. RemoveBackendServers >50% warn capacity drop. VOLCENGINE_SECRET_KEY never.

### Cross-skill: ECS→ve-ecs-ops, VPC→ve-vpc-ops, Billing→ve-billing-ops

## Execution Flows (Agent-Readable)

Every operation: **Pre-flight → Execute → Validate → Recover**.

### Operation: CreateLoadBalancer — Create CLB Instance

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| VPC exists | DescribeVpcs | VPC found | HALT; create VPC first |
| Subnet exists | DescribeSubnets | Subnet found | HALT; create subnet first |
| Quota | Check CLB quota | Sufficient | HALT; request quota increase |

#### Execution

```bash
ve clb CreateLoadBalancer \
  --Region "{{user.region}}" \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --LoadBalancerName "{{user.clb_name}}" \
  --Type "private"
```

**Type Values:** `public` (needs EIP) or `private` (internal only)

#### Post-execution Validation

1. Parse `{{output.clb_id}}` from response
2. Poll DescribeLoadBalancers until status is `active`
3. Report CLB ID, address, type, and creation status

---

### Operation: CreateListener — Create Listener

#### Execution (TCP)

```bash
ve clb CreateListener \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.clb_id}}" \
  --Protocol "TCP" \
  --Port "{{user.listener_port}}" \
  --ListenerName "{{user.listener_name}}"
```

#### Execution (HTTP)

```bash
ve clb CreateListener \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.clb_id}}" \
  --Protocol "HTTP" \
  --Port "80" \
  --ListenerName "http-listener"
```

#### Execution (HTTPS)

```bash
ve clb CreateListener \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.clb_id}}" \
  --Protocol "HTTPS" \
  --Port "443" \
  --ListenerName "https-listener" \
  --CertificateId "{{user.certificate_id}}" \
  --TLSPolicy "tls-1-2"
```

> **Note:** HTTPS listeners require a valid TLS certificate. Create certificates via Volcengine Certificate Manager first. `TLSPolicy` values: `tls-1-0`, `tls-1-1`, `tls-1-2`.

#### Validation

Poll DescribeListeners until the new listener appears:

```bash
ve clb DescribeListeners --Region "{{user.region}}" --LoadBalancerId "{{user.clb_id}}"
```

---

### Operation: AddBackendServers — Add Backend Servers

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| CLB exists | DescribeLoadBalancers | CLB found, status active | HALT |
| ECS instances exist | DescribeInstances | Instances found | HALT |

#### Execution

```bash
ve clb AddBackendServers \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.clb_id}}" \
  --BackendServers '[{"ServerId":"{{user.backend_server_id}}","Port":{{user.backend_port}},"Weight":100}]'
```

Backend server JSON format:
```json
[
  {
    "ServerId": "{{ecs.instance_id}}",
    "Port": {{user.backend_port}},
    "Weight": 100
  }
]
```

#### Validation

```bash
ve clb DescribeBackendServers --Region "{{user.region}}" --LoadBalancerId "{{user.clb_id}}"
```

---

### Operation: DeleteLoadBalancer — Delete CLB

#### Pre-flight (Safety Gate)

1. Verify no listeners exist:
```bash
ve clb DescribeListeners --Region "{{user.region}}" --LoadBalancerId "{{user.clb_id}}"
```
If listeners found, delete each:
```bash
ve clb DeleteListener --Region "{{user.region}}" --ListenerId "{{listener_id}}"
```

2. Verify no backend servers exist:
```bash
ve clb DescribeBackendServers --Region "{{user.region}}" --LoadBalancerId "{{user.clb_id}}"
```
If backends found, remove them:
```bash
ve clb RemoveBackendServers --Region "{{user.region}}" --LoadBalancerId "{{user.clb_id}}" --ServerIds '["{{user.backend_server_id}}"]'
```

3. For public CLB, disassociate EIP:
```bash
ve eip DisassociateEipAddress --Region "{{user.region}}" --AllocationId "{{eip_id}}" --InstanceType "ClbInstance" --InstanceId "{{user.clb_id}}"
```

4. **MUST** obtain explicit user confirmation before deletion

#### Execution

```bash
ve clb DeleteLoadBalancer --Region "{{user.region}}" --LoadBalancerId "{{user.clb_id}}"
```

---

### Operation: SetHealthCheckConfig

#### Execution

```bash
ve clb SetHealthCheckConfig \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.clb_id}}" \
  --ListenerId "{{user.listener_id}}" \
  --HealthyThreshold 3 \
  --UnhealthyThreshold 3 \
  --Interval 5 \
  --Timeout 3 \
  --HttpMethod "GET" \
  --Uri "/health" \
  --HealthyHttpCode "200,201,202,204"
```

---

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
- [GCL Rubric](references/rubric.md)
- [GCL Prompt Templates](references/prompt-templates.md)

## Operational Best Practices

- **Listener Planning:** Create separate listeners for different protocols (TCP:80, HTTP:8080)
- **Health Checks:** Always configure health checks; unconfigured health checks = no backend filtering
- **Backend Weight:** Use weight=100 for equal distribution, adjust for tiered capacity
- **Naming Convention:** Use consistent naming (project-env-clb-tier)
- **Public vs Private:** Use internal CLB for microservices, public CLB + EIP for external-facing
- **Least privilege:** Use IAM policies scoped to CLB APIs only

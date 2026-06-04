---
name: ve-nat-ops
description: >-
  Use when the user needs to create, configure, or manage Volcengine (火山引擎)
  NAT Gateway (NAT网关) — NAT Gateway instances, SNAT rules for outbound traffic,
  and DNAT rules for inbound port mapping. User mentions NAT Gateway, NAT网关,
  SNAT, DNAT, or describes private subnet internet access, port forwarding scenarios.
  Not for billing, IAM, or compute resources.
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
  api_profile: "NAT Gateway API 2020-04-01"
  cli_applicability: dual-path
  cli_support_evidence: >-
    Confirmed via `ve nat --help` or `ve natgateway --help` — NAT Gateway is supported by the ve CLI.
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

> This skill follows the [Agent Skill OpenSpec](https://agentskills.io/specification).

# Volcengine NAT Gateway Operations Skill

## Overview

NAT Gateway (NAT网关) on Volcengine (火山引擎) provides network address translation for private subnets, enabling internet access for resources without public IPs (via SNAT) and inbound port forwarding to private resources (via DNAT). This skill is an **operational runbook** for agents: explicit scope, credential rules, pre-flight checks, **dual-path execution** (official **SDK/API** and **`ve` CLI**), response validation, and failure recovery.

> **UX Compliance:** This skill follows the [User Experience Specification](../../ve-skill-generator/references/user-experience-spec.md).

### CLI applicability

- **`cli_applicability: dual-path`:** Official `ve` CLI supports NAT Gateway. You **MUST** document **both** the SDK step **and** the `ve` CLI step for every operation.

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use conditions with precise triggers and delegation rules |
| 2 | **Structured I/O** | `{{env.VOLCENGINE_*}}` (env), `{{user.*}}` (interactive), `{{output.*}}` (API response) |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | Error taxonomy with ≥ 10 NAT-specific codes; HALT vs retry per type |
| 5 | **Absolute Single Responsibility** | One product (NAT Gateway); cross-product delegation documented |

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- User mentions "Volcengine NAT", "火山引擎 NAT网关", "NAT Gateway", "SNAT", "DNAT"
- Task involves lifecycle operations on **NAT Gateways**: CreateNatGateway, DescribeNatGateways, DeleteNatGateway, ModifyNatGatewayAttribute
- Task involves **SNAT rules**: CreateSnatRule, DescribeSnatRules, DeleteSnatRule
- Task involves **DNAT rules**: CreateDnatRule, DescribeDnatRules, DeleteDnatRule
- Task keywords: 私网上网, outbound internet from private subnet, 端口映射, port forwarding

### SHOULD NOT Use This Skill When

- Task is purely billing / account management → delegate to billing ops
- Task is about **VPC creation** → delegate to: `ve-vpc-ops`
- Task is about **EIP allocation** → delegate to: `ve-eip-ops`
- Task is about **ECS instance** → delegate to: `ve-ecs-ops`
- User insists on **console-only** flows → state limitation

### Delegation Rules

- NAT Gateway requires VPC + Subnet → verify via `ve-vpc-ops`
- NAT Gateway requires EIP (for SNAT) → allocate via `ve-eip-ops`
- DNAT target is ECS → coordinate with `ve-ecs-ops`

## Variable Convention (Agent-Readable)

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_REGION}}` | From runtime environment | Use documented default only if skill explicitly allows |
| `{{user.region}}` | User-supplied region | Ask once; reuse; default from env |
| `{{user.nat_name}}` | User-supplied NAT name | Ask once; reuse |
| `{{user.nat_id}}` | User-supplied NAT ID | Format `ngw-xxxxxxxxx` |
| `{{user.vpc_id}}` | Parent VPC ID | Ask once; format `vpc-xxxxxxxxx` |
| `{{user.subnet_id}}` | NAT deployment subnet | Format `subnet-xxxxxxxxx` |
| `{{user.snat_cidr}}` | Source CIDR for SNAT | e.g., `10.0.2.0/24` |
| `{{user.dnat_external_port}}` | DNAT external port | e.g., `8080` |
| `{{user.dnat_internal_port}}` | DNAT internal port | e.g., `80` |
| `{{user.dnat_internal_ip}}` | DNAT target internal IP | e.g., `10.0.2.10` |
| `{{user.snat_rule_name}}` | SNAT rule name | Ask once; reuse |
| `{{user.dnat_rule_name}}` | DNAT rule name | Ask once; reuse |
| `{{output.nat_id}}` | From CreateNatGateway response | Parse from `$.Result.NatGatewayId` |
| `{{output.snat_rule_id}}` | From CreateSnatRule response | Parse from `$.Result.SnatRuleId` |
| `{{output.dnat_rule_id}}` | From CreateDnatRule response | Parse from `$.Result.DnatRuleId` |

> **Security Warning (Credential Masking — MANDATORY):** **NEVER** log, print, or expose `VOLCENGINE_SECRET_KEY` or any credential field value.

## API and Response Conventions (Agent-Readable)

- **Volcengine NAT Gateway OpenAPI (2020-04-01)** is canonical.
- **Endpoint:** `nat.volcengineapi.com` (default: `open.volcengineapi.com`)
- **Errors:** Responses with `Error` object containing `Code` and `Message` fields.
- **Timestamps:** ISO 8601 format.

### Key Response Fields

| Operation | JSON Path | Type | Description |
|-----------|-----------|------|-------------|
| CreateNatGateway | `$.Result.NatGatewayId` | string | NAT Gateway ID |
| DescribeNatGateways | `$.Result.NatGateways` | array | NAT list |
| DescribeNatGateways | `$.Result.NatGateways[].NatGatewayId` | string | NAT ID |
| DescribeNatGateways | `$.Result.NatGateways[].NatGatewayName` | string | NAT name |
| DescribeNatGateways | `$.Result.NatGateways[].Status` | string | NAT status |
| DescribeNatGateways | `$.Result.NatGateways[].VpcId` | string | Parent VPC |
| DescribeNatGateways | `$.Result.NatGateways[].NatGatewaySpec` | string | Spec (Small/Medium/Large) |
| CreateSnatRule | `$.Result.SnatRuleId` | string | SNAT rule ID |
| DescribeSnatRules | `$.Result.SnatRules` | array | SNAT rules |
| CreateDnatRule | `$.Result.DnatRuleId` | string | DNAT rule ID |
| DescribeDnatRules | `$.Result.DnatRules` | array | DNAT rules |

## Quick Start

### What This Skill Does
This skill enables you to create, configure, and manage Volcengine (火山引擎) NAT Gateways, SNAT rules, and DNAT rules using the `ve` CLI (primary) or JIT Go SDK (fallback).

### Prerequisites
- [ ] `ve` CLI installed (or Go runtime for JIT fallback)
- [ ] Credentials configured: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`
- [ ] Region set: `VOLCENGINE_REGION`
- [ ] VPC and Subnet already created — see `ve-vpc-ops`
- [ ] At least one EIP allocated — see `ve-eip-ops`

### Verify Setup
```bash
ve nat Gateway DescribeNatGateways --Region {{env.VOLCENGINE_REGION}}
```

### Your First Command
```bash
# List all NAT Gateways in the configured region
ve nat Gateway DescribeNatGateways --Region {{env.VOLCENGINE_REGION}}
```

## Capabilities at a Glance

| Operation | Description | Complexity | Risk Level |
|-----------|-------------|------------|------------|
| CreateNatGateway | Create NAT Gateway instance | Medium | Low |
| DescribeNatGateways | Query NAT Gateway list | Low | None |
| DeleteNatGateway | Delete NAT Gateway | Low | **High** — removes SNAT/DNAT |
| ModifyNatGatewayAttribute | Change name/spec | Low | Low |
| CreateSnatRule | Create SNAT rule (outbound internet) | Medium | Low |
| DescribeSnatRules | List SNAT rules | Low | None |
| DeleteSnatRule | Delete SNAT rule | Low | Medium |
| CreateDnatRule | Create DNAT rule (port mapping) | Medium | Medium |
| DescribeDnatRules | List DNAT rules | Low | None |
| DeleteDnatRule | Delete DNAT rule | Low | Medium |

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-25 | Initial release with NAT Gateway, SNAT, DNAT management |
| 1.1.0 | 2026-06-04 | GCL rollout: added ## Quality Gate (GCL), references/rubric.md, references/prompt-templates.md |

## Quality Gate (GCL)

> Mandatory. max_iter=3.

| Tier | Operations | Safety |
|---|---|---|
| **Destructive** | DeleteNatGateway, DeleteSnatRule, DeleteDnatRule | 1.0 |
| **State-changing** | CreateSnatRule, CreateDnatRule | 1.0 |
| **Mutating** | CreateNatGateway, ModifyNatGatewayAttribute | ≥0.5 |
| **Read-only** | DescribeNatGateways, DescribeSnatRules, DescribeDnatRules | ≥0 |

Safety: DeleteNatGateway warn internet access loss for all private subnets. VOLCENGINE_SECRET_KEY never.

### Cross-skill: VPC→ve-vpc-ops, Billing→ve-billing-ops

## Execution Flows (Agent-Readable)

Every operation: **Pre-flight → Execute (ve CLI + JIT Go SDK fallback) → Validate → Recover**.

### Operation: DescribeNatGateways — Query NAT Gateway List

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | `test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY"` | Both set | HALT |
| Region | Verify region is set | Valid region | HALT |
| CLI | `ve version` | Exit code 0 | Install ve CLI |

#### Execution — CLI (`ve`)

```bash
# List all NAT Gateways
ve nat Gateway DescribeNatGateways --Region "{{user.region}}"

# Filter by VPC
ve nat Gateway DescribeNatGateways --Region "{{user.region}}" --VpcId "{{user.vpc_id}}"

# Pagination
ve nat Gateway DescribeNatGateways --Region "{{user.region}}" --PageSize 50 --PageNumber 1
```

#### Validation

1. Check total count for matching NAT Gateways
2. Parse `$.Result.NatGateways[]` for details
3. Report NAT IDs, names, VPCs, specs, and statuses

---

### Operation: CreateNatGateway — Create NAT Gateway

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| VPC exists | DescribeVpcs with VPC ID | VPC found | HALT; create VPC first |
| Subnet exists | DescribeSubnets with Subnet ID | Subnet found | HALT; create subnet first |
| Quota | Check NAT Gateway quota | Sufficient | HALT; request quota increase |

#### Execution

```bash
ve nat Gateway CreateNatGateway \
  --Region "{{user.region}}" \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --NatGatewayName "{{user.nat_name}}" \
  --NatGatewaySpec "Small"
```

**NatGatewaySpec Values:** `Small`, `Medium`, `Large`

#### Post-execution Validation

1. Parse `{{output.nat_id}}` from response
2. Poll DescribeNatGateways until status is `Available`
3. Report NAT Gateway ID, spec, VPC, and creation status

---

### Operation: CreateSnatRule — Create SNAT Rule (Outbound Internet Access)

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| NAT Gateway exists | DescribeNatGateways | NAT found, status Available | HALT |
| EIP available | DescribeEipAddresses | At least one EIP allocated | HALT; allocate EIP first |
| Source CIDR valid | Validate CIDR format | Valid CIDR | HALT |

#### Execution

```bash
ve nat Gateway CreateSnatRule \
  --Region "{{user.region}}" \
  --NatGatewayId "{{user.nat_id}}" \
  --SourceCidr "{{user.snat_cidr}}" \
  --SnatRuleName "{{user.snat_rule_name}}" \
  --EipAddresses '["{{user.eip_address}}"]'
```

#### Validation

Poll DescribeSnatRules until the new rule appears:

```bash
ve nat Gateway DescribeSnatRules --Region "{{user.region}}" --NatGatewayId "{{user.nat_id}}"
```

---

### Operation: CreateDnatRule — Create DNAT Rule (Inbound Port Mapping)

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| NAT Gateway exists | DescribeNatGateways | NAT found, status Available | HALT |
| EIP available | DescribeEipAddresses with EIP in use by NAT | EIP found | HALT |
| IP Protocol valid | TCP or UDP | Valid protocol | HALT |

#### Execution

```bash
ve nat Gateway CreateDnatRule \
  --Region "{{user.region}}" \
  --NatGatewayId "{{user.nat_id}}" \
  --EipAddress "{{user.eip_address}}" \
  --IpProtocol "TCP" \
  --ExternalPort "{{user.dnat_external_port}}" \
  --InternalIp "{{user.dnat_internal_ip}}" \
  --InternalPort "{{user.dnat_internal_port}}" \
  --DnatRuleName "{{user.dnat_rule_name}}"
```

#### Validation

```bash
ve nat Gateway DescribeDnatRules --Region "{{user.region}}" --NatGatewayId "{{user.nat_id}}"
```

---

### Operation: DeleteNatGateway — Delete NAT Gateway

#### Pre-flight (Safety Gate)

1. Verify and delete all SNAT rules:
```bash
ve nat Gateway DescribeSnatRules --Region "{{user.region}}" --NatGatewayId "{{user.nat_id}}"
```
For each rule found, delete:
```bash
ve nat Gateway DeleteSnatRule --Region "{{user.region}}" --SnatRuleId "{{snat_rule_id}}"
```

2. Verify and delete all DNAT rules:
```bash
ve nat Gateway DescribeDnatRules --Region "{{user.region}}" --NatGatewayId "{{user.nat_id}}"
```
For each rule found, delete:
```bash
ve nat Gateway DeleteDnatRule --Region "{{user.region}}" --DnatRuleId "{{dnat_rule_id}}"
```

3. Disassociate EIP from NAT Gateway:
```bash
ve eip DisassociateEipAddress --Region "{{user.region}}" --AllocationId "{{eip_id}}"
```

4. **MUST** obtain explicit confirmation before deletion

#### Execution

```bash
ve nat Gateway DeleteNatGateway --Region "{{user.region}}" --NatGatewayId "{{user.nat_id}}"
```

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
- [GCL Rubric](references/rubric.md)
- [GCL Prompt Templates](references/prompt-templates.md)

## Operational Best Practices

- **SNAT for private subnets:** Always use SNAT rules for private subnet outbound internet access
- **DNAT for specific services:** Use DNAT for port-forwarding (e.g., HTTP/HTTPS to web servers)
- **EIP binding:** Bind EIPs to NAT Gateway before creating SNAT rules
- **Spec sizing:** Start with `Small` spec, upgrade as traffic increases
- **Naming convention:** Use consistent naming (project-env-nat-role)
- **Least privilege:** Use IAM policies scoped to NAT APIs only

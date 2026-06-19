---
name: ve-eip-ops
description: >-
  Use when the user needs to allocate, bind, or manage Volcengine (火山引擎)
  EIP (弹性公网IP) — Elastic IP address lifecycle, bandwidth management, and
  instance association. User mentions EIP, 弹性公网IP, Elastic IP, 公网IP,
  bandwidth, 带宽, or describes public IP allocation, binding, unbinding scenarios.
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
  api_profile: "EIP API 2020-04-01"
  cli_applicability: dual-path
  cli_support_evidence: >-
    Confirmed via `ve eip --help` — EIP is supported by the ve CLI.
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

> This skill follows the [Agent Skill OpenSpec](https://agentskills.io/specification).

# Volcengine EIP Operations Skill

## Overview

EIP (弹性公网IP / Elastic IP) on Volcengine (火山引擎) provides public IP addresses that can be dynamically associated with cloud resources such as ECS instances, CLB, NAT Gateways, and ENIs. This skill is an **operational runbook** for agents: explicit scope, credential rules, pre-flight checks, **dual-path execution** (official **SDK/API** and **`ve` CLI**), response validation, and failure recovery.

> **UX Compliance:** This skill follows the [User Experience Specification](../ve-skill-generator/references/user-experience-spec.md). All operations include onboarding guidance, minimal prompts, smart defaults, clear feedback, and user-friendly error handling.

### CLI applicability

- **`cli_applicability: dual-path`:** Official `ve` CLI supports EIP. You **MUST** document **both** the SDK step **and** the `ve` CLI step for every operation.

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use conditions with precise triggers and delegation rules |
| 2 | **Structured I/O** | `{{env.VOLCENGINE_*}}` (env), `{{user.*}}` (interactive), `{{output.*}}` (API response) |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | Error taxonomy with ≥ 10 EIP-specific codes; HALT vs retry per type |
| 5 | **Absolute Single Responsibility** | One product (EIP), one primary resource (EIP); cross-product delegation documented |

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- User mentions "Volcengine EIP", "火山引擎 EIP", "弹性公网IP", "EIP", "Elastic IP"
- Task involves lifecycle operations on **EIPs**: AllocateEipAddress, DescribeEipAddresses, ReleaseEipAddress
- Task involves **EIP binding**: AssociateEipAddress, DisassociateEipAddress
- Task involves **bandwidth management**: ModifyEipBandwidth, DescribeEipBandwidth
- Task keywords: 公网IP, 弹性IP, 带宽管理, IP绑定, IP解绑, bind public IP

### SHOULD NOT Use This Skill When

- Task is purely billing / account management → delegate to billing ops
- Task is IAM / permission model only → delegate to: `ve-iam-ops` (if not available, use Volcengine IAM API directly)
- Task is about **VPC creation** → delegate to: `ve-vpc-ops`
- Task is about **ECS instance creation** → delegate to: `ve-ecs-ops`
- Task is about **NAT Gateway** → delegate to: `ve-nat-ops`
- Task is about **CLB** → delegate to: `ve-clb-ops`
- User insists on **console-only** flows with no API → state limitation

### Delegation Rules

- EIP allocation needs VPC context → verify VPC exists via `ve-vpc-ops` first
- Binding EIP to ECS → coordinate with `ve-ecs-ops` for instance details
- Binding EIP to NAT Gateway → coordinate with `ve-nat-ops`
- Binding EIP to CLB → coordinate with `ve-clb-ops`

## Variable Convention (Agent-Readable)

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_REGION}}` | From runtime environment | Use documented default only if skill explicitly allows |
| `{{user.region}}` | User-supplied region | Ask once; reuse; default from env |
| `{{user.eip_name}}` | User-supplied EIP name | Ask once; reuse |
| `{{user.eip_id}}` | User-supplied EIP Allocation ID | Ask once; reuse; format `eipalloc-xxxxxxxxx` |
| `{{user.bandwidth}}` | Bandwidth value | Ask once; unit is Mbps |
| `{{user.instance_id}}` | Target instance ID (ECS/CLB/NAT) | Ask once; format varies by type |
| `{{user.instance_type}}` | Target type enum | EcsInstance / ClbInstance / Nat |
| `{{user.eip_address}}` | EIP address string | Can use `{{output.eip_address}}` from allocation |
| `{{user.new_name}}` | New EIP name | For ModifyEipAddressAttributes |
| `{{user.new_description}}` | New EIP description | For ModifyEipAddressAttributes |
| `{{output.eip_id}}` | From AllocateEipAddress response | Parse from `$.Result.AllocationId` |
| `{{output.eip_address}}` | From AllocateEipAddress response | Parse from `$.Result.EipAddress` |

> **Security Warning (Credential Masking — MANDATORY):** **NEVER** log, print, or expose `VOLCENGINE_SECRET_KEY` or any credential field value.

## API and Response Conventions (Agent-Readable)

- **Volcengine EIP OpenAPI (2020-04-01)** is canonical for path, query, body fields, enums, and response shapes.
- **Endpoint:** `eip.volcengineapi.com` (default: `open.volcengineapi.com`)
- **Errors:** Responses with `Error` object containing `Code` and `Message` fields.
- **Timestamps:** ISO 8601 format.

### Key Response Fields

| Operation | JSON Path | Type | Description |
|-----------|-----------|------|-------------|
| AllocateEipAddress | `$.Result.AllocationId` | string | EIP Allocation ID |
| AllocateEipAddress | `$.Result.EipAddress` | string | Allocated IP address |
| DescribeEipAddresses | `$.Result.EipAddresses` | array | EIP list |
| DescribeEipAddresses | `$.Result.EipAddresses[].AllocationId` | string | EIP ID |
| DescribeEipAddresses | `$.Result.EipAddresses[].EipAddress` | string | IP address |
| DescribeEipAddresses | `$.Result.EipAddresses[].Status` | string | EIP status (`Available`, `InUse`) |
| DescribeEipAddresses | `$.Result.EipAddresses[].Bandwidth` | integer | Bandwidth (Mbps) |
| DescribeEipAddresses | `$.Result.EipAddresses[].InstanceType` | string | Binding type |
| DescribeEipAddresses | `$.Result.EipAddresses[].InstanceId` | string | Bound instance ID |
| DescribeEipAddresses | `$.Result.TotalCount` | integer | Total matching EIPs |
| DescribeEipBandwidth | `$.Result.Bandwidth` | integer | Current bandwidth |

## Quick Start

### What This Skill Does
This skill enables you to allocate, bind, unbind, and manage Volcengine (火山引擎) Elastic IP addresses using the `ve` CLI (primary) or JIT Go SDK (fallback).

### Prerequisites
- [ ] `ve` CLI installed (or Go runtime for JIT fallback)
- [ ] Credentials configured: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`
- [ ] Region set: `VOLCENGINE_REGION`
- [ ] VPC already created (for EIP allocation) — see `ve-vpc-ops`

### Verify Setup
```bash
ve eip DescribeEipAddresses --Region {{env.VOLCENGINE_REGION}}
```

### Your First Command
```bash
# List all EIPs in the configured region
ve eip DescribeEipAddresses --Region {{env.VOLCENGINE_REGION}}
```

## Capabilities at a Glance

| Operation | Description | Complexity | Risk Level |
|-----------|-------------|------------|------------|
| AllocateEipAddress | Allocate new EIP | Low | Low |
| DescribeEipAddresses | Query EIP list and details | Low | ✅ None |
| ReleaseEipAddress | Release EIP (must be unbound) | Low | 🔴 **High** — irreversible |
| AssociateEipAddress | Bind EIP to instance | Low | Medium |
| DisassociateEipAddress | Unbind EIP from instance | Low | Medium |
| ModifyEipAddressAttributes | Change EIP name/description | Low | Low |
| ModifyEipBandwidth | Adjust bandwidth | Low | Medium |
| DescribeEipBandwidth | Query bandwidth info | Low | ✅ None |
| RenewEipAddress | Renew prepaid EIP | Low | Low |
| TagEipAddress | Add tags to EIP | Low | ✅ None |

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-25 | Initial release with EIP lifecycle and bandwidth management |
| 1.1.0 | 2026-06-04 | Phase 1 GCL rollout: added `## Quality Gate (GCL)` chapter, `references/rubric.md`, `references/prompt-templates.md` |

## Quality Gate (GCL)

> This chapter is **mandatory** for every execution of `ve-eip-ops`. It implements
> the Generator-Critic-Loop defined in `../../AGENTS.md` §3-§9. Read
> [`references/rubric.md`](references/rubric.md) for scoring and
> [`references/prompt-templates.md`](references/prompt-templates.md) for safety prompts.

### Operation Tiers

| Tier | Operations | `max_iter` | Safety floor |
|---|---|---|---|
| **Destructive** | `ReleaseEipAddress` | 2 | 1.0 (mandatory) |
| **State-changing** | `DisassociateEipAddress`, `AssociateEipAddress`, `ModifyEipBandwidth` | 2 | 1.0 (mandatory) |
| **Mutating** | `AllocateEipAddress`, `ModifyEipAddressAttributes`, `RenewEipAddress`, `TagEipAddress` | 2 | ≥ 0.5 |
| **Read-only** | `DescribeEipAddresses`, `DescribeEipBandwidth`, `DescribeEipAddressAttributes` | 3 | ≥ 0 |

### Loop

1. **Pre-flight** — resolve `{{env.*}}` / `{{user.*}}`; classify tier; load rubric.
2. **Generate** — execute per `## Execution Flows`. Trace to `./audit-results/gcl-trace-*.json`.
3. **Critique** — isolated prompt; score 5 dimensions; MUST NOT see raw request.
4. **Decide** — Safety=0 on Destructive/State-changing → **ABORT**; all pass → return; `iter<max_iter` → loop.

### EIP-specific safety rules

- **ReleaseEipAddress**: MUST check EIP status first (`Available` or warn auto-disassociate). Explicit confirmation naming the IP.
- **DisassociateEipAddress on production**: warn about connectivity loss.
- **AssociateEipAddress** when EIP already bound: force-rebind warning.
- **ModifyEipBandwidth** with significant increase: cost impact warning.

### Trace

`./audit-results/gcl-trace-*.json` — with `redaction_pass: true`.

### Cross-skill delegation

| Critic finding | Delegate to |
|---|---|
| ECS instance details needed | `ve-ecs-ops` |
| CLB binding | `ve-clb-ops` |
| NAT Gateway binding | `ve-nat-ops` |
| VPC network context | `ve-vpc-ops` |
| Billing/quota exceeded | `ve-billing-ops` |

## Execution Flows (Agent-Readable)

Every operation: **Pre-flight → Execute (ve CLI primary + JIT Go SDK fallback) → Validate → Recover**.

### Operation: DescribeEipAddresses — Query EIP List

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | `test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY"` | Both set | HALT; configure credentials |
| Region | Verify `{{user.region}}` or `{{env.VOLCENGINE_REGION}}` is set | Valid region | Suggest: `ve eip DescribeRegions` |
| CLI | `ve version` | Exit code 0 | Install ve CLI |

#### Execution — CLI (`ve`)

```bash
# List all EIPs (JSON output by default)
ve eip DescribeEipAddresses --Region "{{user.region}}"

# Filter by EIP ID
ve eip DescribeEipAddresses --Region "{{user.region}}" --AllocationIds '["{{user.eip_id}}"]'

# Filter by status
ve eip DescribeEipAddresses --Region "{{user.region}}" --Status "Available"

# Pagination
ve eip DescribeEipAddresses --Region "{{user.region}}" --PageSize 50 --PageNumber 1
```

#### Validation

1. Check total count for matching EIPs
2. Parse `$.Result.EipAddresses[]` for EIP details
3. Report EIP IDs, addresses, bandwidth, and binding status

#### Failure Recovery

| Error Pattern | Max Retries | Agent Action |
|--------------|-------------|--------------|
| `InvalidRegion.NotFound` | 0 | List valid regions; HALT |
| `Unauthorized` | 0 | HALT; check IAM permissions |
| `InternalError` | 3 | Retry with exponential backoff |
| Throttling / 429 | 3 | Back off (2s, 4s, 8s); retry |

---

### Operation: AllocateEipAddress — Allocate EIP

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | Env vars set | Non-empty | HALT |
| Quota | Check EIP quota per region | Sufficient | HALT; request quota increase |
| LineType | Validate line type | BGP (default) | Use valid line type |

#### Execution — CLI (`ve`)

```bash
# Allocate EIP with minimal parameters
ve eip AllocateEipAddress \
  --Region "{{user.region}}" \
  --LineType "BGP" \
  --Bandwidth "{{user.bandwidth}}" \
  --Name "{{user.eip_name}}"
```

#### Post-execution Validation

1. Parse `{{output.eip_id}}` and `{{output.eip_address}}` from response
2. Poll DescribeEipAddresses until status is `Available`
3. Report EIP ID, address, and allocation status

#### Failure Recovery

| Error Pattern | Max Retries | Agent Action | UX Feedback |
|--------------|-------------|--------------|-------------|
| `QuotaExceeded.EipAddress` | 0 | HALT | `[ERROR] QuotaExceeded: EIP quota reached. How to fix: Delete unused EIPs or request quota increase.` |
| `InvalidLineType.NotSupported` | 0 | HALT; use BGP | `[ERROR] InvalidLineType: Line type not supported. How to fix: Use BGP.` |
| `InsufficientBalance` | 0 | HALT; recharge | `[ERROR] InsufficientBalance: Account balance insufficient. How to fix: Recharge your account.` |

---

### Operation: AssociateEipAddress — Bind EIP to Instance

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| EIP exists | DescribeEipAddresses with EIP ID | EIP found, status Available | HALT; allocate EIP first |
| Instance exists | Verify target instance exists | Instance found | HALT; create instance first |
| Instance not bound | Check instance has no other EIP | Unbound | HALT; unbind existing EIP first |

#### Execution

```bash
ve eip AssociateEipAddress \
  --Region "{{user.region}}" \
  --AllocationId "{{user.eip_id}}" \
  --InstanceId "{{user.instance_id}}" \
  --InstanceType "{{user.instance_type}}"
```

**Valid InstanceType values:** `EcsInstance`, `ClbInstance`, `Nat`, `HaVip`, `NetworkInterface`

#### Validation

Poll DescribeEipAddresses until status is `InUse` and InstanceId matches.

---

### Operation: DisassociateEipAddress — Unbind EIP

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| EIP exists and bound | DescribeEipAddresses with EIP ID | EIP found, status InUse | HALT; EIP not bound |

#### Execution

```bash
ve eip DisassociateEipAddress \
  --Region "{{user.region}}" \
  --AllocationId "{{user.eip_id}}"
```

#### Validation

Poll DescribeEipAddresses until status is `Available`.

---

### Operation: ReleaseEipAddress — Release EIP

#### Pre-flight (Safety Gate)

1. Verify EIP status (must be `Available`):
```bash
ve eip DescribeEipAddresses --Region "{{user.region}}" --AllocationIds '["{{user.eip_id}}"]'
```
If status is `InUse`, disassociate first:
```bash
ve eip DisassociateEipAddress --Region "{{user.region}}" --AllocationId "{{user.eip_id}}"
```

2. **MUST** obtain explicit confirmation: irreversible release
3. **MUST NOT** proceed without clear user assent

#### Execution

```bash
ve eip ReleaseEipAddress --Region "{{user.region}}" --AllocationId "{{user.eip_id}}"
```

#### Validation

Poll DescribeEipAddresses until EIP not found (404).

---

### Operation: ModifyEipBandwidth — Adjust Bandwidth

#### Execution

```bash
ve eip ModifyEipBandwidth \
  --Region "{{user.region}}" \
  --AllocationId "{{user.eip_id}}" \
  --Bandwidth "{{user.bandwidth}}"
```

#### Validation

```bash
ve eip DescribeEipBandwidth --Region "{{user.region}}" --AllocationId "{{user.eip_id}}"
```

---

### Operation: RenewEipAddress — Renew Prepaid EIP

#### Execution

```bash
ve eip RenewEipAddress \
  --Region "{{user.region}}" \
  --AllocationId "{{user.eip_id}}" \
  --RenewalPeriod "Month" \
  --Quantity 1
```

**RenewalPeriod values:** `Month`, `Year`

#### Validation

```bash
ve eip DescribeEipAddressAttributes --Region "{{user.region}}" --AllocationId "{{user.eip_id}}"
```

---

### Operation: TagEipAddress — Add Tags to EIP

#### Execution

```bash
ve eip TagEipAddress \
  --Region "{{user.region}}" \
  --AllocationId "{{user.eip_id}}" \
  --Tags '[{"Key": "env", "Value": "production"}, {"Key": "team", "Value": "ops"}]'
```

#### Validation

```bash
ve eip DescribeEipAddressAttributes --Region "{{user.region}}" --AllocationId "{{user.eip_id}}" | jq '.Result.Tags'
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
- [GCL Rubric](references/rubric.md) — Scoring dimensions for the Generator-Critic-Loop
- [GCL Prompt Templates](references/prompt-templates.md) — G/C/O prompt skeletons + EIP-specific safety prompts

## Operational Best Practices

- **Bandwidth Planning:** Start with minimum bandwidth, scale up as needed (bandwidth changes are instant)
- **EIP Naming:** Use consistent naming (project-env-eip-role)
- **EIP Recycling:** Release unused EIPs promptly to avoid charges
- **Prepaid vs PayAsYouGo:** Choose billing type based on usage patterns
- **Least privilege:** Use IAM policies scoped to EIP APIs only

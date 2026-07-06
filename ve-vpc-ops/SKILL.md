---
name: ve-vpc-ops
description: >-
  Use when the user needs to create, configure, or manage Volcengine (火山引擎)
  VPC (私有网络) — VPC and Subnet lifecycle, Route Tables, and network topology.
  User mentions VPC, 私有网络, Subnet, 子网, Route Table, 路由表, or describes
  network isolation, CIDR planning, subnet creation, routing configuration scenarios.
  Not for billing, IAM, or compute/application resources that have their own ops skills.
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
  api_profile: "VPC API 2020-04-01"
  cli_applicability: dual-path
  cli_support_evidence: >-
    Confirmed via `ve vpc --help` — VPC is supported by the ve CLI.
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

> This skill follows the [Agent Skill OpenSpec](https://agentskills.io/specification).

# Volcengine VPC Operations Skill

## Overview

VPC (私有网络) on Volcengine (火山引擎) provides isolated network environments including VPCs, Subnets, Route Tables, and Route Entries. This skill is an **operational runbook** for agents: explicit scope, credential rules, pre-flight checks, **dual-path execution** (official **SDK/API** and **`ve` CLI**), response validation, and failure recovery.

> **UX Compliance:** This skill follows the [User Experience Specification](../ve-skill-generator/references/user-experience-spec.md). All operations include onboarding guidance, minimal prompts, smart defaults, clear feedback, and user-friendly error handling.

### CLI applicability

- **`cli_applicability: dual-path`:** Official `ve` CLI supports VPC. You **MUST** document **both** the SDK step **and** the `ve` CLI step for every operation.

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use conditions with precise triggers and delegation rules |
| 2 | **Structured I/O** | `{{env.VOLCENGINE_*}}` (env), `{{user.*}}` (interactive), `{{output.*}}` (API response) |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | Error taxonomy with ≥ 10 VPC-specific codes; HALT vs retry per type |
| 5 | **Absolute Single Responsibility** | One product (VPC), one primary resource (VPC/Subnet); cross-product delegation documented |

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- User mentions "Volcengine VPC", "火山引擎 VPC", "私有网络", "VPC"
- Task involves CRUD or lifecycle operations on **VPCs**: CreateVpc, DescribeVpcs, DeleteVpc, ModifyVpcAttribute
- Task involves **Subnets**: CreateSubnet, DescribeSubnets, DeleteSubnet
- Task involves **Route Tables**: CreateRouteTable, DescribeRouteTables, DeleteRouteTable
- Task involves **Route Entries**: CreateRouteEntry, DeleteRouteEntry
- Task keywords: CIDR, 网段, 子网, 路由, network isolation, 网络隔离

### SHOULD NOT Use This Skill When

- Task is purely billing / account management → delegate to billing ops
- Task is IAM / permission model only → delegate to: `ve-iam-ops` (when present)
- Task is about **EIP binding** → delegate to: `ve-eip-ops`
- Task is about **NAT Gateway** → delegate to: `ve-nat-ops`
- Task is about **Load Balancer** → delegate to: `ve-clb-ops`
- Task is about **ECS instance creation** → delegate to: `ve-ecs-ops`
- User insists on **console-only** flows with no API → state limitation

### Delegation Rules

- EIP binding requires VPC → verify VPC exists via this skill
- NAT Gateway requires VPC and Subnet → verify via this skill first
- CLB requires VPC and Subnet → verify via this skill first
- ECS instance requires VPC and Subnet → verify via this skill first

## Variable Convention (Agent-Readable)

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_REGION}}` | From runtime environment | Use documented default only if skill explicitly allows |
| `{{user.region}}` | User-supplied region | Ask once; reuse; default from env |
| `{{user.vpc_name}}` | User-supplied VPC name | Ask once; reuse |
| `{{user.vpc_id}}` | User-supplied VPC ID | Ask once; reuse; format `vpc-xxxxxxxxx` |
| `{{user.cidr_block}}` | User-supplied CIDR block | Ask once; format `10.0.0.0/16` |
| `{{user.subnet_name}}` | User-supplied subnet name | Ask once; reuse |
| `{{user.subnet_id}}` | User-supplied subnet ID | Format `subnet-xxxxxxxxx` |
| `{{user.subnet_cidr}}` | User-supplied subnet CIDR | Format `10.0.1.0/24` |
| `{{output.vpc_id}}` | From CreateVpc/DescribeVpcs response | Parse from `$.Result.Vpcs[].VpcId` or `$.Result.VpcId` |
| `{{output.subnet_id}}` | From CreateSubnet response | Parse from `$.Result.SubnetId` |

> **Security Warning (Credential Masking — MANDATORY):** **NEVER** log, print, or expose `VOLCENGINE_SECRET_KEY` or any credential field value.

## API and Response Conventions (Agent-Readable)

- **Volcengine VPC OpenAPI (2020-04-01)** is canonical for path, query, body fields, enums, and response shapes.
- **Endpoint:** `vpc.volcengineapi.com` (default: `open.volcengineapi.com`)
- **Errors:** Responses with `Error` object containing `Code` and `Message` fields.
- **Timestamps:** ISO 8601 format.

### Key Response Fields

| Operation | JSON Path | Type | Description |
|-----------|-----------|------|-------------|
| CreateVpc | `$.Result.VpcId` | string | Created VPC ID |
| DescribeVpcs | `$.Result.Vpcs` | array | VPC list |
| DescribeVpcs | `$.Result.Vpcs[].VpcId` | string | VPC ID |
| DescribeVpcs | `$.Result.Vpcs[].VpcName` | string | VPC name |
| DescribeVpcs | `$.Result.Vpcs[].CidrBlock` | string | CIDR block |
| DescribeVpcs | `$.Result.Vpcs[].Status` | string | VPC status |
| DescribeVpcs | `$.Result.TotalCount` | integer | Total matching VPCs |
| CreateSubnet | `$.Result.SubnetId` | string | Created subnet ID |
| DescribeSubnets | `$.Result.Subnets` | array | Subnet list |
| CreateRouteTable | `$.Result.RouteTableId` | string | Route table ID |

## Quick Start

### What This Skill Does
This skill enables you to create, configure, and manage Volcengine (火山引擎) VPCs, Subnets, and Route Tables using the `ve` CLI (primary) or JIT Go SDK (fallback).

### Prerequisites
- [ ] `ve` CLI installed (or Go runtime for JIT fallback)
- [ ] Credentials configured: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`
- [ ] Region set: `VOLCENGINE_REGION`

### Verify Setup
```bash
ve vpc DescribeVpcs --Region {{env.VOLCENGINE_REGION}}
```

### Your First Command
```bash
# List all VPCs in the configured region
ve vpc DescribeVpcs --Region {{env.VOLCENGINE_REGION}}
```

### Next Steps
- [Core Concepts](references/core-concepts.md) — Understand VPC architecture
- [Common Operations](#execution-flows) — Create, manage, and configure virtual networks
- [Troubleshooting](references/troubleshooting.md) — Fix common issues

## Capabilities at a Glance

| Operation | Description | Complexity | Risk Level |
|-----------|-------------|------------|------------|
| CreateVpc | Create a VPC with CIDR block | Medium | Low |
| DescribeVpcs | Query VPC list and details | Low | None |
| DeleteVpc | Delete a VPC (must be empty) | Low | **High** — irreversible |
| ModifyVpcAttribute | Modify VPC name/description | Low | Low |
| CreateSubnet | Create a subnet within VPC | Medium | Low |
| DescribeSubnets | Query subnet list | Low | None |
| DeleteSubnet | Delete a subnet (must be empty) | Low | **High** — irreversible |
| CreateRouteTable | Create a route table | Medium | Low |
| DescribeRouteTables | List route tables | Low | None |
| CreateRouteEntry | Add a route entry | Medium | Medium |
| DeleteRouteEntry | Remove a route entry | Low | Medium |

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-25 | Initial release with VPC, Subnet, Route Table management |
| 1.1.0 | 2026-06-04 | GCL rollout: added ## Quality Gate (GCL), references/rubric.md, references/prompt-templates.md |

## Quality Gate (GCL)

> Mandatory for every execution of `ve-vpc-ops`. Implements GCL per `../../AGENTS.md` §3-§9. Recommended tier: max_iter=3.

| Tier | Operations | max_iter | Safety |
|---|---|---|---|
| **Destructive** | DeleteVpc, DeleteSubnet | 3 | 1.0 |
| **State-changing** | CreateRouteEntry, DeleteRouteEntry, ModifyVpcAttribute | 3 | 1.0 |
| **Mutating** | CreateVpc, CreateSubnet, CreateRouteTable | 3 | ≥0.5 |
| **Read-only** | DescribeVpcs, DescribeSubnets, DescribeRouteTables | 3 | ≥0 |

Safety: DeleteVpc MUST verify empty. VOLCENGINE_SECRET_KEY never in trace.

### Cross-skill
| Finding | Delegate |
|---|---|
| Subnet/ECS | ve-ecs-ops |
| Billing | ve-billing-ops |

## Execution Flows (Agent-Readable)

Every operation: **Pre-flight → Execute (ve CLI primary + JIT Go SDK fallback) → Validate → Recover**.

### Operation: DescribeVpcs — Query VPC List

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | `test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY"` | Both set | HALT; configure credentials |
| Region | Verify `{{user.region}}` or `{{env.VOLCENGINE_REGION}}` is set | Valid region | Suggest: `ve vpc DescribeRegions` |
| CLI | `ve version` | Exit code 0 | Install ve CLI |

#### Execution — CLI (`ve`)

```bash
# List all VPCs (JSON output by default)
ve vpc DescribeVpcs --Region "{{user.region}}"

# Filter by VPC ID
ve vpc DescribeVpcs --Region "{{user.region}}" --VpcIds '["{{user.vpc_id}}"]'

# Pagination
ve vpc DescribeVpcs --Region "{{user.region}}" --MaxResults 50 --NextToken "{{user.next_token}}"
```

#### Validation

1. Check total count for matching VPCs
2. Parse `$.Result.Vpcs[]` for VPC details
3. Report VPC IDs, names, CIDR blocks, and statuses

#### Failure Recovery

| Error Pattern | Max Retries | Agent Action |
|--------------|-------------|--------------|
| `InvalidRegion.NotFound` | 0 | List valid regions; HALT |
| `Unauthorized` | 0 | HALT; check IAM permissions |
| `InternalError` | 3 | Retry with exponential backoff |
| Throttling / 429 | 3 | Back off (2s, 4s, 8s); retry |

---

### Operation: CreateVpc — Create VPC

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | Env vars set | Non-empty | HALT |
| CIDR valid | Validate CIDR format | Valid CIDR | HALT; provide valid CIDR |
| Quota | Check VPC quota per region | Sufficient | HALT; request quota increase |

#### Execution — CLI (`ve`)

```bash
# Create VPC with minimal parameters
ve vpc CreateVpc \
  --Region "{{user.region}}" \
  --CidrBlock "{{user.cidr_block}}" \
  --VpcName "{{user.vpc_name}}"
```

#### Post-execution Validation

1. Parse `{{output.vpc_id}}` from response
2. Poll DescribeVpcs until status is `Available`
3. Report VPC ID, CIDR, and creation status

#### Failure Recovery

| Error Pattern | Max Retries | Agent Action | UX Feedback |
|--------------|-------------|--------------|-------------|
| `InvalidCidrBlock.Malformed` | 0 | HALT; fix CIDR format | `[ERROR] InvalidCidrBlock: CIDR block format invalid. How to fix: Use format like 10.0.0.0/16, 192.168.0.0/16, or 172.16.0.0/12.` |
| `QuotaExceeded.Vpc` | 0 | HALT | `[ERROR] QuotaExceeded: VPC quota reached. How to fix: Delete unused VPCs or request quota increase.` |
| `CidrBlock.Conflict` | 0 | HALT; use different CIDR | `[ERROR] CidrBlockConflict: CIDR overlaps with existing VPC. How to fix: Use a different CIDR block.` |
| `InvalidVpcName.Duplicate` | 0 | HALT; use unique name | `[ERROR] DuplicateName: VPC name already exists. How to fix: Use a unique name.` |

---

### Operation: CreateSubnet — Create Subnet

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| VPC exists | DescribeVpcs with VPC ID | VPC found | HALT; create VPC first |
| CIDR within VPC CIDR | Validate subnet CIDR is subset | Within range | HALT; adjust CIDR |
| Zone available | Check zone availability | Zone available | HALT; choose different zone |

#### Execution — CLI (`ve`)

```bash
ve vpc CreateSubnet \
  --Region "{{user.region}}" \
  --VpcId "{{user.vpc_id}}" \
  --CidrBlock "{{user.subnet_cidr}}" \
  --SubnetName "{{user.subnet_name}}" \
  --ZoneId "{{user.zone_id}}"
```

#### Validation

Poll DescribeSubnets until status is `Available`.

---

### Operation: DeleteVpc — Delete VPC

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: irreversible delete
- **MUST** verify VPC has no subnets, instances, or other resources
- **MUST NOT** proceed without clear user assent

#### Execution

```bash
ve vpc DeleteVpc --Region "{{user.region}}" --VpcId "{{user.vpc_id}}"
```

#### Validation

Poll DescribeVpcs until VPC not found (404).

---

### Operation: CreateRouteTable — Create Route Table

#### Execution

```bash
ve vpc CreateRouteTable \
  --Region "{{user.region}}" \
  --VpcId "{{user.vpc_id}}" \
  --RouteTableName "{{user.route_table_name}}"
```

---

### Operation: CreateRouteEntry — Add Route Entry

#### Execution

```bash
ve vpc CreateRouteEntry \
  --Region "{{user.region}}" \
  --RouteTableId "{{user.route_table_id}}" \
  --DestinationCidrBlock "{{user.destination_cidr}}" \
  --NextHopType "{{user.next_hop_type}}" \
  --NextHopId "{{user.next_hop_id}}"
```

---

## Error Taxonomy

| Code | Description | Resolution |
|------|-------------|------------|
| `InvalidVpcId.NotFound` | 指定的 VPC ID 不存在 | 0 retries; **HALT** |
| `InvalidSubnetId.NotFound` | 指定的子网 ID 不存在 | 0 retries; **HALT** |
| `InvalidCidrBlock.Malformed` | CIDR 地址段格式错误 | 0 retries; **HALT** |
| `CidrBlock.Conflict` | CIDR 与现有 VPC/子网冲突 | 0 retries; **HALT** |
| `QuotaExceeded.Vpc` | VPC 数量超出配额限制 | 0 retries; **HALT** |
| `QuotaExceeded.Subnet` | 子网数量超出配额限制 | 0 retries; **HALT** |
| `QuotaExceeded.RouteTable` | 路由表数量超出配额限制 | 0 retries; **HALT** |
| `Vpc.InUse` | VPC 中存在子网或实例，无法删除 | 0 retries; **HALT** |
| `Subnet.InUse` | 子网中存在实例，无法删除 | 0 retries; **HALT** |
| `InvalidZoneId.NotFound` | 指定的可用区 ID 不存在 | 0 retries; **HALT** |
| `InvalidRouteTableId.NotFound` | 指定的路由表 ID 不存在 | 0 retries; **HALT** |
| `InvalidNextHopType.NotSupported` | 下一跳类型不支持 | 0 retries; **HALT** |
| `Throttling` | VPC API 请求频率超限 | 3 retries/exponential/1s/2s/4s; **RETRY** |
| `InternalError` | VPC 服务端内部错误 | 3 retries/exponential/2s/4s/8s; **RETRY** |
 
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

- **CIDR Planning:** Use non-overlapping CIDR blocks (10.0.0.0/16, 192.168.0.0/16, 172.16.0.0/12)
- **Multi-zone Subnets:** Create subnets in different zones for HA
- **Naming Convention:** Use consistent naming (project-env-vpc, project-env-subnet-zone)
- **Least privilege:** Use IAM policies scoped to VPC APIs only
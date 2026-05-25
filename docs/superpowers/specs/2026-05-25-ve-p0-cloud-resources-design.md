---
name: ve-p0-cloud-resources-design
description: Design specification for 4 P0 Volcengine cloud resource Skills (VPC, CLB, EIP, NAT Gateway)
type: project
---

# Design Specification: P0 Volcengine Cloud Resource Skills

**Date:** 2026-05-25
**Status:** Approved for Implementation
**Approach:** Parallel Generation with Shared References

---

## Overview

This specification defines 4 P0 (highest priority) Volcengine cloud resource operational Skills to be generated using the ve-skill-generator meta-skill. All 4 skills will be generated in parallel for efficiency and consistency.

### Skills Overview

| Skill | Product | Primary Resource | Chinese Name |
|-------|---------|-----------------|--------------|
| `ve-vpc-ops` | VPC | VPC, Subnet, Route Table | 私有网络 |
| `ve-clb-ops` | CLB | Load Balancer, Listener, Backend Server | 负载均衡 |
| `ve-eip-ops` | EIP | Elastic IP Address | 弹性公网IP |
| `ve-nat-ops` | NAT Gateway | NAT Gateway, SNAT/DNAT Rules | NAT网关 |

---

## Why: Infrastructure Dependencies and Operational Frequency

**Why:** These 4 resources form the **core networking infrastructure layer** that all compute and application resources depend on. They are the highest-frequency operations in DevOps workflows:

- **VPC**: Every ECS/RDS/Redis instance requires a VPC and Subnet
- **EIP**: Required for public access (CLB, NAT Gateway, ECS)
- **CLB**: Standard pattern for high-availability service deployment
- **NAT Gateway**: Required for private subnet outbound connectivity

**How to apply:** Generate these 4 skills first, before P1/P2 skills. All compute and application skills will delegate network setup to these P0 skills.

---

## Skill 1: ve-vpc-ops (VPC私有网络)

### Frontmatter

```yaml
name: ve-vpc-ops
description: >-
  Use when the user needs to create, configure, or manage Volcengine (火山引擎)
  VPC (私有网络) — VPC and Subnet lifecycle, Route Tables, and network topology.
  User mentions VPC, 私有网络, Subnet, 子网, Route Table, 路由表, or describes
  network isolation, CIDR planning, subnet creation, routing configuration scenarios.
  Not for billing, IAM, or compute/application resources that have their own ops skills.
cli_applicability: dual-path
cli_support_evidence: Confirmed via `ve vpc --help`
```

### Primary Resources

| Resource | API Prefix | Key Operations |
|----------|------------|----------------|
| VPC | `*Vpc*` | Create, Describe, Delete, Modify |
| Subnet | `*Subnet*` | Create, Describe, Delete |
| Route Table | `*RouteTable*` | Create, Describe, Delete |
| Route Entry | `*RouteEntry*` | Create, Delete |

### Core Operations (≥10)

1. `CreateVpc` — Create VPC with CIDR block
2. `DescribeVpcs` — List/query VPCs
3. `DeleteVpc` — Delete VPC (must be empty)
4. `ModifyVpcAttribute` — Change VPC name/description
5. `CreateSubnet` — Create subnet within VPC
6. `DescribeSubnets` — List/query subnets
7. `DeleteSubnet` — Delete subnet (must be empty)
8. `CreateRouteTable` — Create routing table
9. `DescribeRouteTables` — List routing tables
10. `CreateRouteEntry` — Add route rule
11. `DeleteRouteEntry` — Remove route rule

### Placeholder Conventions

| Placeholder | Type | Source |
|-------------|------|--------|
| `{{user.vpc_name}}` | string | Interactive |
| `{{user.cidr_block}}` | string | Interactive (e.g., 10.0.0.0/16) |
| `{{user.subnet_name}}` | string | Interactive |
| `{{user.subnet_cidr}}` | string | Interactive (e.g., 10.0.1.0/24) |
| `{{output.vpc_id}}` | string | API: `$.Vpcs[].VpcId` |
| `{{output.subnet_id}}` | string | API: `$.Subnets[].SubnetId` |

### Delegation Rules

- EIP binding → `ve-eip-ops`
- NAT Gateway creation → `ve-nat-ops`
- CLB creation → `ve-clb-ops`
- ECS instance → `ve-ecs-ops`

---

## Skill 2: ve-clb-ops (CLB负载均衡)

### Frontmatter

```yaml
name: ve-clb-ops
description: >-
  Use when the user needs to deploy, configure, or troubleshoot Volcengine (火山引擎)
  CLB (负载均衡) — Load Balancer instances, listeners, health checks, and backend
  servers. User mentions CLB, 负载均衡, LoadBalancer, listener, 监听器, backend
  server, 后端服务器, or describes traffic distribution, health check, port
  forwarding scenarios. Not for billing, IAM, or compute resources.
cli_applicability: dual-path
cli_support_evidence: Confirmed via `ve clb --help`
```

### Primary Resources

| Resource | API Prefix | Key Operations |
|----------|------------|----------------|
| LoadBalancer | `*LoadBalancer*` | Create, Describe, Delete, Modify |
| Listener | `*Listener*` | Create, Describe, Delete, Modify |
| BackendServer | `*BackendServer*` | Add, Describe, Remove |

### Core Operations (≥10)

1. `CreateLoadBalancer` — Create CLB instance
2. `DescribeLoadBalancers` — List/query CLBs
3. `DeleteLoadBalancer` — Delete CLB
4. `ModifyLoadBalancerAttributes` — Change name/description
5. `CreateListener` — Create TCP/UDP/HTTP listener
6. `DescribeListeners` — List listeners
7. `DeleteListener` — Delete listener
8. `AddBackendServers` — Add backend servers
9. `DescribeBackendServers` — List backend servers
10. `RemoveBackendServers` — Remove backend servers
11. `SetHealthCheckConfig` — Configure health check

### Placeholder Conventions

| Placeholder | Type | Source |
|-------------|------|--------|
| `{{user.clb_name}}` | string | Interactive |
| `{{user.listener_protocol}}` | enum | TCP/UDP/HTTP |
| `{{user.listener_port}}` | int | Interactive |
| `{{user.backend_server_id}}` | string | ECS instance ID |
| `{{output.clb_id}}` | string | API: `$.LoadBalancers[].LoadBalancerId` |
| `{{output.listener_id}}` | string | API: `$.Listeners[].ListenerId` |

### Delegation Rules

- VPC/Subnet for CLB → `ve-vpc-ops`
- EIP for public CLB → `ve-eip-ops`
- ECS backend → coordinate with `ve-ecs-ops`

---

## Skill 3: ve-eip-ops (EIP弹性公网IP)

### Frontmatter

```yaml
name: ve-eip-ops
description: >-
  Use when the user needs to allocate, bind, or manage Volcengine (火山引擎)
  EIP (弹性公网IP) — Elastic IP address lifecycle, bandwidth management, and
  instance association. User mentions EIP, 弹性公网IP, Elastic IP, 公网IP,
  bandwidth, 带宽, or describes public IP allocation, binding, unbinding scenarios.
  Not for billing, IAM, or compute resources.
cli_applicability: dual-path
cli_support_evidence: Confirmed via `ve eip --help`
```

### Primary Resources

| Resource | API Prefix | Key Operations |
|----------|------------|----------------|
| EipAddress | `*EipAddress*` | Allocate, Describe, Release, Associate, Disassociate |

### Core Operations (≥10)

1. `AllocateEipAddress` — Allocate new EIP
2. `DescribeEipAddresses` — List/query EIPs
3. `ReleaseEipAddress` — Release EIP (must be unbound)
4. `AssociateEipAddress` — Bind EIP to instance
5. `DisassociateEipAddress` — Unbind EIP
6. `ModifyEipAddressAttributes` — Change name
7. `ModifyEipBandwidth` — Adjust bandwidth
8. `DescribeEipBandwidth` — Query bandwidth info
9. `RenewEipAddress` — Renew prepaid EIP
10. `TagEipAddress` — Add tags

### Placeholder Conventions

| Placeholder | Type | Source |
|-------------|------|--------|
| `{{user.eip_name}}` | string | Interactive |
| `{{user.bandwidth}}` | int | Interactive (Mbps) |
| `{{user.instance_id}}` | string | Target instance (ECS/CLB/NAT) |
| `{{output.eip_id}}` | string | API: `$.EipAddresses[].AllocationId` |
| `{{output.eip_address}}` | string | API: `$.EipAddresses[].EipAddress` |

### Delegation Rules

- VPC for EIP → `ve-vpc-ops`
- ECS binding → coordinate with `ve-ecs-ops`
- CLB binding → coordinate with `ve-clb-ops`
- NAT Gateway binding → coordinate with `ve-nat-ops`

---

## Skill 4: ve-nat-ops (NAT网关)

### Frontmatter

```yaml
name: ve-nat-ops
description: >-
  Use when the user needs to create, configure, or manage Volcengine (火山引擎)
  NAT Gateway (NAT网关) — NAT Gateway instances, SNAT rules for outbound traffic,
  and DNAT rules for inbound port mapping. User mentions NAT Gateway, NAT网关,
  SNAT, DNAT, or describes private subnet internet access, port forwarding scenarios.
  Not for billing, IAM, or compute resources.
cli_applicability: dual-path
cli_support_evidence: Confirmed via `ve natgateway --help`
```

### Primary Resources

| Resource | API Prefix | Key Operations |
|----------|------------|----------------|
| NatGateway | `*NatGateway*` | Create, Describe, Delete, Modify |
| SnatRule | `*SnatRule*` | Create, Describe, Delete |
| DnatRule | `*DnatRule*` | Create, Describe, Delete |

### Core Operations (≥10)

1. `CreateNatGateway` — Create NAT Gateway
2. `DescribeNatGateways` — List/query NAT Gateways
3. `DeleteNatGateway` — Delete NAT Gateway
4. `ModifyNatGatewayAttribute` — Change name/spec
5. `CreateSnatRule` — Create SNAT rule (outbound)
6. `DescribeSnatRules` — List SNAT rules
7. `DeleteSnatRule` — Delete SNAT rule
8. `CreateDnatRule` — Create DNAT rule (port mapping)
9. `DescribeDnatRules` — List DNAT rules
10. `DeleteDnatRule` — Delete DNAT rule

### Placeholder Conventions

| Placeholder | Type | Source |
|-------------|------|--------|
| `{{user.nat_name}}` | string | Interactive |
| `{{user.snat_cidr}}` | string | Subnet CIDR for SNAT |
| `{{user.dnat_external_port}}` | int | External port |
| `{{user.dnat_internal_port}}` | int | Internal port |
| `{{output.nat_id}}` | string | API: `$.NatGateways[].NatGatewayId` |
| `{{output.snat_rule_id}}` | string | API: `$.SnatRules[].SnatRuleId` |

### Delegation Rules

- VPC/Subnet for NAT → `ve-vpc-ops`
- EIP for NAT → `ve-eip-ops`

---

## Common Design Elements

### Five Core Standards Compliance

All 4 skills follow the ve-skill-generator Five Core Standards:

| Standard | Implementation |
|----------|---------------|
| 1. Clear Boundaries | SHOULD/SHOULD NOT with specific triggers; delegation to other skills |
| 2. Structured I/O | Typed placeholders; JSON paths from OpenAPI |
| 3. Explicit Actionable Steps | Pre-flight → Execute → Validate → Recover per operation |
| 4. Complete Failure Strategies | ≥10 error codes per skill; HALT vs retry distinction |
| 5. Single Responsibility | One product per skill; cross-product delegation documented |

### Credential Security (Mandatory)

- NEVER log/print `VOLCENGINE_SECRET_KEY`
- Mask all credential output: `VOLCENGINE_SECRET_KEY=<masked>`
- Verify existence only: `test -n "$VOLCENGINE_SECRET_KEY"`

### Directory Structure (Each Skill)

```
ve-[product]-ops/
├── SKILL.md
├── references/
│   ├── core-concepts.md
│   ├── api-sdk-usage.md
│   ├── cli-usage.md (dual-path)
│   ├── troubleshooting.md
│   ├── monitoring.md
│   └── integration.md
└── assets/
    └── example-config.yaml
```

### Shared References (Reused from ve-skill-generator)

- `execution-environment.md` — CLI/SDK setup
- `cli-behavior.md` — ve CLI patterns
- `user-experience-spec.md` — UX requirements
- `enhanced-self-healing-framework.md` — Self-healing patterns

---

## Implementation Approach

### Parallel Generation Strategy

| Phase | Action | Parallelism | Notes |
|-------|--------|-------------|-------|
| 1 | Analyze OpenAPI for each product | 4 agents parallel | Independent research |
| 2 | Scaffold directory structure | Sequential | Single template copy |
| 3 | Populate SKILL.md (VPC first) | 1 agent | Base skill for dependencies |
| 4 | Populate SKILL.md (EIP, NAT, CLB) | 3 agents parallel | After VPC SKILL.md complete |
| 5 | Fill reference files | 4 agents parallel | All SKILL.md ready |
| 6 | Verify P0/P1 checklist | Sequential | Quality gate per skill |
| 7 | Anti-pattern check | Sequential | Final review |

### Dependencies (Order Constraint)

| Skill | Depends On | Execution Order |
|-------|------------|-----------------|
| `ve-vpc-ops` | None | **Phase 3** — First (base skill) |
| `ve-eip-ops` | VPC (for delegation reference) | **Phase 4** — After VPC |
| `ve-nat-ops` | VPC, EIP (for delegation reference) | **Phase 4** — After VPC |
| `ve-clb-ops` | VPC, EIP (for delegation reference) | **Phase 4** — After VPC |

### Estimated Time

| Approach | Total Time |
|----------|------------|
| Parallel | ~45-60 minutes |
| Sequential | ~90-120 minutes |

---

## Error Taxonomy Template (≥10 codes each skill)

| Error Pattern | Category | Action |
|--------------|----------|--------|
| `InvalidParameter.*` | Input error | HALT; fix parameter |
| `InvalidVpcId.NotFound` | Resource error | HALT; verify VPC exists |
| `QuotaExceeded.*` | Quota error | HALT; request increase |
| `ResourceAlreadyExists` | Conflict | Ask reuse vs new name |
| `IncorrectStatus.*` | Status error | HALT; check resource status |
| `Forbidden.RAM` | IAM error | HALT; check permissions |
| `InternalError` | Server error | Retry 3x with backoff |
| `Throttling` | Rate limit | Retry with exponential backoff |
| `ServiceUnavailable` | Availability | Retry; then HALT |
| `InsufficientBalance` | Billing error | HALT; recharge account |

---

## Next Steps

1. Invoke `writing-plans` skill to create detailed implementation plan
2. **Verify JSON paths** against actual OpenAPI response schemas (placeholders in this spec may not match exact API response structure)
3. Execute parallel generation for 4 skills
4. Verify P0/P1 checklist for each skill
5. Anti-pattern check for each skill
6. Final review and commit

---

**Related Skills:** [[ve-skill-generator]], [[ve-ecs-ops]]
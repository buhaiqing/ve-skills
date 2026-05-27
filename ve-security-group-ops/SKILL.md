---
name: ve-security-group-ops
description: >-
  Use when the user needs to audit, optimize, or manage Volcengine (火山引擎)
  Security Groups — rule auditing, risk detection, least-privilege optimization,
  lifecycle management, and conflict detection. User mentions Security Group,
  安全组, sg, security group rules, port exposure, overly permissive rules,
  or describes scenarios like auditing open ports, detecting 0.0.0.0/0 rules,
  finding unused security groups, or optimizing access policies. Not for VPC
  routing, NAT, or network-level configurations that have their own ops skills.
license: MIT
compatibility: >-
  Official Volcengine CLI (`ve`, Go binary, no runtime), Go 1.14+ runtime
  (for JIT SDK fallback), valid API credentials, network access to Volcengine
  endpoints (open.volcengineapi.com).
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-05-27"
  runtime: Harness AI Agent, Claude Code, Cursor, or compatible Agent runtimes
  go_version_minimum: "1.14"
  go_jit_runtime_version: "1.21+"
  api_profile: "VPC API 2020-04-01 (Security Group)"
  cli_applicability: dual-path
  cli_support_evidence: >-
    Confirmed via `ve vpc --help` — Security Group APIs are part of VPC service.
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

> This skill follows the [Agent Skill OpenSpec](https://agentskills.io/specification).

# Volcengine Security Group Operations Skill

## Overview

Security Groups (安全组) on Volcengine (火山引擎) act as virtual firewalls for ECS instances, controlling inbound and outbound traffic. This skill is an **operational runbook** for agents: security group rule auditing, risk detection, least-privilege optimization, lifecycle management, and cross-security-group conflict detection. **Do not use the web console as the primary agent execution path.**

> **UX Compliance:** This skill follows the [User Experience Specification](../../ve-skill-generator/references/user-experience-spec.md). All operations include onboarding guidance, minimal prompts, smart defaults, clear feedback, and user-friendly error handling.

### CLI applicability

- **`cli_applicability: dual-path`:** Official `ve` CLI supports Security Groups via VPC service. You **MUST** document **both** the SDK step **and** the `ve` CLI step for every operation.

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use conditions with precise triggers and delegation rules |
| 2 | **Structured I/O** | `{{env.VOLCENGINE_*}}` (env), `{{user.*}}` (interactive), `{{output.*}}` (API response) |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | Error taxonomy with security-group-specific codes; HALT vs retry per type |
| 5 | **Absolute Single Responsibility** | One product (Security Group); cross-product delegation documented |
| 6 | **FinOps Integration** | Unused security group detection, rule count optimization |
| 7 | **AIOps Integration** | Risk pattern detection, exposure analysis, automated remediation suggestions |

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- User mentions "Security Group", "安全组", "sg-", "security group rules"
- Task involves CRUD operations on **Security Groups**: CreateSecurityGroup, DescribeSecurityGroups, DeleteSecurityGroup, ModifySecurityGroupAttributes
- Task involves **Security Group Rules**: AuthorizeSecurityGroupIngress, AuthorizeSecurityGroupEgress, RevokeSecurityGroupIngress, RevokeSecurityGroupEgress
- Task involves **security auditing**: finding open ports, detecting 0.0.0.0/0 rules, identifying overly permissive policies
- Task involves **risk detection**: exposed sensitive ports (22, 3389, 3306, 6379), unrestricted outbound access
- Task involves **optimization**: least-privilege recommendations, unused rule cleanup, security group consolidation
- Task involves **conflict detection**: overlapping rules, contradictory allow/deny policies

### SHOULD NOT Use This Skill When

- Task is about **VPC/Subnet creation** → delegate to: `ve-vpc-ops`
- Task is about **ECS instance lifecycle** → delegate to: `ve-ecs-ops`
- Task is about **NAT Gateway** → delegate to: `ve-nat-ops`
- Task is about **Load Balancer** → delegate to: `ve-clb-ops`
- Task is about **Network ACL** (subnet-level) → delegate to: `ve-vpc-ops`
- Task is purely billing → delegate to billing ops

### Delegation Rules

- Security group applies to ECS instances → verify instance exists via `ve-ecs-ops`
- Security group resides in a VPC → verify VPC exists via `ve-vpc-ops`
- ECS instance creation requires security group → reference this skill for SG selection

## Variable Convention (Agent-Readable)

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_REGION}}` | From runtime environment | Use documented default only if skill explicitly allows |
| `{{user.region}}` | User-supplied region | Ask once; reuse; default from env |
| `{{user.sg_id}}` | User-supplied security group ID | Format `sg-xxxxxxxxx` |
| `{{user.sg_name}}` | User-supplied security group name | Ask once; reuse |
| `{{user.rule_id}}` | User-supplied rule ID | Format from API response |
| `{{user.vpc_id}}` | User-supplied VPC ID | Format `vpc-xxxxxxxxx` |
| `{{user.cidr}}` | User-supplied CIDR block | Format `10.0.0.0/8` or `0.0.0.0/0` |
| `{{user.port}}` | User-supplied port or port range | Format `22`, `80`, `1024/65535` |
| `{{user.protocol}}` | User-supplied protocol | `tcp`, `udp`, `icmp`, `all` |
| `{{output.sg_id}}` | From CreateSecurityGroup response | Parse from `$.Result.SecurityGroupId` |

> **Security Warning (Credential Masking — MANDATORY):** **NEVER** log, print, or expose `VOLCENGINE_SECRET_KEY`.

## API and Response Conventions (Agent-Readable)

- **Volcengine VPC OpenAPI (2020-04-01)** is canonical for Security Group APIs.
- **Endpoint:** `vpc.volcengineapi.com` (default: `open.volcengineapi.com`)

### Key Response Fields

| Operation | JSON Path | Type | Description |
|-----------|-----------|------|-------------|
| CreateSecurityGroup | `$.Result.SecurityGroupId` | string | Created SG ID |
| DescribeSecurityGroups | `$.Result.SecurityGroups` | array | SG list |
| DescribeSecurityGroups | `$.Result.SecurityGroups[].SecurityGroupId` | string | SG ID |
| DescribeSecurityGroups | `$.Result.SecurityGroups[].SecurityGroupName` | string | SG name |
| DescribeSecurityGroups | `$.Result.SecurityGroups[].VpcId` | string | VPC ID |
| DescribeSecurityGroups | `$.Result.SecurityGroups[].SecurityGroupType` | string | `enterprise` or `basic` |
| DescribeSecurityGroupAttributes | `$.Result.IngressRules` | array | Inbound rules |
| DescribeSecurityGroupAttributes | `$.Result.EgressRules` | array | Outbound rules |
| DescribeSecurityGroupAttributes | `$.Result.IngressRules[].CidrIp` | string | Source CIDR |
| DescribeSecurityGroupAttributes | `$.Result.IngressRules[].PortRange` | string | Port range |
| DescribeSecurityGroupAttributes | `$.Result.IngressRules[].IpProtocol` | string | Protocol |
| DescribeSecurityGroupAttributes | `$.Result.IngressRules[].Policy` | string | `accept` or `drop` |
| DescribeSecurityGroupAttributes | `$.Result.IngressRules[].Priority` | integer | Rule priority (1-100) |
| AuthorizeSecurityGroupIngress | `$.Result.RequestId` | string | Request ID |

## Quick Start

### What This Skill Does
This skill enables you to audit, optimize, and manage Volcengine Security Groups using the `ve` CLI (primary) or JIT Go SDK (fallback).

### Prerequisites
- [ ] `ve` CLI installed (or Go runtime for JIT fallback)
- [ ] Credentials configured: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`
- [ ] Region set: `VOLCENGINE_REGION`

### Verify Setup
```bash
ve vpc DescribeSecurityGroups --Region {{env.VOLCENGINE_REGION}} --MaxResults 1
```

### Your First Command
```bash
# List all security groups
ve vpc DescribeSecurityGroups --Region {{env.VOLCENGINE_REGION}}
```

## Capabilities at a Glance

| Operation | Description | Complexity | Risk Level |
|-----------|-------------|------------|------------|
| DescribeSecurityGroups | Query security group list | Low | None |
| CreateSecurityGroup | Create a security group | Low | Low |
| DeleteSecurityGroup | Delete a security group (must be unused) | Low | **High** |
| DescribeSecurityGroupAttributes | List all rules in a security group | Low | None |
| AuthorizeSecurityGroupIngress | Add inbound rule | Low | Medium |
| AuthorizeSecurityGroupEgress | Add outbound rule | Low | Medium |
| RevokeSecurityGroupIngress | Remove inbound rule | Low | Medium |
| RevokeSecurityGroupEgress | Remove outbound rule | Low | Medium |
| AuditSecurityGroupRules | Audit rules for risks | Medium | None |
| DetectExposedPorts | Find exposed sensitive ports | Medium | None |
| FindUnusedSecurityGroups | Identify SGs not attached to instances | Low | None |
| OptimizeToLeastPrivilege | Generate least-privilege recommendations | High | Medium |
| DetectRuleConflicts | Find conflicting/overlapping rules | Medium | None |

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-27 | Initial release with SG lifecycle, auditing, risk detection, optimization |

## Execution Flows (Agent-Readable)

### Operation: DescribeSecurityGroups — Query Security Group List

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | `test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY"` | Both set | HALT |
| Region | Verify `{{user.region}}` or `{{env.VOLCENGINE_REGION}}` is set | Valid region | Suggest: `ve vpc DescribeRegions` |
| CLI | `ve version` | Exit code 0 | Install ve CLI |

#### Execution — CLI (`ve`)

```bash
# List all security groups
ve vpc DescribeSecurityGroups --Region "{{user.region}}"

# Filter by VPC
ve vpc DescribeSecurityGroups --Region "{{user.region}}" --VpcId "{{user.vpc_id}}"

# Filter by security group ID
ve vpc DescribeSecurityGroups --Region "{{user.region}}" --SecurityGroupIds '["{{user.sg_id}}"]'

# Filter by name
ve vpc DescribeSecurityGroups --Region "{{user.region}}" --SecurityGroupName "{{user.sg_name}}"
```

#### Validation

1. Check `$.Result.TotalCount` for total matching security groups
2. Parse `$.Result.SecurityGroups[]` for SG details
3. Report SG IDs, names, VPC IDs, types, and instance counts

---

### Operation: CreateSecurityGroup — Create Security Group

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| VPC exists | DescribeVpcs with VPC ID | VPC found | HALT; create VPC first |
| Quota | Check SG quota per VPC | Sufficient | HALT; request quota increase |

#### Execution — CLI (`ve`)

```bash
# Create basic security group
ve vpc CreateSecurityGroup \
  --Region "{{user.region}}" \
  --VpcId "{{user.vpc_id}}" \
  --SecurityGroupName "{{user.sg_name}}" \
  --Description "{{user.description}}"

# Create enterprise security group (supports more rules, priority-based)
ve vpc CreateSecurityGroup \
  --Region "{{user.region}}" \
  --VpcId "{{user.vpc_id}}" \
  --SecurityGroupName "{{user.sg_name}}" \
  --SecurityGroupType "enterprise" \
  --Description "{{user.description}}"
```

#### Post-execution Validation

1. Parse `{{output.sg_id}}` from response
2. Verify SG exists via DescribeSecurityGroups
3. Report SG ID, type, and default rules

---

### Operation: DescribeSecurityGroupAttributes — List Rules

#### Execution

```bash
# List all rules for a security group
ve vpc DescribeSecurityGroupAttributes \
  --Region "{{user.region}}" \
  --SecurityGroupId "{{user.sg_id}}"
```

#### Output Analysis

Parse inbound and outbound rules:

```bash
# Show inbound rules
ve vpc DescribeSecurityGroupAttributes --Region "{{user.region}}" --SecurityGroupId "{{user.sg_id}}" | jq '.Result.IngressRules[] | {Priority, IpProtocol, PortRange, CidrIp, Policy}'

# Show outbound rules
ve vpc DescribeSecurityGroupAttributes --Region "{{user.region}}" --SecurityGroupId "{{user.sg_id}}" | jq '.Result.EgressRules[] | {Priority, IpProtocol, PortRange, CidrIp, Policy}'
```

---

### Operation: AuthorizeSecurityGroupIngress — Add Inbound Rule

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| SG exists | DescribeSecurityGroups | SG found | HALT |
| Rule not duplicate | Check existing rules | No exact match | Skip if exists |
| Rule count limit | Count existing rules < quota | Within limit | HALT; remove unused rules |

#### Execution

```bash
# Allow TCP traffic from specific CIDR
ve vpc AuthorizeSecurityGroupIngress \
  --Region "{{user.region}}" \
  --SecurityGroupId "{{user.sg_id}}" \
  --IpProtocol "tcp" \
  --PortRange "{{user.port}}" \
  --CidrIp "{{user.cidr}}" \
  --Priority 1 \
  --Description "Allow HTTP from trusted network"

# Allow from another security group
ve vpc AuthorizeSecurityGroupIngress \
  --Region "{{user.region}}" \
  --SecurityGroupId "{{user.sg_id}}" \
  --IpProtocol "tcp" \
  --PortRange "{{user.port}}" \
  --SourceSecurityGroupId "{{user.source_sg_id}}" \
  --Priority 1
```

#### Safety Warnings

| Risk Pattern | Agent Action |
|-------------|--------------|
| `CidrIp` = `0.0.0.0/0` on sensitive port | Warn user; require explicit confirmation |
| Port range covers all ports (`1/65535`) | Warn user; suggest specific ports |
| Protocol = `all` with `0.0.0.0/0` | **HIGH RISK** — require explicit confirmation with risk explanation |

---

### Operation: RevokeSecurityGroupIngress — Remove Inbound Rule

#### Pre-flight

- Identify exact rule to remove (by priority, protocol, port, CIDR)
- Warn about potential service impact

#### Execution

```bash
# Remove rule by matching attributes
ve vpc RevokeSecurityGroupIngress \
  --Region "{{user.region}}" \
  --SecurityGroupId "{{user.sg_id}}" \
  --IpProtocol "{{user.protocol}}" \
  --PortRange "{{user.port}}" \
  --CidrIp "{{user.cidr}}" \
  --Policy "accept"
```

---

### Operation: DeleteSecurityGroup — Delete Security Group

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: irreversible delete
- **MUST** verify SG is not attached to any ECS instances
- **MUST** warn about default SG — cannot delete

```bash
# Check attached instances
ve ecs DescribeInstances --Region "{{user.region}}" --SecurityGroupIds '["{{user.sg_id}}"]' | jq '.Result.TotalCount'
```

#### Execution

```bash
ve vpc DeleteSecurityGroup --Region "{{user.region}}" --SecurityGroupId "{{user.sg_id}}"
```

---

## Security Audit Operations

### Operation: AuditSecurityGroupRules — Comprehensive Rule Audit

Analyzes all security groups for security risks and compliance issues.

#### Execution

```bash
# Get all security groups
ve vpc DescribeSecurityGroups --Region "{{user.region}}" | jq -r '.Result.SecurityGroups[].SecurityGroupId'

# For each SG, audit rules
for sg_id in $(ve vpc DescribeSecurityGroups --Region "{{user.region}}" | jq -r '.Result.SecurityGroups[].SecurityGroupId'); do
  echo "=== Security Group: $sg_id ==="
  ve vpc DescribeSecurityGroupAttributes --Region "{{user.region}}" --SecurityGroupId "$sg_id" | jq '{
    IngressRules: [.Result.IngressRules[] | select(.CidrIp == "0.0.0.0/0" or .CidrIp == "::/0")],
    WidePortRanges: [.Result.IngressRules[] | select(.PortRange == "1/65535" or .PortRange == "-1/-1")]
  }'
done
```

#### Risk Classification

| Risk Level | Pattern | Example |
|------------|---------|---------|
| **CRITICAL** | 0.0.0.0/0 on port 22, 3389, 3306, 6379, 27017 | SSH/RDP/DB exposed to internet |
| **HIGH** | 0.0.0.0/0 on any TCP/UDP port | Any port open to internet |
| **HIGH** | Protocol `all` from 0.0.0.0/0 | Complete access |
| **MEDIUM** | Port range 1/65535 | All ports from specific CIDR |
| **MEDIUM** | Egress allow all to 0.0.0.0/0 | Unrestricted outbound |
| **LOW** | Rule with priority > 50 in enterprise SG | Low-priority rule may be shadowed |
| **LOW** | Duplicate rules across SGs | Redundant configuration |

#### Output Format

```markdown
## Security Group Audit Report — [Date]

### CRITICAL Risks
| SG ID | SG Name | Rule | Risk | Recommendation |
|-------|---------|------|------|---------------|
| sg-xxx | web-sg | 0.0.0.0/0:22/tcp | SSH exposed | Restrict to bastion IP range |
| sg-yyy | db-sg | 0.0.0.0/0:3306/tcp | MySQL exposed | Restrict to app SG only |

### HIGH Risks
| SG ID | SG Name | Rule | Risk | Recommendation |
|-------|---------|------|------|---------------|
| sg-zzz | app-sg | 0.0.0.0/0:8080/tcp | App port exposed | Restrict to CLB SG |

### Summary
- Total SGs audited: N
- Critical risks: X
- High risks: Y
- Medium risks: Z
- Low risks: W
```

---

### Operation: DetectExposedPorts — Find Exposed Sensitive Ports

Specifically scans for commonly targeted ports exposed to the internet.

#### Sensitive Port List

| Port | Service | Risk |
|------|---------|------|
| 22 | SSH | Brute force, unauthorized access |
| 3389 | RDP | Remote exploitation |
| 3306 | MySQL | Data breach |
| 5432 | PostgreSQL | Data breach |
| 6379 | Redis | Unauthorized access, data theft |
| 27017 | MongoDB | Data breach |
| 9200 | Elasticsearch | Data exposure |
| 2379 | etcd | Cluster compromise |
| 11211 | Memcached | DDoS amplification |
| 1433 | SQL Server | Data breach |

#### Execution

```bash
ve vpc DescribeSecurityGroups --Region "{{user.region}}" | jq -r '.Result.SecurityGroups[].SecurityGroupId' | while read sg_id; do
  ve vpc DescribeSecurityGroupAttributes --Region "{{user.region}}" --SecurityGroupId "$sg_id" | jq --arg sg_id "$sg_id" '
    .Result.IngressRules[] |
    select(.CidrIp == "0.0.0.0/0" or .CidrIp == "::/0") |
    select(.IpProtocol == "tcp") |
    select(
      .PortRange == "22/22" or
      .PortRange == "3389/3389" or
      .PortRange == "3306/3306" or
      .PortRange == "5432/5432" or
      .PortRange == "6379/6379" or
      .PortRange == "27017/27017" or
      .PortRange == "9200/9200" or
      (.PortRange | test("^(\\d+/)?(65535|65534)$"))
    ) |
    {SecurityGroupId: $sg_id, Port: .PortRange, Cidr: .CidrIp}
  '
done
```

---

### Operation: FindUnusedSecurityGroups — Identify Unused SGs

Finds security groups not attached to any ECS instances.

#### Execution

```bash
# Get all SGs with instance counts
ve vpc DescribeSecurityGroups --Region "{{user.region}}" | jq '.Result.SecurityGroups[] | {SecurityGroupId, SecurityGroupName, VpcId, Description}'

# For each SG, check attached instances
for sg_id in $(ve vpc DescribeSecurityGroups --Region "{{user.region}}" | jq -r '.Result.SecurityGroups[].SecurityGroupId'); do
  count=$(ve ecs DescribeInstances --Region "{{user.region}}" --SecurityGroupIds "[\"$sg_id\"]" | jq '.Result.TotalCount')
  echo "$sg_id: $count instances"
done
```

#### Output Format

```markdown
## Unused Security Groups Report

| SG ID | SG Name | VPC | Created | Instances | Recommendation |
|-------|---------|-----|---------|-----------|---------------|
| sg-xxx | old-test-sg | vpc-aaa | 2025-01-15 | 0 | Safe to delete |
| sg-yyy | temp-sg | vpc-bbb | 2025-03-20 | 0 | Verify then delete |
```

---

### Operation: OptimizeToLeastPrivilege — Generate Least-Privilege Recommendations

Analyzes actual traffic patterns (if available) and recommends minimal rule sets.

#### Recommendation Logic

| Current Rule | Recommendation |
|-------------|---------------|
| 0.0.0.0/0 on port 22 | Restrict to bastion host CIDR or SG |
| 0.0.0.0/0 on port 3306 | Restrict to application SG only |
| 0.0.0.0/0:1/65535 | Replace with specific required ports |
| Egress allow all | Restrict to required destinations |
| Multiple SGs with identical rules | Consolidate into shared SG |

#### Execution

```bash
# Generate optimization report
echo "## Least-Privilege Optimization Recommendations"
echo ""
echo "| SG ID | Current Rule | Recommended Rule | Reason |"
echo "|-------|-------------|-----------------|--------|"

ve vpc DescribeSecurityGroups --Region "{{user.region}}" | jq -r '.Result.SecurityGroups[].SecurityGroupId' | while read sg_id; do
  ve vpc DescribeSecurityGroupAttributes --Region "{{user.region}}" --SecurityGroupId "$sg_id" | jq -r --arg sg_id "$sg_id" '
    .Result.IngressRules[] |
    select(.CidrIp == "0.0.0.0/0") |
    "| \($sg_id) | \(.IpProtocol):\(.PortRange) from 0.0.0.0/0 | Restrict to specific CIDR/SG | Overly permissive |"
  '
done
```

---

### Operation: DetectRuleConflicts — Find Conflicting Rules

Identifies overlapping or contradictory rules within and across security groups.

#### Conflict Types

| Type | Description | Example |
|------|-------------|---------|
| **Shadowed Rule** | Higher-priority rule makes lower-priority rule ineffective | Priority 1 allows, Priority 10 denies same traffic |
| **Duplicate Rule** | Identical rules in same SG | Two rules: tcp:80 from 10.0.0.0/8 |
| **Overlapping CIDR** | Rules with overlapping source CIDRs | 10.0.0.0/8 and 10.1.0.0/16 for same port |
| **Contradictory Rules** | Allow and deny for same traffic in enterprise SG | Allow tcp:80 from 10.0.0.0/8, Deny tcp:80 from 10.1.0.0/16 |

#### Execution

```bash
# Check for conflicts within a security group
ve vpc DescribeSecurityGroupAttributes --Region "{{user.region}}" --SecurityGroupId "{{user.sg_id}}" | jq '
  .Result.IngressRules | group_by(.IpProtocol + .PortRange + .CidrIp) |
  map(select(length > 1)) |
  .[] | {ConflictType: "duplicate", Rules: .}
'
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
- [FinOps Best Practices](../../ve-skill-generator/references/finops-best-practices.md)

## Operational Best Practices

- **Least privilege:** Always restrict CIDR to minimum required range; use SG-to-SG references instead of CIDR when possible
- **Default deny:** Enterprise SGs should start with deny-all; add only required allow rules
- **Naming convention:** Use consistent naming (project-env-purpose-sg)
- **Regular audits:** Review security group rules monthly; remove unused rules and SGs
- **Tagging:** Tag security groups with owner, purpose, and environment for lifecycle management
- **Separation of duties:** Use different SGs for different tiers (web, app, db)
- **Egress control:** Restrict outbound traffic; don't allow all egress to 0.0.0.0/0
- **Documentation:** Document the purpose of each rule in the Description field

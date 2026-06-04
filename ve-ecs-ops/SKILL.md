---
name: ve-ecs-ops
description: >-
  Use when the user needs to deploy, configure, troubleshoot, or manage Volcengine
  (火山引擎) ECS (云服务器) — instance lifecycle, disks, snapshots, images, security
  groups, and network. User mentions ECS, 云服务器, 云服务器ECS, or describes
  instance-related scenarios (e.g., instance start/stop, create/delete instances,
  disk attachment, snapshot management, key pair, instance type changes) even
  without naming the product directly. Not for billing, IAM, or related products
  that have their own ops skills. Not for container services (VKE/ACK) or PaaS
  offerings.
license: MIT
compatibility: >-
  Official Volcengine CLI (`ve`, Go binary, no runtime), Go 1.14+ runtime
  (for JIT SDK fallback), valid API credentials, network access to Volcengine
  endpoints (open.volcengineapi.com).
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-05-15"
  runtime: Harness AI Agent, Claude Code, Cursor, or compatible Agent runtimes
  go_version_minimum: "1.14"
  go_version_jit: "1.21+"
  api_profile: "ECS API 2020-04-01 (https://www.volcengine.com/docs/6396/69513)"
  cli_applicability: dual-path
  cli_support_evidence: >-
    Confirmed via `ve ecs --help` — ECS is supported by the ve CLI.
    See: https://github.com/volcengine/volcengine-cli
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

> This skill follows the [Agent Skill OpenSpec](https://agentskills.io/specification).

# Volcengine ECS Operations Skill

## Overview

ECS (云服务器) on Volcengine (火山引擎) provides scalable compute capacity including virtual machine instances, cloud disks, snapshots, images, and security groups. This skill is an **operational runbook** for agents: explicit scope, credential rules, pre-flight checks, **dual-path execution** (official **SDK/API** and **`ve` CLI**), response validation, and failure recovery. **Do not use the web console as the primary agent execution path** in `SKILL.md` or [Volcengine Console](https://console.volcengine.com).

> **UX Compliance:** This skill follows the [User Experience Specification](references/user-experience-spec.md). All operations include onboarding guidance, minimal prompts, smart defaults, clear feedback, and user-friendly error handling.

### CLI applicability

- **`cli_applicability: dual-path`:** Official `ve` CLI supports ECS. You **MUST** document **both** the SDK step **and** the `ve` CLI step for every operation.

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use conditions with precise triggers and delegation rules |
| 2 | **Structured I/O** | `{{env.VOLCENGINE_*}}` (env), `{{user.*}}` (interactive), `{{output.*}}` (API response) |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | Error taxonomy with ≥ 10 ECS-specific codes; HALT vs retry per type |
| 5 | **Absolute Single Responsibility** | One product (ECS), one primary resource (Instance); cross-product delegation documented |
| 6 | **FinOps Integration** | Idle detection, right-sizing, cleanup workflows, cost reports |
| 7 | **AIOps Integration** | Knowledge base with fault patterns, cross-skill diagnosis, alarm storm handling |

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- User mentions "Volcengine ECS", "火山引擎 ECS", "云服务器", or "云服务器ECS"
- Task involves CRUD or lifecycle operations on **ECS Instances**: create (RunInstances), describe (DescribeInstances), start (StartInstance), stop (StopInstance), reboot (RebootInstance), modify (ModifyInstanceAttribute), delete (DeleteInstance)
- Task involves **ECS-attached resources**: cloud disks (CreateVolume, AttachVolume), snapshots (CreateSnapshot), images (CreateImage, DescribeImages), key pairs (CreateKeyPair, ImportKeyPair), security groups (AuthorizeSecurityGroupIngress)
- Task involves instance type changes (ModifyInstanceSpec), security group management (ModifyInstanceVpcAttribute), network interfaces (AssignPrivateIpAddresses)
- Task involves **Cloud Assistant** (云助手): remote command execution, job management, client status check
- Task involves **cost optimization**: finding idle instances, right-sizing recommendations, cost summaries
- Task involves **resource cleanup**: removing stopped instances, orphaned disks, old snapshots, unused images
- Task involves **billing model optimization**: converting PostPaid to PrePaid, using Spot instances

### SHOULD NOT Use This Skill When

- Task is purely billing / account management → delegate to billing ops
- Task is IAM / permission model only → delegate to: `ve-iam-ops` (when present)
- Task is about **VPC networking** (routing, NAT Gateway) → delegate to: `ve-vpc-ops` (when present)
- Task is about **Container Service (VKE)** → delegate to container ops skill
- Task is about **Load Balancer (CLB/ALB)** → delegate to: `ve-slb-ops` (when present)
- User insists on **console-only** flows with no API → state limitation

### Delegation Rules

- Instance creation depends on VPC/VSwitch → verify VPC exists via `ve-vpc-ops`
- Security group rules depend on understanding traffic → reference `ve-vpc-ops` for network context
- Multi-product requests: handle each product with its skill; do not merge unrelated APIs

## Variable Convention (Agent-Readable)

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_REGION}}` | From runtime environment | Use documented default only if skill explicitly allows |
| `{{user.region}}` | User-supplied region | Ask once; reuse; default from env |
| `{{user.instance_name}}` | User-supplied instance name | Ask once; reuse |
| `{{user.instance_id}}` | User-supplied instance ID | Ask once; reuse; format `i-xxxxxxxxx` |
| `{{user.image_id}}` | User-supplied image ID | Format `image-xxxxxxxxx` |
| `{{user.instance_type}}` | User-supplied instance type (e.g., ecs.g3i.large) | Ask once; reuse |
| `{{user.vpc_id}}` | User-supplied VPC ID | Format `vpc-xxxxxxxxx` |
| `{{user.subnet_id}}` | User-supplied subnet ID | Format `subnet-xxxxxxxxx` |
| `{{output.instance_id}}` | From RunInstances/DescribeInstances response | Parse from `$.Result.InstanceIds[0]` or `$.Result.Instances[].InstanceId` |
| `{{output.status}}` | Instance lifecycle state | Parse from `$.Result.Instances[].Status` |

> **`{{env.*}}` MUST NOT** be collected from the user. **`{{user.*}}`** MUST be collected interactively when missing.

> **Security Warning (Credential Masking — MANDATORY):** **NEVER** log, print, or expose `VOLCENGINE_SECRET_KEY`, `secret_key`, `SecretKey`, or any credential field value in console output, debug messages, error messages, or logs.
>
> **Credential verification MUST check existence only**, never echo the value:
> - Bash: `test -n "$VOLCENGINE_SECRET_KEY"` ✅ | `echo $VOLCENGINE_SECRET_KEY` ❌
> - Go: `if os.Getenv("VOLCENGINE_SECRET_KEY") == ""` ✅ | `fmt.Println(os.Getenv("VOLCENGINE_SECRET_KEY"))` ❌

## API and Response Conventions (Agent-Readable)

- **Volcengine ECS OpenAPI (2020-04-01)** is canonical for path, query, body fields, enums, and response shapes.
- **Endpoint:** `ecs.volcengineapi.com` (default: `open.volcengineapi.com`)
- **Errors:** Responses with `Error` object containing `Code` and `Message` fields.
- **Timestamps:** ISO 8601 format (e.g. `2026-04-28T10:00:00Z`).

### Key Response Fields

| Operation | JSON Path | Type | Description |
|-----------|-----------|------|-------------|
| RunInstances | `$.Result.InstanceIds` | array | Created instance IDs |
| RunInstances | `$.Result.InstanceIds[0]` | string | First instance ID |
| DescribeInstances | `$.Result.Instances` | array | Instance list |
| DescribeInstances | `$.Result.Instances[].InstanceId` | string | Instance ID |
| DescribeInstances | `$.Result.Instances[].Status` | string | State: `RUNNING`, `STOPPED`, `CREATING`, etc. |
| DescribeInstances | `$.Result.Instances[].InstanceName` | string | Instance name |
| DescribeInstances | `$.Result.Instances[].InstanceType` | string | Instance spec |
| DescribeInstances | `$.Result.Instances[].PrimaryIpAddress` | string | Primary private IP |
| DescribeInstances | `$.Result.Instances[].VpcId` | string | VPC ID |
| DescribeInstances | `$.Result.TotalCount` | integer | Total matching instances |
| DescribeImages | `$.Result.Images` | array | Image list |
| DescribeSnapshots | `$.Result.Snapshots` | array | Snapshot list |
| CreateVolume | `$.Result.VolumeId` | string | Volume ID |
| CreateSnapshot | `$.Result.SnapshotId` | string | Snapshot ID |
| CreateKeyPair | `$.Result.KeyPairName` | string | Key pair name |
| CreateKeyPair | `$.Result.PrivateKey` | string | PEM private key |
| DeleteInstance | `$.ResponseMetadata.RequestId` | string | Request ID |
| StartInstance | `$.ResponseMetadata.RequestId` | string | Request ID |

### Instance State Transitions

| Operation | Initial State | Target State | Poll Interval | Max Wait |
|-----------|---------------|--------------|---------------|----------|
| RunInstances | — | `RUNNING` | 5s | 600s |
| StartInstances | `STOPPED` | `RUNNING` | 5s | 300s |
| StopInstances | `RUNNING` | `STOPPED` | 5s | 300s |
| RebootInstances | `RUNNING` | `RUNNING` | 5s | 300s |
| DeleteInstances | `STOPPED` | absent | 5s | 300s |

## Quick Start

### What This Skill Does
This skill enables you to deploy, configure, troubleshoot, and manage Volcengine (火山引擎) ECS instances, disks, snapshots, images, and security groups using the `ve` CLI (primary) or JIT Go SDK (fallback).

### Prerequisites
- [ ] `ve` CLI installed (or Go runtime for JIT fallback)
- [ ] Credentials configured: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`
- [ ] Region set: `VOLCENGINE_REGION`

### Verify Setup
```bash
ve ecs DescribeInstances --Region {{env.VOLCENGINE_REGION}} --MaxResults 1
```

### Your First Command
```bash
# List all instances in the configured region
ve ecs DescribeInstances --Region {{env.VOLCENGINE_REGION}}
```

## Capabilities at a Glance

| Operation | Description | Complexity | Risk Level |
|-----------|-------------|------------|------------|
| RunInstances | Create one or more ECS instances | High | Medium |
| DescribeInstances | Query instance list and details | Low | None |
| StartInstance | Start a stopped instance | Low | Low |
| StopInstance | Stop a running instance | Low | Medium |
| RebootInstance | Reboot an instance | Low | Medium |
| ModifyInstanceSpec | Change instance type (resize) | Medium | High |
| ModifyInstanceAttribute | Modify instance name, description, etc. | Low | Low |
| DeleteInstance | Delete an instance | Low | **High** — irreversible |
| CreateVolume | Create a cloud disk | Medium | Low |
| AttachVolume | Attach a disk to an instance | Medium | Medium |
| CreateSnapshot | Create a disk snapshot | Low | Low |
| CreateImage | Create a custom image | Low | Low |
| CreateKeyPair | Create an SSH key pair | Low | Low |
| InvokeCommand | Remote command execution via Cloud Assistant | Medium | Medium |
| DescribeInvocations | Query Cloud Assistant job results | Low | None |
| StopInvocation | Cancel a running Cloud Assistant job | Low | Low |
| AssignPrivateIpAddresses | Assign private IPs to ENI | Low | Low |
| DescribeIdleInstances | Find instances with low utilization | Low | None |
| RightSizeInstance | Recommend optimal instance type | Medium | Low |
| DescribeCostSummary | Generate cost report for ECS resources | Low | None |
| CleanupStoppedInstances | Remove stopped instances older than threshold | Low | **High** |
| CleanupOrphanedDisks | Delete unattached cloud disks | Low | **High** |
| CleanupOldSnapshots | Delete snapshots older than threshold | Low | Medium |

## Execution Flows (Agent-Readable)

Every operation: **Pre-flight → Execute (ve CLI primary + JIT Go SDK fallback) → Validate → Recover**.

### Operation: DescribeInstances — Query Instance List

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | `test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY"` | Both set | HALT; configure credentials |
| Region | Verify `{{user.region}}` or `{{env.VOLCENGINE_REGION}}` is set | Valid region | Suggest: `ve ecs DescribeRegions` |
| CLI | `ve version` | Exit code 0 | Install ve CLI |

#### Execution — CLI (`ve`)

```bash
# List all instances (JSON output by default)
ve ecs DescribeInstances --Region "{{user.region}}"

# Filter by instance ID
ve ecs DescribeInstances --Region "{{user.region}}" --InstanceIds '["{{user.instance_id}}"]'

# Filter by status
ve ecs DescribeInstances --Region "{{user.region}}" --Status "RUNNING"

# Pagination
ve ecs DescribeInstances --Region "{{user.region}}" --MaxResults 50 --NextToken "{{user.next_token}}"

# Filter by VPC
ve ecs DescribeInstances --Region "{{user.region}}" --VpcId "{{user.vpc_id}}"
```

#### Execution — JIT Go SDK (Fallback)

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/ecs"
)

func main() {
    instance := ecs.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "Region":     os.Getenv("VOLCENGINE_REGION"),
        "MaxResults": 50,
    }

    resp, err := instance.Client.Request("DescribeInstances", nil, params)
    if err != nil {
        log.Fatalf("Failed to describe instances: %v", err)
    }
    fmt.Println(string(resp))
}
```

#### Validation

1. Check `$.Result.TotalCount` for total matching instances
2. Parse `$.Result.Instances[]` for instance details
3. Report instance count, IDs, names, and statuses

#### Failure Recovery

| Error Pattern | Max Retries | Agent Action |
|--------------|-------------|--------------|
| `InvalidRegion.NotFound` | 0 | List valid regions via `DescribeRegions`; HALT |
| `Unauthorized` | 0 | HALT; check IAM permissions |
| `InternalError` | 3 | Retry with exponential backoff; HALT after 3 |
| Throttling / 429 | 3 | Back off (2s, 4s, 8s); retry |
| `LimitExceeded.MaxResults` | 0 | Reduce MaxResults; retry once |

---

### Operation: RunInstances — Create ECS Instance

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | Env vars set | Non-empty | HALT |
| Image exists | `ve ecs DescribeImages --ImageIds '["{{user.image_id}}"]'` | Image found | HALT; provide valid image |
| VPC/Subnet exists | Verify `{{user.vpc_id}}` and `{{user.subnet_id}}` | Valid IDs | HALT; create VPC first |
| Instance type available | `ve ecs DescribeInstanceTypes --InstanceTypeIds '["{{user.instance_type}}"]'` | Type exists | HALT; choose available type |
| Quota | Check instance quota per region | Sufficient | HALT; request quota increase |

#### Execution — CLI (`ve`)

```bash
# Create a single instance with minimal parameters
ve ecs RunInstances \
  --Region "{{user.region}}" \
  --InstanceType "{{user.instance_type}}" \
  --ImageId "{{user.image_id}}" \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --InstanceName "{{user.instance_name}}" \
  --Password "{{user.password}}" \
  --ChargeType "PostPaid"
```

**Common parameters:**
| Parameter | Description | Example |
|-----------|-------------|---------|
| `InstanceType` | Instance spec | `ecs.g3i.large` |
| `ImageId` | OS image | `image-xxxxx` |
| `VpcId` | VPC ID | `vpc-xxxxx` |
| `SubnetId` | Subnet ID | `subnet-xxxxx` |
| `InstanceName` | Instance name | `my-web-server` |
| `Password` | Admin password | (user-supplied) |
| `KeyPairName` | SSH key pair | `my-keypair` |
| `ChargeType` | Billing: PostPaid/PrePaid | `PostPaid` |
| `SecurityGroupIds` | Security groups | `["sg-xxxxx"]` |
| `VolumeType` | Root disk type | `ESSD_PL0` |
| `VolumeSize` | Root disk size (GB) | `40` |

#### Execution — JIT Go SDK (Fallback)

```go
params := map[string]interface{}{
    "Region":       os.Getenv("VOLCENGINE_REGION"),
    "InstanceType": "ecs.g3i.large",
    "ImageId":      "image-xxxxx",
    "VpcId":        "vpc-xxxxx",
    "SubnetId":     "subnet-xxxxx",
    "InstanceName": "my-web-server",
    "ChargeType":   "PostPaid",
    "Password":     "YourStrongPassword123!",
    "Volumes": []map[string]interface{}{
        {"VolumeType": "ESSD_PL0", "VolumeSize": 40},
    },
}

resp, err := instance.Client.Request("RunInstances", nil, params)
```

#### Post-execution Validation

1. Parse `$.InstanceIds[]` for created instance IDs → `{{output.instance_id}}`
2. Poll status until `RUNNING`:

```bash
for i in $(seq 1 120); do
  STATUS=$(ve ecs DescribeInstances --Region "{{user.region}}" --InstanceIds '["{{output.instance_id}}"]' | jq -r '.Result.Instances[0].Status')
  [ "$STATUS" = "RUNNING" ] && break
  echo "Current: $STATUS (poll $i/120)"
  sleep 5
done
[ "$STATUS" != "RUNNING" ] && echo "[ERROR] Instance failed to reach RUNNING state (current: $STATUS)" && exit 1
```

3. On success, report instance ID, public/private IPs, and access info
4. On terminal failure (status becomes `ERROR`), go to Failure Recovery

#### Failure Recovery

| Error Pattern | Max Retries | Agent Action | UX Feedback |
|--------------|-------------|--------------|-------------|
| `InvalidImageId.NotFound` | 0 | HALT; verify image ID exists | `[ERROR] InvalidImageId.NotFound: Image not found. What happened: The specified image does not exist or is not available in this region. How to fix: Use DescribeImages to find valid image IDs. Next step: Run "ve ecs DescribeImages --Region {{user.region}}"` |
| `InvalidInstanceType.ValueNotSupported` | 0 | HALT; list available types | `[ERROR] InvalidInstanceType: Instance type not supported. How to fix: Use DescribeInstanceTypes to find available types. Next step: Check available types in this region/zone.` |
| `QuotaExceeded.Instance` | 0 | HALT | `[ERROR] QuotaExceeded: Instance quota reached. How to fix: Request quota increase or use a different region. Next step: Contact Volcengine support.` |
| `InsufficientAvailableStock` | 1 | Retry with different instance type | `[ERROR] InsufficientAvailableStock: Specified instance type currently unavailable. How to fix: Try a different instance type or wait. Next step: Choose alternative type.` |
| `InvalidSubnetId.NotFound` | 0 | HALT; verify subnet | `[ERROR] InvalidSubnetId: Subnet not found. How to fix: Verify the subnet ID and region match. Next step: Run "ve vpc DescribeSubnets --VpcId {{user.vpc_id}}"` |
| `InvalidPasswordFormat` | 0 | HALT; fix password | `[ERROR] InvalidPasswordFormat: Password does not meet requirements. How to fix: 8-30 chars, must include 3 of: uppercase, lowercase, digits, special chars.` |
| `InvalidSecurityGroupId.NotFound` | 0 | HALT; create security group | `[ERROR] InvalidSecurityGroupId: Security group not found. How to fix: Verify the security group ID or create one.` |
| `Unauthorized` | 0 | HALT; check IAM | `[ERROR] Unauthorized: Insufficient permissions. How to fix: Ensure ECSFullAccess policy is attached.` |
| `InternalError` | 3 | Retry with backoff | `[ERROR] InternalError: Server-side error. Will retry automatically.` |
| `Throttling` | 3 | Exponential backoff | `⚠️ Rate limit reached. Retrying...` |
| `ExpiredOrder` | 0 | Retry | `[ERROR] ExpiredOrder: Request expired. How to fix: Retry the creation.` |
| `IncorrectInstanceStatus` | 0 | HALT; check status | `[ERROR] IncorrectInstanceStatus: Instance is not in a valid state for this operation.` |

---

### Operation: StartInstance — Start Instance

#### Pre-flight

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | DescribeInstances with ID | Instance found | HALT |
| Instance is stopped | Status = `STOPPED` | Confirmed | Skip if already running |

#### Execution

```bash
# Start a single instance
ve ecs StartInstance --Region "{{user.region}}" --InstanceId "{{user.instance_id}}"

# Start multiple instances
ve ecs StartInstances --Region "{{user.region}}" --InstanceIds '["i-xxx","i-yyy"]'
```

#### Validation

Poll until `Status` = `RUNNING` (max 300s, interval 5s).

---

### Operation: StopInstance — Stop Instance

#### Pre-flight

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | DescribeInstances | Instance found | HALT |
| Force stop option | Determine if soft or hard stop needed | User preference | Default: soft stop |

#### Execution

```bash
# Soft stop (default, graceful shutdown)
ve ecs StopInstance --Region "{{user.region}}" --InstanceId "{{user.instance_id}}"

# Hard stop (immediate power off)
ve ecs StopInstance --Region "{{user.region}}" --InstanceId "{{user.instance_id}}" --ForceStop true

# Stop multiple instances
ve ecs StopInstances --Region "{{user.region}}" --InstanceIds '["i-xxx","i-yyy"]' --ForceStop false
```

#### Validation

Poll until `Status` = `STOPPED` (max 300s, interval 5s).

---

### Operation: RebootInstance — Reboot Instance

#### Execution

```bash
# Reboot a single instance
ve ecs RebootInstance --Region "{{user.region}}" --InstanceId "{{user.instance_id}}"

# Reboot multiple instances (force)
ve ecs RebootInstances --Region "{{user.region}}" --InstanceIds '["i-xxx","i-yyy"]' --ForceStop true
```

#### Validation

Poll until status transitions `RUNNING` → `REBOOTING` → `RUNNING` (max 300s).

---

### Operation: DeleteInstance — Delete Instance

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: irreversible delete of instance `{{user.instance_name}}` (ID: `{{user.instance_id}}`)
- **MUST NOT** proceed without clear user assent
- **MUST** verify instance is in `STOPPED` state (cannot delete running instances)
- **MUST** warn about attached cloud disks — they may be deleted with the instance
- **MUST** warn about deletion protection — verify it is disabled

```bash
# Check deletion protection status
ve ecs DescribeInstances --Region "{{user.region}}" --InstanceIds '["{{user.instance_id}}"]' | jq '.Result.Instances[0].DeletionProtection'
```

#### Execution

```bash
# Delete instance
ve ecs DeleteInstance --Region "{{user.region}}" --InstanceId "{{user.instance_id}}"

# Delete instance with attached data disks
ve ecs DeleteInstance --Region "{{user.region}}" --InstanceId "{{user.instance_id}}" --TerminateSubscriptions true
```

#### Validation

Poll DescribeInstances until instance not found (404) or `$.TotalCount` = 0 (max 300s).

---

### Operation: ModifyInstanceSpec — Change Instance Type

#### Pre-flight

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance is stopped | Status = `STOPPED` | Confirmed | Stop instance first |
| Target type compatible | Verify same family or supported conversion | Compatible | HALT; choose compatible type |
| Instance has no local disks | Check disk types | No local disks | HALT; local disk instances cannot resize |

#### Execution

```bash
ve ecs ModifyInstanceSpec --Region "{{user.region}}" --InstanceId "{{user.instance_id}}" --InstanceType "{{user.target_instance_type}}"
```

#### Validation

Poll until instance transitions to `STOPPED` → `RUNNING` after start.

---

### Operation: CreateSnapshot — Create Disk Snapshot

#### Execution

```bash
ve ecs CreateSnapshot --Region "{{user.region}}" --VolumeId "{{user.volume_id}}" --SnapshotName "{{user.snapshot_name}}"
```

#### Validation

Poll `DescribeSnapshots` until `Status` = `available` (can take minutes for large disks).

---

### Operation: CreateImage — Create Custom Image

#### Execution

```bash
# From running/stopped instance
ve ecs CreateImage --Region "{{user.region}}" --InstanceId "{{user.instance_id}}" --Name "{{user.image_name}}"

# From disk snapshot
ve ecs CreateImage --Region "{{user.region}}" --SnapshotId "{{user.snapshot_id}}" --Name "{{user.image_name}}"
```

#### Validation

Poll `DescribeImages` until `Status` = `available`.

---

### Operation: CreateKeyPair — Create SSH Key Pair

#### Execution

```bash
ve ecs CreateKeyPair --Region "{{user.region}}" --KeyPairName "{{user.key_pair_name}}"
```

#### Important

The private key is returned **only once** in the API response. User MUST save it immediately.

#### Validation

Verify key pair exists via `ve ecs DescribeKeyPairs --KeyPairNames '["{{user.key_pair_name}}"]'`.

---

### Operation: InvokeCommand — 云助手远程执行命令

云助手（原名"批量作业"）可在 **无需 SSH** 的情况下对一台或多台 ECS 实例执行远程命令。

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Client installed | `ve ecs DescribeCloudAssistantInstanceStatus` — check `"CloudAssistantStatus": "true"` | Status = true | Install client first |
| Instance running | DescribeInstances — Status = `RUNNING` or `STOPPED` | Confirmed | Start instance first |

#### Execution

```bash
# Create and invoke a command on an instance (Linux)
ve ecs InvokeCommand \
  --Region "{{user.region}}" \
  --InstanceIds '["{{user.instance_id}}"]' \
  --CommandContent "$(echo '#!/bin/bash' | base64)" \
  --Type RunShellScript \
  --CommandName "{{user.command_name}}"

# Create and invoke a command on an instance (Windows)
ve ecs InvokeCommand \
  --Region "{{user.region}}" \
  --InstanceIds '["{{user.instance_id}}"]' \
  --CommandContent "$(echo 'Write-Host hello' | base64)" \
  --Type RunBatScript \
  --CommandName "{{user.command_name}}"

# Invoke with timeout (default 600s, max 86400s)
ve ecs InvokeCommand \
  --Region "{{user.region}}" \
  --InstanceIds '["{{user.instance_id}}"]' \
  --CommandContent "$(echo '#!/bin/bash\ndf -h' | base64)" \
  --Type RunShellScript \
  --Timeout 300
```

**Command type options:**

| Type | Platform | Description |
|------|----------|-------------|
| `RunShellScript` | Linux | Shell script (`#!/bin/bash`) |
| `RunBatScript` | Windows | Batch script |
| `RunPowerShellScript` | Windows | PowerShell script |
| `RunPythonScript` | Linux | Python 3 script |

#### Post-execution Validation

1. Capture `{{output.invocation_id}}` from response
2. Poll invocation result:

```bash
ve ecs DescribeInvocationResults \
  --Region "{{user.region}}" \
  --InvocationId "{{output.invocation_id}}"
```

3. Parse `$.Result.CommandInvocationResult[].InvokeStatus` → `Running`, `Success`, `Failed`, `Timeout`
4. On success, display `Output` field from results

#### Failure Recovery

| Error Pattern | Max Retries | Agent Action |
|--------------|-------------|--------------|
| `InvalidInstance.CloudAssistantNotInstalled` | 0 | HALT; install client via `InstallCloudAssistantClient` or instance metadata |
| `InvalidCommandContent.Malformed` | 0 | Base64-encode the command content; retry |
| `InvocationTimeout` | 0 | Increase Timeout parameter; verify command doesn't hang |
| `InvalidInstance.NotFound` | 0 | HALT; verify instance ID |
| `IncorrectInstanceStatus` | 0 | Instance must be `RUNNING` or `STOPPED` |

---

### Operation: DescribeInvocations — 查看作业执行结果

#### Execution

```bash
# List all invocations
ve ecs DescribeInvocations --Region "{{user.region}}"

# Filter by specific invocation ID
ve ecs DescribeInvocations --Region "{{user.region}}" --InvocationId "{{user.invocation_id}}"

# View detailed results
ve ecs DescribeInvocationResults --Region "{{user.region}}" --InvocationId "{{user.invocation_id}}"
```

#### Validation

- Check `$.Result.InvokeStatus` for overall status
- Parse `$.Result.CommandInvocationResult[]` for per-instance results
- Key fields: `Output`, `ExitCode`, `InvokeStatus`, `FinishedTime`

---

### Operation: StopInvocation — 停止运行中的作业

#### Execution

```bash
ve ecs StopInvocation --Region "{{user.region}}" --InvocationId "{{user.invocation_id}}"
```

#### Validation

Poll `DescribeInvocationResults` until status changes from `Running` to `Stopping` or `Timeout`.

---

### Operation: AssignPrivateIpAddresses — Assign Private IPs to ENI

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | DescribeInstances with ID | Instance found | HALT |
| Instance has ENI | DescribeNetworkInterfaces | ENI found | HALT; instance must have network interface |
| Available IPs | Check subnet CIDR available IPs | Sufficient IPs | HALT; expand subnet or use existing IPs |
| IP count limit | Max private IPs per ENI (varies by instance type) | Within limit | HALT; remove unused IPs first |

#### Execution

```bash
# Assign a single private IP to the primary ENI
ve ecs AssignPrivateIpAddresses \
  --Region "{{user.region}}" \
  --InstanceId "{{user.instance_id}}" \
  --PrivateIpAddressCount 1

# Assign specific private IP addresses
ve ecs AssignPrivateIpAddresses \
  --Region "{{user.region}}" \
  --InstanceId "{{user.instance_id}}" \
  --PrivateIpAddresses '["10.0.0.100","10.0.0.101"]'

# Assign IP to a specific ENI
ve ecs AssignPrivateIpAddresses \
  --Region "{{user.region}}" \
  --NetworkInterfaceId "{{user.eni_id}}" \
  --PrivateIpAddressCount 1
```

#### Validation

```bash
# Verify assigned IPs
ve ecs DescribeNetworkInterfaces \
  --Region "{{user.region}}" \
  --InstanceId "{{user.instance_id}}" | jq '.Result.NetworkInterfaceSets.NetworkInterfaceSet[0].PrivateIpSets'
```

#### Failure Recovery

| Error Pattern | Max Retries | Agent Action |
|--------------|-------------|--------------|
| `InvalidInstanceId.NotFound` | 0 | HALT; verify instance ID |
| `InvalidNetworkInterfaceId.NotFound` | 0 | HALT; verify ENI ID |
| `IpAlreadyAssigned` | 0 | IP already assigned; skip |
| `PrivateIpAddressCountExceeded` | 0 | HALT; ENI IP limit reached |
| `InsufficientAvailableIpAddresses` | 0 | HALT; subnet IP exhausted |
| `IncorrectInstanceStatus` | 0 | Instance must be `RUNNING` or `STOPPED` |

---

## FinOps Operations (Agent-Readable)

### Operation: DescribeIdleInstances — Find Idle Instances

Identifies ECS instances with consistently low resource utilization.

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | `test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY"` | Both set | HALT |
| Region | Verify `{{user.region}}` or `{{env.VOLCENGINE_REGION}}` | Valid region | HALT |
| Time window | User specifies analysis period | 7-30 days default | Use default 7 days |

#### Execution

```bash
# List all running instances with details
ve ecs DescribeInstances --Region "{{user.region}}" --Status "RUNNING" | jq -r '.Result.Instances[] | {InstanceId: .InstanceId, InstanceName: .InstanceName, InstanceType: .InstanceType, CreatedAt: .CreatedAt}'
```

For each instance, analyze utilization:

| Signal | Threshold | Classification |
|--------|-----------|---------------|
| CPU avg < 5% for 7 days + Network < 1KB/s | Idle — candidate for deletion |
| CPU avg < 15% for 7 days | Underutilized — candidate for downgrade |
| CPU avg > 85% for 1 hour | Overutilized — candidate for upgrade |
| Memory avg < 30% for 7 days | Memory oversized |

#### Output Format

```markdown
## Idle Instance Analysis — [Date]

| Instance ID | Name | Type | CPU Avg | Memory Avg | Status | Recommendation | Est. Savings/mo |
|------------|------|------|---------|------------|--------|---------------|-----------------|
| i-xxxxx | web-01 | ecs.g3i.xlarge | 3% | 15% | Idle | Delete | ¥450 |
| i-yyyyy | api-02 | ecs.g3i.2xlarge | 12% | 22% | Underutilized | Downgrade to g3i.large | ¥280 |
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `Unauthorized` | HALT; ensure `CMSReadOnlyAccess` IAM policy is attached for metrics |
| `InternalError` | Retry with backoff; HALT after 3 |

---

### Operation: RightSizeInstance — Recommend Optimal Instance Type

Analyzes instance utilization patterns and recommends the most cost-effective instance type.

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | DescribeInstances with ID | Instance found | HALT |
| Utilization data available | Query metrics for past 7-30 days | Data present | Suggest longer monitoring |

#### Execution

```bash
# Get current instance spec
ve ecs DescribeInstances --Region "{{user.region}}" --InstanceIds '["{{user.instance_id}}"]' | jq '.Result.Instances[0] | {InstanceType, Memory, Cpu}'

# Get available instance types in same family
ve ecs DescribeInstanceTypes --Region "{{user.region}}" | jq '.Result.InstanceTypes[] | select(.InstanceTypeFamily == "g3i") | {InstanceType, Cpu, Memory}'
```

#### Recommendation Logic

| Current CPU Avg | Current Memory Avg | Recommendation |
|----------------|-------------------|----------------|
| < 25% | < 40% | Downgrade 1 size (e.g., 2xlarge → xlarge) |
| < 10% | < 20% | Downgrade 2 sizes (e.g., 2xlarge → large) |
| > 80% | > 75% | Upgrade 1 size |
| > 90% | > 85% | Upgrade 2 sizes |
| < 15% | < 25% | Consider Spot if non-critical |

#### Output Format

```markdown
## Right-Sizing Recommendation — {{user.instance_id}}

**Current:** ecs.g3i.2xlarge (8 vCPU, 32 GB) — ¥900/mo
**Recommended:** ecs.g3i.large (2 vCPU, 8 GB) — ¥225/mo
**Estimated Savings:** ¥675/mo (75%)

**Rationale:**
- CPU average: 8% (well within 2 vCPU capacity)
- Memory average: 18% (well within 8 GB capacity)
- Peak CPU: 23% (never exceeds recommended capacity)
```

---

### Operation: CleanupStoppedInstances — Remove Stopped Instances

Identifies and removes ECS instances that have been stopped beyond a threshold.

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: deleting N stopped instances older than {{user.days_threshold}} days
- **MUST NOT** proceed without clear user assent
- **MUST** list all instances to be deleted with details
- **MUST** warn about attached disks and EIPs

#### Execution

```bash
# Find stopped instances older than threshold
ve ecs DescribeInstances --Region "{{user.region}}" --Status "STOPPED" | jq -r '.Result.Instances[] | select(.CreatedAt < "{{user.cutoff_date}}") | {InstanceId, InstanceName, StoppedAt, AttachedDisks}'
```

For each confirmed instance:

```bash
ve ecs DeleteInstance --Region "{{user.region}}" --InstanceId "{{user.instance_id}}" --TerminateSubscriptions true
```

#### Validation

```bash
# Verify instances deleted
ve ecs DescribeInstances --Region "{{user.region}}" --InstanceIds '["{{user.instance_id}}"]' | jq '.Result.TotalCount'
# Expected: 0
```

---

### Operation: CleanupOrphanedDisks — Delete Unattached Cloud Disks

Identifies and removes cloud disks not attached to any instance.

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation before deletion
- **MUST** check disk has no recent snapshots (or create snapshot first)
- **MUST** list all orphaned disks with size and creation date

#### Execution

```bash
# List all disks
ve ecs DescribeDisks --Region "{{user.region}}" | jq -r '.Result.Disks[] | select(.Status == "Available") | {DiskId, DiskName, Size, CreatedAt}'

# Delete orphaned disk (after snapshot)
ve ecs CreateSnapshot --Region "{{user.region}}" --VolumeId "{{user.disk_id}}" --SnapshotName "cleanup-{{user.disk_id}}-$(date +%Y%m%d)"
ve ecs DeleteDisk --Region "{{user.region}}" --DiskId "{{user.disk_id}}"
```

---

### Operation: CleanupOldSnapshots — Delete Old Snapshots

Identifies and removes snapshots older than a threshold.

#### Execution

```bash
# List snapshots older than threshold
ve ecs DescribeSnapshots --Region "{{user.region}}" | jq -r '.Result.Snapshots[] | select(.CreatedAt < "{{user.cutoff_date}}") | {SnapshotId, SnapshotName, Size, CreatedAt, SourceDiskId}'

# Delete old snapshot
ve ecs DeleteSnapshot --Region "{{user.region}}" --SnapshotId "{{user.snapshot_id}}"
```

---

### Operation: DescribeCostSummary — Generate ECS Cost Report

Generates a cost summary for all ECS resources in the account/region.

#### Execution

```bash
# Query billing data for ECS
ve billing DescribeBillDetail --BillingCycle "{{user.billing_cycle}}" --ProductType ecs

# Cross-reference with running instances
ve ecs DescribeInstances --Region "{{user.region}}" | jq '[.Result.Instances[]] | length'
```

#### Output Format

```markdown
## ECS Cost Summary — {{user.billing_cycle}}

| Category | Count | Monthly Cost | Trend |
|----------|-------|-------------|-------|
| Running Instances | 15 | ¥12,500 | ↑ 3% |
| Stopped Instances | 3 | ¥1,200 | — |
| Cloud Disks | 25 | ¥2,800 | ↑ 5% |
| Snapshots | 45 | ¥450 | ↓ 10% |
| **Total** | | **¥16,950** | |

### Optimization Opportunities
- 3 idle instances detected (est. ¥900/mo savings)
- 5 orphaned disks detected (est. ¥150/mo savings)
- 12 old snapshots eligible for cleanup (est. ¥60/mo savings)
```

---

## Quality Gate (GCL)

> This chapter is **mandatory** for every execution of `ve-ecs-ops`. It implements
> the Generator-Critic-Loop defined in `../../AGENTS.md` §3-§9. The Critic and
> Generator MUST live in **isolated prompt contexts** — see AGENTS.md §9
> anti-patterns. Read [`references/rubric.md`](references/rubric.md) for the
> scoring dimensions and [`references/prompt-templates.md`](references/prompt-templates.md)
> for the G/C/O prompt skeletons and verbatim safety prompts.

### Operation Tiers (from `references/rubric.md` §0)

| Tier | Operations | `max_iter` | Safety floor |
|---|---|---|---|
| **Destructive** | `DeleteInstance`, `DeleteDisk`, `DeleteSnapshot`, `DeleteImage`, `DeleteKeyPair`, `DeleteNetworkInterface` | 2 | 1.0 (mandatory) |
| **State-changing** | `StopInstance`, `RebootInstance`, `ModifyInstanceSpec`, `ModifyInstanceAttribute`, `TerminateInstances`, `StopInvocation` | 2 | 1.0 (mandatory) |
| **Mutating** | `RunInstances`, `CreateDisk`, `CreateSnapshot`, `CreateImage`, `CreateKeyPair`, `AttachDisk`, `DetachDisk`, `InvokeCommand` | 2 | ≥ 0.5 |
| **Read-only** | `DescribeInstances`, `DescribeDisks`, `DescribeSnapshots`, `DescribeImages`, `DescribeKeyPairs`, `DescribeRegions`, `DescribeInstanceTypes`, `DescribeInvocationResults`, `DescribeIdleInstances`, `DescribeCostSummary` | 3 | ≥ 0 |

### Loop

1. **Pre-flight (Orchestrator)** — resolve `{{env.*}}` and `{{user.*}}`; classify
   the operation into one of the four tiers above; load
   `references/rubric.md`.
2. **Generate** — execute the operation per the existing
   `## Execution Flows` chapter. Capture full command, parameters, raw response
   excerpt, `RequestId`, validation step output, retries, and final state into
   `./audit-results/gcl-trace-YYYYMMDD-HHMMSS.json` with `redaction_pass: true`.
3. **Critique** — isolated prompt; score correctness / safety / idempotency /
   traceability / spec_compliance per `references/rubric.md`. Emit ≤ 3
   actionable suggestions. The Critic MUST NOT see the raw user request.
4. **Decide** — first match wins:
   - Safety = 0 on a Destructive or State-changing op → **ABORT** (no partial return).
   - All dimensions meet threshold → return Generator output.
   - `iter < max_iter` → inject suggestions into Generator and loop.
   - Else → return best-so-far + unresolved rubric items.

### Safety prompts

The Generator MUST surface a **verbatim** safety prompt to the user before
any Destructive or State-changing operation. Templates are in
[`references/prompt-templates.md` §4](references/prompt-templates.md#4-operation-specific-safety-prompts)
(e.g., §4.1 for `DeleteInstance`, §4.4 for prod `StopInstance`).

The user's literal "yes" / "confirm" reply is what the Critic verifies in the trace.

### Trace

Every GCL run MUST persist a JSON trace to
`./audit-results/gcl-trace-YYYYMMDD-HHMMSS.json`. The trace MUST NOT contain a
real `VOLCENGINE_SECRET_KEY` — only `<masked>` or `sha256:<first-8-hex>` and length.
Add `audit-results/` to `.gitignore` (already done at repo root). See
`references/rubric.md` §4 for the mandatory fields and `references/prompt-templates.md` §3
for the Orchestrator's output shape.

### Cross-skill delegation (extends `## Delegation Rules` above)

When the Critic surfaces a cross-product gap, the **Generator** (not the Critic)
delegates on the next iteration:

| Critic finding | Delegate to |
|---|---|
| IAM policy gap (no permission for `ve ecs <Action>`) | `ve-iam-ops` |
| KMS key / secret needed for the operation | `ve-kms-ops` |
| EIP / VPC network concern raised in a non-network op | `ve-eip-ops` / `ve-vpc-ops` |
| Monitoring / alarm rule change needed after a destructive op | `ve-cms-ops` |
| Billing quota exceeded during the operation | `ve-billing-ops` |

The Critic itself MUST NOT call any of the above — it only emits suggestions.

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-15 | Initial release with instance lifecycle, disk, snapshot, and image management |
| 1.1.0 | 2026-05-27 | Added FinOps operations (idle detection, right-sizing, cleanup, cost reports) and AIOps knowledge base |
| 1.2.0 | 2026-06-04 | Phase 1 GCL rollout: added `## Quality Gate (GCL)` chapter, `references/rubric.md`, `references/prompt-templates.md`; `max_iter=2` for destructive / state_changing ops, `max_iter=3` for read-only ops |

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
- [Enhanced Self-Healing Framework](../../ve-skill-generator/references/enhanced-self-healing-framework.md)
- [FinOps Best Practices](../../ve-skill-generator/references/finops-best-practices.md)
- [Knowledge Base](references/knowledge-base.md)

## Operational Best Practices

- **Least privilege:** Use `ECSAdministratorAccess` or custom IAM policy scoped to required APIs
- **Availability:** Deploy across multiple zones within a region for HA
- **Deletion protection:** Enable for production instances
- **Backup:** Regular snapshots for all production disks
- **Security:** Use key pairs instead of passwords; configure security group least-privilege rules
- **Cost:** Use spot instances for stateless workloads; reserved instances for steady-state

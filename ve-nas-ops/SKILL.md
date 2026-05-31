---
name: ve-nas-ops
description: >-
  Use when the user needs to create, configure, analyze, or manage Volcengine (火山引擎)
  NAS (Network Attached Storage / 文件存储) — file system lifecycle, mount targets,
  access control, performance analysis, storage optimization, and cost management.
  User mentions NAS, 文件存储, file system, mount target, NFS, SMB, EFS, or describes
  scenarios like capacity analysis, performance bottleneck detection, file lifecycle
  management, mount point audit, storage type optimization, or backup management.
  Not for object storage (TOS) or block storage (EBS) that have their own ops skills.
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
  api_profile: "NAS API 2022-01-01"
  cli_applicability: dual-path
  cli_support_evidence: >-
    Confirmed via `ve nas --help` — NAS is supported by the ve CLI.
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

> This skill follows the [Agent Skill OpenSpec](https://agentskills.io/specification).

# Volcengine NAS (Network Attached Storage) Operations Skill

## Overview

NAS (文件存储) on Volcengine (火山引擎) provides scalable, shared file storage accessible by multiple compute instances simultaneously via NFS or SMB protocols. This skill is an **operational runbook** for agents: file system lifecycle management, mount target configuration, performance analysis, storage optimization, file lifecycle management, backup/snapshot management, and cost analysis. **Do not use the web console as the primary agent execution path.**

> **UX Compliance:** This skill follows the [User Experience Specification](../../ve-skill-generator/references/user-experience-spec.md). All operations include onboarding guidance, minimal prompts, smart defaults, clear feedback, and user-friendly error handling.

### CLI applicability

- **`cli_applicability: dual-path`:** Official `ve` CLI supports NAS. You **MUST** document **both** the SDK step **and** the `ve` CLI step for every operation.

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use conditions with precise triggers and delegation rules |
| 2 | **Structured I/O** | `{{env.VOLCENGINE_*}}` (env), `{{user.*}}` (interactive), `{{output.*}}` (API response) |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | Error taxonomy with NAS-specific codes; HALT vs retry per type |
| 5 | **Absolute Single Responsibility** | One product (NAS); cross-product delegation documented |
| 6 | **FinOps Integration** | Capacity planning, storage type optimization, lifecycle-based cost reduction |
| 7 | **AIOps Integration** | Performance bottleneck detection, I/O pattern analysis, anomaly detection |

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- User mentions "NAS", "文件存储", "file system", "mount target", "NFS", "SMB", "EFS"
- Task involves **File System CRUD**: CreateFileSystem, DescribeFileSystems, DeleteFileSystem, ModifyFileSystem
- Task involves **Mount Targets**: CreateMountTarget, DescribeMountTargets, DeleteMountTarget
- Task involves **Access Control**: permission groups, mount permissions, VPC access
- Task involves **performance analysis**: IOPS, throughput, latency, bottleneck detection
- Task involves **storage optimization**: Capacity vs Performance tier selection, right-sizing
- Task involves **file lifecycle**: stale file detection, cleanup recommendations, archiving
- Task involves **backup/snapshot**: snapshot creation, retention policies, restore operations
- Task involves **cost analysis**: storage cost breakdown, growth trends, cost optimization
- Task involves **mount point audit**: verifying mounts, detecting unused mount targets

### SHOULD NOT Use This Skill When

- Task is about **Object Storage (TOS)** → delegate to: `ve-tos-ops`
- Task is about **Block Storage (EBS/Cloud Disk)** → delegate to: `ve-ecs-ops`
- Task is about **ECS instance management** → delegate to: `ve-ecs-ops`
- Task is about **VPC networking** → delegate to: `ve-vpc-ops`
- Task is about **application-level file operations** → not a cloud resource task
- Task is purely billing → delegate to billing ops

### Delegation Rules

- NAS mount targets require VPC and subnet → verify via `ve-vpc-ops`
- NAS is mounted on ECS instances → verify instances exist via `ve-ecs-ops`
- NAS backup to TOS → reference `ve-tos-ops` for destination config

## Variable Convention (Agent-Readable)

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_REGION}}` | From runtime environment | Use documented default only if skill explicitly allows |
| `{{user.region}}` | User-supplied region | Ask once; reuse; default from env |
| `{{user.fs_id}}` | User-supplied file system ID | Format `enas-xxxxxxxxx` |
| `{{user.fs_name}}` | User-supplied file system name | Ask once; reuse |
| `{{user.mount_id}}` | User-supplied mount target ID | Format from API response |
| `{{user.vpc_id}}` | User-supplied VPC ID | Format `vpc-xxxxxxxxx` |
| `{{user.subnet_id}}` | User-supplied subnet ID | Format `subnet-xxxxxxxxx` |
| `{{user.storage_type}}` | User-supplied storage type | `Capacity`, `Performance`, `Extreme` |
| `{{user.protocol}}` | User-supplied protocol | `NFS`, `SMB` |
| `{{user.snapshot_id}}` | User-supplied snapshot ID | Format from API response |
| `{{output.fs_id}}` | From CreateFileSystem response | Parse from `$.FileSystemId` |
| `{{output.mount_id}}` | From CreateMountTarget response | Parse from `$.MountTargetId` |

> **Security Warning (Credential Masking — MANDATORY):** **NEVER** log, print, or expose `VOLCENGINE_SECRET_KEY`.

## API and Response Conventions (Agent-Readable)

- **Volcengine NAS OpenAPI (2022-01-01)** is canonical for NAS APIs.
- **Endpoint:** `nas.volcengineapi.com` (default: `open.volcengineapi.com`)

### Key Response Fields

| Operation | JSON Path | Type | Description |
|-----------|-----------|------|-------------|
| CreateFileSystem | `$.FileSystemId` | string | Created FS ID |
| DescribeFileSystems | `$.FileSystems` | array | FS list |
| DescribeFileSystems | `$.FileSystems[].FileSystemId` | string | FS ID |
| DescribeFileSystems | `$.FileSystems[].FileSystemName` | string | FS name |
| DescribeFileSystems | `$.FileSystems[].StorageType` | string | `Capacity`, `Performance`, `Extreme` |
| DescribeFileSystems | `$.FileSystems[].Protocol` | string | `NFS`, `SMB` |
| DescribeFileSystems | `$.FileSystems[].Status` | string | `running`, `creating`, `deleted` |
| DescribeFileSystems | `$.FileSystems[].UsedCapacityGiB` | number | Used capacity (GiB) |
| DescribeFileSystems | `$.FileSystems[].ChargeType` | string | `PostPaid`, `PrePaid` |
| DescribeFileSystems | `$.FileSystems[].CreateTime` | string | Creation time |
| CreateMountTarget | `$.MountTargetId` | string | Mount target ID |
| DescribeMountTargets | `$.MountTargets` | array | Mount target list |
| DescribeMountTargets | `$.MountTargets[].MountTargetId` | string | Mount target ID |
| DescribeMountTargets | `$.MountTargets[].VpcId` | string | VPC ID |
| DescribeMountTargets | `$.MountTargets[].IpAddress` | string | Mount IP address |
| DescribeMountTargets | `$.MountTargets[].Status` | string | `active`, `creating` |
| DescribeMountTargets | `$.MountTargets[].PermissionGroupName` | string | Permission group |
| CreateSnapshot | `$.SnapshotId` | string | Snapshot ID |
| DescribeSnapshots | `$.Snapshots` | array | Snapshot list |
| DescribeSnapshots | `$.Snapshots[].SnapshotId` | string | Snapshot ID |
| DescribeSnapshots | `$.Snapshots[].Status` | string | `progress`, `accomplished`, `failed` |

## Quick Start

### What This Skill Does
This skill enables you to create, configure, analyze, and manage Volcengine NAS (文件存储) using the `ve` CLI (primary) or JIT Go SDK (fallback).

### Prerequisites
- [ ] `ve` CLI installed (or Go runtime for JIT fallback)
- [ ] Credentials configured: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`
- [ ] Region set: `VOLCENGINE_REGION`

### Verify Setup
```bash
ve nas DescribeFileSystems --Region {{env.VOLCENGINE_REGION}}
```

### Your First Command
```bash
# List all file systems
ve nas DescribeFileSystems --Region {{env.VOLCENGINE_REGION}}
```

## Capabilities at a Glance

| Operation | Description | Complexity | Risk Level |
|-----------|-------------|------------|------------|
| DescribeFileSystems | List file systems | Low | None |
| CreateFileSystem | Create a file system | Medium | Low |
| DeleteFileSystem | Delete a file system | Low | **High** |
| ModifyFileSystem | Modify FS attributes | Low | Low |
| DescribeMountTargets | List mount targets | Low | None |
| CreateMountTarget | Create a mount target | Medium | Low |
| DeleteMountTarget | Delete a mount target | Low | Medium |
| DescribePermissionGroups | List permission groups | Low | None |
| CreatePermissionGroup | Create a permission group | Low | Low |
| DescribeSnapshots | List snapshots | Low | None |
| CreateSnapshot | Create a snapshot | Low | Low |
| DeleteSnapshot | Delete a snapshot | Low | Medium |
| RestoreSnapshot | Restore from snapshot | Medium | **High** |
| AnalyzeCapacity | Analyze capacity and growth trends | Medium | None |
| DetectPerformanceBottlenecks | Identify IOPS/throughput issues | High | None |
| AuditMountPoints | Audit mount point usage | Medium | None |
| OptimizeStorageType | Recommend optimal storage tier | Medium | Medium |
| DetectStaleFiles | Find unused/stale files | High | Low |
| GenerateCostReport | Generate NAS cost analysis | Medium | None |

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.1 | 2026-05-28 | Added FinOps cost calculation reference pricing table; enhanced stale file detection with access age classification; added mount command reference |
| 1.0.0 | 2026-05-27 | Initial release with NAS lifecycle, performance, cost optimization |

## Testing Guide

### Unit Testing Strategy

| Component | Test Approach |
|-----------|---------------|
| Credential validation | Mock environment vars |
| Region validation | Test with valid/invalid regions |
| Storage type selection | Test tier logic |
| Permission group rules | Test CIDR validation |

### Integration Testing

```bash
# Verify credentials
export VOLCENGINE_ACCESS_KEY="test_key"
export VOLCENGINE_SECRET_KEY="test_secret"
export VOLCENGINE_REGION="cn-north-1"

# List file systems
ve nas DescribeFileSystems --Region cn-north-1

# Create test file system
FS_ID=$(ve nas CreateFileSystem --Region cn-north-1 --FileSystemName "test-fs-$(date +%s)" --StorageType "Capacity" --Protocol "NFS" --ChargeType "PostPaid" | jq -r '.FileSystemId')

# Wait for creation
for i in $(seq 1 60); do
  STATUS=$(ve nas DescribeFileSystems --Region cn-north-1 --FileSystemIds '["'$FS_ID'"]' | jq -r '.FileSystems[0].Status')
  [ "$STATUS" = "running" ] && break
  sleep 5
done

# Cleanup
ve nas DeleteFileSystem -- P17 Region cn-north-1 --FileSystemId "$FS_ID"
```

### Test Scenarios

| Scenario | Expected Result |
|----------|-----------------|
| Invalid credentials | Unauthorized error |
| Non-existent region | InvalidRegion error |
| Delete FS with active mounts | FileSystemInUse error |
| Create snapshot | Returns SnapshotId, status progress |

### Smoke Tests

```bash
ve version
ve nas DescribeRegions
```

## Execution Flows (Agent-Readable)

### Operation: DescribeFileSystems

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | `test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY"` | Both set | HALT |
| Region | Verify `{{user.region}}` or `{{env.VOLCENGINE_REGION}}` is set | Valid region | Suggest: `ve nas DescribeRegions` |
| CLI | `ve version` | Exit code 0 | Install ve CLI |

#### Execution

```bash
# List all file systems
ve nas DescribeFileSystems --Region "{{user.region}}"

# Filter by file system ID
ve nas DescribeFileSystems --Region "{{user.region}}" --FileSystemIds '["{{user.fs_id}}"]'

# Filter by storage type
ve nas DescribeFileSystems --Region "{{user.region}}" --StorageType "{{user.storage_type}}"
```

#### Validation

1. Check `$.TotalCount` for total file systems
2. Parse `$.FileSystems[]` for FS details
3. Report FS IDs, names, storage types, protocols, and capacity

---

### Operation: CreateFileSystem — Create File System

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Quota | Check FS quota per region | Sufficient | HALT; request quota increase |
| VPC exists (for mounting) | DescribeVpcs | VPC found | Optional; create later |

#### Execution

```bash
# Create Capacity NAS (cost-effective, lower performance)
ve nas CreateFileSystem \
  --Region "{{user.region}}" \
  --FileSystemName "{{user.fs_name}}" \
  --StorageType "Capacity" \
  --Protocol "NFS" \
  --ChargeType "PostPaid"

# Create Performance NAS (higher IOPS and throughput)
ve nas CreateFileSystem \
  --Region "{{user.region}}" \
  --FileSystemName "{{user.fs_name}}" \
  --StorageType "Performance" \
  --Protocol "NFS" \
  --ChargeType "PostPaid"

# Create Extreme NAS (lowest latency, highest performance)
ve nas CreateFileSystem \
  --Region "{{user.region}}" \
  --FileSystemName "{{user.fs_name}}" \
  --StorageType "Extreme" \
  --Protocol "NFS" \
  --ChargeType "PostPaid"
```

#### Storage Type Comparison

| Type | IOPS (per TB) | Throughput (per TB) | Latency | Use Case | Relative Cost |
|------|---------------|---------------------|---------|----------|--------------|
| Capacity | 50 | 5 MB/s | ~10ms | Archival, backup, infrequent access | 1x |
| Performance | 300 | 15 MB/s | ~1ms | General purpose, dev/test | 3x |
| Extreme | 1000 | 50 MB/s | <0.5ms | High-performance computing, databases | 6x |

#### Post-execution Validation

1. Parse `{{output.fs_id}}` from response
2. Poll status until `running`
3. Report FS ID, storage type, and mount instructions

---

### Operation: CreateMountTarget — Create Mount Target

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| FS exists | DescribeFileSystems | FS found and `running` | HALT |
| VPC exists | DescribeVpcs | VPC found | HALT |
| Subnet exists | DescribeSubnets | Subnet found in VPC | HALT |
| Permission group exists | DescribePermissionGroups | Group found | HALT or create default |

#### Execution

```bash
ve nas CreateMountTarget \
  --Region "{{user.region}}" \
  --FileSystemId "{{user.fs_id}}" \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --PermissionGroupName "VPC_DEFAULT_PERMISSION_GROUP"
```

#### Post-execution Validation

1. Parse `{{output.mount_id}}` and mount IP from response
2. Verify mount target status is `active`
3. Report mount point path for client mounting

---

### Operation: DeleteFileSystem — Delete File System

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: irreversible delete
- **MUST** verify no active mount targets
- **MUST** warn about data loss — all files will be deleted
- **MUST** check if snapshots exist (may be retained)

```bash
# Check mount targets
ve nas DescribeMountTargets --Region "{{user.region}}" --FileSystemId "{{user.fs_id}}"

# Check snapshots
ve nas DescribeSnapshots --Region "{{user.region}}" --FileSystemId "{{user.fs_id}}"
```

#### Execution

```bash
# Delete mount targets first
ve nas DeleteMountTarget --Region "{{user.region}}" --MountTargetId "{{user.mount_id}}"

# Delete file system
ve nas DeleteFileSystem --Region "{{user.region}}" --FileSystemId "{{user.fs_id}}"
```

---

## Performance Analysis

### Operation: AnalyzeCapacity — Analyze Capacity and Growth Trends

#### Execution

```bash
# Get all file systems with capacity info
ve nas DescribeFileSystems --Region "{{user.region}}" | jq '.FileSystems[] | {FileSystemId, FileSystemName, StorageType, Protocol, UsedCapacityGiB, ChargeType, CreateTime, Status}'
```

#### Growth Analysis

| Metric | Calculation | Threshold | Action |
|--------|------------|-----------|--------|
| Daily growth rate | (current - 7d_ago) / 7 | > 10 GB/day | Investigate source |
| Capacity utilization | Used / Quota | > 80% | Plan expansion |
| Capacity utilization | Used / Quota | > 95% | Immediate action needed |
| Storage cost trend | Month-over-month | > 20% increase | Review retention |

#### Output Format

```markdown
## NAS Capacity Analysis Report — [Date]

### File System Inventory
| FS ID | Name | Type | Protocol | Used (GiB) | Status | Created |
|-------|------|------|----------|------------|--------|---------|
| enas-xxx | prod-data | Performance | NFS | 512.5 | running | 2025-06-15 |
| enas-yyy | dev-share | Capacity | NFS | 45.2 | running | 2025-09-20 |

### Growth Trends
| FS ID | Current (GiB) | 7d Ago (GiB) | Daily Growth | 30d Projection |
|-------|--------------|--------------|-------------|---------------|
| enas-xxx | 512.5 | 480.0 | 4.6 GB/day | 650 GB |
| enas-yyy | 45.2 | 44.8 | 0.06 GB/day | 47 GB |

### Alerts
- **enas-xxx**: Growing 4.6 GB/day, projected to exceed quota in ~100 days
```

---

### Operation: DetectPerformanceBottlenecks — Identify I/O Issues

#### Performance Metrics to Check

| Metric | Normal | Warning | Critical | Action |
|--------|--------|---------|----------|--------|
| IOPS utilization | < 60% | 60-80% | > 80% | Consider upgrade or load distribution |
| Throughput utilization | < 60% | 60-80% | > 80% | Consider upgrade |
| Read latency | < 5ms | 5-20ms | > 20ms | Check network, consider Extreme tier |
| Write latency | < 10ms | 10-50ms | > 50ms | Check network, consider Extreme tier |
| Metadata operations | < 1000/s | 1000-5000/s | > 5000/s | Optimize file structure |

#### Execution

```bash
# Get file system details (includes performance tier info)
ve nas DescribeFileSystems --Region "{{user.region}}" --FileSystemIds '["{{user.fs_id}}"]' | jq '.FileSystems[0] | {StorageType, Protocol, Status}'

# Get mount target info (network path affects latency)
ve nas DescribeMountTargets --Region "{{user.region}}" --FileSystemId "{{user.fs_id}}" | jq '.MountTargets[] | {MountTargetId, VpcId, IpAddress, Status}'
```

#### Bottleneck Diagnosis

| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| High latency, low IOPS | Network congestion, wrong storage tier | Check VPC bandwidth; upgrade to Performance/Extreme |
| High latency, high IOPS | Tier limit reached | Upgrade storage type |
| Intermittent timeouts | ECS-NAS network path issue | Check security groups, route tables |
| Slow directory listing | Too many files in single directory | Restructure directory hierarchy |

---

## FinOps Operations

### Operation: OptimizeStorageType — Recommend Optimal Storage Tier

#### Analysis Logic

| Current Usage Pattern | Current Tier | Recommendation | Reason |
|----------------------|--------------|---------------|--------|
| Low IOPS (< 100), low throughput | Performance | **Capacity** | Over-provisioned; save 67% |
| Moderate IOPS (100-500), moderate throughput | Extreme | **Performance** | Over-provisioned; save 50% |
| High IOPS (> 500), low latency needed | Performance | **Extreme** | Under-provisioned; latency issues |
| Very low usage (< 10 GB), infrequent access | Any | **Capacity** | Most cost-effective |
| Archival data, rarely accessed | Performance | **Capacity** or TOS | Significant savings |

#### Execution

```bash
# Get all file systems for analysis
ve nas DescribeFileSystems --Region "{{user.region}}" | jq '.FileSystems[] | {FileSystemId, FileSystemName, StorageType, UsedCapacityGiB, Protocol, ChargeType}'
```

#### Output Format

```markdown
## NAS Storage Optimization Report

| FS ID | Name | Current Tier | Used (GiB) | Recommended Tier | Est. Monthly Savings |
|-------|------|-------------|------------|-----------------|---------------------|
| enas-xxx | dev-share | Performance | 45.2 | Capacity | ¥120 |
| enas-yyy | old-backup | Performance | 200.0 | Capacity or TOS | ¥540 |
| enas-zzz | prod-db | Capacity | 100.0 | Performance | (upgrade needed) |
```

---

### Operation: GenerateCostReport — NAS Cost Analysis

#### Execution

```bash
# Get all file systems with billing info
ve nas DescribeFileSystems --Region "{{user.region}}" | jq '.FileSystems[] | {FileSystemId, FileSystemName, StorageType, UsedCapacityGiB, ChargeType, CreateTime}'
```

#### Cost Calculation (Reference)

| Storage Type | Price (per GB/month, reference) | Notes |
|-------------|-------------------------------|-------|
| Capacity | ¥0.35 | Most cost-effective |
| Performance | ¥1.05 | 3x Capacity |
| Extreme | ¥2.10 | 6x Capacity |

#### Output Format

```markdown
## NAS Cost Report — [Date]

### Cost Breakdown
| FS ID | Name | Tier | Used (GiB) | Monthly Cost | % of Total |
|-------|------|------|------------|-------------|------------|
| enas-xxx | prod-data | Performance | 512 | ¥538 | 65% |
| enas-yyy | dev-share | Performance | 45 | ¥47 | 6% |
| enas-zzz | old-backup | Performance | 200 | ¥210 | 25% |
| **Total** | | | **757** | **¥826** | **100%** |

### Optimization Opportunities
- Move old-backup (200 GB) to Capacity tier: save ¥140/mo
- Archive prod-data files older than 180 days to TOS: save ~¥100/mo
- **Total potential savings: ¥240/mo (29%)**
```

---

### Operation: DetectStaleFiles — Find Unused Files

Identifies files not accessed within a specified threshold.

#### Execution (via ECS mount point)

```bash
# On ECS instance with NAS mounted:
# Find files not accessed in 90 days
find /mnt/nas -type f -atime +90 -ls | sort -k11 -n | head -100

# Find large directories
du -sh /mnt/nas/*/ | sort -hr | head -20

# Find empty directories
find /mnt/nas -type d -empty
```

#### Stale File Classification

| Access Age | Classification | Recommendation |
|-----------|---------------|---------------|
| > 365 days | Archive candidate | Move to TOS or delete |
| 180-365 days | Cold data | Consider Capacity tier |
| 90-180 days | Infrequently accessed | Monitor |
| < 90 days | Active | Keep current tier |

#### Output Format

```markdown
## Stale File Analysis

| Directory | Files > 180d | Files > 365d | Total Size (GB) | Recommendation |
|-----------|-------------|-------------|-----------------|---------------|
| /mnt/nas/logs | 15,230 | 8,450 | 45.2 | Archive to TOS |
| /mnt/nas/backup | 2,100 | 1,800 | 120.5 | Delete or archive |
| /mnt/nas/uploads | 500 | 200 | 12.3 | Review and clean up |
```

---

## Mount Point Audit

### Operation: AuditMountPoints — Verify Mount Health

#### Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Mount target active | DescribeMountTargets | Status = `active` | Recreate mount target |
| VPC connectivity | ECS can reach mount IP | Ping/telnet succeeds | Check security group |
| Permission group | Check permission group rules | Allow client IP | Update permission group |
| Mount command | `df -h` on ECS | NAS appears | Remount file system |
| I/O health | `iostat` on ECS | Normal latency | Investigate network |

#### Execution

```bash
# Get mount targets for all file systems
ve nas DescribeFileSystems --Region "{{user.region}}" | jq -r '.FileSystems[].FileSystemId' | while read fs_id; do
  echo "=== File System: $fs_id ==="
  ve nas DescribeMountTargets --Region "{{user.region}}" --FileSystemId "$fs_id" | jq '.MountTargets[] | {MountTargetId, VpcId, IpAddress, Status, PermissionGroupName}'
done
```

#### Mount Command Reference

```bash
# NFS mount (Linux)
sudo mount -t nfs {{mount_ip}}:/ /mnt/nas

# NFS mount with options (recommended)
sudo mount -t nfs -o vers=4.0,minorversion=0,rsize=1048576,wsize=1048576,hard,timeo=600,retrans=2,noresvport {{mount_ip}}:/ /mnt/nas

# SMB mount (Linux)
sudo mount -t cifs //{{mount_ip}}/share /mnt/nas -o username={{user}},password={{pass}},vers=3.0

# fstab entry for auto-mount
{{mount_ip}}:/ /mnt/nas nfs vers=4.0,hard,timeo=600,retrans=2 0 0
```

---

## Snapshot and Backup Management

### Operation: DescribeSnapshots — List Snapshots

#### Execution

```bash
# List snapshots for a file system
ve nas DescribeSnapshots --Region "{{user.region}}" --FileSystemId "{{user.fs_id}}"

# List all snapshots
ve nas DescribeFileSystems --Region "{{user.region}}" | jq -r '.FileSystems[].FileSystemId' | while read fs_id; do
  ve nas DescribeSnapshots --Region "{{user.region}}" --FileSystemId "$fs_id" | jq '.Snapshots[] | {SnapshotId, FileSystemId, Status, CreateTime, Description}'
done
```

---

### Operation: CreateSnapshot — Create Snapshot

#### Execution

```bash
ve nas CreateSnapshot \
  --Region "{{user.region}}" \
  --FileSystemId "{{user.fs_id}}" \
  --SnapshotName "{{user.snapshot_name}}" \
  --Description "Backup before maintenance window"
```

#### Validation

Poll until status = `accomplished`:

```bash
for i in $(seq 1 120); do
  STATUS=$(ve nas DescribeSnapshots --Region "{{user.region}}" --FileSystemId "{{user.fs_id}}" --SnapshotId "{{output.snapshot_id}}" | jq -r '.Snapshots[0].Status')
  [ "$STATUS" = "accomplished" ] && break
  echo "Current: $STATUS (poll $i/120)"
  sleep 10
done
```

---

### Operation: DeleteSnapshot — Delete Snapshot

#### Pre-flight

- **MUST** verify snapshot is not being used for restore
- **MUST** warn about data loss

#### Execution

```bash
ve nas DeleteSnapshot --Region "{{user.region}}" --SnapshotId "{{user.snapshot_id}}"
```

---

### Operation: RestoreSnapshot — Restore from Snapshot

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: will restore file system to snapshot state
- **MUST** warn that current data after snapshot time will be lost
- **MUST** recommend creating a snapshot of current state first

#### Execution

```bash
# Create snapshot of current state first (recommended)
ve nas CreateSnapshot --Region "{{user.region}}" --FileSystemId "{{user.fs_id}}" --SnapshotName "pre-restore-backup-$(date +%Y%m%d)"

# Restore from snapshot
ve nas RestoreSnapshot --Region "{{user.region}}" --SnapshotId "{{user.snapshot_id}}"
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

- **Storage tier selection:** Start with Performance for general use; downgrade to Capacity for archival workloads
- **Mount security:** Use permission groups to restrict access; never use 0.0.0.0/0 in production
- **Backup strategy:** Regular snapshots before maintenance; keep at least 2 snapshots per file system
- **Directory structure:** Avoid single directories with > 100K files; use hierarchical structure
- **Monitoring:** Set up alerts for capacity > 80% and latency spikes
- **Cost optimization:** Review storage tier monthly; archive cold data to TOS
- **Access pattern:** Use read-only mount targets for backup/analysis workloads
- **Network:** Mount targets should be in same VPC as ECS instances for lowest latency
- **Tagging:** Tag file systems with owner, environment, and purpose for lifecycle management

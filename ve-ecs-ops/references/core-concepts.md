# Core Concepts — Volcengine ECS

## Architecture

Volcengine ECS (云服务器) provides scalable virtual machine instances backed by distributed infrastructure. Key components:

```
┌─────────────────────────────────────────────────┐
│                  VPC (私有网络)                     │
│  ┌──────────┐  ┌──────────┐  ┌───────────────┐  │
│  │ Subnet A │  │ Subnet B │  │ Security Group │  │
│  │ (Zone A) │  │ (Zone B) │  │ (Firewall)     │  │
│  └────┬─────┘  └────┬─────┘  └───────┬───────┘  │
│       │              │               │           │
│  ┌────┴─────┐  ┌────┴─────┐  ┌───────┴───────┐  │
│  │ ECS Inst │  │ ECS Inst │  │ ECS Inst      │  │
│  │ + Disk   │  │ + Disk   │  │ + Local SSD   │  │
│  └──────────┘  └──────────┘  └───────────────┘  │
└─────────────────────────────────────────────────┘
       │              │              │
  ┌────┴─────┐  ┌────┴─────┐  ┌────┴─────┐
  │ Snapshot │  │  Image   │  │  EIP     │
  └──────────┘  └──────────┘  └──────────┘
```

## Instance Lifecycle

```
CREATING → RUNNING ↔ STOPPED ↔ STARTING
   ↓          ↓          ↓
  ERROR    REBOOTING  STOPPING
   ↓
 DELETING → DELETED
```

### Instance States

| State | Description | Available Operations |
|-------|-------------|---------------------|
| `CREATING` | Instance provisioning | Describe only |
| `RUNNING` | Instance is operational | Stop, Reboot, Modify, Delete (if protection off) |
| `STOPPED` | Instance is powered off | Start, Modify, Delete |
| `STARTING` | Instance booting up | Describe only |
| `STOPPING` | Instance shutting down | Describe only |
| `REBOOTING` | Instance restarting | Describe only |
| `ERROR` | Instance has encountered an error | Delete, Diagnose |
| `DELETING` | Instance is being deleted | Describe only |

## Regions and Zones

Query available regions and zones via API — the list changes as new regions are added:

```bash
# List all available regions
ve ecs DescribeRegions

# List zones within a region
ve ecs DescribeZones --Region "{{user.region}}"
```

## Instance Families

| Family | CPU:Memory | Use Case | Examples |
|--------|------------|----------|----------|
| General Purpose (g3i/g3a) | 1:4 | Web servers, app servers | ecs.g3i.large, ecs.g3a.large |
| Compute Optimized (c3i/c3a) | 1:2 | HPC, batch processing | ecs.c3i.large, ecs.c3a.large |
| Memory Optimized (r3i/r3a) | 1:8 | Databases, in-memory apps | ecs.r3i.large, ecs.r3a.large |
| FLEX (f3i) | Variable | General workloads | ecs.f3i.large |
| Big Data (d3i) | 1:4 with local disk | Hadoop, Spark | ecs.d3i.xlarge |
| Shared (s3) | Burstable | Dev/test, low-traffic | ecs.s3.large |

**Query available types:**
```bash
ve ecs DescribeInstanceTypes --InstanceTypeFamilyIds '["g3i"]'
```

## Billing Models

| Model | Chinese | Description |
|-------|---------|-------------|
| PostPaid | 按量计费 | Pay-per-use, billed by the hour/second |
| PrePaid | 包年包月 | Subscription, monthly/yearly upfront |
| Spot | 抢占式实例 | Discounted, can be reclaimed |

## Cloud Disk Types

| Type | Description | ⚡ Max IOPS | ⚡ Max Throughput |
|------|-------------|----------|----------------|
| `ESSD_PL0` | Standard SSD | 10,000 | 150 MB/s |
| `ESSD_FlexPL` | Flexible SSD (performance scalable) | Up to 100,000 | Up to 1,000 MB/s |
| `EHDD`| High-capacity HDD | 6,000 | 150 MB/s |

## Dependency Map

```
RunInstances requires:
  ├── VPC (ve-vpc-ops)
  ├── Subnet (ve-vpc-ops)
  ├── Security Group (ve-vpc-ops)
  ├── Image (this skill)
  ├── Instance Type (this skill)
  └── Optional: Key Pair, Cloud Disk

DeleteInstance affects:
  ├── Attached data disks (may be deleted or detached)
  ├── Attached EIPs (released or disassociated)
  └── Snapshots/Images created from this instance (preserved)
```

## Resource Limits

> **TE-1:** Limits vary by account tier. Query current quota via API:

```bash
ve ecs DescribeResourceQuota --Region "{{user.region}}"
```

| Resource | Query Method |
|----------|-------------|
| Instances per region | `DescribeResourceQuota → $.Result.Quota.InstanceQuota` |
| Security groups per instance | `DescribeResourceQuota → $.Result.Quota.SecurityGroupQuota` |
| Cloud disks per instance | `DescribeResourceQuota → $.Result.Quota.VolumeQuota` |
| Snapshots per disk | `DescribeResourceQuota → $.Result.Quota.SnapshotQuota` |
| Images per account | `DescribeImages — MaxResults=1 → TotalCount` |
| Key pairs per region | `DescribeKeyPairs — MaxResults=1 → TotalCount` |

> Limits can be increased via support ticket.

## FinOps — Cost Optimization

> **TE-7:** Deep FinOps analysis (billing comparison, pricing, cost optimization reference) → [`references/advanced/finops.md`](advanced/finops.md)

See [`references/advanced/finops.md`](advanced/finops.md) for:
- Billing Model Comparison
- Cost Per Instance Type (query API for current prices)
- Cost Optimization Quick Reference

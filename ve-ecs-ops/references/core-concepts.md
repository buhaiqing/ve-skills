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

| Region Code | Region Name | Availability Zones |
|-------------|-------------|-------------------|
| `cn-beijing` | 北京 | cn-beijing-a, cn-beijing-b |
| `cn-shanghai` | 上海 | cn-shanghai-a, cn-shanghai-b |
| `cn-guangzhou` | 广州 | cn-guangzhou-a, cn-guangzhou-b |
| `ap-southeast-1` | 新加坡 | ap-southeast-1a, ap-southeast-1b |

**Query available regions:**
```bash
ve ecs DescribeRegions
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

| Type | Description | Max IOPS | Max Throughput |
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

## Resource Limits (Defaults)

| Resource | Default Limit |
|----------|---------------|
| Instances per account per region | 100 |
| Security groups per instance | 5 |
| Cloud disks per instance | 15 |
| Snapshots per disk | 50 |
| Images per account | 200 |
| Key pairs per region | 500 |

> Limits can be increased via support ticket. Check current quota with: `ve ecs DescribeResourceQuota`

## FinOps — Cost Optimization

### Billing Model Comparison

| Model | Discount | Commitment | Best For |
|-------|----------|------------|----------|
| PostPaid | 0% | None | Variable workloads, testing |
| PrePaid (1 year) | ~35% | 12 months | Steady production workloads |
| PrePaid (3 years) | ~50% | 36 months | Long-term infrastructure |
| Spot | ~60-90% | Interruptible | Batch processing, fault-tolerant |

### Cost Per Instance Type (cn-beijing, PostPaid)

| Type | vCPU | Memory | Hourly | Monthly (~730h) |
|------|------|--------|--------|-----------------|
| ecs.f3i.large | 2 | 4 GB | ¥0.18 | ¥131 |
| ecs.g3i.large | 2 | 8 GB | ¥0.31 | ¥226 |
| ecs.g3i.xlarge | 4 | 16 GB | ¥0.62 | ¥453 |
| ecs.g3i.2xlarge | 8 | 32 GB | ¥1.23 | ¥898 |
| ecs.c3i.large | 2 | 4 GB | ¥0.28 | ¥204 |
| ecs.r3i.large | 2 | 16 GB | ¥0.52 | ¥380 |

### Cost Optimization Quick Reference

| Situation | Action | Savings |
|-----------|--------|---------|
| Instance running > 30 days | Convert PostPaid → PrePaid | ~35% |
| CPU avg < 15% for 7 days | Right-size down | 25-75% |
| Non-critical batch workload | Use Spot | 60-90% |
| Stopped instance > 7 days | Delete or snapshot + delete | 100% |
| Unattached disk | Snapshot + delete | 100% |
| Snapshot > 90 days old | Delete | 100% |

# FinOps — ECS Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7.

## Billing Model Comparison

| Model | Discount | Commitment | Best For |
|-------|----------|------------|----------|
| PostPaid | 0% | None | Variable, testing |
| PrePaid (1yr) | ~35% | 12 months | Steady production |
| PrePaid (3yr) | ~50% | 36 months | Long-term infra |
| Spot | 60-90% | Interruptible | Fault-tolerant batch |

## Product-Specific Cost Optimization

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Use Spot for batch workloads | Stateless/non-critical → Spot | 60-90% |
| Right-size underutilized | CPU avg < 15% for 7d → downsize spec | ~50% |
| Convert steady PostPaid → PrePaid | Running > 30d → 1yr PrePaid | ~35% |
| Delete orphaned volumes | Unattached disk > 7d → snapshot + delete | 100% |

## Cost Anomaly Detection

| Warning Sign | Check Command | Action |
|-------------|---------------|--------|
| 💰 Unexpected instance count spike | `ve ecs DescribeInstances` | Investigate auto-scaling or rogue deployment |
| ⚠️ Spot price surge | `ve ecs DescribeSpotPriceHistory` | Fallback to PostPaid or switch AZ |
| 📊 CPU credit exhaustion (burstable t5) | `ve ecs DescribeInstanceTypes` for t5 specs | Switch to standard instance type |

## Query Current Pricing

```bash
# Query instance type pricing for current region
ve ecs DescribeInstanceTypes --body '{"InstanceTypeIds":["ecs.g3i.large"]}'
```

> 💡 For billing-level queries, use `ve billing DescribeBillSummaryByMonth`.

## Operations

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

| Signal | Threshold | 🔍 Classification |
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

| Current CPU Avg | Current Memory Avg | 💡 Recommendation |
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

> ⚠️ **Destructive Action Confirmation**
> You are about to **permanently delete** N stopped ECS instances older than `{{user.days_threshold}}` days.
> This is **IRREVERSIBLE** — attached cloud disks and EIPs are released; instance data is destroyed.
> Type `confirm cleanup` to confirm, or reply `abort` to cancel.

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation (type-to-confirm above) — **MUST NOT** proceed without clear user assent
- **MUST** list all instances to be deleted with details
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

> ⚠️ **Destructive Action Confirmation**
> You are about to **permanently delete** N unattached cloud disks.
> This is **IRREVERSIBLE** — disk data is permanently destroyed; ensure no valuable data remains.
> Type `confirm cleanup` to confirm, or reply `abort` to cancel.

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation (type-to-confirm above) — **MUST NOT** proceed without clear user assent
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

> ⚠️ **Destructive Action Confirmation**
> You are about to **permanently delete** snapshots older than `{{user.cutoff_date}}`.
> This is **IRREVERSIBLE** — deleted snapshots cannot be recovered; data loss risk.
> Type `confirm cleanup` to confirm, or reply `abort` to cancel.

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation (type-to-confirm above) — **MUST NOT** proceed without clear user assent
- **MUST** list all snapshots to be deleted with size, creation date, and source disk

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

| Category | Count | 💰 Monthly Cost | 📈 Trend |
|----------|-------|-------------|-------|
| Running Instances | 15 | ¥12,500 | ↑ 3% |
| Stopped Instances | 3 | ¥1,200 | — |
| Cloud Disks | 25 | ¥2,800 | ↑ 5% |
| Snapshots | 45 | ¥450 | ↓ 10% |
| **Total** | | **¥16,950** | |

### 💡 Optimization Opportunities
- 3 idle instances detected (est. ¥900/mo savings)
- 5 orphaned disks detected (est. ¥150/mo savings)
- 12 old snapshots eligible for cleanup (est. ¥60/mo savings)
```

---

## Related Resources

- [FinOps — Billing Cost Optimization](../../ve-billing-ops/references/advanced/finops.md)
- [Cloud Monitor — Idle Detection Metrics](../../ve-cms-ops/references/advanced/finops.md) — 用于 Cleanup* 决策

# FinOps Knowledge Base — ve-billing-ops

> **Purpose:** Volcengine-specific billing patterns, gotchas, and institutional knowledge for FinOps practitioners.
> **Version:** 1.0.0>
> **Last Updated:** 2026-06-02

---

## Table of Contents

1. [Volcengine Billing Specifics](#1-volcengine-billing-specifics)
2. [Product-Specific Billing Patterns](#2-product-specific-billing-patterns)
3. [Common Pitfalls](#3-common-pitfalls)
4. [Cost Optimization Recipes](#4-cost-optimization-recipes)
5. [Internal Best Practices](#5-internal-best-practices)

---

## 1. Volcengine Billing Specifics

### 1.1 Billing Cycle

| Aspect | Detail |
|--------|--------|
| Billing Cycle | Calendar month (1st to end of month) |
| Bill Generation | Typically 3-5 days after month end |
| Invoice | Generated after bill confirmation |
| Payment Terms | Prepaid balance or credit |

### 1.2 Pricing Models by Product

| Product | Pay-as-you-go | Subscription | Reserved Instance | Notes |
|---------|--------------|--------------|-------------------|-------|
| ECS | ✅ | ✅ Monthly/Yearly | ✅ 1-Year/3-Year | 30-65% RI savings |
| VKE | ✅ | ✅ Monthly/Yearly | ✅ 1-Year/3-Year | 40-75% RI savings |
| RDS | ✅ | ✅ Monthly/Yearly | ✅ 1-Year/3-Year | 35-55% RI savings |
| Redis | ✅ | ✅ Monthly/Yearly | ✅ 1-Year/3-Year | 40-60% RI savings |
| TOS | ✅ | ✅ (Storage plans) | ❌ | Volume discounts apply |
| SLS | ✅ | ✅ | ❌ | Ingest + storage + query |
| ALB/CLB | ✅ | ✅ | ❌ | LCU/GCU based |
| EIP | ✅ Hourly | ✅ Monthly | ❌ | Peak bandwidth based |

### 1.3 Cost Components by Product

```markdown
## ECS Cost Components
- Instance hours (instance type × hours)
- CPU (if enhanced)
- Memory (if enhanced)
- System disk (GB-hours)
- Data transfer (outbound)
- Public IP (if assigned)

## VKE Cost Components
- Node instance hours (master + worker)
- CloudDisk (per node)
- LoadBalancer (if cluster LB)
- Network traffic (inter-AZ)

## RDS Cost Components
- Instance hours (instance class)
- Storage (GB-month)
- Backup storage (GB-month)
- IOPS (if billable)
- Read replica (if enabled)
- Data transfer

## TOS Cost Components
- Storage (GB-month)
- Requests (GET/PUT)
- Data transfer (outbound)
- Accelerate traffic (if enabled)

## SLS Cost Components
- Ingested data (GB)
- Storage (GB-month)
- Index storage (GB-month)
- Search&Analytics (CU-hours)
```

---

## 2. Product-Specific Billing Patterns

### 2.1 ECS Gotchas

| Pattern | Description | Impact | Mitigation |
|---------|-------------|--------|------------|
| **Stopped Instance Billing** | Stopped instances still incur disk + IP charges | ¥5-20/month | Terminate instead of stop |
| **EIP Auto-Assign** | EIP auto-assigned when stopped instance starts | +¥2-5/month | Disable auto-assign |
| **System Disk Always Charged** | Even for stopped instances | ¥1-3/month/instance | Use ephemeral if possible |
| **Peak Bandwidth vs Traffic** | Billed by bandwidth, not actual traffic | Variable | Right-size bandwidth |
| **Intra-AZ Traffic Free** | Same AZ traffic is free | ✅ Good | Design AZ distribution |

### 2.2 VKE Gotchas

| Pattern | Description | Impact | Mitigation |
|---------|-------------|--------|------------|
| **Master Node Charged** | 3 master nodes always running | ¥50-150/month | Included in some plans |
| **Cluster Load Balancer** | ALB/CLB created per service | ¥20-100/month | Share cluster LB |
| **SNAT Cost** | NAT for pod egress | Variable | Use NAT Gateway efficiently |
| **Volume Snapshots** | CSI snapshots add storage cost | Variable | Policy-based cleanup |
| **Log Agent Overhead** | Log collection adds SLS cost | Variable | Filter noisy logs |

### 2.3 RDS Gotchas

| Pattern | Description | Impact | Mitigation |
|---------|-------------|--------|------------|
| **Backup Storage Overage** | Backups > 100% of data size | Extra cost | Tune retention |
| **Read Replica Always On** | Even during low traffic | +30-50% cost | Scale down during off-peak |
| **IOPS Burst** | Burst credits exhausted | Sudden cost spike | Monitor IOPS |
| **Connection Pool Overage** | Max connections exceeded | Extra charging | Tune connection pool |
| **Multi-AZ Replication** | Primary + standby always | +50-100% cost | Assess HA needs |

### 2.4 TOS Gotchas

| Pattern | Description | Impact | Mitigation |
|---------|-------------|--------|------------|
| **Multipart Upload Cleanup** | Incomplete uploads charge storage | Variable | Set lifecycle rules |
| **Versioning Storage** | Each version charges full storage | 2-3x storage | Use lifecycle policies |
| **Accelerate Egress** | CDN accelerate adds egress cost | Variable | Compare with direct |
| **Request Costs Stack** | GET/PUT/DELETE all charged | Variable | Batch operations |

### 2.5 SLS Gotchas

| Pattern | Description | Impact | Mitigation |
|---------|-------------|--------|------------|
| **Index Bloat** | Full-text indexing increases storage | 3-10x storage | Selective field indexing |
| **Shard Configuration** | Wrong shard sizing wastes CU | Variable | Right-size shards |
| **Long Retention** | 30d vs 90d vs 180d | Significant | Match retention to needs |
| **Continuous Query** | Always-on queries consume CU | Variable | Schedule queries |

---

## 3. Common Pitfalls

### 3.1 Cost Visibility Pitfalls

```markdown
## Pitfall: Relying on Single Dimension

❌ WRONG: Looking at only product-level cost
✅ RIGHT: Break down by environment, project, cost-center

❌ WRONG: Ignoring shared costs
✅ RIGHT: Allocate VPC/NAT/KMS to teams

## Pitfall: Ignoring Time-of-Day Patterns

❌ WRONG: Monthly totals hide daily spikes
✅ RIGHT: Analyze hourly/daily patterns

## Pitfall: No Tag Discipline

❌ WRONG: 60% resources without cost tags
✅ RIGHT: Enforce tag policy; reject untagged resources
```

### 3.2 Cost Optimization Pitfalls

```markdown
## Pitfall: Premature RI Purchase

❌ WRONG: Buying 1-year RI without utilization data
✅ RIGHT: Run 30-60 days of PostPaid first

## Pitfall: Over-optimizing Dev/Test

❌ WRONG: Spending 3 days to save ¥50/month
✅ RIGHT: Focus effort on production savings

## Pitfall: Ignoring Data Transfer

❌ WRONG: Only optimizing compute/storage
✅ RIGHT: Include network egress in analysis

## Pitfall: Short-term Thinking

❌ WRONG: Always choosing cheapest option
✅ RIGHT: Consider total cost (buy + operate)
```

### 3.3 Billing Configuration Pitfalls

```markdown
## Pitfall: Wrong Billing Mode per Product

| Product | Wrong Choice | Right Choice |
|---------|--------------|--------------|
| Dev/Test DB | Always-on PostPaid | Scheduled stop or Spot |
| Batch Processing | Always-on | Spot + Preemptible |
| API Gateway | Fixed capacity | Serverless (per request) |
| Static Storage | Standard | IA/Archive for cold data |

## Pitfall: Unbounded Resources

❌ WRONG: No max instances/size limits
✅ RIGHT: Set quotas per environment

## Pitfall: Missing Alerts

❌ WRONG: No budget alerts configured
✅ RIGHT: Set 80%/90%/100% thresholds
```

---

## 4. Cost Optimization Recipes

### 4.1 Recipe: Batch Workload Cost Reduction

```markdown
## Scenario: Reduce batch processing cost by 60-80%

### Before
- 10x ecs.c6.4xlarge (always-on)
- Cost: ¥15,000/month

### Steps
1. Schedule off-peak execution (night/weekend)
2. Use Spot instances (60% discount)
3. Enable auto-scale to 0 when idle
4. Consider VKE with batch scheduler

### After
- 10x ecs.c6.4xlarge Spot (scheduled)
- Cost: ¥4,500/month
- Savings: ¥10,500/month (70%)
```

### 4.2 Recipe: Database Cost Optimization

```markdown
## Scenario: Reduce RDS cost by 40%

### Before
- 1x rds.mysql.4xlarge (always-on)
- 500GB storage, 100% backup
- Cost: ¥8,000/month

### Steps
1. Enable auto-storage scaling (avoid over-provision)
2. Reduce backup retention to 7 days (if compliance allows)
3. Purchase 1-year RI (40% savings)
4. Enable slow query optimization (may downsize)

### After
- 1x rds.mysql.4xlarge RI
- 300GB storage, 7-day backup
- Cost: ¥4,800/month
- Savings: ¥3,200/month (40%)
```

### 4.3 Recipe: Storage Tiering

```markdown
## Scenario: Optimize object storage cost

### Before
- 10TB Standard storage
- 1M GET requests/month
- Cost: ¥1,200/month

### Steps
1. Analyze access patterns
2. Move data > 30 days → InfrequentAccess
3. Move data > 90 days → Archive
4. Implement lifecycle policies
5. Reduce GET requests with CDN

### After
- 5TB Standard, 3TB IA, 2TB Archive
- 500K GET + CDN cache
- Cost: ¥650/month
- Savings: ¥550/month (46%)
```

### 4.4 Recipe: VKE Node Optimization

```markdown
## Scenario: Reduce Kubernetes cluster cost

### Before
- 6x ecs.g6.2xlarge nodes (over-provisioned)
- CPU: 15% average, Memory: 40%
- Cost: ¥6,000/month

### Steps
1. Profile actual resource needs
2. Downscale to 4x ecs.g6.xlarge (right-sized)
3. Enable cluster autoscaler
4. Purchase 1-year RI for baseline nodes
5. Use Spot for stateless workloads

### After
- 2x ecs.g6.xlarge RI + 2x ecs.g6.xlarge Spot
- HPA enabled for burst
- Cost: ¥2,800/month
- Savings: ¥3,200/month (53%)
```

---

## 5. Internal Best Practices

### 5.1 Cost Allocation Policy

```markdown
## Policy: Mandatory Cost Tags

| Tag Key | Tag Value | Required |
|---------|-----------|----------|
| cost-center | department-name | ✅ |
| project | project-id | ✅ |
| environment | production/staging/dev | ✅ |
| owner | team-email | ✅ |
| managed-by | terraform/manual/ai-agent | ✅ |

## Enforcement
- New resources: Reject if tags missing
- Existing resources: 30-day remediation window
- Monthly audit: Tag compliance report
```

### 5.2 Budget Strategy

```markdown
## Budget Hierarchy

| Level | Scope | Alert Threshold | Owner |
|-------|-------|-----------------|-------|
| Organization | All costs | 80% | FinOps |
| Cost Center | Per department | 80% | Dept Head |
| Project | Per project | 90% | PM |
| Environment | Production only | 85% | DevOps |

## Budget Review Cadence
- Weekly: Budget utilization review
- Monthly: Budget recalibration
- Quarterly: Budget strategy alignment
```

### 5.3 RI Purchase Policy

```markdown
## RI Purchase Decision Matrix

| Condition | Action |
|-----------|--------|
| < 30 days PostPaid data | ❌ No RI purchase |
| 30-60 days stable utilization > 70% | ✅ Consider 1-year RI |
| 60+ days stable utilization > 80% | ✅ Purchase 1-year RI |
| > 90% utilization + growing | ✅ Consider 3-year RI |
| Utilization < 50% | ❌ Return underutilized RIs |

## RI Coverage Target
- Production: > 80% coverage
- Staging: > 50% coverage
- Dev: On-demand only
```

### 5.4 FinOps Metrics

```markdown
## Key Performance Indicators

| KPI | Definition | Target | Review |
|-----|------------|--------|--------|
| Waste Ratio | (Idle + Undersized) / Total | < 10% | Monthly |
| Tag Compliance | Tagged resources / Total | > 95% | Weekly |
| RI Coverage | RI units used / Total units | > 80% | Monthly |
| Cost per User | Total cost / Active users | Decreasing | Monthly |
| Budget Accuracy | Actual / Budget | < 100% | Weekly |
| Savings Achieved | Savings / Total cost | > 5% | Quarterly |

## Monthly FinOps Dashboard

```
┌─────────────────────────────────────────────────────┐
│  Month: [Current]    Budget: ¥XXX,XXX              │
├─────────────────────────────────────────────────────┤
│  Total Spend    │ ████████████░░░░ 80% │ ¥XXX,XXX │
│  Waste Ratio    │ ██░░░░░░░░░░░░░░░░░  8% │ Target: <10% │
│  Tag Compliance │ ██████████████░░░ 92% │ Target: >95% │
│  RI Coverage     │ ████████████░░░░░ 78% │ Target: >80% │
└─────────────────────────────────────────────────────┘
```
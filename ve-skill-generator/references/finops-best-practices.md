# FinOps Best Practices — Volcengine Skill Generator

> **Purpose:** Defines mandatory FinOps patterns for cost visibility, optimization, and measurement across all VE cloud skills.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-27

---

## 1. Cost Visibility

### 1.1 Resource Cost Attribution

Every skill MUST support cost-aware operations:
- Tag resources with cost allocation labels (project, environment, team)
- Query billing data per resource type
- Report cost breakdown in operational summaries

### 1.2 Cost Tags Convention

| Tag Key | Description | Example |
|---------|-------------|---------|
| `cost-center` | Cost allocation unit | `engineering-infra` |
| `project` | Project identifier | `project-alpha` |
| `environment` | Deployment environment | `production`, `staging`, `dev` |
| `owner` | Resource owner | `team-backend` |
| `managed-by` | Provisioning tool | `terraform`, `manual`, `ai-agent` |

### 1.3 Billing Query Pattern

```bash
# Query billing data by tag
ve billing DescribeBillDetail --BillingCycle 2026-05 --ProductType ecs
```

---

## 2. Cost Optimization Workflows

### 2.1 Instance Right-Sizing

Workflow for identifying oversized/undersized instances:

| Signal | Metric | Action |
|--------|--------|--------|
| Oversized | CPU avg < 15% for 7 days | Recommend downgrade |
| Oversized | Memory avg < 30% for 7 days | Recommend downgrade |
| Undersized | CPU avg > 85% for 1 hour | Recommend upgrade |
| Undersized | Memory avg > 90% for 1 hour | Recommend upgrade |

### 2.2 Billing Model Optimization

| Current Model | Recommendation Condition | Savings Potential |
|--------------|-------------------------|-------------------|
| PostPaid | Running > 30 days with steady load | 30-50% → PrePaid |
| PostPaid | Non-critical, interruptible workload | 60-90% → Spot |
| Spot | Production-critical workload | Risk → PrePaid |
| PrePaid | Expiring soon, no longer needed | Cancel renewal |

### 2.3 Idle Resource Detection

| Resource Type | Idle Criteria | Action |
|--------------|--------------|--------|
| ECS Instance | CPU < 5% + Network < 1KB for 7 days | Stop or delete |
| Cloud Disk | Detached > 7 days | Delete or snapshot + delete |
| EIP | Not bound to any resource > 24h | Release |
| TOS Bucket | No access for 90 days | Archive or delete |
| Snapshot | Older than parent image | Delete |

### 2.4 Storage Class Optimization

For object storage:

| Access Pattern | Current Class | Recommended Class | Savings |
|---------------|--------------|-------------------|---------|
| Not accessed > 30 days | Standard | IA | ~40% |
| Not accessed > 90 days | Standard/IA | Archive | ~60% |
| Not accessed > 180 days | Any | ColdArchive | ~80% |

### 2.5 Cleanup Workflows

Mandatory cleanup checks for each skill:

| Skill | Cleanup Target |
|-------|---------------|
| ve-ecs-ops | Stopped instances > 30 days, unattached disks, old snapshots, unused images |
| ve-tos-ops | Incomplete multipart uploads > 7 days, expired objects, orphaned delete markers |

---

## 3. Cost Measurement

### 3.1 Cost Reports

Skills SHOULD support generating cost summaries:

```markdown
## Cost Summary — [Date Range]

| Category | Current Cost | Trend | Optimization Opportunity |
|----------|-------------|-------|-------------------------|
| Compute | ¥X,XXX | ↑ 5% | 3 oversized instances (¥XXX/mo) |
| Storage | ¥X,XXX | → 0% | 2 idle buckets (¥XX/mo) |
| Network | ¥XXX | ↓ 2% | — |
```

### 3.2 ROI Indicators

| Indicator | Formula | Target |
|-----------|---------|--------|
| Resource Utilization | Actual usage / Provisioned capacity | > 60% |
| Cost per Request | Total cost / Total requests | Decreasing |
| Waste Ratio | Idle cost / Total cost | < 10% |

---

## FinOps Compliance Checklist

For all VE cloud skills:
- [ ] Cost attribution tags documented
- [ ] Right-sizing detection workflow included
- [ ] Idle resource detection included
- [ ] Billing model optimization guidance provided
- [ ] Cleanup workflow for orphaned resources
- [ ] Cost summary report pattern available

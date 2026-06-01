# Cost Optimization Guide — ve-billing-ops

> **Purpose:** Systematic cost optimization workflows for Volcengine resources based on billing data analysis.
> **Version:** 1.0.0
> **Last Updated:** 2026-06-02

---

## Table of Contents

1. [Cost Analysis Workflow](#1-cost-analysis-workflow)
2. [Right-Sizing Optimization](#2-right-sizing-optimization)
3. [Reserved Instance Optimization](#3-reserved-instance-optimization)
4. [Idle Resource Detection](#4-idle-resource-detection)
5. [Billing Model Optimization](#5-billing-model-optimization)
6. [Cost Anomaly Detection](#6-cost-anomaly-detection)

---

## 1. Cost Analysis Workflow

### 1.1 Multi-Dimensional Cost Breakdown

```bash
# Query monthly cost breakdown
ve billing DescribeBills --Period "{{user.start_date}}" --Region "{{env.VOLCENGINE_REGION}}"

# Query cost by product type
ve billing DescribeBillDetail --BillingCycle "{{user.start_date}}" --ProductType ecs

# Query cost by tag
ve billing DescribeBillDetail --BillingCycle "{{user.start_date}}" --TagFilters '[{"Key":"cost-center","Value":"engineering-infra"}]'
```

### 1.2 Cost Trend Analysis

| Metric | Formula | Alert Threshold |
|--------|---------|-----------------|
| MoM Growth | `(CurrentMonth - LastMonth) / LastMonth * 100` | > 20% → Warning |
| YoY Growth | `(CurrentMonth - LastYear) / LastYear * 100` | > 50% → Warning |
| Daily Burn Rate | `CurrentMonthCost / DayOfMonth * 100` | > (Budget/30) → Warning |
| Projected Month End | `DailyBurnRate * DaysInMonth` | > Budget → Critical |

### 1.3 Top N Cost Driver Analysis

```bash
# Parse bill detail to identify top 10 cost drivers
# Response: { "BillDetail": [{ "ProductType": "ecs", "ResourceId": "i-xxx", "Cost": 1234.56 }] }
# Sort by Cost descending, report Top N
```

---

## 2. Right-Sizing Optimization

### 2.1 Instance Right-Sizing Signals

| Signal | Metric | Threshold | Recommendation |
|--------|--------|-----------|----------------|
| CPU Underutilized | CPU avg < 15% for 7 days | Warning | Consider downgrading instance type |
| Memory Underutilized | Memory avg < 30% for 7 days | Warning | Consider downgrading instance type |
| CPU Overutilized | CPU avg > 85% for 1 hour | Warning | Consider upgrading instance type |
| Memory Overutilized | Memory avg > 90% for 1 hour | Critical | Consider upgrading instance type |

### 2.2 Right-Sizing Workflow

```markdown
## Right-Sizing Analysis Report

### Step 1: Identify High-Cost Resources
Query DescribeBillDetail, group by ResourceId, sort by Cost desc.

### Step 2: Cross-Reference Utilization
For Top 5 cost drivers, query Cloud Monitor for CPU/Memory metrics.

### Step 3: Generate Recommendations
| ResourceId | CurrentType | AvgCPU | AvgMem | Recommendation | Est.Savings |
|-----------|-------------|--------|--------|----------------|-------------|
| i-xxx | ecs.g6.xlarge | 12% | 25% | → ecs.g6.large | ¥XXX/mo |
```

### 2.3 Instance Type Upgrade/Downgrade Mapping

| Current Type | Downgrade To | Downgrade Condition | Upgrade To | Upgrade Condition |
|-------------|-------------|--------------------|-----------|-------------------|
| ecs.g6.2xlarge | ecs.g6.xlarge | CPU < 20% | — | — |
| ecs.g6.xlarge | ecs.g6.large | CPU < 15% | ecs.g6.2xlarge | CPU > 85% |
| ecs.c6.xlarge | ecs.c6.large | CPU < 15% | ecs.c6.2xlarge | CPU > 85% |

---

## 3. Reserved Instance Optimization

### 3.1 RI Coverage Analysis

```bash
# Query RI utilization
ve billing DescribeReservedInstances --Region "{{env.VOLCENGINE_REGION}}"

# Calculate RI Coverage
RI_Coverage = RI_Units_Used / RI_Units_Purchased * 100
OnDemand_Savings = OnDemand_Cost - RI_Cost
```

| RI Utilization | Status | Action |
|----------------|--------|--------|
| < 50% | Critical | Return unused RIs, analyze why underutilized |
| 50-70% | Warning | Optimize allocation or reduce RI purchases |
| 70-90% | Healthy | Monitor for upcoming renewal decisions |
| > 90% | Opportunity | Consider purchasing more RIs for coverage |

### 3.2 RI Renewal Decision Matrix

| Factor | Renewal Yes | Renewal No |
|--------|------------|-----------|
| Coverage | > 80% | < 50% |
| Trend | Stable/increasing workload | Decreasing workload |
| Price | RI price < OnDemand | RI price increase |
| Contract | Locked-in price beneficial | No significant savings |

---

## 4. Idle Resource Detection

### 4.1 Idle Criteria by Resource Type

| Resource Type | Idle Criteria | Investigation | Action |
|--------------|--------------|---------------|--------|
| ECS Instance | CPU < 5% + Network < 1KB for 7 days | Check if intended | Stop or delete |
| Cloud Disk | Detached > 7 days | Check if snapshot needed | Delete or snapshot + delete |
| EIP | Not bound to any resource > 24h | Check if intended | Release |
| TOS Bucket | No access for 90 days | Archive data first | Archive or delete |
| Snapshot | Older than parent image | Verify no dependency | Delete |
| Database (RDS) | No connections > 7 days | Verify backup policy | Stop (if supported) |

### 4.2 Cross-Reference Workflow

```markdown
## Idle Resource Detection Workflow

### Step 1: Query Bill Detail
ve billing DescribeBillDetail --BillingCycle "{{user.start_date}}" --ProductType ecs

### Step 2: Filter Low-Cost Resources
Resources with cost < ¥10/mo (configurable threshold)

### Step 3: Cross-Reference with Product APIs
- ECS: Query DescribeInstances, filter by Status=Stopped + long duration
- Disk: Query DescribeDisks, filter by Status=Available + unattached
- EIP: Query DescribeEips, filter by Status=Available

### Step 4: Generate Cleanup Report
| ResourceId | ResourceType | MonthlyCost | IdleDays | Recommendation |
|-----------|-------------|------------|---------|----------------|
| d-skp-xxx | disk | ¥50 | 30 | Delete |
```

---

## 5. Billing Model Optimization

### 5.1 PrePaid vs PostPaid Decision

| Condition | Current Model | Recommendation | Savings |
|----------|--------------|----------------|--------|
| Steady workload > 30 days | PostPaid | → PrePaid (Monthly/Yearly) | 30-50% |
| Production-critical workload | Spot | → PrePaid | (Stability priority) |
| Non-critical, interruptible | PostPaid | → Spot | 60-90% |
| Variable workload pattern | PrePaid | → PostPaid | (Flexibility priority) |

### 5.2 Storage Class Optimization

| Access Pattern | Current Class | Recommended Class | Savings |
|---------------|--------------|-------------------|---------|
| Not accessed > 30 days | Standard | InfrequentAccess | ~40% |
| Not accessed > 90 days | Standard/IA | Archive | ~60% |
| Not accessed > 180 days | Any | ColdArchive | ~80% |

### 5.3 Commitment-Based Discounts

| Resource | Commitment Type | Savings vs On-Demand |
|---------|----------------|--------------------|
| VKE Node | 1-Year Reserved | 40-60% |
| VKE Node | 3-Year Reserved | 65-75% |
| RDS | 1-Year Reserved | 35-50% |
| Redis | 1-Year Reserved | 40-55% |

---

## 6. Cost Anomaly Detection

### 6.1 Anomaly Patterns

| Pattern | Detection Logic | Severity | Investigation |
|--------|---------------|----------|---------------|
| Sudden Spike | `Cost > LastMonth * 1.5` | Critical | New resource? Scaling event? |
| Gradual Creep | `MoM Growth > 10%` for 3 consecutive months | Warning | Usage growth or price change? |
| Unusual Pattern | Weekday/Weekend ratio change | Medium | Changed deployment pattern? |
| Tag Gap | Resource without cost-center tag | Low | Add tag for attribution |

### 6.2 Investigation Workflow

```markdown
## Cost Anomaly Investigation

### Step 1: Confirm Anomaly
- Query DescribeBills for current and previous periods
- Calculate percentage change
- Verify against budget threshold

### Step 2: Identify Driver
- Query DescribeBillDetail grouped by ProductType
- Identify ProductType with largest cost change
- Drill down by ResourceId

### Step 3: Root Cause Analysis
- New resource created?
- Configuration change?
- Usage pattern change?
- Price/rate change?

### Step 4: Remediation
- Right-sizing recommendations
- Cleanup idle resources
- Budget adjustment
```

---

## Optimization Report Template

```markdown
# Cost Optimization Report — [Date Range]

## Executive Summary
| Metric | Value | Change |
|--------|-------|--------|
| Total Cost | ¥XXX,XXX | +X% MoM |
| Top Cost Driver | Product/Resource | ¥XXX |
| Optimization Opportunity | Type | ¥XXX/mo |

## 1. Cost Breakdown
| Category | Current | Last Month | Change | % of Total |
|----------|---------|-----------|--------|-----------|
| Compute | ¥XX,XXX | ¥XX,XXX | +X% | XX% |
| Storage | ¥X,XXX | ¥X,XXX | +X% | XX% |
| Network | ¥XXX | ¥XXX | +X% | XX% |

## 2. Right-Sizing Opportunities
| ResourceId | CurrentType | AvgUtil | Recommendation | Est.Savings |
|-----------|-------------|---------|----------------|-------------|
| i-xxx | ecs.g6.xlarge | CPU 12% | → ecs.g6.large | ¥XXX/mo |

## 3. Idle Resource Cleanup
| ResourceId | Type | MonthlyCost | IdleDays | Action |
|-----------|------|------------|---------|--------|
| d-skp-xxx | disk | ¥50 | 30 | Delete |

## 4. RI Optimization
| Metric | Value |
|--------|-------|
| RI Coverage | XX% |
| On-Demand Cost | ¥XXX |
| RI Cost | ¥XXX |
| Potential Savings | ¥XXX |

## 5. Recommended Actions
| Priority | Action | Est.Savings | Owner |
|---------|--------|------------|-------|
| P1 | Right-size 5 oversized instances | ¥X,XXX/mo | @team |
| P2 | Delete 3 idle disks | ¥XXX/mo | @team |
| P3 | Purchase 1-year RI for VKE | ¥X,XXX/mo | @team |
```

---

## Cost Tag Convention

| Tag Key | Description | Example |
|--------|-------------|---------|
| `cost-center` | Cost allocation unit | `engineering-infra` |
| `project` | Project identifier | `project-alpha` |
| `environment` | Deployment environment | `production`, `staging`, `dev` |
| `owner` | Resource owner | `team-backend` |
| `managed-by` | Provisioning tool | `terraform`, `manual`, `ai-agent` |

---

## ROI Indicators

| Indicator | Formula | Target |
|-----------|---------|--------|
| Resource Utilization | Actual usage / Provisioned capacity | > 60% |
| Cost per Request | Total cost / Total requests | Decreasing |
| Waste Ratio | Idle cost / Total cost | < 10% |
| RI Coverage | RI units used / Total units | > 80% |
| PrePay Savings | (OnDemand - PrePaid) / OnDemand | > 30% |
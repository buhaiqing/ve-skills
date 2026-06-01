# Cross-Service Cost Analysis — ve-billing-ops

> **Purpose:** Cost attribution and correlation analysis across multiple Volcengine services. Enable cross-service cost visibility and optimization.
> **Version:** 1.0.0> **Last Updated:** 2026-06-02

---

## Table of Contents

1. [Cross-Service Cost Model](#1-cross-service-cost-model)
2. [Service Correlation Patterns](#2-service-correlation-patterns)
3. [Multi-Service Cost Breakdown](#3-multi-service-cost-breakdown)
4. [Shared Cost Allocation](#4-shared-cost-allocation)
5. [Integration with Other VE Skills](#5-integration-with-other-ve-skills)

---

## 1. Cross-Service Cost Model

### 1.1 Application Cost Hierarchy

```
┌─────────────────────────────────────────────────────────────────┐
│                        Application                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐        │
│  │   Frontend   │  │   Backend    │  │   Database   │        │
│  │   (ALB/CLB)  │  │   (ECS/VKE)  │  │   (RDS/Redis)│        │
│  └──────────────┘  └──────────────┘  └──────────────┘        │
│         │                │                   │                │
│         └────────────────┼───────────────────┘                │
│                          │                                     │
│              ┌───────────▼───────────┐                        │
│              │   Network Layer       │                        │
│              │   (VPC/EIP/NAT/VPN)   │                        │
│              └───────────┬───────────┘                        │
│                          │                                     │
│              ┌───────────▼───────────┐                        │
│              │   Storage Layer       │                        │
│              │   (NAS/TOS/Snapshot)  │                        │
│              └───────────────────────┘                        │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 Cost Attribution by Layer

| Layer | Services | Cost Attribution Key |
|-------|----------|----------------------|
| **Application** | ALB, CLB, ECS, VKE, ARK | Environment, Project |
| **Database** | RDS-MySQL, RDS-PostgreSQL, Redis, MongoDB | Database name, Environment |
| **Storage** | NAS, TOS, Snapshot | Application, Environment |
| **Network** | VPC, EIP, NAT, VPN, CDN | Network zone, Usage pattern |

---

## 2. Service Correlation Patterns

### 2.1 Web Application Pattern

| Component | Primary Service | Cost Driver | Related Costs |
|-----------|----------------|-------------|---------------|
| Load Balancer | ALB/CLB | Request count, Traffic | ECS instances behind |
| API Servers | ECS/VKE | Instance hours, CPU | RDS connections |
| Database | RDS/Redis | Instance hours, Storage, IOPS | Snapshot storage |
| Static Assets | TOS + CDN | Storage, Traffic | CDN traffic |

**Cost Formula:**
```
Total_App_Cost = ALB_Cost + ECS_Cost + RDS_Cost + Redis_Cost + TOS_Cost + CDN_Cost
```

### 2.2 Data Processing Pattern

| Component | Primary Service | Cost Driver | Related Costs |
|-----------|----------------|-------------|---------------|
| Batch Compute | VKE/ECS | Node hours | VKE cluster management |
| Message Queue | Kafka | Instance hours, Storage | Network traffic |
| State Store | RDS/Redis | Instance hours, Storage | Snapshot storage |
| Stream Storage | SLS | Ingested data, Retention | Search storage |

### 2.3 Dev/Test Environment Pattern

| Component | Shared Cost | Allocation Method |
|-----------|------------|-------------------|
| VPC | Network costs | Equal split or tag-based |
| NAT Gateway | Outbound traffic | Actual usage |
| DNS | Domain count | Equal split |
| KMS | API calls | Equal split |
| Security Group | Per SG | By environment |

---

## 3. Multi-Service Cost Breakdown

### 3.1 Query Cross-Service Costs

```bash
# Query costs by product type for current month
CURRENT_MONTH=$(date +%Y-%m)

# ECS costs
ECS_COST=$(ve billing DescribeBillDetail --BillingCycle "$CURRENT_MONTH" --ProductType ecs --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.TotalCost')

# RDS costs
RDS_COST=$(ve billing DescribeBillDetail --BillingCycle "$CURRENT_MONTH" --ProductType rds --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.TotalCost')

# VKE costs
VKE_COST=$(ve billing DescribeBillDetail --BillingCycle "$CURRENT_MONTH" --ProductType vke --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.TotalCost')

# Redis costs
REDIS_COST=$(ve billing DescribeBillDetail --BillingCycle "$CURRENT_MONTH" --ProductType redis --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.TotalCost')

# TOS costs
TOS_COST=$(ve billing DescribeBillDetail --BillingCycle "$CURRENT_MONTH" --ProductType tos --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.TotalCost')

echo "=== Cost Summary ==="
echo "ECS: ¥$ECS_COST"
echo "RDS: ¥$RDS_COST"
echo "VKE: ¥$VKE_COST"
echo "Redis: ¥$REDIS_COST"
echo "TOS: ¥$TOS_COST"
```

### 3.2 Cross-Service Report Template

```markdown
# Cross-Service Cost Report — [Period]

## Application Cost Summary

| Application | ECS | RDS | VKE | Redis | Network | Storage | Total |
|-------------|-----|-----|-----|-------|---------|---------|-------|
| App-A | ¥X | ¥X | ¥X | ¥X | ¥X | ¥X | ¥X |
| App-B | ¥X | ¥X | ¥X | — | ¥X | ¥X | ¥X |
| Shared | — | — | — | — | ¥X | ¥X | ¥X |
| **Total** | ¥X | ¥X | ¥X | ¥X | ¥X | ¥X | ¥X |

## Cost per User/Request (if available)

| Application | Requests/Month | Total Cost | Cost/Request |
|-------------|---------------|------------|--------------|
| App-A | XXX,XXX | ¥X,XXX | ¥0.XXX |
| App-B | XXX,XXX | ¥X,XXX | ¥0.XXX |

## Trend Analysis

| Month | App-A | App-B | Shared | Total | MoM |
|-------|-------|-------|--------|-------|-----|
| Jan | ¥X | ¥X | ¥X | ¥X | — |
| Feb | ¥X | ¥X | ¥X | ¥X | +X% |
```

---

## 4. Shared Cost Allocation

### 4.1 Shared Resource Types

| Resource | Why Shared | Allocation Method |
|----------|-----------|-------------------|
| VPC | Network infrastructure | By subnet usage |
| NAT Gateway | Outbound traffic | By actual traffic |
| EIP | Static IP | By resource binding |
| Security Group | Firewall rules | By rule count |
| KMS | Encryption keys | By API call volume |
| Cloud DNS | Domain management | By zone count |
| KMS | Key management | Equal split |

### 4.2 Allocation Formulas

```markdown
## NAT Gateway Cost Allocation

NAT_Cost_Per_Team = (Team_Outbound_Traffic / Total_Outbound_Traffic) * NAT_Gateway_Cost

## VPC Cost Allocation

VPC_Cost_Per_Team = (Team_Subnet_CIDRs / Total_VPC_CIDRs) * VPC_Base_Cost

## Shared Storage Cost Allocation

Storage_Cost_Per_Team = Team_Storage_GB / Total_Storage_GB * Shared_Storage_Cost
```

### 4.3 Shared Cost Report

```markdown
## Shared Cost Allocation — [Period]

| Cost Type | Total | Team-A | Team-B | Team-C | Unallocated |
|-----------|-------|--------|--------|--------|-------------|
| VPC | ¥X | ¥X (XX%) | ¥X (XX%) | ¥X (XX%) | ¥X (XX%) |
| NAT | ¥X | ¥X (XX%) | ¥X (XX%) | ¥X (XX%) | ¥X (XX%) |
| KMS | ¥X | ¥X (33%) | ¥X (33%) | ¥X (33%) | — |
| DNS | ¥X | ¥X (40%) | ¥X (35%) | ¥X (25%) | — |
| **Total** | ¥X | ¥X | ¥X | ¥X | ¥X |

## Unallocated Cost Investigation
- Reason: Resources without cost-center tag
- Action: Tag untagged resources
```

---

## 5. Integration with Other VE Skills

### 5.1 Cost Query by Resource Tag

```bash
# Query costs filtered by tag (requires tag-based billing enabled)
ve billing DescribeBillDetail \
  --BillingCycle "$(date +%Y-%m)" \
  --TagFilters '[{"Key":"environment","Value":"production"}]' \
  --Region "{{env.VOLCENGINE_REGION}}"
```

### 5.2 Orchestrated Cross-Skill Cost Analysis

```markdown
## Cross-Skill Cost Analysis Workflow

### Step 1: Cost Overview (ve-billing-ops)
Query DescribeBills for total spend and product breakdown.

### Step 2: Top Cost Resources (ve-billing-ops)
Query DescribeBillDetail for Top 10 resource IDs.

### Step 3: Resource Details (Product Skills)
For each Top N resource, query respective product skill:
- ECS: ve ecs-ops DescribeInstances → Get instance specs
- RDS: ve rds-ops DescribeInstances → Get DB specs
- VKE: ve vke-ops DescribeClusters → Get cluster config

### Step 4: Utilization Check (Product Skills)
Query Cloud Monitor for utilization metrics.

### Step 5: Optimization Recommendations
Generate combined report with:
- Right-sizing recommendations
- RI purchase recommendations
- Cleanup recommendations
```

### 5.3 Skill Correlation Matrix

| Billing Query | Primary Skill | Integration Command |
|--------------|---------------|---------------------|
| ECS cost analysis | ve-ecs-ops | `ve ecs DescribeInstances` |
| RDS cost analysis | ve-rds-ops / ve-rds-mysql-ops | `ve rds DescribeInstances` |
| VKE cost analysis | ve-vke-ops | `ve vke DescribeClusters` |
| Redis cost analysis | ve-redis-ops | `ve redis DescribeInstances` |
| TOS cost analysis | ve-tos-ops | `ve tos ListBuckets` |
| Network cost analysis | ve-vpc-ops, ve-eip-ops, ve-nat-ops | `ve vpc DescribeVpcs`, etc. |

### 5.4 Combined Optimization Workflow

```markdown
# Combined Cost Optimization Workflow

## 1. Identify (ve-billing-ops)
- Query Top 10 cost resources
- Identify oversized/undersized candidates

## 2. Analyze (Product Skills)
For each candidate:
```
### 2.1 ECS Candidates
Command: ve ecs DescribeInstances --InstanceIds i-xxx,i-yyy
Analysis: CPU/Memory utilization, instance type

### 2.2 RDS Candidates
Command: ve rds DescribeDBInstances --DBInstanceIds drm-xxx
Analysis: CPU/Memory, storage, connections

### 2.3 VKE Candidates
Command: ve vke DescribeClusters
Analysis: Node count, node specs, pod count
```

## 3. Recommend
| Resource | Current | Recommendation | Est.Savings |
|---------|---------|----------------|-------------|
| i-xxx | ecs.g6.2xlarge | → ecs.g6.xlarge | ¥XXX/mo |
| drm-yyy | rds.mssql.4xlarge | → rds.mssql.2xlarge | ¥XXX/mo |

## 4. Execute (Product Skills)
Execute changes via respective product skills.

## 5. Verify
Re-query billing after 24h for cost change confirmation.
```

---

## Cost Anomaly Cross-Service Investigation

### Scenario: Unexpected Cost Spike

```markdown
## Investigation Playbook

### Day 1: Detection
ve billing DescribeBills → Cost +50% MoM

### Day 2: Attribution
Query DescribeBillDetail:
- Group by ProductType
- Identify highest increase

### Day 3: Root Cause
Cross-reference with product skills:
- New instances created?
- Scaling events?
- Data transfer spike?
- Storage growth?

### Day 4: Resolution
- Stop unexpected resources
- Adjust configurations
- Update budgets if needed
```

### Investigation Command Reference

```bash
# 1. Get overall cost delta
ve billing DescribeBills --Period "$(date +%Y-%m)" --Region "{{env.VOLCENGINE_REGION}}"
ve billing DescribeBills --Period "$(date -v-1m +%Y-%m)" --Region "{{env.VOLCENGINE_REGION}}"

# 2. Drill down by product
for PRODUCT in ecs rds vke redis tos alb clb eip nas; do
  echo "=== $PRODUCT ==="
  ve billing DescribeBillDetail --BillingCycle "$(date +%Y-%m)" --ProductType "$PRODUCT" --Region "{{env.VOLCENGINE_REGION}}" | \
    jq -r '.Result.TotalCost'
done

# 3. Cross-reference with product APIs
# ECS: ve ecs DescribeInstances | count new instances
# RDS: ve rds DescribeDBInstances | check connections
# VKE: ve vke DescribeClusters | check node scaling
```
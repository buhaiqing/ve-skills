# Cross-Skill Orchestration — ve-billing-ops

> **Purpose:** Define integration patterns and orchestration flows between ve-billing-ops and other VE product skills for unified FinOps operations.
> **Version:** 1.0.0>
> **Last Updated:** 2026-06-02

---

## Table of Contents

1. [Orchestration Architecture](#1-orchestration-architecture)
2. [Skill Integration Matrix](#2-skill-integration-matrix)
3. [Orchestrated Workflows](#3-orchestrated-workflows)
4. [Cost-Aware Operation Hooks](#4-cost-aware-operation-hooks)
5. [Unified Dashboard Data Sources](#5-unified-dashboard-data-sources)

---

## 1. Orchestration Architecture

### 1.1 Skill Communication Pattern

```
┌─────────────────────────────────────────────────────────────────────┐
│                        User Request                                │
└───────────────────────────────┬───────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      ve-billing-ops (Orchestrator)                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐   │
│  │ Cost Query  │  │ Anomaly     │  │ Optimization            │   │
│  │ & Analysis │  │ Detection   │  │ Recommendation          │   │
│  └──────┬──────┘  └──────┬──────┘  └───────────┬─────────────┘   │
└─────────┼────────────────┼─────────────────────┼──────────────────┘
          │                │                     │
          ▼                ▼                     ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     Product Skills (Executors)                      │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌─────────────────┐  │
│  │ ve-ecs-ops│  │ ve-rds-ops│  │ ve-vke-ops│  │ ve-tos-ops      │  │
│  └───────────┘  └───────────┘  └───────────┘  └─────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

### 1.2 Orchestration Principles

| Principle | Description | Implementation |
|-----------|-------------|----------------|
| **Cost-First** | Every product operation should consider cost impact | Pre-check cost before changes |
| **Attribution** | All operations must support cost tagging | Enforce tags on create/modify |
| **Optimization** | Product skills should expose optimization opportunities | Return cost recommendations |
| **Audit** | All cost-affecting operations logged | Record in operation audit |

---

## 2. Skill Integration Matrix

### 2.1 Integration Points

| Product Skill | Primary Integration | Commands | Cost Data |
|--------------|-------------------|----------|-----------|
| **ve-ecs-ops** | Instance cost analysis | DescribeInstances | Instance hours, type |
| **ve-rds-ops** | Database cost analysis | DescribeDBInstances | Instance + storage |
| **ve-vke-ops** | Kubernetes cost analysis | DescribeClusters, DescribeNodePools | Node hours |
| **ve-redis-ops** | Cache cost analysis | DescribeInstances | Instance + memory |
| **ve-tos-ops** | Storage cost analysis | ListBuckets, GetBucketStat | Storage + requests |
| **ve-sls-ops** | Log cost analysis | GetProjectStatistics | Ingest + storage |
| **ve-alb-ops** | Load balancer cost | DescribeLoadBalancers | LCU/requests |
| **ve-clb-ops** | CLB cost analysis | DescribeLoadBalancers | LCU/bandwidth |
| **ve-eip-ops** | EIP cost analysis | DescribeEips | Bandwidth + hours |
| **ve-nas-ops** | NAS cost analysis | DescribeFileSystems | Storage + throughput |
| **ve-kafka-ops** | Kafka cost analysis | DescribeInstances | Instance + storage |
| **ve-cdn-ops** | CDN cost analysis | DescribeDomains | Traffic + requests |

### 2.2 Data Flow Summary

```markdown
## ve-billing-ops → Product Skills

| From | To | Data | Trigger |
|------|----|------|---------|
| Cost Analysis | ve-ecs-ops | Top N instance IDs | High ECS cost |
| Cost Analysis | ve-rds-ops | High-cost DB instances | High RDS cost |
| Cost Analysis | ve-vke-ops | Cluster node counts | High VKE cost |
| Idle Detection | Product skills | Resource IDs to investigate | Idle resource found |

## Product Skills → ve-billing-ops

| From | To | Data | Trigger |
|------|----|------|---------|
| ve-ecs-ops | Cost Analysis | Instance specs for right-sizing | User request |
| ve-rds-ops | Cost Analysis | DB specs for optimization | User request |
| ve-ecs-ops | Cost Attribution | Resource tags | Tag compliance check |
| Product skills | Anomaly Detection | Unexpected changes | Config changes |
```

---

## 3. Orchestrated Workflows

### 3.1 Right-Sizing Workflow

```markdown
## Orchestrated Right-Sizing Workflow

### Step 1: ve-billing-ops — Identify High-Cost Resources
```bash
ve billing DescribeBillDetail --BillingCycle "$(date +%Y-%m)" --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '.Result.BillDetails | sort_by(.Cost) | reverse | .[:10]'
```
Output: Top 10 resource IDs with costs

### Step 2: ve-billing-ops → Product Skills — Get Specifications
```bash
# For ECS resources
ve ecs DescribeInstances --InstanceIds i-xxx,i-yyy,i-zzz --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '.Result.Instances[] | {InstanceId, InstanceType, CPU, Memory, Status}'

# For RDS resources
ve rds DescribeDBInstances --DBInstanceIds drm-xxx --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '.Result.DBInstances[] | {DBInstanceId, DBInstanceType, DBInstanceClass, Storage}'

# For VKE clusters
ve vke DescribeClusters --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '.Result.Clusters[] | {ClusterId, NodeCount, NodePoolSpecs}'
```

### Step 3: Product Skills — Query Utilization (via CMS)
```bash
# Query CPU/Memory for ECS instances
ve cms GetMetricStatistics --Namespace acs_ecs_dashboard --MetricName cpu_total --Dimensions '[{"instanceId":"i-xxx"}]'
ve cms GetMetricStatistics --Namespace acs_ecs_dashboard --MetricName.memory_usedutilization --Dimensions '[{"instanceId":"i-xxx"}]'
```

### Step 4: ve-billing-ops — Generate Recommendations
| ResourceId | CurrentType | AvgCPU | AvgMem | Recommendation | Est.Savings |
|-----------|-------------|--------|--------|----------------|-------------|
| i-xxx | ecs.g6.2xlarge | 12% | 25% | → ecs.g6.large | ¥350/mo |
| drm-yyy | rds.mysql.4xlarge | 15% | 30% | → rds.mysql.2xlarge | ¥800/mo |

### Step 5: Execute via Product Skills
```bash
# Execute ECS right-size
ve ecs ModifyInstanceSpec --InstanceId i-xxx --InstanceType ecs.g6.large

# Execute RDS right-size
ve rds ModifyDBInstanceSpec --DBInstanceId drm-yyy --DBInstanceClass rds.mysql.2xlarge
```

### Step 6: Verify Cost Change
```bash
# Check after 24-48 hours
ve billing DescribeBillDetail --BillingCycle "$(date +%Y-%m)" --ResourceIds i-xxx | \
  jq '.Result.BillDetails[0].Cost'
```

### HALT Conditions
- Recommendation savings < ¥50/month
- Risk level > High
- Production critical instance without migration plan
```

### 3.2 Idle Resource Cleanup Workflow

```markdown
## Orchestrated Idle Resource Cleanup

### Step 1: ve-billing-ops — Identify Low-Cost Resources
```bash
ve billing DescribeBillDetail --BillingCycle "$(date +%Y-%m)" --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '.Result.BillDetails | map(select(.Cost < 100)) | .[].ResourceId'
```

### Step 2: Product Skills — Verify Resource Status
```bash
# Check ECS status
ve ecs DescribeInstances --InstanceIds $RESOURCE_IDS --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '.Result.Instances[] | select(.Status == "Stopped") | {InstanceId, InstanceType, StoppedDuration}'

# Check unattached disks
ve ecs DescribeDisks --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '.Result.Disks[] | select(.Status == "Available") | {DiskId, Capacity, CreationTime}'

# Check unbound EIPs
ve eip DescribeEips --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '.Result.Eips[] | select(.InstanceId == null) | {EipId, Bandwidth, CreationTime}'
```

### Step 3: ve-billing-ops — Risk Assessment
| Resource | Cost/Month | IdleDays | Risk | Action |
|----------|-----------|---------|------|--------|
| i-xxx | ¥50 | 45 | Low | Delete |
| d-disk-yyy | ¥30 | 30 | Low | Snapshot + Delete |
| eip-zzz | ¥20 | 15 | Medium | Keep (reserved) |

### Step 4: Execute Cleanup (Product Skills)
```bash
# Delete idle ECS
ve ecs DeleteInstance --InstanceId i-xxx

# Delete unattached disk (with snapshot if valuable)
ve ecs CreateSnapshot --DiskId d-disk-yyy --SnapshotName "backup-$(date +%Y%m%d)"
ve ecs DeleteDisk --DiskId d-disk-yyy

# Release EIP
ve eip ReleaseEipAddress --EipId eip-zzz
```

### Step 5: Verification
```bash
# Verify resources deleted
ve ecs DescribeInstances --InstanceIds i-xxx  # Should return empty
ve billing DescribeBillDetail --ResourceIds i-xxx  # Should reduce
```
```

### 3.3 Cost Anomaly Investigation Workflow

```markdown
## Cross-Skill Anomaly Investigation

### Trigger: ve-billing-ops detects cost spike > 50% MoM

### Step 1: Identify Anomaly Scope
```bash
# Get product breakdown
ve billing DescribeBillDetail --BillingCycle "$ANOMALY_MONTH" | \
  jq '.Result.BillDetails | group_by(.ProductType) | map({product: .[0].ProductType, total: (map(.Cost) | add)})'
```

### Step 2: Delegate to Product Skills for Root Cause

**For ECS spike:**
```bash
ve ecs DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --CreationDateStart "$ANOMALY_MONTH-01" | \
  jq '.Result.Instances | length'  # Count new instances
```

**For RDS spike:**
```bash
ve rds DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '.Result.DBInstances[] | {DBInstanceId, CreationTime, StorageUsed}'
```

**For VKE spike:**
```bash
ve vke DescribeClusters --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '.Result.Clusters[] | {ClusterId, NodeCount, CreationTime}'
```

**For TOS spike:**
```bash
ve tos GetBucketStat --BucketName $BUCKET | \
  jq '{StorageSize, ObjectCount, LastModified}'
```

### Step 3: Correlate with Deployment Logs
Cross-reference timestamps with deployment systems (CI/CD, Terraform state changes)

### Step 4: Generate Root Cause Report
| Finding | Evidence | Owner | Action |
|---------|----------|-------|--------|
| New VKE cluster created | 3 nodes × ecs.g6.xlarge added on 15th | @team-alpha | Tag + Budget |
| Storage auto-scaled | 500GB added to TOS bucket | @team-beta | Review lifecycle |
| RDS backup retention increased | +200GB backup storage | @team-gamma | Adjust retention |
```

---

## 4. Cost-Aware Operation Hooks

### 4.1 Pre-Operation Cost Check

```bash
#!/bin/bash
# Pre-operation cost impact estimation

OPERATION="{{user.operation}}"  # create, modify, delete
RESOURCE_TYPE="{{user.resource_type}}"  # ecs, rds, vke, etc.

case $RESOURCE_TYPE in
  ecs)
    INSTANCE_TYPE="{{user.instance_type}}"  # e.g., ecs.g6.xlarge
    ESTIMATED_COST=$(calculate_ecs_cost "$INSTANCE_TYPE")
    ;;
  rds)
    DB_CLASS="{{user.db_class}}"
    STORAGE_GB="{{user.storage}}"
    ESTIMATED_COST=$(calculate_rds_cost "$DB_CLASS" "$STORAGE_GB")
    ;;
  vke)
    NODE_COUNT="{{user.node_count}}"
    NODE_TYPE="{{user.node_type}}"
    ESTIMATED_COST=$(calculate_vke_cost "$NODE_COUNT" "$NODE_TYPE")
    ;;
esac

# Check against budget
BUDGET_REMAINING=$(ve billing DescribeBudgets --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '[.Result.Budgets[] | .BudgetAmount - .ActualSpent] | add')

if (( $(echo "$ESTIMATED_COST > $BUDGET_REMAINING * 0.2" | bc -l) )); then
  echo "WARNING: This operation adds estimated ¥$ESTIMATED_COST/month"
  echo "         Budget remaining: ¥$BUDGET_REMAINING"
  echo "Proceed? (y/N)"
  read -r confirm
  [ "$confirm" != "y" ] && exit 1
fi
```

### 4.2 Post-Operation Cost Tracking

```bash
#!/bin/bash
# Post-operation cost tagging and tracking

RESOURCE_ID="{{output.resource_id}}"
RESOURCE_TYPE="{{user.resource_type}}"
PROJECT="{{user.project}}"
ENVIRONMENT="{{user.environment}}"
COST_CENTER="{{user.cost_center}}"

# Apply cost tags
case $RESOURCE_TYPE in
  ecs)
    ve ecs TagResources --ResourceIds "$RESOURCE_ID" \
      --Tags "project=$PROJECT,environment=$ENVIRONMENT,cost-center=$COST_CENTER,managed-by=ai-agent" \
      --Region "{{env.VOLCENGINE_REGION}}"
    ;;
  rds)
    ve rds TagResources --ResourceIds "$RESOURCE_ID" \
      --Tags "project=$PROJECT,environment=$ENVIRONMENT,cost-center=$COST_CENTER,managed-by=ai-agent" \
      --Region "{{env.VOLCENGINE_REGION}}"
    ;;
  vke)
    # Apply tags to cluster nodes
    ve vke AddTags --ClusterId "$RESOURCE_ID" \
      --Tags "project=$PROJECT,environment=$ENVIRONMENT,cost-center=$COST_CENTER"
    ;;
esac

echo "Cost attribution tags applied: project=$PROJECT, environment=$ENVIRONMENT, cost-center=$COST_CENTER"
```

---

## 5. Unified Dashboard Data Sources

### 5.1 Dashboard Query Pattern

```bash
#!/bin/bash
# Generate unified FinOps dashboard data

DASHBOARD_DATE="{{user.date}}"  # YYYY-MM-DD

echo "=== FinOps Dashboard Data ==="
echo "Generated: $(date)"
echo "Period: $DASHBOARD_DATE"

# 1. Overall Cost Summary
echo ""
echo "--- Cost Summary ---"
ve billing DescribeBills --Period "${DASHBOARD_DATE:0:7}" --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '{total: .Result.Bills[0].TotalCost, product_count: (.Result.Bills[0] | keys | length)}'

# 2. Cost by Product (from billing)
echo ""
echo "--- Cost by Product ---"
ve billing DescribeBillDetail --BillingCycle "${DASHBOARD_DATE:0:7}" --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '[.Result.BillDetails | group_by(.ProductType)[] | {product: .[0].ProductType, total: (map(.Cost) | add)}]'

# 3. Resource Counts (from product skills)
echo ""
echo "--- Resource Counts ---"
echo "ECS Instances: $(ve ecs DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" 2>/dev/null | jq '.Result.Instances | length')"
echo "RDS Instances: $(ve rds DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}" 2>/dev/null | jq '.Result.DBInstances | length')"
echo "VKE Clusters: $(ve vke DescribeClusters --Region "{{env.VOLCENGINE_REGION}}" 2>/dev/null | jq '.Result.Clusters | length')"
echo "TOS Buckets: $(ve tos ListBuckets --Region "{{env.VOLCENGINE_REGION}}" 2>/dev/null | jq '.Result.Buckets | length')"

# 4. Budget Status
echo ""
echo "--- Budget Status ---"
ve billing DescribeBudgets --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '.Result.Budgets[] | {name: .BudgetName, limit: .BudgetAmount, actual: .ActualSpent, percent: (.ActualSpent / .BudgetAmount * 100 | . * 100 | . / 100)}'

# 5. RI Utilization
echo ""
echo "--- RI Utilization ---"
ve billing DescribeReservedInstances --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '[.Result.ReservedInstances[] | {type: .InstanceType, total: .TotalUnits, used: .UsedUnits, utilization: (.UsedUnits / .TotalUnits * 100 | . * 100 | . / 100)}]'
```

### 5.2 Data Schema for Dashboard Integration

```markdown
## FinOps Dashboard Data Schema

### CostSummary
```json
{
  "period": "2026-06",
  "total_cost": 150000.00,
  "currency": "CNY",
  "product_breakdown": {
    "ecs": 50000.00,
    "rds": 35000.00,
    "vke": 40000.00,
    "tos": 15000.00,
    "other": 10000.00
  }
}
```

### ResourceInventory
```json
{
  "timestamp": "2026-06-02T10:00:00Z",
  "products": [
    {
      "product": "ecs",
      "count": 50,
      "by_type": {
        "ecs.g6.xlarge": 20,
        "ecs.g6.large": 25,
        "ecs.c6.2xlarge": 5
      },
      "cost": 50000.00
    }
  ]
}
```

### BudgetAlert
```json
{
  "budget_name": "monthly-production",
  "limit": 100000.00,
  "actual": 85000.00,
  "utilization_percent": 85.0,
  "projected_end_of_month": 110000.00,
  "status": "warning",
  "days_remaining": 15
}
```
```

### 5.3 Alert Integration with CMS

```bash
#!/bin/bash
# Push billing alerts to Cloud Monitor

BUDGET_NAME="{{user.budget_name}}"
UTILIZATION=$(ve billing DescribeBudgets --Region "{{env.VOLCENGINE_REGION}}" | \
  jq -r ".Result.Budgets[] | select(.BudgetName == \"$BUDGET_NAME\") | .ActualSpent / .BudgetAmount * 100")

# Determine severity
if (( $(echo "$UTILIZATION >= 100" | bc -l) )); then
  SEVERITY="critical"
  MESSAGE="Budget $BUDGET_NAME EXCEEDED!"
elif (( $(echo "$UTILIZATION >= 90" | bc -l) )); then
  SEVERITY="high"
  MESSAGE="Budget $BUDGET_NAME at ${UTILIZATION}%"
elif (( $(echo "$UTILIZATION >= 80" | bc -l) )); then
  SEVERITY="warning"
  MESSAGE="Budget $BUDGET_NAME at ${UTILIZATION}%"
else
  echo "Budget utilization normal: ${UTILIZATION}%"
  exit 0
fi

# Push to CMS
ve cms PutMetricData --Metrics "[{
  \"MetricName\": \"BudgetUtilization\",
  \"Value\": $UTILIZATION,
  \"Dimensions\": {\"BudgetName\": \"$BUDGET_NAME\"},
  \"Timestamp\": $(date +%s)
}]"

echo "CMS alert pushed: $MESSAGE (Severity: $SEVERITY)"
```
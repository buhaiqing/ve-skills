# Billing Operations Framework — ve-billing-ops

> **Purpose:** Standardized periodic inspection framework for Volcengine billing operations. Defines daily, weekly, and monthly review patterns.
> **Version:** 1.0.0
> **Last Updated:** 2026-06-02

---

## Table of Contents

1. [Overview](#1-overview)
2. [Daily Inspection](#2-daily-inspection)
3. [Weekly Inspection](#3-weekly-inspection)
4. [Monthly Inspection](#4-monthly-inspection)
5. [Alert Response Playbook](#5-alert-response-playbook)
6. [Report Templates](#6-report-templates)

---

## 1. Overview

### Inspection Cadence

| Frequency | Focus | SLA |
|-----------|-------|-----|
| **Daily** | Budget alerts, balance warnings, urgent anomalies | Within 2 hours |
| **Weekly** | Cost trends, Top N cost drivers, RI coverage | Within 4 hours |
| **Monthly** | Full bill analysis, optimization planning, RI renewal | Within 1 day |

### Inspection Team Roles

| Role | Responsibility |
|------|----------------|
| **FinOps Engineer** | Daily/Weekly reviews, anomaly investigation |
| **Finance/Manager** | Monthly review approval, budget adjustments |
| **DevOps/Team Lead** | Team-specific optimization execution |

---

## 2. Daily Inspection

### 2.1 Pre-Flight Check

```bash
# Verify environment
echo "Checking credentials..."
[ -z "$VOLCENGINE_ACCESS_KEY" ] && echo "ERROR: VOLCENGINE_ACCESS_KEY not set" && exit 1
[ -z "$VOLCENGINE_SECRET_KEY" ] && echo "ERROR: VOLCENGINE_SECRET_KEY not set" && exit 1

# Verify CLI
ve version
```

### 2.2 Daily Checklist

| Check | Command | Threshold | Severity |
|-------|---------|-----------|----------|
| Budget Alert Check | `ve billing DescribeBudgets` | Any alert triggered | Critical |
| Balance Check | `ve billing DescribeBalance` | < 30 days of burn | Warning |
| Yesterday's Cost | `ve billing DescribeBillDetail` (yesterday) | > Daily budget | Warning |
| Critical Anomaly | Manual review | Cost > 2x normal | Critical |

### 2.3 Daily Execution

```bash
# 1. Check budget status
echo "=== Daily Budget Check ==="
ve billing DescribeBudgets --Region "{{env.VOLCENGINE_REGION}}" | jq '.Result.Budgets[] | select(.AlertTriggered==true)'

# 2. Check balance
echo "=== Balance Check ==="
ve billing DescribeBalance --Region "{{env.VOLCENGINE_REGION}}" | jq '.Result'

# 3. Query yesterday's cost (if supported)
echo "=== Yesterday Cost ==="
ve billing DescribeBillDetail --BillingCycle "$(date -v-1d +%Y-%m)" --Region "{{env.VOLCENGINE_REGION}}" | jq '.Result.TotalCost'
```

### 2.4 Daily Report Format

```markdown
# Daily Billing Report — [YYYY-MM-DD]

## Status Summary
| Check | Status | Details |
|-------|--------|---------|
| Budget Alerts | ✅ No alerts | — |
| Balance | ✅ Healthy | ¥XXX,XXX (XX days remaining) |
| Yesterday Cost | ✅ Normal | ¥X,XXX |
| Critical Alerts | ✅ None | — |

## Action Items
- [ ] None (status green)
- [ ] Item 1 (assigned to @owner)
```

---

## 3. Weekly Inspection

### 3.1 Weekly Cost Analysis

```bash
# 1. Query weekly cost summary
echo "=== Weekly Cost Analysis ==="
CURRENT_MONTH=$(date +%Y-%m)
LAST_MONTH=$(date -v-1m +%Y-%m)

# Current month total
CURRENT_COST=$(ve billing DescribeBills --Period "$CURRENT_MONTH" --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.Bills[0].TotalCost')
echo "Current Month Cost: ¥$CURRENT_COST"

# Calculate week-over-week (if data available)
# Estimated daily burn rate
DAYS_ELAPSED=$(date +%d)
ESTIMATED_MONTHLY=$(echo "$CURRENT_COST / $DAYS_ELAPSED * 30" | bc)
echo "Estimated Month End: ¥$ESTIMATED_MONTHLY"

# 2. Top 5 cost drivers
echo "=== Top 5 Cost Drivers ==="
ve billing DescribeBillDetail --BillingCycle "$CURRENT_MONTH" --Region "{{env.VOLCENGINE_REGION}}" | \
  jq -r '.Result.BillDetails[:5] | .[] | "\(.ProductType): \(.ResourceId) — ¥\(.Cost)"'
```

### 3.2 Weekly Checklist

| Check | Command | Threshold | Action |
|-------|---------|-----------|--------|
| Cost vs Budget | Estimated month vs budget | > 80% | Alert |
| MoM Growth | vs Last Month | > 20% | Investigate |
| Top Cost Driver Change | vs Last Week | New resource in Top 5 | Analyze |
| RI Utilization | From DescribeReservedInstances | < 70% | Recommend action |
| Tag Compliance | % resources with cost-center tag | < 90% | Add tags |

### 3.3 Weekly Report Format

```markdown
# Weekly Billing Report — Week [N] ([Start Date] - [End Date])

## Executive Summary
| Metric | This Week | Last Week | Change |
|--------|-----------|-----------|--------|
| Total Cost | ¥XX,XXX | ¥XX,XXX | +X% |
| Daily Avg | ¥X,XXX | ¥X,XXX | +X% |
| Budget Utilization | XX% | XX% | — |
| RI Coverage | XX% | XX% | — |

## Cost Breakdown by Product
| Product | Cost | % of Total | Change |
|---------|------|-----------|--------|
| ECS | ¥XX,XXX | XX% | +X% |
| RDS | ¥X,XXX | XX% | -X% |
| VKE | ¥X,XXX | XX% | +X% |
| Other | ¥XXX | XX% | — |

## Top 5 Cost Drivers
| Rank | ResourceId | Product | Cost | Notes |
|------|-----------|---------|------|-------|
| 1 | i-xxx | ECS | ¥X,XXX | Production API server |
| 2 | i-yyy | ECS | ¥X,XXX | Database primary |
| 3 | d-skp-zzz | Disk | ¥XXX | Attached to i-xxx |

## Optimization Opportunities
| Type | Description | Est.Savings | Priority |
|------|-------------|------------|----------|
| Right-sizing | 3 oversized instances | ¥X,XXX/mo | P1 |
| Idle cleanup | 2 idle disks | ¥XXX/mo | P2 |

## Action Items
| Item | Owner | Due Date | Status |
|------|-------|---------|--------|
| Right-size instances | @devops | +3 days | In Progress |
| Delete idle disks | @devops | +5 days | Pending |
```

---

## 4. Monthly Inspection

### 4.1 Monthly Deep Dive

```bash
# 1. Full bill analysis
echo "=== Monthly Bill Analysis ==="
CURRENT_MONTH=$(date +%Y-%m)

# Total cost by product
ve billing DescribeBills --Period "$CURRENT_MONTH" --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '.Result.Bills[0]'

# 2. Cross-period comparison
for PERIOD in $(date -v-6m +%Y-%m) $(date -v-5m +%Y-%m) $(date -v-4m +%Y-%m) $(date -v-3m +%Y-%m) \
               $(date -v-2m +%Y-%m) $(date -v-1m +%Y-%m) $(date +%Y-%m); do
  echo -n "$PERIOD: "
  ve billing DescribeBills --Period "$PERIOD" --Region "{{env.VOLCENGINE_REGION}}" | \
    jq -r '.Result.Bills[0].TotalCost' 2>/dev/null || echo "N/A"
done

# 3. RI utilization report
echo "=== RI Utilization Report ==="
ve billing DescribeReservedInstances --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '.Result.ReservedInstances[] | {Name, InstanceType, UsedUnits, TotalUnits, Utilization}'
```

### 4.2 Monthly Checklist

| Check | Focus | Deliverable |
|-------|-------|-------------|
| Bill Reconciliation | Verify invoice vs actual spend | Invoice report |
| Budget Performance | Actual vs budget by category | Variance report |
| Cost Trend | 6-month trend analysis | Trend chart |
| RI Renewal Planning | Upcoming RI expirations | Renewal recommendation |
| Optimization ROI | Savings from actions taken | ROI report |
| Tag Compliance Review | Tag coverage improvement | Compliance report |
| Next Month Forecast | Predict next month's cost | Forecast report |

### 4.3 Monthly Report Format

```markdown
# Monthly Billing Report — [Month YYYY]

## Executive Summary
| Metric | This Month | Last Month | vs Budget | vs Last Year |
|--------|-----------|-----------|-----------|-------------|
| Total Cost | ¥XXX,XXX | ¥XX,XXX | +X% | +X% |
| Daily Avg | ¥XX,XXX | ¥X,XXX | — | — |
| Budget Used | XX% | XX% | — | — |
| Est. Month End | ¥XXX,XXX | — | +X% | — |

## Cost Trend (6 Months)
| Month | Cost | MoM Change | vs Budget |
|-------|------|-----------|-----------|
| Jan | ¥XXX,XXX | — | -X% |
| Feb | ¥XXX,XXX | +X% | -X% |
| ... | ... | ... | ... |

## Cost Breakdown

### By Product Type
| Product | Cost | % of Total | MoM Change |
|---------|------|-----------|-----------|
| ECS | ¥XX,XXX | XX% | +X% |
| RDS | ¥XX,XXX | XX% | +X% |
| VKE | ¥XX,XXX | XX% | +X% |
| Storage | ¥XX,XXX | XX% | +X% |
| Network | ¥XX,XXX | XX% | +X% |
| Other | ¥XX,XXX | XX% | — |

### By Environment
| Environment | Cost | % of Total | Tag Coverage |
|-------------|------|-----------|--------------|
| Production | ¥XX,XXX | XX% | 95% |
| Staging | ¥X,XXX | XX% | 88% |
| Dev | ¥XXX | XX% | 70% |

### By Cost Center
| Cost Center | Cost | % of Total |
|-------------|------|-----------|
| Team-Backend | ¥XX,XXX | XX% |
| Team-Frontend | ¥XX,XXX | XX% |
| Team-ML | ¥XX,XXX | XX% |

## Reserved Instance Analysis
| RI Type | Total Units | Used | Coverage | Cost Savings |
|---------|------------|------|----------|-------------|
| ecs.g6.large | 10 | 8 | 80% | ¥X,XXX |
| ecs.g6.xlarge | 5 | 5 | 100% | ¥X,XXX |

**RI Renewal Recommendations:**
- RI type X expires on [date] — Recommend [renewal/cancel]
- RI type Y utilization < 50% — Recommend reducing

## Optimization Achieved This Month
| Action | Est.Savings | Actual Savings |
|--------|------------|---------------|
| Right-size 5 instances | ¥X,XXX/mo | ¥X,XXX |
| Delete 3 idle resources | ¥XXX/mo | ¥XXX |
| Total | ¥X,XXX/mo | ¥X,XXX |

## Optimization Opportunities Next Month
| Priority | Opportunity | Est.Savings | Complexity |
|----------|-------------|------------|------------|
| P1 | Right-size 10 oversized instances | ¥X,XXX/mo | Medium |
| P2 | Purchase 1-year RI for VKE | ¥X,XXX/mo | Low |
| P3 | Archive cold storage data | ¥XXX/mo | High |

## Tag Compliance
| Metric | Current | Target | Gap |
|--------|---------|--------|-----|
| Tagged Resources | XX% | 95% | -X% |
| Cost Attribution Coverage | XX% | 90% | -X% |

## Budget Recommendations
| Budget Name | Current Limit | Current Spend | Recommendation |
|-------------|--------------|---------------|----------------|
| monthly-production | ¥100,000 | ¥80,000 (80%) | Increase to ¥110,000 |
| monthly-dev | ¥20,000 | ¥5,000 (25%) | Consider reduction |

## Risks & Action Items
| Risk | Impact | Mitigation | Owner |
|------|--------|------------|-------|
| Cost spike in VKE | ¥X,XXX/mo increase | Right-sizing | @team |
| RI expiration without renewal | +¥X,XXX/mo | Renewal review | @finops |

## Sign-Off
| Role | Name | Date | Signature |
|------|------|------|-----------|
| FinOps Engineer | | | |
| Finance/Manager | | | |
| DevOps Lead | | | |
```

---

## 5. Alert Response Playbook

### 5.1 Budget Alert (80%, 90%, 100%)

```
Trigger:   Budget threshold exceeded
Severity: 80%=Warning, 90%=High, 100%=Critical

Response Steps:
1. Acknowledge alert within 1 hour
2. Query DescribeBillDetail for current period
3. Identify cost driver (product type, resource, tag)
4. Determine if spend is expected or anomalous
5. If anomalous:
   - Stop non-critical resources if possible
   - Notify team leads
   - Escalate if > 100% budget
6. If expected:
   - Document reason
   - Update forecast
   - Recommend budget adjustment if recurring
```

### 5.2 Balance Low Alert

```
Trigger:   Balance < estimated 30-day cost
Severity:  Critical

Response Steps:
1. Check current balance: ve billing DescribeBalance
2. Review recent cost trend
3. Calculate days remaining at current burn rate
4. Initiate recharge process (Finance team)
5. If recharge will take > 3 days:
   - Identify non-critical resources for potential shutdown
   - Prepare cost impact assessment
```

### 5.3 Cost Spike Alert

```
Trigger:   Day-over-day cost > 2x average
Severity:  Critical

Response Steps:
1. Confirm spike (not false positive)
2. Query DescribeBillDetail for current day (if available)
3. Identify new resources or configuration changes
4. Check recent deployments with DevOps
5. If resource creation:
   - Verify if intentional
   - Review resource tagging
6. If configuration change:
   - Identify change owner
   - Assess cost impact
7. Implement mitigation if needed
```

### 5.4 RI Underutilization Alert

```
Trigger:   RI utilization < 70% for 7 days
Severity:  Warning

Response Steps:
1. Check RI utilization history
2. Identify underutilized RI type and count
3. Review workload patterns
4. Options:
   - Redeploy workloads to utilize RIs
   - Reduce RI purchases for next term
   - Accept reduced coverage (if seasonal)
5. Document decision and rationale
```

---

## 6. Report Templates

### 6.1 Ad-Hoc Cost Query Report

```markdown
# Cost Query Report — [Title]

**Query Parameters:**
- Period: [YYYY-MM]
- Product: [Product Type]
- Tags: [Tag Filters]

**Results:**
| Metric | Value |
|--------|-------|
| Total Cost | ¥XXX,XXX |
| Record Count | X,XXX |

**Top 10 Cost Drivers:**
[Table]

**Recommendations:**
[Based on analysis]
```

### 6.2 Optimization Proposal

```markdown
# Cost Optimization Proposal

**Date:** [Date]
**Prepared By:** [FinOps Engineer]

## Current State
[Cost overview]

## Proposed Changes
| # | Change | Est.Savings | Effort | Risk |
|---|--------|------------|--------|------|
| 1 | [Change 1] | ¥X,XXX/mo | Low | Low |
| 2 | [Change 2] | ¥X,XXX/mo | Medium | Medium |

## ROI Analysis
[ROI calculation]

## Approval
| Role | Name | Approved | Date |
|------|------|----------|------|
| Finance | | ☐ | |
| Engineering | | ☐ | |
```

---

## Inspection Calendar

| Day | Activity | Owner |
|-----|----------|-------|
| 1-2 | Monthly close reconciliation | FinOps |
| 3 | Monthly report generation | FinOps |
| 5 | Monthly review meeting | All |
| Mon-Fri | Daily inspection | FinOps |
| Friday | Weekly report | FinOps |
| 1st Monday | Monthly inspection kickoff | FinOps |
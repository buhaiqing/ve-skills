# Billing Knowledge Base

> **Purpose:** Fault pattern library and troubleshooting scenarios for Volcengine Billing operations.
> **Version:** 1.3.0>
> **Last Updated:** 2026-06-02

---

## Table of Contents

1. [Cost Anomaly Patterns](#1-cost-anomaly-patterns)
2. [Budget Management Patterns](#2-budget-management-patterns)
3. [Balance & Payment Patterns](#3-balance--payment-patterns)
4. [Reserved Instance Patterns](#4-reserved-instance-patterns)
5. [Invoice Patterns](#5-invoice-patterns)
6. [Tag & Attribution Patterns](#6-tag--attribution-patterns)
7. [Cross-Service Patterns](#7-cross-service-patterns)

---

## 1. Cost Anomaly Patterns

### Pattern: Unexpected Cost Spike

**Description:** Monthly cost increased > 20% without apparent reason.

**Root Causes:**
1. New resource creation without budget tracking
2. Data transfer costs higher than expected (cross-AZ, egress)
3. Reserved Instance expiration (converted to On-Demand)
4. Storage auto-scaling (snapshots, backups)
5. Unscaled development environments

**Detection:**
```bash
# Check for abnormal cost increase
ve billing DescribeBills --Period "$(date +%Y-%m)" --Region "{{env.VOLCENGINE_REGION}}"
ve billing DescribeBills --Period "$(date -v-1m +%Y-%m)" --Region "{{env.VOLCENGINE_REGION}}"
# Compare TotalCost between periods
```

**Resolution Steps:**
1. Query DescribeBillDetail filtered by resource
2. Compare month-over-month with previous periods
3. Identify cost driver resource type via `ProductType` grouping
4. Cross-reference with product APIs to identify new resources
5. Recommend right-sizing, cleanup, or budget adjustment

**Example Investigation:**
```bash
# Step 1: Get product breakdown
ve billing DescribeBillDetail --BillingCycle "$(date +%Y-%m)" --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '[.Result.BillDetails | group_by(.ProductType)[] | {product: .[0].ProductType, cost: (map(.Cost) | add)}]'

# Step 2: Identify top cost increases
ve billing DescribeBillDetail --BillingCycle "$(date +%Y-%m)" --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '.Result.BillDetails | sort_by(.Cost) | reverse | .[:10]'
```

---

### Pattern: Gradual Cost Creep

**Description:** Cost increases 5-10% each month for consecutive months.

**Root Causes:**
1. Organic business growth (legitimate)
2. Unoptimized resource configurations
3. Storage accumulation without lifecycle policies
4. Log volume growth
5. Uncontrolled resource proliferation

**Resolution Steps:**
1. Analyze 6-month trend
2. Identify linear vs exponential growth pattern
3. Check for storage bloat (snapshots, old volumes)
4. Review log retention settings
5. Implement resource quotas

---

### Pattern: Unusual Cost by Product Type

**Description:** Specific product type shows abnormal cost.

**Investigation by Product:**

| Product | Check | Command |
|---------|-------|---------|
| ECS | New instances, unscaled stopped instances | `ve ecs DescribeInstances` |
| RDS | Storage auto-scaling, increased connections | `ve rds DescribeDBInstances` |
| VKE | Node count increase, new clusters | `ve vke DescribeClusters` |
| TOS | Accelerate enabled, versioning storage | `ve tos GetBucketStat` |
| SLS | Index bloat, long retention | `ve sls GetProjectStatistics` |

---

## 2. Budget Management Patterns

### Pattern: Budget Not Triggering Alerts

**Symptoms:** Spending exceeds budget but no alert received.

**Root Causes:**
1. Budget not configured with alert thresholds
2. Alert thresholds set too high
3. Budget scope doesn't match actual spending
4. Notification channel misconfigured

**Resolution Steps:**
1. Check budget configuration:
   ```bash
   ve billing DescribeBudgets --Region "{{env.VOLCENGINE_REGION}}" | \
     jq '.Result.Budgets[] | {name: .BudgetName, thresholds: .AlertThresholds, triggered: .AlertTriggered}'
   ```
2. Verify alert thresholds are set (should be [80, 90, 100] or similar)
3. Check if budget scope (tags) covers actual resources
4. Verify notification channel configuration

---

### Pattern: Budget Frequently Exceeded

**Symptoms:** Consistent budget overruns despite cost control efforts.

**Root Causes:**
1. Budget limit set too low for actual needs
2. Underestimating seasonal variations
3. Incomplete budget scope (missing products)
4. Development environment costs not isolated

**Resolution Steps:**
1. Analyze 3-month actual spend average
2. Add 15-20% buffer for buffer (seasonal, growth)
3. Create separate budgets per environment
4. Set up granular budgets by cost center

**Recommended Budget Structure:**
```
Organization Budget (soft limit)
├── Production Budget (strict)
├── Staging Budget (moderate)
├── Dev/Test Budget (strict)
└── Per-Team Budgets (if needed)
```

---

### Pattern: Multiple Budgets Confusion

**Symptoms:** Conflicting budget alerts from overlapping budgets.

**Root Causes:**
1. Budgets covering same resource scope
2. No clear hierarchy or priority
3. Tag conflicts between budgets

**Resolution Steps:**
1. List all budgets:
   ```bash
   ve billing DescribeBudgets --Region "{{env.VOLCENGINE_REGION}}"
   ```
2. Identify overlapping tag filters
3. Consolidate or delete redundant budgets
4. Establish budget naming convention: `{environment}-{scope}`

---

## 3. Balance & Payment Patterns

### Pattern: Insufficient Balance

**Symptoms:** Account balance approaching zero, service interruptions.

**Root Causes:**
1. Delayed payment or recharge
2. Unexpected cost spike draining balance
3. Budget not monitored
4. Credit limit exhausted

**Resolution Steps:**
1. Check current balance:
   ```bash
   ve billing DescribeBalance --Region "{{env.VOLCENGINE_REGION}}"
   ```
2. Review recent bills for cost trends:
   ```bash
   ve billing DescribeBills --Period "$(date +%Y-%m)" --Region "{{env.VOLCENGINE_REGION}}"
   ```
3. Calculate days remaining at current burn rate
4. Initiate recharge process (requires manual intervention via console or payment portal)
5. Create budget with alert if not exists
6. Consider PrePaid commitment to lock in balance

**Emergency Actions:**
1. Identify non-critical resources that can be stopped
2. Alert team leads for cost control measures
3. Escalate to finance for emergency recharge

---

### Pattern: Unexpected Prepaid Deduction

**Symptoms:** Balance decreases faster than expected from billing.

**Root Causes:**
1. Monthly invoice generation and payment
2. RI renewal charges
3. Subscription auto-renewal
4. Overdue invoice late fees

**Resolution Steps:**
1. Check invoice history:
   ```bash
   ve billing DescribeInvoices --Region "{{env.VOLCENGINE_REGION}}"
   ```
2. Verify scheduled payments or commitments
3. Review RI renewal schedule:
   ```bash
   ve billing DescribeReservedInstances --Region "{{env.VOLCENGINE_REGION}}" | \
     jq '.Result.ReservedInstances[] | {type: .InstanceType, expire: .ExpireTime}'
   ```

---

## 4. Reserved Instance Patterns

### Pattern: RI Expiring Soon

**Symptoms:** RI approaching expiration, risk of on-demand charges.

**Detection:**
```bash
ve billing DescribeReservedInstances --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '.Result.ReservedInstances[] | select(.ExpireTime | fromdate < (now + 2592000))'
```

**Resolution:**
1. Review workload patterns for next period
2. Calculate ROI of renewal vs converting to on-demand
3. If continuing:
   - Renew RI before expiration
   - Consider 1-year or 3-year for better discounts
4. If discontinuing:
   - Plan workload migration
   - Schedule RI expiration (don't renew)

**Decision Matrix:**
| Utilization | Workload Status | Recommendation |
|-------------|----------------|----------------|
| > 80% | Stable/Growing | Renew (1-year or 3-year) |
| 50-80% | Stable | Renew (1-year), monitor |
| 50-80% | Declining | Don't renew |
| < 50% | Any | Don't renew, investigate underutilization |

---

### Pattern: RI Underutilization

**Symptoms:** RI utilization consistently below 70%.

**Root Causes:**
1. Over-purchased RI capacity
2. Workload shifted to different instance types
3. Application architecture changes
4. Seasonal workload reduction

**Resolution Steps:**
1. Analyze utilization by instance type:
   ```bash
   ve billing DescribeReservedInstances --Region "{{env.VOLCENGINE_REGION}}" | \
     jq '.Result.ReservedInstances[] | {type: .InstanceType, util: .Utilization}'
   ```
2. Identify underutilized RI types
3. Cross-reference with actual workload
4. Options:
   - Redeploy workloads to utilize RIs
   - Reduce RI purchases for next term
   - Accept reduced coverage (document rationale)

---

### Pattern: No RI Coverage for Production

**Symptoms:** Production workloads running on-demand despite stable base load.

**Resolution:**
1. Analyze workload stability (30-day minimum data)
2. Calculate potential savings:
   ```
   On-Demand Cost: X CNY/month
   RI Cost (1-year): X * 0.6 = Y CNY/month
   Savings: X - Y = Z CNY/month
   ```
3. Recommend purchasing RIs for stable base capacity
4. Use on-demand for burst capacity

---

## 5. Invoice Patterns

### Pattern: Invoice Status Mismatch

**Symptoms:** Invoice shows different amount than actual spend.

**Resolution Steps:**
1. Verify invoice details:
   ```bash
   ve billing DescribeInvoices --InvoiceType VAT_NORMAL --Region "{{env.VOLCENGINE_REGION}}"
   ```
2. Compare with billing:
   ```bash
   ve billing DescribeBills --Period "2026-05" --Region "{{env.VOLCENGINE_REGION}}"
   ```
3. Common discrepancies:
   - Invoice includes future charges (prepaid)
   - Credit adjustments not reflected
   - Currency conversion (if applicable)

---

### Pattern: Invoice Application Rejected

**Symptoms:** Unable to create or apply for invoice.

**Root Causes:**
1. Invoice amount below minimum threshold
2. Missing company tax information
3. Previous invoice not yet processed
4. Account verification incomplete

**Resolution Steps:**
1. Check invoice requirements in documentation
2. Verify company tax registration is complete
3. Wait for previous invoice processing to complete
4. Contact support for specific rejection reason

---

## 6. Tag & Attribution Patterns

### Pattern: Untagged Resources Causing Cost Gaps

**Symptoms:** Sum of tagged resources doesn't equal total bill.

**Root Causes:**
1. New resources without mandatory tags
2. Tag propagation delay
3. Resources created via APIs without tags
4. Legacy untagged resources

**Resolution Steps:**
1. Identify untagged resources via product APIs:
   ```bash
   # ECS example
   ve ecs DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" | \
     jq '.Result.Instances[] | select(.Tags == null or .Tags | length == 0) | .InstanceId'
   ```
2. Audit tag compliance percentage
3. Implement mandatory tagging via policy
4. Backfill tags on existing resources

**Compliance Enforcement:**
```
New Resources: Require tags before creation → reject if missing
Existing Resources: 30-day remediation window
Monthly Audit: Tag compliance report
```

---

### Pattern: Cost Center Allocation Disputes

**Symptoms:** Teams disputing allocated costs.

**Root Causes:**
1. Shared resources not properly allocated
2. Tag values inconsistent (case, spelling)
3. Project transferred without tag update
4. Resource ownership changes

**Resolution:**
1. Audit tag usage consistency:
   ```bash
   ve billing DescribeResourceTags --Region "{{env.VOLCENGINE_REGION}}"
   ```
2. Standardize tag value conventions
3. Implement tag validation policies
4. Create shared cost allocation rules

**Tag Value Standards:**
| Tag | Standard | Examples |
|-----|----------|-----------|
| cost-center | kebab-case | `engineering-infra`, `ml-platform` |
| environment | lowercase | `production`, `staging`, `dev` |
| project | kebab-case | `payment-service`, `user-api` |
| owner | team-email | `team-backend@company.com` |

---

## 7. Cross-Service Patterns

### Pattern: Cross-Service Cost Spike

**Symptoms:** Overall cost spike involves multiple products.

**Investigation Workflow:**

1. **Identify spike scope:**
   ```bash
   ve billing DescribeBillDetail --BillingCycle "$(date +%Y-%m)" --Region "{{env.VOLCENGINE_REGION}}" | \
     jq '[.Result.BillDetails | group_by(.ProductType)[] | {product: .[0].ProductType, total: (map(.Cost) | add)}]'
   ```

2. **Cross-reference with product APIs:**

   | Product | Check | Command |
   |---------|-------|---------|
   | ECS | New instances created | `ve ecs DescribeInstances --CreationDateStart` |
   | VKE | Cluster/node changes | `ve vke DescribeClusters` |
   | RDS | Storage increases | `ve rds DescribeDBInstances` |
   | TOS | New buckets/objects | `ve tos ListBuckets` |

3. **Correlate with deployment logs:**
   - Check CI/CD pipelines for recent deployments
   - Review Terraform state changes
   - Check manual operations logs

4. **Root cause report:**
   ```markdown
   | Finding | Evidence | Owner | Action |
   |---------|----------|-------|--------|
   | VKE cluster scaled | 5 new nodes created 05-15 | @team-alpha | Tag + monitor |
   | Storage auto-scale | TOS bucket grew 500GB | @team-beta | Add lifecycle |
   ```

---

### Pattern: Hidden Cross-Service Costs

**Symptoms:** Individual product costs seem normal, but total bill is high.

**Common Hidden Costs:**

| Hidden Cost | Source | Investigation |
|-------------|--------|---------------|
| Data transfer | Cross-AZ, egress | Check VKE/ECS network metrics |
| Backup storage | RDS, ECS snapshots | Check DescribeSnapshots |
| Log storage | SLS, CloudTrail | Check retention policies |
| CDN traffic | TOS + CDN | Check CDN bandwidth |
| API Gateway | ALB/CLB LCU | Check LB metrics |

**Investigation:**
```bash
# Check network costs (data transfer)
ve billing DescribeBillDetail --BillingCycle "$(date +%Y-%m)" --ProductType ecs --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '.Result.BillDetails | map(select(.BillItemName | contains("Traffic") or contains("Transfer")))'

# Check snapshot costs
ve ecs DescribeSnapshots --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '[.Result.Snapshots[] | {id: .SnapshotId, size: .Size, created: .CreationTime}]'
```

---

### Pattern: Environment Cost Leakage

**Symptoms:** Dev/Test environments consuming budget meant for production.

**Root Causes:**
1. No environment isolation
2. Shared budgets across environments
3. Dev resources using production configurations
4. Auto-scaling in dev environment

**Resolution:**
1. Implement environment isolation budgets
2. Set resource quotas per environment
3. Enable scheduled scaling (auto-stop dev at night)
4. Tag-based cost allocation

**Best Practice:**
```
production-budget  → { tag: environment=production }
staging-budget     → { tag: environment=staging }
dev-budget         → { tag: environment=dev }
```

---

## Quick Reference: Diagnostic Command Summary

| Issue | First Command | Follow-up |
|-------|--------------|-----------|
| High cost | `ve billing DescribeBillDetail` | `jq group_by(ProductType)` |
| Cost spike | Compare `DescribeBills` periods | Drill into `DescribeBillDetail` |
| Budget alert | `ve billing DescribeBudgets` | Check thresholds |
| Low balance | `ve billing DescribeBalance` | Calculate burn rate |
| RI utilization | `ve billing DescribeReservedInstances` | Calculate utilization % |
| Tag issues | `ve billing DescribeResourceTags` | Cross-check product APIs |
| Invoice problem | `ve billing DescribeInvoices` | Compare with bills |
| Cross-service | `ve billing DescribeBillDetail` + product APIs | Correlate timestamps |

---

## Escalation Matrix

| Severity | Condition | Response Time | Escalation |
|----------|-----------|---------------|------------|
| **P1 - Critical** | Balance < 3 days, Budget exceeded > 20% | 1 hour | FinOps Manager |
| **P2 - High** | Budget exceeded 10-20%, Anomaly detected | 4 hours | FinOps Engineer |
| **P3 - Medium** | Budget warning (80%), RI utilization low | 24 hours | Team Lead |
| **P4 - Low** | Tag compliance, optimization opportunities | 1 week | Self-service |
---
name: ve-billing-ops
description: >-
  Use when the user needs to query, analyze, or manage Volcengine (火山引擎)
  billing, cost, and account — cost explorer, budget management, payment methods,
  invoices, resource tags, reserved instances, and cost optimization. User mentions
  billing, cost, 费用, 账单, invoice, budget, payment, 充值, cost analysis,
  spending, or describes cost-related scenarios even without naming the product
  directly. Not for resource provisioning or IAM permissions.
license: MIT
compatibility: >-
  Official Volcengine CLI (`ve`, Go binary, no runtime), Go 1.21+ runtime
  (JIT SDK fallback; scripts compatible with Go 1.14+ syntax), valid API credentials,
  network access to Volcengine endpoints (open.volcengineapi.com).
metadata:
  author: volcengine
  version: "1.4.0"
  last_updated: "2026-06-02"
  runtime: Harness AI Agent, Claude Code, Cursor, or compatible Agent runtimes
  go_script_syntax_minimum: "1.14"
  go_jit_runtime_version: "1.21+"
  api_profile: "Billing OpenAPI — https://www.volcengine.com/docs/6387"
  cli_applicability: dual-path
  cli_support_evidence: >-
    Confirmed via `ve billing --help` — Billing is supported by the ve CLI.
    Service ID: billing.
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

> This skill follows the [Agent Skill OpenSpec](https://agentskills.io/specification).
>
> **UX Compliance:** This skill follows the [User Experience Specification](../ve-skill-generator/references/user-experience-spec.md). All operations include onboarding guidance, minimal prompts, smart defaults, clear feedback, and user-friendly error handling.

# Volcengine Billing Operations Skill

## Overview

Volcengine Billing (火山引擎费用中心) provides cost management, billing, and account operations. This skill is an **operational runbook** for agents: cost query, budget management, payment operations, invoice management, and cost optimization. **Do not use the web console as the primary agent execution path.**

### CLI applicability

- **`cli_applicability: dual-path`:** Official `ve` CLI supports Billing. Document **both** SDK and CLI steps for every operation.

## Five Core Standards (Quality Gates)

Every generated skill MUST satisfy these five standards. Use them as a design checklist during population:

| # | Standard | How This Skill Fulfills It | Concrete Validation Criteria |
|---|----------|---------------------------|------------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use conditions with precise triggers and delegation rules | ≥ 3 SHOULD entries with specific triggers; ≥ 3 SHOULD NOT entries with named delegation targets |
| 2 | **Structured I/O** | Placeholder conventions (`{{env.*}}`, `{{user.*}}`, `{{output.*}}`) with type and source documented | Zero bare variable names; every input uses a typed placeholder; every output maps to a JSON path |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover, with numbered imperative steps | ≥ 1 operation with all 4 phases present; all steps numbered and imperative (not descriptive) |
| 4 | **Complete Failure Strategies** | Error taxonomy table with ≥ 10 product-specific codes; HALT vs retry per error type | Error table has ≥ 10 rows; each row has: code, max retries, backoff, agent action, UX template |
| 5 | **Absolute Single Responsibility** | One product, one primary resource model; cross-product delegation to other skills | SKILL.md covers exactly 1 product; cross-product ops delegate (not duplicate); naming follows `ve-billing-ops` |

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- User mentions "billing", "cost", "费用", "账单", "invoice", "budget", "payment", "充值"
- Task involves **cost query**: monthly spend, by-resource cost, cost trends
- Task involves **budget management**: create/modify/delete budgets, budget alerts
- Task involves **payment**: payment methods, automatic payment, overdue payment
- Task involves **invoice**: invoice application, invoice history, tax certificates
- Task involves **reserved instances**: purchase, utilization reporting, renewal
- Task involves **resource tags**: cost allocation tags, tag-based cost grouping
- Task involves **account**: balance inquiry, credit limit, account statements
- Task involves **cost optimization**: identifying high-cost resources, right-sizing recommendations

### SHOULD NOT Use This Skill When

- Task is about **creating/changing resources** → delegate to product-specific skill
- Task is about **IAM permissions** → delegate to: `ve-iam-ops`
- Task is about **payment gateway technical integration** → developer documentation

### Delegation Rules

- If resource cost analysis requires product-specific details → delegate to: `ve-ecs-ops`, `ve-rds-ops`, etc.
- Multi-product cost requests: aggregate via ve-billing-ops, then drill down with product skills.
- Cost optimization execution (right-sizing, scale-in) → delegate to respective product skills.

## Variable Convention (Agent-Readable)

Structured placeholders reduce injection ambiguity and unsafe prompts:

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_REGION}}` | From runtime environment | Use documented default only if skill explicitly allows |
| `{{user.start_date}}` | User-supplied start date | Format YYYY-MM; ask if missing |
| `{{user.end_date}}` | User-supplied end date | Format YYYY-MM; ask if missing |
| `{{user.budget_name}}` | User-supplied budget name | Ask once; reuse |
| `{{user.budget_amount}}` | User-supplied budget limit | Decimal, e.g., 10000.00; ask if missing |
| `{{output.total_cost}}` | From billing summary response | Parse `$.Result.Bills[0].TotalCost` |
| `{{output.budget_id}}` | From CreateBudget response | Parse `$.Result.BudgetId` |
| `{{output.balance}}` | From balance response | Parse `$.Result.Balance` |

> **`{{env.*}}` MUST NOT** be collected from the user. **`{{user.*}}` MUST be collected interactively when missing.**

> **Security Warning (Credential Masking — MANDATORY):** **NEVER** log, print, or expose `VOLCENGINE_SECRET_KEY`, `secret_key`, or any credential field value in console output, debug messages, error messages, or logs.
>
> **Masking rules:**
> | Execution Path | Safe Pattern | Unsafe Pattern |
> |----------------|-------------|----------------|
> | Console output | `VOLCENGINE_SECRET_KEY=<masked>` | `VOLCENGINE_SECRET_KEY=AKLT...` |
> | Error messages | `Error: API call failed (credential omitted)` | `Error: InvalidSecretKey... actual secret...` |
> | Verification | `test -n "$VOLCENGINE_SECRET_KEY" && echo "✅ SecretKey is set"` | `echo "SecretKey=$VOLCENGINE_SECRET_KEY"` |

## API and Response Conventions

### Response Field Map

| Operation | JSON Path | Type | Description |
|-----------|-----------|------|-------------|
| DescribeBills | `$.Result.Bills[0].TotalCost` | number | Total cost for period (CNY) |
| DescribeBills | `$.Result.Bills[0].ProductDetail[].ProductType` | string | Product type identifier |
| DescribeBills | `$.Result.Bills[0].ProductDetail[].Cost` | number | Cost by product |
| DescribeBillDetail | `$.Result.BillDetails[].ResourceId` | string | Resource identifier |
| DescribeBillDetail | `$.Result.BillDetails[].Cost` | number | Cost per resource |
| DescribeBalance | `$.Result.Balance` | number | Account balance (CNY) |
| DescribeBudgets | `$.Result.Budgets[].BudgetId` | string | Budget identifier |
| DescribeBudgets | `$.Result.Budgets[].ActualSpent` | number | Actual spend against budget |
| CreateBudget | `$.Result.BudgetId` | string | Newly created budget ID |
| DescribeReservedInstances | `$.Result.ReservedInstances[].InstanceType` | string | RI instance type |

### Expected State Transitions

| Operation | Target State | Poll Interval | Max Wait | Validation |
|-----------|--------------|---------------|----------|------------|
| CreateBudget | Budget appears in DescribeBudgets | 2s | 30s | Verify `{{output.budget_id}}` exists |
| DeleteBudget | Budget removed from list | 2s | 30s | Verify BudgetId no longer present |
| UpdateBudget | Changes reflected in DescribeBudgets | 2s | 30s | Verify updated values |

## Quick Start

### What This Skill Does
Query, analyze, and manage Volcengine billing, cost, and account using the `ve` CLI or JIT Go SDK.

### Prerequisites
- [ ] `ve` CLI installed (or Go runtime for JIT fallback)
- [ ] Credentials configured: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`
- [ ] Region set: `VOLCENGINE_REGION`

### Verify Setup
```bash
ve billing DescribeBills --Period "$(date +%Y-%m)" --Region "{{env.VOLCENGINE_REGION}}"
```

### Your First Command
```bash
# Query current month billing summary
ve billing DescribeBills --Period "$(date +%Y-%m)" --Region "{{env.VOLCENGINE_REGION}}" | jq '.Result.Bills[0]'
```

### Next Steps
- [Core Concepts](references/core-concepts.md) — Understand billing model
- [Cost Optimization](references/cost-optimization.md) — Optimize spending
- [Operations Framework](references/operations-framework.md) — Set up periodic reviews

## Capabilities at a Glance

| Operation | Description | Complexity | Risk Level | | safety_class | blast_radius |
|-----------|-------------|------------|------------||---|---|
| **Core Queries** | | | | | state-changing | single |
| DescribeBills | Query monthly bills | Low | None | | read-only | single |
| DescribeBillDetail | Get bill line items | Low | None | | read-only | single |
| DescribeBalance | Check account balance | Low | None | | read-only | single |
| **Budget Management** | | | | | state-changing | single |
| CreateBudget | Set spending budget | Medium | Low | | state-changing | single |
| DescribeBudgets | List budgets | Low | None | | read-only | single |
| ModifyBudget | Update budget threshold | Medium | Low | | state-changing | single |
| DeleteBudget | Remove budget | Low | Medium | | destructive | single |
| **Reserved Instances** | | | | | state-changing | single |
| DescribeReservedInstances | View RI utilization | Low | None | | read-only | single |
| **Analyzers** | | | | | state-changing | single |
| AnalyzeCostTrend | Month-over-month trend analysis | Medium | None | | state-changing | single |
| FindCostDrivers | Top N cost driver identification | Medium | None | | read-only | single |
| AnalyzeRIUtilization | RI coverage and utilization analysis | Medium | None | | state-changing | single |
| DetectCostAnomaly | Identify abnormal cost changes | Medium | None | | state-changing | single |
| PredictSpending | Month-end cost forecast | Medium | None | | state-changing | single |
| **Optimizers** | | | | | state-changing | single |
| GenerateRightSizingRec | Right-sizing recommendations | High | Medium | | state-changing | single |
| IdentifyIdleResources | Idle resource detection | Medium | None | | state-changing | single |
| **Reports** | | | | | state-changing | single |
| GenerateDailyReport | Daily billing summary | Low | None | | state-changing | single |
| GenerateWeeklyReport | Weekly cost analysis | Medium | None | | state-changing | single |
| GenerateMonthlyReport | Monthly deep-dive analysis | High | None | | state-changing | single |
| GenerateOptimizationReport | Cost optimization proposal | High | None | | state-changing | single |

## Execution Flows (Agent-Readable)

Every operation: **Pre-flight → Execute → Validate → Recover**.

### Operation: DescribeBills — Query Monthly Cost

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| CLI / deps | `ve version` | Exit code 0 | Document CLI install |
| Credentials | Verify env vars `VOLCENGINE_ACCESS_KEY` and `VOLCENGINE_SECRET_KEY` are set | Non-empty keys | HALT; user configures env |
| Region | Verify `{{env.VOLCENGINE_REGION}}` is valid | Region supported | Suggest valid region |

#### Execution — CLI (Primary Path)

```bash
ve billing DescribeBills \
  --Period "{{user.start_date}}" \
  --Region "{{env.VOLCENGINE_REGION}}"
```

#### Execution — JIT Go SDK (Fallback Path)

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/volcengine/volc-sdk-golang/service/billing"
)

func main() {
    instance := billing.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))
    
    params := make(map[string]interface{})
    params["Period"] = os.Getenv("VOLCENGINE_PERIOD") // Format: YYYY-MM
    params["Region"] = os.Getenv("VOLCENGINE_REGION")
    
    resp, err := instance.Client.Request("billing", "DescribeBills", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "API call failed: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println(string(resp))
}
```

#### Validation

1. Parse `$.Result.Bills[0].TotalCost` and confirm non-empty.
2. Report: total cost, by-product breakdown, top 5 cost drivers.

#### Failure Recovery

| Error Pattern | Max Retries | Backoff | Agent Action | UX Feedback |
|---------------|-------------|---------|--------------|-------------|
| `InvalidParameter` / 400 | 1 | — | Fix period format (YYYY-MM); retry | `[ERROR] InvalidParameter: Period format should be YYYY-MM` |
| `Forbidden.RAM` / 403 | 0 | — | HALT; add IAM billing permissions | `[ERROR] Forbidden.RAM: Insufficient permissions for billing operations` |
| `InternalError` / 500 | 3 | 2s, 4s, 8s | Retry; then HALT with RequestId | `[ERROR] InternalError: Server error. Retrying... (Attempt 3/3)` |
| Throttling / 429 | 3 | exponential | Back off; respect Retry-After | `⚠️ Rate limit reached. Retrying in {backoff}s...` |

---

### Operation: CreateBudget — Set Budget Alert

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Budget name unique | Query DescribeBudgets, check existing names | Name not in use | Ask user for different name |
| Amount valid | Validate `{{user.budget_amount}}` > 0 | Positive number | Prompt for valid amount |
| Thresholds valid | Validate alert thresholds in [1, 100] | Valid range | Adjust to 80, 90, 100 |

#### Execution — CLI

```bash
ve billing CreateBudget \
  --BudgetAmount {{user.budget_amount}} \
  --BudgetType MONTHLY \
  --AlertThresholds '[80, 90, 100]' \
  --BudgetName "{{user.budget_name}}" \
  --Region "{{env.VOLCENGINE_REGION}}"
```

#### Validation

1. Read `{{output.budget_id}}` from `$.Result.BudgetId`.
2. Poll DescribeBudgets to verify budget appears.
3. Report budget ID, name, and amount.

#### Failure Recovery

| Error Pattern | Max Retries | Backoff | Agent Action | UX Feedback |
|---------------|-------------|---------|--------------|-------------|
| `BudgetNameExists` | 0 | — | Ask reuse vs new name | `[ERROR] BudgetNameExists: A budget with this name already exists` |
| `InvalidParameter` / 400 | 1 | — | Fix parameters; retry | `[ERROR] InvalidParameter: Budget amount must be positive` |
| `QuotaExceeded` | 0 | — | HALT; list existing budgets | `[ERROR] QuotaExceeded: Maximum budgets reached (limit: 20)` |

---

### Operation: DescribeBillDetail — Get Bill Line Items

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Period valid | Validate YYYY-MM format | Valid format | Prompt for correct format |

#### Execution — CLI

```bash
ve billing DescribeBillDetail \
  --BillingCycle "{{user.start_date}}" \
  --MaxResults 100 \
  --Region "{{env.VOLCENGINE_REGION}}"
```

#### Validation

1. Confirm `$.Result.BillDetails` array is non-empty.
2. If pagination marker present, continue fetching until empty.

#### Present to User

| Field | Path | Notes |
|-------|------|-------|
| TotalCost | `$.Result.TotalCost` | Sum of all line items |
| ResourceId | `$.Result.BillDetails[].ResourceId` | Cloud resource ID |
| ProductType | `$.Result.BillDetails[].ProductType` | ecs, rds, vke, etc. |
| Cost | `$.Result.BillDetails[].Cost` | Cost for this line item |
| BillItemName | `$.Result.BillDetails[].BillItemName` | Item description |

---

### Operation: DescribeBalance — Check Account Balance

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | Verify non-empty AK/SK | Valid credentials | HALT |

#### Execution — CLI

```bash
ve billing DescribeBalance --Region "{{env.VOLCENGINE_REGION}}"
```

#### Validation

1. Parse `$.Result.Balance` and confirm non-negative.
2. Report balance and estimate days remaining at current burn rate.

#### Present to User

| Field | Path | Notes |
|-------|------|-------|
| Balance | `$.Result.Balance` | Current balance (CNY) |
| CreditLimit | `$.Result.CreditLimit` | Credit line if applicable |
| Currency | `$.Result.Currency` | Currency code (CNY) |

#### Failure Recovery

| Error Pattern | Max Retries | Backoff | Agent Action | UX Feedback |
|---------------|-------------|---------|--------------|-------------|
| `InsufficientBalance` | 0 | — | Alert user immediately | `🚨 CRITICAL: Account balance insufficient for operations` |
| `Forbidden.RAM` / 403 | 0 | — | HALT; add IAM billing read permissions | `[ERROR] Forbidden.RAM: Cannot read billing information` |

---

### Operation: DescribeReservedInstances — View RI Utilization

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| CLI / deps | `ve billing DescribeReservedInstances --help` | Shows command | Document CLI installation |

#### Execution — CLI

```bash
ve billing DescribeReservedInstances --Region "{{env.VOLCENGINE_REGION}}"
```

#### Validation

1. Parse `$.Result.ReservedInstances` array.
2. Calculate utilization: `UsedUnits / TotalUnits * 100`.

#### Present to User

| Field | Path | Notes |
|-------|------|-------|
| InstanceType | `$.Result.ReservedInstances[].InstanceType` | RI instance type |
| TotalUnits | `$.Result.ReservedInstances[].TotalUnits` | Purchased units |
| UsedUnits | `$.Result.ReservedInstances[].UsedUnits` | Units currently used |
| Utilization | Calculated | Percentage |
| ExpireTime | `$.Result.ReservedInstances[].ExpireTime` | Expiration timestamp |

---

## Error Taxonomy

| Error Code | Meaning | Resolution |
|------------|---------|-----------|
| `InvalidBillId` | Bill ID does not exist | 0 retries; **HALT** — verify bill ID |
| `InvalidStatementId` | Statement ID does not exist | 0 retries; **HALT** — verify statement ID |
| `InvalidResourceGroupId` | Resource group ID does not exist | 0 retries; **HALT** — verify resource group ID |
| `InvalidOrderId` | Order ID does not exist | 0 retries; **HALT** — verify order ID |
| `InvalidCouponId` | Coupon ID does not exist | 0 retries; **HALT** — verify coupon ID |
| `InvalidInvoiceId` | Invoice ID does not exist | 0 retries; **HALT** — verify invoice ID |
| `QuotaExceeded.Query` | Query quota exceeded | 0 retries; **HALT** — reduce query scope |
| `AccountFrozen` | Account is frozen | 0 retries; **HALT** — contact billing support |
| `InsufficientBalance` | Account balance too low | 0 retries; **HALT** — alert user to recharge |
| `ArrearsInProgress` | Account in arrears | 0 retries; **HALT** — settle outstanding balance |
| `Throttling` | Rate limit exceeded | 3 retries/exponential; **RETRY** — Back off, respect Retry-After |
| `InternalError` | Server error | 3 retries/2s/4s/8s; **RETRY** — Retry, escalate with RequestId |

## Reference Directory

- [Core Concepts](references/core-concepts.md)
- [API & SDK Usage](references/api-sdk-usage.md)
- [CLI Usage](references/cli-usage.md)
- [Troubleshooting Guide](references/troubleshooting.md)
- [Monitoring](references/monitoring.md)
- [Integration](references/integration.md)
- [Knowledge Base](references/knowledge-base.md)
- [Cost Optimization Guide](references/cost-optimization.md)
- [Operations Framework](references/operations-framework.md)
- [Cross-Service Cost Analysis](references/cross-service-cost.md)
- [FinOps Knowledge Base](references/finops-knowledge-base.md)
- [Cost Prediction & Forecasting](references/cost-prediction.md)
- [Cross-Skill Orchestration](references/cross-skill-orchestration.md)
- [Observability Integration](references/observability.md)
- [Prompts Handbook](references/prompts.md)
- [User Experience Specification](../ve-skill-generator/references/user-experience-spec.md)
- [Execution Environment Setup](../ve-skill-generator/references/execution-environment.md)
- [CLI Behavioral Reference](../ve-skill-generator/references/cli-behavior.md)
- [FinOps Best Practices](../ve-skill-generator/references/finops-best-practices.md)
- [GCL Rubric](references/rubric.md)
- [GCL Prompt Templates](references/prompt-templates.md)

## Changelog
| Version | Date | Changes |
|---------|------|---------|
| 1.5.1 | 2026-07-13 | T04: annotate operation table with safety_class + blast_radius leaf-op metadata columns (L3 policy inputs); see ve-skill-generator/references/leaf-op-metadata-spec.md |
| 1.5.0 | 2026-06-04 | GCL rollout: added ## Quality Gate (GCL), references/rubric.md, references/prompt-templates.md |
| 1.4.0 | 2026-06-02 | Complete observability integration (CMS/SLS/Prometheus/Slack/PagerDuty), 3 new example assets, full SDK/CLI scripts for all integration patterns |
| 1.3.0 | 2026-06-02 | Full UX compliance: Complete Execution Flows (4-phase), 16 Error Taxonomy entries, Prerequisites, Response Conventions, State Transitions |
| 1.2.0 | 2026-06-02 | Added FinOps Knowledge Base, Cost Prediction, Cross-Skill Orchestration |
| 1.1.0 | 2026-06-02 | Added Cost Optimization Guide, Operations Framework, Cross-Service Cost Analysis |
| 1.0.0 | 2026-05-31 | Initial release with billing query, budget, balance, cost analysis |

## Operational Best Practices

- **Cost governance framework:** Establish a tagging strategy (environment, project, owner) and enforce it across all resources. Use `ve billing DescribeBillDetail` to track spend by tag group.
- **Budget alerts and notifications:** Create budgets via `ve billing CreateBudget` with alert thresholds at 50%, 80%, and 100% of budget. Integrate with CMS alarms for real-time cost anomaly detection.
- **Resource tagging for cost allocation:** Apply consistent tags (e.g., `Department`, `Project`, `Environment`) to all provisioned resources. Review untagged resources monthly to prevent cost leaks.
- **Cross-account cost visibility:** Aggregate billing data across multiple accounts using consolidated billing. Use `ve billing DescribeBills` to generate organization-wide cost reports.
- **Reserved instance planning:** Purchase reserved instances for steady-state workloads to reduce on-demand spend. Monitor RI utilization via `ve billing DescribeReservedInstances` and modify terms if underutilized.
- **Regular cost review cadence:** Schedule weekly cost reviews using cost explorer. Compare actual versus budgeted spend, identify top-cost resources, and rightsize overprovisioned instances.

## Quality Gate (GCL)

> Optional tier. max_iter=5. Read-mostly.

| Tier | Operations | Safety |
|---|---|---|
| **Destructive** | DeleteBudget | 1.0 |
| **State-changing** | CreateBudget, UpdateBudget | 1.0 |
| **Mutating** | — | ≥0.5 |
| **Read-only** | DescribeBills, DescribeBillDetail, DescribeBalance, DescribeBudgets, DescribeReservedInstances | ≥0 |

Safety: DeleteBudget stops cost alerts. Financial data sensitive. VOLCENGINE_SECRET_KEY never.

### Cross-skill: All product cost analysis→respective product ops
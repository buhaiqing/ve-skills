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
  version: "1.0.0"
  last_updated: "2026-05-31"
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

# Volcengine Billing Operations Skill

## Overview

Volcengine Billing (火山引擎费用中心) provides cost management, billing, and account operations. This skill is an **operational runbook** for agents: cost query, budget management, payment operations, invoice management, and cost optimization. **Do not use the web console as the primary agent execution path.**

### CLI applicability

- **`cli_applicability: dual-path`:** Official `ve` CLI supports Billing. Document **both** SDK and CLI steps for every operation.

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use with billing triggers; resource ops delegated to product skills |
| 2 | **Structured I/O** | `{{env.*}}` for credentials, `{{user.*}}` for billing params, `{{output.*}}` for API responses |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | Error taxonomy with billing-specific codes; HALT vs retry per type |
| 5 | **Absolute Single Responsibility** | Billing only; resource provisioning delegated to product skills |

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

## Variable Convention (Agent-Readable)

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{user.start_date}}` | Start of billing period | Format YYYY-MM |
| `{{user.end_date}}` | End of billing period | Format YYYY-MM |
| `{{user.budget_name}}` | Budget name | Ask once; reuse |
| `{{user.budget_amount}}` | Budget limit amount | Decimal, e.g., 10000.00 |
| `{{output.total_cost}}` | Total cost in period | From billing summary response |
| `{{output.budget_id}}` | Created budget ID | From CreateBudget response |

> **Security Warning (Credential Masking — MANDATORY):** NEVER log, print, or expose credentials.

## Quick Start

### What This Skill Does
Query, analyze, and manage Volcengine billing, cost, and account using the `ve` CLI or JIT Go SDK.

### Prerequisites
- [ ] `ve` CLI installed (or Go runtime for JIT fallback)
- [ ] Credentials configured

### Verify Setup
```bash
ve billing DescribeBills --Period 2026-05 --Region {{env.VOLCENGINE_REGION}}
```

## Capabilities at a Glance

| Operation | Description | Complexity | Risk Level |
|-----------|-------------|------------|------------|
| DescribeBills | Query monthly bills | Low | None |
| DescribeBillDetail | Get bill line items | Low | None |
| DescribeBalance | Check account balance | Low | None |
| CreateBudget | Set spending budget | Low | Low |
| DescribeBudgets | List budgets | Low | None |
| DescribeReservedInstances | View RI utilization | Low | None |

## Execution Flows

### Operation: DescribeBills — Query Monthly Cost

#### Execution

```bash
ve billing DescribeBills --Period "{{user.start_date}}" --Region "{{env.VOLCENGINE_REGION}}"
```

#### Validation
Parse `$.Result.Bills[]` and report: total cost, by-product breakdown, top cost drivers.

### Operation: CreateBudget — Set Budget Alert

#### Execution

```bash
ve billing CreateBudget \
  --BudgetAmount {{user.budget_amount}} \
  --BudgetType MONTHLY \
  --AlertThresholds '[80, 90, 100]' \
  --Region "{{env.VOLCENGINE_REGION}}"
```

#### Validation
Verify `{{output.budget_id}}` returned and budget appears in DescribeBudgets.

## Error Taxonomy

| Error Code | HTTP | Meaning | Max Retries |
|-----------|------|---------|-------------|
| InvalidParameter | 400 | Bad request | 1 |
| Forbidden.RAM | 403 | Insufficient permissions | 0 |
| InsufficientBalance | 403 | Account balance too low | 0 |
| InternalError | 500 | Server error | 3 |
| Throttling | 429 | Rate limit | 3 |

## Reference Directory

- [Core Concepts](references/core-concepts.md)
- [API & SDK Usage](references/api-sdk-usage.md)
- [CLI Usage](references/cli-usage.md)
- [Troubleshooting Guide](references/troubleshooting.md)
- [Monitoring](references/monitoring.md)
- [Integration](references/integration.md)
- [Knowledge Base](references/knowledge-base.md)
- [User Experience Specification](../../ve-skill-generator/references/user-experience-spec.md)
- [Execution Environment Setup](../../ve-skill-generator/references/execution-environment.md)
- [CLI Behavioral Reference](../../ve-skill-generator/references/cli-behavior.md)
- [FinOps Best Practices](../../ve-skill-generator/references/finops-best-practices.md)

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-31 | Initial release with billing query, budget, balance, cost analysis |

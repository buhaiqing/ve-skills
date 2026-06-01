# CLI Usage — ve-billing-ops

> **Purpose:** Complete `ve` CLI command reference for Volcengine Billing operations.
> **Version:** 1.3.0
> **Last Updated:** 2026-06-02

---

## Install and Config

See [Execution Environment Setup](../../ve-skill-generator/references/execution-environment.md).

### Credentials

The `ve` CLI reads credentials from:
- Environment variables: `VOLCENGINE_ACCESS_KEY` / `VOLCENGINE_SECRET_KEY` / `VOLCENGINE_REGION`
- Config file: `~/.volcengine/config.json` (JSON format)

### Verify CLI Installation

```bash
ve version
# Expected: ve version 1.0.x or higher

ve billing --help
# Expected: Shows available billing subcommands
```

---

## Conventions

| Aspect | Convention |
|--------|------------|
| **Command prefix** | `ve billing` |
| **Output format** | JSON by default |
| **Parameter style** | `--ParameterName value` |
| **Date format** | YYYY-MM (e.g., `2026-05`) |
| **Region** | Use `{{env.VOLCENGINE_REGION}}` or explicit region |

---

## CLI vs API Coverage

| Operation | ve CLI | Go SDK | Notes |
|-----------|--------|--------|-------|
| DescribeBills | ✅ Yes | ✅ Yes | Monthly billing summary |
| DescribeBillDetail | ✅ Yes | ✅ Yes | Line-item detail with pagination |
| DescribeBalance | ✅ Yes | ✅ Yes | Account balance inquiry |
| CreateBudget | ✅ Yes | ✅ Yes | Create spending budget |
| DescribeBudgets | ✅ Yes | ✅ Yes | List all budgets |
| ModifyBudget | ✅ Yes | ✅ Yes | Update budget threshold |
| DeleteBudget | ✅ Yes | ✅ Yes | Remove budget |
| DescribeReservedInstances | ✅ Yes | ✅ Yes | RI utilization |
| DescribeInvoices | ✅ Yes | ✅ Yes | Invoice history |
| CreateInvoice | ✅ Yes | ✅ Yes | Apply for invoice |
| DescribeCostAnalysis | ❌ No | ✅ Yes | Advanced cost analysis (SDK only) |
| DescribeResourceTags | ✅ Yes | ✅ Yes | Cost allocation tags |

---

## Command Reference

### Billing Query Commands

#### DescribeBills — Query Monthly Bills

```bash
# Basic usage
ve billing DescribeBills --Period "2026-05" --Region "cn-beijing"

# With product type filter
ve billing DescribeBills --Period "2026-05" --ProductType "ecs" --Region "cn-beijing"

# With tag filter
ve billing DescribeBills --Period "2026-05" --TagFilters '[{"Key":"environment","Value":"production"}]' --Region "cn-beijing"
```

**Parameters:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| Period | Yes | string | Billing period in YYYY-MM format |
| ProductType | No | string | Filter by product type (ecs, rds, vke, etc.) |
| TagFilters | No | JSON | Filter by resource tags |
| MaxResults | No | integer | Page size (default: 20, max: 100) |
| Marker | No | string | Pagination marker |
| Region | Yes | string | Region identifier |

**Output Example:**
```json
{
  "Result": {
    "Bills": [{
      "TotalCost": 12345.67,
      "ProductDetail": [
        {"ProductType": "ecs", "Cost": 5000.00},
        {"ProductType": "rds", "Cost": 3000.00}
      ]
    }]
  },
  "RequestId": "req-xxx"
}
```

---

#### DescribeBillDetail — Get Bill Line Items

```bash
# Basic usage
ve billing DescribeBillDetail --BillingCycle "2026-05" --Region "cn-beijing"

# With resource filter
ve billing DescribeBillDetail --BillingCycle "2026-05" --ResourceIds '["i-xxx","i-yyy"]' --Region "cn-beijing"

# With product type and pagination
ve billing DescribeBillDetail --BillingCycle "2026-05" --ProductType "ecs" --MaxResults 50 --Region "cn-beijing"

# Tag-based filtering
ve billing DescribeBillDetail --BillingCycle "2026-05" --TagFilters '[{"Key":"cost-center","Value":"engineering"}]' --Region "cn-beijing"
```

**Parameters:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| BillingCycle | Yes | string | Billing cycle in YYYY-MM format |
| ProductType | No | string | Filter by product type |
| ResourceIds | No | JSON array | Filter by resource IDs |
| TagFilters | No | JSON array | Filter by tags |
| MaxResults | No | integer | Page size (default: 20, max: 100) |
| Marker | No | string | Pagination marker |
| Region | Yes | string | Region identifier |

**Output Example:**
```json
{
  "Result": {
    "TotalCost": 12345.67,
    "BillDetails": [
      {
        "ResourceId": "i-xxx",
        "ResourceName": "api-server-01",
        "ProductType": "ecs",
        "BillItemName": "Instance: ecs.g6.xlarge",
        "Cost": 350.00
      }
    ],
    "Marker": "next-page-token"
  }
}
```

---

#### DescribeBalance — Check Account Balance

```bash
# Basic usage
ve billing DescribeBalance --Region "cn-beijing"
```

**Parameters:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| Region | Yes | string | Region identifier |

**Output Example:**
```json
{
  "Result": {
    "Balance": 50000.00,
    "CreditLimit": 100000.00,
    "Currency": "CNY"
  }
}
```

---

### Budget Management Commands

#### CreateBudget — Create Spending Budget

```bash
# Basic usage
ve billing CreateBudget \
  --BudgetAmount 50000.00 \
  --BudgetType MONTHLY \
  --AlertThresholds '[80, 90, 100]' \
  --BudgetName "production-monthly" \
  --Region "cn-beijing"

# With tag-based scope
ve billing CreateBudget \
  --BudgetAmount 10000.00 \
  --BudgetType MONTHLY \
  --AlertThresholds '[80, 100]' \
  --BudgetName "dev-team-monthly" \
  --TagFilters '[{"Key":"team","Value":"dev"}]' \
  --Region "cn-beijing"
```

**Parameters:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| BudgetName | Yes | string | Unique budget name |
| BudgetAmount | Yes | number | Budget limit amount |
| BudgetType | Yes | string | MONTHLY, WEEKLY, DAILY |
| AlertThresholds | Yes | JSON array | Alert trigger percentages [80, 90, 100] |
| TagFilters | No | JSON array | Scope to resources with specific tags |
| Region | Yes | string | Region identifier |

**Output Example:**
```json
{
  "Result": {
    "BudgetId": "budget-xxx"
  }
}
```

---

#### DescribeBudgets — List Budgets

```bash
# List all budgets
ve billing DescribeBudgets --Region "cn-beijing"

# List with pagination
ve billing DescribeBudgets --MaxResults 20 --Region "cn-beijing"
```

**Parameters:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| MaxResults | No | integer | Page size |
| Marker | No | string | Pagination marker |
| Region | Yes | string | Region identifier |

**Output Example:**
```json
{
  "Result": {
    "Budgets": [
      {
        "BudgetId": "budget-xxx",
        "BudgetName": "production-monthly",
        "BudgetAmount": 50000.00,
        "BudgetType": "MONTHLY",
        "ActualSpent": 35000.00,
        "AlertThresholds": [80, 90, 100],
        "AlertTriggered": false,
        "Currency": "CNY"
      }
    ]
  }
}
```

---

#### ModifyBudget — Update Budget

```bash
# Update budget amount
ve billing ModifyBudget \
  --BudgetId "budget-xxx" \
  --BudgetAmount 60000.00 \
  --Region "cn-beijing"

# Update alert thresholds
ve billing ModifyBudget \
  --BudgetId "budget-xxx" \
  --AlertThresholds '[70, 85, 100]' \
  --Region "cn-beijing"
```

**Parameters:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| BudgetId | Yes | string | Budget identifier |
| BudgetAmount | No | number | New budget limit |
| AlertThresholds | No | JSON array | New alert thresholds |
| Region | Yes | string | Region identifier |

---

#### DeleteBudget — Delete Budget

```bash
# Delete a budget
ve billing DeleteBudget --BudgetId "budget-xxx" --Region "cn-beijing"
```

**Parameters:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| BudgetId | Yes | string | Budget identifier |
| Region | Yes | string | Region identifier |

---

### Reserved Instance Commands

#### DescribeReservedInstances — View RI Utilization

```bash
# List all RIs
ve billing DescribeReservedInstances --Region "cn-beijing"

# Filter by instance type
ve billing DescribeReservedInstances --InstanceType "ecs.g6.xlarge" --Region "cn-beijing"
```

**Parameters:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| InstanceType | No | string | Filter by RI instance type |
| Region | Yes | string | Region identifier |

**Output Example:**
```json
{
  "Result": {
    "ReservedInstances": [
      {
        "ReservedInstanceId": "ri-xxx",
        "InstanceType": "ecs.g6.xlarge",
        "TotalUnits": 10,
        "UsedUnits": 8,
        "Utilization": 80.0,
        "ExpireTime": "2027-05-31T00:00:00Z",
        "Status": "Active"
      }
    ]
  }
}
```

---

### Invoice Commands

#### DescribeInvoices — List Invoices

```bash
# List invoices for current year
ve billing DescribeInvoices --InvoiceType "VAT_NORMAL" --Region "cn-beijing"

# List with date range
ve billing DescribeInvoices \
  --StartDate "2026-01-01" \
  --EndDate "2026-05-31" \
  --Region "cn-beijing"
```

**Parameters:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| InvoiceType | No | string | VAT_NORMAL, VAT_SPECIAL |
| StartDate | No | string | Start date (YYYY-MM-DD) |
| EndDate | No | string | End date (YYYY-MM-DD) |
| MaxResults | No | integer | Page size |
| Region | Yes | string | Region identifier |

---

#### CreateInvoice — Apply for Invoice

```bash
# Apply for invoice
ve billing CreateInvoice \
  --InvoiceType "VAT_NORMAL" \
  --InvoiceAmount 10000.00 \
  --TaxRate 0.06 \
  --Region "cn-beijing"
```

**Parameters:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| InvoiceType | Yes | string | Invoice type |
| InvoiceAmount | Yes | number | Invoice amount (before tax) |
| TaxRate | Yes | number | Tax rate (0.06 for 6%) |
| Region | Yes | string | Region identifier |

---

### Tag Commands

#### DescribeResourceTags — List Resource Tags

```bash
# List all tags
ve billing DescribeResourceTags --Region "cn-beijing"

# Filter by tag key
ve billing DescribeResourceTags --TagKey "cost-center" --Region "cn-beijing"
```

**Parameters:**

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| TagKey | No | string | Filter by tag key |
| Region | Yes | string | Region identifier |

---

## Common Query Patterns

### Daily Cost Check

```bash
# Get today's accumulated cost (requires bill detail with daily granularity)
TODAY=$(date +%Y-%m-%d)
ve billing DescribeBillDetail --BillingCycle "$(date +%Y-%m)" --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '.Result.BillDetails | map(select(.BillItemName | contains("'${TODAY}'"))) | add'
```

### Top 10 Cost Drivers

```bash
ve billing DescribeBillDetail --BillingCycle "$(date +%Y-%m)" --MaxResults 100 --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '.Result.BillDetails | sort_by(.Cost) | reverse | .[:10]'
```

### Cost by Product Type

```bash
ve billing DescribeBillDetail --BillingCycle "$(date +%Y-%m)" --MaxResults 100 --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '[.Result.BillDetails | group_by(.ProductType)[] | {product: .[0].ProductType, total: (map(.Cost) | add)}]'
```

### Budget Utilization Check

```bash
ve billing DescribeBudgets --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '.Result.Budgets[] | {name: .BudgetName, limit: .BudgetAmount, actual: .ActualSpent, percent: (.ActualSpent / .BudgetAmount * 100)}'
```

### RI Coverage Analysis

```bash
ve billing DescribeReservedInstances --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '[.Result.ReservedInstances[] | {type: .InstanceType, coverage: (.UsedUnits / .TotalUnits * 100)}]'
```

---

## Pagination Handling

```bash
#!/bin/bash
# Fetch all bill details with pagination

BILLING_CYCLE="{{user.billing_cycle}}"
MARKER=""
MAX_RESULTS=100

echo "Fetching bill details for $BILLING_CYCLE..."

while true; do
  RESPONSE=$(ve billing DescribeBillDetail \
    --BillingCycle "$BILLING_CYCLE" \
    --MaxResults $MAX_RESULTS \
    --Region "{{env.VOLCENGINE_REGION}}" \
    ${MARKER:+--Marker "$MARKER"})
  
  # Process current page
  echo "$RESPONSE" | jq '.Result.BillDetails[]'
  
  # Check for next page
  MARKER=$(echo "$RESPONSE" | jq -r '.Result.Marker // empty')
  [ -z "$MARKER" ] && break
  
  echo "Fetching next page..."
done
```

---

## Output Parsing Examples

### Parse Total Cost

```bash
TOTAL_COST=$(ve billing DescribeBills --Period "$(date +%Y-%m)" --Region "{{env.VOLCENGINE_REGION}}" | \
  jq -r '.Result.Bills[0].TotalCost')
echo "Total Cost: ¥$TOTAL_COST"
```

### Parse Balance

```bash
BALANCE=$(ve billing DescribeBalance --Region "{{env.VOLCENGINE_REGION}}" | \
  jq -r '.Result.Balance')
echo "Account Balance: ¥$BALANCE"
```

### Parse Budget Status

```bash
ve billing DescribeBudgets --Region "{{env.VOLCENGINE_REGION}}" | \
  jq -r '.Result.Budgets[] | "\(.BudgetName): \(.ActualSpent) / \(.BudgetAmount) (\(.ActualSpent / .BudgetAmount * 100 | . * 100 | . / 100)%)"'
```

### Parse RI Utilization

```bash
ve billing DescribeReservedInstances --Region "{{env.VOLCENGINE_REGION}}" | \
  jq -r '.Result.ReservedInstances[] | "\(.InstanceType): \(.UsedUnits)/\(.TotalUnits) units (\(.Utilization)%)"'
```

---

## JSON Body Usage

Some billing operations require JSON body for complex parameters:

```bash
# Create budget with tag-based scope (using JSON body)
ve billing CreateBudget \
  --body '{
    "BudgetName": "engineering-monthly",
    "BudgetAmount": 50000.00,
    "BudgetType": "MONTHLY",
    "AlertThresholds": [80, 90, 100],
    "TagFilters": [
      {"Key": "cost-center", "Value": "engineering"},
      {"Key": "environment", "Value": "production"}
    ]
  }' \
  --Region "cn-beijing"
```

---

## CLI Behavioral Notes

| Aspect | Behavior |
|--------|----------|
| **JSON output** | All commands return JSON by default |
| **Empty results** | Returns `{"Result": {}, "RequestId": "..."}` |
| **Error format** | `{"Error": {"Code": "...", "Message": "..."}}` |
| **Pagination** | Marker-based; continue until Marker is empty |
| **Null handling** | Use `jq -r` to avoid "null" string output |
| **Large result sets** | Use `--MaxResults` with pagination for performance |
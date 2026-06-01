# API & SDK Usage — ve-billing-ops

> **Purpose:** Complete API and SDK reference for Volcengine Billing operations.
> **Version:** 1.3.0>
> **Last Updated:** 2026-06-02

---

## OpenAPI

| Item | Value |
|------|-------|
| **Spec** | https://www.volcengine.com/docs/6387 |
| **Service ID** | `billing` |
| **Base URL** | `https://open.volcengineapi.com` |
| **API Version** | `2022-05-01` |

---

## SDK Operations Map

| Goal | API operationId | SDK Method | SDK Package |
|------|----------------|------------|-------------|
| **Billing Queries** | | | |
| Query bills | DescribeBills | `DescribeBills` | `service/billing` |
| Query bill detail | DescribeBillDetail | `DescribeBillDetail` | `service/billing` |
| Check balance | DescribeBalance | `DescribeBalance` | `service/billing` |
| Query cost analysis | DescribeCostAnalysis | `DescribeCostAnalysis` | `service/billing` |
| **Budget Management** | | | |
| Create budget | CreateBudget | `CreateBudget` | `service/billing` |
| List budgets | DescribeBudgets | `DescribeBudgets` | `service/billing` |
| Modify budget | ModifyBudget | `ModifyBudget` | `service/billing` |
| Delete budget | DeleteBudget | `DeleteBudget` | `service/billing` |
| **Reserved Instances** | | | |
| List RIs | DescribeReservedInstances | `DescribeReservedInstances` | `service/billing` |
| **Invoice** | | | |
| List invoices | DescribeInvoices | `DescribeInvoices` | `service/billing` |
| Apply invoice | CreateInvoice | `CreateInvoice` | `service/billing` |
| **Tags** | | | |
| List resource tags | DescribeResourceTags | `DescribeResourceTags` | `service/billing` |
| Update resource tags | UpdateResourceTags | `UpdateResourceTags` | `service/billing` |

---

## Go SDK Package

| Item | Value |
|------|-------|
| **Package** | `github.com/volcengine/volc-sdk-golang/service/billing` |
| **Import Path** | `github.com/volcengine/volc-sdk-golang/service/billing` |
| **GitHub** | https://github.com/volcengine/volc-sdk-golang/tree/main/service/billing |

---

## SDK Script Templates

### Template 1: Billing Query

```go
// main.go — Billing Query Script
package main

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/billing"
)

func main() {
    // Initialize billing client
    client := billing.NewInstance()
    client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    // Prepare request parameters
    params := make(map[string]interface{})
    params["Region"] = os.Getenv("VOLCENGINE_REGION")
    params["Period"] = os.Getenv("BILLING_PERIOD") // Format: YYYY-MM

    // Make API call
    resp, err := client.Client.Request("billing", "DescribeBills", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "API call failed: %v\n", err)
        os.Exit(1)
    }

    // Parse response
    var result map[string]interface{}
    if err := json.Unmarshal(resp, &result); err != nil {
        fmt.Fprintf(os.Stderr, "JSON parse failed: %v\n", err)
        os.Exit(1)
    }

    // Extract and print total cost
    bills := result["Result"].(map[string]interface{})["Bills"].([]interface{})
    if len(bills) > 0 {
        bill := bills[0].(map[string]interface{})
        totalCost := bill["TotalCost"].(float64)
        fmt.Printf("Total Cost: %.2f CNY\n", totalCost)
    }
}
```

### Template 2: Bill Detail with Pagination

```go
// main.go — Bill Detail with Pagination
package main

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/billing"
)

func fetchBillDetails(client *billing.Billing, region, billingCycle string) ([]map[string]interface{}, error) {
    var allDetails []map[string]interface{}
    var marker string

    for {
        params := map[string]interface{}{
            "Region":       region,
            "BillingCycle": billingCycle,
            "MaxResults":   100,
        }
        if marker != "" {
            params["Marker"] = marker
        }

        resp, err := client.Client.Request("billing", "DescribeBillDetail", params)
        if err != nil {
            return nil, err
        }

        var result map[string]interface{}
        json.Unmarshal(resp, &result)

        details := result["Result"].(map[string]interface{})["BillDetails"].([]interface{})
        for _, d := range details {
            allDetails = append(allDetails, d.(map[string]interface{}))
        }

        marker = result["Result"].(map[string]interface{})["Marker"].(string)
        if marker == "" {
            break
        }
    }

    return allDetails, nil
}

func main() {
    client := billing.NewInstance()
    client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    details, err := fetchBillDetails(client, "cn-beijing", "2026-05")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to fetch bill details: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("Total bill details: %d\n", len(details))
}
```

### Template 3: Budget Management

```go
// main.go — Budget Management
package main

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/billing"
)

func main() {
    client := billing.NewInstance()
    client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    // Create budget
    params := map[string]interface{}{
        "Region":          os.Getenv("VOLCENGINE_REGION"),
        "BudgetName":      "production-monthly",
        "BudgetAmount":     50000.00,
        "BudgetType":      "MONTHLY",
        "AlertThresholds": []int{80, 90, 100},
    }

    resp, err := client.Client.Request("billing", "CreateBudget", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create budget: %v\n", err)
        os.Exit(1)
    }

    var result map[string]interface{}
    json.Unmarshal(resp, &result)

    budgetId := result["Result"].(map[string]interface{})["BudgetId"].(string)
    fmt.Printf("Budget created: %s\n", budgetId)

    // Describe budgets to verify
    listParams := map[string]interface{}{
        "Region": os.Getenv("VOLCENGINE_REGION"),
    }

    listResp, err := client.Client.Request("billing", "DescribeBudgets", listParams)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to list budgets: %v\n", err)
        os.Exit(1)
    }

    var listResult map[string]interface{}
    json.Unmarshal(listResp, &listResult)

    budgets := listResult["Result"].(map[string]interface{})["Budgets"].([]interface{})
    fmt.Printf("Total budgets: %d\n", len(budgets))
}
```

### Template 4: RI Utilization Analysis

```go
// main.go — RI Utilization Analysis
package main

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/billing"
)

func analyzeRIUtilization(client *billing.Billing, region string) error {
    params := map[string]interface{}{
        "Region": region,
    }

    resp, err := client.Client.Request("billing", "DescribeReservedInstances", params)
    if err != nil {
        return err
    }

    var result map[string]interface{}
    json.Unmarshal(resp, &result)

    ris := result["Result"].(map[string]interface{})["ReservedInstances"].([]interface{})

    fmt.Println("=== RI Utilization Report ===")
    var totalUnits, usedUnits int

    for _, ri := range ris {
        riMap := ri.(map[string]interface{})
        instanceType := riMap["InstanceType"].(string)
        total := int(riMap["TotalUnits"].(float64))
        used := int(riMap["UsedUnits"].(float64))
        utilization := float64(used) / float64(total) * 100

        totalUnits += total
        usedUnits += used

        fmt.Printf("%s: %d/%d units (%.1f%%)\n", instanceType, used, total, utilization)
    }

    if totalUnits > 0 {
        fmt.Printf("\nOverall: %d/%d units (%.1f%%)\n", usedUnits, totalUnits, float64(usedUnits)/float64(totalUnits)*100)
    }

    return nil
}

func main() {
    client := billing.NewInstance()
    client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    if err := analyzeRIUtilization(client, "cn-beijing"); err != nil {
        fmt.Fprintf(os.Stderr, "Analysis failed: %v\n", err)
        os.Exit(1)
    }
}
```

### Template 5: Balance Check with Days Remaining

```go
// main.go — Balance Check with Forecast
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "time"

    "github.com/volcengine/volc-sdk-golang/service/billing"
)

func main() {
    client := billing.NewInstance()
    client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    // Get balance
    params := map[string]interface{}{
        "Region": os.Getenv("VOLCENGINE_REGION"),
    }

    resp, err := client.Client.Request("billing", "DescribeBalance", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to get balance: %v\n", err)
        os.Exit(1)
    }

    var result map[string]interface{}
    json.Unmarshal(resp, &result)

    balanceData := result["Result"].(map[string]interface{})
    balance := balanceData["Balance"].(float64)
    currency := balanceData["Currency"].(string)

    fmt.Printf("Current Balance: %.2f %s\n", balance, currency)

    // Get current month cost for burn rate calculation
    currentPeriod := time.Now().Format("2006-01")
    billingParams := map[string]interface{}{
        "Region": os.Getenv("VOLCENGINE_REGION"),
        "Period": currentPeriod,
    }

    billingResp, err := client.Client.Request("billing", "DescribeBills", billingParams)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to get billing: %v\n", err)
        os.Exit(1)
    }

    var billingResult map[string]interface{}
    json.Unmarshal(billingResp, &billingResult)

    bills := billingResult["Result"].(map[string]interface{})["Bills"].([]interface{})
    if len(bills) > 0 {
        bill := bills[0].(map[string]interface{})
        totalCost := bill["TotalCost"].(float64)

        dayOfMonth := float64(time.Now().Day())
        dailyBurn := totalCost / dayOfMonth

        if dailyBurn > 0 {
            daysRemaining := balance / dailyBurn
            fmt.Printf("Daily Burn Rate: %.2f %s\n", dailyBurn, currency)
            fmt.Printf("Days of Operation Remaining: %.1f days\n", daysRemaining)
        }
    }
}
```

---

## Request / Response Notes

### Pagination

| Parameter | Description |
|-----------|-------------|
| MaxResults | Page size (default: 20, max: 100) |
| Marker | Pagination token (empty = last page) |

**Pagination Pattern:**
```go
for {
    params["Marker"] = marker
    resp, _ := client.Request("billing", "DescribeBillDetail", params)
    
    // Process page...
    
    marker = result["Result"].(map[string]interface{})["Marker"].(string)
    if marker == "" {
        break
    }
}
```

### Date Formats

| Field | Format | Example |
|-------|--------|---------|
| BillingCycle / Period | YYYY-MM | 2026-05 |
| Date range start/end | YYYY-MM-DD | 2026-05-01 |
| Timestamps | ISO 8601 | 2026-05-01T00:00:00Z |

### Currency

All amounts are in CNY unless specified otherwise.

---

## Error Response Structure

```json
{
  "Error": {
    "Code": "InvalidParameter",
    "Message": "The request parameter is invalid",
    "RequestId": "req-xxx-xxx-xxx"
  }
}
```

---

## Common Request Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| Region | Yes | string | Region identifier (e.g., cn-beijing) |
| Period / BillingCycle | Varies | string | Billing period in YYYY-MM format |
| MaxResults | No | integer | Page size (default: 20) |
| Marker | No | string | Pagination marker |

---

## Advanced: Cost Analysis Query

```go
// Cost Analysis — Group by multiple dimensions
params := map[string]interface{}{
    "Region": "cn-beijing",
    "StartDate": "2026-05-01",
    "EndDate": "2026-05-31",
    "GroupBy": []string{"ProductType", "Tag"},
    "TagFilters": []map[string]string{
        {"Key": "environment", "Value": "production"},
    },
    "Granularity": "DAILY", // DAILY, MONTHLY
}

resp, err := client.Client.Request("billing", "DescribeCostAnalysis", params)
```
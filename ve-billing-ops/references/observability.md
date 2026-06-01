# Observability Integration — ve-billing-ops

> **Purpose:** Integration patterns for connecting ve-billing-ops with Volcengine Cloud Monitor (CMS), Log Service (SLS), and third-party observability platforms.
> **Version:** 1.4.0>
> **Last Updated:** 2026-06-02

---

## Table of Contents

1. [Observability Architecture](#1-observability-architecture)
2. [Cloud Monitor Integration](#2-cloud-monitor-integration)
3. [Log Service Integration](#3-log-service-integration)
4. [Budget Alert Webhook](#4-budget-alert-webhook)
5. [Dashboard Data Sources](#5-dashboard-data-sources)
6. [Alert Escalation Patterns](#6-alert-escalation-patterns)
7. [Complete Integration Scripts](#7-complete-integration-scripts)

---

## 1. Observability Architecture

### 1.1 Data Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        FinOps Observability Stack                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                   │
│  │ ve-billing │    │   Cloud     │    │   Log       │                   │
│  │    -ops    │───▶│   Monitor   │───▶│   Service   │                   │
│  │ (Source)   │    │   (CMS)     │    │   (SLS)     │                   │
│  └─────────────┘    └─────────────┘    └─────────────┘                   │
│         │                  │                  │                            │
│         │                  ▼                  ▼                            │
│         │          ┌─────────────┐    ┌─────────────┐                   │
│         │          │   Alert     │    │   Query &   │                   │
│         │          │   Manager   │    │   Analytics │                   │
│         │          └─────────────┘    └─────────────┘                   │
│         │                  │                                             │
│         ▼                  ▼                                             │
│  ┌─────────────────────────────────────────────────────────────────┐     │
│  │                    Dashboard (Grafana/Console)                │     │
│  └─────────────────────────────────────────────────────────────────┘     │
│                                                                           │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 1.2 Metric Sources from Billing

| Metric Category | Source | Metrics |
|-----------------|--------|---------|
| **Spend Metrics** | DescribeBills | TotalCost, ProductCost[] |
| **Budget Metrics** | DescribeBudgets | BudgetAmount, ActualSpent, Utilization |
| **RI Metrics** | DescribeReservedInstances | TotalUnits, UsedUnits, Utilization |
| **Balance Metrics** | DescribeBalance | Balance, CreditLimit |

### 1.3 Integration Points

| Integration | Purpose | Protocol |
|-------------|---------|----------|
| CMS PushMetrics | Export billing metrics to CMS | HTTPS API |
| CMS AlertRules | Trigger alerts on threshold | CMS Alert API |
| SLS LogStores | Store cost analysis logs | SLS Ingest API |
| Webhook | External notification | HTTP POST |
| OpenAPI Export | Prometheus/Grafana | Prometheus format |

---

## 2. Cloud Monitor Integration

### 2.1 Push Custom Metrics (Go SDK)

```go
// push-billing-metrics.go — Complete CMS metrics pusher
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "time"

    "github.com/volcengine/volc-sdk-golang/service/billing"
    "github.com/volcengine/volc-sdk-golang/service/cms"
)

func main() {
    region := os.Getenv("VOLCENGINE_REGION")
    ak := os.Getenv("VOLCENGINE_ACCESS_KEY")
    sk := os.Getenv("VOLCENGINE_SECRET_KEY")

    // Initialize billing client
    billingClient := billing.NewInstance()
    billingClient.SetAccessKey(ak)
    billingClient.SetSecretKey(sk)

    // Get current billing data
    currentPeriod := time.Now().Format("2006-01")
    billingParams := map[string]interface{}{
        "Region": region,
        "Period": currentPeriod,
    }
    billingResp, _ := billingClient.Client.Request("billing", "DescribeBills", billingParams)
    var billingResult map[string]interface{}
    json.Unmarshal(billingResp, &billingResult)
    totalCost := billingResult["Result"].(map[string]interface{})["Bills"].([]interface{})[0].(map[string]interface{})["TotalCost"].(float64)

    // Get budget data
    budgetParams := map[string]interface{}{"Region": region}
    budgetResp, _ := billingClient.Client.Request("billing", "DescribeBudgets", budgetParams)
    var budgetResult map[string]interface{}
    json.Unmarshal(budgetResp, &budgetResult)
    budgets := budgetResult["Result"].(map[string]interface{})["Budgets"].([]interface{})

    // Initialize CMS client
    cmsClient := cms.NewInstance()
    cmsClient.SetAccessKey(ak)
    cmsClient.SetSecretKey(sk)

    // Build metrics
    metrics := []map[string]interface{}{
        {
            "MetricName": "BillingTotalCost",
            "Value":       totalCost,
            "Unit":        "Count",
            "Timestamp":   time.Now().Unix(),
            "Dimensions": map[string]string{
                "BillingCycle": currentPeriod,
                "Currency":     "CNY",
            },
        },
    }

    // Add budget metrics
    for _, b := range budgets {
        budget := b.(map[string]interface{})
        utilization := budget["ActualSpent"].(float64) / budget["BudgetAmount"].(float64) * 100
        metrics = append(metrics, map[string]interface{}{
            "MetricName": "BudgetUtilization",
            "Value":       utilization,
            "Unit":        "Percent",
            "Timestamp":   time.Now().Unix(),
            "Dimensions": map[string]string{
                "BudgetName": budget["BudgetName"].(string),
            },
        })
    }

    // Push to CMS
    params := map[string]interface{}{
        "MetricData": metrics,
        "Project":    "finops_billing",
    }
    resp, err := cmsClient.Client.Request("cms", "PutMetricData", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to push metrics: %v\n", err)
        os.Exit(1)
    }
    fmt.Printf("Metrics pushed: %s\n", string(resp))
}
```

### 2.2 Push Custom Metrics (CLI)

```bash
#!/bin/bash
# push-billing-metrics.sh — Complete CMS metrics pusher using ve CLI

REGION="{{env.VOLCENGINE_REGION}}"

# Get billing data
CURRENT_MONTH=$(date +%Y-%m)
TOTAL_COST=$(ve billing DescribeBills --Period "$CURRENT_MONTH" --Region "$REGION" | jq -r '.Result.Bills[0].TotalCost')

# Get budget data and build metrics JSON
BUDGET_DATA=$(ve billing DescribeBudgets --Region "$REGION")
BUDGET_UTIL=$(echo "$BUDGET_DATA" | jq -r '[.Result.Budgets[] | {
    MetricName: "BudgetUtilization",
    Value: (.ActualSpent / .BudgetAmount * 100),
    Unit: "Percent",
    Timestamp: now | strftime("%s"),
    Dimensions: {BudgetName: .BudgetName}
}]')

# Build complete metrics JSON
METRICS_JSON=$(jq -n \
    --arg period "$CURRENT_MONTH" \
    --argjson totalCost "$TOTAL_COST" \
    --argjson budgetUtil "$BUDGET_UTIL" \
    '{
        MetricData: (
            [{
                MetricName: "BillingTotalCost",
                Value: $totalCost,
                Unit: "Count",
                Timestamp: now | strftime("%s") | tonumber,
                Dimensions: {BillingCycle: $period, Currency: "CNY"}
            }] + $budgetUtil
        ),
        Project: "finops_billing"
    }')

# Push to CMS using ve cms command
ve cms PutMetricData --body "$METRICS_JSON" --Region "$REGION"

echo "✅ Billing metrics pushed to CMS"
```

### 2.3 CMS Alert Rules (Complete)

```bash
#!/bin/bash
# create-cms-alerts.sh — Create CMS alert rules for billing

REGION="{{env.VOLCENGINE_REGION}}"
PROJECT="finops_billing"

# Alert: Budget 80% warning
ve cms CreateAlertRule \
    --AlertRuleName "billing-budget-warning" \
    --AlertRuleRemark "Alert when budget reaches 80%" \
    --MetricProject "$PROJECT" \
    --MetricName "BudgetUtilization" \
    --Period 300 \
    --Statistics "Average" \
    --ComparisonOperator ">=" \
    --Threshold 80 \
    --AlertLevel "warning" \
    --Region "$REGION"

# Alert: Budget 90% high
ve cms CreateAlertRule \
    --AlertRuleName "billing-budget-high" \
    --AlertRuleRemark "Alert when budget reaches 90%" \
    --MetricProject "$PROJECT" \
    --MetricName "BudgetUtilization" \
    --Period 300 \
    --Statistics "Average" \
    --ComparisonOperator ">=" \
    --Threshold 90 \
    --AlertLevel "high" \
    --Region "$REGION"

# Alert: Budget 100% critical
ve cms CreateAlertRule \
    --AlertRuleName "billing-budget-critical" \
    --AlertRuleRemark "Alert when budget is exceeded" \
    --MetricProject "$PROJECT" \
    --MetricName "BudgetUtilization" \
    --Period 300 \
    --Statistics "Average" \
    --ComparisonOperator ">=" \
    --Threshold 100 \
    --AlertLevel "critical" \
    --Region "$REGION"

# Alert: Low balance
ve cms CreateAlertRule \
    --AlertRuleName "billing-low-balance" \
    --AlertRuleRemark "Alert when balance is below threshold" \
    --MetricProject "$PROJECT" \
    --MetricName "BalanceDaysRemaining" \
    --Period 3600 \
    --Statistics "Average" \
    --ComparisonOperator "<=" \
    --Threshold 7 \
    --AlertLevel "high" \
    --Region "$REGION"

echo "✅ CMS alert rules created"
```

### 2.4 Query CMS Metrics

```bash
#!/bin/bash
# query-cms-metrics.sh — Query billing metrics from CMS

REGION="{{env.VOLCENGINE_REGION}}"
PROJECT="finops_billing"

# Query budget utilization
ve cms GetMetricStatistics \
    --MetricProject "$PROJECT" \
    --MetricName "BudgetUtilization" \
    --StartTime "$(date -v-7d +%s)" \
    --EndTime "$(date +%s)" \
    --Period 3600 \
    --Region "$REGION" | jq '.Datapoints'

# Query total cost
ve cms GetMetricStatistics \
    --MetricProject "$PROJECT" \
    --MetricName "BillingTotalCost" \
    --StartTime "$(date -v-30d +%s)" \
    --EndTime "$(date +%s)" \
    --Period 86400 \
    --Region "$REGION" | jq '.Datapoints'
```

---

## 3. Log Service Integration

### 3.1 Push Cost Analysis Logs (Go SDK)

```go
// push-billing-logs.go — Complete SLS log pusher
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "os"
    "time"

    "github.com/volcengine/volc-sdk-golang/service/sls"
)

type LogGroup struct {
    Topic     string      `json:"topic"`
    Source    string      `json:"source"`
    LogTags   []LogTag    `json:"logtags,omitempty"`
    Logs      []LogEntry  `json:"logs"`
}

type LogTag struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

type LogEntry struct {
    Time int64               `json:"time"`
    Tags []LogTag            `json:"keys,omitempty"`
    Contents map[string]string `json:"contents"`
}

func main() {
    region := os.Getenv("VOLCENGINE_REGION")
    ak := os.Getenv("VOLCENGINE_ACCESS_KEY")
    sk := os.Getenv("VOLCENGINE_SECRET_KEY")

    // Initialize SLS client
    slsClient := sls.NewInstance()
    slsClient.SetAccessKey(ak)
    slsClient.SetSecretKey(sk)

    // Build log group
    logGroup := LogGroup{
        Topic:  "finops-cost-analysis",
        Source: "ve-billing-ops",
        LogTags: []LogTag{
            {Key: "env", Value: os.Getenv("VOLCENGINE_REGION")},
            {Key: "version", Value: "1.4.0"},
        },
        Logs: []LogEntry{
            {
                Time: time.Now().Unix(),
                Contents: map[string]string{
                    "event_type":       "cost_analysis",
                    "billing_period":   time.Now().Format("2006-01"),
                    "total_cost":       "12345.67",
                    "top_cost_driver":  "ecs",
                    "budget_utilization": "0.75",
                    "ri_coverage":      "0.82",
                },
            },
        },
    }

    // Serialize to JSON
    data, _ := json.Marshal([]LogGroup{logGroup})

    // Push to SLS
    params := map[string]interface{}{
        "Project":  "finops",
        "LogStore": "cost-analysis",
        "Region":   region,
        "Body":     string(data),
    }
    resp, err := slsClient.Client.Request("sls", "PostLogStoreLogs", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to push logs: %v\n", err)
        os.Exit(1)
    }
    fmt.Printf("Logs pushed: %s\n", string(resp))
}
```

### 3.2 Push Cost Analysis Logs (CLI)

```bash
#!/bin/bash
# push-billing-logs.sh — Complete SLS log pusher using ve CLI

REGION="{{env.VOLCENGINE_REGION}}"
PROJECT="finops"
LOGSTORE="cost-analysis"

# Generate log entries
CURRENT_MONTH=$(date +%Y-%m)
BILL_DATA=$(ve billing DescribeBills --Period "$CURRENT_MONTH" --Region "$REGION")
TOTAL_COST=$(echo "$BILL_DATA" | jq -r '.Result.Bills[0].TotalCost')

BUDGET_DATA=$(ve billing DescribeBudgets --Region "$REGION")
BUDGET_UTIL=$(echo "$BUDGET_DATA" | jq -r '[.Result.Budgets[] | .ActualSpent / .BudgetAmount * 100] | add / length')

RI_DATA=$(ve billing DescribeReservedInstances --Region "$REGION")
RI_UTIL=$(echo "$RI_DATA" | jq -r '[.Result.ReservedInstances[] | .UsedUnits / .TotalUnits * 100] | add / length')

# Build JSON log entries
LOG_ENTRIES=$(jq -n \
    --arg period "$CURRENT_MONTH" \
    --argjson totalCost "$TOTAL_COST" \
    --argjson budgetUtil "$BUDGET_UTIL" \
    --argjson riUtil "$RI_UTIL" \
    '{
        Topic: "finops-cost-analysis",
        Source: "ve-billing-ops",
        Logs: [{
            Time: now | strftime("%s") | tonumber,
            Contents: {
                billing_period: $period,
                total_cost: ($totalCost | tostring),
                budget_utilization: ($budgetUtil | tostring),
                ri_utilization: ($riUtil | tostring),
                timestamp: now | strftime("%Y-%m-%dT%H:%M:%SZ")
            }
        }]
    }')

# Push to SLS
echo "$LOG_ENTRIES" | ve sls PutLogs \
    --Project "$PROJECT" \
    --LogStore "$LOGSTORE" \
    --Region "$REGION"

echo "✅ Billing logs pushed to SLS"
```

### 3.3 SLS LogQL Queries for Billing

```logql
-- Daily cost trend (last 30 days)
* | where __topic__ = "finops-cost-analysis" 
  | where event_type = "cost_analysis"
  | stats avg(total_cost) as daily_cost by date_trunc('day', timestamp) as day
  | sort by day asc

-- Budget utilization over time
* | where __topic__ = "finops-cost-analysis" 
  | where event_type = "budget_update"
  | stats avg(budget_utilization) as utilization by budget_name, date_trunc('hour', timestamp) as hour
  | sort by hour asc

-- Cost anomaly alerts (change > 20%)
* | where __topic__ = "finops-cost-analysis" 
  | where event_type = "cost_anomaly"
  | where cost_change_rate > 0.2
  | sort by cost_change_rate desc

-- Top cost drivers by day
* | where __topic__ = "finops-cost-analysis" 
  | where event_type = "cost_analysis"
  | stats sum(total_cost) as total by top_cost_driver, date_trunc('day', timestamp) as day
  | sort by total desc
```

### 3.4 Create SLS Index

```bash
#!/bin/bash
# create-sls-index.sh — Configure SLS log indexes

REGION="{{env.VOLCENGINE_REGION}}"
PROJECT="finops"
LOGSTORE="cost-analysis"

# Create index configuration
INDEX_CONFIG='{
    "line": {
        "token": [",", " ", "'"'"';'", "'"'"':", "'"'"'\n'"'"'],
        "caseSensitive": false
    },
    "keys": {
        "event_type": {"type": "text", "token": [",", " ", "'"'"'\n'"'"']},
        "billing_period": {"type": "text", "token": [",", " ", "'"'"'\n'"'"']},
        "total_cost": {"type": "long", "token": [",", " "]},
        "budget_utilization": {"type": "double", "token": [",", " "]},
        "top_cost_driver": {"type": "text", "token": [",", " ", "'"'"'\n'"'"']},
        "ri_utilization": {"type": "double", "token": [",", " "]}
    }
}'

ve sls CreateIndex \
    --Project "$PROJECT" \
    --LogStore "$LOGSTORE" \
    --IndexLine FullText \
    --KeySchema "$INDEX_CONFIG" \
    --Region "$REGION"

echo "✅ SLS index created"
```

---

## 4. Budget Alert Webhook

### 4.1 Webhook Integration (Complete)

```bash
#!/bin/bash
# webhook-alerts.sh — Complete webhook alert integration

REGION="{{env.VOLCENGINE_REGION}}"
WEBHOOK_URL="{{user.webhook_url}}"
WEBHOOK_SECRET="{{user.webhook_secret}}"

# Function to send webhook alert
send_alert() {
    local alert_type="$1"
    local severity="$2"
    local budget_name="$3"
    local utilization="$4"
    local actual_spend="$5"
    local budget_limit="$6"

    # Calculate remaining and projected
    remaining=$(echo "$budget_limit - $actual_spend" | bc -l)
    daily_burn=$(echo "$actual_spend / $(date +%d)" | bc -l)
    days_in_month=$(cal | awk '{print NF}' | tail -1)
    days_elapsed=$(date +%d)
    days_remaining=$((days_in_month - days_elapsed))
    projected=$(echo "$daily_burn * $days_in_month" | bc -l)

    # Build payload
    payload=$(jq -n \
        --arg event "$alert_type" \
        --arg source "ve-billing-ops" \
        --arg timestamp "$(date -Iseconds)" \
        --arg bn "$budget_name" \
        --argjson util "$utilization" \
        --argjson actual "$actual_spend" \
        --argjson limit "$budget_limit" \
        --argjson remaining "$remaining" \
        --argjson projected "$projected" \
        --argjson daysRem "$days_remaining" \
        --argjson dailyBurn "$daily_burn" \
        '{
            event: $event,
            source: $source,
            timestamp: $timestamp,
            severity: if $util >= 100 then "critical" elif $util >= 90 then "high" else "warning" end,
            data: {
                budget_name: $bn,
                utilization_percent: ($util | . * 100 | . / 100),
                actual_spend: $actual,
                budget_limit: $limit,
                remaining: $remaining,
                projected_end_of_month: $projected,
                days_remaining: $daysRem,
                daily_burn_rate: $dailyBurn
            },
            recommended_actions: [
                "Review recent cost increases",
                "Identify top cost drivers",
                "Consider pausing non-critical resources"
            ]
        }')

    # Send webhook
    curl -s -X POST "$WEBHOOK_URL" \
        -H "Content-Type: application/json" \
        -H "X-Webhook-Secret: $WEBHOOK_SECRET" \
        -d "$payload" | jq '.'

    echo "✅ Alert sent: $alert_type ($severity)"
}

# Get budget data and send alerts
BUDGETS=$(ve billing DescribeBudgets --Region "$REGION")
BUDGET_LIST=$(echo "$BUDGETS" | jq -r '.Result.Budgets[] | @base64')

for budget in $BUDGET_LIST; do
    budget_data=$(echo "$budget" | base64 -d | jq -r '.')
    name=$(echo "$budget_data" | jq -r '.BudgetName')
    actual=$(echo "$budget_data" | jq -r '.ActualSpent')
    limit=$(echo "$budget_data" | jq -r '.BudgetAmount')
    util=$(echo "$budget_data" | jq -r '.ActualSpent / .BudgetAmount * 100')

    # Check thresholds
    if (( $(echo "$util >= 100" | bc -l) )); then
        send_alert "budget_exceeded" "critical" "$name" "$util" "$actual" "$limit"
    elif (( $(echo "$util >= 90" | bc -l) )); then
        send_alert "budget_high" "high" "$name" "$util" "$actual" "$limit"
    elif (( $(echo "$util >= 80" | bc -l) )); then
        send_alert "budget_warning" "warning" "$name" "$util" "$actual" "$limit"
    fi
done
```

### 4.2 Slack Integration (Complete)

```bash
#!/bin/bash
# slack-alerts.sh — Complete Slack alert integration

SLACK_WEBHOOK="{{user.slack_webhook}}"

# Function to send Slack alert
send_slack_alert() {
    local alert_type="$1"
    local severity="$2"
    local budget_name="$3"
    local utilization="$4"
    local actual_spend="$5"
    local budget_limit="$6"

    # Emoji based on severity
    case $severity in
        critical) emoji="🚨🚨" ;;
        high) emoji="🚨" ;;
        warning) emoji="⚠️" ;;
        *) emoji="💰" ;;
    esac

    # Color based on severity
    case $severity in
        critical) color="#FF0000" ;;
        high) color="#FFA500" ;;
        warning) color="#FFFF00" ;;
        *) color="#36A64F" ;;
    esac

    # Build Slack payload
    payload=$(jq -n \
        --arg emoji "$emoji" \
        --arg title "$alert_type: $budget_name" \
        --arg color "$color" \
        --argjson util "$utilization" \
        --argjson actual "$actual_spend" \
        --argjson limit "$budget_limit" \
        '{
            attachments: [{
                color: $color,
                blocks: [
                    {
                        type: "header",
                        text: {
                            type: "plain_text",
                            text: ($emoji + " Budget Alert: " + $title),
                            emoji: true
                        }
                    },
                    {
                        type: "section",
                        fields: [
                            {type: "mrkdwn", text: "*Utilization:*\n" + (($util / 100) | . * 100 | . / 100 | tostring) + "%"},
                            {type: "mrkdwn", text: "*Actual Spend:*\n¥" + ($actual | tostring)},
                            {type: "mrkdwn", text: "*Budget Limit:*\n¥" + ($limit | tostring)},
                            {type: "mrkdwn", text: "*Remaining:*\n¥" + (($limit - $actual) | tostring)}
                        ]
                    },
                    {
                        type: "context",
                        elements: [{
                            type: "mrkdwn",
                            text: "Sent by ve-billing-ops at " + (now | strftime("%Y-%m-%d %H:%M:%S"))
                        }]
                    }
                ]
            }]
        }')

    curl -s -X POST "$SLACK_WEBHOOK" \
        -H 'Content-Type: application/json' \
        -d "$payload" | jq '.'

    echo "✅ Slack alert sent"
}

# Main execution (same pattern as webhook)
BUDGETS=$(ve billing DescribeBudgets --Region "{{env.VOLCENGINE_REGION}}")
echo "$BUDGETS" | jq -r '.Result.Budgets[] | @base64' | while read budget; do
    budget_data=$(echo "$budget" | base64 -d | jq -r '.')
    name=$(echo "$budget_data" | jq -r '.BudgetName')
    actual=$(echo "$budget_data" | jq -r '.ActualSpent')
    limit=$(echo "$budget_data" | jq -r '.BudgetAmount')
    util=$(echo "$budget_data" | jq -r '.ActualSpent / .BudgetAmount * 100')

    if (( $(echo "$util >= 80" | bc -l) )); then
        severity=$([ "$util" -ge 100 ] && echo "critical" || [ "$util" -ge 90 ] && echo "high" || echo "warning")
        send_slack_alert "Budget" "$severity" "$name" "$util" "$actual" "$limit"
    fi
done
```

---

## 5. Dashboard Data Sources

### 5.1 Prometheus Exporter (Complete)

```go
// prometheus-exporter.go — Complete Prometheus metrics exporter
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "time"

    "github.com/volcengine/volc-sdk-golang/service/billing"
)

func main() {
    region := os.Getenv("VOLCENGINE_REGION")
    ak := os.Getenv("VOLCENGINE_ACCESS_KEY")
    sk := os.Getenv("VOLCENGINE_SECRET_KEY")

    client := billing.NewInstance()
    client.SetAccessKey(ak)
    client.SetSecretKey(sk)

    http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
        // Set current billing period
        currentPeriod := time.Now().Format("2006-01")

        // Get billing data
        billingParams := map[string]interface{}{
            "Region": region,
            "Period": currentPeriod,
        }
        billingResp, _ := client.Client.Request("billing", "DescribeBills", billingParams)
        var billingResult map[string]interface{}
        json.Unmarshal(billingResp, &billingResult)

        totalCost := billingResult["Result"].(map[string]interface{})["Bills"].([]interface{})[0].(map[string]interface{})["TotalCost"].(float64)

        // Get budget data
        budgetParams := map[string]interface{}{"Region": region}
        budgetResp, _ := client.Client.Request("billing", "DescribeBudgets", budgetParams)
        var budgetResult map[string]interface{}
        json.Unmarshal(budgetResp, &budgetResult)
        budgets := budgetResult["Result"].(map[string]interface{})["Budgets"].([]interface{})

        // Get balance
        balanceParams := map[string]interface{}{"Region": region}
        balanceResp, _ := client.Client.Request("billing", "DescribeBalance", balanceParams)
        var balanceResult map[string]interface{}
        json.Unmarshal(balanceResp, &balanceResult)
        balance := balanceResult["Result"].(map[string]interface{})["Balance"].(float64)

        // Get RI data
        riParams := map[string]interface{}{"Region": region}
        riResp, _ := client.Client.Request("billing", "DescribeReservedInstances", riParams)
        var riResult map[string]interface{}
        json.Unmarshal(riResp, &riResult)
        ris := riResult["Result"].(map[string]interface{})["ReservedInstances"].([]interface{})

        // Output Prometheus format
        fmt.Fprintf(w, "# HELP volcengine_billing_total_cost Monthly billing total cost\n")
        fmt.Fprintf(w, "# TYPE volcengine_billing_total_cost gauge\n")
        fmt.Fprintf(w, "volcengine_billing_total_cost{currency=\"CNY\",period=\"%s\"} %f\n", currentPeriod, totalCost)

        fmt.Fprintf(w, "# HELP volcengine_billing_balance Account balance\n")
        fmt.Fprintf(w, "# TYPE volcengine_billing_balance gauge\n")
        fmt.Fprintf(w, "volcengine_billing_balance{currency=\"CNY\"} %f\n", balance)

        fmt.Fprintf(w, "# HELP volcengine_billing_budget_utilization Budget utilization percentage\n")
        fmt.Fprintf(w, "# TYPE volcengine_billing_budget_utilization gauge\n")
        for _, b := range budgets {
            budget := b.(map[string]interface{})
            name := budget["BudgetName"].(string)
            util := budget["ActualSpent"].(float64) / budget["BudgetAmount"].(float64) * 100
            fmt.Fprintf(w, "volcengine_billing_budget_utilization{budget_name=\"%s\"} %f\n", name, util)
        }

        fmt.Fprintf(w, "# HELP volcengine_billing_ri_utilization Reserved instance utilization\n")
        fmt.Fprintf(w, "# TYPE volcengine_billing_ri_utilization gauge\n")
        for _, ri := range ris {
            riMap := ri.(map[string]interface{})
            instanceType := riMap["InstanceType"].(string)
            util := riMap["UsedUnits"].(float64) / riMap["TotalUnits"].(float64) * 100
            fmt.Fprintf(w, "volcengine_billing_ri_utilization{ri_type=\"%s\"} %f\n", instanceType, util)
        }
    })

    http.ListenAndServe(":8080", nil)
}
```

### 5.2 Prometheus Exporter (Bash)

```bash
#!/bin/bash
# prometheus-exporter.sh — Bash-based Prometheus metrics exporter

# This script generates metrics in Prometheus format
# Use with: python3 -m http.server 8080 --directory /tmp

METRICS_FILE="/tmp/billing-metrics.prom"
REGION="{{env.VOLCENGINE_REGION}}"

generate_metrics() {
    CURRENT_MONTH=$(date +%Y-%m)

    {
        echo '# HELP volcengine_billing_total_cost Monthly billing total cost'
        echo '# TYPE volcengine_billing_total_cost gauge'

        TOTAL_COST=$(ve billing DescribeBills --Period "$CURRENT_MONTH" --Region "$REGION" 2>/dev/null | jq -r '.Result.Bills[0].TotalCost // 0')
        echo "volcengine_billing_total_cost{currency=\"CNY\",period=\"$CURRENT_MONTH\"} ${TOTAL_COST:-0}"

        echo ""
        echo '# HELP volcengine_billing_balance Account balance'
        echo '# TYPE volcengine_billing_balance gauge'

        BALANCE=$(ve billing DescribeBalance --Region "$REGION" 2>/dev/null | jq -r '.Result.Balance // 0')
        echo "volcengine_billing_balance{currency=\"CNY\"} ${BALANCE:-0}"

        echo ""
        echo '# HELP volcengine_billing_budget_utilization Budget utilization percentage'
        echo '# TYPE volcengine_billing_budget_utilization gauge'

        ve billing DescribeBudgets --Region "$REGION" 2>/dev/null | jq -r '.Result.Budgets[] | "volcengine_billing_budget_utilization{budget_name=\"" + .BudgetName + "\"} " + (.ActualSpent / .BudgetAmount * 100 | tostring)'

        echo ""
        echo '# HELP volcengine_billing_ri_utilization Reserved instance utilization'
        echo '# TYPE volcengine_billing_ri_utilization gauge'

        ve billing DescribeReservedInstances --Region "$REGION" 2>/dev/null | jq -r '.Result.ReservedInstances[] | "volcengine_billing_ri_utilization{ri_type=\"" + .InstanceType + "\"} " + (.UsedUnits / .TotalUnits * 100 | tostring)'

    } > "$METRICS_FILE"
}

# Initial generation
generate_metrics

# Watch and regenerate every 5 minutes
while true; do
    sleep 300
    generate_metrics
done
```

### 5.3 Grafana Dashboard JSON (Complete)

```json
{
  "annotations": {
    "list": [
      {
        "builtIn": 1,
        "datasource": "-- Grafana --",
        "enable": true,
        "hide": true,
        "iconColor": "rgba(0, 211, 255, 1)",
        "name": "Annotations & Alerts",
        "type": "dashboard"
      }
    ]
  },
  "description": "Volcengine FinOps Dashboard - ve-billing-ops",
  "editable": true,
  "gnetId": null,
  "graphTooltip": 0,
  "id": null,
  "links": [],
  "panels": [
    {
      "datasource": "Prometheus",
      "fieldConfig": {
        "defaults": {
          "color": {"mode": "palette-classic"},
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {"color": "green", "value": null},
              {"color": "yellow", "value": 70},
              {"color": "orange", "value": 85},
              {"color": "red", "value": 100}
            ]
          },
          "unit": "currencyCNY"
        }
      },
      "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0},
      "id": 1,
      "options": {
        "colorMode": "value",
        "graphMode": "area",
        "justifyMode": "auto",
        "orientation": "auto",
        "reduceOptions": {"calcs": ["lastNotNull"], "fields": "", "values": false},
        "textMode": "auto"
      },
      "pluginVersion": "8.0.0",
      "targets": [
        {
          "expr": "volcengine_billing_total_cost",
          "legendFormat": "Total Cost",
          "refId": "A"
        }
      ],
      "title": "Monthly Billing Cost",
      "type": "stat"
    },
    {
      "datasource": "Prometheus",
      "fieldConfig": {
        "defaults": {
          "color": {"mode": "thresholds"},
          "mappings": [],
          "max": 100,
          "min": 0,
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {"color": "green", "value": null},
              {"color": "yellow", "value": 70},
              {"color": "orange", "value": 85},
              {"color": "red", "value": 100}
            ]
          },
          "unit": "percent"
        }
      },
      "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0},
      "id": 2,
      "options": {
        "orientation": "auto",
        "reduceOptions": {"calcs": ["lastNotNull"], "fields": "", "values": false},
        "showThresholdLabels": false,
        "showThresholdMarkers": true
      },
      "pluginVersion": "8.0.0",
      "targets": [
        {
          "expr": "volcengine_billing_budget_utilization",
          "legendFormat": "{{budget_name}}",
          "refId": "A"
        }
      ],
      "title": "Budget Utilization",
      "type": "gauge"
    },
    {
      "datasource": "Prometheus",
      "fieldConfig": {
        "defaults": {
          "color": {"mode": "palette-classic"},
          "custom": {
            "axisLabel": "",
            "axisPlacement": "auto",
            "barAlignment": 0,
            "drawStyle": "line",
            "fillOpacity": 10,
            "gradientMode": "none",
            "hideFrom": {"legend": false, "tooltip": false, "viz": false},
            "lineInterpolation": "linear",
            "lineWidth": 1,
            "pointSize": 5,
            "scaleDistribution": {"type": "linear"},
            "showPoints": "never",
            "spanNulls": true
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [{"color": "green", "value": null}]
          },
          "unit": "currencyCNY"
        }
      },
      "gridPos": {"h": 8, "w": 24, "x": 0, "y": 8},
      "id": 3,
      "options": {
        "legend": {"calcs": [], "displayMode": "list", "placement": "bottom"},
        "tooltip": {"mode": "single"}
      },
      "pluginVersion": "8.0.0",
      "targets": [
        {
          "expr": "rate(volcengine_billing_total_cost[1h]) * 24 * 30",
          "legendFormat": "Projected Monthly",
          "refId": "A"
        }
      ],
      "title": "Cost Trend & Projection",
      "type": "timeseries"
    },
    {
      "datasource": "Prometheus",
      "fieldConfig": {
        "defaults": {
          "color": {"mode": "palette-classic"},
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [{"color": "green", "value": null}]
          },
          "unit": "percent"
        }
      },
      "gridPos": {"h": 8, "w": 12, "x": 0, "y": 16},
      "id": 4,
      "options": {
        "displayLabels": ["percent"],
        "legend": {"displayMode": "list", "placement": "right"},
        "pieType": "pie",
        "reduceOptions": {"calcs": ["lastNotNull"], "fields": "", "values": false},
        "sortBy": "Percent"
      },
      "pluginVersion": "8.0.0",
      "targets": [
        {
          "expr": "volcengine_billing_ri_utilization",
          "legendFormat": "{{ri_type}}",
          "refId": "A"
        }
      ],
      "title": "RI Utilization",
      "type": "piechart"
    },
    {
      "datasource": "Prometheus",
      "fieldConfig": {
        "defaults": {
          "color": {"mode": "thresholds"},
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {"color": "red", "value": null},
              {"color": "orange", "value": 7},
              {"color": "yellow", "value": 14},
              {"color": "green", "value": 30}
            ]
          },
          "unit": "days"
        }
      },
      "gridPos": {"h": 8, "w": 12, "x": 12, "y": 16},
      "id": 5,
      "options": {
        "colorMode": "value",
        "graphMode": "none",
        "justifyMode": "auto",
        "orientation": "auto",
        "reduceOptions": {"calcs": ["lastNotNull"], "fields": "", "values": false},
        "textMode": "auto"
      },
      "pluginVersion": "8.0.0",
      "targets": [
        {
          "expr": "volcengine_billing_balance / (volcengine_billing_total_cost / 30)",
          "legendFormat": "Days Remaining",
          "refId": "A"
        }
      ],
      "title": "Balance Days Remaining",
      "type": "stat"
    }
  ],
  "refresh": "5m",
  "schemaVersion": 30,
  "style": "dark",
  "tags": ["finops", "billing", "volcengine"],
  "templating": {"list": []},
  "time": {"from": "now-30d", "to": "now"},
  "timepicker": {},
  "timezone": "",
  "title": "Volcengine FinOps Dashboard",
  "uid": "volcengine-finops",
  "version": 1
}
```

---

## 6. Alert Escalation Patterns

### 6.1 Escalation Flow

```markdown
## Alert Escalation Flow

```
Level 1: Warning (80% budget)
    │
    ├──▶ Slack/Teams Notification
    │    └──▶ Team Lead notified
    │
    └──▶ If no action in 24h:
         │
         └──▶ Level 2: High (90% budget)
              │
              ├──▶ Page Team Lead
              ├──▶ Email FinOps Manager
              │
              └──▶ If no action in 12h:
                   │
                   └──▶ Level 3: Critical (100% budget)
                        │
                        ├──▶ Page Director
                        ├──▶ Emergency meeting scheduled
                        └──▶ Auto-freeze non-critical resources
```

### 6.2 On-Call Integration (Complete)

```bash
#!/bin/bash
# oncall-escalation.sh — Complete on-call escalation integration

REGION="{{env.VOLCENGINE_REGION}}"
PAGERDUTY_KEY="{{user.pagerduty_key}}"
VICTOROPS_KEY="{{user.victorops_key}}"

escalate_to_pagerduty() {
    local severity="$1"
    local budget_name="$2"
    local utilization="$3"
    local actual_spend="$4"
    local budget_limit="$5"

    # Map severity to PagerDuty
    case $severity in
        critical) pd_severity="critical" ;;
        high) pd_severity="warning" ;;
        warning) pd_severity="info" ;;
    esac

    payload=$(jq -n \
        --arg routingKey "$PAGERDUTY_KEY" \
        --arg eventAction "trigger" \
        --arg summary "Volcengine Budget Alert: $budget_name at ${utilization}%" \
        --arg severity "$pd_severity" \
        --argjson actual "$actual_spend" \
        --argjson limit "$budget_limit" \
        '{
            routing_key: $routingKey,
            event_action: $eventAction,
            payload: {
                summary: $summary,
                severity: $severity,
                source: "ve-billing-ops",
                custom_details: {
                    budget_name: "ve-billing-ops",
                    actual_spend: ($actual | tostring),
                    budget_limit: ($limit | tostring),
                    utilization: "percent"
                }
            }
        }')

    curl -s -X POST "https://events.pagerduty.com/v2/enqueue" \
        -H "Content-Type: application/json" \
        -d "$payload" | jq '.'

    echo "✅ Escalated to PagerDuty: $severity"
}

escalate_to_victorops() {
    local severity="$1"
    local budget_name="$2"
    local utilization="$3"

    # Map severity
    case $severity in
        critical) vo_severity="critical" ;;
        high) vo_severity="warning" ;;
        warning) vo_severity "info" ;;
    esac

    payload=$(jq -n \
        --arg routingKey "$VICTOROPS_KEY" \
        --arg messageType "$vo_severity" \
        --arg entityId "volcengine-budget-$budget_name" \
        --arg displayName "Budget Alert: $budget_name" \
        --argjson util "$utilization" \
        '{
            message_type: $messageType,
            entity_id: $entityId,
            entity_display_name: $displayName,
            state_message: ("Budget utilization: " + (($util / 100) | . * 100 | . / 100 | tostring) + "%")
        }')

    curl -s -X POST "https://alert.victorops.com/integrations/generic/20171128/alert/$routingKey" \
        -H "Content-Type: application/json" \
        -d "$payload" | jq '.'

    echo "✅ Escalated to VictorOps: $severity"
}

# Main execution
BUDGETS=$(ve billing DescribeBudgets --Region "$REGION")
echo "$BUDGETS" | jq -r '.Result.Budgets[] | @base64' | while read budget; do
    budget_data=$(echo "$budget" | base64 -d | jq -r '.')
    name=$(echo "$budget_data" | jq -r '.BudgetName')
    actual=$(echo "$budget_data" | jq -r '.ActualSpent')
    limit=$(echo "$budget_data" | jq -r '.BudgetAmount')
    util=$(echo "$budget_data" | jq -r '.ActualSpent / .BudgetAmount * 100')

    if (( $(echo "$util >= 100" | bc -l) )); then
        escalate_to_pagerduty "critical" "$name" "$util" "$actual" "$limit"
        escalate_to_victorops "critical" "$name" "$util"
    elif (( $(echo "$util >= 90" | bc -l) )); then
        escalate_to_pagerduty "high" "$name" "$util" "$actual" "$limit"
    fi
done
```

---

## 7. Complete Integration Scripts

### 7.1 Daily FinOps Metrics Collector

```bash
#!/bin/bash
# finops-metrics-collector.sh — Complete daily metrics collection

set -e

REGION="{{env.VOLCENGINE_REGION}}"
PROJECT="finops"
LOGSTORE="cost-analysis"

echo "=== FinOps Metrics Collection ==="
echo "Started: $(date -Iseconds)"

# 1. Collect billing metrics
echo "1/4: Collecting billing metrics..."
CURRENT_MONTH=$(date +%Y-%m)

BILL_DATA=$(ve billing DescribeBills --Period "$CURRENT_MONTH" --Region "$REGION")
TOTAL_COST=$(echo "$BILL_DATA" | jq -r '.Result.Bills[0].TotalCost')
echo "   Total Cost: ¥$TOTAL_COST"

# 2. Collect budget metrics
echo "2/4: Collecting budget metrics..."
BUDGET_DATA=$(ve billing DescribeBudgets --Region "$REGION")
BUDGET_COUNT=$(echo "$BUDGET_DATA" | jq -r '.Result.Budgets | length')
echo "   Active Budgets: $BUDGET_COUNT"

# 3. Collect RI metrics
echo "3/4: Collecting RI metrics..."
RI_DATA=$(ve billing DescribeReservedInstances --Region "$REGION")
RI_COUNT=$(echo "$RI_DATA" | jq -r '.Result.ReservedInstances | length')
RI_UTIL=$(echo "$RI_DATA" | jq -r '[.Result.ReservedInstances[] | .UsedUnits / .TotalUnits * 100] | add / length')
echo "   Reserved Instances: $RI_COUNT (Utilization: ${RI_UTIL}%)"

# 4. Push to SLS
echo "4/4: Pushing logs to SLS..."
LOG_ENTRY=$(jq -n \
    --arg period "$CURRENT_MONTH" \
    --argjson totalCost "$TOTAL_COST" \
    --argjson budgetCount "$BUDGET_COUNT" \
    --argjson riCount "$RI_COUNT" \
    --argjson riUtil "$RI_UTIL" \
    '{
        Topic: "finops-daily-metrics",
        Source: "ve-billing-ops",
        Logs: [{
            Time: now | strftime("%s") | tonumber,
            Contents: {
                billing_period: $period,
                total_cost: ($totalCost | tostring),
                budget_count: ($budgetCount | tostring),
                ri_count: ($riCount | tostring),
                ri_utilization: ($riUtil | tostring),
                timestamp: now | strftime("%Y-%m-%dT%H:%M:%SZ")
            }
        }]
    }')

ve sls PutLogs --Project "$PROJECT" --LogStore "$LOGSTORE" --Body "$LOG_ENTRY" --Region "$REGION" 2>/dev/null || true

# 5. Push to CMS
echo "5/5: Pushing metrics to CMS..."
METRICS_JSON=$(jq -n \
    --arg period "$CURRENT_MONTH" \
    --argjson totalCost "$TOTAL_COST" \
    --argjson budgetCount "$BUDGET_COUNT" \
    --argjson riUtil "$RI_UTIL" \
    '{
        MetricData: [
            {
                MetricName: "BillingTotalCost",
                Value: $totalCost,
                Unit: "Count",
                Timestamp: now | strftime("%s") | tonumber,
                Dimensions: {BillingCycle: $period, Currency: "CNY"}
            },
            {
                MetricName: "BudgetCount",
                Value: $budgetCount,
                Unit: "Count",
                Timestamp: now | strftime("%s") | tonumber,
                Dimensions: {Region: "all"}
            },
            {
                MetricName: "RIUtilization",
                Value: $riUtil,
                Unit: "Percent",
                Timestamp: now | strftime("%s") | tonumber,
                Dimensions: {Region: "all"}
            }
        ],
        Project: "finops_billing"
    }')

ve cms PutMetricData --body "$METRICS_JSON" --Region "$REGION" 2>/dev/null || true

echo ""
echo "=== Collection Complete ==="
echo "Completed: $(date -Iseconds)"
```

---

## Metrics Catalog

### Custom Metrics for Billing

| Metric Name | Type | Unit | Labels | Description |
|-------------|------|------|--------|-------------|
| `billing_total_cost` | Gauge | CNY | period, currency | Total billing cost |
| `billing_cost_by_product` | Gauge | CNY | period, product | Cost breakdown by product |
| `billing_budget_utilization` | Gauge | Percent | budget_name | Budget usage percentage |
| `billing_budget_remaining` | Gauge | CNY | budget_name | Budget remaining amount |
| `billing_balance` | Gauge | CNY | currency | Account balance |
| `billing_ri_utilization` | Gauge | Percent | ri_type | RI utilization rate |
| `billing_ri_coverage` | Gauge | Percent | ri_type | RI coverage rate |
| `billing_daily_burn_rate` | Gauge | CNY/day | period | Daily burn rate |
| `billing_projected_monthly` | Gauge | CNY | period | Projected month-end cost |
| `billing_balance_days_remaining` | Gauge | days | — | Days of operation remaining |
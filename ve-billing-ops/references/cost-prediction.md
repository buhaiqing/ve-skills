# Cost Prediction & Forecasting — ve-billing-ops

> **Purpose:** Models and workflows for predicting future spending, enabling proactive budget management and anomaly detection.
> **Version:** 1.0.0>
> **Last Updated:** 2026-06-02

---

## Table of Contents

1. [Prediction Models](#1-prediction-models)
2. [Forecasting Workflows](#2-forecasting-workflows)
3. [Anomaly Detection](#3-anomaly-detection)
4. [Budget Projection](#4-budget-projection)
5. [Alert Threshold Calculator](#5-alert-threshold-calculator)

---

## 1. Prediction Models

### 1.1 Simple Moving Average (SMA)

**Best for:** Stable, linear cost growth

```python
# Python example for SMA prediction
import json

def sma_predict(costs, periods=3):
    """Simple Moving Average prediction"""
    recent_costs = costs[-periods:]
    avg = sum(recent_costs) / len(recent_costs)
    return avg

# Example usage
monthly_costs = [10000, 10500, 10800, 11000, 11500]
predicted = sma_predict(monthly_costs, periods=3)
print(f"Predicted next month: ¥{predicted:,.2f}")  # ¥11,100.00
```

### 1.2 Weighted Moving Average (WMA)

**Best for:** Recent trends are more indicative

```python
def wma_predict(costs, weights=[0.1, 0.2, 0.3, 0.4]):
    """Weighted Moving Average (recent gets higher weight)"""
    recent = costs[-len(weights):]
    weighted_sum = sum(c * w for c, w in zip(recent, weights))
    return weighted_sum

# Example usage
monthly_costs = [10000, 10500, 10800, 11000, 11500]
predicted = wma_predict(monthly_costs)
print(f"Predicted next month: ¥{predicted:,.2f}")  # ¥11,150.00
```

### 1.3 Linear Regression Trend

**Best for:** Costs growing at a consistent rate

```python
def linear_trend_predict(costs):
    """Linear regression for trend prediction"""
    n = len(costs)
    x = list(range(n))  # [0, 1, 2, 3, 4]
    y = costs
    
    # Calculate slope and intercept
    x_mean = sum(x) / n
    y_mean = sum(y) / n
    
    numerator = sum((x[i] - x_mean) * (y[i] - y_mean) for i in range(n))
    denominator = sum((x[i] - x_mean) ** 2 for i in range(n))
    
    slope = numerator / denominator
    intercept = y_mean - slope * x_mean
    
    # Predict next period
    next_x = n  # 5th month index
    predicted = slope * next_x + intercept
    
    return predicted, slope

# Example usage
monthly_costs = [10000, 10500, 10800, 11000, 11500]
predicted, growth_rate = linear_trend_predict(monthly_costs)
print(f"Predicted: ¥{predicted:,.2f}")
print(f"Monthly growth: ¥{growth_rate:,.2f}")
```

### 1.4 Seasonality-Aware Prediction

**Best for:** Costs with recurring patterns (weekends, month-end)

```python
def seasonality_predict(costs, periods=12):
    """
    Seasonality-aware prediction
    Adjusts for recurring patterns (monthly seasonality)
    """
    n = len(costs)
    if n < periods:
        return sum(costs) / n  # Fall back to average
    
    # Calculate seasonal indices
    seasonal_indices = []
    for i in range(periods):
        period_values = [costs[j] for j in range(i, n, periods) if j < n]
        if period_values:
            avg = sum(period_values) / len(period_values)
            overall_avg = sum(costs) / n
            seasonal_indices.append(avg / overall_avg if overall_avg > 0 else 1.0)
    
    # Base prediction (simple average of last period)
    base_pred = sum(costs[-periods:]) / periods
    
    # Apply seasonal adjustment for next period
    next_seasonal_index = seasonal_indices[n % periods] if seasonal_indices else 1.0
    
    predicted = base_pred * next_seasonal_index
    return predicted

# Example usage
# 12 months of data showing Q4 higher costs
monthly_costs = [9000, 9200, 9500, 10000, 9800, 9600, 9400, 9800, 10200, 11000, 12000, 11500]
predicted = seasonality_predict(monthly_costs)
print(f"Predicted next month (Jan-like): ¥{predicted:,.2f}")
```

---

## 2. Forecasting Workflows

### 2.1 Daily Burn Rate Forecast

```bash
#!/bin/bash
# Daily cost projection to month-end

# Get current date info
CURRENT_MONTH=$(date +%Y-%m)
TODAY=$(date +%d)
DAYS_IN_MONTH=$(date -v-1d -j +%t%B | awk '{print $NF}')  # Days in current month
DAYS_ELAPSED=$(( $(date +%d) ))

# Query current month spend
CURRENT_COST=$(ve billing DescribeBills --Period "$CURRENT_MONTH" --Region "{{env.VOLCENGINE_REGION}}" | \
  jq -r '.Result.Bills[0].TotalCost // 0')

# Calculate daily burn rate
DAILY_BURN=$(echo "scale=2; $CURRENT_COST / $DAYS_ELAPSED" | bc)

# Project to month-end
DAYS_REMAINING=$(( $DAYS_IN_MONTH - $DAYS_ELAPSED + 1 ))
PROJECTED_MONTHLY=$(echo "scale=2; $DAILY_BURN * $DAYS_IN_MONTH" | bc)

echo "=== Cost Forecast ==="
echo "Days Elapsed: $DAYS_ELAPSED / $DAYS_IN_MONTH"
echo "Current Spend: ¥$CURRENT_COST"
echo "Daily Burn Rate: ¥$DAILY_BURN"
echo "Projected Month-End: ¥$PROJECTED_MONTHLY"
```

### 2.2 Weekly Trend Projection

```bash
#!/bin/bash
# Weekly trend analysis and projection

CURRENT_MONTH=$(date +%Y-%m)
LAST_MONTH=$(date -v-1m +%Y-%m)

# Get current month cost
CURRENT_COST=$(ve billing DescribeBills --Period "$CURRENT_MONTH" --Region "{{env.VOLCENGINE_REGION}}" | \
  jq -r '.Result.Bills[0].TotalCost // 0')

# Get last month cost
LAST_MONTH_COST=$(ve billing DescribeBills --Period "$LAST_MONTH" --Region "{{env.VOLCENGINE_REGION}}" | \
  jq -r '.Result.Bills[0].TotalCost // 0')

# Calculate week-over-week if we have partial data
DAYS_ELAPSED=$(date +%d)
WEEKS_ELAPSED=$(echo "scale=2; $DAYS_ELAPSED / 7" | bc)

# Estimate based on last month
WEEKLY_ESTIMATE=$(echo "scale=2; $CURRENT_COST / $WEEKS_ELAPSED * 4.3" | bc)

# Compare with last month
if [ "$LAST_MONTH_COST" != "null" ] && [ "$LAST_MONTH_COST" != "0" ]; then
  MOM_CHANGE=$(echo "scale=2; ($CURRENT_COST - $LAST_MONTH_COST) / $LAST_MONTH_COST * 100" | bc)
  echo "MoM Change: ${MOM_CHANGE}%"
fi

echo "=== Weekly Trend ==="
echo "Weeks Elapsed: $WEEKS_ELAPSED"
echo "Projected (Weekly Burn): ¥$WEEKLY_ESTIMATE"
```

### 2.3 Multi-Scenario Projection

```markdown
## Multi-Scenario Projection Model

### Scenario Parameters
| Parameter | Conservative | Baseline | Optimistic |
|-----------|--------------|----------|------------|
| Growth Rate | +15% MoM | +10% MoM | +5% MoM |
| New Resources | +20% | +10% | 0% |
| Optimization | -5% | -10% | -15% |

### Projection Table

| Month | Conservative | Baseline | Optimistic |
|-------|-------------|----------|------------|
| Current | ¥100,000 | ¥100,000 | ¥100,000 |
| +1 Month | ¥115,000 | ¥110,000 | ¥105,000 |
| +2 Months | ¥132,250 | ¥121,000 | ¥110,250 |
| +3 Months | ¥152,088 | ¥133,100 | ¥115,763 |
| +6 Months | ¥231,106 | ¥176,234 | ¥133,823 |

### Budget Recommendation
Based on Baseline scenario + 10% buffer: **¥140,000/month**

### Confidence Intervals
- 80% CI: ± 8%
- 95% CI: ± 15%
```

---

## 3. Anomaly Detection

### 3.1 Statistical Anomaly Detection

```python
def detect_anomaly_statistical(costs, threshold=2.0):
    """
    Detect anomalies using standard deviation
    threshold: number of standard deviations (2.0 = 95% confidence)
    """
    import math
    
    n = len(costs)
    if n < 3:
        return []  # Need at least 3 data points
    
    mean = sum(costs) / n
    variance = sum((x - mean) ** 2 for x in costs) / n
    std_dev = math.sqrt(variance)
    
    anomalies = []
    for i, cost in enumerate(costs):
        z_score = (cost - mean) / std_dev if std_dev > 0 else 0
        if abs(z_score) > threshold:
            anomalies.append({
                'index': i,
                'cost': cost,
                'z_score': z_score,
                'severity': 'high' if abs(z_score) > 3 else 'medium'
            })
    
    return anomalies

# Example usage
monthly_costs = [10000, 10200, 10100, 10500, 15000, 10800, 11000]
anomalies = detect_anomaly_statistical(monthly_costs)
for a in anomalies:
    print(f"Anomaly at month {a['index']}: ¥{a['cost']} (z-score: {a['z_score']:.2f})")
```

### 3.2 Rate-of-Change Detection

```python
def detect_rate_anomaly(costs, mom_threshold=0.20, mom_critical=0.50):
    """
    Detect anomalies based on rate of change
    mom_threshold: warning threshold (20% change)
    mom_critical: critical threshold (50% change)
    """
    anomalies = []
    
    for i in range(1, len(costs)):
        prev = costs[i-1]
        curr = costs[i]
        
        if prev == 0:
            continue
            
        change_rate = (curr - prev) / prev
        
        if abs(change_rate) > mom_critical:
            severity = 'critical'
        elif abs(change_rate) > mom_threshold:
            severity = 'warning'
        else:
            continue
            
        anomalies.append({
            'index': i,
            'previous_cost': prev,
            'current_cost': curr,
            'change_rate': change_rate * 100,
            'severity': severity
        })
    
    return anomalies

# Example usage
monthly_costs = [10000, 10200, 10100, 10500, 18000, 11000, 11200]
anomalies = detect_rate_anomaly(monthly_costs)
for a in anomalies:
    print(f"{a['severity'].upper()}: Month {a['index']} changed {a['change_rate']:+.1f}% (¥{a['previous_cost']} → ¥{a['current_cost']})")
```

### 3.3 Anomaly Investigation Workflow

```bash
#!/bin/bash
# Anomaly investigation playbook

ANOMALY_MONTH="{{user.month}}"  # e.g., 2026-05
ANOMALY_COST="{{user.anomaly_cost}}"  # e.g., 50000
BASELINE_COST="{{user.baseline_cost}}"  # e.g., 35000

echo "=== Anomaly Investigation ==="
echo "Anomaly Month: $ANOMALY_MONTH"
echo "Anomaly Cost: ¥$ANOMALY_COST"
echo "Baseline Cost: ¥$BASELINE_COST"

# Step 1: Get product breakdown
echo ""
echo "--- Step 1: Product Breakdown ---"
ve billing DescribeBillDetail --BillingCycle "$ANOMALY_MONTH" --Region "{{env.VOLCENGINE_REGION}}" | \
  jq '.Result.BillDetails | group_by(.ProductType) | map({product: .[0].ProductType, total: (map(.Cost) | add)})'

# Step 2: Compare with last month
echo ""
echo "--- Step 2: Month-over-Month Comparison ---"
PREV_MONTH=$(date -v-1m -j -f "%Y-%m" "$ANOMALY_MONTH" +%Y-%m)
ve billing DescribeBillDetail --BillingCycle "$ANOMALY_MONTH" --Region "{{env.VOLCENGINE_REGION}}" | \
  jq --argfile prev <(ve billing DescribeBillDetail --BillingCycle "$PREV_MONTH" --Region "{{env.VOLCENGINE_REGION}}" 2>/dev/null) \
  'def compare: . as $curr | $prev | .Result.BillDetails // [] | map({ResourceId, Cost}) | index(map({ResourceId, Cost})) as $idx | if $idx then "found" else "new" end; "Comparison complete"'

# Step 3: Identify new resources
echo ""
echo "--- Step 3: New Resources ---"
# This would identify resources created in anomaly month
echo "Need to cross-reference with product APIs to identify new resources"

# Step 4: Check for price changes
echo ""
echo "--- Step 4: Price Change Investigation ---"
echo "Check Volcengine pricing updates for $ANOMALY_MONTH"
```

---

## 4. Budget Projection

### 4.1 Dynamic Budget Adjustment

```markdown
## Dynamic Budget Calculation

### Input Parameters
| Parameter | Value | Source |
|-----------|-------|--------|
| Current Spend | ¥80,000 | DescribeBills |
| Days Elapsed | 15 | System date |
| Budget Limit | ¥100,000 | Existing budget |
| RI Commitment | ¥5,000/month | DescribeReservedInstances |

### Calculations
| Metric | Formula | Value |
|--------|---------|-------|
| Daily Burn Rate | ¥80,000 / 15 | ¥5,333/day |
| Projected Month-End | ¥5,333 × 30 | ¥160,000 |
| Remaining Days | 30 - 15 | 15 days |
| Safe Remaining Budget | ¥100,000 - ¥80,000 | ¥20,000 |
| Allowed Daily Spend | ¥20,000 / 15 | ¥1,333/day |

### Decision Matrix
| Scenario | Condition | Action |
|----------|-----------|--------|
| Over Budget | Projected > Budget × 1.1 | Alert + Freeze non-critical |
| On Track | Projected < Budget × 0.9 | Continue monitoring |
| Under Budget | Projected < Budget × 0.7 | Consider budget reduction |
| RI Benefit | RI covers 40% usage | Adjust calculation |
```

### 4.2 Budget Alert Thresholds

```bash
#!/bin/bash
# Calculate adaptive alert thresholds

CURRENT_COST={{output.current_cost}}
MONTHLY_BUDGET={{user.budget_amount}}
DAYS_ELAPSED={{user.days_elapsed}}
DAYS_IN_MONTH={{user.days_in_month}}

# Calculate burn rate
DAILY_BURN=$(echo "scale=2; $CURRENT_COST / $DAYS_ELAPSED" | bc)
PROJECTED=$(echo "scale=2; $DAILY_BURN * $DAYS_IN_MONTH" | bc)
REMAINING_DAYS=$(( $DAYS_IN_MONTH - $DAYS_ELAPSED ))
SAFE_DAILY=$(echo "scale=2; ($MONTHLY_BUDGET - $CURRENT_COST) / $REMAINING_DAYS" | bc)

# Calculate alert thresholds
ALERT_80=$(echo "scale=2; $MONTHLY_BUDGET * 0.8" | bc)
ALERT_90=$(echo "scale=2; $MONTHLY_BUDGET * 0.9" | bc)
ALERT_100=$MONTHLY_BUDGET

echo "=== Budget Status ==="
echo "Current Cost: ¥$CURRENT_COST"
echo "Monthly Budget: ¥$MONTHLY_BUDGET"
echo "Utilization: $(echo "scale=1; $CURRENT_COST / $MONTHLY_BUDGET * 100" | bc)%"
echo ""
echo "=== Alert Thresholds ==="
echo "80% Alert: ¥$ALERT_80 (¥$(echo "scale=2; $ALERT_80 - $CURRENT_COST" | bc) remaining)"
echo "90% Alert: ¥$ALERT_90 (¥$(echo "scale=2; $ALERT_90 - $CURRENT_COST" | bc) remaining)"
echo "100% Alert: ¥$ALERT_100 (¥$(echo "scale=2; $ALERT_100 - $CURRENT_COST" | bc) remaining)"
echo ""
echo "=== Projection ==="
echo "Projected Month-End: ¥$PROJECTED"
echo "Daily Burn Rate: ¥$DAILY_BURN"
echo "Safe Daily Spend: ¥$SAFE_DAILY"
```

---

## 5. Alert Threshold Calculator

### 5.1 Smart Alert Thresholds

```python
def calculate_alert_thresholds(budget, days_elapsed, days_in_month, current_cost, patterns=None):
    """
    Calculate intelligent alert thresholds based on spending patterns
    """
    remaining_budget = budget - current_cost
    remaining_days = days_in_month - days_elapsed
    
    # Base thresholds (static percentages of budget)
    thresholds = {
        'warning': budget * 0.80,
        'high': budget * 0.90,
        'critical': budget * 0.95,
        'exceeded': budget
    }
    
    # Adjust based on spending velocity
    daily_burn = current_cost / days_elapsed if days_elapsed > 0 else 0
    required_daily = remaining_budget / remaining_days if remaining_days > 0 else float('inf')
    
    if daily_burn > required_daily * 1.2:
        # Spending too fast, lower thresholds
        thresholds['warning'] = budget * 0.70
        thresholds['high'] = budget * 0.85
        adjustment_note = "Spending velocity detected - thresholds tightened"
    elif daily_burn < required_daily * 0.8:
        # Spending under control, standard thresholds
        adjustment_note = "On track - standard thresholds"
    else:
        adjustment_note = "Spending within normal range"
    
    # Apply historical patterns if available
    if patterns and 'seasonal_spikes' in patterns:
        spike_months = patterns['seasonal_spikes']
        current_month = (datetime.now().month) % 12 + 1
        if current_month in spike_months:
            # Add buffer for seasonal spike months
            thresholds = {k: v * 1.1 for k, v in thresholds.items()}
            adjustment_note += " (seasonal buffer applied)"
    
    return thresholds, {
        'daily_burn': daily_burn,
        'required_daily': required_daily,
        'velocity_status': 'fast' if daily_burn > required_daily * 1.2 else 'slow' if daily_burn < required_daily * 0.8 else 'normal',
        'adjustment_note': adjustment_note
    }

# Example output
budget = 100000
thresholds, info = calculate_alert_thresholds(
    budget=100000,
    days_elapsed=15,
    days_in_month=30,
    current_cost=65000,
    patterns={'seasonal_spikes': [11, 12]}  # Q4 tends to be higher
)
print(f"Thresholds: {thresholds}")
print(f"Info: {info}")
```

### 5.2 Alert Recommendation Table

```markdown
## Alert Threshold Recommendations

### Standard Environment (Stable Traffic)
| Alert | Threshold | Message | Action |
|-------|-----------|---------|--------|
| Warning | 80% | Budget 80% reached | Review spend |
| High | 90% | Budget 90% reached | Freeze new resources |
| Critical | 100% | Budget exceeded | Immediate action |
| Emergency | 110% | 10% over budget | Escalate to management |

### Production Environment (High Value)
| Alert | Threshold | Message | Action |
|-------|-----------|---------|--------|
| Warning | 75% | Budget 75% reached | Alert team lead |
| High | 85% | Budget 85% reached | Pause non-critical |
| Critical | 95% | Budget 95% reached | Escalate |
| Emergency | 105% | 5% over budget | Executive notification |

### Dev/Test Environment (Cost Sensitive)
| Alert | Threshold | Message | Action |
|-------|-----------|---------|--------|
| Warning | 50% | Budget 50% reached | Notify team |
| High | 70% | Budget 70% reached | Review resources |
| Critical | 90% | Budget 90% reached | Stop non-critical |
| Emergency | 100% | Budget reached | Terminate dev resources |
```
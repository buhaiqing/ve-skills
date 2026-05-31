# Billing Knowledge Base

## Pattern: Unexpected Cost Spike

**Root Causes:**
1. New resource creation without budget tracking
2. Data transfer costs higher than expected
3. Reserved Instance expiration (converted to On-Demand)

**Resolution Steps:**
1. Query DescribeBillDetail filtered by resource
2. Compare month-over-month with previous periods
3. Identify cost driver resource type
4. Recommend right-sizing or RI purchase

## Pattern: Insufficient Balance

**Resolution Steps:**
1. Check current balance: `ve billing DescribeBalance`
2. Review recent bills for cost trends
3. Recommend budget creation
4. Suggest charging or converting to PrePaid

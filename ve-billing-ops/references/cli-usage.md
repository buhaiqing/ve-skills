## Install and Config

See [Execution Environment Setup](../../ve-skill-generator/references/execution-environment.md).

## Conventions

- Command prefix: `ve billing`
- Output is JSON by default

## CLI vs API Coverage

| Operation | Available via `ve`? | Notes |
|-----------|---------------------|-------|
| DescribeBills | Yes | Monthly billing summary |
| DescribeBillDetail | Yes | Line-item detail |
| DescribeBalance | Yes | Account balance |
| CreateBudget | Yes | Spending budget |
| DescribeBudgets | Yes | List budgets |
| DescribeReservedInstances | Yes | RI utilization |

## Command Map

| Goal | Example |
|------|---------|
| Query monthly bill | `ve billing DescribeBills --Period 2026-05` |
| Check balance | `ve billing DescribeBalance` |
| List budgets | `ve billing DescribeBudgets` |

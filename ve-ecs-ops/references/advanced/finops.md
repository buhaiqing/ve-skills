# FinOps — ECS Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Pricing Overview

ECS supports PostPaid (pay-as-you-go), PrePaid (1-year ~35% off, 3-year ~50% off), and Spot instances (60-90% off, interruptible). Right-sizing and converting steady workloads to PrePaid are the highest-leverage savings actions.

## Cost Optimization Tips

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Right-size underutilized instances | CPU avg < 15% for 7d → downsize (e.g., ecs.g3i.2xlarge → ecs.g3i.xlarge) | ~50% |
| Convert steady PostPaid to PrePaid | Running > 30d → 1yr PrePaid | ~35% |
| Use Spot for batch/fault-tolerant | Non-critical or stateless workloads | 60-90% |

## Query Current Prices

```bash
# Describe price for a specific instance type (cn-beijing, PostPaid)
ve ecs DescribeInstanceTypes --InstanceTypeIds '["ecs.g3i.large"]' | jq '.Result.InstanceTypes[] | {InstanceType, Price}'
```

> ⚠️ Pricing data sourced from ve DescribePrice command — use `ve ecs DescribePrice` for current quotes.
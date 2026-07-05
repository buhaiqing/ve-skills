# FinOps — Security Group Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Cost Overview

Security Groups are free of charge. There are no direct billing costs for security group rules or usage.

## Cost Optimization Tips

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Remove unused SGs | Orphaned security groups (not associated with any ENI) — delete | Simplifies management, eliminates audit noise |
| Consolidate rules | Merge duplicate rules across SGs — fewer SGs to manage | Operational efficiency |
| Audit regularly | Periodically review rules against least-privilege principle | Security + governance |

> Security Group usage is free — focus optimization on governance and operational hygiene.
>
> ⚠️ Pricing data sourced from ve DescribePrice command — use `ve security-group DescribePrice` for current quotes.
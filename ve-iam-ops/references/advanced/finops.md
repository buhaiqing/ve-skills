# FinOps — IAM Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Cost Overview

IAM is free of charge. There are no direct billing costs for users, roles, policies, or API calls.

## Cost Optimization Tips

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Remove unused roles/policies | Delete stale IAM roles and orphaned policies | Simplifies governance |
| Consolidate permissions | Merge duplicate inline policies into reusable managed policies | Reduces audit overhead |
| Rotate access keys regularly | Old unused keys → deactivate and delete | Security best practice |

> IAM is free — focus optimization on governance and security hygiene.
>
> ⚠️ Pricing data sourced from ve DescribePrice command — use `ve iam DescribePrice` for current quotes.
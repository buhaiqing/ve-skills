# FinOps — Private DNS Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7.

## Billing Model Comparison

Private DNS typically has flat pricing per zone. Query the billing API for current rates.

## Cost Optimization Quick Reference

| Situation | Action | Savings |
|-----------|--------|---------|
| Many unused zones | Delete stale zones | ~100% |
| Excessive query volume | Optimize record TTL | Reduce API calls |
| Multi-zone duplication | Consolidate zones | Simplify mgmt |

> Private DNS costs are minimal — focus on operational efficiency.

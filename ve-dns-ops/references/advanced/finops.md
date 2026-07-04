# FinOps — Private DNS Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7.

## Billing Model

| Component | Pricing Model | Typical Cost |
|-----------|---------------|--------------|
| Private Zone | Per zone / month | Low (flat rate) |
| DNS Queries | Per million queries | Minimal |
| Cross-region resolution | Per GB data transfer | Variable |

> Private DNS costs are typically flat and minimal — focus on operational efficiency.

## Cost Optimization Quick Reference

| Situation | Action | 💰 Savings |
|-----------|--------|---------|
| Many unused zones | Delete stale zones | 100% |
| Excessive query volume | Optimize record TTL (increase) | Reduce API calls |
| Multi-zone duplication | Consolidate into single zone | Simplify management |
| Orphaned zones (> 30 days no queries) | Delete | 100% |
| Unnecessary wildcard records | Replace with specific records | Reduce query errors |

## Operational Tips

- **TTL tuning**: Increase TTL for stable records (e.g., `3600s` → `86400s`) to reduce query volume
- **Zone consolidation**: Use one private zone per VPC with subdomain delegation instead of multiple zones
- **Resolution path**: Preferred resolution path reduces query hops by using VPC DNS directly
- **Delete before delete**: When deleting zones, verify no dependent services reference them

> Private DNS is low-cost by nature — focus optimization on zone hygiene rather than query volume.
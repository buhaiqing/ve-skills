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

- **TTL tuning**: Stable records: `3600s` → `86400s` to reduce query volume
- **Zone consolidation**: One private zone per VPC + subdomain delegation vs multiple zones
- **Resolution path**: Preferred resolution path reduces query hops (use VPC DNS directly)
- **Delete before delete**: When deleting zones, verify no dependent services reference them

## Cost Anomaly Detection

⚠️ **DNS cost anomalies** — investigate when:

| Anomaly | Investigation | Action |
|---------|---------------|--------|
| Query volume spike > 500% | `ve dns ListZones` → check TTL config for specific records | Lower TTL back if intentional, or audit DDoS |
| Cross-region transfer increase | Check newly added cross-region resolutions | Consolidate resolution path |
| New zone creation burst | Audit recent `CreateZone` events | Confirm authorized |

> **Cost note**: Private DNS is near-free. Focus on zone hygiene, not unit price.

## Related Resources

- [ve-billing-ops FinOps](../../ve-billing-ops/references/advanced/finops.md) — billing models, budget alerts, tag allocation
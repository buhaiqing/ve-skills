# FinOps — CDN Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7.

## Billing Model Comparison

| Model | Description | Best For |
|-------|-------------|----------|
| Pay-by-Traffic | Per GB pricing (tiered: lower per GB at higher volumes) | Variable traffic patterns |
| Pay-by-Bandwidth | Per Mbps/month (95th percentile peak) | Steady high traffic |
| Dynamic Route | Per request — separate from static cache | Dynamic API / real-time content |

## Cost Optimization

| Situation | Action | 💰 Savings |
|-----------|--------|---------|
| High origin traffic | Optimize cache TTL + add cache-control headers | 30-70% |
| Cache miss storm | Pre-warm popular content during off-peak | 20-40% |
| Dynamic content on CDN | Route dynamic API via Dynamic Route (not static tier) | 15-30% |
| Low-hit domains | Consolidate or prune stale domains | 10-30% |
| Multi-region delivery | Use parent group to deduplicate | Variable per zone |

## Cost Anomaly Detection

| Sign | Cause | Investigation |
|------|-------|-------------|
| ⚠️ Bandwidth spike >50% | DDoS, hot content going viral | `ve cdn DescribeCdnUpperUsage --body '{"Domains":["{{env.DOMAIN}}"]}'` |
| ⚠️ Traffic cost > budget | Cache miss ratio high | Check cache hit ratio metrics |
| ⚠️ Origin fetch surge | TTL too short or misconfigured | Review cache rules in CDN console |
| ⚠️ HTTPS request cost jump | Mixed content forcing re-handshake | Enable HTTP/2 + OCSP stapling |

## Query Pricing

```bash
# CDN pricing is bandwidth-tiered by region — query usage to estimate cost
ve cdn DescribeCdnUpperUsage --body '{"Domains":["{{output.DomainName}}"],"StartTime":"{{user.START_TIME}}","EndTime":"{{user.END_TIME}}"}'
# Use Billing API for actual cost
ve billing DescribeBillSummaryByMonth --body '{"BillingPeriod":"{{env.BILLING_MONTH}}"}'
```

## Related Resources

- [ve-billing-ops FinOps](../../ve-billing-ops/references/advanced/finops.md)
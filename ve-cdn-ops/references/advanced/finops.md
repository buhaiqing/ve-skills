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

## Operations

### Operation: AnalyzeBandwidthCost — Generate Bandwidth Cost Report

Analyzes CDN bandwidth usage and costs for optimization opportunities.

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | `test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY"` | Both set | HALT |
| Date range | Valid start/end times | Within last 90 days | Adjust range |

#### Execution

```bash
# Get bandwidth data for all domains
ve cdn DescribeCdnData \
  --Region "{{user.region}}" \
  --StartTime "{{user.start_time}}" \
  --EndTime "{{user.end_time}}" \
  --Metric "bandwidth"

# Get cache hit ratio
ve cdn DescribeCdnDomainHitRate \
  --Region "{{user.region}}" \
  --StartTime "{{user.start_time}}" \
  --EndTime "{{user.end_time}}"
```

#### Analysis Logic

| Signal | Threshold | Classification | Recommendation |
|--------|-----------|----------------|----------------|
| Cache hit rate < 50% | Poor | Optimize cache rules |
| Cache hit rate 50-70% | Fair | Review cache TTL settings |
| Cache hit rate > 90% | Excellent | Maintain current config |
| Origin bandwidth > 30% | High origin dependency | Increase cache TTL |
| Peak bandwidth variance > 50% | Unpredictable traffic | Enable rate limiting |

#### Output Format

```markdown
## CDN Bandwidth Analysis — [Date Range]

| Domain | Bandwidth (Peak) | Traffic (Total) | Cache Hit Rate | Origin Pull | Status |
|--------|-----------------|-----------------|----------------|-------------|--------|
| cdn.example.com | 1.2 Gbps | 15 TB | 92% | 8% | ✅ Excellent |
| cdn2.example.com | 800 Mbps | 8 TB | 65% | 35% | ⚠️ Needs Optimization |

### Optimization Opportunities
- cdn2.example.com: Increase cache TTL for static assets (potential 20% cost reduction)
- All domains: Enable Brotli compression (potential 15% bandwidth savings)

### Estimated Monthly Savings
- Cache optimization: ¥2,400/month
- Compression: ¥1,800/month
- **Total potential savings: ¥4,200/month**
```

---

## Related Resources

- [ve-billing-ops FinOps](../../ve-billing-ops/references/advanced/finops.md)

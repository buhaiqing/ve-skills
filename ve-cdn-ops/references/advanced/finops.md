# FinOps — CDN Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7.

## Billing Model Comparison

| Model | Description | Best For |
|-------|-------------|----------|
| Pay-by-Traffic | Per GB pricing | Variable traffic patterns |
| Pay-by-Bandwidth | Per Mbps/month | Steady high traffic |
| 95th Percentile | Based on 95th bandwidth | Predictable peaks |

## Cost Optimization Quick Reference

| Situation | Action | Savings |
|-----------|--------|---------|
| High origin traffic | Optimize cache TTL | 30-70% |
| Cache miss storm | Pre-warm popular content | 20-40% |
| Unused domains | Clean up stale configs | 10-30% |
| HTTPS over-provisioning | Right-size bandwidth | 5-15% |

> Always query the CDN pricing API — never hardcode rates.

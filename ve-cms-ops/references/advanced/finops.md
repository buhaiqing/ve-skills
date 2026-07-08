# FinOps — Cloud Monitor Cost Optimization

> CMS has free tier + paid custom metrics/API. FinOps = granularity tuning + alarm consolidation.

## Cost Model
| Component | Pricing | Mitigation |
|-----------|---------|------------|
| Built-in metrics | **Free** | No action needed |
| Custom metrics | Per metric/month × granularity | Reduce resolution |
| API calls | Per 1k calls | Batch, reduce polling |
| Alarm rules | Per rule/month | Consolidate |

## Cost Optimization
| Pattern | Action | 💰 Savings |
|---------|--------|---------|
| ⏱️ Lower granularity | 5-min for non-critical, 1-min for prod | ~80% metric cost |
| 🧹 Stale alarms | Delete alarms for deleted resources | 100% idle alarm cost |
| 📊 Batch custom metrics | Aggregate counters → gauge with stats dims | 50-70% metric count |
| 🔄 Reduce polling | 15s → 60s for dashboards | 75% API call reduction |
| 🎯 Alarm consolidation | One composite vs. 5 per-instance alarms | 80% alarm cost |

## Cost Anomaly Detection
| Signal | What to Check | Response |
|--------|---------------|----------|
| ⚠️ Custom metric count spike | New metrics without cost review | Stop unused, aggregate |
| ⚠️ API call surge > 3x | External dashboards polling too frequently | Increase interval, cache |
| ⚠️ Alarm rule explosion | > 50 per service | Consolidate, use composite |
| 🚨 Metric retention > 90d | Long-retained data incurring storage cost | Set TTL, export to TOS |

## Pricing

```bash
ve billing DescribeBillSummaryByMonth --body '{"BillingPeriod":"{{env.BILLING_MONTH}}"}'
```

> 💡 No `ve cms DescribePrice`. Pricing varies by metric count & granularity — use billing or console.

## Related Resources

- [Billing FinOps → budget alerts & cost allocation](../../../ve-billing-ops/references/advanced/finops.md)
- [CMS AIOps → anomaly detection & smart alerting](../advanced/aiops.md)
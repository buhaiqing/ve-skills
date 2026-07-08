# FinOps — KMS Cost Optimization

> KMS has measurable costs: key storage + API usage. FinOps = lifecycle management + tier selection.

## Cost Model
| Component | Pricing | Mitigation |
|-----------|---------|------------|
| Key storage | Per key/month | Delete unused keys |
| API calls | Per 10k calls | Cache decrypted material |
| HSM-backed keys | ~10x software key cost | Software unless compliance mandates |
| Auto rotation | Free | Enable for long-lived keys |

## Cost Optimization
| Pattern | Action | 💰 Savings |
|---------|--------|---------|
| 🗑️ Orphaned keys | Schedule deletion, zero resource refs | 100% key storage |
| 🔄 Key consolidation | One key + aliases vs. multiple keys | Reduces key count |
| 📦 Cache & reuse | Cache decrypted (TTL ≤ security policy) | 60-90% API reduction |
| 🏗️ Software > HSM | Default to software; HSM only for regulatory req | ~90% key cost |
| 🔁 Auto-rotation | Auto-rotate vs. manual re-create (auto cheaper) | Lower ops cost |

## Cost Anomaly Detection
| Signal | What to Check | Response |
|--------|---------------|----------|
| ⚠️ Key creation spike | New keys without retirement plan | Review & consolidate |
| ⚠️ API call surge > 2x | Unexpected re-decryption | Check caching config |
| ⚠️ HSM key ratio ↑ | Team defaulting to HSM | Re-evaluate compliance need |
| 🚨 PendingDeletion > 30d | Keys stuck, not cleaned up | Force delete or cancel |

## Pricing

```bash
ve billing DescribeBillSummaryByMonth --body '{"BillingPeriod":"{{env.BILLING_MONTH}}"}'
```

> 💡 No `ve kms DescribePrice` exists. Query via billing or console pricing page.

## Related Resources

- [Billing FinOps → budget alerts & cost allocation](../../../ve-billing-ops/references/advanced/finops.md)
- [KMS SecurityOps → key lifecycle & rotation](../advanced/securityops.md)
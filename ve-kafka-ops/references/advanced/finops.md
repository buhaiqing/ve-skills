# FinOps — Kafka Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7.

## Billing Model Comparison

| Component | Factor | Optimization |
|-----------|--------|-------------|
| Broker instances | Number × spec type | Right-size broker spec (2-core vs 4-core vs 8-core) |
| Storage | Per GB/month on premium SSD | Set topic retention policy |
| Traffic | Cross-zone transfer at ~¥0.80/GB | Co-locate clients and brokers in same AZ |
| Public bandwidth | Per Mbps egress | Use internal endpoints where possible |

## Cost Optimization

| Situation | Action | 💰 Savings |
|-----------|--------|---------|
| Long retention topics | Reduce `retention.ms` from 7d → 2d | 30-70% storage cost |
| Over-partitioned topic | Merge small partitions (target ≤5K partitions/broker) | Broker CPU/memory ↓ |
| Consumer group lag growing | Add consumers or clean up stale groups | Prevents backlog cost |
| Uncompressed messages | Enable `compression.type=snappy` | 30-60% storage/network |
| Max idle topics | Auto-delete with 7d retention + cleanup.policy=delete | ~100% reclaim |

## Cost Anomaly Detection

| Sign | Cause | Investigation |
|------|-------|-------------|
| ⚠️ Storage usage doubling weekly | Retention too long or no cleanup | `ve kafka DescribeTopicParameters --topic {{output.TopicName}}` |
| ⚠️ Broker CPU >80% sustained | Too many partitions per broker | Check partition distribution across brokers |
| ⚠️ Cross-zone traffic sudden increase | Consumers moved to different AZ | Verify consumer instance AZ placement |
| ⚠️ Message size spike | Uncompressed payloads, large messages | Check `max.message.bytes` config |

## Query Pricing

```bash
# Query instance spec pricing (spec = instance type × broker count)
ve kafka DescribeInstancePrice --body '{"Region":"{{env.REGION}}","InstanceId":"{{output.InstanceId}}"}'
# Check current monthly cost
ve billing DescribeBillSummaryByMonth --body '{"BillingPeriod":"{{env.BILLING_MONTH}}"}'
```

## Related Resources

- [ve-billing-ops FinOps](../../ve-billing-ops/references/advanced/finops.md)
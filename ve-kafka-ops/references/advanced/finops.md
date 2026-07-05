# FinOps — Kafka Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7.

## Billing Model Comparison

| Component | Factor | Optimization |
|-----------|--------|-------------|
| Broker instances | Number × type | Right-size brokers |
| Storage | Per GB/month | Set retention policy |
| Traffic | Cross-zone transfer | Place brokers in same zone |

## Cost Optimization Quick Reference

| Situation | Action | Savings |
|-----------|--------|---------|
| Long retention | Reduce topic retention | 30-70% storage |
| Over-partitioned | Optimize partition count | Broker CPU ↓ |
| Unused topics | Delete stale topics | ~100% |
| Cross-zone traffic | Co-locate clients and brokers | Up to 100% |

> ⚠️ Pricing data sourced from ve DescribePrice command — use `ve kafka DescribePrice` for current quotes.

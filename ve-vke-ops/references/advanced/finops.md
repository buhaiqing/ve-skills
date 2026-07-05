# FinOps — VKE Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Cost Optimization Tips

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Right-size node pools | CPU avg < 20% → use smaller instance type or fewer nodes | ~50% |
| Use Spot for worker nodes | Fault-tolerant workloads → Spot node pool for non-critical pods | 60-90% |
| Auto-scale node pools | Configure Cluster Autoscaler to scale down idle nodes | Up to 100% of idle node cost |

## Query Current Prices

```bash
# Query current VKE node pool pricing
ve vke DescribePrice
```

> ⚠️ Pricing data sourced from ve DescribePrice command — use `ve vke DescribePrice` for current quotes.
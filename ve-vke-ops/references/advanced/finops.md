# FinOps — VKE Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7.

## Billing Model

VKE control plane is free → **cost = ECS worker nodes**.

| Component | Pricing Model | % of Total |
|-----------|---------------|-----------|
| Worker node (PostPaid) | Per second / hour | 100% (dominant) |
| Worker node (PrePaid) | Monthly / yearly | 100% (discounted) |
| Load balancer (CLB/ALB) | Per hour + data | Variable |
| Persistent volume | Per GB / month | Minor |

## Cost Optimization Tips

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Right-size node pools | CPU avg < 20% → smaller instance type or fewer nodes | ~50% |
| Use Spot for batch workloads | Spot node pool for non-critical / fault-tolerant pods | 60-90% |
| Auto-scale node pools | Cluster Autoscaler scales down idle nodes | Up to 100% idle cost |
| Over-provisioning | `Unschedulable` pods → check node utilization; reduce if < 30% | 20-40% |

## Cost Anomaly Detection

⚠️ **VKE cost anomalies** — investigate when:
| Anomaly | Investigation | Action |
|---------|---------------|--------|
| Node pool count spike | `ve vke ListNodePools` → check unscheduled scale-up events | Audit HPA / CA config |
| Spot instance price spike | `ve ecs DescribeSpotAdvice` → check market price trends | Fall back to PostPaid |
| PersistentVolume count surge | `ve vke ListStorageClasses` → review PVC usage | Consolidate storage |
| Cluster node count > 100% expected | Compare to HPA max replicas config | Tune CA limits |

## Query Current Resources

```bash
ve vke ListNodePools --body '{}'; ve ecs DescribeSpotAdvice --body '{"InstanceTypeIds":["ecs.g3i.large"]}'
```
> 💡 VKE cost = ECS node cost. Isolate via `ve billing DescribeInstanceBill --body '{"Product":"VKE"}'`.

## Related Resources

- [ve-billing-ops FinOps](../../ve-billing-ops/references/advanced/finops.md) — billing models, budget alerts, tag allocation
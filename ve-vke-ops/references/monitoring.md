# VKE Multi-Metric Correlation — AIOps

## Anomaly Patterns

| Pattern | Metrics | Detection Logic | Severity |
|---------|---------|-----------------|----------|
| Control Plane Pressure | API Server Latency > 1s + etcd latency > 100ms + CPU master > 80% | All AND over 5min | Critical |
| Node Exhaustion | CPU node > 90% + Memory node > 85% + Pod density > 95% | All AND sustained 10min | Critical |
| VPC-CNI Network Degradation | Pod IP exhaustion + Failed pod assignments > 5/min + DNS failures | Concurrent thresholds | Warning |
| Node Pool Scaling Failure | Auto-scaler triggers + Available ENIs = 0 + ECS quota reached | Cascade detection | Critical |
| Persistent Volume Pressure | Disk node > 85% + etcd disk > 70% + IOPS queue > 4 | Sustained 10min | Warning |
| Cluster Health Degradation | NotReady > 10% + CrashLoopBackOff > 5 + API 5xx > 1% | All AND over 15min | Critical |

## Cross-Skill Diagnosis Decision Tree

```
[Alarm: VKE cluster issue]
    │
    ├── Is the API server affected?
    │   ├── High latency → Check etcd disk I/O → ve-ecs-ops (etcd node disk)
    │   ├── 5xx errors → Check node health → Isolate failing nodes
    │   └── Unresponsive → Check control plane nodes → ve-ecs-ops
    │
    ├── Are nodes NotReady?
    │   ├── Single node → ECS issue → ve-ecs-ops (CPU/Memory/disk)
    │   ├── Multiple same zone → Zone outage → ve-cms-ops
    │   ├── All nodes → Network issue → ve-vpc-ops (VPC route)
    │   └── New nodes can't join → Security group / NodePool config
    │
    ├── Are pods failing?
    │   ├── OOMKilled → Application → Check JVM/memory
    │   ├── CrashLoopBackOff → Check logs → ve-loki-ops
    │   ├── Pending → NodePool capacity → Scale nodes
    │   └── ImagePullBackOff → Registry access
    │
    └── Unknown → ve-cms-ops for correlation analysis
```

## Delegation Matrix

| Alarm Signal | Primary Skill | Secondary Skill | Fallback |
|-------------|--------------|-----------------|----------|
| Node CPU/Memory > 90% | ve-ecs-ops | ve-cms-ops | Manual |
| VPC-CNI IP exhaustion | ve-vpc-ops | ve-ecs-ops | Manual |
| etcd disk I/O bottleneck | ve-ecs-ops | ve-cms-ops | Manual |
| Pod OOM / CrashLoop | Application team | ve-loki-ops (logs) | ve-cms-ops |
| Cluster-wide outage | ve-cms-ops | ve-vpc-ops | Manual |
| NodePool can't scale | ve-ecs-ops (quota) | ve-cms-ops | Manual |

## Proactive Inspection Checklist

### Resource Health
- [ ] Cluster status = Running across all clusters
- [ ] All nodes Ready (NotReady = 0)
- [ ] Node pool actual ≥ MinReplicas
- [ ] Pod density per node < 80% of limit
- [ ] VPC-CNI IP pool > 20% available
- [ ] etcd disk usage < 70%

### Security
- [ ] No public API access without IP whitelist
- [ ] No ClusterRoleBinding to default service account
- [ ] Node pools use dedicated security groups

### Cost
- [ ] No zombie clusters (CPU < 5% for 14 days)
- [ ] Node pools with desired=0 for > 7 days
- [ ] Orphaned volumes from deleted node pools
- [ ] Over-provisioned nodes (CPU avg < 15% for 7 days)

## Alarm Storm Handling

**Detection:** > 10 pod alarms within 5 min from same cluster, or > 50% share one root metric (e.g., node CPU).

**Suppression:**
1. Correlate by cluster + time window
2. Identify root alarm (earliest node NotReady / highest severity)
3. Group: pods on same node → single "node degraded" alarm
4. Throttle: 1 notification per root cause per 5 min
5. Diagnose root cause
6. After fix, verify grouped alarms clear

## Multi-Round Diagnosis Review

1. **Fact Check**: Are node/pod metrics current? Valid region?
2. **Causal Analysis**: Is the identified process the true root?
3. **Solution Validation**: Will scaling/restart resolve? Side effects?

# AIOps — VKE Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[VKE Alarm Triggered]
    │
    ├── Is it node-related?
    │   ├── Node NotReady → Check kubelet status
    │   │   └── Drain node → cordon → repair or replace
    │   ├── Node OOM → Check pod resource limits
    │   │   └── Adjust limits or add nodes
    │   └── Disk pressure → Check disk usage
    │       └── Clean up /var/lib/docker or /var/lib/kubelet
    │
    ├── Is it workload-related?
    │   ├── Pod CrashLoopBackOff → Check logs
    │   │   └── Fix image, config, or resource issues
    │   ├── Pod Evicted → Check resource pressure
    │   │   └── Add nodes or adjust resource requests
    │   └── Deployment stalled → Check readiness
    │       └── Verify health checks and dependencies
    │
    ├── Is it network-related?
    │   ├── DNS resolution fails → Check CoreDNS
    │   │   └── Restart CoreDNS or check network policy
    │   ├── Ingress not working → Check ingress controller
    │   │   └── Verify controller logs
    │   └── Service unreachable → Check endpoints
    │       └── Delegate to ve-vpc-ops for network
    │
    └── Unknown → Delegate to ve-cms-ops
```

## Alarm Storm Handling

**Detection Criteria:**
- > 5 pod failures within 5 minutes
- > 20% nodes NotReady
- Node-level cascade: disk pressure → OOM → NotReady → pod eviction

**Suppression Workflow:**
1. Drain affected nodes (kubectl drain --ignore-daemonsets)
2. Add replacement nodes if needed
3. Investigate root cause (hardware, config, resource)
4. Verify pod scheduling recovers
5. Monitor cluster stability

## Proactive Inspection Checklist

```markdown
## VKE Proactive Inspection — [Date]

### Node Health
- [ ] All nodes Ready
- [ ] No disk pressure on any node
- [ ] No memory pressure on any node
- [ ] Kubelet logs healthy

### Workload
- [ ] No CrashLoopBackOff pods
- [ ] All Deployments have desired replicas
- [ ] No PodDisruptionBudget violations

### Network
- [ ] CoreDNS healthy
- [ ] Ingress controllers healthy
- [ ] Network policies not blocking required traffic

### Capacity
- [ ] Node allocatable headroom > 20%
- [ ] No pods pending scheduling
- [ ] Cluster autoscaler configured if needed
```

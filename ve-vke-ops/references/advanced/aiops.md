# AIOps — VKE Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[VKE Alarm Triggered]
    │
    ├── Is it node-related?
    │   ├── Node NotReady (kubelet heartbeat miss > 40s) → Check kubelet status
    │   │   ├── Node allocatable CPU < 10% remaining → Resource overcommit
    │   │   ├── Node allocatable memory < 10% remaining → Pod memory pressure
    │   │   └── Drain node → cordon → repair or replace
    │   ├── Node OOM (Memory pressure > 85%) → Check pod resource limits
    │   │   └── Top-3 memory consumer pods → Adjust limits or add nodes
    │   ├── Disk pressure (kubelet eviction threshold > 85%) → Check disk usage
    │   │   ├── ImagePull failures > 3/min due to disk → Clean up unused images
    │   │   └── Clean up /var/lib/docker or /var/lib/kubelet
    │   ├── Node CPU usage > 85% sustained 5min → Check resource-heavy pods
    │   │   └── Add nodes or redistribute workloads
    │   └── Node memory usage > 90% → Check for memory leaks
    │       └── Restart containers or set memory limits
    │
    ├── Is it workload-related?
    │   ├── Pod CrashLoopBackOff (restart count > 5/min) → Check logs
    │   │   ├── OOMKilled → Increase memory limit or fix memory leak
    │   │   └── Fix image, config, or resource issues
    │   ├── Pod Evicted (pending > 10 pods) → Resource pressure
    │   │   ├── NodePool capacity exhausted → Scale out node pool
    │   │   └── Add nodes or adjust resource requests
    │   ├── Deployment stalled (progress deadline > 5min) → Check readiness probe
    │   │   ├── Readiness probe failure rate > 50% → Application not ready
    │   │   └── Verify health checks and dependencies
    │   ├── Pod restart count > 5/min → Check liveness probe or OOM
    │   │   └── Increase resources or fix probe logic
    │   └── Container OOMKilled > 3/hour → Review memory limits
    │       └── Increase memory request/limit or profile memory usage
    │
    ├── Is it network-related?
    │   ├── DNS resolution fails (CoreDNS 5xx > 1%) → Check CoreDNS
    │   │   ├── CoreDNS OOM / CPU throttled → Increase CoreDNS resources
    │   │   └── Restart CoreDNS or check network policy
    │   ├── VPC-CNI IP pool < 20% available → Pod IP exhaustion
    │   │   └── Increase subnet CIDR or enable secondary CIDR
    │   ├── Ingress not working (5xx response > 5%) → Check ingress controller
    │   │   ├── Ingress controller OOM → Increase resources
    │   │   └── Verify controller logs
    │   ├── Service unreachable → Check endpoints
    │   │   └── EndpointSlice inconsistency → Restart kube-proxy
    │   ├── CoreDNS lookup failure rate > 1% → Scale CoreDNS pods
    │   │   └── Increase coreDNS replicas or adjust autoscaler
    │   └── Service endpoint not ready > 2min → Check pod readiness
    │       └── Investigate readiness probe failures
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

# VKE Knowledge Base — Fault Patterns

## Pattern 1: Node NotReady — kubelet unresponsive

### Symptoms
- Node status transitions to NotReady
- Pods on node stuck in Unknown/Terminating
- API server reports node heartbeat timeout

### Root Cause
- kubelet process crashed (OOM, disk full, config error)
- Node network partition from API server
- etcd latency causing NodeStatus update timeout

### Resolution
1. SSH to node or check via ECS console
2. `systemctl status kubelet` — restart if dead
3. Check disk: `df -h` — /var/lib/kubelet and container runtime
4. Check kubelet logs: `journalctl -u kubelet --since "10 min ago"`
5. If disk full, clean old container images and logs
6. If etcd latency, check control plane nodes

### Prevention
- Node disk monitoring > 80% alert
- kubelet systemd restart=always
- Container log rotation configured

---

## Pattern 2: Pod Stuck Pending — Insufficient Cluster Resources

### Symptoms
- Pods in Pending state with event "Insufficient cpu" or "Insufficient memory"
- Horizontal Pod Autoscaler not triggering or unable to scale
- Cluster CPU/Memory utilization > 95%

### Root Cause
- NodePool resources exhausted (all nodes fully allocated)
- Auto-scaling disabled or max replicas reached
- Pod resource requests overcommitted (request >> actual usage)

### Resolution
1. Check node pool: actual vs max replicas
2. Check auto-scaling: `ve vke DescribeNodePool --NodePoolId <id>`
3. Scale node pool: increase MaxReplicas or manually add nodes
4. Review pod resource requests — reduce if overcommitted
5. Check ECS quota — may need quota increase

### Prevention
- Auto-scaling enabled on all production node pools
- Headroom: max capacity = 150% of peak load
- Periodic resource right-sizing review

---

## Pattern 3: VPC-CNI Pod IP Exhaustion

### Symptoms
- New pods stuck in Pending with "failed to allocate IP"
- Node has capacity (CPU/memory OK) but can't schedule
- Subnet IP usage > 90%

### Root Cause
- VPC-CNI pod subnet depleted — too many pods or subnet too small
- IPs not reclaimed after pod deletion (stale ENI attachments)

### Resolution
1. Check subnet utilization: `ve vpc DescribeSubnetAttributes`
2. Expand VPC-CNI subnet range or add additional subnets
3. Clean stale ENI attachments if present
4. Update PodConfig with additional subnet IDs
5. Cluster update: `ve vke UpdateClusterConfig`

### Prevention
- /24 subnet minimum for small clusters, /22 for production
- Monitor subnet IP usage with > 70% warning
- Use VpcCniShared mode to reduce per-pod IP consumption

---

## Cascade Pattern: etcd Disk Full → API Server Down → All Pods Degraded

### Trigger Event
- etcd node disk reaches 100% (usually /var/lib/etcd)

### Propagation Path
- **A → etcd compaction fails** (no disk for writes)
- **B → kube-apiserver cannot write/watch** → API responses timeout
- **C → kubelet loses API connection** → Nodes go NotReady
- **D → All pod operations blocked** → no scheduling, no health checks, no restarts

### Breaking the Chain
1. **Critical fix at A**: Clean etcd disk immediately
   ```bash
   # SSH to etcd master node
   du -sh /var/lib/etcd/* | sort -rh | head -20
   # Compact etcd history if possible
   etcdctl defrag
   ```
2. Remove old snapshots or reduce snapshot retention
3. Add monitoring: etcd disk > 70% → Critical alert

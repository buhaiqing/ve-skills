# Knowledge Base — ECS Fault Patterns

## Pattern 1: Instance Unresponsive (CPU High but No Network Response)

### Symptoms
- Instance status shows `RUNNING`
- CPU utilization > 90% sustained
- SSH/RDP connections timeout
- NetworkInPps = 0 or very low
- Security group allows inbound traffic

### Root Cause
- Runaway process consuming all CPU cycles
- Kernel panic or OOM killer silent failure
- Disk I/O starvation causing system hang

### Resolution
1. Try Cloud Assistant command: `ve ecs InvokeCommand --InstanceId i-xxx --CommandContent "top -bn1"`
2. If Cloud Assistant fails, hard reboot: `ve ecs RebootInstance --InstanceId i-xxx --ForceStop true`
3. If still unresponsive, check system log: `ve ecs GetInstanceConsoleOutput --InstanceId i-xxx`
4. As last resort: stop → create snapshot → delete → recreate from snapshot

### Prevention
- Set CPU alert threshold at 85% for 5min
- Enable Cloud Assistant on all instances
- Configure OOM killer notifications

---

## Pattern 2: Disk Full → Service Degradation Cascade

### Symptoms
- Disk usage > 95%
- Application error logs: "No space left on device"
- Database connection failures
- Service health checks failing

### Root Cause
- Log files growing unbounded
- Temporary files not cleaned
- Application writing excessive data to root disk

### Resolution
1. Identify large files: Cloud Assistant `du -sh /* | sort -rh | head -20`
2. Clean log files: `find /var/log -name "*.log" -mtime +7 -delete`
3. Clean temp files: `rm -rf /tmp/* /var/tmp/*`
4. If root disk too small: create snapshot → create larger disk → replace
5. Set up log rotation if not configured

### Prevention
- Monitor disk usage at 80% warning threshold
- Configure log rotation (logrotate)
- Use separate data disk for application data
- Set up automatic cleanup cron jobs

---

## Pattern 3: Instance Type Change Failure

### Symptoms
- `ModifyInstanceSpec` returns error
- Instance stuck in `STOPPED` state after spec change
- Error: `InvalidInstanceType.ValueNotSupported` or `ResourceNotEnough`

### Root Cause
- Target instance type not available in current zone
- Instance has local disks (cannot change type)
- Incompatible instance family conversion

### Resolution
1. Check available types: `ve ecs DescribeInstanceTypes --ZoneId {{user.zone_id}}`
2. If local disk: must recreate instance (cannot migrate)
3. If zone unavailable: try different zone or wait
4. If family incompatible: choose within same family or verify compatibility matrix

### Prevention
- Always verify target type availability before stopping instance
- Document which instances have local disks
- Test instance type changes in non-production first

---

## Pattern 4: Spot Instance Reclamation

### Symptoms
- Spot instance terminated unexpectedly
- Error: `SpotInstanceInterruption` or `InsufficientPoolCapacity`
- Application downtime

### Root Cause
- Spot price exceeds your maximum bid
- Capacity returned to on-demand customers
- Resource pool rebalancing

### Resolution
1. Check interruption reason: `ve ecs DescribeSpotPriceHistory`
2. For stateless workloads: let auto-scaling replace
3. For stateful workloads: migrate to PostPaid or increase bid price
4. Implement graceful shutdown handler for spot interruption notices

### Prevention
- Use spot instances only for fault-tolerant workloads
- Set bid price at least 2x current spot price
- Monitor spot price trends before choosing instance types
- Implement checkpoint/resume for batch workloads

---

## Cascade Pattern: Network Issue → Application Failure → Alert Storm

### Trigger Event
- Security group rule change blocking inbound traffic
- VPC route table modification
- ENI IP address conflict

### Propagation Path
- A (Network change) → B (Instance unreachable) → C (Health check fails) → D (Load balancer removes instance) → E (Auto-scaling launches new instances) → F (Multiple alarms fire simultaneously)

### Breaking the Chain
- **Primary break point:** Verify security group before any network change
- **Secondary break point:** Configure health check grace period (5-10 min) before alarm triggers
- **Tertiary break point:** Use alarm storm suppression — correlate by resource group

### Resolution
1. Stop the cascade: Disable auto-scaling temporarily
2. Fix the root cause: Restore correct security group / route table
3. Verify recovery: Check instance health and connectivity
4. Re-enable auto-scaling
5. Suppress duplicate alarms: Group all related alarms under root cause

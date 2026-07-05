# AIOps — ECS Intelligent Operations

> AIOps deep content extracted per TE-7. Monitoring.md links here.

## Cross-Skill Diagnosis Decision Tree

```
[ECS Alarm Triggered]
    │
    ├── Is it CPU-related?
    │   ├── CPU > 90% sustained for 5min → Is memory also high?
    │   │   ├── CPU > 90% + Memory > 85% → Application-level issue (check logs via Cloud Assistant, GC, threads)
    │   │   │       └── Delegate to application skill if Java/JVM
    │   │   ├── CPU > 90% + Memory normal → Check for runaway processes (Cloud Assistant: top, ps aux)
    │   │   │       └── If single process: kill or restart
    │   │   │       └── If system-wide: reboot
    │   │   └── CPU credit balance < 20% (burstable instances) → Change to compute-optimized instance type
    │   └── CPU < 5% for 7 days → Idle instance — consider stop or downsizing
    │
    ├── Is it disk-related?
    │   ├── Disk usage > 90% → Check log files, temporary files (Cloud Assistant)
    │   │   ├── Disk read latency > 20ms → Check disk type (ESSD vs basic)
    │   │   ├── If log-related: set up rotation → delegate to app team
    │   │   └── If data-related: expand disk or add data disk
    │   └── Disk IOPS > 90% of max throughput → Check database queries, backup jobs
    │       ├── Inode usage > 95% → Too many small files — cleanup or reformat
    │       └── If database: delegate to ve-rds-ops
    │       └── If backup: reschedule to off-peak
    │
    ├── Is it network-related?
    │   ├── Network outbound throughput > 90% of bandwidth limit → Check instance type bandwidth cap
    │   │   └── Upgrade instance type or throttle traffic
    │   ├── Packet loss > 1% → Check security groups, ACLs
    │   │   └── Recent SG change? Rollback → delegate to security team
    │   ├── Connection tracking table > 80% → Too many concurrent connections
    │   │   └── Check application connection pool → Delegate to application skill
    │   └── Latency spike > 200ms → Check upstream dependencies
    │       └── Delegate to ve-vpc-ops for network path analysis
    │
    └── Unknown pattern → Delegate to ve-cms-ops for correlation analysis
```

## Cross-Skill Routing

| Symptom | Delegate To |
|---------|------------|
| IAM permission denied for instance operation | ve-iam-ops |
| KMS key unavailable for encrypted disk attach | ve-kms-ops |
| EIP bandwidth limit exceeded | ve-eip-ops |
| VPC subnet exhausted / route table misconfigured | ve-vpc-ops |
| Alarm correlation across instance group | ve-cms-ops |

## Alarm Storm Handling

**Detection Criteria:**
- > 10 alarms within 5 minutes from same ECS resource group
- > 50% of alarms share the same root cause metric
- Alarm rate exceeds 3x the baseline rate

**Suppression Workflow:**
1. Correlate alarms by instance ID and time window
2. Identify root alarm (earliest or highest severity)
3. Group related alarms under root alarm
4. Notify once per root cause (not per alarm)
5. Execute root cause diagnosis on primary alarm
6. After resolution, verify all related alarms clear

## Proactive Inspection Checklist

```markdown
## ECS Proactive Inspection — [Date]

### Resource Health
- [ ] CPU usage < 70% across all instances (avg over 7 days)
- [ ] Memory usage < 80% across all instances (avg over 7 days)
- [ ] Disk usage < 80% with > 50GB free space
- [ ] Network error rate < 0.1%
- [ ] No instances in ERROR state

### Cost Optimization
- [ ] No idle instances (CPU < 5% for 7 days)
- [ ] No stopped instances > 7 days without planned restart
- [ ] No unattached cloud disks
- [ ] No snapshots older than 90 days without retention policy
- [ ] Reserved instance coverage > 60% for steady workloads

### Security Posture
- [ ] No instances with public IP unless explicitly required
- [ ] Security group rules follow least privilege (no 0.0.0.0/0 on non-HTTP ports)
- [ ] No instances without Cloud Assistant installed
- [ ] Deletion protection enabled for production instances

### Reliability
- [ ] Multi-AZ deployment for production workloads
- [ ] Automated backups configured for instances with data disks
- [ ] Health checks configured for load-balanced instances
```

## Multi-Round Diagnosis Review

Before finalizing any ECS diagnosis:

1. **Fact Check:** Are the ECS metrics current? Are thresholds correct?
2. **Causal Analysis:** Is the identified cause the true root cause? Could something else explain the symptoms?
3. **Solution Validation:** Will the fix actually resolve the issue? Could it cause side effects?
# AIOps — EIP Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[EIP Alarm Triggered]
    │
    ├── Is it connectivity-related?
    │   ├── Instance unreachable → Check EIP binding
    │   │   └── Verify EIP associated to correct instance/NAT
    │   ├── EIP bandwidth saturation > 95% → Check traffic spikes
    │   │   └── Upgrade bandwidth or enable traffic shaping
    │   ├── EIP health check failure > 3 consecutive → Investigate target health
    │   │   └── Delegate to ve-security-group-ops or ve-ecs-ops
    │   └── SNAT port exhaustion on EIP → Check concurrent connections
    │       └── Add more EIPs or switch to NAT gateway
    │
    ├── Is it bandwidth-related?
    │   ├── Bandwidth at limit → Optimize or upgrade
    │   │   └── Review traffic patterns → identify optimization
    │   ├── High packet loss → Check network quality
    │   │   └── Ping/traceroute diagnostics
    │   ├── Inbound traffic > outbound traffic by 5x → Asymmetric routing
    │   │   └── Verify route table symmetry
    │   └── Burst credit exhausted → Continuous high traffic
    │       └── Upgrade to higher bandwidth tier
    │
    ├── Is it configuration-related?
    │   ├── Recent change → Review change log
    │   │   └── Rollback if needed → verify recovery
    │   ├── ACL/firewall rule issue → Review rules
    │   │   └── Delegate to ve-security-group-ops
    │   ├── EIP unbilled / idle > 7 days → Release or associate
    │   │   └── Audit EIP inventory monthly
    │   └── EIP not associated to running instance → Check dangling IPs
    │       └── Delegate to ve-ecs-ops for instance audit
    │
    └── Unknown pattern → Delegate to ve-cms-ops
```

## Alarm Storm Handling

**Detection Criteria:**
- > 10 alarms within 5 minutes from same network resource
- > 50% alarms share same root cause
- Cascade of network → compute alarms

**Suppression Workflow:**
1. Stop cascade: disable dependent auto-scaling temporarily
2. Fix root cause at network layer
3. Verify compute layer recovers
4. Re-enable dependent services
5. Confirm all alarms clear

## Proactive Inspection Checklist

```markdown
## EIP Proactive Inspection — [Date]

### Connectivity
- [ ] All endpoints reachable via expected paths
- [ ] No asymmetric routing detected
- [ ] Health checks configured and passing

### Configuration
- [ ] No overly permissive rules (0.0.0.0/0)
- [ ] Changes reviewed and documented
- [ ] Rollback plan in place for each change

### Performance
- [ ] Bandwidth utilization < 70%
- [ ] No packet loss > 0.1%
- [ ] Latency within SLA
```

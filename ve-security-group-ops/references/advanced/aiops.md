# AIOps — Security Group Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[Security Group Alarm Triggered]
    │
    ├── Is it connectivity-related?
    │   ├── Instance unreachable → Check security group rules
    │   │   └── Delegate to ve-security-group-ops if SG change needed
    │   ├── Route issue → Check VPC route tables
    │   │   └── Delegate to ve-vpc-ops for route diagnosis
    │   └── NAT not working → Check SNAT/DNAT rules
    │       └── Review NAT gateway configuration
    │
    ├── Is it bandwidth-related?
    │   ├── Bandwidth at limit → Optimize or upgrade
    │   │   └── Review traffic patterns → identify optimization
    │   └── High packet loss → Check network quality
    │       └── Ping/traceroute diagnostics
    │
    ├── Is it configuration-related?
    │   ├── Recent change → Review change log
    │   │   └── Rollback if needed → verify recovery
    │   └── ACL/firewall rule issue → Review rules
    │       └── Delegate to ve-security-group-ops
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
## Security Group Proactive Inspection — [Date]

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

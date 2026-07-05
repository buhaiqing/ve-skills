# AIOps — NAT Gateway Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[NAT Gateway Alarm Triggered]
    │
    ├── Is it connectivity-related?
    │   ├── Instance unreachable → Check NAT gateway route
    │   │   └── Verify VPC route table points to NAT gateway
    │   ├── SNAT port exhaustion > 80% → Check concurrent connection count
    │   │   └── Add more EIPs or enable SNAT IP pool
    │   ├── DNAT rule conflict detected → Review port forwarding rules
    │   │   └── Remove duplicate/conflicting DNAT entries
    │   └── DNAT target unhealthy → Check target instance state
    │       └── Delegate to ve-ecs-ops or ve-security-group-ops
    │
    ├── Is it bandwidth-related?
    │   ├── Bandwidth at limit → Optimize or upgrade
    │   │   └── Review traffic patterns → identify optimization
    │   ├── High packet loss → Check network quality
    │   │   └── Ping/traceroute diagnostics
    │   ├── NAT gateway throughput > 80% max → Upgrade NAT spec
    │   │   └── Choose higher spec NAT gateway or split subnets
    │   └── Concurrent connections > 50000 → Connection tracking pressure
    │       └── Increase NAT gateway spec or optimize connection reuse
    │
    ├── Is it configuration-related?
    │   ├── Recent change → Review change log
    │   │   └── Rollback if needed → verify recovery
    │   ├── ACL/firewall rule issue → Review rules
    │   │   └── Delegate to ve-security-group-ops
    │   ├── SNAT entry timeout misconfigured → Verify timeout settings
    │   │   └── Adjust TCP/UDP session timeout for workload pattern
    │   └── DNAT port range exhausted → Check port allocation
    │       └── Expand DNAT port range or add more DNAT entries
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
## NAT Gateway Proactive Inspection — [Date]

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

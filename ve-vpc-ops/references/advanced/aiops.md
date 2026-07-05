# AIOps — VPC Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[VPC Alarm Triggered]
    │
    ├── Is it connectivity-related?
    │   ├── InstanceUnreachable > 2% → Check security group rules
    │   │   └── Delegate to ve-security-group-ops if SG change needed
    │   ├── PacketLoss > 0.5% → Check VPC route tables
    │   │   └── Delegate to ve-vpc-ops for route diagnosis
    │   ├── NAT gateway drop rate > 1% → Check SNAT/DNAT rules
    │   │   └── Review NAT gateway configuration or add more NAT
    │   └── PING round-trip > 10ms → Check AZ placement or latency
    │       └── Consider same-AZ placement for latency-sensitive traffic
    │
    ├── Is it bandwidth-related?
    │   ├── InboundBandwidth > 80% of limit → Optimize or upgrade
    │   │   └── Review traffic patterns → identify optimization
    │   ├── OutboundBandwidth > 80% of limit → Optimize or upgrade
    │   │   └── Review traffic patterns → identify optimization
    │   └── High packet loss (> 0.1%) → Check network quality
    │       └── Ping/traceroute diagnostics → check max MTU
    │
    ├── Is it configuration-related?
    │   ├── Recent change in last 1h → Review change log
    │   │   └── Rollback if needed → verify recovery
    │   ├── SecurityGroupRuleChange > 5 in 24h → Review changes
    │   │   └── Validate all new rules against security policy
    │   └── ACL/firewall rule issue → Review rules
    │       └── Delegate to ve-security-group-ops
    │
    └── Unknown pattern → Delegate to ve-cms-ops
```

## Cross-Skill Routing

| Symptom | Delegate To |
|---------|------------|
| EIP allocation exhausted / bandwidth exceeded | ve-eip-ops |
| NAT gateway SNAT port exhaustion | ve-nat-ops |
| VPN tunnel flapping or BGP session down | ve-vpn-ops |
| CLB health check source IP blocked | ve-clb-ops |
| Alarm correlation across network resources | ve-cms-ops |

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
## VPC Proactive Inspection — [Date]

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

# AIOps — VPN Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[VPN Alarm Triggered]
    │
    ├── Is it connectivity-related?
    │   ├── Instance unreachable → Check security group rules
    │   │   └── Delegate to ve-security-group-ops if SG change needed
    │   ├── Route issue → Check VPC route tables
    │   │   └── Delegate to ve-vpc-ops for route diagnosis
    │   ├── NAT not working → Check SNAT/DNAT rules
    │   │   └── Review NAT gateway configuration
    │   ├── Tunnel status = down > 5min → Check IKE/IPsec phase
    │   │   └── Restart VPN tunnel or verify preshared key
    │   └── IKESASecurityAssociation expiring < 24h → Schedule renegotiation
    │       └── Verify IKE lifetime configuration
    │
    ├── Is it bandwidth-related?
    │   ├── Bandwidth at limit → Optimize or upgrade
    │   │   └── Review traffic patterns → identify optimization
    │   ├── High packet loss → Check network quality
    │   │   └── Ping/traceroute diagnostics
    │   ├── BandwidthUtilization > 85% sustained 10min → Upgrade bandwidth or shape traffic
    │   │   └── Implement QoS for critical traffic flows
    │   └── PacketLossRate > 0.5% → Investigate physical link or ISP
    │       └── Failover to secondary tunnel if available
    │
    ├── Is it configuration-related?
    │   ├── Recent change → Review change log
    │   │   └── Rollback if needed → verify recovery
    │   ├── ACL/firewall rule issue → Review rules
    │   │   └── Delegate to ve-security-group-ops
    │   ├── IPSec policy mismatch (encryption/auth algorithms) → Align on both ends
    │   │   └── Update to compatible IKE/IPsec proposals
    │   └── DPD (Dead Peer Detection) interval > 60s → Reduce to 10s for faster failover
    │       └── Enable DPD retry with 3 attempts
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
## VPN Proactive Inspection — [Date]

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

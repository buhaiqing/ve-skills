# AIOps — Security Group Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[Security Group Alarm Triggered]
    │
    ├── Is it connectivity-related?
    │   ├── Instance unreachable → Check security group rules
    │   │   ├── Inbound rule missing → Add allow rule for required port
    │   │   └── Outbound rule blocking traffic → Verify egress rules
    │   ├── Traffic denied by implicit deny rule → Check denied logs
    │   │   └── Add explicit allow rule if intended traffic
    │   ├── Rule hit count = 0 for 30+ days (stale rule) → Audit and remove
    │   │   └── Document rule purpose before removal
    │   └── Cross-security group communication failing → Check SG-to-SG rules
    │       └── Add source SG reference in inbound rules
    │
    ├── Is it bandwidth-related?
    │   ├── Bandwidth at limit → Optimize or upgrade
    │   │   └── Review traffic patterns → identify optimization
    │   ├── High packet loss → Check network quality
    │   │   └── Ping/traceroute diagnostics
    │   └── SG-bound traffic spike > 200% baseline → Investigate cause
    │       └── Check for DDoS or misconfigured health checks
    │
    ├── Is it configuration-related?
    │   ├── Recent change → Review change log
    │   │   └── Rollback if needed → verify recovery
    │   ├── ACL/firewall rule issue → Review rules
    │   │   └── Delegate to ve-security-group-ops
    │   ├── Rules count > 100 (quota approaching) → Consolidate rules
    │   │   └── Merge overlapping rules or use prefix lists
    │   └── Overly permissive 0.0.0.0/0 on non-HTTP(S) ports → Restrict
    │       └── Limit to specific CIDR blocks or security groups
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

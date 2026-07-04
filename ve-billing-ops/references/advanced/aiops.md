# AIOps — Billing Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[Billing Alarm Triggered]
    │
    ├── Is it budget-related?
    │   ├── Budget threshold exceeded (80%/90%/100%) → Review cost drivers
    │   │   ├── Compute spike → Delegate to ve-ecs-ops for instance audit
    │   │   ├── Storage growth → Delegate to ve-tos-ops for lifecycle review
    │   │   └── Network cost → Delegate to ve-eip-ops / ve-nat-ops for traffic audit
    │   └── Forecast > budget → Adjust budget or implement cost controls
    │
    ├── Is it resource-related?
    │   ├── New resource created without cost tag → Add cost tags
    │   │   └── Enforce tagging policy → delegate to ve-iam-ops
    │   ├── Idle resource detected → Stop or right-size
    │   │   └── Delegate to ve-ecs-ops / ve-rds-ops / ve-redis-ops
    │   └── Reserved instance underutilized → Modify or exchange RI
    │
    ├── Is it subscription-related?
    │   ├── PrePaid instance expiring in < 30 days → Renew
    │   ├── Auto-renewal disabled → Enable auto-renewal
    │   └── Payment overdue → Immediate payment required
    │
    └── Unknown pattern → Delegate to ve-cms-ops for correlation analysis
```

## Alarm Storm Handling

**Detection Criteria:**
- > 5 budget threshold alarms within same billing cycle
- Multiple cost spikes from same resource group

**Suppression Workflow:**
1. Correlate by resource group and time window
2. Identify largest cost contributor
3. Group related alarms under root cost driver
4. Root cause diagnosis → budget adjustment or resource optimization
5. Verify all related alarms clear after action

## Proactive Inspection Checklist

```markdown
## Billing Proactive Inspection — [Date]

### Budget Compliance
- [ ] Month-to-date spend within 80% of budget
- [ ] No unbudgeted resource groups with > 10% spend increase
- [ ] All resources have cost tags assigned

### Resource Efficiency
- [ ] No idle instances running > 7 days
- [ ] Reserved instance coverage > 60% for steady workloads
- [ ] No unattached cloud disks / unused EIPs
- [ ] Snapshot retention policy enforced (auto-delete > 90 days)

### Subscription Health
- [ ] No PrePaid instances expiring in < 30 days
- [ ] Auto-renewal enabled for production resources
- [ ] No overdue payments

### Cost Anomaly
- [ ] No > 20% MoM increase without business justification
- [ ] Cross-region traffic costs within expected range
```

## Multi-Round Diagnosis Review

Before finalizing any billing diagnosis:

1. **Fact Check:** Are the cost metrics for the correct time range? Are budget thresholds correct?
2. **Causal Analysis:** Is the identified cost driver the true root cause? Could another service explain the increase?
3. **Solution Validation:** Will the optimization actually reduce costs? Could it impact service availability?
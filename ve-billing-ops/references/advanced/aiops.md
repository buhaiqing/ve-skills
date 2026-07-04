# AIOps — Billing Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[Billing Alarm Triggered]
    │
    ├── Is it availability-related?
    │   ├── Service unreachable → Check endpoint health
    │   │   └── Verify network connectivity
    │   └── Degraded performance → Check resource usage
    │       └── Scale or optimize as needed
    │
    ├── Is it configuration-related?
    │   ├── Recent change → Review change impact
    │   │   └── Rollback if needed
    │   └── Permission issue → Check IAM/ACL
    │       └── Delegate to ve-iam-ops if needed
    │
    └── Unknown → Delegate to ve-cms-ops for correlation
```

## Alarm Storm Handling

**Detection Criteria:**
- > 10 alarms within 5 minutes
- Multiple correlated alarms from same root cause

**Suppression Workflow:**
1. Correlate alarms by resource and time
2. Address root cause
3. Verify all related alarms clear

## Proactive Inspection Checklist

```markdown
## Billing Proactive Inspection — [Date]

### Health
- [ ] All instances healthy
- [ ] No errors in recent operations
- [ ] Performance within SLA

### Security
- [ ] Access controls appropriate
- [ ] Audit logging enabled
```

# AIOps — IAM Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[IAM Alarm Triggered]
    │
    ├── Is it access-related?
    │   ├── AccessDenied rate > 100/hr → Review recent policy changes
    │   │   ├── Policy modified in last 24h → Rollback or adjust
    │   │   └── Role assumption failed > 20/hr → Check trust policy
    │   └── FailedLogin attempts > 10/min from single IP → Review source IPs
    │       └── Brute force pattern → Rotate credentials, enable MFA
    │
    ├── Is it quota-related?
    │   ├── RoleCount > 90% of limit → Review unused roles
    │   ├── PolicyDocumentSize > 6KB (near 6.5KB max) → Split or simplify
    │   └── API key count > 80% of limit → Audit and revoke unused keys
    │
    ├── Is it security-related?
    │   ├── AccessKeyAge > 90 days → Force rotation
    │   ├── InactiveUser > 90 days → Disable or delete
    │   ├── RootAccountActivity detected → Investigate immediately
    │   └── AccessKey last used > 180 days → Revoke unused keys
    │
    └── Unknown pattern → Delegate to ve-cms-ops for correlation analysis
```

## Alarm Storm Handling

**Detection Criteria:**
- > 20 access denied events within 5 minutes
- > 5 failed login attempts from same source IP
- Multiple users reporting same permission error

**Suppression Workflow:**
1. Correlate by user/role and time window
2. Identify root policy or credential causing cascade failures
3. Group related access denied alarms
4. Address root cause → verify affected users recover

## Proactive Inspection Checklist

```markdown
## IAM Proactive Inspection — [Date]

### Access Hygiene
- [ ] No access keys > 90 days old without rotation
- [ ] No inactive users (no login > 90 days)
- [ ] MFA enabled for all users with console access
- [ ] Root account activity logged and monitored

### Policy Compliance
- [ ] No policies with wildcard (*) on sensitive actions
- [ ] All policies scoped to specific resources (not */*)
- [ ] Least privilege principle verified for production roles

### Security Posture
- [ ] Failed login attempts < 5 per user per day
- [ ] No access denied spikes in last 7 days
- [ ] Credentials reported via audit trail

### Quota Usage
- [ ] Role count < 80% of limit
- [ ] Policy count < 80% of limit
```

## Multi-Round Diagnosis Review

Before finalizing any IAM diagnosis:

1. **Fact Check:** Are the access logs current? Is the time range correct?
2. **Causal Analysis:** Is the denied access due to a recent policy change or pre-existing misconfiguration?
3. **Solution Validation:** Will the policy change grant only the intended access? No privilege escalation risk?
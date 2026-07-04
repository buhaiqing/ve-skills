# AIOps — KMS Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[KMS Alarm Triggered]
    │
    ├── Is it key availability-related?
    │   ├── Key disabled or pending deletion → Check key state
    │   │   ├── Accidental disable → Re-enable key
    │   │   └── Scheduled deletion → Cancel deletion if still needed
    │   └── Key rotation failed → Check permissions
    │       └── Delegate to ve-iam-ops for role/policy audit
    │
    ├── Is it access-related?
    │   ├── Decrypt failures increasing → Verify key access
    │   │   ├── Key not found → Restore from backup
    │   │   └── Permission denied → Check IAM policy
    │   └── GenerateDataKey failures → Check key usage quota
    │
    ├── Is it performance-related?
    │   ├── API latency > 500ms → Check network
    │   │   └── Delegate to ve-vpc-ops
    │   └── Request throttling → Request quota increase
    │
    └── Unknown pattern → Delegate to ve-cms-ops for correlation analysis
```

## Alarm Storm Handling

**Detection Criteria:**
- > 10 decrypt failures within 5 minutes
- Multiple keys reporting access errors simultaneously

**Suppression Workflow:**
1. Correlate by key ID and time window
2. Identify root key causing cascade failures
3. Group related alarms under root key
4. Address root cause → verify dependent services recover

## Proactive Inspection Checklist

```markdown
## KMS Proactive Inspection — [Date]

### Key Health
- [ ] All keys in Enabled state
- [ ] No keys pending deletion without confirmed backup
- [ ] Key rotation successful (automatic or manual)
- [ ] Backup of critical keys verified

### Access Patterns
- [ ] Decrypt success rate > 99.9%
- [ ] Encrypt success rate > 99.9%
- [ ] No unauthorized key access attempts
- [ ] Key usage within 80% of quota

### Security Posture
- [ ] Key deletion protection enabled for production keys
- [ ] Key access logged and monitored
- [ ] No keys shared across environments (dev/prod)

### Compliance
- [ ] Key rotation schedule documented
- [ ] Key material backed up to secure location
```

## Multi-Round Diagnosis Review

Before finalizing any KMS diagnosis:

1. **Fact Check:** Are the key metrics for the correct region? Is the key ID correct?
2. **Causal Analysis:** Is the key issue caused by IAM policy change or KMS service health?
3. **Solution Validation:** Will the key state change impact dependent services? No unintended decryption failures?
# SecurityOps — CMS (Cloud Monitor) Security Operations

> SecurityOps content for non-security-critical skills per TE-7.
> **Purpose**: Alarm webhook security, metric data integrity, notification access

## Security Baseline Checklist

```markdown
## CMS Security Baseline — [Date]

### Access Control
- [ ] IAM policies scoped to cms required APIs only
- [ ] No hardcoded credentials in scripts/configs (use {{env.*}} placeholders)
- [ ] API key rotation schedule documented and tracked
- [ ] Service accounts mapped to specific CMS operations

### Network Security
- [ ] CMS endpoint access restricted to trusted networks
- [ ] Network ACLs / security groups scoped to CMS data plane
- [ ] Access logging enabled for audit trail

### Data Protection
- [ ] Data at rest encryption verified
- [ ] Data in transit encryption enforced (TLS 1.2+)
- [ ] Data retention policy aligned with compliance requirements

### Audit & Monitoring
- [ ] CMS operation logging enabled
- [ ] Security-related events have separate alarm rules
- [ ] Unauthorized access attempts trigger alerts via ve-cms-ops
```

## Product-Specific Security Recommendations

- Use signed webhook URLs for alarm notifications
- Restrict metric data access to monitoring and ops roles
- Set separate alarm contacts for security-critical events


# SecurityOps — SLS (日志服务) Security Operations

> SecurityOps content for non-security-critical skills per TE-7.
> **Purpose**: Log data access control, log encryption, audit trail integrity

## Security Baseline Checklist

```markdown
## SLS Security Baseline — [Date]

### Access Control
- [ ] IAM policies scoped to sls required APIs only
- [ ] No hardcoded credentials in scripts/configs (use {{env.*}} placeholders)
- [ ] API key rotation schedule documented and tracked
- [ ] Service accounts mapped to specific SLS operations

### Network Security
- [ ] SLS endpoint access restricted to trusted networks
- [ ] Network ACLs / security groups scoped to SLS data plane
- [ ] Access logging enabled for audit trail

### Data Protection
- [ ] Data at rest encryption verified
- [ ] Data in transit encryption enforced (TLS 1.2+)
- [ ] Data retention policy aligned with compliance requirements

### Audit & Monitoring
- [ ] SLS operation logging enabled
- [ ] Security-related events have separate alarm rules
- [ ] Unauthorized access attempts trigger alerts via ve-cms-ops
```

## Product-Specific Security Recommendations

- Restrict log read access to authorized ops/security roles
- Enable log encryption at rest for sensitive data
- Set immutable log retention policies to prevent tampering


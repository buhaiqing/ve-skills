# SecurityOps — PolarDB MySQL Security Operations

> SecurityOps content for non-security-critical skills per TE-7.
> **Purpose**: Database access whitelist, query audit log, encryption at rest

## Security Baseline Checklist

```markdown
## PolarDB MySQL Security Baseline — [Date]

### Access Control
- [ ] IAM policies scoped to polardb required APIs only
- [ ] No hardcoded credentials in scripts/configs (use {{env.*}} placeholders)
- [ ] API key rotation schedule documented and tracked
- [ ] Service accounts mapped to specific PolarDB MySQL operations

### Network Security
- [ ] PolarDB MySQL endpoint access restricted to trusted networks
- [ ] Network ACLs / security groups scoped to PolarDB MySQL data plane
- [ ] Access logging enabled for audit trail

### Data Protection
- [ ] Data at rest encryption verified
- [ ] Data in transit encryption enforced (TLS 1.2+)
- [ ] Data retention policy aligned with compliance requirements

### Audit & Monitoring
- [ ] PolarDB MySQL operation logging enabled
- [ ] Security-related events have separate alarm rules
- [ ] Unauthorized access attempts trigger alerts via ve-cms-ops
```

## Product-Specific Security Recommendations

- Configure IP whitelist for database access
- Enable SQL audit logging for production databases
- Use KMS-managed keys for database encryption at rest


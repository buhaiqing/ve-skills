# SecurityOps — RDS PostgreSQL Security Operations

> SecurityOps content for non-security-critical skills per TE-7.
> **Purpose**: PostgreSQL auth (md5/scram), SSL enforcement, pgAudit

## Security Baseline Checklist

```markdown
## RDS PostgreSQL Security Baseline — [Date]

### Access Control
- [ ] IAM policies scoped to rds_pg required APIs only
- [ ] No hardcoded credentials in scripts/configs (use {{env.*}} placeholders)
- [ ] API key rotation schedule documented and tracked
- [ ] Service accounts mapped to specific RDS PostgreSQL operations

### Network Security
- [ ] RDS PostgreSQL endpoint access restricted to trusted networks
- [ ] Network ACLs / security groups scoped to RDS PostgreSQL data plane
- [ ] Access logging enabled for audit trail

### Data Protection
- [ ] Data at rest encryption verified
- [ ] Data in transit encryption enforced (TLS 1.2+)
- [ ] Data retention policy aligned with compliance requirements

### Audit & Monitoring
- [ ] RDS PostgreSQL operation logging enabled
- [ ] Security-related events have separate alarm rules
- [ ] Unauthorized access attempts trigger alerts via ve-cms-ops
```

## Product-Specific Security Recommendations

- Enforce SSL connections for all PostgreSQL clients
- Use SCRAM-SHA-256 authentication where supported
- Enable pgAudit extension for query audit logging


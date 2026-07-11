# SecurityOps — RDS (关系型数据库) Security Operations

> SecurityOps content for non-security-critical skills per TE-7.
> **Purpose**: Database access whitelist, SQL audit, backup encryption

## Security Baseline Checklist

```markdown
## RDS Security Baseline — [Date]

### Access Control
- [ ] IAM policies scoped to rds required APIs only
- [ ] No hardcoded credentials in scripts/configs (use {{env.*}} placeholders)
- [ ] API key rotation schedule documented and tracked
- [ ] Service accounts mapped to specific RDS operations

### Network Security
- [ ] RDS endpoint access restricted to trusted networks
- [ ] Network ACLs / security groups scoped to RDS data plane
- [ ] Access logging enabled for audit trail

### Data Protection
- [ ] Data at rest encryption verified
- [ ] Data in transit encryption enforced (TLS 1.2+)
- [ ] Data retention policy aligned with compliance requirements

### Audit & Monitoring
- [ ] RDS operation logging enabled
- [ ] Security-related events have separate alarm rules
- [ ] Unauthorized access attempts trigger alerts via ve-cms-ops
```

## Product-Specific Security Recommendations

- Configure IP whitelist for database access
- Enable automated backups with KMS encryption
- Enable SQL audit logging for compliance


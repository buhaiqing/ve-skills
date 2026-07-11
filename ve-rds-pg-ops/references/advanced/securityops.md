# SecurityOps — RDS PostgreSQL Security Operations

> SecurityOps content for non-security-critical skills per TE-7.
> **Purpose**: Database access control, SQL audit, data encryption for RDS PostgreSQL.

## Security Baseline Checklist

```markdown
## RDS_POSTGRESQL Security Baseline — [Date]

### Access Control
- [ ] IAM policies scoped to rds_postgresql required APIs only
- [ ] No hardcoded credentials in scripts/configs (use {{env.*}} placeholders)
- [ ] API key rotation schedule documented and tracked
- [ ] Service accounts mapped to specific RDS_POSTGRESQL operations

### Network Security
- [ ] RDS_POSTGRESQL endpoint access restricted to trusted networks
- [ ] Network ACLs / security groups scoped to RDS_POSTGRESQL data plane
- [ ] Access logging enabled for audit trail

### Data Protection
- [ ] Data at rest encryption verified
- [ ] Data in transit encryption enforced (TLS 1.2+)
- [ ] Data retention policy aligned with compliance requirements

### Audit & Monitoring
- [ ] RDS_POSTGRESQL operation logging enabled
- [ ] Security-related events have separate alarm rules
- [ ] Unauthorized access attempts trigger alerts
```
```

## Product-Specific Security Recommendations

1. **SQL Audit Logging**: Enable SQL audit logging for all production RDS PostgreSQL instances. Monitor for suspicious query patterns that may indicate SQL injection or data exfiltration.
2. **Database Account Management**: Use dedicated database accounts per application with minimum required privileges. Leverage PostgreSQL role-based access control for fine-grained permissions.
3. **SSL/TLS Enforcement**: Enforce TLS connections for all client-to-database communications. Configure PostgreSQL `ssl` parameter to require encrypted connections.
4. **Backup and Snapshot Security**: Verify that automated backups are encrypted. Review snapshot sharing permissions to prevent unauthorized data access via shared snapshots.

# SecurityOps — RDS MySQL Security Operations

> SecurityOps content for non-security-critical skills per TE-7.
> **Purpose**: Database access control, SQL audit, data encryption for RDS MySQL.

## Security Baseline Checklist

```markdown
## RDS_MYSQL Security Baseline — [Date]

### Access Control
- [ ] IAM policies scoped to rds_mysql required APIs only
- [ ] No hardcoded credentials in scripts/configs (use {{env.*}} placeholders)
- [ ] API key rotation schedule documented and tracked
- [ ] Service accounts mapped to specific RDS_MYSQL operations

### Network Security
- [ ] RDS_MYSQL endpoint access restricted to trusted networks
- [ ] Network ACLs / security groups scoped to RDS_MYSQL data plane
- [ ] Access logging enabled for audit trail

### Data Protection
- [ ] Data at rest encryption verified
- [ ] Data in transit encryption enforced (TLS 1.2+)
- [ ] Data retention policy aligned with compliance requirements

### Audit & Monitoring
- [ ] RDS_MYSQL operation logging enabled
- [ ] Security-related events have separate alarm rules
- [ ] Unauthorized access attempts trigger alerts
```
```

## Product-Specific Security Recommendations

1. **SQL Audit Logging**: Enable SQL audit logs for all production RDS MySQL instances. Monitor for suspicious query patterns that may indicate SQL injection or data exfiltration.
2. **Database Account Management**: Use dedicated database accounts per application with minimum required privileges. Avoid using the root account for application connections.
3. **SSL/TLS Enforcement**: Enforce TLS connections for all client-to-database communications. Disable non-encrypted connections in production environments.
4. **Backup and Snapshot Security**: Verify that automated backups are encrypted with KMS-managed keys. Review snapshot sharing permissions to prevent unauthorized data access.

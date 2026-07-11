# SecurityOps — ARK (Database Migration) Security Operations

> SecurityOps content for non-security-critical skills per TE-7.
> **Purpose**: Migration data security, source/target connectivity, credential management

## Security Baseline Checklist

```markdown
## ARK Security Baseline — [Date]

### Access Control
- [ ] IAM policies scoped to ark required APIs only
- [ ] No hardcoded credentials in scripts/configs (use {{env.*}} placeholders)
- [ ] API key rotation schedule documented and tracked
- [ ] Service accounts mapped to specific ARK operations

### Network Security
- [ ] ARK endpoint access restricted to trusted networks
- [ ] Network ACLs / security groups scoped to ARK data plane
- [ ] Access logging enabled for audit trail

### Data Protection
- [ ] Data at rest encryption verified
- [ ] Data in transit encryption enforced (TLS 1.2+)
- [ ] Data retention policy aligned with compliance requirements

### Audit & Monitoring
- [ ] ARK operation logging enabled
- [ ] Security-related events have separate alarm rules
- [ ] Unauthorized access attempts trigger alerts via ve-cms-ops
```

## Product-Specific Security Recommendations

- Use encrypted connections (SSL/TLS) for all migration data transfer
- Rotate source/target database credentials after migration completes
- Validate destination ACLs before starting full migration


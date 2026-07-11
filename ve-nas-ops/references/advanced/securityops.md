# SecurityOps — NAS (文件存储) Security Operations

> SecurityOps content for non-security-critical skills per TE-7.
> **Purpose**: NFS/SMB access control, data encryption, snapshot security

## Security Baseline Checklist

```markdown
## NAS Security Baseline — [Date]

### Access Control
- [ ] IAM policies scoped to nas required APIs only
- [ ] No hardcoded credentials in scripts/configs (use {{env.*}} placeholders)
- [ ] API key rotation schedule documented and tracked
- [ ] Service accounts mapped to specific NAS operations

### Network Security
- [ ] NAS endpoint access restricted to trusted networks
- [ ] Network ACLs / security groups scoped to NAS data plane
- [ ] Access logging enabled for audit trail

### Data Protection
- [ ] Data at rest encryption verified
- [ ] Data in transit encryption enforced (TLS 1.2+)
- [ ] Data retention policy aligned with compliance requirements

### Audit & Monitoring
- [ ] NAS operation logging enabled
- [ ] Security-related events have separate alarm rules
- [ ] Unauthorized access attempts trigger alerts via ve-cms-ops
```

## Product-Specific Security Recommendations

- Restrict NAS mount targets to trusted VPC subnets
- Enable encryption at rest for NAS file systems
- Set snapshot retention policy aligned with compliance


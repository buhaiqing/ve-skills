# SecurityOps — TOS (对象存储) Security Operations

> SecurityOps content for non-security-critical skills per TE-7.
> **Purpose**: Bucket policy, object ACL, data encryption, public access blocking

## Security Baseline Checklist

```markdown
## TOS Security Baseline — [Date]

### Access Control
- [ ] IAM policies scoped to tos required APIs only
- [ ] No hardcoded credentials in scripts/configs (use {{env.*}} placeholders)
- [ ] API key rotation schedule documented and tracked
- [ ] Service accounts mapped to specific TOS operations

### Network Security
- [ ] TOS endpoint access restricted to trusted networks
- [ ] Network ACLs / security groups scoped to TOS data plane
- [ ] Access logging enabled for audit trail

### Data Protection
- [ ] Data at rest encryption verified
- [ ] Data in transit encryption enforced (TLS 1.2+)
- [ ] Data retention policy aligned with compliance requirements

### Audit & Monitoring
- [ ] TOS operation logging enabled
- [ ] Security-related events have separate alarm rules
- [ ] Unauthorized access attempts trigger alerts via ve-cms-ops
```

## Product-Specific Security Recommendations

- Block public access at bucket level unless explicitly required
- Use bucket policy condition keys (SourceIp, VpcSourceIp)
- Enable default encryption for all objects


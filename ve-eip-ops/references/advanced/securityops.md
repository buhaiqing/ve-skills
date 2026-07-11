# SecurityOps — EIP (弹性公网IP) Security Operations

> SecurityOps content for non-security-critical skills per TE-7.
> **Purpose**: Public IP ownership, bandwidth abuse protection, DDoS mitigation

## Security Baseline Checklist

```markdown
## EIP Security Baseline — [Date]

### Access Control
- [ ] IAM policies scoped to eip required APIs only
- [ ] No hardcoded credentials in scripts/configs (use {{env.*}} placeholders)
- [ ] API key rotation schedule documented and tracked
- [ ] Service accounts mapped to specific EIP operations

### Network Security
- [ ] EIP endpoint access restricted to trusted networks
- [ ] Network ACLs / security groups scoped to EIP data plane
- [ ] Access logging enabled for audit trail

### Data Protection
- [ ] Data at rest encryption verified
- [ ] Data in transit encryption enforced (TLS 1.2+)
- [ ] Data retention policy aligned with compliance requirements

### Audit & Monitoring
- [ ] EIP operation logging enabled
- [ ] Security-related events have separate alarm rules
- [ ] Unauthorized access attempts trigger alerts via ve-cms-ops
```

## Product-Specific Security Recommendations

- Tag EIPs with owner and purpose
- Enable DDoS mitigation for public-facing EIPs
- Restrict EIP association to authorized instances via IAM policy


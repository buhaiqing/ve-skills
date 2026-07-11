# SecurityOps — CDN (内容分发网络) Security Operations

> SecurityOps content for non-security-critical skills per TE-7.
> **Purpose**: DDoS protection, origin shield, access control, HTTPS enforcement

## Security Baseline Checklist

```markdown
## CDN Security Baseline — [Date]

### Access Control
- [ ] IAM policies scoped to cdn required APIs only
- [ ] No hardcoded credentials in scripts/configs (use {{env.*}} placeholders)
- [ ] API key rotation schedule documented and tracked
- [ ] Service accounts mapped to specific CDN operations

### Network Security
- [ ] CDN endpoint access restricted to trusted networks
- [ ] Network ACLs / security groups scoped to CDN data plane
- [ ] Access logging enabled for audit trail

### Data Protection
- [ ] Data at rest encryption verified
- [ ] Data in transit encryption enforced (TLS 1.2+)
- [ ] Data retention policy aligned with compliance requirements

### Audit & Monitoring
- [ ] CDN operation logging enabled
- [ ] Security-related events have separate alarm rules
- [ ] Unauthorized access attempts trigger alerts via ve-cms-ops
```

## Product-Specific Security Recommendations

- Enforce HTTPS-only access for all CDN domains
- Configure origin shield to protect backend from direct exposure
- Enable WAF and rate limiting for DDoS mitigation at CDN edge


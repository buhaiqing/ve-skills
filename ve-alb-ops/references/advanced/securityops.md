# SecurityOps — ALB (Application Load Balancer) Security Operations

> SecurityOps content for non-security-critical skills per TE-7.
> **Purpose**: Load balancer security, SSL/TLS termination, WAF integration

## Security Baseline Checklist

```markdown
## ALB Security Baseline — [Date]

### Access Control
- [ ] IAM policies scoped to alb required APIs only
- [ ] No hardcoded credentials in scripts/configs (use {{env.*}} placeholders)
- [ ] API key rotation schedule documented and tracked
- [ ] Service accounts mapped to specific ALB operations

### Network Security
- [ ] ALB endpoint access restricted to trusted networks
- [ ] Network ACLs / security groups scoped to ALB data plane
- [ ] Access logging enabled for audit trail

### Data Protection
- [ ] Data at rest encryption verified
- [ ] Data in transit encryption enforced (TLS 1.2+)
- [ ] Data retention policy aligned with compliance requirements

### Audit & Monitoring
- [ ] ALB operation logging enabled
- [ ] Security-related events have separate alarm rules
- [ ] Unauthorized access attempts trigger alerts via ve-cms-ops
```

## Product-Specific Security Recommendations

- Enable WAF integration for public-facing ALBs to filter malicious traffic
- Rotate SSL/TLS certificates before expiry and monitor certificate status
- Restrict listener access to trusted source IPs only


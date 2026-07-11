# SecurityOps — VPC (私有网络) Security Operations

> SecurityOps content for non-security-critical skills per TE-7.
> **Purpose**: Network ACL, subnet isolation, route table security, traffic audit

## Security Baseline Checklist

```markdown
## VPC Security Baseline — [Date]

### Access Control
- [ ] IAM policies scoped to vpc required APIs only
- [ ] No hardcoded credentials in scripts/configs (use {{env.*}} placeholders)
- [ ] API key rotation schedule documented and tracked
- [ ] Service accounts mapped to specific VPC operations

### Network Security
- [ ] VPC endpoint access restricted to trusted networks
- [ ] Network ACLs / security groups scoped to VPC data plane
- [ ] Access logging enabled for audit trail

### Data Protection
- [ ] Data at rest encryption verified
- [ ] Data in transit encryption enforced (TLS 1.2+)
- [ ] Data retention policy aligned with compliance requirements

### Audit & Monitoring
- [ ] VPC operation logging enabled
- [ ] Security-related events have separate alarm rules
- [ ] Unauthorized access attempts trigger alerts via ve-cms-ops
```

## Product-Specific Security Recommendations

- Use network ACLs as stateless firewall for subnets
- Enable VPC flow logs for traffic audit
- Avoid 0.0.0.0/0 route for production subnets


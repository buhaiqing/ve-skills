# SecurityOps — NAT (网关) Security Operations

> SecurityOps content for non-security-critical skills per TE-7.
> **Purpose**: SNAT/DNAT access control, bandwidth governance, source IP spoofing prevention

## Security Baseline Checklist

```markdown
## NAT Security Baseline — [Date]

### Access Control
- [ ] IAM policies scoped to nat required APIs only
- [ ] No hardcoded credentials in scripts/configs (use {{env.*}} placeholders)
- [ ] API key rotation schedule documented and tracked
- [ ] Service accounts mapped to specific NAT operations

### Network Security
- [ ] NAT endpoint access restricted to trusted networks
- [ ] Network ACLs / security groups scoped to NAT data plane
- [ ] Access logging enabled for audit trail

### Data Protection
- [ ] Data at rest encryption verified
- [ ] Data in transit encryption enforced (TLS 1.2+)
- [ ] Data retention policy aligned with compliance requirements

### Audit & Monitoring
- [ ] NAT operation logging enabled
- [ ] Security-related events have separate alarm rules
- [ ] Unauthorized access attempts trigger alerts via ve-cms-ops
```

## Product-Specific Security Recommendations

- Restrict SNAT source CIDRs to trusted VPC subnets
- Monitor NAT gateway bandwidth for abuse
- Audit DNAT port forwarding rules for stale entries


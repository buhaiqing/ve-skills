# SecurityOps — DNS (云解析) Security Operations

> SecurityOps content for non-security-critical skills per TE-7.
> **Purpose**: DNS query security, DDoS mitigation, domain hijack prevention

## Security Baseline Checklist

```markdown
## DNS Security Baseline — [Date]

### Access Control
- [ ] IAM policies scoped to dns required APIs only
- [ ] No hardcoded credentials in scripts/configs (use {{env.*}} placeholders)
- [ ] API key rotation schedule documented and tracked
- [ ] Service accounts mapped to specific DNS operations

### Network Security
- [ ] DNS endpoint access restricted to trusted networks
- [ ] Network ACLs / security groups scoped to DNS data plane
- [ ] Access logging enabled for audit trail

### Data Protection
- [ ] Data at rest encryption verified
- [ ] Data in transit encryption enforced (TLS 1.2+)
- [ ] Data retention policy aligned with compliance requirements

### Audit & Monitoring
- [ ] DNS operation logging enabled
- [ ] Security-related events have separate alarm rules
- [ ] Unauthorized access attempts trigger alerts via ve-cms-ops
```

## Product-Specific Security Recommendations

- Enable DNSSEC for domain validation and hijack prevention
- Configure DDoS mitigation for public DNS zones
- Use TTL lockdown for critical domain records to prevent cache poisoning


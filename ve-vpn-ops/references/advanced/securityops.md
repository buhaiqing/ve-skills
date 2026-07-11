# SecurityOps — VPN (连接) Security Operations

> SecurityOps content for non-security-critical skills per TE-7.
> **Purpose**: Tunnel encryption (IPsec/IKE), peer authentication, pre-shared key rotation

## Security Baseline Checklist

```markdown
## VPN Security Baseline — [Date]

### Access Control
- [ ] IAM policies scoped to vpn required APIs only
- [ ] No hardcoded credentials in scripts/configs (use {{env.*}} placeholders)
- [ ] API key rotation schedule documented and tracked
- [ ] Service accounts mapped to specific VPN operations

### Network Security
- [ ] VPN endpoint access restricted to trusted networks
- [ ] Network ACLs / security groups scoped to VPN data plane
- [ ] Access logging enabled for audit trail

### Data Protection
- [ ] Data at rest encryption verified
- [ ] Data in transit encryption enforced (TLS 1.2+)
- [ ] Data retention policy aligned with compliance requirements

### Audit & Monitoring
- [ ] VPN operation logging enabled
- [ ] Security-related events have separate alarm rules
- [ ] Unauthorized access attempts trigger alerts via ve-cms-ops
```

## Product-Specific Security Recommendations

- Use IKEv2 with strong encryption (AES-256, SHA-256)
- Rotate pre-shared keys periodically
- Restrict VPN peer CIDRs to known on-premise networks


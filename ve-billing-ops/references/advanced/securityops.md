# SecurityOps — Billing (费用中心) Security Operations

> SecurityOps content for non-security-critical skills per TE-7.
> **Purpose**: Cost access control, budget alerts, consumption anomaly detection

## Security Baseline Checklist

```markdown
## Billing Security Baseline — [Date]

### Access Control
- [ ] IAM policies scoped to billing required APIs only
- [ ] No hardcoded credentials in scripts/configs (use {{env.*}} placeholders)
- [ ] API key rotation schedule documented and tracked
- [ ] Service accounts mapped to specific Billing operations

### Network Security
- [ ] Billing endpoint access restricted to trusted networks
- [ ] Network ACLs / security groups scoped to Billing data plane
- [ ] Access logging enabled for audit trail

### Data Protection
- [ ] Data at rest encryption verified
- [ ] Data in transit encryption enforced (TLS 1.2+)
- [ ] Data retention policy aligned with compliance requirements

### Audit & Monitoring
- [ ] Billing operation logging enabled
- [ ] Security-related events have separate alarm rules
- [ ] Unauthorized access attempts trigger alerts via ve-cms-ops
```

## Product-Specific Security Recommendations

- Restrict billing access to finance and admin roles only
- Configure budget alerts for unexpected cost spikes (>20% threshold)
- Enable consumption anomaly detection to catch unauthorized resource provisioning


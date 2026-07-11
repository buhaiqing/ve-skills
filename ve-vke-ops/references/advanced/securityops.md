# SecurityOps — VKE (容器引擎) Security Operations

> SecurityOps content for non-security-critical skills per TE-7.
> **Purpose**: Container runtime security, image scanning, RBAC, network policy

## Security Baseline Checklist

```markdown
## VKE Security Baseline — [Date]

### Access Control
- [ ] IAM policies scoped to vke required APIs only
- [ ] No hardcoded credentials in scripts/configs (use {{env.*}} placeholders)
- [ ] API key rotation schedule documented and tracked
- [ ] Service accounts mapped to specific VKE operations

### Network Security
- [ ] VKE endpoint access restricted to trusted networks
- [ ] Network ACLs / security groups scoped to VKE data plane
- [ ] Access logging enabled for audit trail

### Data Protection
- [ ] Data at rest encryption verified
- [ ] Data in transit encryption enforced (TLS 1.2+)
- [ ] Data retention policy aligned with compliance requirements

### Audit & Monitoring
- [ ] VKE operation logging enabled
- [ ] Security-related events have separate alarm rules
- [ ] Unauthorized access attempts trigger alerts via ve-cms-ops
```

## Product-Specific Security Recommendations

- Enable container image scanning in CI/CD pipeline
- Apply least-privilege RBAC for K8s API access
- Enable network policies to isolate pod traffic by namespace


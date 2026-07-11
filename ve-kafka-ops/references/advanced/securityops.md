# SecurityOps — Kafka (消息队列) Security Operations

> SecurityOps content for non-security-critical skills per TE-7.
> **Purpose**: Message encryption, ACL-based access, consumer group security

## Security Baseline Checklist

```markdown
## Kafka Security Baseline — [Date]

### Access Control
- [ ] IAM policies scoped to kafka required APIs only
- [ ] No hardcoded credentials in scripts/configs (use {{env.*}} placeholders)
- [ ] API key rotation schedule documented and tracked
- [ ] Service accounts mapped to specific Kafka operations

### Network Security
- [ ] Kafka endpoint access restricted to trusted networks
- [ ] Network ACLs / security groups scoped to Kafka data plane
- [ ] Access logging enabled for audit trail

### Data Protection
- [ ] Data at rest encryption verified
- [ ] Data in transit encryption enforced (TLS 1.2+)
- [ ] Data retention policy aligned with compliance requirements

### Audit & Monitoring
- [ ] Kafka operation logging enabled
- [ ] Security-related events have separate alarm rules
- [ ] Unauthorized access attempts trigger alerts via ve-cms-ops
```

## Product-Specific Security Recommendations

- Enable TLS encryption for all Kafka connections
- Use ACL-based access control to restrict topic permissions
- Audit consumer group membership regularly


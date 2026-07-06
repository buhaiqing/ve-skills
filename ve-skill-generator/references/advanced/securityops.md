# SecurityOps — [Product Name] Security Operations Framework

> SecurityOps deep content per TE-7. SKILL.md `/ references/*.md` link here.
> **Purpose**: Security baseline, vulnerability detection, incident response, compliance checks.

## Security Baseline Checklist

```markdown
## [Product Name] Security Baseline — [Date]

### Access Control
- [ ] Least privilege: IAM policies scoped to [product] required APIs only
- [ ] No hardcoded credentials in scripts or configs (use {{env.*}} placeholders)
- [ ] API key rotation schedule documented and tracked
- [ ] Service accounts mapped to specific operations (not wildcard)

### Network Security
- [ ] No public exposure unless explicitly required
- [ ] Security group / ACL rules follow least privilege (no 0.0.0.0/0 on non-standard ports)
- [ ] TLS/SSL enforced for all data-in-transit connections
- [ ] Administrative access restricted to trusted IP ranges

### Data Protection
- [ ] Encryption at rest enabled (KMS-managed or default)
- [ ] Encryption in transit enforced (TLS 1.2+)
- [ ] Backup encryption enabled where applicable
- [ ] Data retention policy aligned with compliance requirements

### Audit & Monitoring
- [ ] CloudTrail / operation logging enabled for [product]
- [ ] Security-related events have separate alarm rules
- [ ] Audit logs retained for minimum compliance period
- [ ] Unauthorized access attempts trigger alerts
```

## Security Incident Response

### Triage Workflow

```
[Security Alert on {Product}]
    │
    ├── Is it authentication-related?
    │   ├── FailedLoginAttempts > threshold → Check credentials
    │   │   ├── Stale/leaked key → Delegate to ve-iam-ops (rotate key)
    │   │   └── Brute force → Enable IP restriction, notify security team
    │   └── UnauthorizedOperation (IAM) → Check policy attachment
    │       └── Delegate to ve-iam-ops for policy audit
    │
    ├── Is it data-related?
    │   ├── DataAccessAnomaly → Review access logs
    │   │   └── Delegate to ve-kms-ops if encryption key usage spike
    │   └── DataExfiltrationRisk → Immediate containment
    │       └── HALT operation, escalate to security team
    │
    ├── Is it network-exposure related?
    │   ├── NewPublicIPAssigned → Verify business justification
    │   ├── UnusualTrafficPattern → Enable VPC Flow Logs analysis
    │   │   └── Delegate to ve-vpc-ops + ve-security-group-ops
    │   └── PortScanDetected → Delegate to ve-security-group-ops
    │       └── Review SG rules, close unnecessary ports
    │
    └── Unknown → Escalate to ve-cms-ops for correlation + security team
```

### Containment Steps (destructive ops — require confirmation)

| Tier | Action | Confirmation Required |
|------|--------|-----------------------|
| 🟢 Low | Rotate credentials, update rules | No |
| 🟡 Medium | Disable service account, restrict network | Yes — via {{user.*}} |
| 🔴 High | Isolate resource (remove from network), delete | Yes + secondary verification |
| 🚨 Critical | Full resource freeze / deletion | Yes + documented approval |

## Vulnerability Scanning Patterns

### Common Vulnerability Classes

| Class | [Product] Specific Risk | Detection Method |
|-------|----------------------|-----------------|
| Misconfiguration | Overly permissive [product-specific] | Audit script / ve [product] Describe* |
| Credential Exposure | Hardcoded secrets in [product] configs | Grep for `secret`/`password` in outputs |
| Unpatched Version | [Product] running outdated version | Compare to latest via Describe*Versions |
| Unrestricted Access | Public exposure of admin interface | Verify security group / ACL rules |
| Weak Encryption | TLS < 1.2 or deprecated ciphers | Check [product] encryption config |

### Automated Scanning

```bash
# [Product] security audit checklist — run quarterly
echo "=== [Product] Security Audit [DATE] ==="

# 1. Check all public-facing resources
# (product-specific command)

# 2. Verify encryption settings
# (product-specific command)

# 3. Audit IAM permissions (delegate)
echo "→ Delegate to ve-iam-ops for permission audit"

# 4. Check compliance with baseline
# (product-specific command)
```

## Compliance Mapping

| Control Framework | [Product] Mapping | Verification |
|-------------------|------------------|--------------|
| **SOC2** — Access Control | IAM policies + SG rules | ve-iam-ops policy audit |
| **SOC2** — Encryption | KMS + [product] encryption at rest | ve kms DescribeKeys |
| **PCI-DSS** — Network Segmentation | VPC + Security Groups | ve-security-group-ops rule audit |
| **ISO 27001** — A.9 Access Control | IAM role-based access | ve-iam-ops |
| **ISO 27001** — A.12 Operations Security | [Product] audit logs | ve-cms-ops alarm rules |

## Cross-Skill Security Routing

| Security Symptom | Delegate To | Action |
|-----------------|-------------|--------|
| IAM permission denied | ve-iam-ops | Policy audit + attach/detach |
| KMS key inaccessible | ve-kms-ops | Key status check + rotation |
| Security group over-permissive | ve-security-group-ops | Rule audit + least-privilege rewrite |
| Public IP on non-public resource | ve-eip-ops | EIP release or associate to bastion |
| Abnormal API call pattern | ve-cms-ops | Alarm correlation + anomaly detection |
| Certificate expiry | ve-cdn-ops / ve-alb-ops | Renew or replace certificate |
| Audit log gap (no trail) | ve-iam-ops | Enable CloudTrail / operation log |

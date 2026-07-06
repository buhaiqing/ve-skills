# SecurityOps — ECS (Elastic Compute Service) Security Operations Framework

> SecurityOps deep content per TE-7. SKILL.md `references/*.md` link here.
> **Purpose**: Instance security baseline, vulnerability detection, incident response, compliance checks.

## Security Baseline Checklist

```markdown
## ECS Security Baseline — [Date]

### Access Control
- [ ] Least privilege: IAM policies scoped to ECS required APIs only (DescribeInstances, RunInstances, etc.)
- [ ] No hardcoded SSH keys or passwords in scripts/configs (use {{env.*}} placeholders)
- [ ] API key rotation schedule documented and tracked
- [ ] Service accounts mapped to specific ECS operations (not wildcard ecs:*)
- [ ] SSH key pair enabled for all instances (password login disabled)

### Network Security
- [ ] No public IP assignment on instances unless explicitly required
- [ ] Security group rules follow least privilege (no 0.0.0.0/0 on SSH/RDP)
- [ ] Instances in private subnets for production workloads
- [ ] Administrative access restricted to trusted IP ranges via SG
- [ ] VPC flow logs enabled for network traffic auditing

### Data Protection
- [ ] System disk and data disk encryption at rest enabled
- [ ] Encryption in transit enforced (TLS 1.2+ for application traffic)
- [ ] Snapshots encrypted with KMS-managed keys
- [ ] Data retention policy for snapshots aligned with compliance requirements

### Audit & Monitoring
- [ ] CloudTrail / operation logging enabled for ECS (Describe*, Run*, Stop* operations)
- [ ] Security-related events (SG rule changes, public IP assignment) have separate alarm rules
- [ ] Instance status change logs retained for minimum compliance period
- [ ] Unauthorized access attempts trigger alerts via ve-cms-ops
```

## Security Incident Response

### Triage Workflow

```
[ECS Security Alert]
    │
    ├── Is it authentication-related?
    │   ├── FailedLoginAttempts > threshold (SSH/RDP) → Check key pair / password
    │   │   ├── Compromised key → Replace key pair, rotate credentials
    │   │   └── Brute force → Restrict SG to trusted IPs, notify security team
    │   └── UnauthorizedOperation (RunInstances/StopInstances) → Check IAM policy
    │       └── Delegate to ve-iam-ops for policy audit
    │
    ├── Is it data-related?
    │   ├── SnapshotExported outside region → Review snapshot permissions
    │   │   └── Revoke cross-region copy if unauthorized
    │   └── DataExfiltrationRisk (large outbound traffic) → Immediate containment
    │       └── HALT operation, isolate instance, escalate to security team
    │
    ├── Is it network-exposure related?
    │   ├── NewPublicIPAssigned to instance → Verify business justification
    │   │   └── If unauthorized → Disassociate EIP, delegate to ve-eip-ops
    │   ├── UnusualTrafficPattern → Enable VPC Flow Logs analysis
    │   │   └── Delegate to ve-vpc-ops + ve-security-group-ops
    │   └── PortScanDetected → Review SG rules, close unnecessary ports
    │
    ├── Is it compute-related?
    │   ├── InstanceHijacked (unexpected process/connection) → Stop instance
    │   │   └── Create forensic snapshot before investigation
    │   ├── CryptoMining (high CPU sustained) → Stop instance, review SG
    │   │   └── Check for unauthorized outbound connections
    │   └── MalwareDetected (Cloud Assistant scan) → Isolate, snapshot, reimage
    │
    └── Unknown → Escalate to ve-cms-ops for correlation + security team
```

### Containment Steps (destructive ops — require confirmation)

| Tier | Action | Confirmation Required |
|------|--------|-----------------------|
| 🟢 Low | Modify SG rules, rotate SSH key | No |
| 🟡 Medium | Stop instance, disable API access | Yes — via {{user.*}} |
| 🔴 High | Isolate instance (remove from LB, restrict SG to null), force stop | Yes + secondary verification |
| 🚨 Critical | Terminate instance, delete snapshot chain | Yes + documented approval |

## Vulnerability Scanning Patterns

### Common Vulnerability Classes

| Class | ECS Specific Risk | Detection Method |
|-------|-------------------|-----------------|
| Misconfiguration | Public SG (0.0.0.0/0 on SSH/22, RDP/3389) | `ve ecs DescribeSecurityGroups` + rule audit |
| Credential Exposure | SSH key pair in instance metadata / user-data | Grep for `ssh-rsa`, `password` in user-data scripts |
| Unpatched OS | Outdated OS image running known CVE | Compare image ID to latest marketplace image |
| CryptoMining | Sustained high CPU with outbound connections | Cloud Assistant `top`, `netstat`, `htop` |
| Weak Encryption | Instance access without SSH key (password only) | Check VNC/console access settings |

### Automated Scanning

```bash
# ECS security audit checklist — run quarterly
echo "=== ECS Security Audit $(date +%Y-%m-%d) ==="

# 1. Check all instances with public IPs
ve ecs DescribeInstances --query 'Instances[?PublicIpAddress!=``]' --output table
echo "→ Verify each public IP has business justification"

# 2. Check all instances with password login (no key pair)
ve ecs DescribeInstances --query 'Instances[?KeyPairName==null]' --output table
echo "→ Ensure key pair is configured, disable password login"

# 3. Check security groups with 0.0.0.0/0 on management ports
echo "→ Delegate to ve-security-group-ops for SG audit"

# 4. Audit IAM permissions (delegate)
echo "→ Delegate to ve-iam-ops for permission audit"

# 5. Check instance age for patch compliance
ve ecs DescribeInstances --query 'Instances[?CreationTime<`2025-01-01`]' --output table
echo "→ Consider rebuilding these instances with latest OS images"
```

## Compliance Mapping

| Control Framework | ECS Mapping | Verification |
|-------------------|-------------|--------------|
| **SOC2** — Access Control | IAM policies + SG rules | ve-iam-ops policy audit |
| **SOC2** — Encryption | KMS + EBS encryption at rest | ve ecs DescribeDisks — encryption check |
| **PCI-DSS** — Network Segmentation | VPC + Security Groups | ve-security-group-ops rule audit |
| **PCI-DSS** — System Hardening | OS baseline + key pair auth | Cloud Assistant security scan |
| **ISO 27001** — A.9 Access Control | IAM role-based access | ve-iam-ops |
| **ISO 27001** — A.12 Operations Security | ECS audit logs + CloudTrail | ve-cms-ops alarm rules |

## Cross-Skill Security Routing

| Security Symptom | Delegate To | Action |
|-----------------|-------------|--------|
| IAM permission denied on EC2 actions | ve-iam-ops | Policy audit + attach/detach |
| Security group over-permissive | ve-security-group-ops | Rule audit + least-privilege rewrite |
| Public IP on non-public instance | ve-eip-ops | EIP release or associate to bastion |
| Abnormal API call pattern from instance | ve-cms-ops | Alarm correlation + anomaly detection |
| Encryption key rotation for EBS | ve-kms-ops | Key status check + rotation |
| Instance network isolation | ve-vpc-ops | ACL rules + subnet isolation |
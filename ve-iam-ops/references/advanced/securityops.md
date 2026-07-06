# SecurityOps — IAM (Identity and Access Management) Security Operations Framework

> SecurityOps deep content per TE-7. SKILL.md `references/*.md` link here.
> **Purpose**: Identity security baseline, credential management, policy audit, incident response.

## Security Baseline Checklist

```markdown
## IAM Security Baseline — [Date]

### Identity Governance
- [ ] Least privilege: IAM policies scoped to required APIs only (no `Action: *`)
- [ ] No root user access keys — use IAM user keys with scoped policies
- [ ] All human users have MFA enabled
- [ ] Service accounts (programmatic users) mapped to specific operations
- [ ] Inactive user accounts (>90d) identified and disabled

### Credential Management
- [ ] Access key rotation schedule: ≤90 days for service accounts
- [ ] No hardcoded credentials in scripts or configs (use {{env.*}} placeholders)
- [ ] Secret key never logged, printed, or echoed — only `<masked>` or existence check
- [ ] Temporary credentials (STS) preferred over long-lived keys for automation

### Policy Security
- [ ] No wildcard `Resource: *` in production policies
- [ ] Policy conditions enforce MFA for sensitive operations (Condition: BoolIfExists aws:MultiFactorAuthPresent)
- [ ] Policy boundaries defined (PermissionsBoundary for delegated users)
- [ ] Deny statements explicitly block dangerous actions (DeleteRole, DetachRolePolicy)

### Audit & Monitoring
- [ ] CloudTrail / operation logging enabled for all IAM actions
- [ ] IAM policy changes trigger alarm rules via ve-cms-ops
- [ ] Failed authentication attempts (>5/min) trigger alerts
- [ ] Cross-account access review performed quarterly
```

## Security Incident Response

### Triage Workflow

```
[IAM Security Alert]
    │
    ├── Is it credential-related?
    │   ├── AccessKeyUsedInAbnormalRegion → Check for key compromise
    │   │   ├── Compromised → Deactivate key immediately, delegate to ve-kms-ops if key material
    │   │   │   └── Rotate key, notify affected services, audit usage
    │   │   └── False alarm → Update alarm threshold, document exception
    │   ├── SecretKeyExposed (GitHub secret scanning) → Emergency rotation
    │   │   └── Deactivate old key → Generate new key → Update consuming services
    │   └── MultipleFailedLoginAttempts → Check credential source
    │       ├── Stale automation script → Update script, rotate key
    │       └── Brute force attack → Enable IP restriction, notify security team
    │
    ├── Is it policy-related?
    │   ├── PolicyAttachedToUnknownUser → Audit policy attachment
    │   │   └── Detach if unauthorized, delegate to ve-cms-ops for alarm tuning
    │   ├── OverlyPermissivePolicyDetected → Review policy document
    │   │   └── Rewrite with least privilege, restrict Resource/ Action scope
    │   └── CrossAccountAccessAdded → Verify business need
    │       └── If unauthorized → Remove trust relationship, notify security team
    │
    ├── Is it user-related?
    │   ├── NewUserCreated → Verify creation source (automation vs manual)
    │   │   └── If unauthorized → Disable user, audit creation trail
    │   ├── UserAddedToAdminGroup → Review role elevation
    │   │   └── If temporary → Set expiry, document Jira ticket
    │   └── OrphanedRoleDetected → Check if role has trust relationships
    │       └── No active trust → Delete role
    │
    └── Unknown → Escalate to ve-cms-ops for correlation + security team
```

### Containment Steps (destructive ops — require confirmation)

| Tier | Action | Confirmation Required |
|------|--------|-----------------------|
| 🟢 Low | Rotate access key, update alarm threshold | No |
| 🟡 Medium | Disable IAM user, detach policy | Yes — via {{user.*}} |
| 🔴 High | Delete access key, delete IAM role | Yes + secondary verification |
| 🚨 Critical | Delete IAM user, revoke trust relationship | Yes + documented approval |

## Vulnerability Scanning Patterns

### Common Vulnerability Classes

| Class | IAM Specific Risk | Detection Method |
|-------|-------------------|-----------------|
| Overly Permissive Policy | `Action: *` or `Resource: *` on production | `ve iam ListPolicies` + policy document audit |
| Stale Credentials | Access key not rotated >90 days | `ve iam ListAccessKeys` + creation date check |
| Inactive Users | No API calls in 90+ days | CloudTrail log analysis |
| Privilege Escalation | Role with trust to external account | `ve iam ListRoles` + trust policy document check |
| MFA Gap | IAM user without MFA enabled | `ve iam ListUsers` + MFA status check |

### Automated Scanning

```bash
# IAM security audit checklist — run quarterly
echo "=== IAM Security Audit $(date +%Y-%m-%d) ==="

# 1. List all IAM users and check MFA status
echo "→ Checking MFA status for all users..."
ve iam ListUsers --output table
echo "→ Verify each user has MFA enabled"

# 2. List all access keys older than 90 days
echo "→ Checking key rotation..."
ve iam ListAccessKeys --output table
echo "→ Rotate keys older than 90 days"

# 3. Audit policies for wildcards
echo "→ Checking for overly permissive policies..."
ve iam ListPolicies --query 'Policies[?PolicyDocument.contains(@,`\"*\"`)]' --output table
echo "→ Review and rewrite policies with wildcards"

# 4. Check for inactive users
echo "→ Delegate to ve-cms-ops for user activity report"

# 5. Verify cross-account trust relationships
ve iam ListRoles --query 'Roles[?AssumeRolePolicyDocument.contains(@,`\"AWS\"`)]' --output table
echo "→ Review cross-account trust relationships"
```

## Compliance Mapping

| Control Framework | IAM Mapping | Verification |
|-------------------|-------------|--------------|
| **SOC2** — Access Control | IAM policies + MFA | ve-iam-ops policy audit + MFA check |
| **SOC2** — Credential Management | Access key rotation | ve iam ListAccessKeys — age check |
| **PCI-DSS** — Authentication | MFA for all privileged users | ve iam ListUsers — MFA status |
| **ISO 27001** — A.9 Access Control | Policy-based RBAC | ve-iam-ops role audit |
| **ISO 27001** — A.9.2 User Access | User lifecycle management | ve iam ListUsers — status check |
| **NIST 800-53** — AC-6 Least Privilege | Scoped IAM policies | Policy document review |

## Cross-Skill Security Routing

| Security Symptom | Delegate To | Action |
|-----------------|-------------|--------|
| ECS instance unauthorized | ve-ecs-ops | Instance security group + key audit |
| Encryption key compromised | ve-kms-ops | Key rotation + deactivation |
| Security group rule over-permissive | ve-security-group-ops | SG rule audit + rewrite |
| Abnormal API call pattern | ve-cms-ops | Alarm correlation + anomaly detection |
| Instance public IP exposure | ve-eip-ops | EIP release or associate to bastion |
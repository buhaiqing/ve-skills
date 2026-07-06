# SecurityOps — KMS (Key Management Service) Security Operations Framework

> SecurityOps deep content per TE-7. SKILL.md `references/*.md` link here.
> **Purpose**: Key security baseline, encryption governance, key lifecycle incident response, compliance checks.

## Security Baseline Checklist

```markdown
## KMS Security Baseline — [Date]

### Key Governance
- [ ] Customer master keys (CMKs) created with explicit key administrators (not root)
- [ ] Key rotation enabled (automatic annual rotation or manual quarterly)
- [ ] Key deletion disabled by default (scheduled deletion with confirmation)
- [ ] Key usage limited to specific services via key policies
- [ ] Key alias naming convention enforced (e.g., `alias/prod/<service>/<purpose>`)

### Access Control
- [ ] Key policies scoped to specific IAM roles, not wildcard principals
- [ ] Grants limited to specific operations (Encrypt, Decrypt — not all)
- [ ] No cross-account key sharing unless explicitly required
- [ ] Key administrators separate from key users (separation of duties)
- [ ] KMS API calls restricted to VPC endpoints (no public KMS access)

### Data Protection
- [ ] All production EBS/RDS volumes encrypted with KMS keys
- [ ] S3/TOS bucket encryption using KMS keys (not default S3-managed)
- [ ] Backup snapshots encrypted with KMS-managed keys
- [ ] Encryption context enforced for all Encrypt/Decrypt calls

### Audit & Monitoring
- [ ] KMS key lifecycle events logged (CreateKey, ScheduleKeyDeletion, EnableKey, DisableKey)
- [ ] Failed Decrypt attempts (>5/min) trigger alarm rules via ve-cms-ops
- [ ] Key material expiry tracked for imported keys
- [ ] KMS API usage monitored for unusual patterns (geographic anomalies, volume spikes)
```

## Security Incident Response

### Triage Workflow

```
[KMS Security Alert]
    │
    ├── Is it key-access related?
    │   ├── DecryptCallFailed repeatedly → Check key status
    │   │   ├── Key disabled → EnableKey (requires confirmation)
    │   │   ├── Key pending deletion → CancelKeyDeletion (requires confirmation)
    │   │   └── Incorrect encryption context → Verify context string
    │   └── UnauthorizedKmsOperation → Check IAM policy
    │       └── Delegate to ve-iam-ops for policy audit
    │
    ├── Is it key-lifecycle related?
    │   ├── KeyDeletionScheduled → Verify intent
    │   │   ├── Intentional → Ensure all dependent services migrated, proceed
    │   │   └── Unauthorized → CancelKeyDeletion, audit who scheduled it
    │   ├── KeyRotationMissed → Check rotation schedule
    │   │   └── Enable automatic rotation or trigger manual rotation
    │   └── ImportedKeyMaterialExpiring → Create new key material
    │       └── ImportKeyMaterial before expiry date
    │
    ├── Is it compliance related?
    │   ├── KeyCreatedWithoutAlias → Tag with proper naming convention
    │   ├── EncryptionContextMissing → Update Encrypt/Decrypt callers
    │   └── CrossRegionKeyUsage → Verify business requirement
    │
    └── Unknown → Escalate to ve-cms-ops for correlation + security team
```

### Containment Steps (destructive ops — require confirmation)

| Tier | Action | Confirmation Required |
|------|--------|-----------------------|
| 🟢 Low | Enable key, rotate key material | No |
| 🟡 Medium | Disable key (temporarily), update key policy | Yes — via {{user.*}} |
| 🔴 High | Schedule key deletion, revoke grant | Yes + secondary verification |
| 🚨 Critical | Delete key material (imported keys), permanently delete CMK | Yes + documented approval |

## Vulnerability Scanning Patterns

### Common Vulnerability Classes

| Class | KMS Specific Risk | Detection Method |
|-------|-------------------|-----------------|
| Key Exposure | Key policy allows anonymous access | `ve kms DescribeKeys` + key policy review |
| Weak Key Protection | No rotation enabled for >1 year | `ve kms DescribeKeys` — rotation status check |
| Key Drift | Unused keys not deleted (orphaned grants) | `ve kms ListKeys` + grant association check |
| Imported Key Expiry | Imported key material nearing expiry | `ve kms DescribeKey` — validTo time check |
| Overly Permissive Grants | Grant allows all operations on key | `ve kms ListGrants` — operations constraint check |

### Automated Scanning

```bash
# KMS security audit checklist — run quarterly
echo "=== KMS Security Audit $(date +%Y-%m-%d) ==="

# 1. List all keys and check rotation status
ve kms DescribeKeys --query 'Keys[?RotationInterval!=`yearly`]' --output table
echo "→ Ensure all keys have rotation enabled"

# 2. Check keys pending deletion
ve kms DescribeKeys --query 'Keys[?KeyState==`PendingDeletion`]' --output table
echo "→ Verify all pending deletions are authorized"

# 3. Check key policies for wildcard principals
echo "→ Delegate to ve-iam-ops for key policy audit"

# 4. List all grants and verify scope
ve kms ListGrants --output table
echo "→ Review grants for overly permissive operations"

# 5. Check imported key material expiry
ve kms DescribeKeys --query 'Keys[?Origin==`EXTERNAL`]' --output table
echo "→ Renew imported key material before expiry"
```

## Compliance Mapping

| Control Framework | KMS Mapping | Verification |
|-------------------|-------------|--------------|
| **SOC2** — Encryption | CMK with automatic rotation | `ve kms DescribeKeys` — rotation check |
| **PCI-DSS** — Key Management | Secure key storage + rotation | KMS key policy audit |
| **ISO 27001** — A.10 Cryptography | Encryption key lifecycle management | Key rotation + deletion audit |
| **NIST 800-57** — Key Management | Key rotation schedule | `ve kms DescribeKeys` — rotation interval |
| **GDPR** — Data Protection | Encryption at rest + access logging | KMS CloudTrail audit |

## Cross-Skill Security Routing

| Security Symptom | Delegate To | Action |
|-----------------|-------------|--------|
| IAM permission denied on KMS actions | ve-iam-ops | Policy audit + attach/detach |
| EBS volume encryption issue | ve-ecs-ops | Disk encryption status check |
| RDS encryption at rest issue | ve-rds-mysql-ops | DB instance encryption check |
| Security group exposing KMS VPC endpoint | ve-security-group-ops | SG rule audit |
| Abnormal KMS API call pattern | ve-cms-ops | Alarm correlation + anomaly detection |
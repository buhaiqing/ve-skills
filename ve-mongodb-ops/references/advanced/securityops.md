# SecurityOps — MongoDB Security Operations Framework

> SecurityOps deep content per TE-7. SKILL.md `references/*.md` link here.
> **Purpose**: Database security baseline, access control, injection prevention, incident response.

## Security Baseline Checklist

```markdown
## MongoDB Security Baseline — [Date]

### Access Control
- [ ] MongoDB authentication enabled (--auth) for all production instances
- [ ] Least privilege: DB users scoped to specific databases/collections (no root/dbo)
- [ ] Application users with readWrite only (no dbAdmin, no userAdminAnyDatabase)
- [ ] IP whitelist restricted to application server SGs
- [ ] MongoDB admin user password rotated quarterly

### Network Security
- [ ] MongoDB instance in private subnet (no public accessibility)
- [ ] No 0.0.0.0/0 on MongoDB port (27017) — deny by default
- [ ] Security group references for cross-tier connectivity
- [ ] VPC peering/endpoint for cross-VPC access (no public endpoint)

### Data Protection
- [ ] Encryption at rest enabled (KMS-managed or instance encryption)
- [ ] Encryption in transit (TLS) enabled for all client connections
- [ ] Automated backups enabled with 7-35 day retention
- [ ] Deletion protection enabled (prevent accidental drop)

### Audit & Monitoring
- [ ] MongoDB audit log enabled (auditAuthorizationSuccess, auditCommand operations)
- [ ] Slow query monitoring with profiler (db.setProfilingLevel(1, 100))
- [ ] Connection count monitoring with alarm via ve-cms-ops
- [ ] Replication lag monitoring for replica sets
- [ ] dropDatabase / dropCollection operations trigger immediate alerts
```

## Security Incident Response

### Triage Workflow

```
[MongoDB Security Alert]
    │
    ├── Is it access-related?
    │   ├── AuthenticationFailed repeatedly → Check credentials
    │   │   ├── Stale application user → Rotate password
    │   │   └── Brute force → Restrict IP allowlist
    │   └── UnauthorizedConnectionByIP → Check allowlist
    │       └── Update allowlist, delegate to ve-security-group-ops
    │
    ├── Is it data-related?
    │   ├── dropDatabase detected → Check backup retention
    │   │   └── Restore from backup if unauthorized
    │   ├── dropCollection detected → Verify intent
    │   │   └── Restore collection from backup if unauthorized
    │   ├── UnauthorizedDataExport (find() with large batch) → Slow query log review
    │   │   └── Revoke read access if unauthorized
    │   └── NoSQLInjectionDetected → Check query patterns
    │       └── Sanitize input, use parameterized queries
    │
    ├── Is it performance/configuration-related?
    │   ├── ReplicationLagHigh → Network or secondary overload
    │   │   └── Check secondary node status, scale if needed
    │   ├── ConnectionExhaustion → Number of connections > max
    │   │   └── Increase connection limit or scale instance
    │   └── OplogTooSmall → Risk of secondary falling behind
    │       └── Increase oplog size (requires primary restart)
    │
    └── Unknown → Escalate to ve-cms-ops for correlation
```

### Containment Steps (destructive ops — require confirmation)

| Tier | Action | Confirmation Required |
|------|--------|-----------------------|
| 🟢 Low | Rotate user password, update allowlist, enable TLS | No |
| 🟡 Medium | Disable public access, revoke user privileges | Yes — via {{user.*}} |
| 🔴 High | Snapshot + terminate instance, dropDatabase (confirmed backup) | Yes + secondary verification |
| 🚨 Critical | Delete instance + backups (non-reversible) | Yes + documented approval |

## Vulnerability Scanning Patterns

### Common Vulnerability Classes

| Class | MongoDB Specific Risk | Detection Method |
|-------|----------------------|-----------------|
| NoSQL Injection | Unsanitized query operators ($where, $regex, $gt) | Application code review + slow query log |
| Auth Disabled | No authentication configured | `ve mongodb DescribeDBInstances` — auth check |
| Weak Roles | User with root/__system role | `use admin; db.system.users.find()` role audit |
| Public Exposure | Publicly accessible MongoDB | Network config check |
| Missing Encryption | TLS not enforced | SSL status parameter check |
| Oplog Size Risk | Oplog too small (< 24h capacity) | Replication config audit |

### Automated Scanning

```bash
# MongoDB security audit — run quarterly
echo "=== MongoDB Security Audit $(date +%Y-%m-%d) ==="

# 1. Check all instances for public accessibility
ve mongodb DescribeDBInstances --query 'Instances[?PublicAccessibility==`true`]' --output table
echo "→ Disable public access on ALL production instances"

# 2. Check auth enabled
ve mongodb DescribeDBInstances --query 'Instances[?AuthEnabled==`false`]' --output table
echo "→ Enable authentication on unprotected instances"

# 3. Check TLS/SSL status
ve mongodb DescribeDBInstanceSSL --output table
echo "→ Enable TLS for all client connections"

# 4. Check backup retention
ve mongodb DescribeBackupPolicy --output table
echo "→ Ensure backup retention ≥ 7 days"

# 5. Delegate SG audit
echo "→ Delegate to ve-security-group-ops for SG rule audit"
```

## Compliance Mapping

| Control Framework | MongoDB Mapping | Verification |
|-------------------|-----------------|--------------|
| **SOC2** — Access Control | Auth + RBAC + allowlist | `ve mongodb DescribeDBInstances` — auth check |
| **SOC2** — Encryption | TLS + encryption at rest | SSL + storage encryption check |
| **PCI-DSS** — Data Security | Access control + audit logging | MongoDB audit log config |
| **ISO 27001** — A.9 Access Control | Database user roles | User privilege audit |
| **GDPR** — Data Protection | Backup + encryption | Backup retention + encryption check |

## Cross-Skill Security Routing

| Security Symptom | Delegate To | Action |
|-----------------|-------------|--------|
| IAM permission denied on MongoDB | ve-iam-ops | Policy audit |
| SG over-permissive on 27017 | ve-security-group-ops | SG rule restrict |
| KMS key for encrypted MongoDB | ve-kms-ops | Key status check |
| Abnormal connection pattern | ve-cms-ops | Alarm correlation |
| ECS app unable to connect to MongoDB | ve-ecs-ops | App server network check |
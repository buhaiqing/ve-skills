# SecurityOps — RDS MySQL Security Operations Framework

> SecurityOps deep content per TE-7. SKILL.md `references/*.md` link here.
> **Purpose**: Database security baseline, SQL injection detection, access control, incident response.

## Security Baseline Checklist

```markdown
## RDS MySQL Security Baseline — [Date]

### Access Control
- [ ] Least privilege: DB accounts scoped to specific databases and operations (no wildcard GRANT)
- [ ] No default `root` user accessible from outside VPC
- [ ] Database passwords rotated quarterly, no shared passwords across instances
- [ ] IAM-based authentication preferred over password auth (where supported)
- [ ] VPN/bastion host required for all administrative DB access
- [ ] Application accounts with only DML permissions (no DDL/Schema changes)

### Network Security
- [ ] RDS instance in private subnet (no public accessibility = false)
- [ ] IP whitelist (SG) restricted to application server SGs only
- [ ] No 0.0.0.0/0 on MySQL port (3306) — deny by default
- [ ] SSL/TLS 1.2+ enforced for all client connections (require_secure_transport=ON)
- [ ] VPC endpoint used for cross-VPC access, never public endpoint

### Data Protection
- [ ] Encryption at rest enabled (KMS-managed or RDS-managed)
- [ ] Encryption in transit enforced (TLS for all connections)
- [ ] Automated backups enabled with 7-35 day retention
- [ ] Backup encryption enabled (inherited from instance encryption)
- [ ] Deletion protection enabled (prevent accidental drop)

### Audit & Monitoring
- [ ] RDS audit logs enabled (log_queries_not_using_indexes, slow_query_log)
- [ ] Security-related events (failed login, privilege escalation) have alarm rules
- [ ] SQL injection detection via WAF or SQL audit
- [ ] DDL operations logged and monitored (CREATE/DROP/ALTER)
- [ ] Connection count monitoring with thresholds (max_connections)
```

## Security Incident Response

### Triage Workflow

```
[RDS MySQL Security Alert]
    │
    ├── Is it access-related?
    │   ├── FailedLoginAttempts > threshold → Check credentials
    │   │   ├── Brute force → Restrict IP in SG, enable WAF
    │   │   ├── Stale application credentials → Rotate DB password
    │   │   └── Compromised account → Revoke access, audit queries
    │   └── UnauthorizedAccessAttempt from unknown IP → Check SG rules
    │       └── Update SG to block source IP, delegate to ve-security-group-ops
    │
    ├── Is it data-related?
    │   ├── LargeDataExport (SELECT INTO OUTFILE, mysqldump) → Check authorization
    │   │   ├── Authorized backup → No action
    │   │   └── Unauthorized → HALT, revoke FILE privilege, audit source process
    │   ├── UnexpectedTRUNCATE/DROP → Check binlog
    │   │   └── Point-in-time recovery if unauthorized
    │   └── DataExfiltrationRisk → Immediate containment
    │       └── HALT all connections, isolate instance, delegate to security team
    │
    ├── Is it SQL injection related?
    │   ├── AnomalousQueryPattern → Review slow query log
    │   │   ├── ' OR '1'='1 pattern detected → HALT, WAF rule update
    │   │   ├── UNION SELECT pattern → Application code review required
    │   │   └── Time-based blind injection → Rate limit + parameterized queries
    │   └── Known malicious IP pattern → Block at SG level
    │
    ├── Is it configuration-related?
    │   ├── PublicAccessEnabled → Disable immediately
    │   │   └── Delegate to ve-security-group-ops for SG rule audit
    │   ├── SSLDisabled for connection → Enable require_secure_transport
    │   └── EncryptionAtRestDisabled → Enable (requires migration)
    │
    └── Unknown → Escalate to ve-cms-ops for correlation + security team
```

### Containment Steps (destructive ops — require confirmation)

| Tier | Action | Confirmation Required |
|------|--------|-----------------------|
| 🟢 Low | Rotate DB password, update SG rule, enable SSL | No |
| 🟡 Medium | Restrict instance to private subnet, revoke user privileges | Yes — via {{user.*}} |
| 🔴 High | Snapshot + terminate instance, restore from backup | Yes + secondary verification |
| 🚨 Critical | Drop database, delete instance + backups (non-reversible) | Yes + documented approval |

## Vulnerability Scanning Patterns

### Common Vulnerability Classes

| Class | RDS MySQL Specific Risk | Detection Method |
|-------|------------------------|-----------------|
| SQL Injection | Unsanitized user input in queries | Slow query log analysis + WAF audit |
| Weak Password | Default/weak DB user passwords | Password complexity check |
| Overly Permissive Grants | DB user with GRANT ALL privileges | `SHOW GRANTS` per user audit |
| Unpatched MySQL | Running deprecated MySQL version | `ve rds_mysql DescribeDBInstances` — engine version |
| Public Exposure | Public accessibility enabled | `ve rds_mysql DescribeDBInstances` — public check |
| Missing Encryption | SSL/TLS not enforced | `ve rds_mysql DescribeDBInstances` — SSL status |

### Automated Scanning

```bash
# RDS MySQL security audit checklist — run quarterly
echo "=== RDS MySQL Security Audit $(date +%Y-%m-%d) ==="

# 1. Check all instances for public accessibility
ve rds_mysql DescribeDBInstances --query 'Instances[?PublicAccessibility==`true`]' --output table
echo "→ Disable public access on ALL production instances"

# 2. Check SSL/TLS enforcement
ve rds_mysql DescribeDBInstances --query 'Instances[?SSLEnabled==`false`]' --output table
echo "→ Enable SSL/TLS enforcement"

# 3. Check encryption at rest
ve rds_mysql DescribeDBInstances --query 'Instances[?StorageEncrypted==`false`]' --output table
echo "→ Enable encryption at rest"

# 4. Check outdated MySQL engine versions
ve rds_mysql DescribeDBEngineVersions --Engine MySQL --output table
echo "→ Plan upgrade for deprecated versions"

# 5. Check SG rules for MySQL port
echo "→ Delegate to ve-security-group-ops for SG audit"

# 6. Audit IAM permissions (delegate)
echo "→ Delegate to ve-iam-ops for permission audit"
```

## Compliance Mapping

| Control Framework | RDS MySQL Mapping | Verification |
|-------------------|-------------------|--------------|
| **SOC2** — Access Control | IAM + SG + DB user ACL | Multi-layer access audit |
| **SOC2** — Encryption | KMS + RDS encryption at rest | `ve rds_mysql DescribeDBInstances` |
| **PCI-DSS** — Data Security | Encryption + audit logging | SSL check + audit log review |
| **PCI-DSS** — SQL Injection | WAF + parameterized queries | Slow query audit + WAF config |
| **ISO 27001** — A.10 Cryptography | Encryption at rest + in transit | SSL + storage encryption check |
| **GDPR** — Data Protection | Backup retention + encryption | Backup config + encryption status |

## Cross-Skill Security Routing

| Security Symptom | Delegate To | Action |
|-----------------|-------------|--------|
| IAM permission denied on RDS actions | ve-iam-ops | Policy audit + attach/detach |
| Security group over-permissive on 3306 | ve-security-group-ops | SG rule audit |
| KMS key rotation for encrypted RDS | ve-kms-ops | Key status check + rotation |
| Abnormal API call pattern | ve-cms-ops | Alarm correlation + anomaly detection |
| VPC ACL issues affecting RDS access | ve-vpc-ops | ACL rule audit |
| EC2 application unable to connect | ve-ecs-ops | App server SG + network check |
# SecurityOps — Redis Cache Security Operations Framework

> SecurityOps deep content per TE-7. SKILL.md `references/*.md` link here.
> **Purpose**: Cache security baseline, access control, data protection, incident response.

## Security Baseline Checklist

```markdown
## Redis Security Baseline — [Date]

### Access Control
- [ ] Redis instance in private subnet (no public accessibility)
- [ ] IP allowlist (whitelist) restricted to application server SGs only
- [ ] No 0.0.0.0/0 on Redis port (6379) — deny by default
- [ ] Authentication password enabled for all production instances
- [ ] Application connection strings use {{env.*}} for credentials (no hardcoding)

### Network Security
- [ ] Security group limits Redis access to specific source CIDR/SGs
- [ ] VPC peering or private link for cross-VPC access (no public endpoint)
- [ ] Redis sentinel/node ports also restricted (26379, 6380 etc.)
- [ ] No wide port ranges in SG rules for Redis (restrict to exact port)

### Data Protection
- [ ] AOF persistence enabled for data durability (appendonly yes)
- [ ] RDB snapshots enabled with defined backup schedule
- [ ] Encryption in transit (TLS) enforced where supported
- [ ] Key naming convention to prevent accidental overwrites

### Audit & Monitoring
- [ ] Redis slow log enabled (slowlog-log-slower-than 10000)
- [ ] Connection count monitoring with alarm thresholds
- [ ] Memory usage monitoring (>80% triggers alarm via ve-cms-ops)
- [ ] Failed authentication attempts monitored (>5/min triggers alert)
- [ ] Instance state change (restart/flush/delete) triggers alarm
```

## Security Incident Response

### Triage Workflow

```
[Redis Security Alert]
    │
    ├── Is it access-related?
    │   ├── AuthFailed repeatedly → Check credentials
    │   │   ├── Stale application password → Rotate Redis auth string
    │   │   └── Brute force attempt → Restrict IP allowlist, enable TLS
    │   └── ConnectionFromUnknownIP → Check allowlist rules
    │       └── Update allowlist, delegate to ve-security-group-ops
    │
    ├── Is it data-related?
    │   ├── FLUSHALL/FLUSHDB detected → Check AOF/RDB for recovery
    │   │   ├── Authorized (maintenance window) → No action
    │   │   └── Unauthorized → Restore from RDB/AOF backup
    │   ├── LargeKEYDEL (unexpected eviction) → Check key expiry
    │   │   └── Set TTL on cached keys, prevent memory exhaustion
    │   └── KEYS * or SCAN with large count → Performance impact
    │       └── Rate limit scans, use SCAN with MATCH + COUNT
    │
    ├── Is it configuration-related?
    │   ├── ConfigSet notify-keyspace-events → Risk of info leak
    │   │   └── Restrict keyspace notifications if not needed
    │   ├── CONFIG SET require-pass '' → Auth disabled risk
    │   │   └── Set require-pass immediately (requires confirmation)
    │   └── SLAVEOF (replication attack) → Verify intent
    │       └── If unauthorized → SLAVEOF NO ONE, audit source
    │
    └── Unknown → Escalate to ve-cms-ops for correlation
```

### Containment Steps (destructive ops — require confirmation)

| Tier | Action | Confirmation Required |
|------|--------|-----------------------|
| 🟢 Low | Rotate auth string, update allowlist, enable TLS | No |
| 🟡 Medium | Flush DB (with backup confirmation), disable public access | Yes — via {{user.*}} |
| 🔴 High | Delete instance (with final snapshot), restore from backup | Yes + secondary verification |
| 🚨 Critical | Purge all backups + delete instance (non-reversible) | Yes + documented approval |

## Vulnerability Scanning Patterns

### Common Vulnerability Classes

| Class | Redis Specific Risk | Detection Method |
|-------|---------------------|-----------------|
| Auth Bypass | No password set on Redis instance | `ve redis DescribeDBInstances` — auth check |
| Public Exposure | Publicly accessible Redis | Instance network config check |
| Memory Exhaustion | No maxmemory-policy set | Config parameter audit |
| No Persistence | AOF/RDB disabled | `ve redis DescribeDBInstanceParameters` |
| Weak Password | Default/simple auth string | Password complexity check (manual) |
| Slow Query Risk | No slowlog configured | `ve redis DescribeDBInstanceParameters` |

### Automated Scanning

```bash
# Redis security audit — run quarterly
echo "=== Redis Security Audit $(date +%Y-%m-%d) ==="

# 1. Check all instances for public accessibility
ve redis DescribeDBInstances | jq '.Instances[] | select(.PublicAccessibility == true)'
echo "→ Disable public access on ALL production instances"

# 2. Check auth enabled
ve redis DescribeDBInstances | jq '.Instances[] | select(.AuthEnabled == false)'
echo "→ Enable authentication on unprotected instances"

# 3. Check persistence config
ve redis DescribeDBInstanceParameters | jq '.'
echo "→ Verify AOF and/or RDB persistence enabled"

# 4. Check maxmemory-policy
echo "→ Delegate to ve-redis-ops for eviction policy audit"

# 5. Delegate SG audit
echo "→ Delegate to ve-security-group-ops for SG rule audit"
```

## Compliance Mapping

| Control Framework | Redis Mapping | Verification |
|-------------------|---------------|--------------|
| **SOC2** — Access Control | Auth password + IP allowlist | `ve redis DescribeDBInstances` — auth check |
| **PCI-DSS** — Data Security | Data persistence + backup | AOF/RDB config audit |
| **ISO 27001** — A.10 Cryptography | TLS in transit | TLS config check |
| **NIST 800-53** — SC-13 Cryptography | Encryption at rest + transit | Instance encryption config |

## Cross-Skill Security Routing

| Security Symptom | Delegate To | Action |
|-----------------|-------------|--------|
| IAM permission denied on Redis actions | ve-iam-ops | Policy audit |
| SG over-permissive on 6379 | ve-security-group-ops | SG rule restrict |
| KMS key for encrypted Redis | ve-kms-ops | Key status check |
| Abnormal connection pattern | ve-cms-ops | Alarm correlation |
| VPC endpoint misconfiguration | ve-vpc-ops | Endpoint policy audit |
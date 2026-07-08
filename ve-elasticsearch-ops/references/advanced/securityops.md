# SecurityOps — Elasticsearch Security Operations Framework

> SecurityOps deep content per TE-7. SKILL.md `references/*.md` link here.
> **Purpose**: Cluster security baseline, access control, data protection, incident response.

## Security Baseline Checklist

```markdown
## Elasticsearch Security Baseline — [Date]

### Access Control
- [ ] Authentication enabled (X-Pack / Open Distro Security) for all production clusters
- [ ] Least privilege: ES users scoped to specific indices and operations
- [ ] Read-only roles for analytics/BI tool connections
- [ ] Application credentials rotated quarterly
- [ ] API key authentication preferred over basic auth for automation

### Network Security
- [ ] ES cluster in private subnet (no public accessibility)
- [ ] No 0.0.0.0/0 on ES HTTP port (9200) or transport port (9300-9400)
- [ ] Security group restricts ES access to application server SGs only
- [ ] Kibana access restricted in the same security group
- [ ] VPC endpoint for cross-VPC access (never public endpoint)

### Data Protection
- [ ] Encryption at rest enabled (KMS-managed or cluster encryption)
- [ ] Encryption in transit (HTTPS/TLS) for all HTTP and transport connections
- [ ] Index lifecycle management (ILM) with defined hot/warm/cold/delete phases
- [ ] Snapshot lifecycle management (SLM) with 7-30 day retention
- [ ] Deletion protection enabled (delete index requires confirmation)

### Audit & Monitoring
- [ ] Audit logs enabled (audit.log for authentication, access, index operations)
- [ ] Slow search/query logging enabled
- [ ] Cluster health monitoring (green status) via ve-cms-ops
- [ ] Disk usage monitoring (>80% triggers alarm)
- [ ] Index deletion/modification operations logged and alerted
- [ ] Shard allocation monitoring for node failures
```

## Security Incident Response

### Triage Workflow

```
[Elasticsearch Security Alert]
    │
    ├── Is it access-related?
    │   ├── AuthenticationFailed repeatedly → Check credentials
    │   │   ├── Stale application user → Rotate ES credentials
    │   │   └── Brute force → Restrict IP in SG, enable rate limiting
    │   ├── UnauthorizedIndexAccess → Check user roles
    │   │   └── Revoke access, update role mappings
    │   └── SSLHandshakeError → Check certificate validity
    │       └── Renew TLS certificate, update trust store
    │
    ├── Is it data-related?
    │   ├── IndexDeletedDetected → Check SLM for recovery
    │   │   ├── Accidental → Restore from snapshot
    │   │   └── Malicious → Security incident, audit access logs
    │   ├── LargeBulkIndexFromUnknownSource → Data injection risk
    │   │   └── Review source IP, validate input, restrict write access
    │   └── UnauthorizedExport (scroll+search) → Data exfiltration check
    │       └── Rate limit, revoke search privileges if unauthorized
    │
    ├── Is it cluster health related?
    │   ├── ClusterRedStatus → Missing primary shards
    │   │   └── Re-route shards or restore from snapshot
    │   ├── DiskWatermarkBreached (>95%) → Risk of read-only
    │   │   └── Scale cluster, force merge, delete old indices
    │   └── CircuitBreakerTripped → Memory pressure
    │       └── Review field mapping, increase heap
    │
    └── Unknown → Escalate to ve-cms-ops for correlation
```

### Containment Steps (destructive ops — require confirmation)

| Tier | Action | Confirmation Required |
|------|--------|-----------------------|
| 🟢 Low | Rotate user password, update SG rule, re-enable SSL | No |
| 🟡 Medium | Restrict cluster to private subnet, update role mapping | Yes — via {{user.*}} |
| 🔴 High | Delete index (with snapshot confirmation), force-merge nodes | Yes + secondary verification |
| 🚨 Critical | Delete cluster + snapshots (non-reversible) | Yes + documented approval |

## Vulnerability Scanning Patterns

### Common Vulnerability Classes

| Class | ES Specific Risk | Detection Method |
|-------|------------------|-----------------|
| Auth Disabled | X-Pack security not enabled | `ve elasticsearch DescribeInstance` — security config |
| Public Exposure | Publicly accessible ES cluster | Network access config check |
| Index Injection | Unvalidated index names in input | Application code review |
| Data Exposure | Open indices without field-level security | Index/role mapping audit |
| Weak Shard Distribution | Over-sharding (>1000 shards/node) | `_cat/shards` node count check |
| Missing Snapshots | No SLM configured | Snapshot policy audit |

### Automated Scanning

```bash
# Elasticsearch security audit — run quarterly
echo "=== Elasticsearch Security Audit $(date +%Y-%m-%d) ==="

# 1. Check all instances for public accessibility
ve elasticsearch DescribeInstances | jq '.Instances[] | select(.PublicAccessibility == true)'
echo "→ Disable public access on ALL production clusters"

# 2. Check security/authentication enabled
ve elasticsearch DescribeInstance | jq '.'
echo "→ Verify security plugin / authentication enabled"

# 3. Check snapshot configuration
ve elasticsearch DescribeSnapshotConfiguration | jq '.'
echo "→ Ensure SLM enabled with adequate retention"

# 4. Check cluster health
ve elasticsearch DescribeInstanceHealth | jq '.'
echo "→ All clusters should be green status"

# 5. Delegate SG audit
echo "→ Delegate to ve-security-group-ops for SG rule audit"

# 6. Delegate KMS audit if encryption managed keys
echo "→ Delegate to ve-kms-ops for encryption key audit"
```

## Compliance Mapping

| Control Framework | ES Mapping | Verification |
|-------------------|------------|--------------|
| **SOC2** — Access Control | X-Pack/security auth + RBAC | `ve elasticsearch DescribeInstance` — auth check |
| **SOC2** — Encryption | HTTPS + encryption at rest | TLS + storage encryption check |
| **PCI-DSS** — Logging | Audit logging + slow query log | Audit log config check |
| **ISO 27001** — A.12 Operations | SLM + ILM policies | Snapshot + index lifecycle audit |
| **GDPR** — Data Retention | ILM deletion phases | Index lifecycle policy review |

## Cross-Skill Security Routing

| Security Symptom | Delegate To | Action |
|-----------------|-------------|--------|
| IAM permission denied on ES actions | ve-iam-ops | Policy audit |
| SG over-permissive on 9200/9300 | ve-security-group-ops | SG rule restrict |
| KMS key for encrypted ES cluster | ve-kms-ops | Key rotation + status |
| Abnormal search/query pattern | ve-cms-ops | Alarm correlation + anomaly detection |
| ECS app unable to connect to ES | ve-ecs-ops | App server network + SG check |
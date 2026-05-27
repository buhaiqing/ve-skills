# Monitoring — TOS

## TOS does not expose built-in metrics by default unlike ECS. Metrics are available via Volcengine Cloud Monitor (CMS).

## Key Metrics

| Metric | Description | Threshold (Warning) | Threshold (Critical) |
|--------|-------------|---------------------|---------------------|
| `RequestCount` | Requests per minute | Monitor trend | — |
| `4xxErrorRate` | Client error rate (%) | > 1% | > 5% |
| `5xxErrorRate` | Server error rate (%) | > 0.1% | > 1% |
| `BandwidthIn` | Inbound bandwidth (bytes/s) | — | — |
| `BandwidthOut` | Outbound bandwidth (bytes/s) | — | — |
| `StorageUsed` | Total storage (bytes) | — | — |
| `ObjectCount` | Total object count | — | — |
| `FirstByteLatency` | Time to first byte (ms) | > 1000ms | > 5000ms |
| `ActiveConnections` | Active connections | — | — |

## Anomaly Patterns

| Pattern | Metrics | Detection Logic | Severity |
|---------|---------|-----------------|----------|
| Error Spike | 4xxErrorRate > 5% | Sustained for 5min | Warning |
| Access Denied Flood | AccessDenied rate spikes | Sudden increase > 10x | Critical |
| Bandwidth Saturation | BandwidthOut > 90% of limit | Sustained for 10min | Warning |
| High Latency | FirstByteLatency > 5s for GET | Sustained for 5min | Warning |
| Storage Surge | StorageUsed growth > 50% in 1hr | Rapid increase | Warning |

## Monitoring Query

```bash
# Query TOS metrics via CMS API
ve cms DescribeMetricData --Region cn-beijing --Namespace Volcengine_TOS --MetricName RequestCount
```

## Alert Delegation Matrix

| Alarm Source | Primary Diagnosis Skill | Secondary Skill |
|-------------|------------------------|-----------------|
| TOS 4xx/5xx error spike | ve-tos-ops | ve-cms-ops |
| TOS storage growth anomaly | ve-tos-ops | ve-cms-ops |
| TOS access denied flood | ve-tos-ops | ve-iam-ops |
| TOS bandwidth saturation | ve-tos-ops | ve-vpc-ops |

## AIOps — Intelligent Operations

### Cross-Skill Diagnosis Decision Tree

```
[TOS Alarm Triggered]
    │
    ├── Is it error-related?
    │   ├── 4xx spike → Check client permissions and ACLs
    │   │   ├── AccessDenied → Check IAM policy, bucket ACL
    │   │   │   └── Recent ACL change? Rollback
    │   │   ├── NoSuchKey → Application referencing deleted objects
    │   │   │   └── Check versioning, enable lifecycle
    │   │   └── SignatureDoesNotMatch → Clock skew or credential issue
    │   │       └── Verify system time, rotate credentials
    │   └── 5xx spike → Server-side issue
    │       └── Delegate to ve-cms-ops for service health check
    │
    ├── Is it storage-related?
    │   ├── Storage growth > 50%/day → Check upload sources
    │   │   ├── Application writing logs? → Redirect to log service
    │   │   ├── Backup misconfigured? → Fix backup job
    │   │   └── Credential abuse? → Rotate keys
    │   └── Storage > 90% of quota → Cleanup or increase quota
    │
    ├── Is it performance-related?
    │   ├── High latency → Check bandwidth, CDN, network path
    │   │   ├── Bandwidth saturated → Enable CDN for hot objects
    │   │   ├── Cross-region access → Use VPC endpoint
    │   │   └── Hot object → Implement client-side caching
    │   └── Slow uploads → Check part size, network, concurrency
    │       └── Reduce part size, enable retry
    │
    └── Unknown pattern → Delegate to ve-cms-ops for correlation analysis
```

### Proactive Inspection Checklist

```markdown
## TOS Proactive Inspection — [Date]

### Resource Health
- [ ] 4xx error rate < 1% across all buckets
- [ ] 5xx error rate < 0.1% across all buckets
- [ ] FirstByteLatency < 1000ms for GET requests
- [ ] No incomplete multipart uploads > 7 days

### Cost Optimization
- [ ] No Standard objects idle > 30 days (recommend IA)
- [ ] No IA objects idle > 90 days (recommend Archive)
- [ ] Lifecycle rules configured for all production buckets
- [ ] Bucket versioning cleanup rules active (if versioning enabled)

### Security Posture
- [ ] No buckets with public-read-write ACL
- [ ] Bucket policies restrict access by IP/role where needed
- [ ] Pre-signed URLs with appropriate expiration (< 24h)
- [ ] No credentials exposed in code or configs

### Reliability
- [ ] Cross-region replication configured for critical buckets
- [ ] Backup strategy documented and tested
- [ ] No single bucket exceeding 80% of account storage quota
```

### Multi-Round Diagnosis Review

Before finalizing any TOS diagnosis:

1. **Fact Check:** Are the TOS metrics current? Are thresholds correct for this bucket's access pattern?
2. **Causal Analysis:** Is the identified cause the true root cause? Could a client change explain the symptoms?
3. **Solution Validation:** Will the fix actually resolve the issue? Could lifecycle changes affect active data?

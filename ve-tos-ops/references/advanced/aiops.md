# AIOps — TOS (Object Storage) Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[TOS (Object Storage) Alarm Triggered]
    │
    ├── Is it capacity-related?
    │   ├── StorageUsed > 80% of bucket quota → Plan capacity expansion
    │   │   └── Review lifecycle policies → clean up old data
    │   ├── BucketObjectCount > 2.5B (near 5B max) → Archive old objects
    │   │   └── Enable transition to cold storage tier
    │   └── Quota approaching → Alert team → schedule cleanup
    │
    ├── Is it performance-related?
    │   ├── PutObject latency > 500ms → Check throughput limits
    │   │   └── Review access patterns → optimize requests
    │   ├── GetObject latency > 300ms → Review object size/network
    │   │   └── Enable CDN for hot objects via ve-cdn-ops
    │   └── Throughput < 50% of expected → Check network bandwidth
    │       └── Delegate to ve-vpc-ops if network issue
    │
    ├── Is it availability-related?
    │   ├── 5xx error rate > 2% → Check service health
    │   │   └── Review recent API call errors
    │   ├── Throttling (503 SlowDown) > 10/min → Review rate limits
    │   │   └── Optimize request patterns or request quota increase
    │   └── ListBucket errors > 1% → Check bucket policy or ACLs
    │       └── Delegate to ve-iam-ops if permission issue
    │
    └── Unknown pattern → Delegate to ve-cms-ops for correlation
```

## Cross-Skill Routing

| Symptom | Delegate To |
|---------|------------|
| CDN cache miss or stale cache for hot objects | ve-cdn-ops |
| IAM bucket policy denied | ve-iam-ops |
| Unexpected cost surge from storage usage | ve-billing-ops |
| Network bandwidth / connectivity issue | ve-vpc-ops |

## Alarm Storm Handling

**Detection Criteria:**
- > 10 alarms within 5 minutes for same bucket/instance
- > 3 concurrent throttling alarms

**Suppression Workflow:**
1. Correlate by bucket/instance and time window
2. Identify root cause (capacity vs performance vs availability)
3. Group related alarms
4. Address root cause → verify all clear

## Proactive Inspection Checklist

```markdown
## TOS (Object Storage) Proactive Inspection — [Date]

### Capacity
- [ ] Usage < 80% across all buckets/instances
- [ ] No buckets approaching quota limits
- [ ] Lifecycle policies configured for old data

### Performance
- [ ] Average latency within SLA
- [ ] No throttling events in past 7 days
- [ ] Request success rate > 99.5%

### Security
- [ ] No public access on sensitive buckets
- [ ] Encryption enabled for all data
- [ ] Access logs enabled and monitored
```

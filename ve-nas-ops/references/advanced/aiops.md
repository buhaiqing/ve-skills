# AIOps — NAS Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[NAS Alarm Triggered]
    │
    ├── Is it capacity-related?
    │   ├── Usage > 80% → Plan capacity expansion
    │   │   └── Review lifecycle policies → clean up old data
    │   ├── Quota approaching → Alert team → schedule cleanup
    │   ├── StorageUsed growth > 10% per month → Investigate data accumulation
    │   │   └── Implement lifecycle rules for auto-deletion
    │   └── InodeUsage > 80% → Check for small-file explosion
    │       └── Consolidate small files or increase inode quota
    │
    ├── Is it performance-related?
    │   ├── High latency → Check throughput limits
    │   │   └── Review access patterns → optimize requests
    │   ├── Low throughput → Check network bandwidth
    │   │   └── Delegate to ve-vpc-ops if network issue
    │   ├── Read latency > 10ms sustained → Check IOPS budget
    │   │   └── Upgrade to higher IOPS tier or cache hot data
    │   └── Write latency > 20ms sustained → Check write amplification
    │       └── Batch small writes or switch to async writes
    │
    ├── Is it availability-related?
    │   ├── Request errors > 1% → Check service health
    │   │   └── Review recent API call errors
    │   ├── Throttling detected → Review rate limits
    │   │   └── Optimize request patterns or request quota increase
    │   ├── Filesystem mounting failures > 3/day → Check NFS client configuration
    │   │   └── Verify mount options and network path
    │   └── Throughput < 50% provisioned → Check for bottleneck
    │       └── Review access patterns (sequential vs random I/O)
    │
    └── Unknown pattern → Delegate to ve-cms-ops for correlation
```

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
## NAS Proactive Inspection — [Date]

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

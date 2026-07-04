# AIOps — NAS Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[NAS Alarm Triggered]
    │
    ├── Is it capacity-related?
    │   ├── Usage > 80% → Plan capacity expansion
    │   │   └── Review lifecycle policies → clean up old data
    │   └── Quota approaching → Alert team → schedule cleanup
    │
    ├── Is it performance-related?
    │   ├── High latency → Check throughput limits
    │   │   └── Review access patterns → optimize requests
    │   └── Low throughput → Check network bandwidth
    │       └── Delegate to ve-vpc-ops if network issue
    │
    ├── Is it availability-related?
    │   ├── Request errors > 1% → Check service health
    │   │   └── Review recent API call errors
    │   └── Throttling detected → Review rate limits
    │       └── Optimize request patterns or request quota increase
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

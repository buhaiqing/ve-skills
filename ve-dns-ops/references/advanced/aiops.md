# AIOps — Private DNS Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[Private DNS Alarm Triggered]
    │
    ├── Is it resolution-related?
    │   ├── High DNS query failure rate (> 2%) → Check zone health
    │   │   ├── NXDOMAIN ratio > 5% → Missing records or expired entries
    │   │   ├── SERVFAIL ratio > 1% → Upstream resolver unreachable → Check VPC DNS settings
    │   │   └── Zone misconfigured → Verify record sets
    │   └── Slow resolution (> 200ms) → Check recursive resolver latency
    │       ├── Authoritative response time > 100ms → Delegate to ve-vpc-ops for VPC DNS endpoint check
    │       └── TTL too short causing upstream bursts → Increase TTL on stable records
    │
    ├── Is it configuration-related?
    │   ├── Recent zone change → Review change log
    │   │   ├── More than 5 changes in 10min → Possible misconfiguration — rollback
    │   │   └── CNAME conflict with apex record → Remove conflicting record
    │   ├── Record conflict detected → Review overlapping records
    │   │   └── Same name mapped to multiple IPs → Validate intended resolution
    │   └── Zone transfer failed (> 3 retries) → Check authorization
    │       └── TSIG key mismatch or expired → Reconfigure AXFR/IXFR
    │
    ├── Is it quota-related?
    │   ├── Zone count approaching limit → Consolidate zones
    │   ├── Record count per zone > 80% → Add recordsets or split zone
    │   └── Query rate throttled → Request quota increase
    │
    └── Unknown pattern → Delegate to ve-cms-ops for correlation analysis
```

## Alarm Storm Handling

**Detection Criteria:**
- > 50 resolution failures within 5 minutes
- > 10% increase in resolution latency

**Suppression Workflow:**
1. Correlate by zone and time window
2. Identify root zone or record causing failures
3. Group related alarms under root zone
4. Address root cause → verify resolution recovers

## Proactive Inspection Checklist

```markdown
## Private DNS Proactive Inspection — [Date]

### Resolution Health
- [ ] Query failure rate < 1%
- [ ] Average resolution latency < 50ms
- [ ] No zones in ERROR state

### Configuration Hygiene
- [ ] No duplicate record sets across zones
- [ ] All zones have appropriate TTL settings
- [ ] No orphaned zones (unused > 30 days)
- [ ] Change log enabled and reviewed

### Quota & Capacity
- [ ] Zone count < 80% of limit
- [ ] Records per zone < 80% of limit
- [ ] Query rate < 70% of throttling limit

### Security
- [ ] No public zones exposing internal records
- [ ] Zone transfer restricted to authorized VPCs
```

## Multi-Round Diagnosis Review

Before finalizing any DNS diagnosis:

1. **Fact Check:** Are the resolution metrics from the affected VPC? Is the time window correct?
2. **Causal Analysis:** Is the resolution failure due to DNS misconfiguration or upstream network issue?
3. **Solution Validation:** Will the record change resolve the issue without introducing conflicts?
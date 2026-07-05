# AIOps — SLS (Log Service) Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[SLS (Log Service) Alarm Triggered]
    │
    ├── Is it ingestion-related?
    │   ├── LogIngestionRate dropping > 30% → Check source connectivity
    │   │   ├── Agent/collector down → Restart log agent
    │   │   │   └── Delegate to ve-ecs-ops for instance health
    │   │   └── Network loss > 0.1% → Check VPC connectivity
    │   │       └── Delegate to ve-vpc-ops
    │   ├── IngestionThrottling > 100 req/s → Request quota increase
    │   ├── DataFormatError > 10/min → Check log parsing config
    │   └── LogShipperErrorRate > 1% → Reinstall or update shipper
    │       └── Check shipper version compatibility
    │
    ├── Is it storage-related?
    │   ├── LogStorage usage > 80% → Adjust retention or expand
    │   ├── IndexSize growing > 10% per day → Review index config
    │   │   └── Consider full-text vs word split indexing
    │   └── ColdStorage > HotStorage ratio > 3:1 → Adjust tiering policy
    │       └── Review hot retention window
    │
    ├── Is it query-related?
    │   ├── QueryLatency p99 > 10s → Optimize query patterns
    │   │   ├── FullScanQuery → Add time range filter
    │   │   └── Missing index → Create appropriate index
    │   ├── ConcurrentQueryLimit reached > 5/min → Queue or throttle
    │   └── DashboardLoad time > 30s → Optimize dashboard queries
    │       └── Reduce query time range or add pre-aggregation
    │
    └── Unknown pattern → Delegate to ve-cms-ops for correlation analysis
```

## Alarm Storm Handling

**Detection Criteria:**
- > 5 log shippers reporting errors simultaneously
- Ingestion rate drop > 50% across multiple logstores

**Suppression Workflow:**
1. Correlate by project/logstore and time window
2. Identify root cause (agent vs network vs SLS service)
3. Group related shipper alarms
4. Address root cause → verify ingestion resumes

## Proactive Inspection Checklist

```markdown
## SLS (Log Service) Proactive Inspection — [Date]

### Ingestion Health
- [ ] All logstores receiving data within expected rate ± 20%
- [ ] No ingestion errors in past 24 hours
- [ ] Log agent/collector running on all source instances

### Storage & Retention
- [ ] Total storage usage < 75% of quota
- [ ] Retention policy configured per logstore (hot → cold → delete)
- [ ] Index size < 50% of total storage

### Query Performance
- [ ] Average query latency < 5s
- [ ] No queries hitting concurrent limit
- [ ] Dashboard refresh within expected interval

### Security
- [ ] No public access to log data
- [ ] Log data encrypted at rest
- [ ] Audit trail for log access enabled
```

## Multi-Round Diagnosis Review

Before finalizing any SLS diagnosis:

1. **Fact Check:** Are the ingestion metrics for the correct project/logstore? Is the time range correct?
2. **Causal Analysis:** Is the ingestion drop due to source failure or SLS service degradation?
3. **Solution Validation:** Will the config change fix the ingestion without losing log data?
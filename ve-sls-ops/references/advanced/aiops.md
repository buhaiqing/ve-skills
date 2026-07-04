# AIOps — SLS (Log Service) Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[SLS (Log Service) Alarm Triggered]
    │
    ├── Is it ingestion-related?
    │   ├── Log ingestion rate dropping → Check source connectivity
    │   │   ├── Agent/collector down → Restart log agent
    │   │   │   └── Delegate to ve-ecs-ops for instance health
    │   │   └── Network issue → Check VPC connectivity
    │   │       └── Delegate to ve-vpc-ops
    │   ├── Throttling detected → Request quota increase
    │   └── Data format errors → Check log parsing config
    │
    ├── Is it storage-related?
    │   ├── Log storage > 80% → Adjust retention or expand
    │   ├── Index size growing fast → Review index config
    │   └── Cold storage > hot storage → Adjust tiering policy
    │
    ├── Is it query-related?
    │   ├── Query latency > 10s → Optimize query patterns
    │   │   ├── Full scan query → Add time range filter
    │   │   └── Missing index → Create appropriate index
    │   └── Concurrent query limit reached → Queue or throttle
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
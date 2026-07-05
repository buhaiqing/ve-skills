# AIOps — Cloud Monitor Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[Cloud Monitor Alarm Triggered]
    │
    ├── Is it metric-related?
    │   ├── Missing data > 30% in evaluation window → Check agent/metric source
    │   │   └── Verify monitoring agent on target
    │   ├── MetricIngestionLag > 120s → Check metric ingestion pipeline
    │   │   └── Verify API endpoint reachability
    │   ├── DataPoints < 50% of expected → Check metric collection interval
    │   │   └── Review metric push frequency from target
    │   └── Incorrect values (NaN, spikes > 3σ) → Check metric definition
    │       └── Review metric calculation logic
    │
    ├── Is it alarm-related?
    │   ├── Alarm evaluation failure > 5% → Check alarm rule syntax
    │   │   └── Verify threshold and notification channel
    │   ├── Alarm storm > 100 alarms/5min → Correlate and suppress
    │   │   └── Use alarm grouping rules
    │   ├── Notification delivery latency > 60s → Check channel
    │   │   └── Verify webhook/email/SMS configuration
    │   └── Same alarm firing > 10 times without resolution → Escalate
    │       └── Review alarm threshold — may be too sensitive
    │
    └── Unknown → Escalate to monitoring platform team
```

## Alarm Storm Handling

**Detection Criteria:**
- > 100 alarms within 5 minutes
- Same alarm firing repeatedly without resolution

**Suppression Workflow:**
1. Enable bulk suppression for affected scope
2. Identify root alarm
3. Resolve root cause
4. Bulk unsuppress after resolution
5. Review and tune alarm rules

## Proactive Inspection Checklist

```markdown
## Cloud Monitor Proactive Inspection — [Date]

### Coverage
- [ ] All critical resources have alarm rules
- [ ] No duplicate/conflicting alarm rules
- [ ] Alarm notifications reaching correct channels

### Delivery
- [ ] Notification channels tested monthly
- [ ] Escalation policies configured
- [ ] On-call rotation current

### Quality
- [ ] Alarm precision > 80%
- [ ] Alarm recall > 90%
- [ ] MTTR trending down
```

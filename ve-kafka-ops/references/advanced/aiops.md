# AIOps — Kafka Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[Kafka Alarm Triggered]
    │
    ├── Is it broker-related?
    │   ├── Broker down → Check broker health
    │   │   └── Review recent broker operations
    │   ├── Disk full → Check log retention
    │   │   └── Adjust retention policy or expand disk
    │   └── Leader election → Check partition health
    │       └── Review ISR (In-Sync Replicas) status
    │
    ├── Is it consumer-related?
    │   ├── Consumer lag → Check consumer health
    │   │   └── Scale consumers or optimize processing
    │   ├── Consumer group rebalance → Check consumer stability
    │   │   └── Review consumer session timeout
    │   └── Message processing failure → Check DLQ
    │       └── Analyze failed messages
    │
    ├── Is it producer-related?
    │   ├── Produce failures → Check broker connectivity
    │   │   └── Verify broker addresses and SSL
    │   └── Ack timeout → Check broker performance
    │       └── Optimize batch settings
    │
    └── Unknown → Check controller logs
```

## Alarm Storm Handling

**Detection Criteria:**
- > 3 broker failures within 10 minutes
- Consumer lag > 100,000 messages
- Disk usage > 80% on any broker

**Suppression Workflow:**
1. Identify affected topics and partitions
2. Ensure minimum replicas available
3. Address broker/disk issues
4. Verify consumer catch-up
5. Monitor backlog clearing

## Proactive Inspection Checklist

```markdown
## Kafka Proactive Inspection — [Date]

### Cluster Health
- [ ] All brokers healthy
- [ ] No under-replicated partitions
- [ ] Controller election stable

### Performance
- [ ] Consumer lag < 10,000 messages
- [ ] Produce latency < 100ms p99
- [ ] Disk usage < 70%

### Security
- [ ] SASL/SSL enabled
- [ ] Access controls configured
```

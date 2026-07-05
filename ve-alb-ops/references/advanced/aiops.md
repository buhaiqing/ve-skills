# AIOps — ALB Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[ALB Alarm Triggered]
    │
    ├── Is it backend-related?
    │   ├── HealthyHostCount < 80% of target → Check backend health
    │   │   └── Delegate to ve-ecs-ops / ve-vke-ops as needed
    │   ├── BackendResponseTime p99 > 2s → Check backend metrics
    │   │   └── Scale backend or adjust load distribution
    │   ├── ActiveConnections > 80% of limit → Check connection pool
    │   │   └── Optimize backend connection settings
    │   └── RequestCount > 100000/min → Scale backend group
    │       └── Consider adding more server group members
    │
    ├── Is it listener-related?
    │   ├── Listener error > 1% → Check listener configuration
    │   │   └── Verify protocol, port, certificate
    │   ├── SSL handshake failure > 10/min → Check certificates
    │   │   └── Renew or update certificate
    │   └── Listener drop count > 100/min → Check connection limit
    │       └── Increase listener connection limit
    │
    ├── Is it routing-related?
    │   ├── Requests not reaching backends > 1% → Check rules
    │   │   └── Verify routing rules and target groups
    │   ├── ALB 5xx error rate > 2% → Check LB health
    │   │   └── Review LB logs for error patterns
    │   └── RuleEvaluationError > 1% → Review custom rules
    │       └── Simplify complex routing rules
    │
    └── Unknown → Delegate to ve-cms-ops
```

## Alarm Storm Handling

**Detection Criteria:**
- > 5 backend failures within 5 minutes
- > 20% of backends unhealthy simultaneously

**Suppression Workflow:**
1. Isolate failing backends (drain connections)
2. Investigate backend issues
3. Restore backends or adjust routing
4. Re-enable traffic gradually
5. Verify all backends healthy

## Proactive Inspection Checklist

```markdown
## ALB Proactive Inspection — [Date]

### Health
- [ ] All backends healthy (100% pass rate)
- [ ] No backend with consecutive failures
- [ ] Health check settings appropriate

### Performance
- [ ] LB latency < 100ms p99
- [ ] Connection count < 80% of limit
- [ ] Throughput within contracted limits

### Security
- [ ] WAF enabled if required
- [ ] No open listener on unnecessary ports
- [ ] SSL certificates valid
```

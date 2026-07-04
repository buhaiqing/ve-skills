# AIOps — CLB Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[CLB Alarm Triggered]
    │
    ├── Is it backend-related?
    │   ├── Backend unhealthy → Check backend health
    │   │   └── Delegate to ve-ecs-ops / ve-vke-ops as needed
    │   ├── Backend overload → Check backend metrics
    │   │   └── Scale backend or adjust load distribution
    │   └── Connection limit → Check connection pool
    │       └── Optimize backend connection settings
    │
    ├── Is it listener-related?
    │   ├── Listener error → Check listener configuration
    │   │   └── Verify protocol, port, certificate
    │   └── SSL handshake failure → Check certificates
    │       └── Renew or update certificate
    │
    ├── Is it routing-related?
    │   ├── Requests not reaching backends → Check rules
    │   │   └── Verify routing rules and target groups
    │   └── 5xx errors from LB → Check LB health
    │       └── Review LB logs for error patterns
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
## CLB Proactive Inspection — [Date]

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

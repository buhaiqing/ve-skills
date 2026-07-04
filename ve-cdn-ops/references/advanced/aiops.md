# AIOps — CDN Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[CDN Alarm Triggered]
    │
    ├── Is it origin-related?
    │   ├── Origin error spike → Check origin health
    │   │   └── Delegate to origin service ops as needed
    │   ├── High origin latency → Check origin response time
    │   │   └── Optimize origin or enable origin shield
    │   └── Origin unreachable → Check network path
    │       └── Verify origin configuration
    │
    ├── Is it cache-related?
    │   ├── Low hit ratio → Check cache headers
    │   │   └── Review cache TTL settings
    │   ├── Cache miss storm → Check cache warming
    │   │   └── Pre-populate popular content
    │   └── Stale content → Check refresh mechanism
    │       └── Purge cache or adjust TTL
    │
    ├── Is it configuration-related?
    │   ├── SSL error → Check certificate
    │   │   └── Renew certificate
    │   └── DNS propagation issue → Check DNS
    │       └── Wait for propagation or force refresh
    │
    └── Unknown → Check CDN provider status
```

## Alarm Storm Handling

**Detection Criteria:**
- > 10 5xx errors within 5 minutes
- Cache miss rate > 50%
- Origin error rate > 10%

**Suppression Workflow:**
1. Enable detailed logging temporarily
2. Check CDN provider incident page
3. Verify origin health
4. Adjust CDN configuration as needed
5. Monitor recovery

## Proactive Inspection Checklist

```markdown
## CDN Proactive Inspection — [Date]

### Performance
- [ ] Cache hit ratio > 85%
- [ ] Average response time within SLA
- [ ] 4xx error rate < 1%

### Security
- [ ] DDoS protection enabled
- [ ] Hotlink protection configured
- [ ] SSL certificates valid

### Cost
- [ ] Bandwidth usage within budget
- [ ] No unexpected traffic spikes
```

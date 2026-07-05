# AIOps — CDN Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[CDN Alarm Triggered]
    │
    ├── Is it origin-related?
    │   ├── Origin 5xx rate > 5% → Check origin health
    │   │   └── Delegate to origin service ops as needed
    │   ├── OriginResponseTime > 5s → Check origin response time
    │   │   └── Optimize origin or enable origin shield
    │   ├── OriginUnreachable > 2% → Check network path
    │   │   └── Verify origin configuration
    │   └── Origin bandwidth > 80% of limit → Scale origin capacity
    │       └── Consider enabling CDN origin shield
    │
    ├── Is it cache-related?
    │   ├── CacheHitRatio < 60% → Check cache headers
    │   │   └── Review cache TTL settings
    │   ├── CacheMiss rate > 50% → Check cache warming
    │   │   └── Pre-populate popular content
    │   └── Stale content (> 10% age-out requests) → Check refresh mechanism
    │       └── Purge cache or adjust TTL
    │
    ├── Is it configuration-related?
    │   ├── SSL certificate expires < 30 days → Renew immediately
    │   │   └── Renew certificate
    │   ├── DNSResolution time > 5s → Check DNS configuration
    │   │   └── Optimize DNS TTL or wait for propagation
    │   └── ConfigChangeError > 1% → Review recent config changes
    │       └── Rollback CDN config changes
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

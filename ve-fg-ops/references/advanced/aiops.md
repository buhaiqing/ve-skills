# AIOps — Function Compute Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[Function Compute Alarm Triggered]
    │
    ├── Is it invocation-related?
    │   ├── High error rate → Check function logs
    │   │   └── Review recent deployments
    │   ├── Timeout → Check downstream dependencies
    │   │   └── Optimize function or increase timeout
    │   └── Throttling → Check concurrent executions
    │       └── Request quota increase or optimize
    │
    ├── Is it resource-related?
    │   ├── OOM → Increase memory or optimize code
    │   │   └── Review memory usage patterns
    │   ├── High CPU → Optimize function code
    │   │   └── Check for infinite loops or heavy computation
    │   └── Disk full → Clean up /tmp directory
    │       └── Review temp file cleanup logic
    │
    ├── Is it cold start-related?
    │   ├── Latency spike → Check function size
    │   │   └── Enable provisioned concurrency if needed
    │   └── Frequent cold starts → Review deployment package
    │       └── Minimize package size
    │
    └── Unknown → Check function execution logs
```

## Alarm Storm Handling

**Detection Criteria:**
- > 5 function errors within 1 minute
- Error rate > 5% for any function

**Suppression Workflow:**
1. Identify affected function(s)
2. Enable function logging if not already
3. Analyze error patterns in logs
4. Fix or roll back as needed
5. Monitor recovery

## Proactive Inspection Checklist

```markdown
## Function Compute Proactive Inspection — [Date]

### Performance
- [ ] Average latency within SLA
- [ ] Cold start rate < 5%
- [ ] No timeout errors

### Resource Usage
- [ ] Memory usage < 80%
- [ ] No OOM errors in past 7 days
- [ ] Execution duration stable

### Reliability
- [ ] Error rate < 1%
- [ ] Retry policies configured for critical functions
- [ ] DLQ configured for failed messages
```

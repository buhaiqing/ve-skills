# AIOps — ARK Intelligent Operations

> AIOps deep content extracted per TE-7.

## Cross-Skill Diagnosis Decision Tree

```
[ARK Alarm Triggered]
    │
    ├── Is it endpoint-related?
    │   ├── Endpoint unreachable → Check endpoint state
    │   │   ├── Stopped → Restart endpoint
    │   │   └── Quota exceeded → Request quota increase
    │   └── High latency (> 2s) → Check model type
    │       ├── Large model → Review context length
    │       └── Small model → Check concurrent requests
    │
    ├── Is it model-related?
    │   ├── Model deployment failed → Check model status
    │   │   ├── Insufficient resources → Add nodes
    │   │   └── Version mismatch → Update model version
    │   ├── Inference errors increasing → Check input format
    │   │   └── Validate request schema
    │   └── Throughput dropping → Check scaling policy
    │
    ├── Is it billing-related?
    │   ├── Usage cost > budget → Review request volume
    │   │   └── Optimize prompt length or batch requests
    │   └── PrePaid endpoint expiring → Renew or convert
    │
    └── Unknown pattern → Delegate to ve-cms-ops for correlation analysis
```

## Alarm Storm Handling

**Detection Criteria:**
- > 10 inference errors within 5 minutes
- Multiple endpoints reporting high latency simultaneously

**Suppression Workflow:**
1. Correlate by endpoint and time window
2. Identify root endpoint or model causing errors
3. Group related alarms under root endpoint
4. Address root cause → verify inference recovers

## Proactive Inspection Checklist

```markdown
## ARK Proactive Inspection — [Date]

### Endpoint Health
- [ ] All endpoints in Running state
- [ ] Inference error rate < 1%
- [ ] Average latency within service level
- [ ] No endpoints at 100% concurrent request limit

### Model & Deployment
- [ ] All models in active state
- [ ] No deprecated model versions in use
- [ ] Scaling policy configured for production endpoints
- [ ] Canary testing before full deployment

### Cost & Capacity
- [ ] Monthly inference cost within budget
- [ ] No idle endpoints running > 7 days
- [ ] Autoscaling enabled for variable workloads

### Security
- [ ] Endpoint access restricted to authorized services
- [ ] API key rotation completed within policy
```

## Multi-Round Diagnosis Review

Before finalizing any ARK diagnosis:

1. **Fact Check:** Are the endpoint metrics current? Is the model version correct?
2. **Causal Analysis:** Is the inference issue caused by endpoint configuration or upstream model quality?
3. **Solution Validation:** Will the fix resolve the issue without introducing downtime?
# Monitoring Ark

## Key Metrics

| Metric | Description |
|--------|-------------|
| InferenceLatency | Endpoint response time (ms) |
| InferenceCount | Number of inference requests |
| InferenceError | Failed inference count |
| TokenCount | Total tokens processed |
| EndpointThroughput | Requests per second |

## Alert Recommendations

| Alert | Threshold | Severity |
|-------|-----------|----------|
| High latency | p95 > 5000ms | Warning |
| Error rate | > 5% failure rate | Critical |
| Endpoint down | Endpoint status != Running | Critical |
| Quota near limit | > 80% of endpoint quota | Warning |

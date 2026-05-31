# Monitoring FunctionGraph

## Key Metrics

| Metric | Description | Namespace |
|--------|-------------|-----------|
| `InvocationCount` | Number of function invocations | `Volcengine_FunctionGraph` |
| `InvocationError` | Number of invocation errors | `Volcengine_FunctionGraph` |
| `Duration` | Function execution duration in ms | `Volcengine_FunctionGraph` |
| `ConcurrentExecutions` | Current concurrent executions | `Volcengine_FunctionGraph` |
| `ThrottleCount` | Number of throttled requests | `Volcengine_FunctionGraph` |
| `MemoryUsage` | Memory usage in MB | `Volcengine_FunctionGraph` |

## Alert Recommendations

| Alert Name | Condition | Severity |
|------------|-----------|----------|
| High Error Rate | `InvocationError / InvocationCount > 0.05` | Critical |
| High Duration | `p95(Duration) > 5000` | Warning |
| Throttling Occurring | `ThrottleCount > 0` | Warning |
| Concurrent Limit Near | `ConcurrentExecutions > 80% of quota` | Warning |

## Logging

Function logs are stored in SLS (Simple Log Service). Configure log collection during function creation.

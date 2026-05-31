# Monitoring SLS

## Key Metrics

| Metric | Description | Namespace |
|--------|-------------|-----------|
| `LogIncomingBytes` | Incoming log data volume (bytes) | `Volcengine_TLS` |
| `LogIncomingCount` | Number of log entries received | `Volcengine_TLS` |
| `LogStorageBytes` | Total stored log data (bytes) | `Volcengine_TLS` |
| `LogIndexBytes` | Index storage size (bytes) | `Volcengine_TLS` |
| `LogReadCount` | Number of log read requests | `Volcengine_TLS` |

## Alert Recommendations

| Alert Name | Condition | Severity |
|------------|-----------|----------|
| Log Volume Spike | `LogIncomingBytes > 2x baseline` | Warning |
| Log Volume Drop | `LogIncomingBytes < 0.1x baseline` | Critical |
| Storage Near Limit | `LogStorageBytes > 80% of quota` | Warning |
| Index Oversized | `LogIndexBytes > 3x LogStorageBytes` | Warning |

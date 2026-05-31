# Monitoring DNS

## Key Metrics

| Metric | Description |
|--------|-------------|
| DNSQueryCount | Number of DNS queries |
| DNSBlockedCount | Number of blocked queries (DDoS) |
| DNSLatency | Query resolution latency |

## Alert Recommendations

| Alert | Threshold | Severity |
|-------|-----------|----------|
| Query spike | > 5x baseline | Warning |
| High latency | p95 > 500ms | Warning |
| Domain resolution failure | Any failure | Critical |

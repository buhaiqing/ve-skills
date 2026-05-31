# Monitoring — Volcengine DNS

## Key Metrics

Volcengine DNS provides operational statistics through the `DescribeDomainStatistics` API.

| Metric | Description | Source |
|--------|-------------|--------|
| DNS Query Count | Total number of DNS queries | `DescribeDomainStatistics.TotalRequests` |
| Successful Resolutions | Queries that returned a valid result | `DescribeDomainStatistics.SuccessfulResolutions` |
| Failed Resolutions | Queries that failed | `DescribeDomainStatistics.FailedResolutions` |
| Query Traffic | Total traffic volume in bytes | `DescribeDomainStatistics.Traffic` |
| Peak QPS | Peak queries per second | `DescribeDomainStatistics.PeakQPS` |

## Viewing Statistics

### Using CLI

```bash
# Get basic statistics
ve dns DescribeDomainStatistics --Region "cn-beijing" --DomainName "example.com"

# Get statistics for specific time range
ve dns DescribeDomainStatistics \
  --Region "cn-beijing" \
  --DomainName "example.com" \
  --StartTime "2026-05-01T00:00:00Z" \
  --EndTime "2026-05-31T00:00:00Z"
```

### Using JIT Go SDK

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/volcengine/volc-sdk-golang/service/dns"
)

func main() {
    instance := dns.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "Region":     os.Getenv("VOLCENGINE_REGION"),
        "DomainName": os.Getenv("DOMAIN_NAME"),
        "StartTime":  "2026-05-01T00:00:00Z",
        "EndTime":    "2026-05-31T00:00:00Z",
    }

    resp, err := instance.Client.Request("dns", "DescribeDomainStatistics", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "API call failed: %v\n", err)
        os.Exit(1)
    }

    fmt.Println(string(resp))
}
```

## Alert Configuration

### Threshold-Based Alerts

| Alert Name | Metric | Threshold | Severity | Action |
|------------|--------|-----------|----------|--------|
| High Query Volume | TotalRequests | > 2x baseline | Warning | Investigate traffic source |
| Resolution Spike | PeakQPS | > 10000 QPS | Warning | Check for DDoS or traffic surge |
| Resolution Failure Spike | FailedResolutions | > 1% of total | Critical | Investigate DNS configuration |
| Traffic Anomaly | Traffic | > 5x baseline | Warning | Check for abnormal patterns |
| Zero Traffic | TotalRequests | = 0 (for active domain) | Warning | DNS may be misconfigured |

### Monitoring Frequency

| Interval | Use Case |
|----------|----------|
| Every 5 minutes | Production domains with active traffic |
| Every 1 hour | Staging or low-traffic domains |
| Daily | Batch monitoring report |

## Regular Monitoring Tasks

### Daily

```bash
# Quick health check for all domains
ve dns ListDomains --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Domains[].DomainName' | while read domain; do
  echo "=== $domain ==="
  ve dns DescribeDNSResolution --Region "{{env.VOLCENGINE_REGION}}" --DomainName "$domain"
  ve dns DescribeDomainStatistics --Region "{{env.VOLCENGINE_REGION}}" --DomainName "$domain"
done
```

### Weekly

```bash
# Export all records for backup
ve dns ListDomains --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Domains[].DomainName' | while read domain; do
  ve dns ListRecords --Region "{{env.VOLCENGINE_REGION}}" --DomainName "$domain" \
    > "dns-backup-$(date +%Y%m%d)-$domain.json"
done
```

### Monthly

- Review DNS query volume trends
- Audit and remove stale/unused records
- Review CAA record configuration
- Check for expired domains or certificates

## Integration with CMS (Cloud Monitor)

When available, DNS metrics can be integrated with Volcengine CMS (Cloud Monitor Service):

1. Navigate to CMS console
2. Create alarm rules based on DNS metrics
3. Configure notification channels (email, SMS, webhook)

Reference: `ve-cms-ops` for CMS integration details.

# Monitoring — Volcengine CDN

## Overview

Effective CDN monitoring involves tracking performance metrics, cache efficiency, origin health, and cost optimization opportunities.

## Key Metrics

### Performance Metrics

| Metric | Description | Unit | Alert Threshold |
|--------|-------------|------|-----------------|
| `bandwidth` | Data transfer rate | bps | > 80% of capacity |
| `traffic` | Total data transferred | bytes | > expected baseline |
| `requests` | HTTP request count | count | > normal pattern |
| `qps` | Queries per second | count/s | > 2x baseline |
| `response_time` | Edge response time | ms | > 500ms |
| `origin_response_time` | Origin response time | ms | > 2000ms |

### Cache Efficiency Metrics

| Metric | Description | Target | Alert Threshold |
|--------|-------------|--------|-----------------|
| `cache_hit_rate` | % served from cache | > 85% | < 70% |
| `cache_miss_rate` | % fetched from origin | < 15% | > 30% |
| `hit_traffic` | Traffic from cache | Maximize | N/A |
| `miss_traffic` | Traffic from origin | Minimize | > 2x baseline |

### Error Metrics

| Metric | Description | Target | Alert Threshold |
|--------|-------------|--------|-----------------|
| `4xx_rate` | Client error rate | < 1% | > 5% |
| `5xx_rate` | Server error rate | < 0.1% | > 1% |
| `origin_error_rate` | Origin error rate | < 0.5% | > 2% |

## Metric Collection

### Bandwidth Monitoring

```bash
#!/bin/bash
# Collect bandwidth metrics for last hour

REGION="cn-beijing"
DOMAIN="cdn.example.com"
START_TIME=$(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%SZ)
END_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Get bandwidth data (5-minute intervals)
ve cdn DescribeCdnData \
  --Region $REGION \
  --StartTime "$START_TIME" \
  --EndTime "$END_TIME" \
  --Metric bandwidth \
  --Interval 300 \
  --Domain $DOMAIN | jq '.Result.Data'
```

### Cache Hit Rate Monitoring

```bash
#!/bin/bash
# Monitor cache hit rate

REGION="cn-beijing"
DOMAIN="cdn.example.com"
START_TIME=$(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%SZ)
END_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Get hit rate data
ve cdn DescribeCdnDomainHitRate \
  --Region $REGION \
  --StartTime "$START_TIME" \
  --EndTime "$END_TIME" \
  --Domain $DOMAIN | jq '.Result.Data | {
    latest_hit_rate: (.[-1].HitRate // 0),
    avg_hit_rate: (map(.HitRate) | add / length),
    total_hit_traffic: (map(.HitTraffic) | add),
    total_miss_traffic: (map(.MissTraffic) | add)
  }'
```

### Origin Health Monitoring

```bash
#!/bin/bash
# Monitor origin pull metrics

REGION="cn-beijing"
DOMAIN="cdn.example.com"
START_TIME=$(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%SZ)
END_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Get origin bandwidth
ve cdn DescribeOriginData \
  --Region $REGION \
  --StartTime "$START_TIME" \
  --EndTime "$END_TIME" \
  --Metric bandwidth \
  --Domain $DOMAIN | jq '.Result.Data | {
    origin_bandwidth_peak: (map(.Value) | max),
    origin_bandwidth_avg: (map(.Value) | add / length)
  }'
```

## Dashboard Queries

### Daily Summary

```bash
#!/bin/bash
# Generate daily CDN summary

REGION="cn-beijing"
START_TIME=$(date -u -d '1 day ago' +%Y-%m-%dT%H:%M:%SZ)
END_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

echo "=== CDN Daily Summary ==="
echo "Period: $START_TIME to $END_TIME"

# List all domains
echo -e "\n--- Domains ---"
ve cdn ListCdnDomains --Region $REGION | jq -r '.Result.Domains[] | "\(.Domain): \(.Status)"'

# Aggregate metrics
echo -e "\n--- Aggregate Metrics ---"
for DOMAIN in $(ve cdn ListCdnDomains --Region $REGION | jq -r '.Result.Domains[].Domain'); do
  echo -e "\nDomain: $DOMAIN"

  # Traffic
  TRAFFIC=$(ve cdn DescribeCdnData --Region $REGION --StartTime "$START_TIME" --EndTime "$END_TIME" --Metric traffic --Domain $DOMAIN | jq '[.Result.Data[].Value] | add')
  echo "  Total Traffic: $(numfmt --to=iec $TRAFFIC 2>/dev/null || echo $TRAFFIC bytes)"

  # Hit rate
  HIT_RATE=$(ve cdn DescribeCdnDomainHitRate --Region $REGION --StartTime "$START_TIME" --EndTime "$END_TIME" --Domain $DOMAIN | jq '[.Result.Data[].HitRate] | add / length * 100')
  echo "  Avg Cache Hit Rate: ${HIT_RATE}%"
done
```

### Regional Distribution Analysis

```bash
#!/bin/bash
# Analyze traffic distribution by region

REGION="cn-beijing"
DOMAIN="cdn.example.com"
START_TIME=$(date -u -d '1 day ago' +%Y-%m-%dT%H:%M:%SZ)
END_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

echo "=== Regional Distribution ==="
ve cdn DescribeCdnRegionData \
  --Region $REGION \
  --StartTime "$START_TIME" \
  --EndTime "$END_TIME" \
  --Domain $DOMAIN | jq '.Result.Data | group_by(.Region) | map({
    region: .[0].Region,
    total_traffic: (map(.Traffic) | add),
    total_requests: (map(.Requests) | add)
  }) | sort_by(-.total_traffic)'
```

## Alerting Rules

### High Bandwidth Alert

```bash
#!/bin/bash
# Check if bandwidth exceeds threshold

THRESHOLD=1000000000  # 1 Gbps in bps
REGION="cn-beijing"
DOMAIN="cdn.example.com"
START_TIME=$(date -u -d '5 minutes ago' +%Y-%m-%dT%H:%M:%SZ)
END_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

MAX_BANDWIDTH=$(ve cdn DescribeCdnData \
  --Region $REGION \
  --StartTime "$START_TIME" \
  --EndTime "$END_TIME" \
  --Metric bandwidth \
  --Domain $DOMAIN | jq '[.Result.Data[].Value] | max')

if [ "$MAX_BANDWIDTH" -gt "$THRESHOLD" ]; then
  echo "ALERT: High bandwidth detected for $DOMAIN"
  echo "Peak: $MAX_BANDWIDTH bps (Threshold: $THRESHOLD bps)"
  # Send notification here
fi
```

### Low Cache Hit Rate Alert

```bash
#!/bin/bash
# Alert when cache hit rate drops below threshold

THRESHOLD=70  # 70%
REGION="cn-beijing"
DOMAIN="cdn.example.com"
START_TIME=$(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%SZ)
END_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

AVG_HIT_RATE=$(ve cdn DescribeCdnDomainHitRate \
  --Region $REGION \
  --StartTime "$START_TIME" \
  --EndTime "$END_TIME" \
  --Domain $DOMAIN | jq '[.Result.Data[].HitRate] | add / length * 100')

if (( $(echo "$AVG_HIT_RATE < $THRESHOLD" | bc -l) )); then
  echo "ALERT: Low cache hit rate for $DOMAIN"
  echo "Current: ${AVG_HIT_RATE}% (Threshold: ${THRESHOLD}%)"
  # Send notification here
fi
```

### Origin Error Rate Alert

```bash
#!/bin/bash
# Monitor origin error rates

REGION="cn-beijing"
DOMAIN="cdn.example.com"

# Check origin health by measuring response times
START_TIME=$(date -u -d '10 minutes ago' +%Y-%m-%dT%H:%M:%SZ)
END_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

ORIGIN_DATA=$(ve cdn DescribeOriginData \
  --Region $REGION \
  --StartTime "$START_TIME" \
  --EndTime "$END_TIME" \
  --Metric bandwidth \
  --Domain $DOMAIN)

# Check if origin data shows anomalies
# (Add your specific anomaly detection logic here)
echo "$ORIGIN_DATA" | jq '.Result.Data | length'
```

## Log Monitoring

### Access Log Analysis

```bash
#!/bin/bash
# Analyze CDN access logs

LOG_FILE="/var/log/cdn/access.log"

echo "=== Access Log Analysis ==="

# Top 10 requested URLs
echo -e "\nTop 10 URLs:"
grep -oE 'GET [^ ]+' $LOG_FILE | sort | uniq -c | sort -rn | head -10

# Status code distribution
echo -e "\nStatus Code Distribution:"
grep -oE '" [0-9]{3} ' $LOG_FILE | sort | uniq -c | sort -rn

# Cache status distribution
echo -e "\nCache Status:"
grep -oE 'HIT|MISS' $LOG_FILE | sort | uniq -c | sort -rn

# Top client IPs
echo -e "\nTop 10 Client IPs:"
awk '{print $2}' $LOG_FILE | sort | uniq -c | sort -rn | head -10
```

### Error Log Analysis

```bash
#!/bin/bash
# Analyze error patterns

LOG_FILE="/var/log/cdn/access.log"

echo "=== Error Analysis ==="

# 4xx errors by URL
echo -e "\n4xx Errors by URL:"
grep -E '" 4[0-9]{2} ' $LOG_FILE | awk '{print $4}' | sort | uniq -c | sort -rn | head -10

# 5xx errors by time
echo -e "\n5xx Errors by Hour:"
grep -E '" 5[0-9]{2} ' $LOG_FILE | cut -d'[' -f2 | cut -d':' -f1 | sort | uniq -c

# Slow requests (origin time > 2s)
echo -e "\nSlow Requests (>2s origin time):"
awk '$NF > 2000 {print $0}' $LOG_FILE | tail -20
```

## Cost Monitoring

### Bandwidth Cost Analysis

```bash
#!/bin/bash
# Estimate bandwidth costs

REGION="cn-beijing"
START_TIME=$(date -u -d '1 month ago' +%Y-%m-%dT%H:%M:%SZ)
END_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Pricing (example - check current pricing)
PRICE_PER_GB=0.15  # Adjust based on actual pricing

echo "=== Bandwidth Cost Analysis ==="

for DOMAIN in $(ve cdn ListCdnDomains --Region $REGION | jq -r '.Result.Domains[].Domain'); do
  TRAFFIC_BYTES=$(ve cdn DescribeCdnData --Region $REGION --StartTime "$START_TIME" --EndTime "$END_TIME" --Metric traffic --Domain $DOMAIN | jq '[.Result.Data[].Value] | add')
  TRAFFIC_GB=$(echo "scale=2; $TRAFFIC_BYTES / 1024 / 1024 / 1024" | bc)
  COST=$(echo "scale=2; $TRAFFIC_GB * $PRICE_PER_GB" | bc)

  echo "$DOMAIN: ${TRAFFIC_GB}GB - Estimated cost: \$${COST}"
done
```

### Cache Efficiency Cost Impact

```bash
#!/bin/bash
# Calculate cost impact of cache hit rate

REGION="cn-beijing"
DOMAIN="cdn.example.com"
START_TIME=$(date -u -d '1 day ago' +%Y-%m-%dT%H:%M:%SZ)
END_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Get hit rate data
HIT_RATE_DATA=$(ve cdn DescribeCdnDomainHitRate \
  --Region $REGION \
  --StartTime "$START_TIME" \
  --EndTime "$END_TIME" \
  --Domain $DOMAIN)

AVG_HIT_RATE=$(echo "$HIT_RATE_DATA" | jq '[.Result.Data[].HitRate] | add / length * 100')
TOTAL_HIT_TRAFFIC=$(echo "$HIT_RATE_DATA" | jq '[.Result.Data[].HitTraffic] | add')
TOTAL_MISS_TRAFFIC=$(echo "$HIT_RATE_DATA" | jq '[.Result.Data[].MissTraffic] | add')

echo "=== Cache Efficiency Cost Impact ==="
echo "Domain: $DOMAIN"
echo "Average Hit Rate: ${AVG_HIT_RATE}%"
echo "Cache Hit Traffic: $(echo "scale=2; $TOTAL_HIT_TRAFFIC / 1024 / 1024 / 1024" | bc) GB"
echo "Origin Pull Traffic: $(echo "scale=2; $TOTAL_MISS_TRAFFIC / 1024 / 1024 / 1024" | bc) GB"
echo ""
echo "If hit rate improved to 95%:"
NEW_MISS_TRAFFIC=$(echo "scale=0; $TOTAL_HIT_TRAFFIC * 0.05 / 0.95" | bc)
echo "Estimated origin traffic would be: $(echo "scale=2; $NEW_MISS_TRAFFIC / 1024 / 1024 / 1024" | bc) GB"
```

## SLA Monitoring

### Availability Check

```bash
#!/bin/bash
# Check CDN domain availability

DOMAIN="cdn.example.com"
TIMEOUT=10

echo "=== Availability Check ==="

# HTTP check
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" --max-time $TIMEOUT http://$DOMAIN/)
if [ "$HTTP_STATUS" = "200" ] || [ "$HTTP_STATUS" = "301" ] || [ "$HTTP_STATUS" = "302" ]; then
  echo "HTTP: OK (Status: $HTTP_STATUS)"
else
  echo "HTTP: FAIL (Status: $HTTP_STATUS)"
fi

# HTTPS check
HTTPS_STATUS=$(curl -s -o /dev/null -w "%{http_code}" --max-time $TIMEOUT https://$DOMAIN/)
if [ "$HTTPS_STATUS" = "200" ] || [ "$HTTPS_STATUS" = "301" ] || [ "$HTTPS_STATUS" = "302" ]; then
  echo "HTTPS: OK (Status: $HTTPS_STATUS)"
else
  echo "HTTPS: FAIL (Status: $HTTPS_STATUS)"
fi

# Response time check
RESPONSE_TIME=$(curl -s -o /dev/null -w "%{time_total}" --max-time $TIMEOUT https://$DOMAIN/)
echo "Response Time: ${RESPONSE_TIME}s"
```

## Performance Benchmarking

### Latency Test

```bash
#!/bin/bash
# Test CDN latency from multiple locations

DOMAIN="cdn.example.com"

echo "=== Latency Test ==="

# Test with curl timing
echo -e "\nDetailed Timing:"
curl -w "
DNS Lookup: %{time_namelookup}s
Connect: %{time_connect}s
TLS Handshake: %{time_appconnect}s
TTFB: %{time_starttransfer}s
Total: %{time_total}s
" -o /dev/null -s https://$DOMAIN/
```

### Throughput Test

```bash
#!/bin/bash
# Test CDN throughput

DOMAIN="cdn.example.com"
TEST_FILE="/large-test-file.zip"

echo "=== Throughput Test ==="

# Download test file and measure speed
curl -o /dev/null -w "Download Speed: %{speed_download} bytes/s (%{speed_download} Mbps)\n" \
  https://$DOMAIN$TEST_FILE
```

## Reporting

### Weekly Report Generation

```bash
#!/bin/bash
# Generate weekly CDN report

REGION="cn-beijing"
START_TIME=$(date -u -d '7 days ago' +%Y-%m-%dT%H:%M:%SZ)
END_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

REPORT_FILE="/tmp/cdn-weekly-report-$(date +%Y%m%d).md"

cat > $REPORT_FILE << EOF
# CDN Weekly Report

**Period:** $START_TIME to $END_TIME  
**Region:** $REGION

## Domain Summary

EOF

# Add domain details
for DOMAIN in $(ve cdn ListCdnDomains --Region $REGION | jq -r '.Result.Domains[].Domain'); do
  echo "### $DOMAIN" >> $REPORT_FILE

  # Status
  STATUS=$(ve cdn ListCdnDomains --Region $REGION --Domain $DOMAIN | jq -r '.Result.Domains[0].Status')
  echo "- **Status:** $STATUS" >> $REPORT_FILE

  # Traffic
  TRAFFIC=$(ve cdn DescribeCdnData --Region $REGION --StartTime "$START_TIME" --EndTime "$END_TIME" --Metric traffic --Domain $DOMAIN | jq '[.Result.Data[].Value] | add')
  echo "- **Total Traffic:** $(echo "scale=2; $TRAFFIC / 1024 / 1024 / 1024" | bc) GB" >> $REPORT_FILE

  # Hit rate
  HIT_RATE=$(ve cdn DescribeCdnDomainHitRate --Region $REGION --StartTime "$START_TIME" --EndTime "$END_TIME" --Domain $DOMAIN | jq '[.Result.Data[].HitRate] | add / length * 100')
  echo "- **Avg Cache Hit Rate:** ${HIT_RATE}%" >> $REPORT_FILE

  echo "" >> $REPORT_FILE
done

echo "Report generated: $REPORT_FILE"
cat $REPORT_FILE
```

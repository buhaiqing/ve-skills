# Troubleshooting Guide — Volcengine CDN

## Error Code Reference

### Domain Management Errors

| Error Code | Description | Common Causes | Recovery Actions |
|------------|-------------|---------------|------------------|
| `DomainAlreadyExists` | Domain already registered in CDN | Attempting to add duplicate domain | Use existing domain or choose different name |
| `DomainNotFound` | Specified domain not found | Wrong domain name, domain deleted, wrong region | Verify domain name and region |
| `InvalidDomain.Format` | Invalid domain name format | Missing TLD, invalid characters, IP used | Use valid FQDN (e.g., cdn.example.com) |
| `DomainStatusError` | Domain in invalid state for operation | Trying to delete online domain, modify processing domain | Stop domain first, wait for processing to complete |
| `DomainQuotaExceeded` | Domain quota limit reached | Account has max domains (default 100) | Contact support to increase quota or delete unused domains |

### Origin Configuration Errors

| Error Code | Description | Common Causes | Recovery Actions |
|------------|-------------|---------------|------------------|
| `InvalidOrigin.NotReachable` | Origin server not responding | Origin down, network issue, wrong IP/domain | Verify origin is accessible via curl/ping |
| `InvalidOriginType` | Invalid origin type specified | Typo in origin type value | Use: domain, ip, or tos |
| `OriginConnectionTimeout` | Cannot connect to origin within timeout | Origin slow, network latency, firewall blocking | Check origin health, increase timeout |
| `OriginConnectionRefused` | Connection actively refused | Origin service not running, wrong port | Verify origin service and port |
| `OriginSslError` | SSL/TLS handshake failed with origin | Invalid certificate, TLS version mismatch | Check origin SSL config |
| `OriginDnsError` | Cannot resolve origin domain | DNS misconfiguration, domain expired | Check DNS settings |

### HTTPS/Certificate Errors

| Error Code | Description | Common Causes | Recovery Actions |
|------------|-------------|---------------|------------------|
| `InvalidCertificate.Format` | Certificate not in valid PEM format | Wrong encoding, missing headers, binary file | Convert to PEM format with proper headers |
| `InvalidCertificate.Expired` | Certificate has expired | Certificate past validity date | Renew certificate |
| `CertificateMismatch.Domain` | Certificate CN/SAN doesn't match domain | Wrong certificate for domain, wildcard mismatch | Use correct certificate |
| `CertificateMismatch.Key` | Private key doesn't match certificate | Key from different pair | Use matching key |
| `InvalidPrivateKey.Format` | Private key not in valid format | Wrong encoding, encrypted key | Provide unencrypted PEM key |
| `TlsVersionNotSupported` | Specified TLS version not available | Using deprecated TLS 1.0/1.1 | Use tlsv1.2+ or tlsv1.3 |
| `TlsHandshakeError` | TLS handshake failed | Cipher mismatch, protocol error | Check TLS configuration |

### Cache Configuration Errors

| Error Code | Description | Common Causes | Recovery Actions |
|------------|-------------|---------------|------------------|
| `InvalidCacheRule.Syntax` | Cache rule syntax error | Invalid JSON, missing required fields | Check rule JSON format |
| `InvalidCacheRule.Type` | Invalid rule type | Typo in RuleType | Use: path, filetype, directory, fullpath, home |
| `InvalidCacheRule.TTL` | TTL value out of range | Negative number, exceeds max | Use 0-31536000 seconds |
| `CacheRuleConflict` | Overlapping or conflicting rules | Same priority, overlapping patterns | Adjust priorities or patterns |
| `CacheRuleLimitExceeded` | Too many cache rules | More than 50 rules | Remove unused rules, consolidate |
| `InvalidCacheRule.Priority` | Invalid priority value | Negative, duplicate, out of order | Use positive integers, unique values |

### Access Control Errors

| Error Code | Description | Common Causes | Recovery Actions |
|------------|-------------|---------------|------------------|
| `InvalidIpFilter.Format` | IP address format invalid | Wrong CIDR notation, invalid IP | Use valid IP or CIDR (e.g., 192.168.1.0/24) |
| `IpFilterLimitExceeded` | Too many IP addresses | Exceeds limit per filter | Consolidate using CIDR ranges |
| `InvalidReferer.Format` | Referer format invalid | Wildcard misuse, invalid characters | Use valid domain patterns |
| `InvalidUaFilter.Format` | User-Agent format invalid | Invalid characters, too long | Use valid string |
| `AccessControlConflict` | Conflicting access rules | Whitelist and blacklist both active | Choose one filter type |

### Cache Purge Errors

| Error Code | Description | Common Causes | Recovery Actions |
|------------|-------------|---------------|------------------|
| `QuotaExceeded.Refresh` | Daily URL purge quota exhausted | Exceeded 10,000 URL limit | Wait for next day or contact support |
| `QuotaExceeded.DirRefresh` | Daily directory purge quota exhausted | Exceeded 100 directory limit | Wait for next day |
| `InvalidUrl.Format` | URL format invalid | Missing protocol, malformed URL | Use full HTTPS URL |
| `InvalidUrl.Domain` | URL domain not in CDN | Wrong domain, domain deleted | Verify domain is in CDN |
| `InvalidUrl.Path` | URL path invalid | Special characters not encoded | URL-encode special characters |
| `TaskSubmitFailed` | Failed to submit purge task | Temporary error, rate limit | Retry after delay |
| `TaskNotFound` | Specified task ID not found | Wrong task ID, task expired | Verify task ID |

### API/Authentication Errors

| Error Code | HTTP Status | Description | Recovery Actions |
|------------|-------------|-------------|------------------|
| `Unauthorized` | 403 | Insufficient permissions | Attach CDNFullAccess IAM policy |
| `InvalidAccessKey` | 403 | Access key invalid or expired | Check VOLCENGINE_ACCESS_KEY |
| `SignatureDoesNotMatch` | 403 | Request signature invalid | Check VOLCENGINE_SECRET_KEY |
| `MissingAuthenticationToken` | 401 | Missing authentication | Provide access credentials |
| `InvalidRegion` | 400 | Region code invalid | Use valid region code |
| `Throttling` | 429 | Rate limit exceeded | Implement backoff, reduce request rate |
| `InternalError` | 500 | Internal server error | Retry with exponential backoff |
| `ServiceUnavailable` | 503 | Service temporarily unavailable | Retry after delay |

## Common Issues and Solutions

### Issue: Domain Stuck in "processing" State

**Symptoms:**
- Domain status doesn't change from `processing`
- Cannot modify domain configuration

**Causes:**
- Configuration validation in progress
- Certificate being deployed
- DNS propagation issues

**Resolution:**
```bash
# Check domain status
ve cdn ListCdnDomains --Region cn-beijing --Domain "cdn.example.com" | jq '.Result.Domains[0].Status'

# Wait and retry (can take up to 10 minutes)
for i in {1..60}; do
  STATUS=$(ve cdn ListCdnDomains --Region cn-beijing --Domain "cdn.example.com" | jq -r '.Result.Domains[0].Status')
  [ "$STATUS" != "processing" ] && break
  echo "Still processing... ($i/60)"
  sleep 10
done

# If still processing after 10 minutes, contact support
```

### Issue: HTTPS Not Working After Configuration

**Symptoms:**
- HTTPS requests fail
- Certificate error in browser
- TLS handshake fails

**Diagnosis:**
```bash
# Check HTTPS configuration
ve cdn DescribeCdnDomainDetail --Region cn-beijing --Domain "cdn.example.com" | jq '.Result.DomainConfig.Https'

# Test HTTPS connection
curl -v https://cdn.example.com 2>&1 | grep -E "(SSL|TLS|certificate|error)"

# Check certificate expiry
echo | openssl s_client -servername cdn.example.com -connect cdn.example.com:443 2>/dev/null | openssl x509 -noout -dates
```

**Resolution Steps:**
1. Verify certificate is not expired
2. Check certificate matches domain name
3. Ensure private key matches certificate
4. Verify TLS version compatibility
5. Re-upload certificate if needed

```bash
# Re-configure HTTPS
ve cdn UpdateCdnDomainHttps \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --HttpsSwitch on \
  --Certificate "$(cat new-cert.pem)" \
  --PrivateKey "$(cat new-key.pem)" \
  --TlsVersion "tlsv1.2+"
```

### Issue: Low Cache Hit Rate

**Symptoms:**
- Cache hit rate below 50%
- High origin bandwidth usage
- Increased costs

**Diagnosis:**
```bash
# Check cache hit rate
ve cdn DescribeCdnDomainHitRate \
  --Region cn-beijing \
  --StartTime "$(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%SZ)" \
  --EndTime "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --Domain "cdn.example.com" | jq '.Result.Data[].HitRate'

# Check current cache rules
ve cdn DescribeCdnDomainDetail --Region cn-beijing --Domain "cdn.example.com" | jq '.Result.DomainConfig.CacheRules'
```

**Resolution:**
1. Review cache rules for appropriate TTLs
2. Increase TTL for static content
3. Verify cache key settings
4. Check for cache-busting query parameters
5. Enable query string normalization if needed

```bash
# Update cache rules with longer TTL
ve cdn UpdateCdnDomainConfig \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --CacheRules '[
    {"RuleType":"path","RuleValue":"/*.jpg","TTL":604800,"Priority":1},
    {"RuleType":"path","RuleValue":"/*.css","TTL":604800,"Priority":2},
    {"RuleType":"path","RuleValue":"/*.js","TTL":604800,"Priority":3}
  ]'
```

### Issue: Origin Connection Failures

**Symptoms:**
- 5xx errors from CDN
- Content not loading
- Origin timeout errors

**Diagnosis:**
```bash
# Test origin connectivity
curl -I http://origin.example.com/health-check

# Check origin response time
time curl -o /dev/null -s http://origin.example.com/

# Verify DNS resolution
nslookup origin.example.com

# Check CDN origin configuration
ve cdn DescribeCdnDomainDetail --Region cn-beijing --Domain "cdn.example.com" | jq '.Result.DomainConfig.Origin'
```

**Resolution:**
1. Verify origin server is running
2. Check firewall rules allow CDN edge IPs
3. Increase origin timeout if slow
4. Verify origin protocol matches origin capability

### Issue: Purge Not Taking Effect

**Symptoms:**
- Old content still served after purge
- Purge task shows success but content unchanged

**Diagnosis:**
```bash
# Check purge task status
ve cdn DescribeContentTasks --Region cn-beijing --TaskId "your-task-id" | jq '.Result.Tasks[0].Status'

# Check cache headers
 curl -I https://cdn.example.com/path/to/file.jpg | grep -i cache
```

**Resolution:**
1. Wait for purge propagation (can take 5-10 minutes)
2. Verify URL exactly matches cached URL
3. Use directory purge for bulk updates
4. Check if content has query string variations
5. Prefetch updated content after purge

```bash
# Re-purge with exact URL
ve cdn SubmitRefreshTask \
  --Region cn-beijing \
  --Type url \
  --Urls '["https://cdn.example.com/exact/path/file.jpg"]'

# Prefetch updated content
ve cdn SubmitPreloadTask \
  --Region cn-beijing \
  --Urls '["https://cdn.example.com/exact/path/file.jpg"]'
```

### Issue: Access Control Blocking Legitimate Requests

**Symptoms:**
- 403 Forbidden errors
- Users unable to access content
- Specific IPs/regions blocked

**Diagnosis:**
```bash
# Check access control configuration
ve cdn DescribeCdnDomainDetail --Region cn-beijing --Domain "cdn.example.com" | jq '.Result.DomainConfig.AccessControl'

# Test from affected IP
curl -I https://cdn.example.com/ -H "X-Forwarded-For: affected.ip.address"
```

**Resolution:**
1. Review IP whitelist/blacklist
2. Check referer filter settings
3. Verify User-Agent filters
4. Temporarily disable access control for testing

```bash
# Disable IP filtering
ve cdn UpdateCdnDomainAccessControl \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --IpFilterType off
```

## Diagnostic Commands

### Domain Health Check

```bash
#!/bin/bash
DOMAIN="cdn.example.com"
REGION="cn-beijing"

echo "=== Domain Health Check: $DOMAIN ==="

# Check domain status
echo -e "\n1. Domain Status:"
ve cdn ListCdnDomains --Region $REGION --Domain $DOMAIN | jq '.Result.Domains[0] | {Domain, Status, Cname, CreateTime}'

# Check HTTPS config
echo -e "\n2. HTTPS Configuration:"
ve cdn DescribeCdnDomainDetail --Region $REGION --Domain $DOMAIN | jq '.Result.DomainConfig.Https.HttpsSwitch'

# Test HTTPS
echo -e "\n3. HTTPS Test:"
curl -s -o /dev/null -w "%{http_code}" https://$DOMAIN/ || echo "FAILED"

# Check cache hit rate (last hour)
echo -e "\n4. Cache Hit Rate (last hour):"
START=$(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%SZ)
END=$(date -u +%Y-%m-%dT%H:%M:%SZ)
ve cdn DescribeCdnDomainHitRate --Region $REGION --StartTime "$START" --EndTime "$END" --Domain $DOMAIN | jq '.Result.Data[0].HitRate'

echo -e "\n=== Check Complete ==="
```

### Origin Health Check

```bash
#!/bin/bash
ORIGIN="origin.example.com"

echo "=== Origin Health Check: $ORIGIN ==="

# DNS resolution
echo -e "\n1. DNS Resolution:"
nslookup $ORIGIN || echo "DNS FAILED"

# HTTP connectivity
echo -e "\n2. HTTP Test:"
curl -s -o /dev/null -w "HTTP Status: %{http_code}, Time: %{time_total}s\n" http://$ORIGIN/ || echo "HTTP FAILED"

# HTTPS connectivity (if applicable)
echo -e "\n3. HTTPS Test:"
curl -s -o /dev/null -w "HTTP Status: %{http_code}, Time: %{time_total}s\n" https://$ORIGIN/ || echo "HTTPS FAILED"

# Response headers
echo -e "\n4. Response Headers:"
curl -I http://$ORIGIN/ 2>/dev/null | head -20

echo -e "\n=== Check Complete ==="
```

## Log Analysis

### Access Log Format

CDN access logs contain the following fields:

| Field | Description | Example |
|-------|-------------|---------|
| `timestamp` | Request timestamp | `2024-05-27T12:00:00Z` |
| `client_ip` | Client IP address | `203.0.113.45` |
| `method` | HTTP method | `GET` |
| `url` | Request URL | `/images/photo.jpg` |
| `status` | HTTP status code | `200` |
| `bytes_sent` | Bytes sent to client | `15432` |
| `referer` | Referer header | `https://example.com/` |
| `user_agent` | User-Agent header | `Mozilla/5.0...` |
| `cache_status` | Cache HIT/MISS | `HIT` |
| `edge_node` | Edge node ID | `PEK-001` |
| `origin_time` | Origin response time (ms) | `150` |

### Common Log Patterns

**High Cache Miss Rate:**
```bash
# Find URLs with frequent cache misses
grep "MISS" access.log | cut -d' ' -f4 | sort | uniq -c | sort -rn | head -20
```

**5xx Errors:**
```bash
# Find 5xx errors
grep -E '" 5[0-9][0-9] ' access.log | tail -20

# Count by status code
grep -oE '" 5[0-9][0-9] ' access.log | sort | uniq -c
```

**Slow Requests:**
```bash
# Find slow origin responses (adjust threshold as needed)
awk '$NF > 1000 {print}' access.log | tail -20
```

## Escalation Procedures

### When to Contact Support

| Issue | Self-Resolution Time | Escalation |
|-------|---------------------|------------|
| Domain stuck processing | 30 minutes | Yes |
| InternalError persists | 3 retries | Yes |
| Certificate upload fails | 5 attempts | Yes |
| Origin connection works direct but fails via CDN | 15 minutes | Yes |
| Purge quota increase needed | N/A | Yes |
| Billing/cost questions | N/A | Yes |

### Support Information to Provide

When contacting support, include:

1. **Domain name** affected
2. **Region** being used
3. **Error codes** and messages
4. **Request IDs** from API responses
5. **Time range** when issue occurred
6. **Steps already attempted**
7. **Impact scope** (single domain vs multiple)

```bash
# Collect diagnostic info
ve cdn ListCdnDomains --Region cn-beijing --Domain "cdn.example.com" > /tmp/domain-info.json
echo "Request ID from last command: $(jq -r '.ResponseMetadata.RequestId' /tmp/domain-info.json)"
```

## Recovery Procedures

### Emergency Domain Disable

If domain is causing issues:

```bash
# Immediately stop domain
ve cdn StopCdnDomain --Region cn-beijing --Domain "cdn.example.com"

# Verify stopped
ve cdn ListCdnDomains --Region cn-beijing --Domain "cdn.example.com" | jq '.Result.Domains[0].Status'
```

### Rollback Configuration

When configuration changes cause issues:

```bash
# Note: Volcengine CDN does not have native versioning
# Document previous configuration before changes

# To rollback:
# 1. Re-apply previous configuration
# 2. Verify functionality
# 3. Monitor for issues

# Example: Rollback cache rules
ve cdn UpdateCdnDomainConfig \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --CacheRules '[previous-rules-json]'
```

### Certificate Emergency Replacement

```bash
# If certificate is compromised or expired:

# 1. Generate new certificate (via ACME/Let's Encrypt or CA)

# 2. Upload new certificate
ve cdn UpdateCdnDomainHttps \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --HttpsSwitch on \
  --Certificate "$(cat new-cert.pem)" \
  --PrivateKey "$(cat new-key.pem)"

# 3. Verify HTTPS works
curl -I https://cdn.example.com/

# 4. Revoke old certificate if needed
```

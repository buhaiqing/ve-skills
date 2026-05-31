# CLI Usage — Volcengine CDN

## ve CLI CDN Commands

All CDN operations are available through the `ve cdn` command prefix.

### Installation

```bash
# Download ve CLI
curl -fsSL https://github.com/volcengine/volcengine-cli/releases/latest/download/ve-linux-amd64 -o /usr/local/bin/ve
chmod +x /usr/local/bin/ve

# Verify installation
ve version

# Check CDN support
ve cdn --help
```

## Domain Management Commands

### List Domains

```bash
# List all domains
ve cdn ListCdnDomains --Region cn-beijing

# With pagination
ve cdn ListCdnDomains --Region cn-beijing --PageNum 1 --PageSize 50

# Filter by domain name
ve cdn ListCdnDomains --Region cn-beijing --Domain "cdn.example.com"

# Filter by status
ve cdn ListCdnDomains --Region cn-beijing --Status online
ve cdn ListCdnDomains --Region cn-beijing --Status offline
```

### Add Domain

```bash
# Basic domain addition
ve cdn AddCdnDomain \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --OriginDomain "origin.example.com" \
  --OriginType domain \
  --Protocol http \
  --ServiceType static

# With custom origin host header
ve cdn AddCdnDomain \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --OriginDomain "192.168.1.100" \
  --OriginType ip \
  --Protocol https \
  --ServiceType static \
  --OriginHost "origin.example.com"

# TOS bucket as origin
ve cdn AddCdnDomain \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --OriginDomain "my-bucket" \
  --OriginType tos \
  --Protocol http \
  --ServiceType static
```

### Domain Details

```bash
# Get domain configuration details
ve cdn DescribeCdnDomainDetail --Region cn-beijing --Domain "cdn.example.com"
```

### Start/Stop Domain

```bash
# Enable domain
ve cdn StartCdnDomain --Region cn-beijing --Domain "cdn.example.com"

# Disable domain
ve cdn StopCdnDomain --Region cn-beijing --Domain "cdn.example.com"
```

### Delete Domain

```bash
# Delete domain (irreversible)
ve cdn DeleteCdnDomain --Region cn-beijing --Domain "cdn.example.com"
```

## Origin Configuration Commands

### Update Origin

```bash
# Update origin domain
ve cdn UpdateCdnDomainOrigin \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --OriginDomain "new-origin.example.com" \
  --OriginType domain \
  --Protocol https \
  --OriginHost "new-origin.example.com"

# Update to IP origin
ve cdn UpdateCdnDomainOrigin \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --OriginDomain "203.0.113.10" \
  --OriginType ip \
  --Protocol http
```

## Cache Configuration Commands

### Update Cache Rules

```bash
# Set cache rules via JSON
ve cdn UpdateCdnDomainConfig \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --CacheRules '[
    {"RuleType":"path","RuleValue":"/*.jpg","TTL":86400,"Priority":1},
    {"RuleType":"path","RuleValue":"/*.css","TTL":604800,"Priority":2},
    {"RuleType":"filetype","RuleValue":"html,htm","TTL":300,"Priority":3}
  ]'
```

### Cache Rule Templates

**Static Website:**
```bash
ve cdn UpdateCdnDomainConfig \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --CacheRules '[
    {"RuleType":"path","RuleValue":"/*.jpg","TTL":2592000,"Priority":1},
    {"RuleType":"path","RuleValue":"/*.png","TTL":2592000,"Priority":2},
    {"RuleType":"path","RuleValue":"/*.css","TTL":604800,"Priority":3},
    {"RuleType":"path","RuleValue":"/*.js","TTL":604800,"Priority":4},
    {"RuleType":"filetype","RuleValue":"html,htm","TTL":300,"Priority":5}
  ]'
```

**Media Streaming:**
```bash
ve cdn UpdateCdnDomainConfig \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --CacheRules '[
    {"RuleType":"path","RuleValue":"/*.mp4","TTL":604800,"Priority":1},
    {"RuleType":"path","RuleValue":"/*.m3u8","TTL":5,"Priority":2},
    {"RuleType":"path","RuleValue":"/*.ts","TTL":86400,"Priority":3}
  ]'
```

**API/No Cache:**
```bash
ve cdn UpdateCdnDomainConfig \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --CacheRules '[
    {"RuleType":"path","RuleValue":"/api/*","TTL":0,"Priority":1}
  ]'
```

## HTTPS Configuration Commands

### Enable HTTPS

```bash
# Enable HTTPS with certificate from files
ve cdn UpdateCdnDomainHttps \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --HttpsSwitch on \
  --Certificate "$(cat cert.pem)" \
  --PrivateKey "$(cat key.pem)" \
  --TlsVersion "tlsv1.2+"

# Enable HTTP/2
ve cdn UpdateCdnDomainHttps \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --HttpsSwitch on \
  --Certificate "$(cat cert.pem)" \
  --PrivateKey "$(cat key.pem)" \
  --Http2 on
```

### Disable HTTPS

```bash
ve cdn UpdateCdnDomainHttps \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --HttpsSwitch off
```

## Access Control Commands

### IP Filtering

```bash
# Whitelist specific IPs
ve cdn UpdateCdnDomainAccessControl \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --IpFilter "192.168.1.1,192.168.1.2,10.0.0.0/8" \
  --IpFilterType whitelist

# Blacklist specific IPs
ve cdn UpdateCdnDomainAccessControl \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --IpFilter "203.0.113.0/24,198.51.100.1" \
  --IpFilterType blacklist

# Disable IP filtering
ve cdn UpdateCdnDomainAccessControl \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --IpFilterType off
```

### Referer Filtering

```bash
# Whitelist referers
ve cdn UpdateCdnDomainAccessControl \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --RefererFilter "*.example.com,example.com" \
  --RefererFilterType whitelist \
  --RefererAllowEmpty true

# Blacklist referers
ve cdn UpdateCdnDomainAccessControl \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --RefererFilter "bad-site.com,spam.com" \
  --RefererFilterType blacklist

# Disable referer filtering
ve cdn UpdateCdnDomainAccessControl \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --RefererFilterType off
```

### User-Agent Filtering

```bash
# Block specific User-Agents
ve cdn UpdateCdnDomainAccessControl \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --UaFilter "BadBot,SpamCrawler,FakeBrowser" \
  --UaFilterType blacklist
```

## Cache Purging Commands

### URL Refresh

```bash
# Refresh single URL
ve cdn SubmitRefreshTask \
  --Region cn-beijing \
  --Type url \
  --Urls '["https://cdn.example.com/image.jpg"]'

# Refresh multiple URLs
ve cdn SubmitRefreshTask \
  --Region cn-beijing \
  --Type url \
  --Urls '["https://cdn.example.com/file1.jpg","https://cdn.example.com/file2.css"]'

# Refresh with URL encoding
ve cdn SubmitRefreshTask \
  --Region cn-beijing \
  --Type url \
  --Urls '["https://cdn.example.com/path%20with%20spaces.jpg"]'
```

### Directory Refresh

```bash
# Refresh directory
ve cdn SubmitRefreshTask \
  --Region cn-beijing \
  --Type dir \
  --Dirs '["https://cdn.example.com/images/"]'

# Refresh multiple directories
ve cdn SubmitRefreshTask \
  --Region cn-beijing \
  --Type dir \
  --Dirs '["https://cdn.example.com/images/","https://cdn.example.com/css/"]'
```

### Prefetch

```bash
# Prefetch URLs
ve cdn SubmitPreloadTask \
  --Region cn-beijing \
  --Urls '["https://cdn.example.com/important/file.jpg"]'
```

### Query Task Status

```bash
# List tasks
ve cdn DescribeContentTasks --Region cn-beijing

# Filter by task ID
ve cdn DescribeContentTasks --Region cn-beijing --TaskId "refresh-abc123"

# Filter by type
ve cdn DescribeContentTasks --Region cn-beijing --Type url

# Filter by status
ve cdn DescribeContentTasks --Region cn-beijing --Status done
```

### Query Quotas

```bash
ve cdn DescribeContentQuota --Region cn-beijing
```

## Metrics and Monitoring Commands

### Bandwidth and Traffic

```bash
# Query bandwidth
ve cdn DescribeCdnData \
  --Region cn-beijing \
  --StartTime "2024-05-20T00:00:00Z" \
  --EndTime "2024-05-27T00:00:00Z" \
  --Metric bandwidth \
  --Domain "cdn.example.com"

# Query traffic
ve cdn DescribeCdnData \
  --Region cn-beijing \
  --StartTime "2024-05-20T00:00:00Z" \
  --EndTime "2024-05-27T00:00:00Z" \
  --Metric traffic \
  --Domain "cdn.example.com"

# Query requests
ve cdn DescribeCdnData \
  --Region cn-beijing \
  --StartTime "2024-05-20T00:00:00Z" \
  --EndTime "2024-05-27T00:00:00Z" \
  --Metric requests \
  --Domain "cdn.example.com"

# 5-minute intervals
ve cdn DescribeCdnData \
  --Region cn-beijing \
  --StartTime "2024-05-27T00:00:00Z" \
  --EndTime "2024-05-27T01:00:00Z" \
  --Metric bandwidth \
  --Interval 300 \
  --Domain "cdn.example.com"
```

### Origin Statistics

```bash
# Origin bandwidth
ve cdn DescribeOriginData \
  --Region cn-beijing \
  --StartTime "2024-05-20T00:00:00Z" \
  --EndTime "2024-05-27T00:00:00Z" \
  --Metric bandwidth \
  --Domain "cdn.example.com"

# Origin traffic
ve cdn DescribeOriginData \
  --Region cn-beijing \
  --StartTime "2024-05-20T00:00:00Z" \
  --EndTime "2024-05-27T00:00:00Z" \
  --Metric traffic \
  --Domain "cdn.example.com"
```

### Cache Hit Rate

```bash
ve cdn DescribeCdnDomainHitRate \
  --Region cn-beijing \
  --StartTime "2024-05-20T00:00:00Z" \
  --EndTime "2024-05-27T00:00:00Z" \
  --Domain "cdn.example.com"
```

### Regional Distribution

```bash
ve cdn DescribeCdnRegionData \
  --Region cn-beijing \
  --StartTime "2024-05-20T00:00:00Z" \
  --EndTime "2024-05-27T00:00:00Z" \
  --Domain "cdn.example.com"
```

## Compression Commands

### Enable Gzip

```bash
ve cdn UpdateCdnDomainCompression \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --GzipSwitch on \
  --GzipTypes "text/html,text/css,application/javascript,application/json"
```

### Enable Brotli

```bash
ve cdn UpdateCdnDomainCompression \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --BrotliSwitch on
```

## Common Workflows

### Add Domain with HTTPS

```bash
# Step 1: Add domain
ve cdn AddCdnDomain \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --OriginDomain "origin.example.com" \
  --OriginType domain \
  --Protocol http \
  --ServiceType static

# Step 2: Wait for domain to be online
sleep 60

# Step 3: Configure HTTPS
ve cdn UpdateCdnDomainHttps \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --HttpsSwitch on \
  --Certificate "$(cat cert.pem)" \
  --PrivateKey "$(cat key.pem)"

# Step 4: Configure cache rules
ve cdn UpdateCdnDomainConfig \
  --Region cn-beijing \
  --Domain "cdn.example.com" \
  --CacheRules '[{"RuleType":"path","RuleValue":"/*.jpg","TTL":86400,"Priority":1}]'
```

### Purge and Prefetch Workflow

```bash
# Step 1: Check quota
ve cdn DescribeContentQuota --Region cn-beijing

# Step 2: Submit refresh task
ve cdn SubmitRefreshTask \
  --Region cn-beijing \
  --Type url \
  --Urls '["https://cdn.example.com/critical-file.js"]'

# Step 3: Check task status
ve cdn DescribeContentTasks --Region cn-beijing

# Step 4: Prefetch updated content
ve cdn SubmitPreloadTask \
  --Region cn-beijing \
  --Urls '["https://cdn.example.com/critical-file.js"]'
```

## Output Formatting with jq

```bash
# Extract domain list
ve cdn ListCdnDomains --Region cn-beijing | jq '.Result.Domains[].Domain'

# Extract domain status
ve cdn ListCdnDomains --Region cn-beijing --Domain "cdn.example.com" | jq '.Result.Domains[0].Status'

# Extract CNAME
ve cdn ListCdnDomains --Region cn-beijing --Domain "cdn.example.com" | jq -r '.Result.Domains[0].Cname'

# Count domains by status
ve cdn ListCdnDomains --Region cn-beijing | jq '.Result.Domains | group_by(.Status) | map({status: .[0].Status, count: length})'

# Extract task status
ve cdn DescribeContentTasks --Region cn-beijing | jq '.Result.Tasks[] | {TaskId, Type, Status}'

# Calculate cache hit rate percentage
ve cdn DescribeCdnDomainHitRate --Region cn-beijing --StartTime "2024-05-27T00:00:00Z" --EndTime "2024-05-28T00:00:00Z" --Domain "cdn.example.com" | jq '.Result.Data[0].HitRate * 100'
```

## Environment Variables

Set these for easier command execution:

```bash
export VOLCENGINE_ACCESS_KEY="your-access-key"
export VOLCENGINE_SECRET_KEY="<masked>"
export VOLCENGINE_REGION="cn-beijing"

# Then commands can omit --Region
ve cdn ListCdnDomains
```

## Common Options

| Option | Description | Example |
|--------|-------------|---------|
| `--Region` | Region code | `cn-beijing` |
| `--Domain` | Domain name | `cdn.example.com` |
| `--PageNum` | Page number | `1` |
| `--PageSize` | Items per page | `50` |
| `--StartTime` | Start time (ISO 8601) | `2024-05-27T00:00:00Z` |
| `--EndTime` | End time (ISO 8601) | `2024-05-28T00:00:00Z` |
| `--Metric` | Metric type | `bandwidth`, `traffic`, `requests` |
| `--Interval` | Data interval (seconds) | `300` (5 minutes) |

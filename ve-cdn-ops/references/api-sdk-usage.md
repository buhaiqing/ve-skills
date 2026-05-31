# API & SDK Usage — Volcengine CDN

## API Overview

Volcengine CDN provides RESTful APIs for all operations. API version: **2021-03-01**.

### Base Endpoint

```
https://open.volcengineapi.com
```

### Authentication

All API requests require:
- `AccessKeyId` / `SecretAccessKey` (via headers or query params)
- `Region` parameter
- Request signature (handled by SDK/CLI)

### Common Headers

| Header | Description | Required |
|--------|-------------|----------|
| `Content-Type` | `application/json` | Yes (for POST/PUT) |
| `X-Date` | ISO 8601 timestamp | Yes |
| `Authorization` | Request signature | Yes |

## SDK Installation

### Go SDK

```bash
go get github.com/volcengine/volc-sdk-golang
```

### Python SDK

```bash
pip install volcengine
```

## Core API Operations

### Domain Management

#### AddCdnDomain

Creates a new CDN domain.

**Request:**
```json
{
  "Region": "cn-beijing",
  "Domain": "cdn.example.com",
  "OriginDomain": "origin.example.com",
  "OriginType": "domain",
  "Protocol": "http",
  "ServiceType": "static",
  "OriginHost": "origin.example.com"
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "20240527120000000000000000000000000",
    "Action": "AddCdnDomain",
    "Version": "2021-03-01",
    "Service": "cdn"
  },
  "Result": {
    "DomainId": "domain-123456789",
    "Domain": "cdn.example.com",
    "Cname": "cdn.example.com.volcgslb.com"
  }
}
```

#### ListCdnDomains

Lists all CDN domains.

**Request:**
```json
{
  "Region": "cn-beijing",
  "PageNum": 1,
  "PageSize": 50
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "20240527120000000000000000000000001",
    "Action": "ListCdnDomains"
  },
  "Result": {
    "Total": 5,
    "Domains": [
      {
        "DomainId": "domain-123456789",
        "Domain": "cdn.example.com",
        "Status": "online",
        "Cname": "cdn.example.com.volcgslb.com",
        "ServiceType": "static",
        "Protocol": "http",
        "OriginDomain": "origin.example.com",
        "CreateTime": "2024-05-20T10:00:00Z"
      }
    ]
  }
}
```

#### DescribeCdnDomainDetail

Gets detailed domain configuration.

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `DomainConfig.Domain` | string | Domain name |
| `DomainConfig.Status` | string | online/offline/processing/error |
| `DomainConfig.Origin` | object | Origin configuration |
| `DomainConfig.CacheRules` | array | Cache rule list |
| `DomainConfig.Https` | object | HTTPS configuration |
| `DomainConfig.AccessControl` | object | Access control settings |
| `DomainConfig.Compression` | object | Compression settings |

### Origin Configuration

#### UpdateCdnDomainOrigin

Updates origin server configuration.

**Request:**
```json
{
  "Region": "cn-beijing",
  "Domain": "cdn.example.com",
  "OriginDomain": "new-origin.example.com",
  "OriginType": "domain",
  "Protocol": "https",
  "OriginHost": "new-origin.example.com",
  "OriginPort": 443
}
```

**Origin Types:**

| Type | Description | Example |
|------|-------------|---------|
| `domain` | Domain name origin | `origin.example.com` |
| `ip` | IP address origin | `192.168.1.100` |
| `tos` | TOS bucket origin | `my-bucket` |

### Cache Configuration

#### UpdateCdnDomainCacheRule

Configures caching rules.

**Request:**
```json
{
  "Region": "cn-beijing",
  "Domain": "cdn.example.com",
  "CacheRules": [
    {
      "RuleType": "path",
      "RuleValue": "/*.jpg",
      "TTL": 86400,
      "Priority": 1,
      "CacheBehavior": "cache"
    },
    {
      "RuleType": "filetype",
      "RuleValue": "html,htm",
      "TTL": 300,
      "Priority": 2,
      "CacheBehavior": "cache"
    }
  ]
}
```

**Rule Types:**

| Type | Description | Example |
|------|-------------|---------|
| `path` | URL path pattern | `/*.jpg`, `/static/*` |
| `filetype` | File extension | `jpg,png,css` |
| `directory` | Directory path | `/images/`, `/assets/` |
| `fullpath` | Exact full path | `/api/data.json` |
| `home` | Homepage | `/` |

### HTTPS Configuration

#### UpdateCdnDomainHttps

Configures HTTPS/TLS.

**Request:**
```json
{
  "Region": "cn-beijing",
  "Domain": "cdn.example.com",
  "HttpsSwitch": "on",
  "Certificate": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
  "PrivateKey": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
  "TlsVersion": "tlsv1.2+",
  "Http2": "on"
}
```

**TLS Version Options:**

| Value | Description |
|-------|-------------|
| `tlsv1.0+` | TLS 1.0 and above |
| `tlsv1.1+` | TLS 1.1 and above |
| `tlsv1.2+` | TLS 1.2 and above |
| `tlsv1.3` | TLS 1.3 only |

### Access Control

#### UpdateCdnDomainAccessControl

Configures access control policies.

**IP Filtering:**
```json
{
  "Region": "cn-beijing",
  "Domain": "cdn.example.com",
  "IpFilterType": "whitelist",
  "IpFilter": "192.168.1.1,192.168.1.0/24,10.0.0.0/8"
}
```

**Referer Filtering:**
```json
{
  "Region": "cn-beijing",
  "Domain": "cdn.example.com",
  "RefererFilterType": "whitelist",
  "RefererFilter": "*.example.com,example.com",
  "RefererAllowEmpty": "true"
}
```

### Cache Purging

#### SubmitRefreshTask

Submits URL or directory refresh (purge) task.

**URL Refresh:**
```json
{
  "Region": "cn-beijing",
  "Type": "url",
  "Urls": [
    "https://cdn.example.com/image1.jpg",
    "https://cdn.example.com/style.css"
  ]
}
```

**Directory Refresh:**
```json
{
  "Region": "cn-beijing",
  "Type": "dir",
  "Dirs": [
    "https://cdn.example.com/images/",
    "https://cdn.example.com/assets/"
  ]
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "20240527120000000000000000000000002"
  },
  "Result": {
    "TaskId": "refresh-abc123def456",
    "SubmitCount": 2
  }
}
```

#### DescribeContentTasks

Queries purge/prefetch task status.

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "20240527120000000000000000000000003"
  },
  "Result": {
    "Tasks": [
      {
        "TaskId": "refresh-abc123def456",
        "Type": "url",
        "Status": "done",
        "SubmitTime": "2024-05-27T12:00:00Z",
        "CompleteTime": "2024-05-27T12:05:00Z",
        "SuccessCount": 2,
        "FailCount": 0
      }
    ]
  }
}

```

#### DescribeContentQuota

Queries daily purge/prefetch quotas.

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "20240527120000000000000000000000004"
  },
  "Result": {
    "RefreshQuota": 9850,
    "RefreshTotal": 10000,
    "DirQuota": 95,
    "DirTotal": 100,
    "PreloadQuota": 980,
    "PreloadTotal": 1000
  }
}
```

### Metrics and Statistics

#### DescribeCdnData

Queries CDN metrics (bandwidth, traffic, requests).

**Request:**
```json
{
  "Region": "cn-beijing",
  "StartTime": "2024-05-20T00:00:00Z",
  "EndTime": "2024-05-27T00:00:00Z",
  "Metric": "bandwidth",
  "Interval": "300",
  "Domain": "cdn.example.com"
}
```

**Metrics:**

| Metric | Description | Unit |
|--------|-------------|------|
| `bandwidth` | Bandwidth | bps |
| `traffic` | Traffic | bytes |
| `requests` | Request count | count |
| `qps` | QPS | count/s |

**Interval Options:**

| Interval | Description |
|----------|-------------|
| `60` | 1 minute |
| `300` | 5 minutes |
| `3600` | 1 hour |
| `86400` | 1 day |

#### DescribeCdnDomainHitRate

Queries cache hit ratio.

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "20240527120000000000000000000000005"
  },
  "Result": {
    "Data": [
      {
        "Time": "2024-05-27T12:00:00Z",
        "HitRate": 0.9234,
        "HitTraffic": 9876543210,
        "MissTraffic": 823456789
      }
    ]
  }
}
```

## Error Codes Reference

| Code | HTTP Status | Description | Recovery Action |
|------|-------------|-------------|-----------------|
| `InvalidDomain` | 400 | Invalid domain format | Check domain name format |
| `DomainAlreadyExists` | 400 | Domain already exists | Use different domain or modify existing |
| `DomainNotFound` | 404 | Domain not found | Verify domain exists |
| `InvalidOrigin` | 400 | Invalid origin configuration | Check origin domain/IP |
| `InvalidCertificate` | 400 | Invalid certificate format | Verify PEM format |
| `CertificateMismatch` | 400 | Certificate doesn't match domain | Check CN/SAN |
| `QuotaExceeded` | 429 | Quota exceeded | Contact support to increase |
| `InvalidCacheRule` | 400 | Invalid cache rule configuration | Check rule syntax |
| `TaskSubmitFailed` | 500 | Failed to submit purge task | Retry after delay |
| `Unauthorized` | 403 | Insufficient permissions | Check IAM policies |
| `InternalError` | 500 | Internal server error | Retry with backoff |
| `Throttling` | 429 | Rate limit exceeded | Implement backoff |
| `InvalidParameter` | 400 | Invalid parameter value | Check parameter format |

## Request Signing (Go Example)

```go
package main

import (
    "github.com/volcengine/volc-sdk-golang/base"
    "github.com/volcengine/volc-sdk-golang/service/cdn"
)

func main() {
    instance := cdn.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    // Request signing is handled automatically by SDK
    resp, err := instance.Client.Request("ListCdnDomains", nil, params)
}
```

## Pagination Handling

```go
pageNum := 1
pageSize := 50

for {
    params := map[string]interface{}{
        "Region":   region,
        "PageNum":  pageNum,
        "PageSize": pageSize,
    }

    resp, err := instance.Client.Request("ListCdnDomains", nil, params)
    // ... handle response

    // Check if there are more pages
    total := gjson.Get(string(resp), "Result.Total").Int()
    if pageNum*pageSize >= int(total) {
        break
    }
    pageNum++
}
```

## Rate Limiting

| API Category | Default Limit |
|--------------|---------------|
| Domain operations | 100/minute |
| Cache purge | 10/minute |
| Metrics query | 100/minute |
| Configuration | 50/minute |

**Recommended backoff strategy:**
- Retry 1: wait 2 seconds
- Retry 2: wait 4 seconds
- Retry 3: wait 8 seconds
- Max retries: 3

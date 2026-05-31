# Core Concepts — Volcengine CDN

## Architecture

CDN (Content Delivery Network) accelerates content delivery through globally distributed edge nodes:

```
┌─────────────────────────────────────────────────────────────┐
│                        CDN Service                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ Edge Node 1 │  │ Edge Node 2 │  │ Edge Node N         │ │
│  │ (Beijing)   │  │ (Shanghai)  │  │ (Global)            │ │
│  │ ┌─────────┐ │  │ ┌─────────┐ │  │ ┌─────────────────┐ │ │
│  │ │ Cache   │ │  │ │ Cache   │ │  │ │ Cache           │ │ │
│  │ │ Layer   │ │  │ │ Layer   │ │  │ │ Layer           │ │ │
│  │ └────┬────┘ │  │ └────┬────┘ │  │ └────────┬────────┘ │ │
│  └──────┼──────┘  └──────┼──────┘  └──────────┼──────────┘ │
│         │                │                    │            │
│         └────────────────┼────────────────────┘            │
│                          │                                 │
│                   ┌──────┴──────┐                          │
│                   │   Origin    │                          │
│                   │   Server    │                          │
│                   │ (Source)    │                          │
│                   └─────────────┘                          │
└─────────────────────────────────────────────────────────────┘
```

### Request Flow

1. **User Request** → DNS resolves CNAME to nearest edge node
2. **Edge Cache Check** → If cached, return directly (cache hit)
3. **Origin Pull** → If not cached, fetch from origin and cache (cache miss)
4. **Response** → Return content to user and store in cache

## Service Types

| Type | Description | Use Cases | Default Cache TTL |
|------|-------------|-----------|-------------------|
| `static` | Static file acceleration | Images, CSS, JS, HTML | 30 days |
| `media` | Streaming media | Video, audio streaming | 7 days |
| `download` | Large file download | Software packages, ISOs | 1 day |

## Domain States

| State | Description | Allowed Operations |
|-------|-------------|-------------------|
| `online` | Domain is active and serving traffic | Stop, Delete, Update config |
| `offline` | Domain is disabled | Start, Delete |
| `processing` | Configuration is being applied | Wait, Query status |
| `error` | Domain has configuration errors | Update config, Delete |

## Origin Types

| Type | Description | Example | Use When |
|------|-------------|---------|----------|
| `domain` | Origin server domain name | `origin.example.com` | Origin has domain name |
| `ip` | Origin server IP address | `192.168.1.100` | Origin uses IP directly |
| `tos` | TOS bucket as origin | `bucket-name` | Static content in TOS |

## Origin Protocol

| Protocol | Description | Use Case |
|----------|-------------|----------|
| `http` | HTTP only | Non-sensitive content |
| `https` | HTTPS only | Secure content origin |
| `follow` | Follow client protocol | Match client request |

## Regions and Endpoints

| Region Code | Region Name | CDN Endpoint |
|-------------|-------------|--------------|
| `cn-beijing` | 北京 | `https://cdn-cn-beijing.volcengineapi.com` |
| `cn-shanghai` | 上海 | `https://cdn-cn-shanghai.volcengineapi.com` |
| `cn-guangzhou` | 广州 | `https://cdn-cn-guangzhou.volcengineapi.com` |
| `ap-singapore` | 新加坡 | `https://cdn-ap-singapore.volcengineapi.com` |

## Caching Concepts

### Cache Key

The unique identifier for cached content. Default cache key includes:
- Domain name
- Full URL path
- Query string (optional, configurable)

### Cache Rules Priority

Rules are evaluated in priority order (1 = highest):

```
Request → Match Rule 1? → Yes → Apply TTL
              ↓ No
         Match Rule 2? → Yes → Apply TTL
              ↓ No
         Match Rule 3? → Yes → Apply TTL
              ↓ No
         Default TTL
```

### Cache Behaviors

| Behavior | Description |
|----------|-------------|
| `cache` | Store content in cache |
| `no-cache` | Always fetch from origin |
| `ignore-cache-control` | Override origin Cache-Control headers |

## HTTPS/TLS Configuration

### Certificate Requirements

- Format: PEM encoded
- Key: RSA 2048-bit or higher
- Validity: Not expired
- Match: Domain name must match certificate CN/SAN

### TLS Versions

| Version | Security Level | Browser Support |
|---------|---------------|-----------------|
| TLS 1.0 | Weak | Legacy only |
| TLS 1.1 | Weak | Legacy only |
| TLS 1.2 | Strong | Universal |
| TLS 1.3 | Strongest | Modern browsers |

**Recommendation:** Use `tlsv1.2+` for best balance of security and compatibility.

## Access Control Types

### IP Filtering

- **Whitelist**: Only allow specified IPs (all others blocked)
- **Blacklist**: Block specified IPs (all others allowed)
- Format: Single IP (`192.168.1.1`) or CIDR (`192.168.1.0/24`)

### Referer Filtering

- **Whitelist**: Only allow requests from specified referers
- **Blacklist**: Block requests from specified referers
- Supports wildcards: `*.example.com`

### User-Agent Filtering

- Block/allow based on User-Agent string
- Supports substring matching
- Common use: Block bots, crawlers

## Content Optimization

### Compression

| Algorithm | Compression Ratio | Browser Support |
|-----------|------------------|-----------------|
| Gzip | 60-80% | Universal |
| Brotli | 70-85% | Modern browsers |
| None | 0% | All |

**Recommended content types for compression:**
- `text/html`
- `text/css`
- `text/javascript`
- `application/javascript`
- `application/json`
- `text/xml`
- `application/xml`

### Image Optimization

| Feature | Description | Benefit |
|---------|-------------|---------|
| WebP conversion | Convert JPEG/PNG to WebP | 25-35% smaller |
| Responsive images | Serve different sizes per device | Bandwidth savings |
| Quality adjustment | Dynamic quality based on network | Faster loading |

## Resource Limits (Defaults)

| Resource | Default Limit |
|----------|---------------|
| Domains per account | 100 |
| Cache rules per domain | 50 |
| URL purge quota (daily) | 10,000 |
| Directory purge quota (daily) | 100 |
| Prefetch quota (daily) | 1,000 |
| Max file size (cache) | 32 GB |
| Certificate size | 2 MB |

## Metrics and Monitoring

### Key Metrics

| Metric | Description | Unit |
|--------|-------------|------|
| Bandwidth | Data transfer rate | bps (bits per second) |
| Traffic | Total data transferred | bytes |
| Requests | HTTP request count | count |
| QPS | Queries per second | count/second |
| Cache Hit Rate | % served from cache | ratio (0.0-1.0) |
| Origin Pull Rate | % fetched from origin | ratio (0.0-1.0) |

### Cache Hit Rate Benchmarks

| Hit Rate | Grade | Action |
|----------|-------|--------|
| > 95% | Excellent | Maintain |
| 90-95% | Good | Monitor |
| 80-90% | Fair | Optimize rules |
| 70-80% | Poor | Review cache config |
| < 70% | Critical | Immediate attention |

## Dependency Map

```
CDN Operations depend on:
  ├── IAM policies (ve-iam-ops) for permissions
  ├── TOS bucket (ve-tos-ops) when using TOS as origin
  ├── ECS/CLB (ve-ecs-ops/ve-clb-ops) when using as origin
  └── DNS configuration for CNAME setup
```

## FinOps — CDN Cost Optimization

### Pricing Components

| Component | Billing Basis | Optimization |
|-----------|---------------|--------------|
| Outbound traffic | GB transferred | Increase cache hit rate |
| Requests | Per 10K requests | Optimize file sizes |
| Origin pull | GB from origin | Maximize cache efficiency |
| HTTPS requests | Per 10K requests | Combine small files |

### Cost Optimization Strategies

1. **Increase Cache TTL** → Reduce origin pull
2. **Enable Compression** → Reduce transfer size
3. **Optimize Cache Rules** → Improve hit rate
4. **Use Appropriate Service Type** → Match billing to usage
5. **Prefetch Critical Content** → Reduce origin spikes

### Cost Calculation Example

```
Scenario: 10 TB/month traffic, 80% hit rate

Without optimization:
- Origin pull: 2 TB
- Cost: Base traffic cost + origin pull cost

With optimization (90% hit rate):
- Origin pull: 1 TB
- Savings: ~50% on origin pull costs
```

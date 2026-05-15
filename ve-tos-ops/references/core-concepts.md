# Core Concepts — Volcengine TOS

## Architecture

TOS (Torch Object Storage) provides S3-compatible object storage with RESTful APIs:

```
┌──────────────────────────────────────────┐
│              TOS Service                  │
│  ┌─────────┐ ┌─────────┐ ┌───────────┐  │
│  │ Bucket A│ │ Bucket B│ │ Bucket C  │  │
│  │ /logs/  │ │ /data/  │ │ /images/  │  │
│  │ obj1.gz │ │ db.bak  │ │ photo.jpg │  │
│  └─────────┘ └─────────┘ └───────────┘  │
└──────────────────────────────────────────┘
        │              │              │
   ┌────┴─────┐  ┌────┴─────┐  ┌────┴─────┐
   │ tosutil   │  │ ve CLI   │  │ Go SDK   │
   │ (bulk)    │  │ (API)    │  │ (app)    │
   └──────────┘  └──────────┘  └──────────┘
```

## Regions and Endpoints

| Region Code | Region Name | Endpoint |
|-------------|-------------|----------|
| `cn-beijing` | 北京 | `https://tos-cn-beijing.volces.com` |
| `cn-shanghai` | 上海 | `https://tos-cn-shanghai.volces.com` |
| `cn-guangzhou` | 广州 | `https://tos-cn-guangzhou.volces.com` |
| `ap-southeast-1` | 新加坡 | `https://tos-ap-southeast-1.volces.com` |

**Endpoint format:** `https://tos-{region}.volces.com`

## Storage Classes

| Class | Description | Min Storage | Retrieval Cost | Use Case |
|-------|-------------|-------------|----------------|----------|
| `Standard` | Hot data | None | Low | Frequently accessed data |
| `IA` (Infrequent Access) | Warm data | 30 days | Medium | Backup, archives |
| `Archive` | Cold data | 90 days | High (restore needed) | Compliance, long-term retention |
| `ColdArchive` | Very cold data | 180 days | Highest | Disaster recovery |

## Bucket Properties

| Property | Description |
|----------|-------------|
| **ACL** | `private`, `public-read`, `public-read-write` |
| **Versioning** | Enable/disable to keep multiple versions of objects |
| **Lifecycle** | Auto-transition or expire objects by age/prefix |
| **CORS** | Cross-origin request rules for browser access |
| **Encryption** | Server-side encryption (SSE-KMS, SSE-TOSS3) |
| **Replication** | Cross-region replication for DR |

## Object Key Conventions

- **Key format:** URL-safe string, can use `/` as delimiter to simulate folders
- **Max key length:** 1024 bytes
- **No naming restrictions** beyond valid UTF-8

## Access Methods

| Method | Tool | Best For |
|--------|------|---------|
| S3 REST API | Any S3-compatible client | Programmatic access |
| tosutil CLI | `tosutil` command | Bulk data transfer |
| ve CLI API | `ve tos` command | API-based operations |
| Go SDK | `ve-tos-golang-sdk` | Application integration |
| Python SDK | `tos-python-sdk` | Python scripts |

## Resource Limits (Defaults)

| Resource | Default Limit |
|----------|---------------|
| Buckets per account | 100 (can be increased) |
| Object size (single PUT) | 5 GB |
| Object size (multipart) | 48.8 TB |
| Part size range | 5 MB – 5 GB |
| Max parts per upload | 10,000 |

## Dependency Map

```
TOS Operations depend on:
  ├── IAM policies (ve-iam-ops) for permissions
  └── VPC endpoints (ve-vpc-ops) for private access (optional)
```

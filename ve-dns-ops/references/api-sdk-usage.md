# API & SDK Usage — Volcengine DNS

## OpenAPI Reference

- **Doc:** [Volcengine DNS OpenAPI](https://www.volcengine.com/docs/6634)
- **Base URL:** `https://open.volcengineapi.com`
- **Service Name:** `dns`

## Operations Map

| Goal | API Action | SDK Method | HTTP Method |
|------|-----------|------------|-------------|
| List domains | `ListDomains` | `dns.ListDomains` | GET |
| Describe domain | `DescribeDomain` | `dns.DescribeDomain` | GET |
| Create domain | `CreateDomain` | `dns.CreateDomain` | POST |
| Delete domain | `DeleteDomain` | `dns.DeleteDomain` | POST |
| Modify domain | `ModifyDomain` | `dns.ModifyDomain` | POST |
| Add record | `AddRecord` | `dns.AddRecord` | POST |
| Update record | `UpdateRecord` | `dns.UpdateRecord` | POST |
| Delete record | `DeleteRecord` | `dns.DeleteRecord` | POST |
| List records | `ListRecords` | `dns.ListRecords` | GET |
| Domain statistics | `DescribeDomainStatistics` | `dns.DescribeDomainStatistics` | GET |
| Resolution status | `DescribeDNSResolution` | `dns.DescribeDNSResolution` | GET |
| Batch import | `BatchImportRecords` | `dns.BatchImportRecords` | POST |

## Common Request / Response Patterns

### CreateDomain

**Request:**
```json
{
  "DomainName": "example.com"
}
```

**Response:**
```json
{
  "DomainId": "d-xxxxxxxxxxxxx"
}
```

### AddRecord

**Request:**
```json
{
  "DomainName": "example.com",
  "RR": "www",
  "Type": "A",
  "Value": "192.168.1.1",
  "TTL": 600
}
```

**Response:**
```json
{
  "RecordId": "r-xxxxxxxxxxxxx"
}
```

### ListRecords

**Response:**
```json
{
  "Records": [
    {
      "RecordId": "r-xxx",
      "RR": "www",
      "Type": "A",
      "Value": "192.168.1.1",
      "TTL": 600,
      "Status": "active"
    }
  ]
}
```

### Error Response Shape

```json
{
  "Error": {
    "Code": "InvalidDomainName",
    "Message": "The domain name format is invalid."
  }
}
```

## Request / Response Notes

### Required Fields by Operation

| Operation | Required Fields | Optional Fields |
|-----------|----------------|----------------|
| CreateDomain | `DomainName` | — |
| DeleteDomain | `DomainName` or `DomainId` | — |
| DescribeDomain | `DomainName` or `DomainId` | — |
| ModifyDomain | `DomainName` or `DomainId` | `Description` |
| AddRecord | `DomainName`, `RR`, `Type`, `Value` | `TTL`, `Priority` |
| UpdateRecord | `DomainName`, `RecordId` | `Type`, `Value`, `TTL`, `Priority` |
| DeleteRecord | `DomainName`, `RecordId` | — |
| ListRecords | `DomainName` | `Type`, `PageNumber`, `PageSize` |
| DescribeDomainStatistics | `DomainName` | `StartTime`, `EndTime` |

### Pagination

Paginated operations (like `ListDomains`, `ListRecords`) support:
- `PageNumber` — page index (default: 1)
- `PageSize` — items per page (default: 20, max: 100)

Response includes:
- `TotalCount` — total matching items
- `PageNumber` — current page
- `PageSize` — items per page

### Idempotency

- `CreateDomain` is **not** idempotent — creating an already-existing domain returns `DomainAlreadyExists`
- `AddRecord` is **not** idempotent — duplicate records return `DuplicateRecord`
- `UpdateRecord` is **idempotent** — updating with the same values is safe
- `DeleteDomain` / `DeleteRecord` are **idempotent** — deleting a non-existent resource returns `ResourceNotFound`

### Rate Limits

- API rate limits apply per account
- HTTP 429 (`Throttling`) — implement exponential backoff with `Retry-After` header
- Default limit: consult [Volcengine API Rate Limits](https://www.volcengine.com/docs/6634/rate-limits)

## Go SDK Package

```go
import "github.com/volcengine/volc-sdk-golang/service/dns"
```

### SDK Initialization

```go
instance := dns.NewInstance()
instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))
```

### Generic Request Pattern

```go
params := map[string]interface{}{
    "Region":     os.Getenv("VOLCENGINE_REGION"),
    "DomainName": "example.com",
}

resp, err := instance.Client.Request("dns", "AddRecord", params)
if err != nil {
    // handle error
}
fmt.Println(string(resp))
```

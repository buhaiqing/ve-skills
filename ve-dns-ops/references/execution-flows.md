## Execution Flows (Agent-Readable)

Every operation: **Pre-flight → Execute → Validate → Recover**. Do not skip phases.

**Preference hint:** CLI is preferred for coverage and simplicity; Go SDK is used for operations CLI does not expose.

---

### Operation: ListDomains — List All DNS Domains

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | `test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY"` | Both set | HALT; configure credentials |
| Region | Verify `{{env.VOLCENGINE_REGION}}` is set | Valid region | HALT; set VOLCENGINE_REGION |
| CLI installed | `ve version` | Exit code 0 | Install ve CLI |

#### Execution — CLI (`ve`) (Primary Path)

```bash
# List all domains
ve dns ListDomains --Region "{{env.VOLCENGINE_REGION}}"

# List with pagination (if API supports)
ve dns ListDomains --Region "{{env.VOLCENGINE_REGION}}" --Limit 20
```

#### Execution — JIT Go SDK (Fallback Path)

```go
// main.go (generated dynamically in /tmp/ve-sdk-workspace)
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
        "Region": os.Getenv("VOLCENGINE_REGION"),
        "Limit":  20,
    }

    resp, err := instance.Client.Request("dns", "ListDomains", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "API call failed: %v\n", err)
        os.Exit(1)
    }

    fmt.Println(string(resp))
}
```

#### Validation

1. Parse `$.Domains[]` array from response
2. Count total domains: `echo "$RESPONSE" | jq '.Domains | length'`
3. Verify each domain has expected fields (DomainId, DomainName, Status)
4. Present in tabular format to user

```bash
# Pretty-print domains
ve dns ListDomains --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Domains[] | [.DomainId, .DomainName, .Status] | @tsv'
```

#### Failure Recovery

| Error Pattern | Max Retries | Backoff | Agent Action | UX Template |
|--------------|-------------|---------|--------------|-------------|
| `InvalidParameter` | 0 | — | Fix pagination params | `[ERROR] InvalidParameter: Check pagination parameters.` |
| `InternalError` | 3 | 2s, 4s, 8s | Retry with backoff | `[ERROR] InternalError: Server-side error. Retrying...` |

---

### Operation: CreateDomain — Add a New Domain

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | Env vars set | Both set | HALT |
| Region | Env var set | Valid | HALT |
| Domain format | Validate FQDN format | Valid DNS name | HALT; provide correct format |
| Domain available | `ve dns ListDomains` | Domain not already added | HALT; domain already exists |
| Quota | Check account quota | Under limit | HALT; request quota increase |

#### Execution — CLI (`ve`)

```bash
ve dns CreateDomain \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --DomainName "{{user.domain}}"
```

#### Execution — JIT Go SDK (Fallback)

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
    }

    resp, err := instance.Client.Request("dns", "CreateDomain", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "API call failed: %v\n", err)
        os.Exit(1)
    }

    fmt.Println(string(resp))
}
```

#### Post-execution Validation

1. Parse `{{output.domain_id}}` from response path `$.DomainId`
2. Poll until domain status is `active`:

```bash
for i in $(seq 1 12); do
  STATUS=$(ve dns DescribeDomain --Region "{{env.VOLCENGINE_REGION}}" --DomainId "{{output.domain_id}}" | jq -r '.Status')
  echo "Status: $STATUS (attempt $i/12)"
  [ "$STATUS" = "active" ] && break
  sleep 5
done
```

3. List domains to confirm: `ve dns ListDomains --Region "{{env.VOLCENGINE_REGION}}" | jq '.Domains[] | select(.DomainName=="{{user.domain}}")'`
4. Report `{{output.domain_id}}` and `{{user.domain}}` to user

#### Failure Recovery

| Error Pattern | Max Retries | Backoff | Agent Action | UX Template |
|--------------|-------------|---------|--------------|-------------|
| `InvalidDomainName` | 0 | — | HALT; correct domain format | `[ERROR] Invalid domain name format. Use FQDN (e.g., example.com).` |
| `DomainAlreadyExists` | 0 | — | HALT; domain already added | `[ERROR] Domain already exists in your account. Check ListDomains.` |
| `QuotaExceeded` | 0 | — | HALT; raise quota | `[ERROR] Domain quota exceeded. Request increase from support.` |
| `InsufficientBalance` | 0 | — | HALT; recharge | `[ERROR] Insufficient balance. Recharge your account.` |
| `InternalError` | 3 | 2s, 4s, 8s | Retry with backoff | `[ERROR] Internal server error. Retrying (attempt {n}/3)...` |
| `Throttling` / 429 | 3 | 1s, 2s, 4s | Back off and retry | `[WARNING] Rate limited. Retrying in {n}s...` |

---

### Operation: DescribeDomain — View Domain Details

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Domain exists | `ve dns ListDomains` | DomainId or DomainName valid | HALT |

#### Execution

```bash
# By DomainName (if supported)
ve dns DescribeDomain \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --DomainName "{{user.domain}}"

# By DomainId
ve dns DescribeDomain \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --DomainId "{{user.domain_id}}"
```

#### Present to User

| Field | Path | Notes |
|-------|------|-------|
| Domain ID | `$.DomainId` | Internal ID |
| Domain Name | `$.DomainName` | FQDN |
| Status | `$.Status` | `active`, `pending`, etc. |
| Created At | `$.CreatedAt` | ISO 8601 |
| Record Count | `$.RecordCount` | Number of records |
| DNS Resolution | `$.DNSResolution` | Resolution status |

---

### Operation: DeleteDomain — Remove a Domain (Irreversible)

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: irreversible delete of domain `{{user.domain}}` (`{{user.domain_id}}`).
- **MUST NOT** proceed without clear user assent.
- **MUST** warn: all DNS records under this domain will be permanently deleted.
- **MUST** list records before deleting so user is aware:

```bash
# List all records before deletion
ve dns ListRecords --Region "{{env.VOLCENGINE_REGION}}" --DomainName "{{user.domain}}"

echo "WARNING: Deleting domain '{{user.domain}}' will permanently remove ALL DNS records."
echo "Total records: $(ve dns ListRecords --Region "{{env.VOLCENGINE_REGION}}" --DomainName "{{user.domain}}" | jq '.Records | length')"
```

#### Execution

```bash
# Delete the domain
ve dns DeleteDomain \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --DomainName "{{user.domain}}"
```

#### Post-execution Validation

```bash
# Verify domain is deleted (ListDomains should no longer show it)
if ve dns ListDomains --Region "{{env.VOLCENGINE_REGION}}" | jq -e '.Domains[] | select(.DomainName=="'"{{user.domain}}"'")' > /dev/null 2>&1; then
  echo "Domain may still exist — check manually"
else
  echo "Domain deleted successfully"
fi
```

#### Failure Recovery

| Error Pattern | Max Retries | Backoff | Agent Action | UX Template |
|--------------|-------------|---------|--------------|-------------|
| `DomainNotFound` | 0 | — | Domain already deleted; skip | `[INFO] Domain not found — may already be deleted.` |
| `DependencyViolation` | 0 | — | HALT; check for dependencies | `[ERROR] Domain has dependencies. Check if DNS is in use.` |
| `Unauthorized` | 0 | — | HALT; IAM permission | `[ERROR] Unauthorized. Check IAM policies for dns:DeleteDomain.` |
| `InternalError` | 3 | 2s, 4s, 8s | Retry | `[ERROR] Internal error. Retrying...` |

---

### Operation: ModifyDomain — Update Domain Configuration

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Domain exists | `ve dns DescribeDomain` | Domain found | HALT |

#### Execution

```bash
# Modify domain (e.g., update related configuration)
ve dns ModifyDomain \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --DomainName "{{user.domain}}" \
  --Description "{{user.description}}"
```

#### Validation

```bash
# Verify domain configuration updated
ve dns DescribeDomain --Region "{{env.VOLCENGINE_REGION}}" --DomainName "{{user.domain}}"
```

---

### Operation: AddRecord — Add a DNS Record

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Domain exists | `ve dns DescribeDomain` | Domain active | HALT |
| Record type valid | Validate type | A, AAAA, CNAME, MX, TXT, NS, SRV, CAA | HALT; provide valid type |
| Record value format | Validate per type | Correct format | HALT; fix format |
| Duplicate check | `ve dns ListRecords` | No conflicting record | HALT or warn |
| TTL valid | Range check | 1-86400 | Use default 600 |

#### Execution — CLI (`ve`)

```bash
# Add an A record
ve dns AddRecord \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --DomainName "{{user.domain}}" \
  --RR "www" \
  --Type "A" \
  --Value "192.168.1.1" \
  --TTL 600

# Add a CNAME record
ve dns AddRecord \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --DomainName "{{user.domain}}" \
  --RR "mail" \
  --Type "CNAME" \
  --Value "mail.example.com" \
  --TTL 600

# Add an MX record with priority
ve dns AddRecord \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --DomainName "{{user.domain}}" \
  --RR "@" \
  --Type "MX" \
  --Value "mail.example.com" \
  --Priority 10 \
  --TTL 600

# Add a TXT record (e.g., for SPF or domain verification)
ve dns AddRecord \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --DomainName "{{user.domain}}" \
  --RR "@" \
  --Type "TXT" \
  --Value "v=spf1 include:_spf.example.com ~all" \
  --TTL 600

# Add an AAAA record (IPv6)
ve dns AddRecord \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --DomainName "{{user.domain}}" \
  --RR "www" \
  --Type "AAAA" \
  --Value "2001:db8::1" \
  --TTL 600

# Add an NS record (delegation)
ve dns AddRecord \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --DomainName "{{user.domain}}" \
  --RR "subdomain" \
  --Type "NS" \
  --Value "ns1.example.com" \
  --TTL 86400

# Add a SRV record
ve dns AddRecord \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --DomainName "{{user.domain}}" \
  --RR "_sip._tcp" \
  --Type "SRV" \
  --Value "10 60 5060 sip.example.com" \
  --TTL 600

# Add a CAA record
ve dns AddRecord \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --DomainName "{{user.domain}}" \
  --RR "@" \
  --Type "CAA" \
  --Value "0 issue "letsencrypt.org"" \
  --TTL 600
```

#### Execution — JIT Go SDK (Fallback)

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
        "RR":         os.Getenv("RR"),
        "Type":       os.Getenv("RECORD_TYPE"),
        "Value":      os.Getenv("RECORD_VALUE"),
        "TTL":        600,
    }

    resp, err := instance.Client.Request("dns", "AddRecord", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "API call failed: %v\n", err)
        os.Exit(1)
    }

    fmt.Println(string(resp))
}
```

#### Post-execution Validation

1. Parse `{{output.record_id}}` from `$.RecordId`
2. Verify record in list:

```bash
# Verify the record was added
ve dns ListRecords --Region "{{env.VOLCENGINE_REGION}}" --DomainName "{{user.domain}}" \
  | jq '.Records[] | select(.RecordId=="'"{{output.record_id}}"'")'
```

3. Test resolution (after propagation):

```bash
# Test DNS resolution from public DNS (if A/CNAME record)
dig @8.8.8.8 "{{user.rr}}.{{user.domain}}" +short
nslookup "{{user.rr}}.{{user.domain}}" 8.8.8.8
```

#### Failure Recovery

| Error Pattern | Max Retries | Backoff | Agent Action | UX Template |
|--------------|-------------|---------|--------------|-------------|
| `InvalidRecordType` | 0 | — | HALT; use valid type | `[ERROR] Invalid record type. Use A, AAAA, CNAME, MX, TXT, NS, SRV, or CAA.` |
| `InvalidRecordValue` | 0 | — | HALT; fix value format | `[ERROR] Invalid record value for type {{user.record_type}}.` |
| `DuplicateRecord` | 0 | — | HALT; record exists | `[ERROR] Duplicate record. Use UpdateRecord to modify.` |
| `DomainNotFound` | 0 | — | HALT; create domain first | `[ERROR] Domain not found. Create domain first.` |
| `RecordLimitExceeded` | 0 | — | HALT; record quota reached | `[ERROR] Record limit exceeded for this domain.` |
| `InternalError` | 3 | 2s, 4s, 8s | Retry with backoff | `[ERROR] Internal error. Retrying...` |

---

### Operation: UpdateRecord — Modify an Existing DNS Record

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Domain exists | `ve dns DescribeDomain` | Domain active | HALT |
| Record exists | `ve dns ListRecords` | RecordId valid | HALT |
| Record type unchanged | Compare to existing | Same type | HALT; delete and re-add for type change |

#### Execution — CLI (`ve`)

```bash
# Update a record value
ve dns UpdateRecord \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --DomainName "{{user.domain}}" \
  --RecordId "{{user.record_id}}" \
  --Type "A" \
  --Value "10.0.0.1" \
  --TTL 600

# Update only TTL
ve dns UpdateRecord \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --DomainName "{{user.domain}}" \
  --RecordId "{{user.record_id}}" \
  --TTL 300
```

#### Post-execution Validation

```bash
# Verify the record was updated
ve dns ListRecords --Region "{{env.VOLCENGINE_REGION}}" --DomainName "{{user.domain}}" \
  | jq '.Records[] | select(.RecordId=="'"{{user.record_id}}"'") | {RecordId, Type, Value, TTL}'
```

#### Failure Recovery

| Error Pattern | Max Retries | Backoff | Agent Action | UX Template |
|--------------|-------------|---------|--------------|-------------|
| `RecordNotFound` | 0 | — | HALT; verify RecordId | `[ERROR] Record not found. Verify RecordId.` |
| `RecordTypeMismatch` | 0 | — | HALT; cannot change type | `[ERROR] Cannot change record type. Delete and re-add if needed.` |
| `InvalidRecordValue` | 0 | — | HALT; fix value | `[ERROR] Invalid record value.` |

---

### Operation: DeleteRecord — Remove a DNS Record

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Domain exists | `ve dns DescribeDomain` | Domain active | HALT |
| Record exists | `ve dns ListRecords` | RecordId valid | HALT |

#### Execution — CLI (`ve`)

```bash
# Get the current record for confirmation
echo "Current record:"
ve dns ListRecords --Region "{{env.VOLCENGINE_REGION}}" --DomainName "{{user.domain}}" \
  | jq '.Records[] | select(.RecordId=="'"{{user.record_id}}"'")'

# Prompt user to confirm before deleting
# (User must confirm the RecordId and value)

ve dns DeleteRecord \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --DomainName "{{user.domain}}" \
  --RecordId "{{user.record_id}}"
```

#### Post-execution Validation

```bash
# Verify the record is gone
if ve dns ListRecords --Region "{{env.VOLCENGINE_REGION}}" --DomainName "{{user.domain}}" \
  | jq -e '.Records[] | select(.RecordId=="'"{{user.record_id}}"'")' > /dev/null 2>&1; then
  echo "WARNING: Record may still exist"
else
  echo "Record deleted successfully"
fi
```

#### Failure Recovery

| Error Pattern | Max Retries | Backoff | Agent Action | UX Template |
|--------------|-------------|---------|--------------|-------------|
| `RecordNotFound` | 0 | — | Record already deleted; skip | `[INFO] Record not found — may already be deleted.` |
| `DomainNotFound` | 0 | — | HALT; verify domain | `[ERROR] Domain not found.` |

---

### Operation: ListRecords — List All Records for a Domain

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Domain exists | `ve dns ListDomains` | Domain found | HALT |

#### Execution — CLI (`ve`)

```bash
# List all records for a domain
ve dns ListRecords --Region "{{env.VOLCENGINE_REGION}}" --DomainName "{{user.domain}}"

# Filter by record type
ve dns ListRecords --Region "{{env.VOLCENGINE_REGION}}" --DomainName "{{user.domain}}" --Type "A"

# Get count of records
ve dns ListRecords --Region "{{env.VOLCENGINE_REGION}}" --DomainName "{{user.domain}}" | jq '.Records | length'
```

#### Present to User (Tabular Format)

```bash
ve dns ListRecords --Region "{{env.VOLCENGINE_REGION}}" --DomainName "{{user.domain}}" \
  | jq -r '.Records[] | [.RecordId, .RR, .Type, .Value, .TTL, .Status] | @tsv'
```

#### Failure Recovery

| Error Pattern | Max Retries | Backoff | Agent Action | UX Template |
|--------------|-------------|---------|--------------|-------------|
| `DomainNotFound` | 0 | — | HALT; verify domain name | `[ERROR] Domain not found.` |
| `InternalError` | 3 | 2s, 4s, 8s | Retry | `[ERROR] Internal error. Retrying...` |

---

### Operation: DescribeDomainStatistics — View Resolution Statistics

#### Execution

```bash
# Get domain statistics
ve dns DescribeDomainStatistics \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --DomainName "{{user.domain}}"

# Get statistics for a specific time period
ve dns DescribeDomainStatistics \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --DomainName "{{user.domain}}" \
  --StartTime "2026-05-01T00:00:00Z" \
  --EndTime "2026-05-31T00:00:00Z"
```

#### Present to User

| Field | Path | Description |
|-------|------|-------------|
| Total Requests | `$.TotalRequests` | Total DNS queries |
| Successful Resolutions | `$.SuccessfulResolutions` | Successful responses |
| Failed Resolutions | `$.FailedResolutions` | Failed queries |
| Traffic | `$.Traffic` | DNS query traffic volume |
| Peak QPS | `$.PeakQPS` | Peak queries per second |

---

### Operation: DescribeDNSResolution — Get Resolution Status

#### Execution

```bash
# Get resolution details for a domain
ve dns DescribeDNSResolution \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --DomainName "{{user.domain}}"
```

#### Present to User

| Field | Path | Description |
|-------|------|-------------|
| Resolution Status | `$.ResolutionStatus` | Current resolution health |
| DNS Servers | `$.DNSServers` | Authoritative name servers |
| Last Check | `$.LastCheckTime` | Last resolution check timestamp |

---

### Operation: BatchImportRecords — Import Records in Bulk

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Domain exists | `ve dns DescribeDomain` | Domain active | HALT |
| File format valid | Validate input format | Correct CSV/JSON | HALT; provide template |
| Record count | Count input records | Under domain quota | HALT; reduce batch size |

#### Execution

```bash
# Import records from file (format depends on API spec)
ve dns BatchImportRecords \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --DomainName "{{user.domain}}" \
  --Body '{"Records": [{"RR":"www","Type":"A","Value":"1.2.3.4","TTL":600}]}'
```

#### Validation

```bash
# Verify total record count after import
ve dns ListRecords --Region "{{env.VOLCENGINE_REGION}}" --DomainName "{{user.domain}}" | jq '.Records | length'
```

---

### Operation: BatchExportRecords — Export Records in Bulk

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Domain exists | `ve dns DescribeDomain` | Domain active | HALT |
| Records exist | `ve dns ListRecords` | At least 1 record | HALT; domain has no records |

#### Execution

```bash
# Export all records for a domain to JSON
ve dns ListRecords --Region "{{env.VOLCENGINE_REGION}}" --DomainName "{{user.domain}}" | jq '.Records' > dns-records-{{user.domain}}.json

# Export as formatted table (for display)
echo "=== DNS Records for {{user.domain}} ==="
echo ""
echo "| Type | Name | Value | TTL | Priority |"
echo "|------|------|-------|-----|----------|"
ve dns ListRecords --Region "{{env.VOLCENGINE_REGION}}" --DomainName "{{user.domain}}" | jq -r '.Records[] | "|\(.Type) | \(.RR) | \(.Value) | \(.TTL // "-") | \(.Priority // "-") |"'
```

#### Present to User

```bash
# Summary of exported records
TOTAL=$(ve dns ListRecords --Region "{{env.VOLCENGINE_REGION}}" --DomainName "{{user.domain}}" | jq '.Records | length')
echo "Exported $TOTAL records for domain {{user.domain}}"
```

#### Failure Recovery

| Error pattern | Max retries | Recovery |
|--------------|-------------|---------|
| `DomainNotFound` | 0 | HALT; verify domain — `[ERROR] Domain not found.` |
| `InvalidParameter` | 1 | Fix parameter — `[ERROR] Invalid parameter.` |

---


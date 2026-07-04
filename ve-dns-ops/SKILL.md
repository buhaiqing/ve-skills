---
name: ve-dns-ops
description: >-
  Use when the user needs to configure, troubleshoot, or manage Volcengine (火山引擎)
  DNS (DNS解析 / DNS Resolution) — domain management, record set management,
  resolution policies, DNS security, and monitoring. User mentions DNS, 域名解析,
  domain, record set, A record, CNAME, MX, TXT, NS, resolution, or describes
  domain resolution scenarios even without naming the product directly.
  Not for CDN, load balancer, or web application firewall.
license: MIT
compatibility: >-
  Official Volcengine CLI (`ve`, Go binary, no runtime), Go 1.21+ runtime
  (JIT SDK fallback; scripts compatible with Go 1.14+ syntax), valid API credentials,
  network access to Volcengine endpoints (open.volcengineapi.com).
metadata:
  author: volcengine
  version: "1.1.0"
  last_updated: "2026-06-04"
  runtime: Harness AI Agent, Claude Code, Cursor, or compatible Agent runtimes
  go_script_syntax_minimum: "1.14"
  go_jit_runtime_version: "1.21+"
  api_profile: "DNS OpenAPI — https://www.volcengine.com/docs/6634"
  cli_applicability: dual-path
  cli_support_evidence: >-
    Confirmed via `ve dns --help` — DNS is supported by the ve CLI.
    Service ID: dns.
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

# Volcengine DNS (DNS解析) Operations Skill

> This skill follows the [Agent Skill OpenSpec](https://agentskills.io/specification).

## Overview

DNS (DNS解析) on Volcengine (火山引擎) provides domain name resolution services with high availability, low latency, and anti-DDoS protection. This skill is an **operational runbook** for agents with **dual-path execution**: `ve` CLI for API calls and JIT Go SDK fallback. **Do not use the web console as the primary agent execution path.**

### CLI applicability

- **`cli_applicability: dual-path`:** Official `ve` CLI supports DNS operations.
  - **`ve dns`**: Domain management, record set CRUD, resolution management
  - Covered operations: ListDomains, DescribeDomain, CreateDomain, DeleteDomain, AddRecord, UpdateRecord, DeleteRecord, ListRecords, DescribeDomainStatistics, DescribeDNSResolution, ModifyDomain

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use conditions with precise triggers and delegation rules |
| 2 | **Structured I/O** | `{{env.VOLCENGINE_*}}` (env), `{{user.*}}` (interactive), `{{output.*}}` (API response) |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | Error taxonomy with DNS-specific codes; HALT vs retry per type |
| 5 | **Absolute Single Responsibility** | One product (DNS); cross-product delegation documented |

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- User mentions "Volcengine DNS", "DNS解析", "domain", "record set", "DNS", "resolution", "A record", "CNAME", "MX", "TXT", "NS", "SRV", "CAA"
- Task involves **domain operations**: CreateDomain, ListDomains, DescribeDomain, DeleteDomain, ModifyDomain
- Task involves **record set operations**: AddRecord, UpdateRecord, DeleteRecord, ListRecords
- Task involves **resolution management**: DescribeDNSResolution, resolution policies
- Task involves **DNS security**: anti-DDoS, rate limiting, DNS security settings
- Task involves **monitoring**: DescribeDomainStatistics, resolution volume, traffic analysis
- Task involves **batch operations**: importing/exporting DNS records
- User asks to configure, troubleshoot, or manage DNS **via API, SDK, CLI, or automation**

### SHOULD NOT Use This Skill When

- Task is about **Content Delivery Network (CDN)** → delegate to: `ve-cdn-ops` (when present)
- Task is about **Cloud Load Balancer (CLB)** → delegate to: `ve-clb-ops`
- Task is about **Web Application Firewall (WAF)** → delegate to: `ve-waf-ops` (when present)
- Task is about **ECS instance management** → delegate to: `ve-ecs-ops`
- Task is purely billing / account management → delegate to billing ops
- User insists on **console-only** flows with no API → state limitation; do not invent undocumented HTTP steps

### Delegation Rules

- DNS domains and record sets are independent resources (no VPC/subnet dependency)
- If a task involves DNS + ECS (e.g., pointing DNS to ECS IP), complete the ECS step via `ve-ecs-ops` first, then create the DNS record
- For HTTPS certificates referenced by DNS, delegate to `ve-ssl-ops` (when present)

## Variable Convention (Agent-Readable)

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | Access key from runtime | NEVER ask user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | Secret key from runtime | NEVER ask user; fail if unset; **ALWAYS mask as `<masked>`** |
| `{{env.VOLCENGINE_REGION}}` | Region from runtime (e.g., cn-beijing) | Use documented default if skill allows |
| `{{user.domain}}` | Domain name (e.g., example.com) | Ask once; validate DNS format |
| `{{user.domain_id}}` | Domain ID from API | Ask once; format from API response |
| `{{user.record_id}}` | Record ID from API | Ask once; format from API response |
| `{{user.record_type}}` | DNS record type | Ask; one of A, AAAA, CNAME, MX, TXT, NS, SRV, CAA |
| `{{user.record_value}}` | Record value | Ask; validate per record type |
| `{{user.ttl}}` | Time-to-live in seconds | Ask; default 600; range 1-86400 |
| `{{user.priority}}` | Priority (MX/SRV only) | Ask for MX, SRV records; omit otherwise |
| `{{output.domain_id}}` | Domain ID from create response | Parse from `$.DomainId` |
| `{{output.record_id}}` | Record ID from add response | Parse from `$.RecordId` |
| `{{output.request_id}}` | Request ID for tracing | Parse from response |

> **Security Warning (Credential Masking — MANDATORY):** NEVER echo or log `VOLCENGINE_SECRET_KEY` or any credential values. Verify existence only with `test -n "$VOLCENGINE_SECRET_KEY"`.

## API and Response Conventions (Agent-Readable)

- **DNS uses RESTful API** with JSON responses
- **Endpoint:** `open.volcengineapi.com` (default Volcengine API endpoint)
- **API Version:** See [DNS OpenAPI docs](https://www.volcengine.com/docs/6634)
- **Go SDK:** `github.com/volcengine/volc-sdk-golang/service/dns`
- **Error responses:** JSON with `Error` object containing `Code` and `Message`

### Key Response Fields

| Operation | Response Field | Type | Description |
|-----------|---------------|------|-------------|
| CreateDomain | `$.DomainId` | string | Created domain ID |
| ListDomains | `$.Domains[]` | array | Domain list |
| ListDomains | `$.Domains[].DomainId` | string | Domain ID |
| ListDomains | `$.Domains[].DomainName` | string | Domain name |
| ListDomains | `$.Domains[].Status` | string | Domain status |
| ListDomains | `$.Domains[].CreatedAt` | string | Creation timestamp |
| DescribeDomain | `$.DomainId` | string | Domain ID |
| DescribeDomain | `$.DomainName` | string | Domain name |
| DescribeDomain | `$.Status` | string | Domain status |
| AddRecord | `$.RecordId` | string | Created record ID |
| ListRecords | `$.Records[]` | array | Record list |
| ListRecords | `$.Records[].RecordId` | string | Record ID |
| ListRecords | `$.Records[].Type` | string | Record type (A, CNAME, MX, etc.) |
| ListRecords | `$.Records[].Value` | string | Record value |
| ListRecords | `$.Records[].TTL` | integer | TTL in seconds |
| ListRecords | `$.Records[].Status` | string | Record status |
| DescribeDomainStatistics | `$.Traffic` | number | DNS resolution traffic |
| DescribeDomainStatistics | `$.RequestCount` | number | Total request count |

### Expected State Transitions

| Operation | Initial State | Target State | Poll Interval | Max Wait |
|-----------|---------------|--------------|---------------|----------|
| CreateDomain | — | `active` | 5s | 60s |
| ModifyDomain | `active` | `active` | 5s | 30s |
| DeleteDomain | any stable | absent | 5s | 60s |
| AddRecord | — | `active` | 5s | 30s |
| UpdateRecord | `active` | `active` | 5s | 30s |
| DeleteRecord | any stable | absent | 5s | 30s |

## Quick Start

### What This Skill Does

This skill enables you to manage Volcengine DNS domains and record sets — create and delete domains, manage A/AAAA/CNAME/MX/TXT/NS/SRV/CAA records, monitor resolution statistics, and troubleshoot DNS issues using the `ve` CLI (primary) or JIT Go SDK (fallback).

### Prerequisites

- [ ] `ve` CLI installed (or Go runtime for JIT fallback)
- [ ] Credentials configured: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`
- [ ] Region set: `VOLCENGINE_REGION`

### Verify Setup

```bash
ve dns ListDomains --Region {{env.VOLCENGINE_REGION}}
```

### Your First Command

```bash
# List all domains
ve dns ListDomains --Region {{env.VOLCENGINE_REGION}}

# Add an A record to a domain
ve dns AddRecord \
  --Region {{env.VOLCENGINE_REGION}} \
  --DomainName "{{user.domain}}" \
  --RR "www" \
  --Type "A" \
  --Value "192.168.1.1"
```

## Capabilities at a Glance

| Operation | Description | Complexity | Risk Level |
|-----------|-------------|------------|------------|
| ListDomains | List all domains | Low | None |
| DescribeDomain | View domain details | Low | None |
| CreateDomain | Add a new domain | Low | Low |
| DeleteDomain | Remove a domain | Low | **High** — irreversible |
| ModifyDomain | Change domain configuration | Low | Low |
| AddRecord | Add a DNS record | Low | Low |
| UpdateRecord | Modify an existing record | Low | Medium |
| DeleteRecord | Remove a DNS record | Low | Medium |
| ListRecords | List all records for a domain | Low | None |
| DescribeDomainStatistics | View domain resolution stats | Low | None |
| DescribeDNSResolution | Get resolution details | Low | None |
| BatchImportRecords | Import records in bulk | Medium | Medium |
| BatchExportRecords | Export all records | Low | None |

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

## Error Taxonomy (≥ 10 Codes)

| Error Code | Meaning | Resolution |
|------------|---------|-----------|
| `InvalidDomainName` | Domain name format invalid | 0 retries; HALT — provide FQDN format |
| `DomainAlreadyExists` | Domain already in account | 0 retries; HALT — check ListDomains |
| `DomainNotFound` | Domain does not exist | 0 retries; HALT — verify domain name |
| `InvalidRecordType` | Record type not supported | 0 retries; HALT — use valid type |
| `InvalidRecordValue` | Record value format invalid | 0 retries; HALT — fix per type |
| `DuplicateRecord` | Duplicate DNS record | 0 retries; HALT — use UpdateRecord instead |
| `RecordNotFound` | Record does not exist | 0 retries; HALT — verify RecordId |
| `RecordLimitExceeded` | Record count limit reached | 0 retries; HALT — delete unused records |
| `QuotaExceeded` | Domain quota exceeded | 0 retries; HALT — request quota increase |
| `InsufficientBalance` | Account balance low | 0 retries; HALT — recharge account |
| `Unauthorized` | IAM permission denied | 0 retries; HALT — check IAM policies |
| `RecordTypeMismatch` | Cannot change record type | 0 retries; HALT — delete and re-add |
| `DependencyViolation` | Domain has dependencies | 0 retries; HALT — remove dependencies |
| `Throttling` | Rate limit exceeded | 3 retries/1s/2s/4s; Back off and retry |
| `InternalError` | Server-side error | 3 retries/2s/4s/8s; Retry then HALT |

## Safety Gate: DeleteDomain (Irreversible)

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| User confirmation | Ask user explicitly | Written "yes" | Do NOT proceed |
| Records warning | List all records | User acknowledges | Do NOT proceed |
| Domain in use | Check if DNS is referenced | Not in active use | Warn about service impact |

**Confirmation prompt template:**
```
/!\\ WARNING: You are about to DELETE domain '{{user.domain}}' ({{user.domain_id}}).

This will permanently remove:
- The domain itself
- ALL DNS records ({{record_count}} total)
- DNS resolution for this domain will stop

Type "DELETE" to confirm, or anything else to cancel.
```

## Prerequisites

### 1. Install `ve` CLI (Primary Execution Path)

```bash
# Download from GitHub releases
# macOS (ARM64)
curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/v1.0.42/ve-darwin-arm64 -o /usr/local/bin/ve
chmod +x /usr/local/bin/ve

# Verify
ve version
```

### 2. Configure Credentials

```bash
export VOLCENGINE_ACCESS_KEY="{{env.VOLCENGINE_ACCESS_KEY}}"
export VOLCENGINE_SECRET_KEY="{{env.VOLCENGINE_SECRET_KEY}}"
export VOLCENGINE_REGION="{{env.VOLCENGINE_REGION}}"
```

### 3. Verify Configuration

```bash
# List domains to verify credentials and CLI
ve dns ListDomains --Region {{env.VOLCENGINE_REGION}}
```

## Records by Type — Quick Reference

| Type | Purpose | Example Value | TTL (typical) | Notes |
|------|---------|---------------|---------------|-------|
| **A** | Map hostname to IPv4 | `192.168.1.1` | 600 | Most common |
| **AAAA** | Map hostname to IPv6 | `2001:db8::1` | 600 | IPv6 resolution |
| **CNAME** | Alias to another domain | `example.com` | 600 | Can't be at zone apex |
| **MX** | Mail exchange server | `mail.example.com` | 600 | Requires Priority (lower = higher priority) |
| **TXT** | Text metadata | `v=spf1 include:_spf...` | 600 | SPF, DKIM, domain verification |
| **NS** | Name server delegation | `ns1.example.com` | 86400 | Subdomain delegation |
| **SRV** | Service location | `10 60 5060 sip.example.com` | 600 | Priority Weight Port Target |
| **CAA** | Certificate authority auth | `0 issue "letsencrypt.org"` | 600 | CA authorization |

## Reference Directory

- [Core Concepts](references/core-concepts.md) — DNS architecture and terminology
- [API & SDK Usage](references/api-sdk-usage.md) — OpenAPI operations and Go SDK
- [CLI Usage](references/cli-usage.md) — `ve dns` commands and coverage
- [Troubleshooting Guide](references/troubleshooting.md) — Common issues and diagnostics
- [Monitoring & Alerts](references/monitoring.md) — DNS metrics and alerting
- [Integration](references/integration.md) — Environment setup and SDK workflow
- [Knowledge Base](references/knowledge-base.md) — DNS fault patterns and diagnostics
- [GCL Rubric](references/rubric.md)
- [GCL Prompt Templates](references/prompt-templates.md)

## Operational Best Practices

- **TTL strategy:** Use low TTL (60-300s) before planned changes, restore to normal (600-3600s) after
- **Record hygiene:** Regularly audit and remove unused records; document record purpose in comments
- **MX records:** Always configure at least 2 MX servers with different priorities for redundancy
- **SPF/DKIM/DMARC:** Use TXT records for email authentication to improve deliverability
- **CNAME restrictions:** CNAME records cannot coexist with other record types at the same name; use ALIAS/ANAME if needed
- **Security:** Enable DNSSEC if available; use CAA records to restrict certificate issuance
- **Monitoring:** Set up DNS resolution success rate alerts; monitor query volume anomalies
- **Backup:** Export DNS records before making bulk changes; maintain off-platform backups
- **Delegation:** Use NS records for subdomain delegation; verify delegation with `dig`
- **Load distribution:** Use weighted SRV records or multiple A records with DNS-based load balancing

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-31 | Initial release with DNS domain, record, and resolution management |
| 1.1.0 | 2026-06-04 | GCL rollout: added ## Quality Gate (GCL), references/rubric.md, references/prompt-templates.md |

## Quality Gate (GCL)

> Optional tier. max_iter=5.

| Tier | Operations | Safety |
|---|---|---|
| **Destructive** | DeleteDNSRecord, DeleteDNSDomain | 1.0 |
| **State-changing** | CreateDNSRecord, ModifyDNSRecord, ModifyDNSDomain | 1.0 |
| **Mutating** | CreateDNSDomain | ≥0.5 |
| **Read-only** | DescribeDNSDomains, DescribeDNSRecords, DescribeDomainStatistics | ≥0 |

Safety: DeleteDomain ALL records lost. DeleteRecord resolution breaks. VOLCENGINE_SECRET_KEY never.

### Cross-skill: Billing→ve-billing-ops

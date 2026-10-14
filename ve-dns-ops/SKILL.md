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

### Next Steps
- [Core Concepts](references/core-concepts.md) — Understand DNS architecture
- [Common Operations](#execution-flows) — Create, manage, and manage domain resolution
- [Troubleshooting](references/troubleshooting.md) — Fix common issues

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

# See Execution Flows in references/execution-flows.md

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
curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/{{env.ve_version}}/ve-darwin-arm64 -o /usr/local/bin/ve
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

## Error Taxonomy

| Error Code | Meaning | Resolution |
|------------|---------|-----------|
| `InvalidDomainName` | Domain name format invalid | 0 retries; **HALT** — check domain format |
| `InvalidRecordId` | Record ID does not exist | 0 retries; **HALT** — verify record ID |
| `InvalidRecordSet` | Record set configuration invalid | 0 retries; **RETRY** — fix record parameters |
| `RecordConflict` | Record conflicts with existing entry | 0 retries; **RETRY** — use different record name |
| `QuotaExceeded.Domain` | Domain quota reached | 0 retries; **HALT** — request domain quota increase |
| `QuotaExceeded.Record` | Record quota per domain reached | 0 retries; **HALT** — delete unused records |
| `DomainNotRegistered` | Domain not registered | 0 retries; **HALT** — register domain first |
| `DomainLocked` | Domain is locked / transfer prohibited | 0 retries; **HALT** — unlock domain first |
| `InvalidLineType` | DNS line type unsupported | 0 retries; **HALT** — check supported line types |
| `InvalidTtl` | TTL value out of allowed range | 0 retries; **RETRY** — set TTL within 60-86400s |
| `Throttling` | Rate limit exceeded | 3 retries/exponential; **RETRY** — Back off and retry |
| `InternalError` | Server-side error | 3 retries/2s/4s/8s; **RETRY** — Retry, escalate with RequestId |

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

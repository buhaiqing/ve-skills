---
name: ve-cdn-ops
description: >-
  Use when the user needs to deploy, configure, troubleshoot, or manage Volcengine
  (火山引擎) CDN (内容分发网络) — domain management, origin configuration, caching rules,
  HTTPS/TLS certificates, access control (IP whitelist/blacklist, referer, UA filter),
  content optimization (compression, image optimization), cache purging, and monitoring.
  User mentions CDN, 内容分发网络, 加速域名, cache purge, 刷新预热, or describes
  content delivery scenarios (e.g., speeding up website access, configuring origin servers,
  setting cache rules, enabling HTTPS, purging cache) even without naming the product directly.
  Not for ECS, TOS, or other compute/storage products that have their own ops skills.
license: MIT
compatibility: >-
  Official Volcengine CLI (`ve`, Go binary, no runtime), Go 1.14+ runtime
  (for JIT SDK fallback), valid API credentials, network access to Volcengine
  endpoints (open.volcengineapi.com).
metadata:
  author: volcengine
  version: "1.1.1"
  last_updated: "2026-07-11"
  runtime: Harness AI Agent, Claude Code, Cursor, or compatible Agent runtimes
  go_version_minimum: "1.14"
  go_version_jit: "1.21+"
  api_profile: "CDN API 2021-03-01 (https://www.volcengine.com/docs/6454/147453)"
  cli_applicability: dual-path
  cli_support_evidence: >-
    Confirmed via `ve cdn --help` — CDN is supported by the ve CLI.
    See: https://github.com/volcengine/volcengine-cli
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

> This skill follows the [Agent Skill OpenSpec](https://agentskills.io/specification).

# Volcengine CDN Operations Skill

## Overview

CDN (内容分发网络 / Content Delivery Network) on Volcengine (火山引擎) provides global content acceleration with edge caching, origin pull optimization, HTTPS support, and access control. This skill is an **operational runbook** for agents: explicit scope, credential rules, pre-flight checks, **dual-path execution** (official **SDK/API** and **`ve` CLI**), response validation, and failure recovery.

### CLI applicability

- **`cli_applicability: dual-path`:** Official `ve` CLI supports CDN. You **MUST** document **both** the SDK step **and** the `ve` CLI step for every operation.

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use conditions with precise triggers and delegation rules |
| 2 | **Structured I/O** | `{{env.VOLCENGINE_*}}` (env), `{{user.*}}` (interactive), `{{output.*}}` (API response) |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | Error taxonomy with ≥ 10 CDN-specific codes; HALT vs retry per type |
| 5 | **Absolute Single Responsibility** | One product (CDN), one primary resource (Domain); cross-product delegation documented |
| 6 | **FinOps Integration** | Bandwidth analysis, cache hit ratio optimization, cost reports |
| 7 | **AIOps Integration** | Knowledge base with fault patterns, cross-skill diagnosis, alarm storm handling |

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- User mentions "Volcengine CDN", "火山引擎 CDN", "内容分发网络", "加速域名", or "CDN"
- Task involves **domain management**: AddCdnDomain, DeleteCdnDomain, ListCdnDomains, DescribeCdnDomainDetail
- Task involves **origin configuration**: origin server setup, origin pull policies, origin host header
- Task involves **caching rules**: cache expiration, cache key configuration, custom cache rules
- Task involves **HTTPS/TLS**: certificate upload, HTTPS enablement, TLS version configuration
- Task involves **access control**: IP whitelist/blacklist, referer filtering, User-Agent filtering
- Task involves **content optimization**: Gzip/Brotli compression, image optimization, WebP conversion
- Task involves **cache purging**: URL purge, directory purge, cache prefetch
- Task involves **monitoring**: bandwidth/traffic metrics, cache hit ratio, origin pull statistics
- Task involves **cost optimization**: bandwidth analysis, cache efficiency improvements

### SHOULD NOT Use This Skill When

- Task is about **ECS** (云服务器) → delegate to: `ve-ecs-ops`
- Task is about **TOS** (对象存储) → delegate to: `ve-tos-ops`
- Task is about **VKE** (容器服务) → delegate to: `ve-vke-ops`
- Task is about **DNS** management (outside CDN context) → delegate to DNS ops skill
- Task is purely billing / account management → delegate to billing ops
- Task is IAM / permission model only → delegate to: `ve-iam-ops`

### Delegation Rules

- CDN origin can be TOS bucket → reference `ve-tos-ops` for bucket configuration
- CDN origin can be ECS instance/CLB → reference `ve-ecs-ops` or `ve-clb-ops`
- HTTPS certificates may need upload → document certificate requirements

## Variable Convention (Agent-Readable)

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_REGION}}` | From runtime environment | Use documented default only if skill explicitly allows |
| `{{user.region}}` | User-supplied region | Ask once; reuse; default from env |
| `{{user.domain}}` | CDN domain name | Ask once; format: `www.example.com` or `cdn.example.com` |
| `{{user.origin_domain}}` | Origin server domain/IP | Ask once; can be IP or domain |
| `{{user.origin_type}}` | Origin type: `domain`, `ip`, `tos` | Ask once; default `domain` |
| `{{user.protocol}}` | Protocol: `http`, `https`, `follow` | Ask once; default `http` |
| `{{user.service_type}}` | Service type: `static`, `media`, `download` | Ask once; default `static` |
| `{{output.domain_id}}` | Domain ID from API response | Parse from `$.Result.DomainId` or `$.Result.Domain` |
| `{{output.cname}}` | CNAME record for DNS configuration | Parse from response |
| `{{output.status}}` | Domain status | Parse from `$.Result.Status` |

> **`{{env.*}}` MUST NOT** be collected from the user. **`{{user.*}}`** MUST be collected interactively when missing.

> **Security Warning (Credential Masking — MANDATORY):** **NEVER** log, print, or expose `VOLCENGINE_SECRET_KEY`, `secret_key`, `SecretKey`, or any credential field value in console output, debug messages, error messages, or logs.
>
> **Credential verification MUST check existence only**, never echo the value:
> - Bash: `test -n "$VOLCENGINE_SECRET_KEY"` ✅ | `echo $VOLCENGINE_SECRET_KEY` ❌
> - Go: `if os.Getenv("VOLCENGINE_SECRET_KEY") == ""` ✅ | `fmt.Println(os.Getenv("VOLCENGINE_SECRET_KEY"))` ❌

## API and Response Conventions (Agent-Readable)

- **Volcengine CDN OpenAPI (2021-03-01)** is canonical for path, query, body fields, enums, and response shapes.
- **Endpoint:** `cdn.volcengineapi.com` (default: `open.volcengineapi.com`)
- **Errors:** Responses with `Error` object containing `Code` and `Message` fields.
- **Timestamps:** ISO 8601 format (e.g. `2026-04-28T10:00:00Z`).

### Key Response Fields

| Operation | JSON Path | Type | Description |
|-----------|-----------|------|-------------|
| AddCdnDomain | `$.Result.DomainId` | string | Created domain ID |
| AddCdnDomain | `$.Result.Domain` | string | Domain name |
| AddCdnDomain | `$.Result.Cname` | string | CNAME record for DNS |
| ListCdnDomains | `$.Result.Domains` | array | Domain list |
| ListCdnDomains | `$.Result.Domains[].Domain` | string | Domain name |
| ListCdnDomains | `$.Result.Domains[].DomainId` | string | Domain ID |
| ListCdnDomains | `$.Result.Domains[].Status` | string | `online`, `offline`, `processing`, `error` |
| ListCdnDomains | `$.Result.Domains[].Cname` | string | CNAME record |
| DescribeCdnDomainDetail | `$.Result.DomainConfig` | object | Full domain configuration |
| SubmitRefreshTask | `$.Result.TaskId` | string | Purge task ID |
| DescribeContentQuota | `$.Result.RefreshQuota` | integer | Daily URL purge quota |
| DescribeContentQuota | `$.Result.DirQuota` | integer | Daily directory purge quota |
| DescribeCdnData | `$.Result.Data` | array | Metrics time series data |

### Domain State Transitions

| Operation | Initial State | Target State | Poll Interval | Max Wait |
|-----------|---------------|--------------|---------------|----------|
| AddCdnDomain | — | `online` | 10s | 600s |
| StartCdnDomain | `offline` | `online` | 10s | 300s |
| StopCdnDomain | `online` | `offline` | 10s | 300s |
| DeleteCdnDomain | any | absent | 10s | 300s |

## Quick Start

### What This Skill Does
This skill enables you to deploy, configure, troubleshoot, and manage Volcengine CDN domains, origins, caching, HTTPS, access control, and cache purging using the `ve` CLI (primary) or JIT Go SDK (fallback).

### Prerequisites
- [ ] `ve` CLI installed (or Go runtime for JIT fallback)
- [ ] Credentials configured: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`
- [ ] Region set: `VOLCENGINE_REGION`

### Verify Setup
```bash
ve cdn ListCdnDomains --Region {{env.VOLCENGINE_REGION}} --PageSize 10
```

### Your First Command
```bash
# List all CDN domains
ve cdn ListCdnDomains --Region {{env.VOLCENGINE_REGION}}
```

### Next Steps
- [Core Concepts](references/core-concepts.md) — Understand CDN architecture
- [Common Operations](#execution-flows) — Create, manage, and accelerate content delivery
- [Troubleshooting](references/troubleshooting.md) — Fix common issues

## Capabilities at a Glance

| Operation | Description | Complexity | Risk Level |
|-----------|-------------|------------|------------|
| AddCdnDomain | Add a new CDN domain | Medium | Medium |
| ListCdnDomains | List all CDN domains | Low | None |
| DescribeCdnDomainDetail | Get domain configuration details | Low | None |
| UpdateCdnDomainConfig | Update domain configuration | Medium | Medium |
| DeleteCdnDomain | Delete a CDN domain | Low | **High** — irreversible |
| StartCdnDomain | Enable a disabled domain | Low | Low |
| StopCdnDomain | Disable a domain | Low | Medium |
| UpdateCdnDomainOrigin | Update origin server configuration | Medium | Medium |
| UpdateCdnDomainCacheRule | Update caching rules | Medium | Low |
| UpdateCdnDomainHttps | Configure HTTPS/TLS | Medium | Medium |
| UpdateCdnDomainAccessControl | Configure IP/referer/UA filtering | Medium | Low |
| SubmitRefreshTask | Purge cache by URL | Low | Low |
| SubmitDirRefreshTask | Purge cache by directory | Low | Low |
| SubmitPreloadTask | Prefetch content to edge | Medium | Low |
| DescribeContentQuota | Query purge/prefetch quotas | Low | None |
| DescribeCdnData | Query bandwidth/traffic metrics | Low | None |
| DescribeEdgeData | Query edge node statistics | Low | None |
| DescribeOriginData | Query origin pull statistics | Low | None |
| DescribeCdnRegionData | Query regional distribution | Low | None |
| DescribeContentTasks | List purge/prefetch tasks | Low | None |
| DescribeCdnDomainLog | Download access logs | Low | None |
| UpdateCdnDomainCompression | Enable Gzip/Brotli compression | Low | Low |
| DescribeCdnDomainHitRate | Query cache hit ratio | Low | None |
| AnalyzeBandwidthCost | Generate bandwidth cost report | Low | None |

> FinOps Operations 详情 → [references/advanced/finops.md#operations](references/advanced/finops.md#operations)

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-27 | Initial release with domain lifecycle, origin config, caching, HTTPS, access control, purging, and monitoring |
| 1.1.0 | 2026-06-04 | GCL rollout: added ## Quality Gate (GCL), references/rubric.md, references/prompt-templates.md |
| 1.1.1 | 2026-07-11 | Split FinOps Operations (1 op: AnalyzeBandwidthCost) to `references/advanced/finops.md` (TE-7); SKILL.md retains entry link only |

## Quality Gate (GCL)

> Optional tier. max_iter=5.

| Tier | Operations | Safety |
|---|---|---|
| **Destructive** | DeleteDomain | 1.0 |
| **State-changing** | StartDomain, StopDomain, CreateDomain, ModifyDomainConfig, SubmitRefreshTask | 1.0 |
| **Mutating** | PreLoadCache | ≥0.5 |
| **Read-only** | DescribeDomains, DescribeDomainDetail, DescribeCdnData, DescribeCdnOrigin, DescribeCdnTopData, DescribeContentQuota | ≥0 |

Safety: DeleteDomain config lost. StopDomain traffic stops. RefreshTask purge irreversible. VOLCENGINE_SECRET_KEY never.

### Cross-skill: Billing→ve-billing-ops

## Execution Flows (Agent-Readable)

Every operation: **Pre-flight → Execute (ve CLI primary + JIT Go SDK fallback) → Validate → Recover**.

### Operation: ListCdnDomains — Query CDN Domain List

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | `test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY"` | Both set | HALT; configure credentials |
| Region | Verify `{{user.region}}` or `{{env.VOLCENGINE_REGION}}` is set | Valid region | Suggest: `ve cdn DescribeCdnRegion` |
| CLI | `ve version` | Exit code 0 | Install ve CLI |

#### Execution — CLI (`ve`)

```bash
# List all domains (JSON output by default)
ve cdn ListCdnDomains --Region "{{user.region}}"

# List with pagination
ve cdn ListCdnDomains --Region "{{user.region}}" --PageNum 1 --PageSize 50

# Filter by domain name
ve cdn ListCdnDomains --Region "{{user.region}}" --Domain "{{user.domain}}"

# Filter by status
ve cdn ListCdnDomains --Region "{{user.region}}" --Status "online"
```

#### Execution — JIT Go SDK (Fallback)

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/cdn"
)

func main() {
    instance := cdn.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "Region":   os.Getenv("VOLCENGINE_REGION"),
        "PageNum":  1,
        "PageSize": 50,
    }

    resp, err := instance.Client.Request("ListCdnDomains", nil, params)
    if err != nil {
        log.Fatalf("Failed to list domains: %v", err)
    }
    fmt.Println(string(resp))
}
```

#### Validation

1. Check `$.Result.Total` for total domain count
2. Parse `$.Result.Domains[]` for domain details
3. Report domain count, names, statuses, and CNAMEs

#### Failure Recovery

| Error Pattern | Max Retries | Agent Action |
|--------------|-------------|--------------|
| `InvalidRegion.NotFound` | 0 | List valid regions via `DescribeCdnRegion`; HALT |
| `Unauthorized` | 0 | HALT; check IAM permissions |
| `InternalError` | 3 | Retry with exponential backoff; HALT after 3 |
| Throttling / 429 | 3 | Back off (2s, 4s, 8s); retry |

---

### Operation: AddCdnDomain — Add CDN Domain

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | Env vars set | Non-empty | HALT |
| Domain format | Validate domain name format | Valid FQDN | HALT; fix domain format |
| Origin valid | Verify origin domain/IP is reachable | Origin responds | HALT; check origin |
| Domain not duplicate | `ListCdnDomains` filter | Domain not exists | HALT; domain already added |
| Service type | Determine if static/media/download | Valid type | Default: `static` |

#### Execution — CLI (`ve`)

```bash
# Add a domain with minimal parameters
ve cdn AddCdnDomain \
  --Region "{{user.region}}" \
  --Domain "{{user.domain}}" \
  --OriginDomain "{{user.origin_domain}}" \
  --OriginType "{{user.origin_type}}" \
  --Protocol "{{user.protocol}}" \
  --ServiceType "{{user.service_type}}"
```

**Common parameters:**
| Parameter | Description | Example |
|-----------|-------------|---------|
| `Domain` | CDN domain name | `cdn.example.com` |
| `OriginDomain` | Origin server address | `origin.example.com` or `192.168.1.1` |
| `OriginType` | Origin type | `domain`, `ip`, `tos` |
| `Protocol` | Origin protocol | `http`, `https`, `follow` |
| `ServiceType` | Service type | `static`, `media`, `download` |
| `OriginHost` | Host header for origin pull | `origin.example.com` |
| `OriginPort` | Origin port | `80`, `443`, `8080` |

#### Execution — JIT Go SDK (Fallback)

```go
params := map[string]interface{}{
    "Region":       os.Getenv("VOLCENGINE_REGION"),
    "Domain":       "cdn.example.com",
    "OriginDomain": "origin.example.com",
    "OriginType":   "domain",
    "Protocol":     "http",
    "ServiceType":  "static",
}

resp, err := instance.Client.Request("AddCdnDomain", nil, params)
```

#### Post-execution Validation

1. Parse `$.Result.DomainId` → `{{output.domain_id}}`
2. Parse `$.Result.Cname` → `{{output.cname}}`
3. Poll status until `online`:

```bash
for i in $(seq 1 60); do
  STATUS=$(ve cdn ListCdnDomains --Region "{{user.region}}" --Domain "{{user.domain}}" | jq -r '.Result.Domains[0].Status')
  [ "$STATUS" = "online" ] && break
  echo "Current: $STATUS (poll $i/60)"
  sleep 10
done
[ "$STATUS" != "online" ] && echo "[ERROR] Domain failed to become online (current: $STATUS)" && exit 1
```

4. On success, report domain ID, CNAME, and DNS configuration instructions

#### Failure Recovery

| Error Pattern | Max Retries | Agent Action | UX Feedback |
|--------------|-------------|--------------|-------------|
| `DomainAlreadyExists` | 0 | HALT; domain already in CDN | `[ERROR] DomainAlreadyExists: Domain is already registered in CDN. How to fix: Use the existing domain or choose a different domain name.` |
| `InvalidDomain.Format` | 0 | HALT; invalid domain format | `[ERROR] InvalidDomain.Format: Domain name format is invalid. How to fix: Use a valid FQDN (e.g., cdn.example.com).` |
| `InvalidOrigin.NotReachable` | 0 | HALT; origin not responding | `[ERROR] InvalidOrigin: Origin server is not reachable. How to fix: Verify origin domain/IP and ensure it's accessible.` |
| `InvalidOriginType` | 0 | HALT; invalid origin type | `[ERROR] InvalidOriginType: Origin type must be domain, ip, or tos.` |
| `QuotaExceeded.Domain` | 0 | HALT | `[ERROR] QuotaExceeded: Domain quota reached. How to fix: Contact Volcengine support to increase quota.` |
| `Unauthorized` | 0 | HALT; check IAM | `[ERROR] Unauthorized: Insufficient permissions. How to fix: Ensure CDNFullAccess policy is attached.` |
| `InternalError` | 3 | Retry with backoff | `[ERROR] InternalError: Server-side error. Will retry automatically.` |
| `Throttling` | 3 | Exponential backoff | `⚠️ Rate limit reached. Retrying...` |

---

### Operation: DeleteCdnDomain — Delete CDN Domain

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: irreversible delete of domain `{{user.domain}}`
- **MUST NOT** proceed without clear user assent
- **MUST** verify domain is in `offline` status (stop domain first if online)
- **MUST** warn about DNS impact — CNAME will stop working immediately

```bash
# Check current status
ve cdn ListCdnDomains --Region "{{user.region}}" --Domain "{{user.domain}}" | jq '.Result.Domains[0].Status'

# If online, stop first
ve cdn StopCdnDomain --Region "{{user.region}}" --Domain "{{user.domain}}"
```

#### Execution

```bash
# Delete the domain
ve cdn DeleteCdnDomain --Region "{{user.region}}" --Domain "{{user.domain}}"
```

#### Validation

Poll `ListCdnDomains` until domain not found or status is absent (max 300s).

---

### Operation: StartCdnDomain — Enable Domain

#### Execution

```bash
ve cdn StartCdnDomain --Region "{{user.region}}" --Domain "{{user.domain}}"
```

#### Validation

Poll until `Status` = `online` (max 300s, interval 10s).

---

### Operation: StopCdnDomain — Disable Domain

#### Execution

```bash
ve cdn StopCdnDomain --Region "{{user.region}}" --Domain "{{user.domain}}"
```

#### Validation

Poll until `Status` = `offline` (max 300s, interval 10s).

---

### Operation: UpdateCdnDomainOrigin — Update Origin Configuration

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Domain exists | `ListCdnDomains` with filter | Domain found | HALT |
| Origin reachable | Verify new origin responds | Origin accessible | HALT |

#### Execution

```bash
ve cdn UpdateCdnDomainOrigin \
  --Region "{{user.region}}" \
  --Domain "{{user.domain}}" \
  --OriginDomain "{{user.origin_domain}}" \
  --OriginType "{{user.origin_type}}" \
  --Protocol "{{user.protocol}}" \
  --OriginHost "{{user.origin_host}}"
```

#### Validation

1. Query domain detail to verify origin config updated
2. Test content delivery from edge node

---

### Operation: UpdateCdnDomainCacheRule — Configure Caching Rules

#### Execution

```bash
# Update cache rules via JSON configuration
ve cdn UpdateCdnDomainConfig \
  --Region "{{user.region}}" \
  --Domain "{{user.domain}}" \
  --CacheRules '[{"RuleType":"path","RuleValue":"/*.jpg","TTL":86400,"Priority":1}]'
```

**Cache rule types:**

| RuleType | Description | Example |
|----------|-------------|---------|
| `path` | Match by file path | `/*.jpg`, `/static/*` |
| `filetype` | Match by file extension | `jpg,png,css,js` |
| `directory` | Match by directory | `/images/`, `/assets/` |
| `fullpath` | Exact path match | `/api/v1/data.json` |
| `home` | Homepage | `/` |

**TTL values:**
- `0` = No cache (always fetch from origin)
- `1-31536000` = Cache time in seconds (max 1 year)
- Common: `3600` (1 hour), `86400` (1 day), `604800` (7 days)

---

### Operation: UpdateCdnDomainHttps — Configure HTTPS

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Certificate valid | Verify certificate and key | Valid PEM format | HALT |
| Certificate matches domain | Check CN/SAN | Domain in cert | HALT |
| Private key matches cert | Verify key pair | Match confirmed | HALT |

#### Execution

```bash
# Enable HTTPS with certificate
ve cdn UpdateCdnDomainHttps \
  --Region "{{user.region}}" \
  --Domain "{{user.domain}}" \
  --HttpsSwitch "on" \
  --Certificate "$(cat cert.pem)" \
  --PrivateKey "$(cat key.pem)" \
  --TlsVersion "tlsv1.2+"
```

**TLS version options:**
- `tlsv1.0+` = TLS 1.0 and above (not recommended)
- `tlsv1.1+` = TLS 1.1 and above
- `tlsv1.2+` = TLS 1.2 and above (recommended)
- `tlsv1.3` = TLS 1.3 only

#### Validation

1. Query domain detail to verify HTTPS config
2. Test HTTPS access: `curl -I https://{{user.domain}}`

---

### Operation: UpdateCdnDomainAccessControl — Configure Access Control

#### Execution — IP Whitelist/Blacklist

```bash
# Configure IP whitelist (only allow specified IPs)
ve cdn UpdateCdnDomainAccessControl \
  --Region "{{user.region}}" \
  --Domain "{{user.domain}}" \
  --IpFilter "{{user.ip_list}}" \
  --IpFilterType "whitelist"

# Configure IP blacklist (block specified IPs)
ve cdn UpdateCdnDomainAccessControl \
  --Region "{{user.region}}" \
  --Domain "{{user.domain}}" \
  --IpFilter "{{user.ip_list}}" \
  --IpFilterType "blacklist"
```

#### Execution — Referer Filtering

```bash
# Referer whitelist (only allow requests from specified referers)
ve cdn UpdateCdnDomainAccessControl \
  --Region "{{user.region}}" \
  --Domain "{{user.domain}}" \
  --RefererFilter "{{user.referer_list}}" \
  --RefererFilterType "whitelist" \
  --RefererAllowEmpty "true"

# Referer blacklist (block requests from specified referers)
ve cdn UpdateCdnDomainAccessControl \
  --Region "{{user.region}}" \
  --Domain "{{user.domain}}" \
  --RefererFilter "{{user.referer_list}}" \
  --RefererFilterType "blacklist"
```

#### Execution — User-Agent Filtering

```bash
# Block specific User-Agents
ve cdn UpdateCdnDomainAccessControl \
  --Region "{{user.region}}" \
  --Domain "{{user.domain}}" \
  --UaFilter "{{user.ua_list}}" \
  --UaFilterType "blacklist"
```

---

### Operation: SubmitRefreshTask — URL Cache Purge

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Domain exists | `ListCdnDomains` | Domain found | HALT |
| Quota available | `DescribeContentQuota` | Quota > 0 | HALT; quota exhausted |

#### Execution

```bash
# Purge single URL
ve cdn SubmitRefreshTask \
  --Region "{{user.region}}" \
  --Type "url" \
  --Urls '["https://{{user.domain}}/path/to/file.jpg"]'

# Purge multiple URLs
ve cdn SubmitRefreshTask \
  --Region "{{user.region}}" \
  --Type "url" \
  --Urls '["https://{{user.domain}}/file1.jpg","https://{{user.domain}}/file2.css"]'

# Purge with URL encoding for special characters
ve cdn SubmitRefreshTask \
  --Region "{{user.region}}" \
  --Type "url" \
  --Urls '["https://{{user.domain}}/path%20with%20spaces.jpg"]'
```

#### Validation

1. Capture `$.Result.TaskId` from response
2. Query task status:

```bash
ve cdn DescribeContentTasks --Region "{{user.region}}" --TaskId "{{output.task_id}}"
```

3. Check `$.Result.Tasks[].Status` → `processing`, `done`, `failed`
4. On completion, verify cache is purged by fetching URL again

#### Failure Recovery

| Error Pattern | Max Retries | Agent Action |
|--------------|-------------|--------------|
| `QuotaExceeded.Refresh` | 0 | HALT; daily purge quota exhausted |
| `InvalidUrl.Format` | 0 | HALT; URL format invalid |
| `DomainNotFound` | 0 | HALT; domain doesn't exist |
| `TaskSubmitFailed` | 2 | Retry; HALT after 2 |

---

### Operation: SubmitDirRefreshTask — Directory Cache Purge

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Quota available | `DescribeContentQuota` | DirQuota > 0 | HALT; quota exhausted |

#### Execution

```bash
# Purge entire directory
ve cdn SubmitRefreshTask \
  --Region "{{user.region}}" \
  --Type "dir" \
  --Dirs '["https://{{user.domain}}/static/"]'

# Purge multiple directories
ve cdn SubmitRefreshTask \
  --Region "{{user.region}}" \
  --Type "dir" \
  --Dirs '["https://{{user.domain}}/images/","https://{{user.domain}}/css/"]'
```

> **Note:** Directory purge removes all cached content under the specified path.

---

### Operation: SubmitPreloadTask — Cache Prefetch

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| URLs valid | Verify URL format | Valid URLs | HALT |

#### Execution

```bash
# Prefetch URLs to edge nodes
ve cdn SubmitPreloadTask \
  --Region "{{user.region}}" \
  --Urls '["https://{{user.domain}}/important/file.jpg"]'

# Prefetch multiple URLs
ve cdn SubmitPreloadTask \
  --Region "{{user.region}}" \
  --Urls '["https://{{user.domain}}/file1.jpg","https://{{user.domain}}/file2.css"]'
```

#### Validation

Poll `DescribeContentTasks` until status = `done`.

---

### Operation: DescribeContentQuota — Query Purge/Prefetch Quotas

#### Execution

```bash
ve cdn DescribeContentQuota --Region "{{user.region}}"
```

#### Response Fields

| Field | Description |
|-------|-------------|
| `RefreshQuota` | Daily URL purge quota remaining |
| `RefreshTotal` | Total daily URL purge quota |
| `DirQuota` | Daily directory purge quota remaining |
| `DirTotal` | Total daily directory purge quota |
| `PreloadQuota` | Daily prefetch quota remaining |
| `PreloadTotal` | Total daily prefetch quota |

---

### Operation: DescribeCdnData — Query Bandwidth/Traffic Metrics

#### Execution

```bash
# Query bandwidth data
ve cdn DescribeCdnData \
  --Region "{{user.region}}" \
  --StartTime "{{user.start_time}}" \
  --EndTime "{{user.end_time}}" \
  --Metric "bandwidth" \
  --Domain "{{user.domain}}"

# Query traffic data
ve cdn DescribeCdnData \
  --Region "{{user.region}}" \
  --StartTime "{{user.start_time}}" \
  --EndTime "{{user.end_time}}" \
  --Metric "traffic" \
  --Domain "{{user.domain}}"

# Query request count
ve cdn DescribeCdnData \
  --Region "{{user.region}}" \
  --StartTime "{{user.start_time}}" \
  --EndTime "{{user.end_time}}" \
  --Metric "requests" \
  --Domain "{{user.domain}}"
```

**Time format:** `YYYY-MM-DDTHH:MM:SSZ` (UTC)

**Metric types:**
- `bandwidth` = Bandwidth in bps
- `traffic` = Traffic in bytes
- `requests` = Request count
- `qps` = Queries per second

---

### Operation: DescribeOriginData — Query Origin Pull Statistics

#### Execution

```bash
# Query origin bandwidth
ve cdn DescribeOriginData \
  --Region "{{user.region}}" \
  --StartTime "{{user.start_time}}" \
  --EndTime "{{user.end_time}}" \
  --Metric "bandwidth" \
  --Domain "{{user.domain}}"

# Query origin traffic
ve cdn DescribeOriginData \
  --Region "{{user.region}}" \
  --StartTime "{{user.start_time}}" \
  --EndTime "{{user.end_time}}" \
  --Metric "traffic" \
  --Domain "{{user.domain}}"
```

---

### Operation: DescribeCdnDomainHitRate — Query Cache Hit Ratio

#### Execution

```bash
ve cdn DescribeCdnDomainHitRate \
  --Region "{{user.region}}" \
  --StartTime "{{user.start_time}}" \
  --EndTime "{{user.end_time}}" \
  --Domain "{{user.domain}}"
```

#### Response

| Field | Description |
|-------|-------------|
| `HitRate` | Cache hit ratio (0.0-1.0, multiply by 100 for percentage) |
| `HitTraffic` | Traffic served from cache |
| `MissTraffic` | Traffic from origin pull |

**Hit rate interpretation:**
- `> 0.90` = Excellent (90%+ cache hit)
- `0.70-0.90` = Good
- `0.50-0.70` = Fair
- `< 0.50` = Poor (consider optimizing cache rules)

---

### Operation: UpdateCdnDomainCompression — Enable Compression

#### Execution

```bash
# Enable Gzip compression
ve cdn UpdateCdnDomainCompression \
  --Region "{{user.region}}" \
  --Domain "{{user.domain}}" \
  --GzipSwitch "on" \
  --GzipTypes "text/html,text/css,application/javascript"

# Enable Brotli compression (if supported)
ve cdn UpdateCdnDomainCompression \
  --Region "{{user.region}}" \
  --Domain "{{user.domain}}" \
  --BrotliSwitch "on"
```

-## FinOps Operations (Advanced)

详细 FinOps Operation 步骤（AnalyzeBandwidthCost）见 [`references/advanced/finops.md`](references/advanced/finops.md)。

> SKILL.md 中仅保留 FinOps 入口；具体执行步骤在 `references/advanced/finops.md`（TE-7 专业内容分层）。

## Error Taxonomy

| `code` | 含义 | 分辨率 |
|--------|------|--------|
| `InvalidDomainName` | 加速域名格式无效或不符合规范 | 0 retries; **HALT** — 使用有效 FQDN (如 cdn.example.com) |
| `DomainNotInStatus` | 域名当前状态不允许执行该操作 | 0 retries; **HALT** — 等待域名状态变为 online/offline 后重试 |
| `QuotaExceeded.Domain` | 加速域名数量超出配额限制 | 0 retries; **HALT** — 删除未使用域名或联系技术支持提额 |
| `InvalidOriginConfig` | 源站配置无效 (地址/类型/端口) | 0 retries; **HALT** — 检查源站地址、类型和端口是否正确 |
| `InvalidCacheRule` | 缓存规则配置无效 | 0 retries; **HALT** — 验证规则类型 (path/filetype/directory) 和 TTL 值 |
| `InvalidHttpsConfig` | HTTPS 配置无效 (证书/密钥/TLS) | 0 retries; **HALT** — 验证证书与域名匹配、密钥格式正确 |
| `InvalidCertId` | 证书 ID 不存在或已过期 | 0 retries; **HALT** — 检查证书是否已上传且未过期 |
| `DomainLocked` | 域名被锁定，无法执行操作 | 0 retries; **HALT** — 联系技术支持解锁域名 |
| `QuotaExceeded.Refresh` | 每日刷新/预热配额不足 | 0 retries; **HALT** — 等待配额重置或联系提额 |
| `Throttling` | 请求频率过高触发限流 | 3 retries/exponential/2s/4s/8s; **RETRY** — 背退等待后重试 |
| `InternalError` | 服务端内部错误 | 3 retries/exponential/2s/4s/8s; **RETRY** — 超过重试次数后 HALT 并记录 RequestId |
| `Unauthorized` | 鉴权失败，权限不足 | 0 retries; **HALT** — 检查 IAM 策略是否包含 CDNFullAccess |

## Reference Directory

- [Core Concepts](references/core-concepts.md)
- [API & SDK Usage](references/api-sdk-usage.md)
- [CLI Usage](references/cli-usage.md)
- [Troubleshooting Guide](references/troubleshooting.md)
- [Monitoring](references/monitoring.md)
- [Integration](references/integration.md)
- [User Experience Specification](../ve-skill-generator/references/user-experience-spec.md)
- [Execution Environment Setup](../ve-skill-generator/references/execution-environment.md)
- [CLI Behavioral Reference](../ve-skill-generator/references/cli-behavior.md)
- [GCL Rubric](references/rubric.md)
- [GCL Prompt Templates](references/prompt-templates.md)

## Operational Best Practices

- **Origin health:** Monitor origin server health; CDN cannot serve if origin is down
- **Cache TTL:** Set appropriate TTL based on content update frequency
- **HTTPS:** Always enable HTTPS for production domains
- **Compression:** Enable Gzip/Brotli for text-based content (HTML, CSS, JS, JSON)
- **Purge strategy:** Use URL purge for single files; directory purge for bulk updates
- **Prefetch:** Preload critical content before peak traffic periods
- **Access control:** Use IP whitelist for admin panels; referer filter for hotlink protection
- **Monitoring:** Track cache hit ratio; aim for > 80%
- **Cost optimization:** Optimize cache rules to reduce origin pull bandwidth

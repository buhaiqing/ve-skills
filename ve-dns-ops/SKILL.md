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
  version: "1.0.0"
  last_updated: "2026-05-31"
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

# Volcengine DNS Operations Skill

## Overview

Volcengine DNS (DNS解析) provides domain name resolution services. This skill is an **operational runbook** for agents: domain management, record sets, resolution policies, security, and monitoring. **Do not use the web console as the primary agent execution path.**

### CLI applicability

- **`cli_applicability: dual-path`:** Official `ve` CLI supports DNS. Document **both** SDK and CLI steps.

## Five Core Standards

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | DNS triggers; CDN/CLB delegated |
| 2 | **Structured I/O** | `{{env.*}}`, `{{user.*}}`, `{{output.*}}` |
| 3 | **Explicit Actionable Steps** | Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | DNS error taxonomy; HALT vs retry |
| 5 | **Absolute Single Responsibility** | DNS only; CDN/CLB/ECS delegated |

## Quick Start

```bash
ve dns ListDomains --Region {{env.VOLCENGINE_REGION}}
```

## Reference Directory

- [Core Concepts](references/core-concepts.md)
- [API & SDK Usage](references/api-sdk-usage.md)
- [CLI Usage](references/cli-usage.md)
- [Troubleshooting](references/troubleshooting.md)
- [Monitoring](references/monitoring.md)
- [Integration](references/integration.md)
- [Knowledge Base](references/knowledge-base.md)

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-31 | Initial release with DNS domain, record, and resolution management |

---
name: ve-ark-ops
description: >-
  Use when the user needs to deploy, configure, troubleshoot, or manage Volcengine
  (火山引擎) Ark (方舟大模型平台) — model inference endpoints, model training,
  dataset management, evaluation, and model marketplace. User mentions Ark, 方舟,
  大模型, LLM, model inference, model training, fine-tuning, endpoint, or describes
  AI/LLM deployment scenarios even without naming the product directly.
  Not for ECS, VKE, or traditional compute services.
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
  api_profile: "Ark OpenAPI — https://www.volcengine.com/docs/82379"
  cli_applicability: dual-path
  cli_support_evidence: >-
    Confirmed via `ve ark --help` — Ark is supported by the ve CLI.
    Service ID: ark.
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

# Volcengine Ark (方舟大模型平台) Operations Skill

## Overview

Volcengine Ark (火山引擎方舟大模型平台) provides LLM model inference, fine-tuning, dataset management, and model marketplace services. This skill is an **operational runbook** for agents: endpoint management, model deployment, training jobs, dataset operations, and monitoring. **Do not use the web console as the primary agent execution path.**

### CLI applicability

- **`cli_applicability: dual-path`:** Official `ve` CLI supports Ark. Document **both** SDK and CLI steps.

## Five Core Standards

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | Ark triggers; compute/network delegated |
| 2 | **Structured I/O** | `{{env.*}}`, `{{user.*}}`, `{{output.*}}` |
| 3 | **Explicit Actionable Steps** | Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | Ark error taxonomy; HALT vs retry |
| 5 | **Absolute Single Responsibility** | Ark only; ECS/VKE delegated |

## Quick Start

```bash
ve ark ListEndpoints --Region {{env.VOLCENGINE_REGION}}
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
| 1.0.0 | 2026-05-31 | Initial release with Ark endpoint, model, and training management |

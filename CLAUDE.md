# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is a **Skills Farm** — a collection of Agent Skills for Volcengine (火山引擎) cloud product operations. Each skill is an **operational runbook** for AI agents: structured, parseable, executable specifications for cloud resource management.

The repository follows the [Agent Skill OpenSpec](https://agentskills.io/specification).

## Technology Stack

- **Primary execution**: `ve` CLI (Volcengine CLI — static Go binary, no runtime dependencies)
- **Fallback**: JIT Go SDK scripts (`github.com/volcengine/volc-sdk-golang`)
- **Credentials**: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`, `VOLCENGINE_REGION`

## Skill Structure

Each product skill (`ve-[product]-ops`) has this layout:

```
ve-[product]-ops/
├── SKILL.md               # Main runbook — triggers, flows, recovery
├── references/
│   ├── core-concepts.md   # Architecture, limits, regions
│   ├── api-sdk-usage.md   # Operation map, request/response
│   ├── cli-usage.md       # CLI cheat sheet (when dual-path)
│   ├── troubleshooting.md # Error codes, diagnostics
│   ├── monitoring.md      # Metrics, alerts
│   └── integration.md     # Go bootstrap, JIT SDK setup
└── assets/
    └── example-config.yaml
```

## Key Conventions

### Placeholder System
- `{{env.*}}` — Environment variables (NEVER ask user, fail if unset)
- `{{user.*}}` — Interactive inputs (collect when missing)
- `{{output.*}}` — Captured from API/CLI responses

### Credential Security (MANDATORY)
- **NEVER** log, print, or expose `VOLCENGINE_SECRET_KEY` or any credential value
- Mask all secrets: `VOLCENGINE_SECRET_KEY=<masked>`
- Verify existence only: `test -n "$VOLCENGINE_SECRET_KEY"`

### CLI Behavior
- Command prefix: `ve` (changed from `volcengine-cli` since v1.0.20)
- Output: **JSON by default**
- Help: `ve <service> --help` or `ve <service> <action> --help`
- Env vars: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`, `VOLCENGINE_REGION`

### Five Core Standards
1. **Clear Boundaries**: SHOULD/SHOULD NOT use conditions with delegation targets
2. **Structured I/O**: Typed placeholders, JSON paths from OpenAPI
3. **Explicit Actionable Steps**: Pre-flight → Execute → Validate → Recover
4. **Complete Failure Strategies**: Error taxonomy ≥ 10 codes with recovery actions
5. **Single Responsibility**: One product, one primary resource model

## Generating New Skills

Use `ve-skill-generator` meta-skill to create new product skills. The generator follows an evaluation-driven workflow:

1. Define evaluation targets
2. Analyze OpenAPI sources
3. Scaffold directory layout
4. Populate SKILL.md from template
5. Fill reference files
6. Verify with P0/P1 checklist
7. Final anti-pattern check

See `ve-skill-generator/SKILL.md` for the complete workflow and `ve-skill-generator/references/ve-skill-template.md` for the template.

## Available Skills

| Skill | Product | Description |
|-------|---------|-------------|
| `ve-skill-generator` | Meta-skill | Generates new product skills |
| `ve-ecs-ops` | ECS (云服务器) | Instance lifecycle, disks, snapshots |
| `ve-rds-mysql-ops` | RDS MySQL | Database instances, backups |
| `ve-rds-pg-ops` | RDS PostgreSQL | Database instances |
| `ve-redis-ops` | Redis | Cache instances |
| `ve-tos-ops` | TOS (对象存储) | Object storage buckets |
| `ve-cms-ops` | CMS (云监控) | Monitoring, alerts |
| `ve-vke-ops` | VKE (容器服务) | Kubernetes clusters |

## Working with Skills

When working on a skill:
- Start from OpenAPI/official docs — never guess field names
- Cross-reference `ve` CLI coverage via `ve <service> --help`
- Use `cli_applicability: dual-path` when CLI supports the product
- Add safety gates for all destructive operations (delete, stop)
- Include ≥ 10 product-specific error codes in failure recovery

## CLI Installation

```bash
# Download from GitHub releases
curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/v1.0.42/ve-linux-amd64 -o /usr/local/bin/ve
chmod +x /usr/local/bin/ve
ve version
```

## Reference Files

- `ve-skill-generator/references/execution-environment.md` — CLI and Go SDK setup details
- `ve-skill-generator/references/cli-behavior.md` — Verified `ve` CLI conventions
- `ve-skill-generator/references/ve-skill-template.md` — Template for new skills
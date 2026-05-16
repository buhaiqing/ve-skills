---
name: ve-[product-name]-ops
description: >-
  Use when the user needs to deploy, configure, troubleshoot, or monitor Volcengine
  (火山引擎) [Product Name] — [Resource Type] lifecycle, configuration, and
  diagnostics. User mentions [Product Name], [Product Chinese Name],
  [Product Alias], or describes product-specific scenarios (e.g., connection
  drops, performance degradation, resource creation failures) even without
  naming the product directly. Not for billing, IAM, or related products that
  have their own ops skills.
license: MIT
compatibility: >-
  Official Volcengine CLI (`ve`, Go binary, no runtime), Go 1.21+ runtime
  (JIT SDK fallback; scripts compatible with Go 1.14+ syntax), valid API credentials,
  network access to Volcengine endpoints.
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-05-15"
  runtime: Harness AI Agent, Claude Code, Cursor, or compatible Agent runtimes
  go_script_syntax_minimum: "1.14"
  go_jit_runtime_version: "1.21+"
  api_profile: "[Paste OpenAPI title/version or doc link]"
  cli_applicability: dual-path
  cli_support_evidence: >-
    [If CLI covers this product: cite confirmation via `ve <product> --help`.
    If CLI does NOT cover: note JIT Go SDK fallback required.]
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

> This template follows the [Agent Skill OpenSpec](https://agentskills.io/specification).

# Volcengine [Product Name] Operations Skill

## Overview

[Product Name] on Volcengine (火山引擎) provides [brief description]. This skill is an **operational runbook** for agents: explicit scope, credential rules, pre-flight checks, **dual-path execution** (official **SDK/API** and, when the product is supported by official **`ve` CLI**, the matching **CLI** flows), response validation, and failure recovery. **Do not use the web console as the primary agent execution path** in `SKILL.md` or [Volcengine Console](https://console.volcengine.com).

> **UX Compliance:** This skill follows the [User Experience Specification](../references/user-experience-spec.md). All operations include onboarding guidance, minimal prompts, smart defaults, clear feedback, and user-friendly error handling.

### CLI applicability (repository policy)

- **`cli_applicability: dual-path`:** Official `ve` CLI supports this product. You **MUST** ship **`references/cli-usage.md`** and, in **each** execution flow below, document **both** the SDK step **and** the `ve` CLI step for every operation the CLI exposes. If the CLI covers **only part** of the API, add a **coverage gap** table (SDK-only operations) in `references/cli-usage.md`.
- **`cli_applicability: sdk-only`:** Official `ve` CLI does **not** expose this product. **Omit** `references/cli-usage.md`. Keep **`cli_support_evidence`** pointing at official proof. SDK/API remains mandatory for all operations.

## Five Core Standards (Quality Gates)

Every generated skill MUST satisfy these five standards. Use them as a design checklist during population:

| # | Standard | How This Skill Fulfills It | Concrete Validation Criteria |
|---|----------|---------------------------|------------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use conditions with precise triggers and delegation rules | ≥ 3 SHOULD entries with specific triggers; ≥ 3 SHOULD NOT entries with named delegation targets |
| 2 | **Structured I/O** | Placeholder conventions (`{{env.*}}`, `{{user.*}}`, `{{output.*}}`) with type and source documented | Zero bare variable names; every input uses a typed placeholder; every output maps to a JSON path |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover, with numbered imperative steps | ≥ 1 operation with all 4 phases present; all steps numbered and imperative (not descriptive) |
| 4 | **Complete Failure Strategies** | Error taxonomy table with ≥ 10 product-specific codes; HALT vs retry per error type | Error table has ≥ 10 rows; each row has: code, max retries, backoff, agent action, UX template |
| 5 | **Absolute Single Responsibility** | One product, one primary resource model; cross-product delegation to other skills | SKILL.md covers exactly 1 product; cross-product ops delegate (not duplicate); naming follows `ve-[product]-ops` |

Refer to the [meta-skill](../SKILL.md#five-core-standards-quality-gates) for detailed descriptions of each standard.

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- User mentions "Volcengine [Product Name]" OR "[Product Chinese Name]" OR "[Product Alias]"
- Task involves CRUD or lifecycle operations on **[Resource Type]** (create, describe, modify, delete, list, and product-specific actions)
- Task keywords: [keyword1], [keyword2], [keyword3], …
- User asks to deploy, configure, troubleshoot, or monitor [Product Name] **via API, SDK, CLI, or automation**

### SHOULD NOT Use This Skill When

- Task is purely billing / account management → delegate to: `ve-billing-ops` (when present)
- Task is IAM / permission model only → delegate to: `ve-iam-ops` (when present)
- Task is about **[related product]** → delegate to: `ve-[other]-ops`
- User insists on **console-only** flows with no API → state limitation; do not invent undocumented HTTP steps

### Delegation Rules

- If resource B depends on resource A, complete or verify A (via the A skill) before B's SDK or CLI steps.
- Multi-product requests: handle each product with its skill; do not merge unrelated APIs into one ambiguous flow.

## Variable Convention (Agent-Readable)

Structured placeholders reduce injection ambiguity and unsafe prompts:

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_REGION}}` | From runtime environment | Use documented default only if skill explicitly allows |
| `{{user.region}}` | User-supplied region | Ask once; reuse |
| `{{user.resource_name}}` | User-supplied name | Ask once; reuse |
| `{{output.resource_id}}` | From last API or CLI JSON response | Parse per **OpenAPI** (SDK) or **verified CLI** path for this operation |

> **`{{env.*}}` MUST NOT** be collected from the user. **`{{user.*}}`** MUST be collected interactively when missing.

> **Security Warning (Credential Masking — MANDATORY):** **NEVER** log, print, or expose `VOLCENGINE_SECRET_KEY`, `secret_key`, `SecretKey`, or any credential field value in console output, debug messages, error messages, or logs.
>
> **Masking rules across all execution paths:**
> | Execution Path | Safe Pattern | Unsafe Pattern |
> |----------------|-------------|----------------|
> | Console output | `VOLCENGINE_SECRET_KEY=<masked>` | `VOLCENGINE_SECRET_KEY=AKLT...` |
> | Error messages | `Error: API call failed (credential omitted)` | `Error: InvalidSecretKey.XXX ... actual secret...` |
> | Log files | `[INFO] Credentials configured: SecretKey=***` | `[INFO] SK: AKLT...` |
> | Verification | `test -n "$VOLCENGINE_SECRET_KEY" && echo "✅ SecretKey is set"` | `echo "SecretKey=$VOLCENGINE_SECRET_KEY"` |
> | JIT Go SDK | `instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))` (env read is safe) | `fmt.Printf("Config: %+v", config)` or `log.Printf("%+v", ...)` |
> | Debug/verbose | `⚠️ Debug mode may expose credential values` (warning only) | `--debug` with un-masked credential output |
>
> **Credential verification MUST check existence only**, never echo the value:
> - Bash: `test -n "$VOLCENGINE_SECRET_KEY"` ✅ | `echo $VOLCENGINE_SECRET_KEY` ❌
> - Go: `if os.Getenv("VOLCENGINE_SECRET_KEY") == ""` ✅ | `fmt.Println(os.Getenv("VOLCENGINE_SECRET_KEY"))` ❌
>
> **If any execution flow violates this rule, the skill SHALL be blocked from merge as a security incident.**

## API and Response Conventions (Agent-Readable)

- **OpenAPI/official docs are canonical** for path, query, body fields, enums, and response shapes. Replace generic JSON paths below with **real** schema field names.
- **Errors:** Map SDK/HTTP errors to `code` / `status` / message fields per spec. Do not assume a single global shape across products.
- **Timestamps:** ISO 8601 with timezone when the API returns strings (e.g. `2026-04-28T10:00:00+08:00`).
- **Idempotency:** Document client request tokens, duplicate names, and `ResourceAlreadyExists` behavior per API.

### Example Response Field Table (Replace with OpenAPI-Accurate Paths)

| Operation | JSON Path (example) | Type | Description |
|-----------|---------------------|------|-------------|
| Create | `$.Result.InstanceId` | string | New resource ID (verify name in spec) |
| Describe | `$.Result.Status` | string | Lifecycle state |
| List | `$.Result.Items[].Id` | array | IDs (verify array structure) |
| Modify / Delete | `$.Metadata.RequestId` or `$.Error` | string / object | Per spec |

### Expected State Transitions (Adjust to Product)

| Operation | Initial State | Target State | Poll Interval | Max Wait |
|-----------|---------------|--------------|---------------|----------|
| Create | — | `running` or product equivalent | 5s | 300s |
| Start | `stopped` | `running` | 5s | 120s |
| Stop | `running` | `stopped` | 5s | 120s |
| Delete | any stable state | absent or `deleted` per describe | 5s | 300s |

## Quick Start

### What This Skill Does
This skill enables you to deploy, configure, troubleshoot, and monitor [Product Name] resources on Volcengine (火山引擎) using the `ve` CLI (primary) or JIT Go SDK (fallback).

### Prerequisites
- [ ] `ve` CLI installed (or Go runtime for JIT fallback)
- [ ] Credentials configured: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`
- [ ] Region set: `VOLCENGINE_REGION`

### Verify Setup
```bash
# Check CLI and credentials
ve [product] Describe[Resources] --Region {{env.VOLCENGINE_REGION}}
```

### Your First Command
```bash
# Example: List resources
ve [product] Describe[Resources] --Region {{env.VOLCENGINE_REGION}}
```

### Next Steps
- [Core Concepts](references/core-concepts.md) — Understand [Product Name] architecture
- [Common Operations](#execution-flows) — Create, manage, and delete resources
- [Troubleshooting](references/troubleshooting.md) — Fix common issues

## Capabilities at a Glance

| Operation | Description | Complexity | Risk Level |
|-----------|-------------|------------|------------|
| Create | Create a new [Resource] | Medium | Low |
| Describe | View [Resource] details | Low | None |
| Modify | Change [Resource] configuration | Medium | Medium |
| Delete | Remove a [Resource] | Low | **High** — irreversible |
| List | View all [Resources] | Low | None |

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-15 | Initial template with ve CLI and JIT Go SDK support |

## Execution Flows (Agent-Readable)

Every operation: **Pre-flight → Execute (SDK/API and, when applicable, `ve` CLI) → Validate → Recover**. Do not skip phases.

**Preference hint:** When CLI does not support a specific operation, JIT build a Go SDK script. CLI is preferred for coverage and simplicity; Go SDK is used for operations CLI does not expose.

### Operation: Create [Resource]

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| SDK / deps | Import client from `github.com/volcengine/volc-sdk-golang` | No import error | Document install pin |
| CLI / deps | `ve version` (**required** when `cli_applicability: dual-path`) | Exit code 0 | Document CLI install |
| Credentials | Verify env vars `VOLCENGINE_ACCESS_KEY` and `VOLCENGINE_SECRET_KEY` are set | Non-empty keys | HALT; user configures env |
| Region | Verify `{{user.region}}` is a valid Volcengine region | Region supported | Suggest valid region |
| Quota | Call quota/describe API per OpenAPI | Sufficient quota | HALT; user raises quota |

#### Execution — CLI (`ve`) (Primary Path)

Use the [Volcengine CLI](https://github.com/volcengine/volcengine-cli) as the **primary execution path**.

> **Critical CLI Notes** (verified through source code analysis):
> - Since v1.0.20, command prefix is `ve` (was `volcengine-cli`)
> - Output is **JSON by default**
> - `--help` shows parameter list for any action: `ve <service> <action> --help`
> - Credentials can be passed via environment variables `VOLCENGINE_ACCESS_KEY` / `VOLCENGINE_SECRET_KEY` / `VOLCENGINE_REGION`
> - For API calls: `ve <service> <Action> --ParameterName value`
> - JSON body can be passed directly: `ve rds_mysql ModifyDBInstanceIPList --body '{"InstanceId":"xxx"}'`

```bash
# API call (JSON output by default)
ve [product] Create[Resource] \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --[ParamName] "value"
  # See `ve [product] Create[Resource] --help` for full parameter list
```

#### Execution — JIT Go SDK (Fallback Path)

When `ve` CLI does not support a specific operation, **JIT build a Go SDK script** dynamically:

```go
// main.go (generated dynamically in /tmp/ve-sdk-workspace)
package main

import (
    "fmt"
    "os"
    
    "github.com/volcengine/volc-sdk-golang/base"
    "[product]" "github.com/volcengine/volc-sdk-golang/service/[product]"
)

func main() {
    // Initialize service instance
    instance := [product].NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))
    
    // Prepare request parameters
    params := make(map[string]interface{})
    params["Region"] = os.Getenv("VOLCENGINE_REGION")
    // Add fields per OpenAPI request schema
    
    // Make API call
    resp, err := instance.Client.Request("[product]", "Create[Resource]", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "API call failed: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println(string(resp))
}
```

Execute:
```bash
# In /tmp/ve-sdk-workspace
go mod init ve-sdk-script
go get -u github.com/volcengine/volc-sdk-golang
go run ./main.go
```

#### Post-execution Validation

1. Read `{{output.resource_id}}` from the **documented** response path.
2. Poll **Describe** until terminal success state or timeout:

```bash
# CLI polling with shell loop
for i in $(seq 1 60); do
  STATUS=$(ve [product] Describe[Resource] --[IdName] "{{output.resource_id}}" | jq -r '.Result.Status')
  [ "$STATUS" = "Running" ] && break
  sleep 5
done
```

3. On success, report `{{output.resource_id}}` and key fields to the user.
4. On terminal failure, go to **Failure Recovery**.

#### Failure Recovery

| Error pattern (from API/SDK or parsed CLI JSON) | Max retries | Backoff | Agent Action | UX Feedback |
|------------------------------|-------------|---------|--------------|-------------|
| `InvalidParameter` / 400 invalid input | 0–1 | — | Fix args from OpenAPI; retry once if safe | `[ERROR] InvalidParameter: The request parameter is invalid. What happened: One or more parameters do not meet the API specification. How to fix: Check the parameter against OpenAPI docs and retry. Next step: Review the parameter table above.` |
| `QuotaExceeded` / resource quota limit | 0 | — | HALT | `[ERROR] QuotaExceeded: Resource quota limit reached. What happened: Your account has reached the maximum allowed number of this resource type. How to fix: Delete unused resources or request a quota increase. Next step: Contact support or delete unused resources.` |
| `InsufficientBalance` / balance insufficient | 0 | — | HALT | `[ERROR] InsufficientBalance: Account balance insufficient. What happened: Your account does not have enough balance to complete this operation. How to fix: Recharge your account. Next step: Go to Volcengine billing console to recharge.` |
| `ResourceAlreadyExists` | 0 | — | Ask reuse vs new name | `[ERROR] ResourceAlreadyExists: A resource with this name already exists. What happened: The specified resource name is already in use. How to fix: Use a different name or reuse the existing resource. Next step: Choose a unique name or describe the existing resource.` |
| Throttling / 429 | 3 | exponential | Back off; respect `Retry-After` if present | `⚠️ Rate limit reached. Retrying in {backoff}s... (Attempt {current}/{max})` |
| `InternalError` / 5xx | 3 | 2s, 4s, 8s | Retry; then HALT with correlation id (RequestId) if any | `[ERROR] InternalError: Server-side error occurred. What happened: Volcengine encountered an internal error processing your request. How to fix: Retry the operation. If it persists, escalate with RequestId. Next step: Retry now or escalate with RequestId: {RequestId}.` |

### Operation: Describe [Resource]

#### Execution

Use the SDK **describe** or **get** API matching OpenAPI. When **`cli_applicability: dual-path`**, also document the equivalent `ve [product] Describe[Resource] ...`, passing `{{user.resource_id}}` and region.

```bash
# CLI — plain JSON (default output format)
ve [product] Describe[Resource] --Region cn-beijing --[IdName] "{{user.resource_id}}"
```

#### Present to User

| Field | Path (example) | Notes |
|-------|----------------|-------|
| ID | from describe result | Plain text |
| Name | from describe result | Plain text |
| Status | from describe result | Human-readable state |
| Created time | from describe result | Format ISO per API |

### Operation: Delete [Resource]

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: irreversible delete of `{{user.resource_name}}` (`{{user.resource_id}}`).
- **MUST NOT** proceed without clear user assent.

#### Execution

Call delete API per OpenAPI. When **`cli_applicability: dual-path`**, also document the `ve` delete subcommand; capture `requestId`, success flag, or error per **verified** output shape for **each** path.

#### Post-execution Validation

Poll describe (or head/get) until **404**, **NotFound**, or status indicates deleted—per API semantics—within **max wait**.

## Prerequisites

1. **Install `ve` CLI** (primary execution path — static Go binary, no runtime dependencies):

   ```bash
   # Download from GitHub releases
   # Latest releases: https://github.com/volcengine/volcengine-cli/releases
   
   # Example for macOS (ARM64 / Apple Silicon)
   curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/v1.0.42/ve-darwin-arm64 -o /usr/local/bin/ve
   chmod +x /usr/local/bin/ve

   # Example for macOS (x86_64 / Intel)
   curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/v1.0.42/ve-darwin-amd64 -o /usr/local/bin/ve
   chmod +x /usr/local/bin/ve

   # Example for Linux (x86_64)
   curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/v1.0.42/ve-linux-amd64 -o /usr/local/bin/ve
   chmod +x /usr/local/bin/ve

   # Example for Linux (ARM64)
   curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/v1.0.42/ve-linux-arm64 -o /usr/local/bin/ve
   chmod +x /usr/local/bin/ve
   ```

   > **Note:** CLI binary naming follows pattern `ve-{os}-{arch}`. Check [releases page](https://github.com/volcengine/volcengine-cli/releases) for the latest version and available binaries.

   ```bash
   # Verify installation
   ve version
   ```

2. **Bootstrap Go runtime** (for JIT SDK fallback — only needed if CLI does not support operation):

   ```bash
   # Check if Go exists
   if ! command -v go &> /dev/null; then
       # JIT download Go 1.21 (auto-detects OS and architecture)
       OS=$(uname -s | tr '[:upper:]' '[:lower:]')
       ARCH=$(uname -m)
       [ "$ARCH" = "x86_64" ] && ARCH="amd64"
       [ "$ARCH" = "aarch64" ] && ARCH="arm64"
       
       mkdir -p /tmp/go-runtime
       curl -fsSL "https://go.dev/dl/go1.21.0.${OS}-${ARCH}.tar.gz" | tar -xz -C /tmp/go-runtime
       
       # Set environment variables
       export PATH="/tmp/go-runtime/go/bin:$PATH"
       export GOPATH="/tmp/go-workspace"
       export GOCACHE="/tmp/go-cache"
       export GOMODCACHE="/tmp/go-modcache"
       export GOPROXY="https://goproxy.cn,direct"  # China CDN mirror
   fi
   
   go version
   ```

   > Go version strategy: **JIT download Go 1.21+**, **Script compatibility Go 1.14+** (minimum).

3. **Configure Credentials** — Environment variables (recommended for Agent execution):

   ```bash
   export VOLCENGINE_ACCESS_KEY="{{env.VOLCENGINE_ACCESS_KEY}}"
   export VOLCENGINE_SECRET_KEY="{{env.VOLCENGINE_SECRET_KEY}}"
   export VOLCENGINE_REGION="{{env.VOLCENGINE_REGION}}"
   ```

   **Alternative — Interactive CLI Configuration:**
   ```bash
   ve configure set --profile default --region cn-beijing --access-key "{{user.access_key}}" --secret-key "{{user.secret_key}}"
   ```

   **Alternative — Config File (`~/.volcengine/config.json`):**
   ```bash
   mkdir -p ~/.volcengine
   cat > ~/.volcengine/config.json << 'CONFIGEOF'
   {
     "current": "default",
     "profiles": [
       {
         "name": "default",
         "mode": "AK",
         "access_key": "{{user.access_key}}",
         "secret_key": "{{user.secret_key}}",
         "region": "{{user.region}}"
       }
     ]
   }
   CONFIGEOF
   ```

4. **Verify Configuration**:
   ```bash
   # Quick validation (JSON output by default)
   ve [product] Describe[Resources] --Region "{{env.VOLCENGINE_REGION}}"
   ```

> **Security:** Never commit `.env` to version control (already in `.gitignore`). All credentials use `{{env.*}}` placeholders in generated Skills — never real values.

## Reference Directory

- [Core Concepts](references/core-concepts.md)
- [API & SDK Usage](references/api-sdk-usage.md)
- [CLI Usage](references/cli-usage.md) (**required** when `cli_applicability: dual-path`; omit only for `sdk-only` with evidence in frontmatter)
- [Troubleshooting Guide](references/troubleshooting.md)
- [Monitoring & Alerts](references/monitoring.md)
- [Integration](references/integration.md)
- [Knowledge Base](references/knowledge-base.md) — fault pattern library (AIOps diagnostic skills)
- [Observability Integration](references/observability.md) — Metrics/Logs/Traces linkage (AIOps diagnostic skills)
- [Prompts Handbook](references/prompts.md) — common prompt templates (AIOps diagnostic skills)
- [User Experience Specification](references/user-experience-spec.md) — mandatory UX compliance reference
- [AIOps Best Practices](references/aiops-best-practices.md) — mandatory AIOps patterns for monitoring/diagnosis skills
- [Optimization Analysis](references/optimization-analysis.md) — three-dimensional optimization framework
- [Execution Environment Setup](references/execution-environment.md) — CLI install, Go JIT download, credential config, verification
- [CLI Behavioral Reference](references/cli-behavior.md) — verified `ve` CLI conventions (JSON output, env vars, invocation patterns)
- [Enhanced Self-Healing Framework](references/enhanced-self-healing-framework.md) — **MANDATORY** self-healing patterns for all installation flows

## Operational Best Practices

- **Least privilege:** IAM policies scoped to required APIs only.
- **Availability:** Multi-AZ or product-specific HA patterns per docs.
- **Cost:** Right-size resources; use product cost controls where applicable.

---

# Appendix: Reference File Templates

## references/troubleshooting.md

```markdown
# Troubleshooting [Product Name]

## Common API Error Codes
| Code / HTTP | Meaning | Agent Action |
|-------------|---------|--------------|
| InvalidParameter | Request failed validation | Align body with OpenAPI |
| Forbidden.RAM | Insufficient IAM permissions | User adds IAM policy |
| InternalError | Server-side error | Retry with backoff; then HALT |

## Diagnostic Order
1. Describe resource by ID.
2. List related resources if API supports filters.
3. Check regional endpoint and `Region` consistency.
4. Verify CLI metadata coverage: `ve [product] --help`
```

## references/api-sdk-usage.md

```markdown
# API & SDK — [Product Name]

## OpenAPI
- Spec: [link or path]
- Base path and version: …

## SDK Operations Map
| Goal | API operationId | SDK method (if known) |
|------|-----------------|------------------------|
| Create | … | … |
| Describe | … | … |

## Request / Response Notes
- Required fields: …
- Pagination: …
```

## references/cli-usage.md

```markdown
# CLI — [Product Name] (`ve`)

## Install and config
- Install: see [Volcengine CLI](https://github.com/volcengine/volcengine-cli)
- **CRITICAL Credentials:** The `ve` CLI reads from env vars `VOLCENGINE_ACCESS_KEY` / `VOLCENGINE_SECRET_KEY` OR `~/.volcengine/config.json` (JSON format).
- For sandbox environments, set env vars directly (preferred).

## Conventions (agent execution)
- Output is **JSON by default**
- Document **exact** JSON paths after verifying with a real invocation
- CLI invocation: `ve <service> <action> --parameter value`

## CLI vs API coverage gap
| Operation (API / SDK) | Available via `ve`? | Notes |
|------------------------|---------------------|-------|
| Create | yes / no | … |
| Describe | yes / no | … |

## Command map
| Goal | Example `ve` invocation | Notes |
|------|--------------------------|-------|
| Create | `ve [product] Create[Resource] --Region cn-beijing` | JSON output by default |
| Describe | `ve [product] Describe[Resource] --Region cn-beijing` | JSON output by default |
```

## references/monitoring.md

```markdown
# Monitoring [Product Name]

## Key Metrics (examples — replace with product namespaces)
- Metric A: `Volcengine_[product]_[metric]`
- Metric B: `Volcengine_[product]_[metric]`

## Alert Example (structure only)
```

## references/integration.md

````markdown
# Integration

## Environment Setup

**Primary path:** `ve` CLI (static Go binary, no runtime dependencies)

**Fallback path:** JIT Go SDK (dynamic script generation + `go run`)

### Go Runtime Bootstrap

If Agent Runtime lacks Go, JIT download from official source:

```bash
# Check Go runtime
if ! command -v go &> /dev/null; then
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    [ "$ARCH" = "x86_64" ] && ARCH="amd64"
    [ "$ARCH" = "aarch64" ] && ARCH="arm64"
    
    mkdir -p /tmp/go-runtime
    curl -fsSL "https://go.dev/dl/go1.21.0.${OS}-${ARCH}.tar.gz" | tar -xz -C /tmp/go-runtime
    export PATH="/tmp/go-runtime/go/bin:$PATH"
    export GOPATH="/tmp/go-workspace"
    export GOCACHE="/tmp/go-cache"
fi

go version
```

> **Go version strategy:**
> - **JIT download:** Go 1.21+ (stable)
> - **Script compatibility:** Go 1.14+ (minimum)

### JIT Go SDK Workflow

1. **Initialize workspace:**
   ```bash
   mkdir -p /tmp/ve-sdk-workspace
   cd /tmp/ve-sdk-workspace
   go mod init ve-sdk-script
   ```

2. **Get dependencies:**
   ```bash
   # Set proxy for China CDN mirror (faster download)
   export GOPROXY="https://goproxy.cn,direct"
   
   # Volcengine SDK
   go get -u github.com/volcengine/volc-sdk-golang
   ```

3. **Generate script** (Agent dynamically creates operation-specific .go file)

4. **Execute:**
   ```bash
   go run ./main.go
   ```

### SDK Package Structure

| Product | Go SDK Package |
|---------|---------------|
| ECS | `github.com/volcengine/volc-sdk-golang/service/ecs` |
| RDS MySQL | `github.com/volcengine/volc-sdk-golang/service/rds_mysql` |
| VPC | `github.com/volcengine/volc-sdk-golang/service/vpc` |
| Redis | `github.com/volcengine/volc-sdk-golang/service/redis` |

> Find package names at: https://github.com/volcengine/volc-sdk-golang/tree/main/service

## SDK Script Template

```go
// main.go (single-file script template)
package main

import (
    "fmt"
    "os"
    
    "[product]" "github.com/volcengine/volc-sdk-golang/service/[product]"
)

func main() {
    // Initialize service instance
    instance := [product].NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))
    
    // Prepare request parameters
    params := make(map[string]interface{})
    params["Region"] = os.Getenv("VOLCENGINE_REGION")
    
    // Make API call
    resp, err := instance.Client.Request("[product]", "[ActionName]", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "API call failed: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println(string(resp))
}
```

> Use `os.Getenv("KEY")` for all credentials. Never hardcode secrets in scripts.
````

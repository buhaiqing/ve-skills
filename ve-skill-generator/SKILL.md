---
name: ve-skill-generator
description: >-
  Use when the user needs to create or update a Volcengine (火山引擎) Agent Skill
  (`ve-[product]-ops`) in this repository — even if they don't explicitly ask for
  scaffolding or generation. Triggers include: user wants to "add a skill for
  product X", "regenerate from OpenAPI", "generate a Volcengine ECS/RDS skill",
  or "fix gaps found during review". Also use when an existing skill needs
  realignment after API doc changes or fails a governance/adversarial review.
  Not for executing live changes against cloud accounts or for one-off debugging
  with no intent to maintain.
license: MIT
compatibility: >-
  Access to Volcengine (火山引擎) official documentation, OpenAPI/Swagger for the product,
  `ve-skill-generator/references/ve-skill-template.md`,
  `references/evaluation-driven-workflow.md`,
  `references/governance-and-adversarial-review.md` (when present),
  `references/prompt-library.md` (structured prompt repository),
  `references/optimization-analysis.md` (three-dimensional optimization framework),
  `references/user-experience-spec.md` (mandatory UX requirements for generated skills),
  `references/execution-environment.md` (CLI + Go SDK setup details),
  `references/cli-behavior.md` (verified `ve` CLI behavioral notes),
  `references/rubric.md` (GCL rubric for generated skills),
  `references/prompt-templates.md` (GCL prompt skeletons for generated skills),
  and agentskills.io frontmatter conventions.
metadata:
  author: volcengine
  version: "1.3.0"
  last_updated: "2026-07-05"
  runtime: Harness AI Agent, Claude Code, Cursor, or compatible Agent runtimes
  type: meta-skill
  guidance_freedom_level: medium
  go_script_syntax_minimum: "1.14"
  go_jit_runtime_version: "1.21+"
  cli_applicability: sdk-only
---

# Volcengine Skill Generator (Meta-Skill)

## Quick Start

### What This Skill Does
Scaffolds new or updates existing `ve-[product]-ops` skills in this repository, based on official Volcengine OpenAPI specs. This is a **meta-skill** — it generates runbooks for agents, not operational execution against cloud accounts.

### Prerequisites
- [ ] Access to OpenAPI/Swagger spec for the target Volcengine product
- [ ] Read access to this repository's template files
- [ ] Network access to Volcengine documentation URLs

### Your First Generation
```
Input: "Generate ve-ecs-ops for ECS instances, disks, and snapshots"
Output: ve-ecs-ops/ directory with SKILL.md and references/
```

### Next Steps
- [Generation Workflow](#evaluation-driven-generation-workflow) — Step-by-step generation process
- [Anti-Pattern Checklist](#anti-pattern-checklist) — Common mistakes to avoid
- [P0/P1 Checklist](#p0--must-pass) — Quality gates for generated skills

---

## Overview

This **meta-skill** defines **how** to author a new **product-scoped** operational skill (e.g. `ve-ecs-ops`) **inside this repo**. It does **not** perform maintenance against a user's cloud account. Live work uses the generated `ve-[product]-ops` skills (official `ve` CLI with **JIT Go SDK fallback**).

### Guidance Freedom Level: Medium (Provide Templates)

This meta-skill operates at **Medium** guidance level: it provides **templates and frameworks** ([ve-skill-template.md](references/ve-skill-template.md), prompt library, UX spec) while allowing the agent to adapt based on product-specific context. Low-level scripts (CLI installation, Go runtime JIT download) are detailed in [references/execution-environment.md](references/execution-environment.md).

### Core Principle

Generated skills are **agent-readable runbooks**: triggers, env vs user placeholders, pre-flight → execute → validate → recover, safety gates, and outputs **grounded in OpenAPI and verified CLI behavior**, not guessed.

### Technology Stack
- **CLI:** `ve` (Volcengine CLI, Go binary, static, no dependencies) — primary execution path. Command prefix changed from `volcengine-cli` to `ve` since v1.0.20.
- **SDK:** Volcengine Go SDK (`github.com/volcengine/volc-sdk-golang`) — JIT fallback
- **JIT execution:** `go run` (script mode, dynamic generation)

### Repository Scope
All generated layout and policies apply **only** to the `ve-skills` monorepo unless explicitly stated elsewhere.

---

## Role Boundary (Agent-Readable)

| This meta-skill **does** | This meta-skill **does not** |
|--------------------------|------------------------------|
| Choose **extend** vs **new** `ve-[product]-ops` | Replace deep product knowledge already in an existing ops skill |
| Scaffold `SKILL.md`, `references/*`, `assets/*` from the template | Call Volcengine APIs on behalf of the user |
| Enforce naming, frontmatter, P0/P1, delegation, and **governance** hooks | Invent request/response fields or CLI flags without official doc verification |
| Point authors to **adversarial review** before merge (when governance doc exists) | Store or echo real credentials |

If the user wants **operational execution** (e.g. "create a resource"), load the appropriate `ve-*-ops` skill for that product — not this generator.

---

## When to Use / Not Use

### Use When
- A new Volcengine product needs a **first** ops skill in **this repo**
- An existing skill lacks P0 elements (triggers, placeholders, flows, recovery, destructive gates)
- OpenAPI or official docs changed; the skill should be **realigned** (bump version/changelog)
- A contributor needs the **standard directory layout** for a new `ve-[product]-ops`

### Do NOT Use When
- One-off debugging with no intent to maintain a reusable skill
- Non-Volcengine application work
- You only need billing/IAM execution — use dedicated ops skills when they exist

---

## Input / Output Structure

### Input

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `product.name` | string | Yes | English product name (e.g., ECS, RDS MySQL) |
| `product.slug` | string | Yes | CLI service slug (e.g., ecs, rds_mysql) — verify via `ve <slug> --help` |
| `product.chinese_name` | string | No | Chinese name for trigger matching (e.g., 云服务器) |
| `primary_resource` | string | Yes | Primary resource type (e.g., Instance, DBInstance) |
| `api_service_id` | string | Yes | API service identifier from OpenAPI tags or SDK package |
| `openapi_url` | string | Recommended | OpenAPI/Swagger URL or path — required for API-accurate fields |
| `operation_list` | string[] | Yes | List of operations (create, describe, modify, delete, list, product-specific) |
| `doc_urls` | string[] | Recommended | Official documentation URLs |
| `cli_support_evidence` | string | Yes | Confirmation that `ve` exposes this product (or JIT SDK fallback needed) |

### Output

| Artifact | Description |
|----------|-------------|
| `ve-[product]-ops/SKILL.md` | Main skill runbook — triggers, flows, recovery, safety gates |
| `references/core-concepts.md` | Architecture, limits, regions, quotas |
| `references/api-sdk-usage.md` | Operation map, required fields, pagination, request/response snippets |
| `references/cli-usage.md` | `ve` CLI cheat sheet (primary path) — when `cli_applicability: dual-path` |
| `references/troubleshooting.md` | Error codes, ordered diagnostics |
| `references/monitoring.md` | Metrics, dashboards, alerts (when monitoring in scope) |
| `references/integration.md` | Go bootstrap, JIT SDK setup, env vars |
| `assets/example-config.yaml` | Example configuration |

---

## Five Core Standards (Quality Gates)

Every generated skill MUST satisfy these five standards. Reference them throughout the generation workflow.

### Standard 1: Clear Boundaries (边界明确)
- **SHOULD use** conditions: precise, with keywords and intent matching
- **SHOULD NOT use** conditions: explicit negative cases that prevent misfire
- **Delegation rules**: clear pointers to related skills

### Standard 2: Structured I/O (输入输出结构化)
- Input parameters defined with types and sources (`{{env.*}}`, `{{user.*}}`)
- Output fields defined with JSON paths from OpenAPI response schemas
- Placeholder conventions: `{{env.*}}` (from runtime, NEVER ask user), `{{user.*}}` (interactive collect), `{{output.*}}` (from API response)

### Standard 3: Explicit Actionable Steps (步骤明确可执行)
- Every operation: Pre-flight → Execute → Validate → Recover
- Steps are numbered, imperative, specific — not descriptive summaries
- CLI and SDK paths documented separately when both apply

### Standard 4: Complete Failure Strategies (失败策略完备)
- Error taxonomy with product-specific error codes (≥ 10)
- Each error pattern: max retries, backoff strategy, agent action, UX feedback
- HALT vs retry distinction; credential, quota, and business errors clearly separated

### Standard 5: Absolute Single Responsibility (职责绝对单一)
- One skill = one product = one primary resource model
- Cross-product delegation: document in Trigger & Scope, do NOT duplicate full flows
- Naming: `ve-[product]-ops` (lowercase, hyphenated)

---

## Anti-Pattern Checklist

Before and during generation, check against these common anti-patterns:

| # | Anti-Pattern | How It Manifests | Correction |
|---|-------------|-----------------|------------|
| 1 | ❌ **Skill = Prompt** | Writing conversational instructions instead of executable steps | Use imperative numbered steps; define I/O; separate triggers from execution |
| 2 | ❌ **Skill = Human Doc** | Explaining concepts instead of instructing the agent | Use model-parsable structured language; define behavior boundaries |
| 3 | ❌ **Feature Bundling** | One skill tries to do everything (create + monitor + backup + billing) | Split into single-responsibility skills; delegate to existing skills |
| 4 | ❌ **API Hallucination** | Inventing field names, JSON paths, or CLI flags not in official docs | Cross-reference every field against OpenAPI or verified CLI output |
| 5 | ❌ **Credential Leaking** | Printing, logging, or echoing secret values in any execution path | Mask all credentials with `***` / `<masked>`; check existence only |
| 6 | ❌ **No Safety Gate** | Destructive operations (delete, stop, release) without explicit confirmation | Add confirmation step before every destructive path (CLI + SDK) |
| 7 | ❌ **Hardcoded Values** | Regions, timeouts, or limits baked into instructions | Use `{{env.*}}` / `{{user.*}}` placeholders; document defaults separately |
| 8 | ❌ **Missing Failure Path** | Only documenting the success path; no error handling | Add failure recovery table with error codes, retry logic, HALT conditions |
| 9 | ❌ **Over-Engineering** | Adding advanced features before core flow works | Follow evaluation-driven approach: start minimal, expand step by step |
| 10 | ❌ **Redundant Redundancy** | Repeating the same info across SKILL.md and references | SKILL.md is entry point; references provide depth — no duplication |

---

## Evaluation-Driven Generation Workflow

This workflow follows the **"fail first, evaluate first"** principle: define what "good" looks like before generating. At each critical node, validate the output and loop back for corrections.

> **Copy the checklist below before starting, and mark each step as you complete it.**

### Workflow Checklist

```
[ ] Step 1: Define Evaluation Targets — What does success look like?
[ ] Step 2: Analyze Sources — Extract operations, fields, errors from OpenAPI
    ↓ [Feedback Loop: Sources complete? If gaps found → research, then return]
[ ] Step 3: Scaffold Layout — Create directory from template
[ ] Step 4: Populate SKILL.md — Fill template with verified data
    ↓ [Feedback Loop: Five core standards satisfied? If not → fix and re-verify]
[ ] Step 5: Fill Reference Files — Complete all references/
    ↓ [Feedback Loop: All files populated? If gaps → fix]
[ ] Step 6: Verify & Review — P0/P1 checklist + adversarial review
    ↓ [Feedback Loop: Any failures? → return to Step 4 or 5; re-verify after fix]
[ ] Step 7: Final Anti-Pattern Check — Run anti-pattern checklist above
```

---

### Step 1: Define Evaluation Targets

Before generating anything, define **3-5 evaluation cases** for the target skill. Each case has a clear PASS/FAIL criterion.

**Template:**
```markdown
| ID | Scenario | Expected Behavior | PASS Condition |
|----|----------|-----------------|----------------|
| E1 | User asks to create a resource with minimal input | Skill prompts for required fields, uses smart defaults for optional | ≤ 2 prompts before execution |
| E2 | User asks to delete a resource | Skill asks for explicit confirmation with resource identifier | Confirmation step present |
| E3 | API returns QuotaExceeded | Skill returns clear error message with remediation steps | Error follows `[ERROR] code → explanation → fix → next step` |
| E4 | User asks about a non-existent resource | Skill checks existence first, returns "not found" with list suggestion | Resource existence check in pre-flight |
| E5 | User asks for a related product operation (e.g., VPC when using ECS) | Skill delegates to the correct skill or documents the limitation | Delegation rule present in Trigger & Scope |
```

**Purpose:** These cases anchor the generation process. Every feature in the generated skill must trace back to at least one evaluation case.

---

### Step 2: Analyze Sources

Extract from OpenAPI and official docs:

- **Operations**: OperationIds grouped by resource tag
- **Parameters**: Required vs optional, types, enums, defaults
- **Response schemas**: JSON paths, terminal states, pagination
- **Error codes**: Product-specific error taxonomy (≥ 10 codes)
- **Async behavior**: Polling intervals, terminal state names
- **CLI coverage**: Which operations `ve` supports vs SDK-only

**Validation checkpoint:** Before proceeding, confirm:
- [ ] All operationIds are real (not invented)
- [ ] JSON paths are from actual response schemas
- [ ] Error codes are documented in OpenAPI or official docs
- [ ] `cli_applicability` is correctly determined (`dual-path` vs `sdk-only`)

---

### Step 3: Scaffold Directory Layout

```text
ve-[product]-ops/
├── SKILL.md
├── references/
│   ├── core-concepts.md
│   ├── api-sdk-usage.md
│   ├── cli-usage.md              # Required when cli_applicability: dual-path
│   ├── troubleshooting.md
│   ├── monitoring.md              # When monitoring in scope
│   └── integration.md
├── assets/
│   └── example-config.yaml
```

Add `references/idempotency-checklist.md` when retries or automation require idempotent behavior.

---

### Step 4: Populate SKILL.md

Base: [ve-skill-template.md](references/ve-skill-template.md).

Replace all `[Placeholder]` with product-specific content derived from Step 2. Every field, JSON path, and CLI command MUST be traceable to OpenAPI or verified CLI output.

**Frontmatter requirements:**
| Field | Rule |
|-------|------|
| `name` | `ve-[product]-ops` — lowercase, hyphens, ≤ 64 chars |
| `description` | Third person, triggers only (per OpenSpec) |
| `cli_applicability` | `dual-path` (CLI available) or `sdk-only` (JIT Go SDK only) |
| `cli_support_evidence` | Cite confirmation via `ve <product> --help` or official docs |

**Validation checkpoint (Five Core Standards):**
- [ ] **Boundary**: SHOULD/SHOULD NOT use conditions complete?
- [ ] **I/O**: All placeholders (`{{env.*}}`, `{{user.*}}`, `{{output.*}}`) correctly typed?
- [ ] **Steps**: Every operation has Pre-flight → Execute → Validate → Recover?
- [ ] **Failure**: Error taxonomy ≥ 10 codes, each with recovery action?
- [ ] **Single Responsibility**: One product, one resource model, clear delegation?

**If any standard fails → FIX before proceeding to Step 5.**

---

### Step 5: Fill Reference Files

| File | Content | Source |
|------|---------|--------|
| `core-concepts.md` | Architecture, limits, regions, quotas, resource relationships | Official docs |
| `api-sdk-usage.md` | Operation map, required fields, pagination, request/response snippets | OpenAPI |
| `cli-usage.md` | `ve` command map, coverage gap table, JSON output paths | Verified CLI output |
| `troubleshooting.md` | Error code table, ordered diagnostic steps, product-specific patterns | OpenAPI + experience |
| `monitoring.md` | Metrics, dashboards, alarms, anomaly patterns | Volcengine monitoring docs |
| `integration.md` | Go bootstrap, JIT SDK setup, dependency config | Execution environment |

**Validation checkpoint:** All reference files populated with real content (not template placeholders)?

---

### Step 6: Verify & Review

Run the [P0/P1 Checklist](#p0--must-pass) below against the generated skill. Run the [Adversarial Review](references/governance-and-adversarial-review.md) scenarios (when present).

**Reference File Existence Validation (automated):**
Before running P0/P1 checklist, verify all expected files exist and contain real content:
```bash
# Check all required files exist and are not empty
REQUIRED_FILES=(
  "ve-[product]-ops/SKILL.md"
  "ve-[product]-ops/references/core-concepts.md"
  "ve-[product]-ops/references/api-sdk-usage.md"
  "ve-[product]-ops/references/troubleshooting.md"
  "ve-[product]-ops/references/integration.md"
  "ve-[product]-ops/assets/example-config.yaml"
)

# Conditionally required files
if [ "$CLI_APPLICABILITY" = "dual-path" ]; then
  REQUIRED_FILES+=("ve-[product]-ops/references/cli-usage.md")
fi

for f in "${REQUIRED_FILES[@]}"; do
  if [ ! -s "$f" ]; then
    echo "FAIL: $f is missing or empty"
    exit 1
  fi
  # Check for unreplaced placeholders
  if grep -q '\[Product Name\]\|\[product\]\|\[Resource Type\]\|\[Placeholder\]' "$f"; then
    echo "FAIL: $f contains unreplaced placeholders"
    exit 1
  fi
done
```

**For any failure:**
1. Identify the gap
2. Return to Step 4 (SKILL.md) or Step 5 (references)
3. Fix the gap
4. Re-verify the full checklist

**Re-verify after fixes — do not skip re-runs.**

---

### Step 7: Final Anti-Pattern Check

Run the [Anti-Pattern Checklist](#anti-pattern-checklist) above against the generated skill. Every item must pass.

**If an anti-pattern is detected:**
- Document the instance
- Fix according to the "Correction" column
- Re-run the P0/P1 checklist

---

## Description Optimization & Generation Telemetry

Moved to **[references/generation-guidance.md](references/generation-guidance.md)** for better separation of concerns. Covers:
- Description optimization principles (agentskills.io)
- Eval query design and optimization loop
- Generation telemetry & observability metrics
- Version pinning strategy for CLI and SDK

**Quick summary:** The `description` field is the sole trigger mechanism. Target ≤ 600 characters. Create `assets/eval_queries.json` with ≥ 20 queries. Track generation metrics for continuous improvement.

---

## Before You Generate: Decisions

### Extend vs New Directory
- **Extend** same product and resource model (new operation section, paths, troubleshooting rows)
- **New** `ve-[product]-ops` when the **service/API surface** or **primary resource** is distinct

### Naming
- Pattern: `ve-[product]-ops` (lowercase, hyphenated)
- Search the repo for collisions before creating

### Dependencies
- Cross-product chains: document **delegation** in Trigger & Scope
- Avoid duplicating another product's full flows

### Sources of Truth
- **OpenAPI + official docs** beat forums and chat logs
- Pin an API/SDK profile in skill `metadata` or `references/integration.md`

### Secrets
- Only `{{env.*}}` **names** and documentation; never real keys or customer data
- Credential masking is MANDATORY — see [references/execution-environment.md](references/execution-environment.md#credential-security)

### CLI-First with JIT Go SDK Fallback
- Primary path: `ve` CLI (static Go binary, covers most APIs)
- Fallback path: JIT Go SDK (dynamic script + `go run`)
- Execution environment details: [references/execution-environment.md](references/execution-environment.md)

---

## Governance (Expert Recommendation)

**Minimal adversarial review** gives high return for low cost: it catches destructive-action shortcuts, credential leaks in instructions, and API hallucination **before** merge. Treat [governance-and-adversarial-review.md](references/governance-and-adversarial-review.md) (when present) as the **reviewer companion** to this meta-skill.

Optional later improvements: PR template checkbox linking to that doc; periodic check that CLI-documented skills stay aligned with OpenAPI when APIs change.

---

## Agent-Ready Quality Checklist

### Generator Self-Checks (P0 — for THIS meta-skill's process)

- [ ] **Template Integrity:** All `[Placeholder]` tokens replaced with product-specific content — no literal placeholder strings remain in output
- [ ] **Reference File Completeness:** Every file listed in Reference Directory exists and contains real content (not template stubs)
- [ ] **Frontmatter Valid:** `name` matches `ve-[product]-ops` pattern, `cli_applicability` is `dual-path` or `sdk-only`, `cli_support_evidence` cites real verification
- [ ] **Description Length:** Generated `description` is ≤ 1024 characters (target: ≤ 600)
- [ ] **JSON Validity:** `assets/eval_queries.json` is valid JSON with correct trigger classifications
- [ ] **No Credential Leakage:** No real AK/SK values appear anywhere in generated files (template or data)
- [ ] **File Naming Consistency:** All filenames use lowercase-hyphen convention; no underscores or mixed case
- [ ] **Anti-Pattern Clean:** Ran Anti-Pattern Checklist — all 10 items pass

### Generated Skill Requirements (P0 — the OUTPUT must satisfy these)

#### 1. Trigger & Scope
- [ ] SHOULD/SHOULD NOT use conditions present with specific delegation targets (≥ 3 SHOULD NOT entries)
- [ ] Delegation rules point to named skills (e.g., `ve-billing-ops`, `ve-iam-ops`)

#### 2. Variables & Security
- [ ] `{{env.*}}` placeholders use `VOLCENGINE_*` prefix; none collected from user
- [ ] `{{user.*}}` placeholders only for interactive inputs (name, region, etc.)
- [ ] No secret literals anywhere; credential masking pattern enforced
- [ ] SDK scripts use `os.Getenv()` — never print config structs or credential values

#### 3. Execution Flows
- [ ] Every critical operation: Pre-flight → Execute → Validate → Recover
- [ ] `ve` CLI documented as primary path; JIT Go SDK as fallback (when `cli_applicability: dual-path`)
- [ ] Safety gate before each destructive operation (both CLI and SDK paths)
- [ ] Timeouts for polling (default: 5s interval, 300s max wait)

#### 4. Failure Recovery
- [ ] Error taxonomy ≥ 10 product-specific codes
- [ ] Each entry: max retries, backoff strategy, agent action, UX feedback
- [ ] HALT vs retry clearly distinguished; credential, quota, business errors separated
- [ ] Error messages follow `[ERROR] code: summary → explanation → fix → next step` format

#### 5. API & CLI Fidelity
- [ ] All operationIds, fields, JSON paths traceable to OpenAPI or verified CLI output
- [ ] CLI default output is JSON; commands match official docs
- [ ] Version pin: SDK/API baseline documented in metadata or integration.md

#### 6. UX Compliance (per user-experience-spec.md)
- [ ] Quick Start section; first command executable within 60s
- [ ] Common operations require ≤ 3 prompts; smart defaults documented
- [ ] Success/failure messages follow standardized format; progress for ops > 5s

#### 7. Self-Healing (per enhanced-self-healing-framework.md)
- [ ] All installation flows: pre-flight checks, error classification, multi-path recovery, health verification, graceful degradation
- [ ] CLI install, Go runtime JIT, dependency download each have ≥ 3 self-healing paths per error type
- [ ] Health score ≥ 8/10, self-healing duration < 30s, user intervention rate < 20% documented

#### 8. Prompt & Description Quality
- [ ] `description` field: imperative phrasing, user-intent focused, implicit triggers, negative boundaries
- [ ] `assets/eval_queries.json` created with ≥ 20 should/should-not trigger queries
- [ ] Generation used structured prompts from `prompt-library.md`

#### 9. AIOps Compliance (when monitoring/alarm/diagnosis in scope)
- [ ] Multi-metric correlation (≥ 4 anomaly patterns), cross-skill diagnosis decision tree
- [ ] Delegation matrix, proactive inspection, alarm storm handling per `references/advanced/aiops-best-practices.md`

#### 10. **IMPORTANT: Mandatory Sections**
- [ ] `### What This Skill Does` — **MUST** exist with a clear 2-3 sentence description of the skill's purpose
- [ ] `## Operational Best Practices` — **MUST** exist with actionable operational guidance (monitoring, backup, security patterns)
- [ ] `### Next Steps` — **MUST** exist (even if brief)

### P1 — SHOULD PASS

- [ ] **Chaining:** Stable output fields for downstream skills (via `{{output.*}}` placeholders)
- [ ] **Naming:** `ve-[product]-ops` consistent with repo conventions
- [ ] **Idempotency** or duplicate-resource behavior documented
- [ ] **Adversarial scenarios** considered using the governance doc
- [ ] **Path preference:** SKILL.md states when to prefer `ve` vs SDK fallback if non-obvious

---

## Example Request

> Add a Volcengine skill for ECS in this repo: instances, disks, snapshots. Docs: `https://www.volcengine.com/docs/6396`. Go SDK (JIT fallback). CLI: `ve ecs`.

**Expected output:** → `ve-ecs-ops` tree with **real** operationIds, Go SDK types, response paths, **and** matching `ve` commands (primary path), plus JIT Go SDK fallback documentation.

---

## Quality Gate (GCL)

> This chapter implements the Generator-Critic-Loop defined in `../../AGENTS.md`
> §3-§9. For this **meta-skill**, GCL is **optional** (`max_iter=3`) because the
> skill's output is generated skill files, not cloud operations. The GCL loop
> verifies that **generated skill artifacts** meet the Five Core Standards and
> inherit GCL correctly. The Generator (the agent executing generation) and the
> Critic (an isolated prompt scoring the output) MUST live in separate contexts
> — see `../../AGENTS.md` §9 anti-patterns.

### 4-Tier Operation Classification

Every operation this meta-skill exposes MUST be classified into exactly one of
the four tiers below. The classification drives `max_iter` and the Safety
floor in the GCL loop.

| Tier | Typical operations | `max_iter` | Safety floor |
|---|---|---|---|
| **Destructive** | — (generation is additive) | — | — |
| **State-changing** | `GenerateSkill`, `UpdateSkill` | 3 | 1.0 (mandatory) |
| **Mutating** | — (all generation is state-changing) | — | — |
| **Read-only** | `AnalyzeOpenAPI`, `ReviewSkill` | 3 | ≥ 0 |

### Loop contract (Generator → Critic → Orchestrator)

1. **Pre-flight (Orchestrator)** — resolve `{{user.*}}` inputs; classify the
   operation into one of the four tiers above; load `references/rubric.md` (GCL
   rubric reference for generated skills) and `references/prompt-templates.md`
   (GCL prompt skeleton reference for generated skills).
2. **Generate** — execute the generation workflow per the
   `## Evaluation-Driven Generation Workflow` chapter. Capture the generated
   file tree, template-versus-output diff, unresolved placeholder report, and
   P0/P1 checklist results into
   `./audit-results/gcl-trace-YYYYMMDD-HHMMSS.json` with `redaction_pass: true`.
3. **Critique** — isolated prompt; score the generated output for correctness /
   safety / idempotency / traceability / spec_compliance per the rubric. Emit
   ≤ 3 actionable suggestions. The Critic MUST NOT see the raw user request.
4. **Decide** — first match wins:
   - Safety = 0 on a State-changing op → **ABORT** (no partial return).
   - All dimensions meet threshold → return Generator output.
   - `iter < max_iter` → inject suggestions into Generator and loop.
   - Else → return best-so-far + unresolved rubric items.

### Mandatory deliverables for every generated skill

A generated skill is **incomplete** if it does NOT ship all three GCL
deliverables:

| Deliverable | Path in generated skill | Required content |
|---|---|---|
| **GCL chapter** | `SKILL.md` section `## Quality Gate (GCL)` | 4-tier table; loop contract (4 steps); cross-skill delegation table; trace spec; explicit pointer to the rubric and prompt templates below |
| **Rubric** | `references/rubric.md` | 5-dimension scoring (Correctness / Safety / Idempotency / Traceability / Spec Compliance); product-specific safety rules; product-specific correctness checks; ≥ 10 product-specific error codes mapped to HALT vs retry; **Safety = 0 → ABORT** rule |
| **Prompt templates** | `references/prompt-templates.md` | Generator prompt (with `{{env.*}}` / `{{user.*}}` / `{{output.*}}` placeholders); Critic prompt (Critic MUST NOT see the raw user request); Orchestrator prompt; ≥ 1 verbatim safety prompt per Destructive / State-changing op |

### Meta-skill-specific safety rules

These rules apply to the **generation process itself** — they verify the
generated skill is safe to ship:

- [ ] Generated skill MUST NOT contain real credentials — only `<masked>` and
      `{{env.*}}` placeholders (zero credential literals).
- [ ] Generated rubric (`references/rubric.md`) MUST include all 5 scoring
      dimensions (Correctness / Safety / Idempotency / Traceability / Spec
      Compliance).
- [ ] Generated prompt templates (`references/prompt-templates.md`) MUST include
      a Critic prompt that hides the raw user request (Critic MUST NOT see it).
- [ ] Generated skill MUST reference `references/rubric.md` and
      `references/prompt-templates.md` in its `## Reference Directory`.
- [ ] Generated skill MUST have valid YAML frontmatter with `name` matching
      `ve-[product]-ops` pattern and `cli_applicability` set correctly.

### Trace (mandatory for every GCL run)

Every GCL run MUST persist a JSON trace to
`./audit-results/gcl-trace-YYYYMMDD-HHMMSS.json` with these fields:

- `skill`, `request` (sanitized), `rubric_version`
- `iterations[]` — each with `iter`, `generator` (generated file tree /
  placeholder report / checklist results), `critic` (scores / suggestions /
  blocking), `decision`
- `final` — `status` (`PASS` / `MAX_ITER` / `SAFETY_FAIL`), `iter`, `output`
- `redaction_pass: true` — the trace MUST NOT contain real credential values;
  only `<masked>` or `sha256:<first-8-hex>` and length

`audit-results/` is in the repo's `.gitignore` (added 2026-06-04).

### Cross-skill delegation (extends `## Role Boundary` above)

When the Critic surfaces a cross-product gap, the **Generator** (not the
Critic) delegates on the next iteration. The Critic only emits suggestions.

| Critic finding | Delegate to |
|---|---|
| Generated skill references a product that already has an ops skill | `ve-[product]-ops` — delegate for verification of accuracy |
| Generated skill needs billing/IAM operation verification | `ve-billing-ops` / `ve-iam-ops` (when present) |
| Generated skill includes monitoring/alarm patterns | `ve-cms-ops` — delegate for monitoring accuracy check |
| OpenAPI spec analysis reveals missing operations | `AnalyzeOpenAPI` read-only pass — re-analyze before next iteration |

### Anti-patterns (from `../../AGENTS.md` §9 — banned)

Generated skills MUST NOT introduce any of these:
- Shared context for G and C (defeats independence).
- Subjective scoring without the rubric.
- Unbounded iteration loops.
- Critic seeing the raw user request (encourages rubber-stamping).
- Silently downgrading on Safety = 0.
- Trace not persisted.
- Critic mutating resources.
- Real credential values in trace.
- GCL bypass for "obviously safe" ops.

---

## Reference Directory

| File | Purpose |
|------|---------|
| [ve-skill-template.md](references/ve-skill-template.md) | Base template for generated SKILL.md |
| [execution-environment.md](references/execution-environment.md) | CLI install, Go JIT download, credential config, verification **(progressive disclosure)** |
| [cli-behavior.md](references/cli-behavior.md) | Verified `ve` CLI behavioral notes (output format, env vars, patterns) **(progressive disclosure)** |
| [enhanced-self-healing-framework.md](references/enhanced-self-healing-framework.md) | **MANDATORY** self-healing patterns for installation flows |
| [governance-and-adversarial-review.md](references/governance-and-adversarial-review.md) | (when present) Adversarial review scenarios and governance checklist |
| [prompt-library.md](references/prompt-library.md) | Structured prompts for the generation lifecycle |
| [optimization-analysis.md](references/optimization-analysis.md) | Three-dimensional optimization framework |
| [user-experience-spec.md](references/user-experience-spec.md) | Mandatory UX requirements for all generated skills |
| [aiops-best-practices.md](references/advanced/aiops-best-practices.md) | Mandatory AIOps patterns for monitoring/diagnosis skills |
| [generation-guidance.md](references/generation-guidance.md) | Description optimization, eval queries, generation telemetry, version pinning |
| [rubric.md](references/rubric.md) | GCL rubric — scoring dimensions for generated skills |
| [prompt-templates.md](references/prompt-templates.md) | GCL prompt skeletons — Generator/Critic/Orchestrator templates |
| [assets/eval_queries.json](assets/eval_queries.json) | Eval queries for testing the meta-skill's description trigger accuracy |

### External References

- [Volcengine CLI](https://github.com/volcengine/volcengine-cli)
- [Volcengine SDK for Go](https://github.com/volcengine/volc-sdk-golang)
- [Agent Skills Open Specification](https://agentskills.io/specification)
- [Volcengine Documentation](https://www.volcengine.com/docs)

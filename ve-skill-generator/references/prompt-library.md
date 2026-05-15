# Prompt Library — Volcengine Skill Generator

> **Purpose:** Structured prompts for the generation lifecycle. Use these templates during skill generation to ensure consistent output quality.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-15

---

## Table of Contents

1. [Generation Prompts](#1-generation-prompts)
2. [Research Prompts](#2-research-prompts)
3. [Review Prompts](#3-review-prompts)
4. [Optimization Prompts](#4-optimization-prompts)
5. [Prompt Effectiveness Tracking](#5-prompt-effectiveness-tracking)

---

## 1. Generation Prompts

### 1.1 Initial Skill Generation

```
Generate a Volcengine operational skill for [Product Name] ([Product Chinese Name]).

Product Details:
- Product slug: [slug, e.g., ecs, rds_mysql]
- Primary resource: [e.g., Instance, DBInstance]
- API documentation: [URL]
- OpenAPI spec: [URL if available]
- Core operations: create, describe, modify, delete, list, [product-specific actions]

Requirements:
1. Use the ve-skill-template.md as the base template
2. Populate all placeholders with product-specific, verified data
3. Include dual-path execution (ve CLI + JIT Go SDK fallback)
4. Document all CLI commands with verified parameter names
5. Create error taxonomy with ≥ 10 product-specific error codes
6. Add safety gates for all destructive operations

Output structure:
ve-[product-slug]-ops/
├── SKILL.md
├── references/
│   ├── core-concepts.md
│   ├── api-sdk-usage.md
│   ├── cli-usage.md
│   ├── troubleshooting.md
│   └── integration.md
└── assets/
    └── example-config.yaml
```

### 1.2 SKILL.md Population

```
Populate the SKILL.md for ve-[product]-ops using the verified data below:

Product: [name]
Slug: [slug]
Primary Resource: [resource]
Operations: [list]
Error Codes: [list with descriptions]
CLI Commands: [list with parameter names]
SDK Package: github.com/volcengine/volc-sdk-golang/service/[product]

For each operation:
- Document Pre-flight → Execute (CLI + SDK) → Validate → Recover
- Include exact CLI command syntax with verified parameters
- Include SDK code template with correct imports
- Map response JSON paths to {{output.*}} placeholders
- Add failure recovery entries for each error pattern
```

---

## 2. Research Prompts

### 2.1 API Discovery

```
Research the Volcengine [Product] API and extract:

1. All available API operations (operation names)
2. Required parameters for each operation
3. Response schema structure and key JSON paths
4. Product-specific error codes
5. Async/polling behavior (if any)

Source: https://www.volcengine.com/docs/[product-doc-id]
Focus on: [specific area, e.g., instance lifecycle, disk management]
```

### 2.2 CLI Coverage Analysis

```
Determine ve CLI coverage for Volcengine [Product]:

1. Run `ve [product] --help` and list all available actions
2. Compare CLI actions against API operations
3. Identify coverage gaps (API-only operations)
4. Note any CLI-specific behaviors or parameter differences

Document findings in cli-usage.md format with:
- Command map (goal → CLI invocation)
- Coverage gap table (API vs CLI)
- Verified JSON output paths
```

### 2.3 Go SDK Mapping

```
Map the Volcengine Go SDK for [Product]:

1. SDK package path: github.com/volcengine/volc-sdk-golang/service/[product]
2. List available client methods/functions
3. Identify initialization pattern (NewInstance, SetAccessKey, etc.)
4. Note request parameter structure
5. Document response parsing pattern

Create a quick reference table:
| Operation | SDK Method | Request Type | Response Type |
```

---

## 3. Review Prompts

### 3.1 P0 Checklist Review

```
Review the generated ve-[product]-ops skill against P0 requirements:

1. Trigger & Scope: Are SHOULD/SHOULD NOT use conditions clear?
2. Variables: Are all {{env.*}} placeholders using correct VOLCENGINE_* names?
3. Flows: Does every operation have Pre-flight → Execute → Validate → Recover?
4. CLI fidelity: Do ve commands match official CLI docs?
5. Safety gates: Are destructive operations gated with explicit confirmation?
6. Error taxonomy: Are there ≥ 10 product-specific error codes?
7. Credential masking: Is VOLCENGINE_SECRET_KEY never echoed?

For each failing item: document the gap and specific fix needed.
```

### 3.2 Anti-Pattern Review

```
Scan the generated skill for these anti-patterns:

1. API Hallucination: Any invented field names or JSON paths?
2. Hardcoded Values: Any hardcoded regions, timeouts, or limits?
3. Missing Failure Path: Are all operations missing error recovery?
4. Redundant Redundancy: Is the same content duplicated in SKILL.md and references?
5. Credential Leak: Any secret value in example output or logs?

For each finding: location, anti-pattern type, recommended fix.
```

---

## 4. Optimization Prompts

### 4.1 Description Optimization

```
Optimize the skill description for trigger accuracy:

Current description: [paste current description]

Requirements:
- Under 1024 characters
- Third person, imperative phrasing
- Include implicit trigger scenarios
- Add negative boundaries (what this skill is NOT for)
- Mention product Chinese name for trigger matching

Generate 3 variations and score each against these test queries:
[Provide 10 test queries: 5 should-trigger, 5 should-not-trigger]
```

### 4.2 Eval Query Generation

```
Generate eval queries for testing ve-[product]-ops skill trigger accuracy:

Product: [name]
Chinese name: [name]
Related products to exclude: [list]

Generate 20 queries:
- 10 that SHOULD trigger this skill (vary phrasing, explicitness, use Chinese names)
- 10 that should NOT trigger (near-misses with related products, billing, IAM)

Include realistic variations: typos, casual language, abbreviations, context-heavy.
```

---

## 5. Prompt Effectiveness Tracking

Track prompt effectiveness using this template:

| Prompt ID | Purpose | Success Rate | Avg Iterations | Notes |
|-----------|---------|-------------|----------------|-------|
| GEN-001 | Initial skill generation | ~70% | 2-3 | Usually needs 1-2 fix passes |
| RES-001 | API discovery | ~85% | 1-2 | Depends on doc quality |
| REV-001 | P0 checklist review | ~95% | 1 | Reliable for catching gaps |
| OPT-001 | Description optimization | ~60% | 3-4 | Often needs human refinement |

Update this table monthly based on generation outcomes.

---

*Use these prompts as starting points. Adapt based on product complexity and doc quality.*

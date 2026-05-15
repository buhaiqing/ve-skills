# User Experience Specification — Volcengine Skill Generator

> **Purpose:** Defines user experience (UX) requirements and design patterns that MUST be integrated into every generated `ve-[product]-ops` skill. This specification ensures generated skills are intuitive, accessible, and confidence-inspiring for operators.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-15
> **Status:** MANDATORY — all generated skills MUST pass UX review against this spec

---

## Table of Contents

1. [UX Design Principles](#1-ux-design-principles)
2. [Onboarding & Guidance](#2-onboarding--guidance)
3. [Interaction Design](#3-interaction-design)
4. [Feedback Mechanisms](#4-feedback-mechanisms)
5. [Error Handling & Recovery](#5-error-handling--recovery)
6. [UX Review & Validation](#6-ux-review--validation)
7. [Appendix: UX Patterns Library](#7-appendix-ux-patterns-library)

---

## 1. UX Design Principles

### 1.1 Core Principles

All generated skills MUST adhere to these five core principles:

| Principle | Description | Success Criteria |
|-----------|-------------|------------------|
| **Clarity** | Every action and its consequence is unambiguous | User never wonders "what just happened?" |
| **Efficiency** | Common tasks require minimal steps | 80% of tasks complete in ≤ 3 prompts |
| **Forgiveness** | Mistakes are recoverable with clear guidance | User can undo or recover from any non-destructive error |
| **Consistency** | Patterns are uniform across all ve skills | User learns once, applies everywhere |
| **Transparency** | System state is always visible | User always knows what the system is doing |

### 1.2 UX Maturity Model for Generated Skills

| Level | Name | Characteristics |
|-------|------|-----------------|
| 1 | Functional | Skill works; minimal UX consideration |
| 2 | Usable | Basic guidance; clear error messages |
| 3 | Comfortable | Onboarding flow; consistent patterns; helpful defaults |
| 4 | Delightful | Anticipates needs; proactive suggestions; minimal friction |
| 5 | Intuitive | Feels like natural conversation; zero learning curve |

**Target:** All generated skills MUST achieve **Level 3 (Comfortable)** minimum.

---

## 2. Onboarding & Guidance

### 2.1 Quick Start Section (Mandatory)

Every `SKILL.md` MUST include a **Quick Start** section immediately after the Overview. This section must enable a first-time user to execute their first command within 60 seconds.

**Required Structure:**

```markdown
## Quick Start

### What This Skill Does
[One sentence describing the skill's primary purpose]

### Prerequisites
- [ ] `ve` CLI installed (or Go runtime for JIT fallback)
- [ ] Credentials configured: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`
- [ ] Region set: `VOLCENGINE_REGION`

### Verify Setup
```bash
ve {{product}} DescribeInstances --Region {{env.VOLCENGINE_REGION}}
```

### Your First Command
```bash
ve {{product}} Describe{{Resources}} --Region {{env.VOLCENGINE_REGION}}
```
```

### 2.2 Capability Overview

After Quick Start, provide a **Capability Overview** table showing operations, complexity, and risk levels.

### 2.3 Progressive Disclosure

Information MUST be presented progressively:
- **Level 1 (Summary):** One-line description + command example
- **Level 2 (Details):** Parameter table + common options
- **Level 3 (Advanced):** Full parameter reference + edge cases

---

## 3. Interaction Design

### 3.1 Prompt Minimization

**Rule:** Ask the user for information ONLY when it cannot be inferred from environment variables, defaulted safely, or derived from previous context.

**Prompt Budget per Operation:**

| Operation Type | Max Prompts | Notes |
|----------------|-------------|-------|
| Describe / List | 0–1 | Should be fully automated with env vars |
| Create | 1–2 | Name + region (region can default from env) |
| Modify | 1–2 | Resource ID + change specification |
| Delete | 1 + confirmation | Resource ID + explicit confirmation |

### 3.2 Smart Defaults

Every optional parameter SHOULD have a smart default:

| Parameter | Smart Default | Rationale |
|-----------|---------------|-----------|
| Region | `{{env.VOLCENGINE_REGION}}` | User's configured region |
| Name | `[product]-[resource]-[timestamp]` | Unique, descriptive, sortable |
| Timeout | 300s | Balances patience and responsiveness |
| PageSize | 50 | Reasonable batch size for list operations |

### 3.3 Confirmation Patterns

**Destructive Operations MUST use explicit confirmation:**

```markdown
⚠️ **Destructive Action Confirmation**

You are about to DELETE the following resource:
- Name: {{user.resource_name}}
- ID: {{user.resource_id}}
- Region: {{user.region}}

This action is **IRREVERSIBLE**. All data will be permanently lost.

Type the resource name "{{user.resource_name}}" to confirm: _
```

### 3.4 Progress Indication

For operations taking > 5 seconds, MUST show progress with elapsed time and estimated remaining time.

---

## 4. Feedback Mechanisms

### 4.1 Success Feedback

Every successful operation MUST provide resource ID, key fields, and suggested next steps.

### 4.2 Failure Feedback

Every failed operation MUST follow the standardized error format:

```markdown
❌ **Operation Failed**

Error: {{error.code}}
Message: {{error.human_readable}}

**What happened:**
{{error.explanation}}

**How to fix:**
{{error.remediation}}

**Next steps:**
1. {{error.next_action_1}}
2. {{error.next_action_2}}
```

### 4.3 Feedback Timing

| Operation Duration | Feedback Type |
|-------------------|---------------|
| < 1s | Immediate result |
| 1–5s | Result with brief "Done" message |
| 5–30s | Progress indicator + final result |
| > 30s | Detailed progress + ETA + final result |

---

## 5. Error Handling & Recovery

### 5.1 Error Message Design

All error messages MUST follow this format:

```
[ERROR] {error.code}: {human_readable_summary}

What happened:
{2-3 sentence explanation in plain language}

How to fix:
{1-3 concrete steps}

Next step:
{single actionable instruction}
```

### 5.2 Error Categories and Handling

| Category | User-Friendly Prefix | Auto-Recoverable | Action |
|----------|---------------------|-------------------|--------|
| Credential | "Authentication failed" | No | HALT with setup instructions |
| Region | "Region not available" | No | Suggest valid regions |
| Resource Not Found | "Resource not found" | No | Suggest list command |
| Quota | "Quota exceeded" | No | HALT with quota increase link |
| Throttling | "Rate limit reached" | Yes (retry) | Auto-retry with backoff |
| Invalid Parameter | "Invalid input" | Yes (fix) | Prompt for correction |
| Internal Error | "Server error" | Yes (retry) | Retry 3x, then HALT |
| Network | "Connection failed" | Yes (retry) | Retry with exponential backoff |

### 5.3 Escalation Template

When HALT is necessary, provide standardized escalation block with Request ID, operation name, resource ID, error code, and timestamp.

---

## 6. UX Review & Validation

### 6.1 UX Review Checklist

#### Onboarding
- [ ] Quick Start section exists and is ≤ 30 seconds to read
- [ ] Prerequisites are clearly listed with verification commands
- [ ] First command example is copy-paste ready
- [ ] Capability Overview table is present

#### Interaction
- [ ] Common operations require ≤ 3 prompts
- [ ] Smart defaults are documented for all optional parameters
- [ ] Destructive operations have explicit confirmation
- [ ] Progress is shown for operations > 5s

#### Feedback
- [ ] Success messages include resource ID and next steps
- [ ] Failure messages include error code, explanation, and fix steps
- [ ] Long-running operations show progress and ETA
- [ ] All feedback is human-readable (not raw JSON)

#### Error Handling
- [ ] All error categories have user-friendly messages
- [ ] Recovery steps are concrete and actionable
- [ ] Escalation template includes all required fields
- [ ] No secret values are exposed in error messages

### 6.2 UX Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Time to First Success | < 60s | From first prompt to successful operation |
| Task Completion Rate | > 90% | % of users who complete task without escalation |
| Error Recovery Rate | > 80% | % of errors resolved without human support |
| Prompt Count (common tasks) | ≤ 3 | Average prompts per common operation |

---

## 7. Appendix: UX Patterns Library

### Pattern: Guided Wizard

For complex multi-step operations, provide step-by-step guided flow with review before execution.

### Pattern: Batch Operation

For operating on multiple resources, show resource list, skip already-correct items, confirm count, and show progress.

### Pattern: Dry Run

For validating operations without executing, show what WOULD happen and provide a confirm flag to actually execute.

---

*This specification is mandatory for all skills generated by `ve-skill-generator`. Update it as UX best practices evolve.*

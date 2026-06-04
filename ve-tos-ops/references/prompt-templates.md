---
name: ve-tos-ops-prompt-templates
description: >-
  GCL prompt templates for ve-tos-ops. Generator / Critic / Orchestrator roles
  plus TOS-specific safety prompts for DeleteBucket, DeleteObject prefix-pattern,
  PutBucketACL (public access), OptimizeStorageClass (Archive cost),
  PutBucketVersioning (suspend).
  NOTE: TOS uses TOS_ACCESS_KEY/TOS_SECRET_KEY (not VOLCENGINE_*). 
  All placeholders: {{env.*}} / {{user.*}} / {{output.*}}.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-tos-ops
  gcl_role: prompt_skeletons
  roles: [generator, critic, orchestrator]
  default_max_iter: 2
  env_var_notice: "TOS_ACCESS_KEY / TOS_SECRET_KEY (not VOLCENGINE_*)"
---

# GCL Prompt Templates — ve-tos-ops

---

## 1. Generator Prompt (role: G)

```text
You are the Generator for the Volcengine TOS skill (ve-tos-ops).
You execute TOS operations using `tosutil` (bulk data), `ve tos` (API),
or JIT Go SDK (fallback). You MUST NOT self-score or modify the rubric.

# Inputs
- user_request: {{user.request}}
- critic_feedback_from_previous_iter: {{output.critic_feedback}}
- rubric: {{output.rubric}}
- operation_tier: {{output.operation_tier}}

# Execution contract
1. Resolve placeholders. {{env.TOS_ACCESS_KEY}} / {{env.TOS_SECRET_KEY}} /
   {{env.VOLCENGINE_REGION}} from runtime — NEVER ask user.
2. Pre-flight: walk the skill's pre-flight table. HALT on failure.
3. Execute: prefer `tosutil` for bulk data, `ve tos <Action>` for API.
   Dual-path: ALSO write Go SDK snippet.
4. Validate: ls / stat / Get* API to confirm.
5. Trace: persist to ./audit-results/gcl-trace-*.json with redaction_pass: true.

# Safety gate
- Surface verbatim confirmation prompt from references/prompt-templates.md §4.
- For DeleteBucket: MUST check bucket emptiness (objects + versions + multiparts).
- For DeleteObject -r: MUST show file list before executing.
- For PutBucketACL public-read: warn about public access.
- For OptimizeStorageClass → Archive: warn about retrieval costs.
- NEVER print TOS_SECRET_KEY (or VOLCENGINE_SECRET_KEY).

# Output JSON
{ "status": "OK" | "HALT", "operation": "...", "bucket": "...",
  "command": "...", "request_id": "...", "validation": { ... },
  "trace_path": "...", "issues_encountered": [...], "next_action": "..." }
```

---

## 2. Critic Prompt (role: C)

> The Critic MUST NOT see the raw user request.

```text
You are a Volcengine TOS auditor. Score the execution result STRICTLY
against the rubric. Do NOT consider the original user request.

# Inputs
- rubric: {{output.rubric}}
- generator_output: {{output.generator_output}}
- trace: {{output.trace}}

# Scoring (0|0.5|1)
- correctness      : bucket/object state matches request
- safety           : destructive/state-changing confirmed (0/1; 0 → ABORT)
- idempotency      : retry safe; pre-checks before CreateBucket/PutObject
- traceability     : command, params, response, validation captured
- spec_compliance  : dual-path (ve tos + tosutil); ≥ 10 TOS error codes

# TOS-specific checks
- Verify TOS_SECRET_KEY NOT in trace (not just VOLCENGINE_SECRET_KEY).
- Verify DeleteBucket: emptiness check + user confirmation in trace.
- Verify DeleteObject -r: file list shown to user.
- Verify PutBucketACL public-read: warning present.
- Verify PutBucketLifecycle with Expiration: warning about permanent deletion.
- Verify OptimizeStorageClass → Archive: retrieval cost warning present.

# Output (strict JSON)
{ "scores": { "correctness": 0|0.5|1, "safety": 0|0.5|1,
  "idempotency": 0|0.5|1, "traceability": 0|0.5|1,
  "spec_compliance": 0|0.5|1 },
  "suggestions": ["≤ 3 concrete improvements"],
  "blocking": true|false }
```

---

## 3. Orchestrator Prompt (role: O)

```text
You are the Orchestrator for ve-tos-ops. Control iteration/termination.
Do NOT call `ve tos`, `tosutil`, or Go SDK.

# Decision (first match)
1. Safety=0 AND op in {destructive, state_changing} → ABORT.
2. All pass → return Generator output.
3. iter < max_iter → inject suggestions and loop.
4. Else → return best-so-far + unresolved.

# Output
{ "status": "PASS" | "MAX_ITER" | "SAFETY_FAIL",
  "iter": <int>, "final_output": <generator output>,
  "unresolved_rubric_items": [...] }
```

---

## 4. Operation-Specific Safety Prompts

### 4.1 `DeleteBucket`

```text
[DESTRUCTIVE] About to delete TOS bucket {{user.bucket}} (region: cn-beijing).

Pre-deletion checks:
  - Objects found: {{output.object_count}}
  - Versioning enabled: {{output.versioning_enabled}}
  - Non-current versions: {{output.noncurrent_version_count}}
  - In-progress multipart uploads: {{output.multipart_count}}

WARNING: This is IRREVERSIBLE. The bucket and ALL its contents will be
permanently deleted. If versioning is enabled, all versions and delete
markers will be removed. In-progress multipart uploads will be aborted.

Type "yes" to delete, or "cancel" to abort.
```

### 4.2 `DeleteObject` (prefix-pattern variant)

```text
[DESTRUCTIVE] About to delete objects with prefix
"{{user.object_prefix}}" from bucket {{user.bucket}}.

The following objects will be deleted (showing first 20):
{{output.object_list}}

Total: {{output.total_count}} objects
This is IRREVERSIBLE. Type "yes" to delete all, or "cancel" to abort.
```

### 4.3 `PutBucketACL` (public-access variant)

```text
[PERMISSION-CHANGE] About to set ACL on bucket {{user.bucket}} to
{{user.acl}}.

{{#if is_public}}
WARNING: {{user.acl}} allows ANYONE on the internet to
{{#if is_public_read_write}}read and write{{else}}read{{/if}}
objects in this bucket. This is a significant security risk.
{{/if}}

Type "yes" to apply, or "cancel" to abort.
```

### 4.4 `OptimizeStorageClass` (Archive variant)

```text
[COST-IMPACT] About to transition objects in bucket {{user.bucket}}
with prefix "{{user.prefix}}" to {{user.target_storage_class}}.

WARNING: Archive storage class has retrieval costs:
  - Per-GB retrieval fee
  - 1-12 hours restore time before objects are readable
  - Objects cannot be deleted for 180 days minimum

Type "yes" to proceed, or "cancel" to abort.
```

### 4.5 `PutBucketVersioning` (suspend variant)

```text
[STATE-CHANGING] About to SUSPEND versioning on bucket {{user.bucket}}.

Current state: {{output.current_versioning}}.
Target state: Suspended.

WARNING: With versioning suspended, new overwrites will permanently
replace (not version) existing objects. You will lose the ability to
recover previous versions from new uploads. Existing versions are preserved.

Type "yes" to suspend, or "cancel" to abort.
```

### 4.6 `PutBucketLifecycle` (expiration variant)

```text
[STATE-CHANGING] About to set lifecycle rule on bucket {{user.bucket}}:

Rule: {{output.lifecycle_rule_summary}}

{{#if has_expiration}}
WARNING: Objects matching prefix "{{user.prefix}}" will be PERMANENTLY
DELETED after {{output.expiration_days}} days. This is IRREVERSIBLE.
{{/if}}

Type "yes" to apply, or "cancel" to abort.
```

---

## 5. Changelog

| Version | Date | Change |
|---|---|---|
| 1.0.0 | 2026-06-04 | Initial GCL prompt templates for ve-tos-ops (6 operation-specific safety prompts; DeleteBucket emptiness/versioning guard; DeleteObject prefix-pattern review; PutBucketACL public-warning; Archive-cost warning; TOS_* env-var convention) |
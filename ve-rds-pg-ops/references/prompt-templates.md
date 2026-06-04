---
name: ve-rds-pg-ops-prompt-templates
description: GCL prompt templates for ve-rds-pg-ops. G/C/O + safety prompts for DeleteDBInstance, ModifyDBInstanceSpec, DeleteDBAccount, ModifyDBInstanceParameter.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-rds-pg-ops
  gcl_role: prompt_skeletons
  roles: [generator, critic, orchestrator]
  default_max_iter: 2
---

# GCL Prompt Templates — ve-rds-pg-ops

## 1. Generator Prompt

```text
Generator for ve-rds-pg-ops. Execute via `ve rds_postgresql <Action>` or Go SDK.
# Inputs: {{user.request}}, {{output.critic_feedback}}, {{output.rubric}}
- NEVER ask {{env.*}} from user.
- Dual-path: document both CLI and Go SDK.
- Trace to ./audit-results/gcl-trace-*.json with redaction_pass: true.
- Surface safety prompts from §4 for destructive/state-changing ops.
- Password masked as <masked> in trace. NEVER print VOLCENGINE_SECRET_KEY.

Output JSON: { "status", "operation", "instance_id", "command", "request_id", "validation", "trace_path", "issues_encountered", "next_action" }
```

## 2. Critic Prompt (MUST NOT see raw user request)

```text
Independent PostgreSQL RDS auditor. Score against rubric. Do NOT consider user request.
# Inputs: {{output.rubric}}, {{output.generator_output}}, {{output.trace}}
- Verify VOLCENGINE_SECRET_KEY not in trace; DB password masked.
- Verify DeleteDBInstance: deletion protection check in trace.
- Verify ModifyDBInstanceSpec: downtime warning present.
- Verify `ve rds_postgresql <Action> --<Param>` PascalCase shape.
Output: { "scores": {correctness, safety, idempotency, traceability, spec_compliance}, "suggestions": [...], "blocking": bool }
```

## 3. Orchestrator Prompt

```text
Orchestrator for ve-rds-pg-ops. Do NOT call CLI/SDK.
Safety=0 AND destructive/state_changing → ABORT. All pass → return. iter<max_iter → loop. Else → best+unresolved.
Output: { "status", "iter", "final_output", "unresolved_rubric_items" }
```

## 4. Safety Prompts

### 4.1 DeleteDBInstance
```text
[DESTRUCTIVE] Delete PG instance {{user.instance_name}} ({{user.instance_id}}).
IRREVERSIBLE. All data lost. Type "yes" to delete, or "cancel".
```

### 4.2 ModifyDBInstanceSpec
```text
[STATE-CHANGING] Modify PG instance {{user.instance_name}} spec. 60-900s downtime.
Type "yes" to proceed, or "cancel".
```

### 4.3 DeleteDBAccount
```text
[DESTRUCTIVE] Delete account {{user.account_name}} from {{user.instance_name}}.
Apps lose DB access. Type "yes" to delete, or "cancel".
```

### 4.4 ModifyDBInstanceParameter (ForceRestart)
```text
[STATE-CHANGING] Modify params on {{user.instance_name}}. ForceRestart=true: instance restarts.
Type "yes" to apply, or "cancel".
```

## 5. Changelog
| Version | Date | Change |
|---|---|---|
| 1.0.0 | 2026-06-04 | Initial GCL prompt templates for ve-rds-pg-ops |
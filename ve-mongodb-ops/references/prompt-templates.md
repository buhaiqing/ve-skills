---
name: ve-mongodb-ops-prompt-templates
description: GCL prompt templates for ve-mongodb-ops. G/C/O + safety prompts for DeleteDBInstance, ModifyDBInstanceSpec, RestartDBInstance.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-mongodb-ops
  gcl_role: prompt_skeletons
  roles: [generator, critic, orchestrator]
  default_max_iter: 2
---

# GCL Prompt Templates — ve-mongodb-ops

## 1. Generator Prompt

```text
Generator for ve-mongodb-ops. Execute via `ve mongodb <Action>` or Go SDK.
- NEVER ask {{env.*}}. Dual-path. Trace to gcl-trace-*.json.
- Safety prompts from §4. Password masked. NEVER print VOLCENGINE_SECRET_KEY.
Output JSON: { "status", "operation", "instance_id", "command", "request_id", "validation", "trace_path" }
```

## 2. Critic Prompt (MUST NOT see raw user request)

```text
Independent MongoDB auditor. Score against rubric.
- Verify VOLCENGINE_SECRET_KEY not in trace. DB password masked.
- Verify DeleteDBInstance: data-loss warning. Replica set election warning for spec change.
- Verify `ve mongodb <Action>` PascalCase shape.
Output: { "scores": {...}, "suggestions": [...], "blocking": bool }
```

## 3. Orchestrator Prompt

```text
Orchestrator. Safety=0 AND destructive/state-changing → ABORT. All pass → return. iter<max_iter → loop.
```

## 4. Safety Prompts

### 4.1 DeleteDBInstance
```text
[DESTRUCTIVE] Delete MongoDB instance {{user.instance_name}} ({{user.instance_id}}).
ALL data + backups + configs LOST. IRREVERSIBLE. Type "yes", or "cancel".
```

### 4.2 ModifyDBInstanceSpec
```text
[STATE-CHANGING] Modify MongoDB {{user.instance_name}} spec. 60-900s downtime (replica set election).
Type "yes", or "cancel".
```

### 4.3 RestartDBInstance
```text
[STATE-CHANGING] Restart MongoDB {{user.instance_name}}. Brief connection interruption.
Type "yes", or "cancel".
```

## 5. Changelog
| Version | Date | Change |
|---|---|---|
| 1.0.0 | 2026-06-04 | Initial GCL prompt templates for ve-mongodb-ops |
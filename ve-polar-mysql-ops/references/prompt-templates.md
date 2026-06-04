---
name: ve-polar-mysql-ops-prompt-templates
description: GCL prompt templates for ve-polar-mysql-ops. G/C/O + safety prompts for DeleteDBCluster, Failover, ModifyDBNodeSpec, ScaleStorage.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-polar-mysql-ops
  gcl_role: prompt_skeletons
  roles: [generator, critic, orchestrator]
  default_max_iter: 2
---

# GCL Prompt Templates — ve-polar-mysql-ops

## 1. Generator Prompt

```text
Generator for ve-polar-mysql-ops. Execute via `ve polardb_mysql <Action>` or Go SDK.
- NEVER ask {{env.*}}. Dual-path: CLI + Go SDK. Trace to gcl-trace-*.json.
- Surface safety prompts from §4. Cluster password masked in trace.
- NEVER print VOLCENGINE_SECRET_KEY.

Output JSON: { "status", "operation", "cluster_id", "command", "request_id", "validation", "trace_path", "issues", "next_action" }
```

## 2. Critic Prompt (MUST NOT see raw user request)

```text
Independent PolarDB MySQL auditor. Score against rubric.
- Verify VOLCENGINE_SECRET_KEY not in trace. DB password masked.
- Verify DeleteDBCluster: data-loss warning present.
- Verify Failover: production impact warning present.
- Verify `ve polardb_mysql <Action>` PascalCase shape.
Output: { "scores": {correctness, safety, idempotency, traceability, spec_compliance}, "suggestions": [...], "blocking": bool }
```

## 3. Orchestrator Prompt

```text
Orchestrator. Safety=0 AND destructive/state-changing → ABORT. All pass → return. iter<max_iter → loop.
```

## 4. Safety Prompts

### 4.1 DeleteDBCluster
```text
[DESTRUCTIVE] Delete PolarDB cluster {{user.cluster_name}} ({{user.cluster_id}}).
ALL compute nodes + shared storage + data LOST. IRREVERSIBLE. Type "yes", or "cancel".
```

### 4.2 Failover
```text
[STATE-CHANGING] Failover PolarDB cluster {{user.cluster_name}}.
Primary→Secondary role swap. ~30-60s write interruption. Type "yes", or "cancel".
```

### 4.3 ModifyDBNodeSpec
```text
[STATE-CHANGING] Scale compute nodes of {{user.cluster_name}}. 60-900s downtime.
Type "yes", or "cancel".
```

### 4.4 ScaleStorage
```text
[STATE-CHANGING] Scale storage of {{user.cluster_name}} to {{user.storage_size}} GB.
Storage scaling is IRREVERSIBLE (can only increase). Type "yes", or "cancel".
```

## 5. Changelog
| Version | Date | Change |
|---|---|---|
| 1.0.0 | 2026-06-04 | Initial GCL prompt templates for ve-polar-mysql-ops |
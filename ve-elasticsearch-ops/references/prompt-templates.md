---
name: ve-elasticsearch-ops-prompt-templates
description: GCL prompt templates for ve-elasticsearch-ops. G/C/O + safety prompts for DeleteInstance, DeleteIndex, ModifyInstanceSpec, RestartInstance, UpgradeVersion, UninstallPlugin.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-elasticsearch-ops
  gcl_role: prompt_skeletons
  roles: [generator, critic, orchestrator]
  default_max_iter: 2
---

# GCL Prompt Templates — ve-elasticsearch-ops

## 1. Generator Prompt

```text
Generator for ve-elasticsearch-ops. Execute via `ve elasticsearch <Action>` or Go SDK.
- NEVER ask {{env.*}}. Dual-path. Trace to gcl-trace-*.json.
- Surface safety prompts from §4. Kibana password masked. NEVER print VOLCENGINE_SECRET_KEY.
Output JSON: { "status", "operation", "instance_id", "command", "request_id", "validation", "trace_path" }
```

## 2. Critic Prompt (MUST NOT see raw user request)

```text
Independent ES auditor. Score against rubric.
- Verify VOLCENGINE_SECRET_KEY not in trace. Kibana password masked.
- Verify DeleteInstance: ALL indices+snapshots lost warning.
- Verify ModifyInstanceSpec: cluster rebalance downtime warning.
- Verify `ve elasticsearch <Action>` PascalCase shape.
Output: { "scores": {...}, "suggestions": [...], "blocking": bool }
```

## 3. Orchestrator Prompt

```text
Orchestrator. Safety=0 AND destructive/state-changing → ABORT. All pass → return. iter<max_iter → loop.
```

## 4. Safety Prompts

### 4.1 DeleteInstance
```text
[DESTRUCTIVE] Delete ES instance {{user.instance_name}} ({{user.instance_id}}).
ALL indices, data, snapshots, Kibana config LOST. IRREVERSIBLE. Type "yes", or "cancel".
```

### 4.2 DeleteIndex
```text
[DESTRUCTIVE] Delete index {{user.index_name}} from {{user.instance_id}}.
IRREVERSIBLE. Type "yes", or "cancel".
```

### 4.3 ModifyInstanceSpec
```text
[STATE-CHANGING] Scale ES {{user.instance_name}}. 60-1800s cluster rebalancing downtime.
Type "yes", or "cancel".
```

### 4.4 RestartInstance
```text
[STATE-CHANGING] Restart ES {{user.instance_name}}. Rolling restart: 30-60s per node.
Search/indexing interrupted. Type "yes", or "cancel".
```

### 4.5 UpgradeVersion
```text
[STATE-CHANGING] Upgrade ES {{user.instance_name}} to {{user.target_version}}.
Rolling upgrade: 30-60s per node interruption. Cannot downgrade after upgrade.
Type "yes", or "cancel".
```

### 4.6 UninstallPlugin
```text
[STATE-CHANGING] Uninstall plugin {{user.plugin_name}} from {{user.instance_name}}.
Plugin functionality stops immediately. Type "yes", or "cancel".
```

## 5. Changelog
| Version | Date | Change |
|---|---|---|
| 1.0.0 | 2026-06-04 | Initial GCL prompt templates for ve-elasticsearch-ops |
---
name: ve-cms-ops-prompt-templates
description: GCL prompt templates for ve-cms-ops.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-cms-ops
  gcl_role: prompt_skeletons
  default_max_iter: 3
---
# GCL Prompt Templates — ve-cms-ops
## 1. Generator Prompt
Generator for ve-cms-ops. Execute via `ve cms <Action>` or Go SDK. Dual-path. Trace. NEVER print VOLCENGINE_SECRET_KEY.
## 2. Critic Prompt (MUST NOT see user request)
Independent CMS auditor. Verify DeleteAlarmRule: monitoring blackout for resources. Verify CreateAlarmRule with empty notification: no alert sent.
## 3. Orchestrator Prompt
Safety=0 AND destructive/state-changing → ABORT.
## 4. Safety Prompts
### 4.1 DeleteAlarmRule
[DESTRUCTIVE] Delete alarm rule {{user.rule_id}} ({{user.rule_name}}). Resources {{output.resource_description}} will NOT be monitored for {{output.metric}}. Type "yes", or "cancel".
### 4.2 DisableAlarmRule
[STATE-CHANGING] Disable alarm rule {{user.rule_id}}. Alerts will NOT be sent while disabled. Type "yes", or "cancel".
### 4.3 CreateAlarmRule (no notification variant)
[STATE-CHANGING] Create alarm rule on {{user.resource}} for {{user.metric}}. Notification channels: {{output.channels}}. WARNING: empty channels = no alert sent. Type "yes" to proceed, or "cancel".
## 5. Changelog
| 1.0.0 | 2026-06-04 | Initial GCL prompt templates for ve-cms-ops |
---
name: ve-fg-ops-prompt-templates
description: GCL prompt templates for ve-fg-ops.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-fg-ops
  gcl_role: prompt_skeletons
  default_max_iter: 3
---
# GCL Prompt Templates — ve-fg-ops
## 1. Generator Prompt
Generator for ve-fg-ops. Execute via `ve fg <Action>` or Go SDK. Dual-path. Trace. NEVER print VOLCENGINE_SECRET_KEY.
## 2. Critic Prompt (MUST NOT see user request)
Independent FG auditor. Verify DeleteFunction: ALL versions/triggers lost warning. Verify UpdateFunction: in-flight invocations warning.
## 3. Orchestrator Prompt
Safety=0 AND destructive/state-changing → ABORT.
## 4. Safety Prompts
### 4.1 DeleteFunction
[DESTRUCTIVE] Delete function {{user.fn_name}} ({{user.fn_id}}). ALL versions, aliases, triggers LOST. Type "yes", or "cancel".
### 4.2 DeleteTrigger
[DESTRUCTIVE] Delete trigger from {{user.fn_name}}. Function stops responding to {{user.trigger_type}} events. Type "yes", or "cancel".
### 4.3 UpdateFunction
[STATE-CHANGING] Update function code/config for {{user.fn_name}}. In-flight invocations ({{output.in_flight}}) will complete with old code. New invocations use new code. Type "yes", or "cancel".
## 5. Changelog
| 1.0.0 | 2026-06-04 | Initial GCL prompt templates for ve-fg-ops |
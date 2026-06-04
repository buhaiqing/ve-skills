---
name: ve-alb-ops-prompt-templates
description: GCL prompt templates for ve-alb-ops.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-alb-ops
  gcl_role: prompt_skeletons
  default_max_iter: 3
---
# GCL Prompt Templates — ve-alb-ops
## 1. Generator Prompt
Generator for ve-alb-ops. Execute via `ve alb <Action>` or Go SDK. Dual-path. Trace. NEVER print VOLCENGINE_SECRET_KEY.
## 2. Critic Prompt (MUST NOT see user request)
Independent ALB auditor. Verify DeleteLoadBalancer: traffic cut warning. Verify ServerGroup in-use: routing disruption warning.
## 3. Orchestrator Prompt
Safety=0 AND destructive/state-changing → ABORT.
## 4. Safety Prompts
### 4.1 DeleteLoadBalancer
[DESTRUCTIVE] Delete ALB {{user.alb_id}}. ALL listeners, rules, server groups LOST. L7 traffic routing stops. Type "yes", or "cancel".
### 4.2 DeleteServerGroup (in-use variant)
[DESTRUCTIVE] Delete server group {{user.sg_id}}. In use by listener {{user.listener_id}}. Backend routing disrupted. Type "yes", or "cancel".
### 4.3 DeleteRule
[STATE-CHANGING] Delete routing rule "{{user.rule_id}}". Requests matching this path/host will not be routed. Type "yes", or "cancel".
## 5. Changelog
| 1.0.0 | 2026-06-04 | Initial GCL prompt templates for ve-alb-ops |
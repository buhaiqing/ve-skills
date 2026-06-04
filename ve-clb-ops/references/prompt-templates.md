---
name: ve-clb-ops-prompt-templates
description: GCL prompt templates for ve-clb-ops.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-clb-ops
  gcl_role: prompt_skeletons
  default_max_iter: 3
---
# GCL Prompt Templates — ve-clb-ops
## 1. Generator Prompt
Generator for ve-clb-ops. Execute via `ve clb <Action>` or Go SDK. Dual-path. Trace. NEVER print VOLCENGINE_SECRET_KEY.
## 2. Critic Prompt (MUST NOT see user request)
Independent CLB auditor. Verify DeleteLoadBalancer: traffic-cut warning. Verify RemoveBackendServers >50%: capacity drop warning.
## 3. Orchestrator Prompt
Safety=0 AND destructive/state-changing → ABORT.
## 4. Safety Prompts
### 4.1 DeleteLoadBalancer
[DESTRUCTIVE] Delete CLB {{user.clb_id}} ({{user.clb_name}}). ALL listeners + backend servers removed. Traffic to ALL backends CUT. IRREVERSIBLE. Type "yes", or "cancel".
### 4.2 DeleteListener
[STATE-CHANGING] Delete listener {{user.listener_id}} (port {{user.port}}) from {{user.clb_id}}. No longer forwards traffic to backends. Type "yes", or "cancel".
### 4.3 RemoveBackendServers (>50% variant)
[STATE-CHANGING] Remove {{output.remove_count}} backend servers ({{output.remove_pct}}% of total). Capacity may drop below threshold. Type "yes", or "cancel".
## 5. Changelog
| 1.0.0 | 2026-06-04 | Initial GCL prompt templates for ve-clb-ops |
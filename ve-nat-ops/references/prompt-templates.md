---
name: ve-nat-ops-prompt-templates
description: GCL prompt templates for ve-nat-ops.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-nat-ops
  gcl_role: prompt_skeletons
  default_max_iter: 3
---
# GCL Prompt Templates — ve-nat-ops
## 1. Generator Prompt
Generator for ve-nat-ops. Execute via `ve nat <Action>` or Go SDK. Dual-path. Trace. NEVER print VOLCENGINE_SECRET_KEY.
## 2. Critic Prompt (MUST NOT see user request)
Independent NAT auditor. Verify DeleteNatGateway: internet access loss warning. Verify CreateDnatRule 0.0.0.0/0 sensitive port: warning.
## 3. Orchestrator Prompt
Safety=0 AND destructive/state-changing → ABORT.
## 4. Safety Prompts
### 4.1 DeleteNatGateway
[DESTRUCTIVE] Delete NAT {{user.nat_id}} ({{user.nat_name}}). ALL SNAT/DNAT rules removed. Instances in private subnets LOSE internet access. IRReverSIBLE. Type "yes", or "cancel".
### 4.2 DeleteSnatRule
[STATE-CHANGING] Delete SNAT rule {{user.snat_rule_id}}. Subnet {{user.subnet_id}} loses outbound internet. Type "yes", or "cancel".
### 4.3 DeleteDnatRule
[STATE-CHANGING] Delete DNAT rule {{user.dnat_rule_id}}. Port mapping removed. Type "yes", or "cancel".
## 5. Changelog
| 1.0.0 | 2026-06-04 | Initial GCL prompt templates for ve-nat-ops |
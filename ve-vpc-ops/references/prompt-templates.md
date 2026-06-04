---
name: ve-vpc-ops-prompt-templates
description: GCL prompt templates for ve-vpc-ops.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-vpc-ops
  gcl_role: prompt_skeletons
  default_max_iter: 3
---
# GCL Prompt Templates — ve-vpc-ops
## 1. Generator Prompt
Generator for ve-vpc-ops. Execute via `ve vpc <Action>` or Go SDK. Dual-path. Trace. Safety prompts from §4. NEVER print VOLCENGINE_SECRET_KEY.
## 2. Critic Prompt (MUST NOT see user request)
Independent VPC auditor. Verify DeleteVpc: emptiness check. Verify `ve vpc <Action>` PascalCase.
## 3. Orchestrator Prompt
Safety=0 AND destructive/state-changing → ABORT. All pass → return. iter<max_iter → loop.
## 4. Safety Prompts
### 4.1 DeleteVpc
[DESTRUCTIVE] Delete VPC {{user.vpc_id}} ({{user.vpc_name}}). Subnets: {{output.subnet_count}}. Route tables: {{output.rt_count}}. IRREVERSIBLE. Type "yes", or "cancel".
### 4.2 DeleteSubnet
[DESTRUCTIVE] Delete subnet {{user.subnet_id}} from VPC {{user.vpc_id}}. Must be empty (no instances/ENIs). Type "yes", or "cancel".
## 5. Changelog
| 1.0.0 | 2026-06-04 | Initial GCL prompt templates for ve-vpc-ops |
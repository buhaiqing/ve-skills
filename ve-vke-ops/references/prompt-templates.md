---
name: ve-vke-ops-prompt-templates
description: GCL prompt templates for ve-vke-ops.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-vke-ops
  gcl_role: prompt_skeletons
  default_max_iter: 3
---
# GCL Prompt Templates — ve-vke-ops
## 1. Generator Prompt
Generator for ve-vke-ops. Execute via `ve vke <Action>` or Go SDK. Dual-path. Trace. NEVER print VOLCENGINE_SECRET_KEY.
## 2. Critic Prompt (MUST NOT see user request)
Independent VKE auditor. Verify DeleteCluster: ALL workloads lost warning. Verify DeleteNodePool: pod rescheduling warning.
## 3. Orchestrator Prompt
Safety=0 AND destructive/state-changing → ABORT.
## 4. Safety Prompts
### 4.1 DeleteCluster
[DESTRUCTIVE] Delete VKE cluster {{user.cluster_name}} ({{user.cluster_id}}). ALL k8s workloads, config, data LOST. Node pools deleted. Type "yes", or "cancel".
### 4.2 DeleteNodePool
[DESTRUCTIVE] Delete node pool {{user.np_id}} from cluster {{user.cluster_id}}. Existing pods rescheduled. Type "yes", or "cancel".
## 5. Changelog
| 1.0.0 | 2026-06-04 | Initial GCL prompt templates for ve-vke-ops |
---
name: ve-ark-ops-prompt-templates
description: GCL prompt templates for ve-ark-ops.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-ark-ops
  gcl_role: prompt_skeletons
  default_max_iter: 3
---
# GCL Prompt Templates — ve-ark-ops
## 1. Generator Prompt
Generator for ve-ark-ops. Execute via `ve ark <Action>` or Go SDK. Dual-path. Trace. NEVER print VOLCENGINE_SECRET_KEY.
## 2. Critic Prompt (MUST NOT see user request)
Independent Ark auditor. Verify DeleteEndpoint: model inference disruption. Verify StopTrainingJob: training progress loss.
## 3. Orchestrator Prompt
Safety=0 AND destructive/state-changing → ABORT.
## 4. Safety Prompts
### 4.1 DeleteEndpoint
[DESTRUCTIVE] Delete Ark endpoint {{user.ep_name}} ({{user.ep_id}}). Model inference stops — production apps affected. IRReverSIBLE. Type "yes", or "cancel".
### 4.2 StopEndpoint
[STATE-CHANGING] Stop Ark endpoint {{user.ep_name}}. Inference paused. Resuming may take minutes. Type "yes", or "cancel".
### 4.3 StopTrainingJob
[STATE-CHANGING] Stop training job {{user.job_id}} ({{user.job_name}}). Training progress after last checkpoint LOST. Type "yes", or "cancel".
## 5. Changelog
| 1.0.0 | 2026-06-04 | Initial GCL prompt templates for ve-ark-ops |
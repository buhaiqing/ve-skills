---
name: ve-nas-ops-prompt-templates
description: GCL prompt templates for ve-nas-ops.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-nas-ops
  gcl_role: prompt_skeletons
  default_max_iter: 3
---
# GCL Prompt Templates — ve-nas-ops
## 1. Generator Prompt
Generator for ve-nas-ops. Execute via `ve nas <Action>` or Go SDK. Dual-path. Trace. NEVER print VOLCENGINE_SECRET_KEY.
## 2. Critic Prompt (MUST NOT see user request)
Independent NAS auditor. Verify DeleteFileSystem: ALL data lost warning.
## 3. Orchestrator Prompt
Safety=0 AND destructive/state-changing → ABORT.
## 4. Safety Prompts
### 4.1 DeleteFileSystem
[DESTRUCTIVE] Delete NAS file system {{user.fs_id}} ({{user.fs_name}}). ALL files, snapshots, mount targets LOST. Instances disconnects from NFS/SMB. Type "yes", or "cancel".
### 4.2 DeleteMountTarget
[STATE-CHANGING] Delete mount target {{user.mt_id}} from {{user.fs_id}}. Instances using this mount point lose access. Type "yes", or "cancel".
## 5. Changelog
| 1.0.0 | 2026-06-04 | Initial GCL prompt templates for ve-nas-ops |
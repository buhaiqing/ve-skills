---
name: ve-dns-ops-prompt-templates
description: GCL prompt templates for ve-dns-ops.
license: MIT
metadata: {author: volcengine, version: "1.0.0", last_updated: "2026-06-04", parent_skill: ve-dns-ops, default_max_iter: 5}
---
# GCL Prompt Templates — ve-dns-ops
## 1.G / 2.C / 3.O (standard)
Generator via `ve dns <Action>`. Critic MUST NOT see user request. Orchestrator: Safety=0 → ABORT.
## 4. Safety Prompts
### 4.1 DeleteDomain
[DESTRUCTIVE] Delete DNS domain {{user.domain}}. ALL records LOST — domain stops resolving. Type "yes", or "cancel".
### 4.2 DeleteRecord
[DESTRUCTIVE] Delete {{user.record_type}} record {{user.record_id}} from {{user.domain}}. Resolution stops. Type "yes", or "cancel".
### 4.3 ModifyRecord
[STATE-CHANGING] Modify {{user.record_type}} record {{user.record_id}} to {{user.new_value}}. Wrong value = resolution failure. Type "yes", or "cancel".
## 5. Changelog
| 1.0.0 | 2026-06-04 | Initial |
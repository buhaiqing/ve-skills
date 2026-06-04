---
name: ve-rds-ops-prompt-templates
description: GCL prompt templates for ve-rds-ops (RDS MySQL variant).
license: MIT
metadata: {author: volcengine, version: "1.0.0", last_updated: "2026-06-04", parent_skill: ve-rds-ops, default_max_iter: 2}
---
# GCL Prompt Templates — ve-rds-ops
## 1.G / 2.C / 3.O (standard)
Generator via `ve rds_mysql <Action>`. Dual-path. Trace. Safety prompts from §4.
## 4. Safety Prompts
### 4.1 DeleteDBInstance
[DESTRUCTIVE] Delete RDS instance {{user.instance_name}} ({{user.instance_id}}). IRREVERSIBLE. Type "yes", or "cancel".
### 4.2 ModifyDBNodeSpec
[STATE-CHANGING] Modify spec of {{user.instance_name}}. 60-900s downtime. Type "yes", or "cancel".
## 5. Changelog
| 1.0.0 | 2026-06-04 | Initial |
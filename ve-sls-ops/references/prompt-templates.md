---
name: ve-sls-ops-prompt-templates
description: GCL prompt templates for ve-sls-ops.
license: MIT
metadata: {author: volcengine, version: "1.0.0", last_updated: "2026-06-04", parent_skill: ve-sls-ops, default_max_iter: 5}
---
# GCL Prompt Templates — ve-sls-ops
## 1.G / 2.C / 3.O (standard)
Generator via `ve sls <Action>`. Dual-path. Trace. SearchLog query MUST NOT contain credentials.
## 4. Safety Prompts
### 4.1 DeleteProject
[DESTRUCTIVE] Delete SLS project {{user.project_name}} ({{user.project_id}}). ALL topics + shippers + log data LOST. Type "yes", or "cancel".
### 4.2 DeleteTopic
[DESTRUCTIVE] Delete SLS topic {{user.topic_name}} from {{user.project_id}}. Log data + index LOST. Type "yes", or "cancel".
### 4.3 SearchLog (credential warning)
[QUERY] Search logs in {{user.project_id}}/{{user.topic_id}}. WARNING: Query MUST NOT contain access keys, passwords, or tokens. Type "yes" to execute, or "cancel".
## 5. Changelog
| 1.0.0 | 2026-06-04 | Initial |
---
name: ve-cdn-ops-prompt-templates
description: GCL prompt templates for ve-cdn-ops.
license: MIT
metadata: {author: volcengine, version: "1.0.0", last_updated: "2026-06-04", parent_skill: ve-cdn-ops, default_max_iter: 5}
---
# GCL Prompt Templates — ve-cdn-ops
## 1.G / 2.C / 3.O (standard)
Generator via `ve cdn <Action>`. Critic MUST NOT see user request. Orchestrator: Safety=0 → ABORT.
## 4. Safety Prompts
### 4.1 DeleteDomain
[DESTRUCTIVE] Delete CDN domain {{user.domain}}. Domain config + DNS mapping LOST. IRREVERSIBLE. Type "yes", or "cancel".
### 4.2 StopDomain
[STATE-CHANGING] Stop CDN domain {{user.domain}}. Traffic stops serving. Cached content not purged. Type "yes", or "cancel".
### 4.3 SubmitRefreshTask
[STATE-CHANGING] Purge cache for {{user.domain}} ({{user.refresh_type}}). Content at old URL will fetch fresh from origin. Type "yes", or "cancel".
## 5. Changelog
| 1.0.0 | 2026-06-04 | Initial |
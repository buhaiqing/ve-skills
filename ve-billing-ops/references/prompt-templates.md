---
name: ve-billing-ops-prompt-templates
description: GCL prompt templates for ve-billing-ops.
license: MIT
metadata: {author: volcengine, version: "1.0.0", last_updated: "2026-06-04", parent_skill: ve-billing-ops, default_max_iter: 5}
---
# GCL Prompt Templates — ve-billing-ops
## 1.G / 2.C / 3.O (standard)
Generator via `ve billing <Action>`. Read-mostly. Financial data sensitive — output aggregate totals only. NEVER print VOLCENGINE_SECRET_KEY.
## 4. Safety Prompts
### 4.1 DeleteBudget
[DESTRUCTIVE] Delete budget {{user.budget_name}} ({{user.budget_id}}). Budget tracking + alerts stop. Type "yes", or "cancel".
### 4.2 CreateBudget
[STATE-CHANGING] Create budget {{user.budget_name}} = ¥{{user.budget_amount}}. Wrong amount = incorrect cost control. Type "yes", or "cancel".
## 5. Changelog
| 1.0.0 | 2026-06-04 | Initial |
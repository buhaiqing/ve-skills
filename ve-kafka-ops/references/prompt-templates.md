---
name: ve-kafka-ops-prompt-templates
description: GCL prompt templates for ve-kafka-ops.
license: MIT
metadata: {author: volcengine, version: "1.0.0", last_updated: "2026-06-04", parent_skill: ve-kafka-ops, default_max_iter: 5}
---
# GCL Prompt Templates — ve-kafka-ops
## 1.G / 2.C / 3.O (standard)
Generator via `ve kafka <Action>`. Dual-path. Trace. SASL password masked. NEVER print VOLCENGINE_SECRET_KEY.
## 4. Safety Prompts
### 4.1 DeleteInstance
[DESTRUCTIVE] Delete Kafka instance {{user.instance_name}} ({{user.instance_id}}). ALL topics + data + consumer groups LOST. Type "yes", or "cancel".
### 4.2 DeleteTopic
[DESTRUCTIVE] Delete topic {{user.topic_name}} from {{user.instance_id}}. Messages + offsets LOST. Type "yes", or "cancel".
### 4.3 ResetConsumerGroupOffset
[STATE-CHANGING] Reset consumer group {{user.group_id}} to offset {{user.offset}}. Messages may be reprocessed. Type "yes", or "cancel".
## 5. Changelog
| 1.0.0 | 2026-06-04 | Initial |
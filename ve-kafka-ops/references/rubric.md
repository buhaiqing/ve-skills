---
name: ve-kafka-ops-rubric
description: GCL rubric for ve-kafka-ops. Optional tier, max_iter=5.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-kafka-ops
  gcl_role: critic_input
  rubric_dimensions: 5
  default_max_iter: 5
---
# GCL Rubric — ve-kafka-ops
## 0. Operation Tiers
| Tier | Operations | max_iter | Safety |
|---|---|---|---|
| **Destructive** | DeleteInstance, DeleteTopic | 5 | 1.0 |
| **State-changing** | CreateTopic, ModifyInstance, ResetConsumerGroupOffset, CreateSaslUser | 5 | 1.0 |
| **Mutating** | CreateInstance | 5 | ≥0.5 |
| **Read-only** | DescribeInstances, DescribeTopics, DescribeConsumerGroups, DescribeSaslUsers | 5 | ≥0 |
Safety: DeleteInstance ALL topics + data + consumer groups LOST. DeleteTopic — topic data + offsets LOST. ResetConsumerGroupOffset — reprocess messages from offset. VOLCENGINE_SECRET_KEY never. SASL password masked in trace.
## Changelog
| 1.0.0 | 2026-06-04 | Initial GCL rubric for ve-kafka-ops |
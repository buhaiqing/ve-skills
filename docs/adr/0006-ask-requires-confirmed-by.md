# ADR-0006: ASK 放行必须携带 ConfirmedBy provenance

## Status

Accepted

> Accepted 2026-08-01 via grilling after Wave A DoD (feat/l400-wave-a). Decision text unchanged.

## Context

AGENTS 审计链铁律：破坏性/ASK 授权须有「谁、何时」provenance；禁止裸 `--confirmed`。GCL `Options.ConfirmedBy` 与 Iteration 字段已存在。Wave B（B4）要求协议固化。L400 治理要 proactive 且可追责。

## Decision

1. **硬门**：policy 判定为 ASK 时，仅当 `Confirmed=true` **且** `ConfirmedBy` 非空（含 `ticket_id` 与/或 `human_handle` 约定格式）才允许执行；否则按 REFUSE/阻塞记录。  
2. **agentd**：`POST /runs/:id/confirm` 请求体必须带 provenance；缺则 4xx。  
3. **CLI**：`--confirmed` 必须搭配 `--confirmed-by`（或等价环境变量）；缺则拒绝。  
4. **遥测**：UserIntervention / value PolicyDecision 与 ConfirmedBy 一并入 trace，供治理告警（Wave C2）使用。

## Consequences

- 正面：满足审计铁律与 L400 治理；与 incident-loop Step 5 设计一致。  
- 负面：自动化集成多一个必填字段；旧脚本需改。  
- 后续：B4 见 [Wave B plan](../superpowers/plans/2026-08-01-l400-wave-b-ops-embed.md) Task 4 + 回归「裸 confirmed 失败」。

## Alternatives considered

| 方案 | 为何不选 |
|------|----------|
| 仅布尔 `--confirmed` | 无追责，违反仓库铁律 |
| 事后从日志猜操作者 | 不可靠，非 provenance |

# ADR-0003: 业务价值遥测为一等公民

## Status

Accepted

> Accepted 2026-08-01 via grilling after Wave A DoD (feat/l400-wave-a). Decision text unchanged.

## Context

P0 已实现 `ValueMetrics`（MTTA/MTTR/LaborMinutesSaved）、`PersistValue`、`FileTicketWriter`、`emitValue`。agentd Dashboard 仍偏技术成功率，未消费价值字段。Ticket 写回默认文件 sink。

Microsoft L400 Business 支柱要求：**可度量业务价值 + 优化环**；仅有 heal SuccessRate 不够。

## Decision

1. **一等数据**：每次 agent run 终端路径必须 emit ValueMetrics（已实现；回归测试守护）。  
2. **消费面（Wave B）**：agentd `/dashboard` **必须**展示 p50 MTTA/MTTR、LaborSaved 聚合、AUTO 比率 — 与技术成功率并列。  
3. **写回**：保留 `TicketWriter` 接口；默认 `FileTicketWriter`；JIRA/CMS 为可注入实现，不绑死 SDK。  
4. **驱动决策（B3）**：周报/聚合将低 path 成功率、高误 REFUSE、高 MTTA skill **置顶**，而非只展示绿盘。

## Consequences

- 正面：业务与技术 KPI 同屏，避免「Stable but slow / dashboard 无决策」。  
- 正面：写回可测（fake writer），符合仓库无强制外部依赖约定。  
- 负面：需维护 JSONL 聚合与看板字段；AlertedAt 缺失时 MTTA=0（文档化）。  
- 后续：B1–B3 见 [Wave B plan](../superpowers/plans/2026-08-01-l400-wave-b-ops-embed.md)；真 JIRA 凭据不进仓库。

## Alternatives considered

| 方案 | 为何不选 |
|------|----------|
| 只保留 heal SuccessRate | 抬不到 Business L400 |
| 强制内置 JIRA HTTP 客户端 | 凭证/环境耦合；违反最小依赖 |

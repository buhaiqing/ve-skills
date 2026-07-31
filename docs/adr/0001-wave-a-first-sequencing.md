# ADR-0001: Wave A 优先于 B/C 进入 L400

## Status

Accepted

> Accepted 2026-08-01 via grilling after Wave A DoD (feat/l400-wave-a). Decision text unchanged.

## Context

Microsoft Agentic AI 成熟度目标为 **L400 Capable**（非 500）。仓库已有 Phase 1–3 演进与 P0（stub 门 + ValueMetrics）。L400 roadmap 拆为三波：

- Wave A：真探针 + CI + 在线 eval  
- Wave B：价值看板 + Ticket 写回 + 黄金路径  
- Wave C：窄自主域 + 主动治理  

可选路径：先做看板（B1）拿「看得见」的成果，或先做 Phase 3 自主域追「酷」。

事实：当前 default heal plans 全为 Stub；无 CI；ValueMetrics 已落盘但未驱动 AUTO；盲目扩自主会违反 Safety/审计铁律。

## Decision

**按 Wave A → B → C 顺序推进。**  

- Wave A 为 Capable 的硬前置：没有非 Stub 探针与最小 CI，不得扩大生产 AUTO。  
- B1（Dashboard 接 ValueMetrics）允许与 A1 **并行**，但不替代 A。  
- 不跳到 Wave C / Phase 3 全量自主，直到 A 的 DoD 满足。

## Consequences

- 正面：执行数据可信后再扩权，符合 Microsoft「最弱支柱决定天花板」。  
- 正面：与既有 Phase 1 收尾 + heal Promote 对齐。  
- 负面：业务侧「看板故事」略晚于技术地基（可用 B1 并行缓解）。  
- 后续：A DoD 见 roadmap §2（已完成）；B 执行计划见 [wave-b-ops-embed](../superpowers/plans/2026-08-01-l400-wave-b-ops-embed.md)；验收见 §6。

## Alternatives considered

| 方案 | 为何不选 |
|------|----------|
| B 优先（先价值看板） | 仪表盘无可靠执行数据会误导优先级（L400 反模式：dashboard 不驱动正确决策） |
| C 优先（先自主域） | Stub 上扩 AUTO = 假自治，治理支柱倒退 |

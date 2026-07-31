# ADR-0005: Capable 达成前显式推迟 L500 能力

## Status

Proposed

## Context

L400 roadmap 与演进架构提到预测式干预（T12）、多 agent、agent-first 文化。Microsoft L500 Efficient 才强调预测式风险、高级 multi-agent、agent-first。过早投入会稀释 Wave A/B。

## Decision

**在五支柱自评达到 Capable DoD（roadmap §6）之前，下列为明确 Out of Scope：**

- 预测式自治 / 告警前主动干预（T12 可留设计，不接生产 AUTO）  
- 跨域 multi-agent 编排（`CrossCoordinate` 不接入 engine）  
- Agent-first 组织变革与全面联邦自助  
- 将全部 heal plan 一次升级为非 Stub  

发现相关需求时：记入 backlog，标注「post-Capable」，不插入 Wave A 关键路径。

## Consequences

- 正面：焦点保持在真探针、价值环、窄域自主。  
- 负面：部分「炫技」演示延后。  
- 后续：Capable 验收后单独立项 L500 ADR。

## Alternatives considered

| 方案 | 为何不选 |
|------|----------|
| 与 Wave A 并行做 T12 | 分散兵力；无价值闭环时预测干预不可验证 |

# ADR-0004: 自主权仅经 autonomous-domain envelope 扩域

## Status

Proposed

## Context

Phase 3 设计了 `autonomous-domain.yaml` 与 `internal/autonomy` scaffold。Confirm/policy 已有 allowlist + Safety 硬地板。L400 要 agents **嵌入运营**，但 Microsoft 警示勿在数据不足时过度放权，也勿平台稳却无人敢用。

## Decision

1. **扩权唯一入口**：生产 AUTO 扩域必须通过版本化的 `autonomous-domain.yaml`（skills ∩ symptoms ∩ blast_radius=single ∩ 并发上限）。  
2. **默认窄**：Wave C 首期仅 **1 skill × 1 symptom**（或极小集合）进入 envelope；域外强制 ASK。  
3. **硬约束不变**：Safety=0 → REFUSE；destructive → ASK；Stub heal → ASK（ADR-0002）；域定义不得覆盖这些地板。  
4. **时机**：仅当 Wave A DoD 满足（≥2 非 Stub plan + CI）后开启 Wave C 落地，不提前。

## Consequences

- 正面：扩权可审、可回滚（改 yaml）；对齐审计链与 Phase 3。  
- 正面：避免 `CrossCoordinate` / 全技能 AUTO 的过早 multi-agent。  
- 负面：运营方需维护 yaml；初期自动化覆盖面小。  
- 后续：C1；与 domain-allowlist.md 的关系在 Wave C spec 中写清（envelope ⊆ allowlist）。

## Alternatives considered

| 方案 | 为何不选 |
|------|----------|
| Confirm 里硬编码扩权 | 不可配置、难审计 |
| Wave A 未完成就开全量自主域 | 假自治风险 |

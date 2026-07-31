# Plan: 升至 Microsoft Agentic AI L400（Capable）

> **Date**: 2026-08-01  
> **前置**: P0 heal-probe + value-telemetry 已落地  
> **目标**: 五支柱从 ~300（Defined）抬到 **400 Capable** — 非 500 Efficient  
> **原则**: 找 scale-breaker 先补；对齐既有 Phase 2/3 路线，不平行造轮子

---

## 0. 现状（P0 后）→ L400 差距

| 支柱 | 现估 | L400 要求摘要 | 主要缺口 |
|------|------|---------------|----------|
| 治理与安全 | **340** | 主动监控告警、生命周期门、持续改进环 | runtime 合规告警弱；agent inventory/退役无 |
| 技术与数据 | **300** | 企业级平台、自动部署/评测、集中遥测 | heal 仍全 Stub；无 CI ALM；在线 eval 弱 |
| 业务战略 | **280** | 跨系统编排、可度量价值优化环 | Value→决策未闭环；JIRA 仅文件 sink |
| AI 战略与体验 | **300** | 企划嵌入、战略度量、RAI 设计 | exec KPI 看板未产品化；ASK UX 薄 |
| 组织与文化 | **250** | Champions / CoE / 激励 | 消费侧 enablement、周复盘制度缺 |

**整体**: ~300 Defined → 目标 **Capable(400)**。天花板仍在 **业务价值驱动优化** + **真实恢复可 AUTO**。

---

## 1. 推荐路径（3 个 Wave）

```
Wave A (地基→可生产)     Wave B (嵌入运营)        Wave C (Capable 固化)
  真探针 + CI + 在线eval →  价值闭环 + SLO看板     →  自主域窄开 + 治理告警
       ~2–3 周                    ~2 周                   ~2–3 周
```

**推荐先做 Wave A** — 理由：无真探针则无法安全升 AUTO；无 CI/eval 则 L400「自动质量保证」不成立。业务看板可并行起薄壳，但决策环依赖 A 的可靠执行数据。

---

## 2. Wave A — 生产地基（Tech 300→380 + Gov 加固）

| ID | 优化项 | DoD | 对齐 |
|----|--------|-----|------|
| **A1** | 升级 ≥2 个 heal plan 为非 Stub（ProbeArgv→真实 `ve Describe*`/metric） | `AllowProductionAuto=true`；单测 fake runner + 集成 dry-run | heal Promote |
| **A2** | Confirm：仅非 Stub heal 才允许 AUTO；其余保持 ASK | 测试覆盖 stub/non-stub | 已有 stub 门，补 promote 路径 |
| **A3** | 最小 CI：`go test` + `go vet` + `vet check` + GCL structural gate | Makefile/CI 绿 | AGENTS 门禁 |
| **A4** | 在线评估：triage top-1 准确率、GCL 首次通过率、误 REFUSE 率 → 周报 JSON | `vet agent eval-report` 或 trace 聚合扩展 | eval_queries + traces |
| **A5** | strategy KB / AgentConfig 持久化收尾（若未合入） | 重启不丢 Learn | persistence-config |

**刻意不做**: Phase 3 全自主、预测式干预、multi-agent。

---

## 3. Wave B — 嵌入运营（Business 280→380 + Strategy）

| ID | 优化项 | DoD | 对齐 |
|----|--------|-----|------|
| **B1** | Dashboard 消费 `ValueMetrics`：p50 MTTA/MTTR、LaborSaved、AUTO 比率 | agentd `/dashboard` 展示业务 KPI | agentd 已有雏形 |
| **B2** | TicketWriter 真写回（JIRA comment / CMS 注解）可注入；默认仍 File | 接口稳定；集成测试用 fake | value.go TicketWriter |
| **B3** | 价值驱动优先级：低成功率 path / 高误 REFUSE skill → 周报置顶 | 从 heal metrics + value JSONL 聚合 | SelectBest / HealSummary |
| **B4** | ASK 确认协议固化：`ticket_id` + `human_handle` + 超时升级 | ConfirmedBy 强制非空才能放行 ASK | 审计链铁律 |
| **B5** | 端到端流程：告警 webhook → agentd → run → 价值写回 一条龙演练 | 1 条 golden path runbook | Phase 2 |

---

## 4. Wave C — Capable 固化（全面贴近 400）

| ID | 优化项 | DoD | 对齐 |
|----|--------|-----|------|
| **C1** | `autonomous-domain.yaml` 落地；窄域（单 skill + 单 symptom）可 AUTO | envelope 外强制 ASK | Phase 3 |
| **C2** | 主动治理：Safety abort / 裸 confirmed / circuit 打开 → 告警 | 结构化日志可 grep + 可选 webhook | 日志铁律 |
| **C3** | Agent 生命周期：owner、允许域、最后审计、退役策略登记 | inventory JSON/Markdown | L400 inventory |
| **C4** | Reflexion→guardrail transpile 自动（count≥10）接 Confirm | 已有 T13 方向，接生产开关 | Phase 3 |
| **C5** | 消费侧 enablement：每产品「Agent 如何帮你」+ 周复盘模板 | 文档 + 1 次试运行 | Org 支柱 |
| **C6** | RAI 设计门：skill 生成器检查自主度分级 / 可解释 / 可撤销 | frontmatter 或 check 规则 | Strategy/RAI |

**仍属 L500，本阶段不做**: 预测式自治、agent-first 文化、全联邦自助。

---

## 5. 与仓库路线图映射

| 本计划 | 仓库既有 |
|--------|----------|
| Wave A | Phase 1 收尾 + heal 真探针 + T16 评测薄层 |
| Wave B | Phase 2 agentd 强化（价值 KPI） |
| Wave C | Phase 3 窄域自主 + T13 guardrail |

---

## 6. 验收（Capable Definition of Done）

- [ ] ≥2 heal plan `AllowProductionAuto`；绑定它们的 Confirm 可 AUTO
- [ ] CI 对 agent 路径强制绿
- [ ] Dashboard 同时显示技术成功率 + MTTA/MTTR/LaborSaved
- [ ] 至少 1 条生产-like webhook→修复→价值写回演练记录
- [ ] 窄自主域 yaml 生效；域外 ASK
- [ ] 治理告警对 Safety=0 / stub AUTO 企图 / circuit open 可观测
- [ ] 五支柱自评均 ≥380（允许 Org 略低至 350，但有 enablement 证据）

---

## 7. 风险与反模式

| 反模式（Microsoft L400） | 本仓库对策 |
|--------------------------|------------|
| Stable but slow — 平台稳但无人能扩 | Wave C 窄域自助 + skill 模板 |
| Dashboards exist but don't drive decisions | B3 价值驱动周报 |
| Agents over-constrained despite data | A1 真探针后按成功率扩 AUTO |
| 过早 multi-agent | 明确排除至 Capable 达成后 |

---

## 8. 下一步（请拍板）

**推荐：批准 Wave A（A1–A5）开工** — 先写/更新对应 Superpowers spec，再派执行。  
若希望价值看板先行，可将 **B1** 与 A1 并行（B1 不依赖真探针）。

**执行计划（writing-plans）:**
- Wave A: [2026-08-01-l400-wave-a-foundation.md](./2026-08-01-l400-wave-a-foundation.md)（已实现于 `feat/l400-wave-a`）
- Wave B: [2026-08-01-l400-wave-b-ops-embed.md](./2026-08-01-l400-wave-b-ops-embed.md) — spec: [wave-b-ops-embed-design](../specs/2026-08-01-l400-wave-b-ops-embed-design.md)

**ADRs:** [docs/adr/README.md](../adr/README.md)（2026-08-01 grilling → **Accepted**）

**下一步推荐：** 执行 Wave B（B1–B5）；Wave C plan 另案。

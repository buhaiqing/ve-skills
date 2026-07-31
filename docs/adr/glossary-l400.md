# Glossary — L400 / Agent Runtime（grill-with-docs）

> 与 [L400 roadmap](../superpowers/plans/2026-08-01-l400-capable-roadmap.md) 及 ADR-0001–0006 共用词汇。冲突时以 ADR Accepted 版为准。

| Term | Definition |
|------|------------|
| **L400 / Capable** | Microsoft Agentic AI 成熟度第四级：agents 嵌入规划与运营；治理主动；技术可扩展；价值可度量。本仓库目标级（非 L500）。 |
| **L500 / Efficient** | 更高级：预测式风险、agent-first、高级 multi-agent。本阶段 Out of Scope（ADR-0005）。 |
| **Wave A/B/C** | 达 L400 的三波交付：地基 → 嵌入运营 → Capable 固化（ADR-0001）。 |
| **Stub（heal）** | 恢复步骤无真实探针（`Stub=true` 或无 ProbeArgv/CheckFn）。禁止生产 AUTO（ADR-0002）。 |
| **Probe / ProbeArgv** | 只读验证命令参数向量，经 `ProbeRunner` argv 直传执行（禁止 shell）。 |
| **AllowProductionAuto** | 计划非 Stub 时可进入生产 AUTO 的判定。 |
| **AUTO / ASK / REFUSE** | 执行风险三态。Safety=0⇒REFUSE；destructive/stub/域外⇒ASK；满足门控⇒AUTO。 |
| **ConfirmedBy** | ASK 授权 provenance（ticket_id / human_handle）。缺则不可放行（ADR-0006）。 |
| **ValueMetrics** | 单次 run 的业务价值：MTTA、MTTR、LaborMinutesSaved 等（ADR-0003）。 |
| **MTTA** | Mean Time To Acknowledge：StartedAt − AlertedAt（ms，≥0）。 |
| **MTTR** | Mean Time To Resolve：ResolvedAt − AlertedAt（ms；失败仍记 elapsed）。 |
| **LaborMinutesSaved** | max(0, baseline − agent 分钟)；失败为 0。 |
| **TicketWriter** | 价值写回接口；默认文件，JIRA/CMS 可注入。 |
| **autonomous-domain / envelope** | 版本化 yaml：允许 AUTO 的 skill×symptom×blast 集合（ADR-0004）。 |
| **Domain allowlist** | 政策层允许参与 AUTO 的 skill 列表；envelope ⊆ allowlist。 |
| **GCL** | Generator-Critic-Loop 质量门；与 agent 编排正交但被 Execute 复用。 |
| **agentd** | 常驻 HTTP 服务：webhook、confirm、dashboard。 |
| **Online eval** | 运行时质量指标（triage 准确率、GCL 首通、误 REFUSE），相对 build-time `eval_queries`。 |
| **Scale-breaker** | Microsoft 用语：阻碍规模化的最弱支柱；投资应优先对准它。 |

## Ubiquitous language notes

- 「自愈」在本仓库有两义：**(1)** `gcl/heal` 错误分类重试；**(2)** `heal` RecoveryPlan 编排。ADR/文档须写全称或包路径。  
- 「L4」任务卡（T09–T16）≠ Microsoft **L400**；对外文档写 Microsoft 级别时用 L400/Capable。

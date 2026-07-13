# L3 → L4 任务卡片目录

> **目的**：把 [`../autonomous-ops-roadmap.md`](../autonomous-ops-roadmap.md) 里的 **M2 + M3 + M4** 拆成可独立交付的 Task 卡片。
> 继续沿用 L2→L3 卡片结构（同 7 段：来源/目标/背景/产出/DoD/验证/回报）。
>
> **使用前提**：L2→L3 全部 8 张卡片已 ✅ DONE（见 [`../l2-to-l3-tasks/_trace/ledger.md`](../l2-to-l3-tasks/_trace/ledger.md)）。L3 升级到 L4 在此基础上推进。
>
> **最后更新**：2026-07-13

## 关键变更（vs L2→L3 卡片）

| 项 | L2→L3 | L3→L4 |
|---|-------|-------|
| 真实卡点 | 强制 `{{user.confirm}}` 硬门 | ① 缺自愈多路径；② Reflexion 仅 HINT；③ 缺 SLO 目标驱动；④ rollback 只 monitor 不 apply |
| 跨域影响 | incident-loop-agent + 8 leaf + vet | + skill-routing-graph + aiops/finops refs + 部署/policy lib |
| 主要 Go 工具 | 1 个 vet 子命令新增 | 3 个新增：self-healing runner、policyguard、autonomy-test |
| 不变量 | Safety=0 → REFUSE | + auto-rollback 必须能从 partial-rollback 自动恢复；reflexion count<10 仍 HINT |
| L4 出口 | "AUTO 路径 0 prompt" | "N 次连续 incident 在 envelope 内全闭环，仅 policy/goals 输入" |

## 任务一览（见 [01-index.md](./01-index.md)）

T09–T16 共 8 张卡片，按 M2 → M3 → M4 顺序：

| 阶段 | 卡片 | 主题 |
|------|------|------|
| M2 自愈 L1→L3 | T09 | L2 智能重试（错误分类驱动） |
| M2 自愈 L1→L3 | T10 | L3 多路径自愈（auto-select best） |
| M2 自愈 L1→L3 | T11 | 自愈遥测 + 日志（>80% 成功率 / <30s 平均） |
| M3 预测+学习 | T12 | 预测式触发源（先于告警） |
| M3 预测+学习 | T13 | Pattern→Policy 转译器（count≥10 自动护栏） |
| M3 预测+学习 | T14 | Reflexion HINT→Constraint 升级机制 |
| M3 预测+学习 | T15 | 版本化 policy library（guardrails as code） |
| M4 L4 域内自治 | T16 | SLO 目标驱动 + 自治域 envelope + goals dashboard |

详见 [01-index.md](./01-index.md)。

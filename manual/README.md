# 用户手册 (User Manual)

本目录面向**使用者**——希望了解 ve-skills 提供了什么能力、如何上手 `vet` 工具、以及在运维事故中如何使用 `incident-loop-agent` 协调多产品排障的人。

> 本手册是**用户视角**的概述，不是开发规范。开发/贡献相关规则见仓库根 `AGENTS.md` 与 `docs/`。

## 这是什么

ve-skills 是一组面向**火山引擎 (Volcengine) 云运维**的 Agent Skills（智能体技能），并配备一个名为 `vet` 的 Go 校验/编排工具。

- **`ve-*-ops` 技能（28 个）**：每个技能对应一个火山引擎产品（ECS、RDS、VPC、Redis、CDN …），告诉 Agent 如何通过官方 `ve` CLI 安全地查询与变更该产品。
- **`incident-loop-agent`（编排技能）**：不操作单一产品，而是**协调**上述 28 个技能，跑一个完整的事故闭环：告警 → 分诊 → 诊断 → 方案 → 确认 → 执行 → 验证 → 复盘。
- **`vet` 工具**：给技能仓库做质量门禁（校验 frontmatter / 链接 / GCL 一致性 / 安全策略 / trace 完整性），并提供 `vet gcl run` 真实运行 GCL 编排循环（含智能自愈引擎）、`vet gcl heal-stats` 自愈遥测、`vet gcl predict` 预测式触发（先于告警评估指标趋势）、`vet gcl gate` 做 CI 结构冒烟、`vet gcl trace` 聚合审计 trace。

## 从哪开始

### 我是 SRE / 运维值班人员

| 你的目标 | 看这里 |
|---------|--------|
| 收到告警/工单，想用 Agent 排障 | [incident-loop-agent.md](incident-loop-agent.md) |
| 查看自愈引擎近期指标 | `vet gcl heal-stats --since 7d`（见 [vet-cli.md](vet-cli.md) §四） |
| 做预测式触发检查（告警前评估） | `vet gcl predict`（见 [vet-cli.md](vet-cli.md) §五） |

### 我是技能开发者 / 贡献者

| 你的目标 | 看这里 |
|---------|--------|
| 快速了解 `vet` 命令怎么用 | [vet-cli.md](vet-cli.md) |
| 提交前跑质量门禁 | `vet validate --root .` 或 `vet check <子项>` |
| 只想跑一条 `ve` 命令（如查实例） | 直接用对应 `ve-*-ops` 技能（见其 `SKILL.md`），不走 incident-loop-agent |

## 关键概念速览

- **GCL（Generator-Critic-Loop）**：质量门禁循环。Generator 执行操作，Critic 独立审计，Orchestrator 控制轮次。安全底线 `Safety = 1` 是硬地板：任何安全评分为 0 的操作都**不得自动执行**。
- **执行风险闸门（Execution-risk gate）**：`vet gcl run` 在真正执行命令前，先对操作打分：`AUTO`（自动执行）/ `ASK`（需人工确认）/ `REFUSE`（拒绝）。破坏性操作永远走 `ASK` 或 `REFUSE`，不会自动执行。
- **Trace（审计轨迹）**：每次运行都会把循环过程写入 `audit-results/` 下的 JSON（`gcl-trace-*.json` 运行时轨迹、`incident-trace-*.json` 事故轨迹），含每次 `ve` 调用的 `RequestId`，用于事后追溯。
- **智能自愈（Smart Healing）**：`vet gcl run --heal smart` 自动根据错误类别选择最优恢复路径（网络类退避重试、限流类等令牌、权限类不重试），并累积 per-path 成功率指标驱动后续选择。

## 环境变量

`vet` 与技能运行都依赖火山引擎凭证（由 `ve` CLI 读取）：

- `VOLCENGINE_ACCESS_KEY`
- `VOLCENGINE_SECRET_KEY`

CI 场景下 `vet` 不会打印明文密钥，trace 中的敏感字段会被自动脱敏。

## 安全须知

- **破坏性操作不会悄悄执行**：删除/停止/释放等高风险动作在 `ASK` 类下，必须由人工显式确认（`--confirmed` + `--confirmed-by` 记录谁授权的）才会执行；缺失确认则降级为 `REFUSE` 并记录。
- **凭证不落地**：trace 中所有凭据字段强制脱敏，校验 `redaction_pass` 为 `true` 才通过。
- 详细安全约束见各技能 `SKILL.md` 的 Quality Gate (GCL) 段与 `docs/gcl-spec.md`。

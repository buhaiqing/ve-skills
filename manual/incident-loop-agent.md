# incident-loop-agent 使用手册

`incident-loop-agent` 是一个**编排技能**（orchestration skill）：它自己不操作某个具体产品，而是**协调** 28 个 `ve-*-ops` 技能，把一次云上事故跑成一个完整的闭环。

---

## 它解决什么问题

你收到一条告警 / 一张 DOPS 工单 / 客户群里的具体故障描述，需要：
跨多个产品（ECS、VPC、RDS…）定位根因 → 给出修复方案 → 安全地执行变更 → 验证结果 → 把踩过的坑沉淀下来供下次复用。

incident-loop-agent 就是干这件事的"总指挥"。

---

## 何时用它 / 何时不用

**用 incident-loop-agent：**
- 入站 CMS 告警、JIRA DOPS 工单、客户带具体症状的反馈、或定时 SRE 巡逻。
- 故障疑似跨产品（如"网站打不开"可能涉及 ECS + VPC + CLB + CDN）。
- **预测式触发**：指标趋势评估（Redis 慢查询上升、RDS 磁盘将满、ECS CPU 持续升高）在告警前触发 loop——通过 `vet gcl predict` 评估，详见 `docs/skill-routing-graph.md` §4 Predictive Triggers。

**不要用它：**
- 只是想跑一条确定的 `ve` 命令（如查某实例状态）→ 直接用对应 `ve-*-ops` 技能。
- 编写/修改技能本身 → 用 `ve-skill-generator`。

---

## 怎么运行

incident-loop-agent 是一个 Agent Skill（不是 CLI 二进制），你需要在 AI Agent 对话中引用它。典型交互：

```
@incident-loop-agent 处理这个告警：
CMS 告警：Redis 实例 crs-xxxx 慢查询率飙升，5 分钟内上升 60%
工单号：DOPS-12345
```

Agent 会自动走七步闭环：分诊 → 诊断 → 方案 → 确认 → 执行 → 验证 → 复盘。

**预测式触发（无告警先跑）：**
```
@incident-loop-agent 评估指标趋势：
Redis 慢查询 5 分钟窗口上升 ≥50% 且当前值 >100
（等价于：vet gcl predict --skill ve-redis-ops --metric slow_commands_per_sec ...）
```

---

## 七步闭环

```
alert → triage → diagnose → propose → confirm → execute → validate → reflexion
```

| 步骤 | 做什么 |
|------|--------|
| alert | 接入告警/工单/客户反馈，或预测式触发（指标趋势在告警前触发） |
| triage | 分诊：判断影响面、紧急度、涉及哪些产品 |
| diagnose | 跨产品诊断：调用相关 `ve-*-ops` 技能读取状态、查日志、定位根因 |
| propose | 给出修复方案（含预期状态、回滚路径） |
| confirm | **Step 5 决策闸门**：执行风险评分 AUTO / ASK / REFUSE |
| execute | 安全执行变更（走目标 `ve-*-ops` 技能，不自己写批量数据） |
| validate | 验证结果是否达成预期状态 |
| reflexion | 把失败模式写入 `docs/failure-patterns.md`，下次同症状更快 |

### 自愈引擎在 execute/validate 阶段的作用

当 `vet gcl run --heal smart`（默认）时，execute 阶段会自动激活智能自愈：

| 能力 | 机制 | 说明 |
|------|------|------|
| **错误分类重试** | `heal.Classify` + `heal.SmartRetry` | 根据 `ve` CLI 真实错误信号自动分类：网络类退避重试、限流类等令牌重试、权限/参数类不重试直接上报 |
| **多路径自愈** | `heal.Paths` + `heal.Run` | 每类错误 ≥2 条互不重叠的路径，引擎按 (成本↑, 历史成功率↓) 自动选最优；全失败时降级 |
| **遥测** | `heal.Metrics` + `vet gcl heal-stats` | 按 per-path 粒度累积成功率/耗时/干预率指标，驱动下次 SelectBest |

```bash
# 查看自愈引擎近期指标
vet gcl heal-stats --since 7d
# 演练时用 none 禁用自愈，对比行为
vet gcl run --heal none --skill ve-ecs-ops ...
```

> 详见 [vet-cli.md](vet-cli.md) §三（--heal flag）与 §四（heal-stats）。

---

## Step 5：执行风险闸门（你最需要理解的部分）

在真正执行任何变更前，Agent 会对操作打分：

| 决策 | 含义 | 是否需要你确认 |
|------|------|---------------|
| `AUTO` | 低风险、高置信、单资源（如只读查询、单资源低风险变更） | 不需要，自动执行 |
| `ASK` | 中高风险（如**破坏性**操作、缺元数据的操作） | **需要你显式确认** |
| `REFUSE` | 安全评分 = 0（硬地板） | 永不执行，直接拒绝 |

- **破坏性操作（删除 / 停止 / 释放 / 吊销…）永远走 `ASK` 或 `REFUSE`**，不会自动执行。
- 当走到 `ASK` 时，Agent 会停下等你确认（`{{user.confirm}}`）。**只有你确认后**，才带上授权来源继续执行——这个授权会被记入 trace 的 `confirmed_by` 字段，回答"谁批准了这次操作"。
- 任何 `Safety = 0` 的操作都被硬地板挡住，无论其它条件如何。

> 边界：incident-loop-agent 是"读 + 决策 + 轻量写"。真正的批量写入（如 `ve rds CreateDBInstance`）仍由目标 `ve-*-ops` 技能执行。

---

## 审计与可追溯

每次运行都会产出 trace：
- `audit-results/gcl-trace-<时间戳>.json`：运行时 GCL 轨迹，包含每轮迭代、Critic 评分、决策、以及**每次 `ve` 调用的 `RequestId`**。
- `audit-results/incident-trace-<ticket_id>-<ISO>.json`：事故轨迹（含 `ticket_id`、各 `ve_calls` 的 `request_id`）。

这些轨迹是事后追溯与合规审计的依据。CI 会通过 `vet check trace` 校验：缺 `RequestId` 或脱敏失败会直接让流水线变红。

---

## 前置条件

- 仓库 checkout 含全部 `ve-*-ops` 技能目录。
- `ve` CLI 已安装并认证（`VOLCENGINE_ACCESS_KEY` / `VOLCENGINE_SECRET_KEY`）。
- 可达 `docs/skill-routing-graph.md`（跨产品路由，含 §4 Predictive Triggers 预测触发规则）与 `docs/failure-patterns.md`（复盘写入）。
- GCL runner（`vet gcl run`）可用。

---

## 与其它文档的关系

- 想了解 `vet` 命令细节 → [vet-cli.md](vet-cli.md)
- 想看完整 GCL 规范（角色/评分/终止） → `docs/gcl-spec.md`
- 想看执行风险策略与 schema → `incident-loop-agent/references/policies/`、`incident-loop-agent/assets/`

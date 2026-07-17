# `vet` 命令行手册

`vet` 是 ve-skills 仓库的 Go 工具，提供**质量检查**与 **GCL 编排**两类能力。从仓库内构建：

```bash
cd cmd/vet
go build -o ../bin/vet .
# 或直接运行
go run . <子命令> --root <仓库根>
```

所有命令都接受 `--root <path>`（默认当前仓库根），绝大多数支持 `--json` 输出机器可读结果。

---

## 一、质量检查：`vet check <子项>`

| 子项 | 作用 | 失败含义 |
|------|------|---------|
| `frontmatter` | 校验每个 `ve-*-ops/SKILL.md` 的 frontmatter 元信息 | 缺字段/格式错误 |
| `links` | 校验技能文档间的 Markdown 链接可达 | 死链 |
| `gcl` | 校验 GCL 一致性（rubric / prompt 模板存在） | GCL 配置缺失 |
| `aiops` | 校验 AIOps/FinOps 覆盖（R+R 级技能需有 advanced 文档） | 覆盖缺口 |
| `assessment` | 校验文档示例 JSON 合法 | 示例损坏 |
| `eval` | 校验 `eval_queries.json` 意图分类用例充足（可加 `--git-diff origin/main` 仅查改动技能） | 用例不足 |
| `policyguard` | 校验 dispatch 计划满足安全不变量（Safety=0→不 AUTO、破坏性→不 AUTO、缺元数据→不 AUTO） | 安全策略将被破坏 |
| `trace` | **双 schema 校验**：`gcl-trace-*.json`（运行时轨迹，要求 `request_id` 非空且脱敏）与 `incident-trace-*.json`（事故轨迹，要求 `ticket_id`/`request_id` 等齐全） | 轨迹不可追溯或脱敏失败 |
| `routing` | 校验 `docs/skill-routing-graph.md` 中 predictive/reactive 路由行的结构完整性（symptom / primary / secondary / action / source） | 路由行结构不完整 |

示例：

```bash
vet check trace --root .          # P5：运行时轨迹缺 RequestId 会 exit 1
vet check policyguard --root .    # P7：安全不变量红线
vet check links --root .
```

> `vet check trace` 只会扫描 `audit-results/` 下 `gcl-trace-*` 与 `incident-trace-*` 前缀的文件，其它 JSON 会被跳过。

---

## 二、整体校验：`vet validate --root .`

一次性跑完上述适用的检查集合，等价于本地模拟 CI。提交前建议先跑一遍：

```bash
vet validate --root .
# 仅列出将要执行的检查
vet validate --root . --list-only
```

---

## 三、GCL 编排：`vet gcl <run|gate|trace>`

### `vet gcl run` — 真实运行一次 GCL 循环

让 Orchestrator 执行某个技能的一条生成器命令，跑 Critic 审计、重试、写 trace。

```bash
vet gcl run \
  --root . \
  --skill ve-ecs-ops \
  --request "客户报 ECS 实例无法访问" \
  --command "ve ecs DescribeInstances --InstanceIds i-xxxx" \
  --structural-critic-only        # 仅规则审计，不调用外部 Critic（CI/演练）
```

常用参数：

| 参数 | 说明 |
|------|------|
| `--skill` | 技能 id，如 `ve-ecs-ops`（必填） |
| `--request` | 脱敏后的用户请求（记入 trace，必填） |
| `--command` | 生成器要执行的 shell 命令（必填） |
| `--max-iter` | 最大循环轮次（0 = 该技能默认值） |
| `--timeout` | 单条命令超时秒数（默认 120） |
| `--structural-critic-only` | 用内置规则 Critic（不需要外部 Critic 输入） |
| `--critic-json <file>` | 外部 Critic 评分 JSON 文件 |
| `--critic-stdin` | 从 stdin 读取 Critic JSON |
| `--critic-command <cmd>` | 独立进程运行的隔离 Critic 命令 |
| `--confirmed` | 为 `ASK` 类操作背书（见下方安全说明） |
| `--confirmed-by <id>` | 背书的来源（工单号/人工 handle），写入 trace 留痕 |

**安全说明（重要）**：
- 非交互运行时，`ASK` 类操作若没有 `--confirmed` 会被当作 `REFUSE`（没人可问），并记录 `POLICY_BLOCK`。
- 只有在上游人工确认闸门（如 incident-loop-agent Step 5 的 `{{user.confirm}}`）显式授权后，才传 `--confirmed`，并**必须**同时传 `--confirmed-by` 记录"谁授权了"。裸用 `--confirmed` 而无 `--confirmed-by` 违反审计规则。
- 退出码：`0` PASS · `1` MAX_ITER · `2` Critic 无效 · `3` SAFETY_FAIL · `4` POLICY_BLOCK。

### `vet gcl gate` — CI 结构冒烟

对所有（或指定）技能跑结构性冒烟，确认 GCL 装配没坏。

```bash
vet gcl gate --root . --skip-incident-loop   # 跳过编排技能本身
vet gcl gate --root . --skills ve-ecs-ops,ve-vpc-ops
```

输出形如 `GCL CI gate: 28/29 skills pass structural smoke.`，任一失败则 exit 1。

### `vet gcl trace` — 聚合审计轨迹

把 `audit-results/gcl-trace-*.json` 汇总成质量概览（通过率、各维度评分、未决项）。

```bash
vet gcl trace --root .
vet gcl trace --root . --since-hours 24     # 只看最近 24 小时
vet gcl trace --root . --input audit-results/gcl-trace-20260717-*.json
```

---

## 六、预测式触发：`vet gcl predict`

在告警发生**之前**，根据指标时间序列趋势评估是否需要触发 GCL 循环。内置 3 个预测器：

| 预测器 | 指标 | 触发条件 |
|--------|------|---------|
| `redis-slow-query-degrade` | `slow_commands_per_sec` | 最近 5 样本上升 ≥50% 且当前值 >100 |
| `rds-capacity-waterline` | `disk_usage_percent` | 线性外推 24h 内磁盘使用率将达 ≥90% |
| `ecs-cpu-trend` | `cpu_usage_percent` | 1h 均值 >70% 且趋势斜率为正 |

使用方式：

```bash
# 从 JSON 文件或 stdin 读取指标
echo '{"skill":"ve-redis-ops","name":"slow_commands_per_sec","current":150,"history":[60,70,80,90,110]}' \
  | vet gcl predict --input -

# 从命令行参数指定
vet gcl predict \
  --skill ve-redis-ops \
  --metric slow_commands_per_sec \
  --current 150 \
  --history 60,70,80,90,110

# JSON 输出
vet gcl predict --input metrics.json --json
```

输出示例：
```
🚨 risk=high trigger=true predictor=redis-slow-query-degrade
  detail: slow_commands_per_sec up >=50% over last 5 samples and exceeds threshold 100
```

- `risk=low`：继续监控，不触发
- `risk=medium`：记录 HINT（供 Reflexion 消费，T14），不触发
- `risk=high`：触发 loop，`trigger=true`

> 预测器只消费喂入的 `Metric`，不自己拉取云监控数据。真实数据采集由 incident-loop-agent（T16）负责。

---

## 七、版本

```bash
vet version
```

---

## 八、典型工作流

```bash
# 1) 开发/改完技能后，本地跑质量门禁
vet validate --root .

# 2) 演练某技能的一次 GCL 运行（不真实变更）
vet gcl run --skill ve-ecs-ops --request "演练" \
  --command "ve ecs DescribeInstances" --structural-critic-only

# 3) 提交前，确认 CI 会跑的检查都绿
vet check trace --root .
vet check policyguard --root .
```

> CI（`.github/workflows/validate.yml`）已自动执行：`vet check trace`、`vet check policyguard`、`vet gcl gate` 等。本地结果与 CI 一致。

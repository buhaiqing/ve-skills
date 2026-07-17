# `vet` 命令行手册

`vet` 是 ve-skills 仓库的 Go 工具，提供以下能力：

| 类别 | 命令 | 用途 |
|------|------|------|
| **质量检查** | `vet check <子项>` | 提交前校验 frontmatter / 链接 / GCL / 安全策略 / trace |
| | `vet validate --root .` | 一次性跑完所有检查（等价 CI） |
| **GCL 编排** | `vet gcl run` | 运行 GCL 循环：Generator 执行 → Critic 审计 → 重试 |
| | `vet gcl gate` | CI 结构冒烟：确认所有技能的 GCL 装配完好 |
| | `vet gcl trace` | 聚合 audit-results/ 下的轨迹 JSON 成质量概览 |
| **自愈** | `vet gcl heal-stats` | 查看自愈引擎指标：成功率、耗时、干预率 |
| **预测触发** | `vet gcl predict` | 指标趋势评估：在告警发生前触发 loop |

从仓库内构建：

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

## 三、GCL 编排：`vet gcl <run|gate|trace|predict|heal-stats>`

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
| `--heal` | 自愈策略：`smart`（默认，错误分类重试 + 多路径自愈）、`none`（固定循环重试） |

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

## 四、自愈遥测：`vet gcl heal-stats`

查看自愈引擎的运行指标（成功率、平均耗时、人工干预率、降级率），支持 `--since 7d / 1w / 24h` 时间窗口过滤。

```bash
vet gcl heal-stats --since 7d
vet gcl heal-stats --since 7d --json
```

输出示例：
```
Self-healing stats (since 7d, events=42, skipped=3):
Success rate            94.3% (target: >= 90.0%) ✅
Avg duration             2.1s (target: <=  5.0s) ✅
User intervention        5.7% (target: <= 10.0%) ✅
```

**自愈引擎工作原理（T09–T10）：**

| 组件 | 文件 | 作用 |
|------|------|------|
| 错误分类 | `heal/retry.go` | 根据 `ve` CLI 真实输出信号将错误分为 `retryable`/`rate_limit`/`fatal`/`unknown` 四类 |
| 多路径自愈 | `heal/paths.go` + `heal/runner.go` | 每类错误 ≥2 条互不重叠的自愈路径，按 (Cost↑, 历史成功率↓) 自动选最优 |
| 指标收集 | `heal/metrics.go` | 按 per-path 粒度记录成功/失败/降级，累积历史供 SelectBest 消费 |

- `fatal` 类错误（权限/参数问题）不重试，直接上报
- `rate_limit` 类等令牌后重试 1 次
- 所有路径都失败时降级（fallback），记录 `result=fail`
- trace 中 `self_healing` 段记录每次自愈的 path_name / cost / result / duration

---

## 五、预测式触发：`vet gcl predict`

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

## 六、版本

```bash
vet version
```

---

## 七、典型工作流

### 场景 A：提交前质量门禁

```bash
# 1) 本地跑全量门禁（等价 CI 会跑的）
vet validate --root .

# 2) 只看关键安全红线
vet check trace --root .          # trace 脱敏 + RequestId 完整性
vet check policyguard --root .    # Safety=0 硬地板
vet check routing --root .        # 路由图结构
```

### 场景 B：演练 GCL 循环（不真实变更）

```bash
# 不带自愈的纯 GCL 演练
vet gcl run \
  --skill ve-ecs-ops \
  --request "演练：查询实例" \
  --command "ve ecs DescribeInstances" \
  --structural-critic-only

# 带智能自愈的演练（观察错误分类与路径选择）
vet gcl run \
  --skill ve-redis-ops \
  --request "演练：慢查询检测" \
  --command "ve redis DescribeSlowLogRecords --InstanceId xxxxx" \
  --heal smart \
  --structural-critic-only

# 查看自愈引擎近期指标
vet gcl heal-stats --since 7d
```

### 场景 C：预测式触发检查（先于告警）

```bash
# 检查 Redis 慢查询趋势
echo '{"skill":"ve-redis-ops","name":"slow_commands_per_sec","current":150,"history":[60,70,80,90,110]}' \
  | vet gcl predict --input -

# 检查 RDS 磁盘水位
vet gcl predict \
  --skill ve-rds-mysql-ops \
  --metric disk_usage_percent \
  --current 78 \
  --history 75,76,76,77,78,78

# 如果 risk=high 且 trigger=true，应启动 incident-loop-agent
```

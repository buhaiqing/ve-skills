# SDD: L4 T12 — 预测式触发源（先于告警）

> **源任务卡:** `docs/l3-to-l4-tasks/T12-predictive-trigger.md`
> **Plan:** `docs/superpowers/l3-to-l4/plans/2026-07-17-l4-t12-predictive-trigger.md`
> **专题:** `docs/superpowers/l3-to-l4/`
> **Status:** ready to implement (spec approved)

---

## 1. 功能描述

把 GCL 的触发模型从「告警/工单驱动（reactive）」扩展到「指标趋势预测驱动（predictive）」：

- 给定一组时序指标（容量水位、慢查询劣化、错误率上升），预测器在**告警发生之前**给出风险等级。
- 风险等级映射到动作：`low` → 仅监控；`medium` → 记入 HINT（供 Reflexion，T14 消费）；`high` → 触发 loop。
- 预测触发的症状通过 `docs/skill-routing-graph.md` §4（新增 Predictive Triggers 表）路由到主/次 skill，与既有 reactive 路由表并存。
- 路由表由新增的 JSON schema（`docs/skill-routing-graph.schema.json`）约束，并由 `vet check routing` 子命令校验，纳入 CI。

**Out of scope（本卡不做）：**
- 预测器**不主动调用任何 `ve` CLI 拉指标**——它只接收调用方喂入的 `Metric`（含 `History`）。真实指标采集由 T16 / incident-loop-agent 接入。
- 不实现「触发后自动执行 heal」——触发后按既有 loop 流程先 triage（T12 卡 §7 风险缓解）。
- 不扩展 `incident-loop-agent` 的运行时调度（只扩路由表与 schema，运行时会消费）。

---

## 2. 接口契约 / 状态机

### 2.1 `predict` 包公共 API

```go
package predict

// RiskLevel 是预测器输出的风险等级。
type RiskLevel string

const (
    RiskLow    RiskLevel = "low"    // 监控继续，不触发
    RiskMedium RiskLevel = "medium" // 记入 HINT，不触发 loop
    RiskHigh   RiskLevel = "high"   // 触发 loop
)

// Metric 是调用方喂入的单条时序指标。
type Metric struct {
    Skill   string    // 来源 skill，如 "ve-redis-ops"
    Name    string    // 指标名，如 "slow_commands_per_sec"
    Current float64   // 当前值
    History []float64 // 最近 N 次采样（按时间升序）
}

// Evaluation 是单次预测结果。
type Evaluation struct {
    Predictor string    // 预测器名
    Risk      RiskLevel // 风险等级
    Trigger   bool      // 是否应触发 loop（RiskHigh）
    Detail    string    // 人类可读的判定理由
}

// Predictor 是单一预测器接口。
type Predictor interface {
    Name() string
    Evaluate(ctx context.Context, m Metric) (Evaluation, bool) // 第二个返回值=false 表示"不适用此预测器"
}
```

### 2.2 三个内置预测器

| 预测器名 | 适用 Metric.Name | 触发条件（RiskHigh） | Medium 条件 |
|---------|----------------|---------------------|------------|
| `redis-slow-query-degrade` | `slow_commands_per_sec` | `len(History)>=5` 且 近 5 点上升 ≥50% 且 `Current > threshold`（默认 100） | 上升 ≥50% 但未超阈值 |
| `rds-capacity-waterline` | `disk_usage_percent` | 线性外推（最小二乘斜率）预测 24h 内 `Current + slope*24*每小时采样数` ≥ 90 | 外推 24h 内 ≥ 80 |
| `ecs-cpu-trend` | `cpu_usage_percent` | `len(History)>=5` 且 近 1h 均值 > 70 且 斜率为正 | 近 1h 均值 > 70 |

**历史数据不足保护：** 任一预测器若 `len(History) < 5`（或所需最小窗口），直接返回 `RiskLow`、不触发，避免误报抖动。

### 2.3 CLI 契约

**新增 `vet gcl predict` 子命令（`cmd/vet/gcl.go`）：**
```
vet gcl predict --input <path|->
vet gcl predict --skill <name> --metric <name> --current <float> --history <csv>
```
- `--input -` 读取一行 JSON（`Metric` 编码），输出 `Evaluation` JSON（含 Risk/Trigger/Detail）。
- 内置预测器全部注册到 `Registry`，依次尝试；命中第一个 `applicable` 的预测器即返回其 `Evaluation`。

**新增 `vet check routing` 子命令（`cmd/vet/check.go`）：**
- 读取 `docs/skill-routing-graph.md`，解析所有 `| ... | ... | ... |` 表格行，对照 `docs/skill-routing-graph.schema.json` 校验。
- DoD 要求至少校验每条 trigger 行的必填字段：`symptom`、`primary`、`secondary[]`、`action`、`source`（枚举 `predictive`/`reactive`）。
- 通过 `--json` 输出机器可读报告；失败退出码 1。

---

## 3. 异常边界 / 状态机

| 场景 | 行为 |
|------|------|
| `History` 长度 < 最小窗口 | `RiskLow`，不触发（降噪） |
| `Current` 或 `History` 含 NaN / 非有限值 | 该预测器返回 `RiskLow`，不触发 |
| 无预测器适用该 `Metric.Name` | 输出 `RiskLow`，`Detail="no predictor matches metric name"` |
| `ctx` 取消 | 立即返回 `ctx.Err()` 包装的 error，不触发 |
| `--input` JSON 解析失败 | 退出码 2，打印 stderr 用法 |
| `secondary` 字段为空 | schema 校验允许（`secondary` 可为空数组）；reactive 表已有 precedent |

---

## 3.5 实现锚点（磁盘目标，均为新增）

| 交付物 | 目标路径 | 当前状态 |
|--------|----------|----------|
| `predict` 包（接口 + 3 预测器 + Registry） | `cmd/vet/internal/gcl/predict/predict.go`（新） | 不存在，需新建 |
| `predict` 单测 | `cmd/vet/internal/gcl/predict/predict_test.go`（新） | 不存在，需新建 |
| `vet gcl predict` 子命令 | `cmd/vet/gcl.go`（在 `runGCL` 的 `case` 链新增 `"predict"`） | 当前仅 run/gate/trace/heal-stats，需新增 |
| 路由图 schema | `docs/skill-routing-graph.schema.json`（新） | 不存在，需新建 |
| 路由图 §4 Predictive Triggers | `docs/skill-routing-graph.md`（在 §3 Critical Routing Rules 后插入 §4） | 当前仅到 §3，需新增 |
| `vet check routing` 子命令 | `cmd/vet/check.go`（在 `runCheck` 的 `switch` 新增 `"routing"`） | 当前无 routing case，需新增 |
| CI 接线 | `.github/workflows/validate.yml`（新增 `vet check routing --root .`） | 需新增 step |

> 所有交付物均为新增，与 T12 卡 DoD 一致。不修改既有 `vet gcl` / `vet check` 子命令行为，只在 `case` 链追加分支。

---

## 4. 验收标准（DoD 映射）

```
✅ 1. cmd/vet/internal/gcl/predict/predict.go：Predictor 接口 + 3 内置实现（redis/rds/ecs）
✅ 2. cmd/vet/gcl.go 注册 vet gcl predict 子命令
✅ 3. 写入 docs/skill-routing-graph.schema.json
✅ 4. docs/skill-routing-graph.md 新增 §4 Predictive Triggers（≥3 行，含 source=predictive）
✅ 5. cmd/vet/check.go 新增 routing 子命令，读取 schema 校验
✅ 6. go build + go vet + go test 全绿
✅ 7. predict_test.go 覆盖：3 个预测器的触发/不触发边界（含历史不足、阈值边界）
✅ 8. CI（validate.yml）新增 vet check routing（可阻断或 || true 仅警告，见 plan §M5）
```

---

## 5. 风险 / 偏差

- **偏差（与源卡 §3.1 字段）：** 源卡 `Predictor.Evaluate(ctx, metric) (Risk, bool)` 返回裸 `Risk` 字符串；本 spec 改为返回结构化 `Evaluation`（含 `Predictor`/`Risk`/`Trigger`/`Detail`），更利于日志与 HINT 消费，且 `bool` 语义明确为「是否适用」。
- **误报抖动：** 触发后不自动执行，先 triage（与 T12 卡 §7 一致）。
- **预测器不拉真实指标：** 仅消费喂入数据，保持包纯净、可单测（无外部依赖）。
- **`vet check routing` 解析策略：** 用正则从 markdown 表格提取行，不引入 markdown 解析库；schema 用标准 `encoding/json` + 简单手写校验（不引入 jsonschema 依赖），保持零三方依赖。

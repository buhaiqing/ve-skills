# Plan: L4 T12 — 预测式触发源（先于告警）

> **源任务卡:** `docs/l3-to-l4-tasks/T12-predictive-trigger.md`
> **Parent spec:** `docs/superpowers/l3-to-l4/specs/2026-07-17-l4-t12-predictive-trigger-design.md`
> **专题:** `docs/superpowers/l3-to-l4/`
> **Status:** ✅ DONE (2026-07-17) — 实现完成，go build/vet/test 全绿，vet check routing 通过。

---

## 0. Scope

把 GCL 触发模型从「告警/工单驱动」扩到「指标趋势预测驱动」。新增 `cmd/vet/internal/gcl/predict` 包（接口 + 3 内置预测器 + Registry），注册 `vet gcl predict` 与 `vet check routing` 两个子命令，并在 `docs/skill-routing-graph.md` 新增 §4 Predictive Triggers + 配套 JSON schema 约束，CI 接线 `vet check routing`。

**实现注记（与源卡偏差）：** 源卡 `Predictor.Evaluate(ctx, metric) (Risk, bool)` 返回裸 `Risk`；本 plan 按 spec §2.1 改为返回结构化 `Evaluation`（含 `Predictor`/`Risk`/`Trigger`/`Detail`），`bool` 语义为「是否适用」。预测器不拉真实指标（仅消费喂入 `Metric`，spec §1 Out of scope）。`vet check routing` 用正则从 markdown 表格提取行 + 手写 schema 校验，不引入 markdown / jsonschema 第三方依赖（spec §5）。

**Out of scope:** 触发后自动执行 heal（先 triage）；真实指标采集（T16 / incident-loop-agent 接入）；扩展 incident-loop-agent 运行时调度。

---

## 1. 里程碑拆分

### M1 — 新增 `predict` 包（`cmd/vet/internal/gcl/predict/predict.go`）
**File:** `cmd/vet/internal/gcl/predict/predict.go`（新）

- 定义 `RiskLevel`（`RiskLow`/`RiskMedium`/`RiskHigh`）、`Metric`、`Evaluation`、`Predictor` 接口（spec §2.1）。
- `Registry`：`[]Predictor`，`Evaluate(ctx, m) (Evaluation, bool)` 依次尝试，命中第一个 `applicable` 预测器即返回。
- 3 个内置预测器（spec §2.2）：
  - `redisSlowQueryDegrade`：`slow_commands_per_sec`，近 5 点上升 ≥50% 且 `Current > 100`（默认阈值）→ RiskHigh；上升 ≥50% 但未超阈值 → RiskMedium。
  - `rdsCapacityWaterline`：`disk_usage_percent`，最小二乘斜率外推 24h 内 ≥90 → RiskHigh；≥80 → RiskMedium。
  - `ecsCPUTrend`：`cpu_usage_percent`，近 1h 均值 >70 且斜率为正 → RiskHigh；均值 >70 → RiskMedium。
- 历史数据不足（`len(History) < 5`）/ NaN / 非有限值 → `RiskLow`，不触发（spec §3）。

### M2 — `vet gcl predict` 子命令（`cmd/vet/gcl.go`）
**File:** `cmd/vet/gcl.go`（在 `runGCL` 的 `case` 链新增 `"predict"`，`runGCLPredict` 函数）

- `--input <path|->`：读一行 `Metric` JSON，输出 `Evaluation` JSON。
- `--skill/--metric/--current/--history <csv>`：CLI 直接喂入。
- 调用 `Registry.Evaluate`，输出 `Evaluation`；`Trigger=true` 时退出码 0 但 `Detail` 提示触发。

### M3 — 路由图 schema + §4（`docs/`）
**File:** `docs/skill-routing-graph.schema.json`（新）+ `docs/skill-routing-graph.md`（新增 §4）

- schema 约束每行 trigger 必含 `symptom`/`primary`/`secondary[]`/`action`/`source`（枚举 `predictive`/`reactive`）。
- §4 Predictive Triggers 表（spec §2.3 + 卡 §3.3）：Redis 慢查询 5min +50%、RDS 磁盘 24h 内将满、ECS CPU 1h 趋势上，各含 primary/secondary/action/source=predictive。

### M4 — `vet check routing` 子命令（`cmd/vet/check.go`）
**File:** `cmd/vet/check.go`（在 `runCheck` 的 `switch` 新增 `"routing"`，`runCheckRouting` 函数）

- 正则解析 `docs/skill-routing-graph.md` 所有 `| ... | ... | ... |` 表格行，对照 schema 校验必填字段 + `source` 枚举。
- `--json` 输出机器可读报告；失败退出码 1。

### M5 — 单测 + CI（`predict_test.go` + `validate.yml`）
**File:** `cmd/vet/internal/gcl/predict/predict_test.go`（新）+ `.github/workflows/validate.yml`

- `predict_test.go` 覆盖：3 个预测器触发/不触发边界（含历史不足、阈值边界、NaN）。
- CI 新增 `vet check routing --root .`（`|| true` 仅警告，不阻断，见卡 DoD #8）。

### M6 — 翻卡 + 提交
**File:** `docs/l3-to-l4-tasks/T12-predictive-trigger.md`

- 状态 🟡 TODO → ✅ DONE，追加完成回报（含偏差说明：结构化 `Evaluation`、`secondary` 可为空）。
- commit spec+plan+代码到 `docs/superpowers/l3-to-l4/` + `cmd/vet/` + `docs/`。

---

## 2. 函数级 / 命令契约

| Spec 组件 | Code anchor（权威） |
|-----------|---------------------|
| `Predictor` 接口 | `cmd/vet/internal/gcl/predict/predict.go` |
| `Registry.Evaluate` | `cmd/vet/internal/gcl/predict/predict.go` |
| 3 内置预测器 | `cmd/vet/internal/gcl/predict/predict.go` |
| `vet gcl predict` | `cmd/vet/gcl.go` `runGCLPredict`（新增 `case "predict"`） |
| 路由 schema | `docs/skill-routing-graph.schema.json`（新） |
| 路由 §4 | `docs/skill-routing-graph.md`（新增 `## §4 Predictive Triggers`） |
| `vet check routing` | `cmd/vet/check.go` `runCheckRouting`（新增 `case "routing"`） |

---

## 3. DoD

```
✅ 1. cmd/vet/internal/gcl/predict/predict.go：Predictor 接口 + 3 内置实现（redis/rds/ecs）
✅ 2. cmd/vet/gcl.go 注册 vet gcl predict 子命令
✅ 3. 写入 docs/skill-routing-graph.schema.json
✅ 4. docs/skill-routing-graph.md 新增 §4 Predictive Triggers（≥3 行，含 source=predictive）
✅ 5. cmd/vet/check.go 新增 routing 子命令，读取 schema 校验
✅ 6. go build + go vet + go test 全绿
✅ 7. predict_test.go 覆盖：3 个预测器的触发/不触发边界（含历史不足、阈值边界）
✅ 8. CI（validate.yml）新增 vet check routing（|| true 仅警告）
✅ 9. T12 卡状态 ✅ DONE + 偏差说明
```

---

## 4. 验证

```bash
cd cmd/vet
go build ./... && go vet ./...
go test -run TestPredict ./internal/gcl/predict/ -v
go test ./...
go build -o /tmp/vet .
/tmp/vet gcl predict --help
echo '{"skill":"ve-redis-ops","name":"slow_commands_per_sec","current":150,"history":[100,80,70,60,50]}' \
  | /tmp/vet gcl predict --input -
# 期望：risk=high, trigger=true
/tmp/vet check routing --root .     # 校验 docs/skill-routing-graph.md
```

---

## 5. 三者一致性锚点

- **Spec ↔ Plan:** DoD（§3）== spec §4 验收标准。✅ 对齐。
- **Plan ↔ Code:** anchors（§2）指向真实目标路径；M1–M5 为新增文件/分支，非凭空符号（已 grep 确认 `vet gcl` 当前无 predict case、`vet check` 无 routing case、predict 包与 schema 均不存在）。✅ 对齐。
- **Spec ↔ Code:** 实现以 spec §2.1 结构化 `Evaluation` 为准，不退回源卡裸 `Risk`（spec §5 偏差）。✅ 对齐。
- **任务卡 ↔ Plan:** `docs/l3-to-l4-tasks/T12` 为 plan 源，本 plan 是其 Superpowers 规范拆分。✅ 引用一致。

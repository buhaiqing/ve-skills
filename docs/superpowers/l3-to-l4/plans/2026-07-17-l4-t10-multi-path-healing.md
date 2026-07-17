# Plan: L4 T10 — 多路径自愈（auto-select best）

> **源任务卡:** `docs/l3-to-l4-tasks/T10-multi-path-healing.md`
> **Parent spec:** `docs/superpowers/l3-to-l4/specs/2026-07-17-l4-t10-multi-path-healing-design.md`
> **专题:** `docs/superpowers/l3-to-l4/`
> **Status:** ready to implement (spec approved, path 语义已与用户确认=策略包装器).

---

## 0. Scope

把 T09 单条智能重试扩为多路径自动选择。新增 `heal/paths.go`（注册表 + SelectBest）与
`heal/runner.go`（Classify→Select→Execute→Verify 闭环）；扩展 T11 `Metrics` 加 per-path 历史；
`trace` 加 `self_healing` 段。Path.Execute 为策略包装器（用户确认，非 infra 动作）。

**Out of scope:** 具体基础设施自愈动作（换镜像/提权/重启）—— `heal` 包不执行，由调用方注入 env/策略。

---

## 1. Milestones

### M1 — `heal/paths.go`（注册表 + SelectBest）
- `Path{Name, Class, Cost, Execute func(ctx, op func() error) error}`。
- `Paths` 注册表：4 真实类 × ≥2 路径（≥8 条，见 spec §2.1）。
  - retryable: `backoff-retry`(Cost1) / `endpoint-switch-retry`(Cost2)
  - rate_limit: `wait-retry`(Cost1) / `wait-retry-long`(Cost2)
  - fatal: `escalate`(Cost4) / `degrade-manual`(Cost3)
  - unknown: `single-retry`(Cost1) / `escalate`(Cost4)
- `SelectBest(class ErrorClass, hist History) *Path`：按 (Cost↑, SuccessRate↓) 选；空历史→Cost 最低；无路径→nil。
- `History` 接口：`SuccessRate(class, name string) float64`（由 Metrics 实现）。

### M2 — `Metrics.PerPath`（消费 T11）
**文件:** `metrics.go`
- `PathStat{Success int64; Total int64}`；`Metrics.PerPath map[string]*PathStat`（key=`class/name`）。
- `Record` 时对 `ErrorCode`(class)+`Action`(name) 累加 PerPath。
- `SuccessRate(class, name)` 实现 `History` 接口（分母 0 → 0.0）。

### M3 — `heal/runner.go`（闭环）
- `Run(ctx, class ErrorClass, op func() error, hist History) (PathResult, error)`
  - `p := SelectBest(class, hist)`；nil → 退化为单次 `op()`（T09 行为）。
  - 执行 `p.Execute(ctx, op)`；成功 → `PathResult{Name, Cost, Result:"ok", DurationMs}`。
  - 失败 → 尝试同 class 其余路径（按 Cost 升序）；全失败 → 返回末错 + `result=fail`。
- `PathResult{Name string; Cost int; Result string; DurationMs int64}`。
- ctx 默认 30s 上限（每条路径 `context.WithTimeout`）。

### M4 — trace `self_healing` 段
**文件:** `trace.go` `Final` 加 `SelfHealing *SelfHealingRecord json:"self_healing,omitempty"`
```go
type SelfHealingRecord struct {
    Class     string `json:"class"`
    PathName  string `json:"path_name"`
    Cost      int    `json:"cost"`
    Result    string `json:"result"` // ok|fail
    DurationMs int64 `json:"duration_ms"`
}
```
`run.go` 在 `runGeneratorWithHeal` 中：当 `Heal=="smart"` 且走多路径时，用 `heal.Run` 结果填 `SelfHealing`，挂到本次迭代（或 Final）。为简单且兼容 T07，`self_healing` 放到 `Iteration`（每次自愈尝试一条）。

### M5 — 单测
**文件:** `paths_test.go` / `runner_test.go`
- `TestSelectBest_*`：每类选到预期最优（Cost 最低 / 历史高优先）。
- `TestRunner_*`：选最优执行成功；全失败→返回错误 + result=fail；fatal 类不重试直接 escalate；nil 路径退化单次 op。

### M6 — CI + 标记 DoD
**文件:** 复用现有 CI；`run.go` 多路径接入可行为对比。
- `docs/l3-to-l4-tasks/T10` 状态 🟡 TODO → ✅ DONE（追加完成回报 + 语义偏差说明）。

---

## 2. 函数级 / 命令契约

| Spec 组件 | Code anchor（权威） |
|-----------|---------------------|
| 路径注册表 | `cmd/vet/internal/gcl/heal/paths.go` |
| 闭环 | `cmd/vet/internal/gcl/heal/runner.go` |
| 历史接口 | `metrics.go` `History` / `Metrics.SuccessRate` |
| trace 段 | `trace.go` `Final.SelfHealing` / `Iteration` |
| 接入点 | `run.go` `runGeneratorWithHeal` |

---

## 3. DoD

```
✅ 1. heal/paths.go：4 真实类 × ≥2 路径（≥8 条）+ SelectBest 加权
✅ 2. heal/runner.go：Classify→Select→Execute→Verify 闭环
✅ 3. SelectBest 按 (Cost, T11 历史成功率) 加权；历史空退 Cost 最低
✅ 4. go build + go vet + go test 全绿
✅ 5. runner_test.go：每类选最优 + 全失败降级/上报
✅ 6. trace self_healing 段必填（omitempty，兼容 T07 schema）
✅ 7. T10 卡 ✅ DONE + 语义偏差说明
```

---

## 4. 验证

```bash
cd cmd/vet
go build ./... && go vet ./...
go test -run 'TestSelectBest|TestRunner' ./internal/gcl/heal/ -v
go test ./...
go build -o /tmp/vet .
/tmp/vet gcl run --skill ve-ecs-ops --request x --command "..." --structural-critic-only --heal smart --max-iter 3
# 检查产出的 trace 含 self_healing 段
```

---

## 5. 三者一致性锚点

- **Spec ↔ Plan:** DoD（§3）== spec §5 验收标准。✅ 对齐。
- **Plan ↔ Code:** anchors（§2）指向真实符号（Paths/SelectBest/Metrics.SuccessRate）。✅ 对齐。
- **Spec ↔ Code:** 仅 4 真实类（spec §2.1），不混入 framework 安装码；Path 为策略包装器（spec §1）。✅ 对齐。
- **任务卡 ↔ Plan:** T10 卡为 plan 源；本 plan 是其 Superpowers 规范拆分 + 语义偏差记录。✅ 引用一致。

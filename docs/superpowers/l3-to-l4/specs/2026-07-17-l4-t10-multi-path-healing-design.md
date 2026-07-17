# Design: L4 T10 — 多路径自愈（auto-select best）

**Date:** 2026-07-17
**Author:** ve-skills (loop engineering)
**Status:** Approved
**专题目录:** `docs/superpowers/l3-to-l4/`
**Companion plan:** `docs/superpowers/l3-to-l4/plans/2026-07-17-l4-t10-multi-path-healing.md`
**源任务卡:** `docs/l3-to-l4-tasks/T10-multi-path-healing.md`（plan/DoD 来源，本文件为其 SDD 补充）
**依赖:** T09（heal.SmartRetry / Classify，已 ✅）；T11（Metrics 历史成功率，已 ✅）
**可并行:** 无

---

## 1. 功能描述

把 T09「单条智能重试」扩成「多条候选自愈路径，按 (Cost, 历史成功率) 自动选最优」。
`heal` 包拥有**路径注册表** + `SelectBest`，每条 `Path` 是一个围绕 generator 命令的
**自愈策略包装器** `Execute func(ctx context.Context, op func() error) error`；具体命令变换
（如切换 endpoint 环境变量）由调用方注入，包本身不执行任何基础设施动作。

**用户确认的方向（2026-07-17）：** Path.Execute 采用「策略包装器」语义，而非源卡字面的
「执行具体自愈动作（switch-mirror/prompt-sudo/restart）」—— 因为 `vet gcl run` 是 GCL
编排器，不具备也不能越权执行基础设施自愈。

---

## 2. 路径模型（关键决策）

### 2.1 真实错误类别（非 framework 安装码）

与 T09/T11 一致，本包只有 4 个真实 `ErrorClass`：`retryable` / `rate_limit` / `fatal` / `unknown`。
源卡 §3.3 的 NET_TIMEOUT/PERM_*/GO_*/RES_* 是 framework 安装码，**不采用**。改为按 4 真实类
每类 ≥ 2 条策略路径：

| ErrorClass | 路径 1（低 Cost） | 路径 2 | 说明 |
|------------|-------------------|--------|------|
| `retryable` | `backoff-retry` | `endpoint-switch-retry` | 退避重试 / 切换 endpoint 环境变量后重试 |
| `rate_limit` | `wait-retry` | `wait-retry-long` | 短等重试 / 长等重试（更保守） |
| `fatal` | `escalate` | `degrade-manual` | 升级人工 / 降级手动（均不自动重试） |
| `unknown` | `single-retry` | `escalate` | 重试 1 次 / 升级 |

> 每条路径的 `Execute` 是策略包装器：例如 `endpoint-switch-retry` 在重试前给 `op` 闭包注入
> 一个不同的 endpoint env（由调用方提供 env 映射），`heal` 包不感知具体 env 值。

### 2.2 Path 结构

```go
type Path struct {
    Name    string
    Class   ErrorClass
    Cost    int            // 0=最低（如退避重试）, 5=高（如升级人工）
    Execute func(ctx context.Context, op func() error) error
}
```

### 2.3 SelectBest 加权

`SelectBest(class, hist History)` 返回该 class 下最优 `Path`：
- 主排序：`Cost` 升序（成本低优先）
- 次排序：`hist.SuccessRate(class, name)` 降序（历史成功率高优先）
- 历史为空 → 退到 Cost 最低（源卡 §7 风险缓解）
- `History` 由 T11 `Metrics` 提供（新增 per-class 计数，见 §3）

---

## 3. 历史成功率来源（消费 T11）

T11 `Metrics` 当前只有整体聚合。T10 需要**按 (class, path) 维度**的历史成功率。方案：
- `Metrics` 新增 `PerPath map[string]*PathStat`（`PathStat{Success, Total}`），`Record` 时按
  `ErrorCode`(class 名) + `Action`(path 名) 累加。
- `History.SuccessRate(class, name)` 读 `Metrics.PerPath[class+"/"+name]`。
- 这样 T11 的 `heal-stats` 自然也能按路径展示成功率（T10 不破坏 T11 既有聚合）。

---

## 4. 异常边界

- **路径超时**：每条 `Execute` 内部用 `ctx` + 默认 30s 上限；超时 → 下一条路径（源卡 §7）。
- **全路径失败**：所有候选路径失败 → 返回最后一个错误，runner 标记 `result=fail` 并上报（不静默吞）。
- **fatal 类不自动重试**：`escalate` / `degrade-manual` 立即返回（不进入重试循环），符合 T09 语义。
- **空注册表**：某 class 无路径 → `SelectBest` 返回 nil，runner 退化为 T09 单次重试。

---

## 5. 验收标准

- `heal/paths.go`：4 真实类 × ≥2 路径 = ≥8 条 `Path`；`SelectBest(class, hist)` 正确加权。
- `heal/runner.go`：`Run(ctx, class, op, hist)` 完成 Classify→Select→Execute→Verify 闭环；
  全失败时返回错误并标记。
- `Metrics.PerPath` 支持 T11 历史成功率查询。
- `trace` 新增 `self_healing` 段（path_name / cost / result / duration），与 T07 schema 兼容（omitempty）。
- `go build` + `go vet` + `go test` 全绿（runner_test 覆盖每类选最优 + 全失败降级）。
- 不引入 framework 安装错误码；不执行任何基础设施动作。

---

## 6. 风险与回滚

| 风险 | 缓解 |
|------|------|
| 路径副作用（如 endpoint 切换踩坑） | 每条路径 ctx timeout（默认 30s）；超时走下一条 |
| SelectBest 退化 | 历史空 → Cost 最低；T11 累积后切真实权重 |
| 回滚 | `git checkout cmd/vet/internal/gcl/heal/` |

---

## 7. 参考

- `cmd/vet/internal/gcl/heal/retry.go` — `ErrorClass` / `Classify` / `SmartRetry`（T09）
- `cmd/vet/internal/gcl/heal/metrics.go` — `Metrics`（T11，本卡扩展 PerPath）
- `cmd/vet/internal/gcl/trace/trace.go` — `Final` / `Iteration`（self_healing 段落点）
- `docs/l3-to-l4-tasks/T10-multi-path-healing.md` — 源任务卡

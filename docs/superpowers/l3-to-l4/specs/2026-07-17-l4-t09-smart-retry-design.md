# Design: L4 T09 — L2 智能重试（错误分类驱动）

**Date:** 2026-07-17
**Author:** ve-skills (loop engineering)
**Status:** Approved
**专题目录:** `docs/superpowers/l3-to-l4/` （L3→L4 大专题需求目录）
**Companion plan:** `docs/superpowers/l3-to-l4/plans/2026-07-17-l4-t09-smart-retry.md`
**源任务卡:** `docs/l3-to-l4-tasks/T09-smart-retry.md` （plan/DoD 来源，本文件为其 SDD 补充）
**依赖:** T06 (`vet gcl run` — L2→L3，已 ✅)
**可并行:** T12

---

## 1. 功能描述

把 `vet gcl run` 现有的"固定次数重试"（L1，见 `run.go` 的 `MaxIter` 循环）升级为**基于错误分类的针对性重试**（L2）：

- **网络/运行时类错误** → 指数退避重试（+ 可选切换 endpoint/镜像）
- **限流类错误** → 等待令牌/Retry-After 后重试
- **权限/参数类错误** → **不重试**，直接上报（重试无意义且可能放大故障）

产出：`cmd/vet/internal/gcl/heal/retry.go`（新 Go 包 `heal`），被 `vet gcl run` 的 retry 路径调用。

---

## 2. 错误分类契约（状态机）

```
                  ┌─────────────┐
   error_code ──▶ │  Classify   │
                  └──────┬──────┘
        ┌────────────────┼─────────────────┐
        ▼                ▼                 ▼
  ClassRetryable    ClassRateLimit     ClassFatal
  (退避重试)         (等令牌重试)        (不重试, 上报)
        │                │                 │
        ▼                ▼                 ▼
  SmartRetry 循环   SmartRetry 循环   立即返回错误
```

### 2.1 分类信号源（关键决策 — 与源任务卡偏差）

> **偏差说明（重要）：** 源任务卡 `T09-smart-retry.md` §3.2 的错误码映射表混入了两类本包**不会触发**的码：
> - `RequestLimitExceeded` / `Throttling` / `InvalidParameter` 在 `enhanced-self-healing-framework.md` 中 **0 次出现**；
> - `NET_*` / `PERM_*` / `GO_*` 是 **self-healing 安装器**的错误码（另一个工具链），不在 `vet gcl run` 的 retry 路径内。
>
> 本 SDD 以 **`vet gcl run` 实际可提取的 `ve` CLI 错误信号** 为准，直接复用 `run.go` 已有的 `failureSignatures` 正则分类（五类：`cli_parameter` / `runtime` / `cross_skill` / `token_efficiency` / `skill_generation`）。理由：真实可触发、零重复、符合简单优先。framework 安装码不纳入本包。

### 2.2 分类映射（基于 `run.go` `failureSignatures`）

| `failureSignatures` 类 | 代表错误信号（正则片段） | `ErrorClass` | 重试策略 |
|------------------------|--------------------------|--------------|----------|
| `runtime` | `TIMEOUT` `RequestLimitExceeded` `InternalError` `ConnectionError` | `ClassRetryable` | 指数退避 200ms→1.6s，max=3 |
| `runtime`（限流子类） | `RequestLimitExceeded` | `ClassRateLimit` | 等待 Retry-After（默认 1s）+ 重试 1 次 |
| `cli_parameter` | `InvalidParameter` `MissingParameter` `AuthFailure` `UnauthorizedOperation` | `ClassFatal` | **不重试**，直接上报 |
| `cross_skill` / `token_efficiency` / `skill_generation` | 跨技能/ token / 生成类 | `ClassFatal` | **不重试**（需人工或上游干预） |
| 未匹配 / 空 | — | `ClassUnknown` | 重试 1 次 + 标记（fail-safe） |

> 限流判定：`runtime` 类中若错误串含 `RequestLimitExceeded` 则降为 `ClassRateLimit`（更温和的等待策略）；其余 runtime 归 `ClassRetryable`。

---

## 3. 异常边界

- **退避上限**：`BackoffMax` 封顶（默认 1.6s），避免长尾阻塞 runtime。
- **重试计数**：`MaxAttempts` 默认 3；超过即终止并返回最后一次错误（trace 记 `MAX_ITER` 语义由 `vet gcl run` 外层循环承接）。
- **上下文取消**：`SmartRetry(ctx, ...)` 尊重 `ctx.Done()`，外部超时（run.go `--timeout`）可中断重试。
- **分类未知**：`ClassUnknown` 只重试 1 次并打指标标记，不无限重试。
- **不静默吞错**：Fatal 类立即返回原错误，`vet gcl run` 据此进入 SAFETY/上报分支。

---

## 4. 验收标准

- `heal.Classify(errorCode string) ErrorClass` 覆盖五类信号 + Unknown，行为可单测。
- `heal.SmartRetry`：Retryable 退避重试到 max；RateLimit 等待后重试 1 次；Fatal 立即返回不重试；Unknown 重试 1 次。
- 每次重试决策写入指标（供 T11 telemetry 消费）：分类、attempt、最终 outcome。
- `go build` + `go vet` + `go test ./...` 全绿（新增 `heal/retry_test.go`）。
- `vet gcl run` 仍可通过 `--heal=none`（或保持现状固定循环）与之对比，CI 加 1 case。
- 不引入 framework 安装错误码（NET_*/PERM_*/GO_*）到本包。

---

## 5. 风险与回滚

| 风险 | 缓解 |
|------|------|
| `ve` CLI 新增错误信号未覆盖 | `Classify` 默认 `ClassUnknown`（重试 1 次 + 标记）；月度 audit 补充正则 |
| 重试放大下游压力 | `BackoffMax` + `MaxAttempts` 封顶；指标可观测 |
| 回滚 | `git checkout cmd/vet/internal/gcl/heal/` |

---

## 6. 参考

- `cmd/vet/internal/gcl/run/run.go` — `failureSignatures`（L36-45）、`Run()` 循环（L460）
- `docs/l3-to-l4-tasks/T09-smart-retry.md` — 源任务卡（plan/DoD）
- `docs/l3-to-l4-tasks/01-index.md` — L4 依赖图（T09 为关键路径起点）
- `ve-skill-generator/references/enhanced-self-healing-framework.md` — 自愈框架（安装器错误码，本包**不**采用）

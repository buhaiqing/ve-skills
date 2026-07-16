# Plan: L4 T09 — L2 智能重试（错误分类驱动）

> **源任务卡:** `docs/l3-to-l4-tasks/T09-smart-retry.md`
> **Parent spec:** `docs/superpowers/l3-to-l4/specs/2026-07-17-l4-t09-smart-retry-design.md`
> **专题:** `docs/superpowers/l3-to-l4/` （L3→L4 需求目录）
> **Status:** ready to implement (spec approved).

---

## 0. Scope

把 `vet gcl run` 的固定次数重试升级为错误分类驱动的智能重试。新增 `cmd/vet/internal/gcl/heal` 包（`retry.go` + `retry_test.go`），被 `Run()` 的 retry 路径调用。

**实现注记（与源卡偏差）：** 错误分类信号以 `ve` CLI 真实输出为准，复用 `run.go` 的 `failureSignatures`（五类正则），**不引入** framework 的安装错误码（NET_*/PERM_*/GO_*）。详见 spec §2.2。

**Out of scope:** T10 多路径自愈、T11 遥测落库（本卡仅暴露指标钩子供其消费）；切换镜像/endpoint 的真实逻辑（保留钩子，本卡只做分类+退避决策）。

---

## 1. Milestones

### M1 — 新增 `heal` 包：`Classify` + `RetryPolicy`（`cmd/vet/internal/gcl/heal/retry.go`）
**File:** `cmd/vet/internal/gcl/heal/retry.go`（新）

- 定义 `ErrorClass`（Retryable / RateLimit / Fatal / Unknown）。
- `Classify(errorCode string) ErrorClass`：基于 `run.go` 的 `failureSignatures` 五类正则判定；`RequestLimitExceeded` 子类降为 RateLimit；未匹配 → Unknown。
- `RetryPolicy` 结构体：`MaxAttempts` / `BackoffBase` / `BackoffMax` / `Jitter`。
- 默认策略常量（Retryable: 200ms→1.6s max3；RateLimit: 等待 1s 重试 1 次；Fatal: 0；Unknown: 重试 1 次）。

### M2 — `SmartRetry` 实现（`retry.go`）
**File:** `cmd/vet/internal/gcl/heal/retry.go`

- `SmartRetry(ctx context.Context, op func() error, policy RetryPolicy, classify func(string) ErrorClass, record func(attempt int, class ErrorClass, outcome string)) error`
  - 先执行 `op()`；成功返回 nil。
  - 失败 → 用 `classify(errorCode)` 取分类（errorCode 由调用方从 `op` 错误中提取；本包提供 `ExtractErrorCode` 辅助，从 run.go 的 `failureSignatures` 反查）。
  - Retryable/Unknown：退避后重试至 `MaxAttempts`。
  - RateLimit：等待后重试 1 次。
  - Fatal：立即返回原错误，不重试。
  - 每次决策通过 `record` 钩子写指标（供 T11）。
  - 尊重 `ctx.Done()`。

### M3 — 接入 `vet gcl run`（`cmd/vet/internal/gcl/run/run.go`）
**File:** `run.go` `Run()`

- 在 `runCommand` 失败路径（或迭代内失败分支）调用 `heal.SmartRetry` 替代"无条件进下一 iter"。
- 新增 `Options.Heal` 字段（`"smart"` 默认 / `"none"` 保持现状固定循环），`gcl.go` 暴露 `--heal` flag。
- 分类结果写入 `trace.Iteration`（新增 `heal_class` 可选字段，供审计）。

### M4 — 单测（`cmd/vet/internal/gcl/heal/retry_test.go`）
**File:** `retry_test.go`（新）

- `TestClassify_*`：五类信号 + Unknown 分类正确。
- `TestSmartRetry_Retryable`：退避重试到 max 后返回错误。
- `TestSmartRetry_Fatal`：立即返回，0 次重试。
- `TestSmartRetry_RateLimit`：等待后重试 1 次。
- `TestSmartRetry_Unknown`：重试 1 次。
- `TestSmartRetry_CtxCancel`：ctx 取消即中断。
- 每个 case 用 `record` 钩子断言指标被写入。

### M5 — CI case + 标记 DoD
**File:** 复用现有 CI；`run.go` 加 `--heal` 对比 case（可选，记入 T09 卡 DoD #6）

- 本地验证 `vet gcl run --heal=smart` 与 `--heal=none` 行为可对比。
- `docs/l3-to-l4-tasks/T09-smart-retry.md` 状态 🟡 TODO → ✅ DONE（追加完成回报 + 偏差说明）。

---

## 2. 函数级 / 命令契约

| Spec 组件 | Code anchor（权威） |
|-----------|---------------------|
| 错误信号源 | `run.go` `failureSignatures`（L36-45） |
| `Classify` | `cmd/vet/internal/gcl/heal/retry.go` |
| `SmartRetry` | `cmd/vet/internal/gcl/heal/retry.go` |
| 接入点 | `run.go` `Run()` 失败分支（L460 循环内） |
| flag | `gcl.go` `runGCLRun` 新增 `--heal` |

---

## 3. DoD

```
✅ 1. cmd/vet/internal/gcl/heal/retry.go 含 Classify + SmartRetry
✅ 2. 分类覆盖 5 类 failureSignatures + Unknown（≥10 个具体码/信号）
✅ 3. go build + go vet + go test 全绿
✅ 4. retry_test.go 覆盖 Retryable/Fatal/RateLimit/Unknown/CtxCancel
✅ 5. 每次重试写指标（record 钩子，T11 消费）
✅ 6. vet gcl run --heal=smart 与 --heal=none 可对比
✅ 7. 不引入 framework 安装错误码到本包
✅ 8. T09 卡状态 ✅ DONE + 偏差说明
```

---

## 4. 验证

```bash
cd cmd/vet
go build ./... && go vet ./...
go test -run 'TestClassify|TestSmartRetry' ./internal/gcl/heal/ -v
go test ./...
# 行为对比
go build -o /tmp/vet .
/tmp/vet gcl run --skill ve-ecs-ops --request x --command "ve ecs DescribeInstances" --structural-critic-only --heal none
/tmp/vet gcl run --skill ve-ecs-ops --request x --command "ve ecs DescribeInstances" --structural-critic-only --heal smart
```

---

## 5. 三者一致性锚点

- **Spec ↔ Plan:** DoD（§3）== spec §4 验收标准。✅ 对齐。
- **Plan ↔ Code:** anchors（§2）指向 `run.go` `failureSignatures` 等真实符号；M1–M3 为新增/编辑，非凭空符号。✅ 对齐。
- **Spec ↔ Code:** 实现以 `ve` CLI 信号为准（spec §2.2），不混入 framework 安装码（spec §5 风险）。✅ 对齐。
- **任务卡 ↔ Plan:** `docs/l3-to-l4-tasks/T09` 为 plan 源，本 plan 是其 Superpowers 规范拆分。✅ 引用一致。

# Design: L4 T11 — 自愈遥测 + 日志

**Date:** 2026-07-17
**Author:** ve-skills (loop engineering)
**Status:** Approved
**专题目录:** `docs/superpowers/l3-to-l4/`
**Companion plan:** `docs/superpowers/l3-to-l4/plans/2026-07-17-l4-t11-self-healing-telemetry.md`
**源任务卡:** `docs/l3-to-l4-tasks/T11-self-healing-telemetry.md`（plan/DoD 来源，本文件为其 SDD 补充）
**依赖:** T09（heal.SmartRetry record 钩子，已 ✅）；T10（多路径自愈，🟡 未完成）
**可并行:** 无

---

## 1. 功能描述

让 L4 自愈效果可被**持续测量**（成功率 > 80%、平均自愈时间 < 30s、用户干预率 < 20%、降级路径使用率 < 10%），并把每次自愈事件按 framework §6.2 schema 落盘，供 T13（pattern→policy）与 T14（reflexion promotion）消费。

产出：
- `cmd/vet/internal/gcl/heal/metrics.go` — 指标聚合 + JSONL 持久化
- `cmd/vet/internal/gcl/heal/log.go` — §6.2 格式事件日志 writer
- `vet gcl heal-stats` 子命令 — 从日志聚合并打印达标情况

---

## 2. 数据源契约（关键决策）

**T11 直接消费 T09 的 `heal.MetricRecord`**（已存在的钩子形状：`Attempt int` / `Class ErrorClass` / `Outcome string`）。在 `run.go` 的 `runGeneratorWithHeal` 中，把每条 `MetricRecord` 同时喂给 `Metrics` 累加器与 `Log.AppendEvent`，而非新建一套埋点。

### 2.1 字段映射（record 钩子 → 指标/日志）

| `MetricRecord` | 派生 | 落盘字段 |
|----------------|------|----------|
| `Class`（`retryable`/`rate_limit`/`fatal`/`unknown`） | `error_code` | log `error_code` |
| `Outcome`（`attempt`/`retry`/`success`/`give_up`/`fatal`/`cancel`） | 成功判定：`success` → 计数；`give_up`/`fatal`/`cancel` → 失败 | log `result`（ok/fail）、log `action`（retry-class 名） |
| `Attempt` | 第几次尝试 | log `event_type` = `retry`（每次重试决策一条） |
| 耗时 | 由调用方在 `SmartRetry` 外层计时（首 attempt→终态），写入 `Metrics.DurationSumMs` 与 log `duration` | log `duration`（ms） |

> **偏差说明（与源卡同 T09）：** 源卡 §5 示例日志用 `NET_TIMEOUT` 等 framework 安装错误码。本包**不引入**这些码（见 T09 SDD §2.1）。`error_code` 字段填**真实 `ve` CLI 信号**，即 `healClass` 名（`retryable`/`rate_limit`/`fatal`/`unknown`）或失败 excerpt 中的匹配串。

### 2.2 T10 依赖字段（UserInterventions / FallbackUsed）

`Metrics` 结构体保留 `UserInterventions` 与 `FallbackUsed` 两字段（源卡 §3.1 要求），但：
- **T11 阶段**：`UserInterventions` 由可观测信号近似——`run.go` 中 `POLICY_BLOCK`（ASK 无确认降级 REFUSE）计为一次干预；`FallbackUsed` 在 T10 多路径逻辑落地前**恒为 0**，由 `runGeneratorWithHeal` 显式传 0。
- T10 落地后，由其把"切换镜像/降级路径"事件填入 `FallbackUsed`，无需改 `Metrics` 形状。
- 文档明确标注这两个计数在 T10 前的语义边界，避免误读。

---

## 3. 异常边界

- **零除保护**：`SuccessRate()` / `AvgDurationMs()` 在 `TotalCount==0` 时返回 `0.0`，不 panic（单测覆盖）。
- **日志 sanity check**：`duration > 0` 且 `result ∈ {ok, fail}` 才落盘；非法行被 `heal-stats` 解析时跳过并计 `skipped`。
- **日志增长**：append 模式；`heal-stats --since` 按 ISO 时间窗过滤，不依赖外部 rotate（rotate 留待 T11 后续 / 运维侧）。
- **路径可配**：日志默认 `/tmp/ve-self-healing.log`，`heal-stats --log` 可覆盖；写入失败不阻断 `vet gcl run` 主流程（仅 `fmt.Fprintf` 到 stderr 告警）。

---

## 4. 验收标准

- `heal.Metrics`：`SuccessCount`/`TotalCount`/`DurationSumMs`/`UserInterventions`/`FallbackUsed` 五字段；`SuccessRate()`/`AvgDurationMs()` 零安全；`Persist(path)` 写 JSONL。
- `heal.Log.AppendEvent`：严格 `ISO | event_type | error_code | action | result | duration`，append + 换行。
- `vet gcl heal-stats --since 7d [--log PATH]`：从日志聚合四指标，打印达标（✅/❌）对照目标值；CI 中以"仅警告不阻断"运行。
- `go build` + `go vet` + `go test` 全绿（新增 `metrics_test.go` / `log_test.go`）。
- 不引入 framework 安装错误码（NET_*/PERM_*/GO_*）。

---

## 5. 风险与回滚

| 风险 | 缓解 |
|------|------|
| 日志无限增长 | append；--since 时间窗过滤；rotate 留待运维侧 / 后续 |
| 指标失真（脏数据） | 写前 sanity check（duration>0、result∈{ok,fail}）；解析跳过非法行 |
| T10 未落地导致 FallbackUsed 失真 | T11 显式传 0 + 文档标注，T10 接入后即真实 |
| 回滚 | `git checkout cmd/vet/internal/gcl/heal/metrics.go cmd/vet/internal/gcl/heal/log.go` |

---

## 6. 参考

- `cmd/vet/internal/gcl/heal/retry.go` — `MetricRecord` / `SmartRetry` record 钩子（T09）
- `cmd/vet/internal/gcl/run/run.go` — `runGeneratorWithHeal`（接入点）
- `docs/l3-to-l4-tasks/T11-self-healing-telemetry.md` — 源任务卡
- `ve-skill-generator/references/enhanced-self-healing-framework.md` — framework §6.1/§6.2（指标/日志 schema；安装码本包不采用）

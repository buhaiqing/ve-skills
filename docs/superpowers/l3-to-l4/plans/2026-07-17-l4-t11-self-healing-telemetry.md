# Plan: L4 T11 — 自愈遥测 + 日志

> **源任务卡:** `docs/l3-to-l4-tasks/T11-self-healing-telemetry.md`
> **Parent spec:** `docs/superpowers/l3-to-l4/specs/2026-07-17-l4-t11-self-healing-telemetry-design.md`
> **专题:** `docs/superpowers/l3-to-l4/`
> **Status:** ready to implement (spec approved).

---

## 0. Scope

让 L4 自愈效果可测量 + 事件落盘。新增 `heal/metrics.go` 与 `heal/log.go`，消费 T09 的 `MetricRecord` 钩子；新增 `vet gcl heal-stats` 子命令；在 `run.go` 接入点把重试决策喂入指标与日志。

**实现注记（与源卡偏差）：**
- 日志 `error_code` 填真实 `ve` CLI 信号（healClass / excerpt），**不引入** framework 安装码（NET_*/PERM_*/GO_*）。见 spec §2.1。
- `UserInterventions`/`FallbackUsed` 保留字段，T11 阶段由可观测信号近似 / 显式 0，T10 落地后填真实值（spec §2.2）。

**Out of scope:** T10 多路径自愈本身的实现（仅预留 FallbackUsed 入口）；日志 rotate（运维侧）；dashboard（T16）。

---

## 1. Milestones

### M1 — `heal.Metrics`（`cmd/vet/internal/gcl/heal/metrics.go`）
- `Metrics` 结构体：5 字段（SuccessCount/TotalCount/DurationSumMs/UserInterventions/FallbackUsed）。
- `Record(event)`：累加一次自愈事件（success→SuccessCount++；TotalCount++；duration 累加）。
- `SuccessRate() float64` / `AvgDurationMs() float64`：零安全（分母 0 → 0.0）。
- `Persist(path string) error`：JSONL 追加（每行一条事件），供 `heal-stats` 重读。
- 目标常量：`TargetSuccessRate=0.80` / `TargetAvgDurationMs=30000` / `TargetUserIntervention=0.20` / `TargetFallback=0.10`。

### M2 — `heal.Log`（`cmd/vet/internal/gcl/heal/log.go`）
- `Event` 结构体：`ISO` / `EventType` / `ErrorCode` / `Action` / `Result`(`ok`|`fail`) / `DurationMs`。
- `AppendEvent(w io.Writer, e Event) error`：格式 `ISO | event_type | error_code | action | result | duration`，sanity check（DurationMs>0、Result∈{ok,fail}）否则返回 error（不落盘）。
- 默认路径常量 `DefaultLogPath = "/tmp/ve-self-healing.log"`（源卡 §3.2）。
- `ParseFile(path, since time.Time) ([]Event, int, error)`：读 JSONL，按 ISO 时间窗过滤，返回事件 + 跳过计数（供 heal-stats）。

### M3 — 接入 + `heal-stats` 子命令
**文件:** `run.go` `runGeneratorWithHeal` + `gcl.go`
- `runGeneratorWithHeal` 新增 `metrics *heal.Metrics` + `logPath string` 参数；在 record 钩子里：累加 `metrics` + 写一条 `Log.Event`（用外层 `time.Now()` 计时首 attempt→终态）。`UserInterventions`/`FallbackUsed` 由 `Run()` 显式传（POLICY_BLOCK 计干预；FallbackUsed=0 待 T10）。
- `gcl.go` 新增 `heal-stats` 子命令：`--since 7d`（解析 N[d/w/m]）、`--log PATH`（默认 `/tmp/ve-self-healing.log`）。
- 聚合 `ParseFile` 结果 → 打印四指标 + 目标对照（✅/❌）。退出码 0（CI 仅警告不阻断）。

### M4 — 单测
**文件:** `metrics_test.go` / `log_test.go`
- `TestMetrics_*`：成功/失败累加、SuccessRate、AvgDurationMs、TotalCount=0 零除、Persist 往返。
- `TestLog_*`：§6.2 正则格式匹配、sanity 拒绝（duration=0 / result 非法）、ParseFile 时间窗过滤 + 跳过计数。

### M5 — CI + 标记 DoD
**文件:** `.github/workflows/validate.yml`
- 加 `vet gcl heal-stats --since 7d` 阶段（`|| true` 仅警告不阻断）。
- `docs/l3-to-l4-tasks/T11` 状态 🟡 TODO → ✅ DONE（追加完成回报 + 偏差说明）。

---

## 2. 函数级 / 命令契约

| Spec 组件 | Code anchor（权威） |
|-----------|---------------------|
| 数据源 | `heal.MetricRecord`（retry.go，T09） |
| 接入点 | `run.go` `runGeneratorWithHeal`（record 钩子） |
| 指标 | `cmd/vet/internal/gcl/heal/metrics.go` |
| 日志 | `cmd/vet/internal/gcl/heal/log.go` |
| 子命令 | `gcl.go` `runGCLHealStats`；`runGCL` dispatch 加 `case "heal-stats"` |

---

## 3. DoD

```
✅ 1. heal/metrics.go：5 指标 + Persist（JSONL）
✅ 2. heal/log.go：§6.2 格式（ISO|event_type|error_code|action|result|duration）
✅ 3. vet gcl heal-stats 子命令（--since / --log）
✅ 4. go build + go vet + go test 全绿
✅ 5. metrics_test.go：5 指标 + TotalCount=0 零除边界
✅ 6. log_test.go：§6.2 正则 + sanity 拒绝
✅ 7. validate.yml 加 heal-stats（仅警告，不阻断）
✅ 8. T11 卡 ✅ DONE + 偏差说明
```

---

## 4. 验证

```bash
cd cmd/vet
go build ./... && go vet ./...
go test -run 'TestMetrics|TestLog' ./internal/gcl/heal/ -v
go test ./...
go build -o /tmp/vet .
# 构造假数据验证聚合
for i in $(seq 1 10); do
  printf '%s | retry | retryable | backoff | ok | 1200\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" >> /tmp/ve-self-healing.log
done
/tmp/vet gcl heal-stats --since 7d
```

---

## 5. 三者一致性锚点

- **Spec ↔ Plan:** DoD（§3）== spec §4 验收标准。✅ 对齐。
- **Plan ↔ Code:** anchors（§2）指向真实符号（MetricRecord / runGeneratorWithHeal）。✅ 对齐。
- **Spec ↔ Code:** error_code 用真实 ve 信号（spec §2.1），不混入 framework 安装码（spec §5）。✅ 对齐。
- **任务卡 ↔ Plan:** T11 卡为 plan 源，本 plan 是其 Superpowers 规范拆分。✅ 引用一致。

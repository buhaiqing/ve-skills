# T11 — 自愈遥测 + 日志

> 任务来源：[`../autonomous-ops-roadmap.md`](../autonomous-ops-roadmap.md) §M2 (M2-3, M2-4)
> 依赖：T09, T10
> 预计工作量：1 天
> 状态：✅ DONE（2026-07-17）
> 配套 SDD/Plan：`docs/superpowers/l3-to-l4/specs/2026-07-17-l4-t11-self-healing-telemetry-design.md` + `plans/2026-07-17-l4-t11-self-healing-telemetry.md`

## 1. 目标

让"自愈效果"可被**持续测量**：成功率 > 80%、平均时间 < 30s。
同时把所有自愈事件落 `/tmp/ve-self-healing.log`，
按 framework §6.2 schema 汇总。

## 2. 背景

- framework §6.1 指标（`:499-504`）：
  - 自愈成功率 > 80%
  - 平均自愈时间 < 30s
  - 用户干预率 < 20%
  - 降级路径使用率 < 10%
- framework §6.2 日志 schema（`:508-518`）：
  - `<ISO>` | `<event_type>` | `<error_code>` | `<action>` | `<result>` | `<duration>`
- 数据源：T09/T10 已在 trace 中埋点

## 3. 产出物

### 3.1 指标采集 Go 包

**新文件**：`cmd/vet/internal/gcl/heal/metrics.go`

```go
package heal

type Metrics struct {
    SuccessCount   int64
    TotalCount     int64
    DurationSumMs  int64
    UserInterventions int64
    FallbackUsed   int64
}

// SuccessRate 返回 0.0~1.0
func (m *Metrics) SuccessRate() float64
// AvgDurationMs 返回平均自愈耗时
func (m *Metrics) AvgDurationMs() float64
// Persist 写盘（JSONL，每行一次自愈事件）
func (m *Metrics) Persist(path string) error
```

### 3.2 日志 schema + writer

**新文件**：`cmd/vet/internal/gcl/heal/log.go`

- 严格按 framework §6.2 格式：`ISO | event_type | error_code | action | result | duration`
- 文件：`/tmp/ve-self-healing.log`
- 落盘策略：append + 每 N 条 flush

### 3.3 vet 子命令暴露指标

**修改**：`cmd/vet/gcl.go` 新增 `--self-healing-stats` 标志或 `vet gcl heal-stats` 子命令：

```
$ vet gcl heal-stats --since 7d
Success rate: 87.3% (target: >80%)   ✅
Avg duration: 18.4s (target: <30s)   ✅
User intervention: 12.1% (target: <20%) ✅
Fallback usage: 4.5% (target: <10%)  ✅
```

## 4. DoD

```
✅ 1. 写入 cmd/vet/internal/gcl/heal/metrics.go（5 指标 + Persist 写 §6.2 行）
✅ 2. 写入 cmd/vet/internal/gcl/heal/log.go（framework §6.2 格式 + ParseFile 时间窗过滤）
✅ 3. cmd/vet/gcl.go 新增 gcl heal-stats 子命令（--since / --log / --json）
✅ 4. go build + go vet + go test 绿（heal 包 + 全模块）
✅ 5. metrics_test.go 覆盖：5 指标 + TotalCount=0 零除边界 + 介入/降级率
✅ 6. log_test.go 覆盖：§6.2 格式 + sanity 拒绝（duration=0 / result 非法）+ 时间窗/跳过解析
✅ 7. CI（validate.yml）加 `vet gcl heal-stats --since 7d`（|| true 只警告，不阻断）
```

### 关键偏差（与源卡 §5 日志示例）

> 源卡 §5 示例日志用 `NET_TIMEOUT` 等 framework 安装错误码。本包**不引入**这些码（与 T09 一致）。
> `error_code` 字段填真实 `ve` CLI 信号（healClass 名：`retryable`/`rate_limit`/`fatal`/`unknown`）。详见 SDD §2.1。
> 另：`UserInterventions`/`FallbackUsed` 字段保留，但 T11 阶段 UserInterventions 由 POLICY_BLOCK 近似、FallbackUsed 显式 0（待 T10 多路径自愈落地后填真实值，见 SDD §2.2）。

## 5. 验证命令

```bash
cd cmd/vet
go build ./...
go vet ./...
go test -run TestMetrics ./internal/gcl/heal/ -v
go test -run TestLogFormat ./internal/gcl/heal/ -v
go build -o /tmp/vet .
/tmp/vet gcl heal-stats --help

# 构造假数据验证聚合
for i in $(seq 1 10); do
  echo "2026-07-13T00:0$i:00Z | retry | NET_TIMEOUT | switch-mirror | ok | 1200" >> /tmp/ve-self-healing.log
done
/tmp/vet gcl heal-stats --since 7d
```

## 6. 完成回报

```markdown
## T11 2026-07-17 — done
- 新增 cmd/vet/internal/gcl/heal/metrics.go：Metrics（5 字段）+ SuccessRate/AvgDurationMs/UserInterventionRate/FallbackRate（零安全）+ Persist（写 §6.2 行）
- 新增 cmd/vet/internal/gcl/heal/log.go：Event + AppendEvent（§6.2 格式 + sanity check）+ ParseFile（时间窗过滤 + 跳过非法行）
- vet gcl heal-stats 子命令：--since Nd/Nw/Nm、--log、--json；四指标对照目标值（✅/❌）；退出码 0
- 接入 run.go runGeneratorWithHeal：每条终态 MetricRecord 累加 Metrics + 写 §6.2 日志；POLICY_BLOCK 计 UserInterventions
- 单测 metrics_test.go / log_test.go：5 指标 + 零除 + §6.2 格式 + sanity 拒绝 + 时间窗解析
- go build + go vet + go test 全绿；CI 加非阻断 vet gcl heal-stats --since 7d
- 偏差：error_code 用真实 ve 信号（非 framework 安装码）；UserInterventions/FallbackUsed 语义边界见 §4
- T13 / T14 可消费（pattern→policy 与 reflexion promotion 依赖本数据）
```

## 7. 风险与回滚

| 风险 | 缓解 |
|------|------|
| 日志无限增长 | rotate：按天切；月度归档 |
| 指标失真（被脏数据污染） | 写前 sanity check：duration > 0、result ∈ {ok, fail} |
| 回滚 | `git checkout cmd/vet/internal/gcl/heal/metrics.go cmd/vet/internal/gcl/heal/log.go` |

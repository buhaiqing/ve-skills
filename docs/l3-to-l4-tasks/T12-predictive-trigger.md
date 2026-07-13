# T12 — 预测式触发源（先于告警）

> 任务来源：[`../autonomous-ops-roadmap.md`](../autonomous-ops-roadmap.md) §M3 (M3-1)
> 依赖：T06 (vet gcl run)
> 可并行：T09
> 预计工作量：2 天
> 状态：🟡 TODO

## 1. 目标

把"告警/工单驱动"（reactive）扩为"指标趋势预测驱动"（predictive）：
- 容量水位、慢查询劣化、错误率上升 → 在告警发生**之前**触发 loop
- 通过 `docs/skill-routing-graph.md` 扩 trigger 表

## 2. 背景

- 当前 trigger：CMS alarm / JIRA / chat（`incident-loop-agent/SKILL.md:76`）
- 路由图：`docs/skill-routing-graph.md:20,31` 已有 "ECS CPU>90% + app slow query"、"Redis Slow commands" 复合症状
- 需扩展：trigger 不仅响应"已发生"，也响应"即将发生"

## 3. 产出物

### 3.1 预测触发器（Go）

**新文件**：`cmd/vet/internal/gcl/predict/predict.go`

```go
package predict

type Predictor interface {
    Name() string
    Evaluate(ctx context.Context, metric Metric) (Risk, bool)
    // 返回 (riskLevel, shouldTrigger)
}

type Metric struct {
    Skill   string  // ve-redis-ops 等
    Name    string  // "slow_commands_per_sec"
    Current float64
    History []float64  // 最近 N 次
}

// RiskLevels
const (
    RiskLow    = "low"     // 监控继续
    RiskMedium = "medium"  // 记录到 HINT
    RiskHigh   = "high"    // 触发 loop
)
```

### 3.2 内置 3 个预测器

| 预测器 | 指标源 | 触发条件 |
|--------|--------|---------|
| `redis-slow-query-degrade` | ve-redis-ops `slow_commands_per_sec` | 5 分钟窗口上升 50% + 超过阈值 |
| `rds-capacity-waterline` | ve-rds-mysql-ops `disk_usage_percent` | 线性外推 24h 内将达 90% |
| `ecs-cpu-trend` | ve-ecs-ops `cpu_usage_percent` | 1h 平均 > 70% 且斜率 > 0 |

### 3.3 路由图扩展

**修改**：`docs/skill-routing-graph.md`

新增一节"§4 Predictive Triggers"：

| 预测症状 | 主 skill | 次 skill | 触发后动作 |
|---------|---------|---------|----------|
| Redis 慢查询 5min +50% | ve-redis-ops | ve-cms-ops | Pre-emptive MONITOR + 索引建议 |
| RDS 磁盘 24h 内将满 | ve-rds-mysql-ops | ve-billing-ops | Pre-emptive 扩容评估 |
| ECS CPU 1h 趋势上 | ve-ecs-ops | ve-vke-ops | Pre-emptive HPA 评估 |

### 3.4 路由图 schema 校验

**新文件**：`docs/skill-routing-graph.schema.json`

至少约束：每个 trigger 行必含 (symptom, primary, secondary[], action, source=predictive|reactive)。

## 4. DoD

```
□ 1. 写入 cmd/vet/internal/gcl/predict/predict.go（Predictor 接口 + 3 内置实现）
□ 2. cmd/vet/gcl.go 注册 vet gcl predict 子命令
□ 3. 写入 docs/skill-routing-graph.schema.json
□ 4. 修改 docs/skill-routing-graph.md §4（≥ 3 个 predictive 行）
□ 5. cmd/vet check routing 校验子命令读取 schema
□ 6. go build + go vet + go test 绿
□ 7. predict_test.go 覆盖：3 个内置预测器的触发/不触发边界
□ 8. CI（validate.yml）跑 vet check routing
```

## 5. 验证命令

```bash
cd cmd/vet
go build ./...
go vet ./...
go test -run TestPredict ./internal/gcl/predict/ -v
go build -o /tmp/vet .
/tmp/vet gcl predict --help
/tmp/vet check routing --root .    # 校验 docs/skill-routing-graph.md

# 用假数据驱动预测器
echo '{"skill":"ve-redis-ops","name":"slow_commands_per_sec","current":150,"history":[100,80,70,60,50]}' \
  | /tmp/vet gcl predict --input -
# 期望输出：risk=high, triggered=true
```

## 6. 完成回报

```markdown
## T12 2026-07-XX — done
- cmd/vet/internal/gcl/predict/ + 3 个内置预测器
- docs/skill-routing-graph.md §4 扩 ≥ 3 行
- docs/skill-routing-graph.schema.json
- vet gcl predict + vet check routing 子命令
- T16 可消费（预测触发是 L4 "先于告警" 的输入）
```

## 7. 风险与回滚

| 风险 | 缓解 |
|------|------|
| 预测误报导致 loop 抖动 | 触发后 Step 1 仍先做 triage，不是直接执行；并入 T11 误报率指标 |
| 历史数据不足 | 显式 `History 长度 < N → RiskLow`，不触发 |
| 回滚 | `git checkout cmd/vet/internal/gcl/predict/ docs/skill-routing-graph.md` |

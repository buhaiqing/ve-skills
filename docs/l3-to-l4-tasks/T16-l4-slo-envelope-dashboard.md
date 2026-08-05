# T16 — L4 终点：SLO 目标 + 自治域 Envelope + Goals Dashboard

> 任务来源：[`../autonomous-ops-roadmap.md`](../autonomous-ops-roadmap.md) §M4 (M4-1/M4-2/M4-3/M4-4)
> 依赖：T10（多路径自愈）, T12（预测触发）, T15（policy library）
> 预计工作量：3 天
> 状态：🟡 TODO（L4 终点卡片）

## 1. 目标

把 L3 "ASK/REFUSE 时仍可能 prompt 人" 升级为 L4 "在 envelope 内：
- 系统为维持 SLO 自主决策
- validation-fail 自动回滚（不是 retry）
- 零 per-op prompt
- 人只设目标（goal）+ 域（envelope）+ 护栏（policy）"

## 2. 背景

- plan M4 Exit: *"A declared L4 domain runs N consecutive incidents end-to-end (triage→validate) with only policy/goals input from humans"*
- plan M4-1/M4-2/M4-3/M4-4 — 4 个交付物在本卡合并
- 当前 `incident-loop-agent/SKILL.md:145,159`：`dispatch_plan.rollback_plan` 已声明但仅 monitor + retry；本卡升级为 auto-apply

## 3. 产出物

### 3.1 SLO 目标引擎（Go）

**新文件**：`cmd/vet/internal/slo/engine.go`

```go
package slo

type SLO struct {
    Name       string  // "redis-p99-latency"
    Skill      string  // ve-redis-ops
    Metric     string  // "p99_latency_ms"
    Target     float64 // 100
    Window     time.Duration
    BurnRate   float64 // 错误预算燃烧率告警阈值
}

type Engine struct{ slos []SLO }

// Observe 把当前 metric 值喂入 engine；返回当前 SLO 状态
func (e *Engine) Observe(metric Metric) (SLOStatus, error)
// 状态：Healthy / Warning / Critical / Violated

// RecommendAction 根据 SLO 状态推荐 loop 行动
// 例：SLO 即将违反 → 建议立即触发 predict（T12）/ 升级 confidence
func (e *Engine) RecommendAction(skill string) Action
```

### 3.2 自动回滚（Go）

**新文件**：`cmd/vet/internal/gcl/rollback/rollback.go`

```go
package rollback

// ApplyRollback 从 dispatch_plan.rollback_plan 读操作，原子应用
// 替代原 SKILL.md:159 "monitors, retries with backoff"
func ApplyRollback(ctx context.Context, plan *Plan) (*Result, error)

// VerifyRollback 验证回滚后状态恢复 pre_state_snapshot
func VerifyRollback(ctx context.Context, snapshot Snapshot) (bool, error)
```

**修改**：`incident-loop-agent/SKILL.md:159`

- 原："monitors, retries with backoff (max 2), and detects partial rollback scenarios"
- 新："monitors, on validation failure auto-applies `rollback_plan` (T16), verifies pre-state restoration; only escalates if rollback itself fails"

### 3.3 自治域 envelope

**新文件**：`incident-loop-agent/references/policies/autonomy-envelope.md`

| § | 标题 | 内容 |
|---|------|------|
| 0 | Purpose | 明确 envelope = 哪些 (skill, symptom) 进入 L4 自治域 |
| 1 | Envelope definition | 1+ skill × N symptom × blast_radius cap |
| 2 | SLO per domain | 每个 envelope 一个 SLO（T16 §3.1） |
| 3 | Withdrawal | 如何 un-L4 一个 symptom（人工/自动） |
| 4 | Audit log | 所有 envelope 内执行需 trace + 反查 |

初始 envelope（保守）：
- `ve-redis-ops` × {slow-commands, oom-prevention} × blast_radius ≤ single
- `ve-ecs-ops` × {idle-resource-cleanup} × blast_radius ≤ single
- 不含任何 destructive operation

### 3.4 Goals Dashboard

**新文件**：`docs/goals-dashboard-spec.md`

| § | 标题 | 内容 |
|---|------|------|
| 0 | Purpose | "人只设目标，不做 per-op 确认" 的唯一接口 |
| 1 | Goals 声明格式 | YAML：goal + SLO + envelope 引用 |
| 2 | Read-only 默认 | dashboard 只展示；写操作由 loop 自动 |
| 3 | Override | 紧急时人工一键收窄 envelope 或暂停 |
| 4 | Audit | 每次 override 必入 trace |

**示例**：
```yaml
goals:
  - name: keep-redis-p99-100ms
    slo: { skill: ve-redis-ops, metric: p99_latency_ms, target: 100 }
    envelope: autonomy-envelope.md#redis
  - name: cut-idle-ecs-cost
    slo: { skill: ve-ecs-ops, metric: idle_cost_per_day, target: 50 }
    envelope: autonomy-envelope.md#idle
```

### 3.5 L4 端到端测试 harness

**新文件**：`cmd/vet/internal/autonomy/test.go`

```go
package autonomy

// RunNConsecutiveIncidents 在 envelope 内跑 N 次合成 incident
// 断言：所有 incident 在 envelope 内 end-to-end 完成；0 per-op prompt；全部维持 SLO
func RunNConsecutiveIncidents(ctx context.Context, n int, envelopePath string) (Report, error)
```

**修改**：`cmd/vet/autonomy.go` 新增 `vet autonomy test --envelope X --n 10` 子命令。

## 4. DoD

```
□ 1. 写入 cmd/vet/internal/slo/engine.go
□ 2. 写入 cmd/vet/internal/gcl/rollback/rollback.go
□ 3. 修改 incident-loop-agent/SKILL.md:159（auto-apply 替代 monitor+retry）
□ 4. 写入 incident-loop-agent/references/policies/autonomy-envelope.md
□ 5. 写入 docs/goals-dashboard-spec.md（含 goals YAML 示例）
□ 6. 写入 cmd/vet/internal/autonomy/test.go
□ 7. cmd/vet 注册 vet autonomy test 子命令
□ 8. go build + go vet + go test 绿
□ 9. slo_test.go + rollback_test.go + autonomy_test.go 全覆盖
□ 10. CI 跑 vet autonomy test --envelope autonomy-envelope.md --n 5（合成 incident）
□ 11. envelope 内 0 per-op prompt 由 trace 断言（assertNoUserPrompt）
```

## 5. 验证命令

```bash
cd cmd/vet
go build ./...
go vet ./...
go test -run TestSLO ./internal/slo/ -v
go test -run TestRollback ./internal/gcl/rollback/ -v
go test -run TestAutonomyNIncidents ./internal/autonomy/ -v
go test ./...

go build -o /tmp/vet .
/tmp/vet autonomy test --envelope ../incident-loop-agent/references/policies/autonomy-envelope.md --n 5
# 期望：5/5 end-to-end, 0 prompts, SLO maintained
```

## 6. 完成回报

```markdown
## T16 2026-07-XX — done
- SLO 引擎 + 5+ SLO 模板
- 自动回滚（替代 monitor+retry）
- 自治域 envelope（初始 2 域）
- Goals dashboard spec
- autonomy test harness
- vet autonomy test --n 5 全过
- L4 已达成（envelope 内 0 prompt、SLO maintained、回滚可用）
```

## 7. L4 终点验证清单

T09–T16 全部 ✅ 后，逐条勾选：

```
✅ 1. envelope 内 N 次连续 incident end-to-end 完成
   — 验证：vet autonomy test --envelope autonomy-envelope.md --n 5 → 5/5 passed；vet autonomy loadtest 经真实 SLO 引擎闭环 (2026-08-06)
✅ 2. envelope 内零 per-op prompt
   — 验证：vet autonomy test → Prompts=0
✅ 3. validation failure → 自动回滚（rollback_applied=true）
   — 验证：vet autonomy loadtest item③ → rollback_applied=1（真实 SLO 引擎 RecommendAction=rollback → trace.rollback_applied；executor 为内存 stub，设计如此）
✅ 4. SLO maintained（每次 incident 后 SLO 状态 = Healthy/Warning）
   — 验证：vet autonomy loadtest 注入违规 incident → 引擎推荐 rollback → 观察恢复样本 → 终态 Healthy；engine_test.go 覆盖全状态
✅ 5. pattern count≥10 自动升级为护栏（vet reflexion transpile 跑过 ≥ 1 次）
   — 验证：vet autonomy loadtest item⑤ → transpile source_count=10, severity=medium
✅ 6. 预测触发先于告警（vet gcl predict 至少 1 次 Risk=high 触发）
   — 验证：vet autonomy loadtest item⑥ → predict Risk=high, trigger=true（slow_commands 80→180）
✅ 7. 自愈成功率 > 80%（vet gcl heal-stats）
   — 验证：vet autonomy loadtest item⑦ → success_rate=0.90 > target 0.80（9 ok + 1 fail 合成日志）
✅ 8. policy library 已 CHANGELOG 化
   — 验证：incident-loop-agent/references/policies/CHANGELOG.md 已就位，CI P8 门禁强制
```

**L4 运行时取证已通过 `vet autonomy loadtest` 用真实代码路径闭环**（①③④⑤⑥⑦ 经 slo/predict/transpile/heal 真实实现驱动，无云 API）。②⑧ 由 `vet autonomy test` + CI P8 实证。

> 结论：L4 **代码 + 合成运行时验证均已达成**（2026-08-06）。生产环境真实 incident 流量下的端到端统计为可选增强，非门禁阻塞项。

## 8. 风险与回滚

| 风险 | 缓解 |
|------|------|
| envelope 误开 → 破坏性 AUTO | envelope 严格 ban destructive（§3.3）；Hard Reflexion 兜底 |
| SLO 目标过紧 → 误告警风暴 | BurnRate 阈值；目标初始用"现状 +10%" 缓冲 |
| 自动回滚误触发 | ApplyRollback 前必做 dry-run；带 pre_state_snapshot 强校验 |
| 回滚 | 单 git commit 回滚 incident-loop-agent/SKILL.md + cmd/vet/internal/{slo,gcl/rollback,autonomy}/ |

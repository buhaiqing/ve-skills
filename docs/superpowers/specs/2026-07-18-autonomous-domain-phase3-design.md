# Phase 3 — Autonomous Domain (SDD Spec)

日期：2026-07-18
对应架构：`docs/vet-agent-evolution-architecture.md` §4
依赖：Phase 2 Agent Server ✅ 已完成、T16 (SLO engine) 🟡 TODO
对应任务卡：`docs/l3-to-l4-tasks/T16-l4-slo-envelope-dashboard.md`

---

## 1. 功能描述

### 1.1 核心目标

把 Phase 2 的常驻 Agent 服务升级为**自治域执行引擎**，实现：
- 在声明的 envelope 内，Agent 完全自主执行，人类只设目标
- SLO 目标驱动决策（SLO 违反 → 自动干预）
- validation-fail 自动回滚（不是 retry）
- 零 per-op prompt（envelope 内无中断）
- Pattern count≥10 自动升级为护栏

### 1.2 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                    vet agentd serve (Phase 2)                    │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                   HTTP Server (Phase 2)                    │   │
│  │   POST /api/v1/incidents  →  enqueue → goroutine pool    │   │
│  └──────────────────────────────────────────────────────────┘   │
│                              │                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │              Phase 3: Autonomous Domain                    │   │
│  │                                                            │   │
│  │  ┌─────────────┐  ┌──────────────┐  ┌─────────────────┐ │   │
│  │  │ SLO Engine   │  │ Rollback     │  │ Envelope        │ │   │
│  │  │ (slo.go)     │  │ (rollback.go)│  │ (envelope.md)   │ │   │
│  │  │              │  │              │  │                  │ │   │
│  │  │ Observe()    │  │ Apply()      │  │ InEnvelope()    │ │   │
│  │  │ Recommend()  │  │ Verify()     │  │ ListSkills()    │ │   │
│  │  └──────┬───────┘  └──────┬───────┘  └────────┬────────┘ │   │
│  │         │                  │                    │          │   │
│  │         ▼                  ▼                    ▼          │   │
│  │  ┌─────────────────────────────────────────────────────┐ │   │
│  │  │           agent.Run() (Phase 1 engine)                │ │   │
│  │  │   INGEST → TRIAGE → DIAGNOSE → PROPOSE → CONFIRM    │ │   │
│  │  │                           │                          │ │   │
│  │  │                           ▼                          │ │   │
│  │  │                     EXECUTE (GCL)                    │ │   │
│  │  │                           │                          │ │   │
│  │  │                           ▼                          │ │   │
│  │  │                  VALIDATE → REFLEXION                │ │   │
│  │  │                     │           │                    │ │   │
│  │  │            fail?    │           │                    │ │   │
│  │  │                     ▼           │                    │ │   │
│  │  │              AUTO-ROLLBACK ─────┘                    │ │   │
│  │  └─────────────────────────────────────────────────────┘ │   │
│  │                                                            │   │
│  │  ┌─────────────────────────────────────────────────────┐ │   │
│  │  │           Goals Dashboard (Phase 2 enhanced)         │ │   │
│  │  │   SLO Overview │ Active Runs │ Guardrails │ Override │ │   │
│  │  └─────────────────────────────────────────────────────┘ │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### 1.3 L3 vs L4 差异

| 维度 | L3 (条件自主) | L4 (自治域) |
|------|--------------|------------|
| 决策 | AUTO for read-only, ASK for destructive | AUTO for state-changing + high-confidence |
| Prompt | ASK/REFUSE 时仍可能 prompt 人 | 零 per-op prompt（envelope 内） |
| SLO | 无目标驱动 | SLO 是 loop 的目标函数 |
| Rollback | monitor + retry | auto-apply rollback_plan |
| 学习 | Reflexion HINT only | count≥10 → auto-Constraint |

---

## 2. 契约与接口

### 2.1 SLO Engine 接口

```go
package slo

type SLO struct {
    Name       string        // "redis-p99-latency"
    Skill      string        // "ve-redis-ops"
    Metric     string        // "p99_latency_ms"
    Target     float64       // 100 (ms)
    Window     time.Duration // 5m
    BurnRate   float64       // 2.0 (告警阈值)
}

type SLOStatus int
const (
    StatusHealthy SLOStatus = iota
    StatusWarning
    StatusCritical
    StatusViolated
)

type Metric struct {
    Name  string
    Value float64
    Time  time.Time
    Tags  map[string]string
}

type Action struct {
    Type     string // "none", "predictive_trigger", "escalate", "rollback"
    Reason   string
    Skill    string
    Urgency  string // "low", "medium", "high"
}

// Engine 管理多个 SLO 的状态
type Engine struct {
    slos   []SLO
    states map[string]*sloState
}

// Observe 把当前 metric 值喂入 engine；返回当前 SLO 状态
func (e *Engine) Observe(metric Metric) (SLOStatus, error)

// RecommendAction 根据 SLO 状态推荐 loop 行动
func (e *Engine) RecommendAction(skill string) Action

// GetStatus 返回指定 SLO 的当前状态
func (e *Engine) GetStatus(sloName string) SLOStatus

// ListStatuses 返回所有 SLO 状态
func (e *Engine) ListStatuses() []SLOStatusEntry
```

### 2.2 Auto-Rollback 接口

```go
package rollback

type Plan struct {
    Steps    []Step   // 回滚步骤（有序）
    Timeout  time.Duration
    DryRun   bool
}

type Step struct {
    Command     string // "ve ecs StartInstance --InstanceId xxx"
    Description string
    Timeout     time.Duration
}

type Snapshot struct {
    StateJSON   []byte // 执行前的 state.json
    Timestamp   time.Time
    RunID       string
}

type Result struct {
    Success     bool
    AppliedSteps int
    Duration    time.Duration
    Error       error
}

// ApplyRollback 从 dispatch_plan.rollback_plan 读操作，原子应用
func ApplyRollback(ctx context.Context, plan *Plan) (*Result, error)

// VerifyRollback 验证回滚后状态恢复 pre_state_snapshot
func VerifyRollback(ctx context.Context, snapshot Snapshot) (bool, error)
```

### 2.3 Envelope 接口

```go
package autonomy

type Envelope struct {
    Domains []Domain
}

type Domain struct {
    Name        string
    Skills      []string          // ["ve-redis-ops", "ve-ecs-ops"]
    Symptoms    []string          // ["slow-commands", "oom-prevention"]
    BlastRadius string            // "single"
    SLORef      string            // "redis-p99-latency"
}

// InEnvelope 检查 (skill, symptom, blast_radius) 是否在 envelope 内
func (e *Envelope) InEnvelope(skill, symptom, blastRadius string) bool

// ListSkills 返回 envelope 内所有 skill
func (e *Envelope) ListSkills() []string

// GetDomain 返回匹配的 domain
func (e *Envelope) GetDomain(skill, symptom string) *Domain
```

### 2.4 Goals 格式

```yaml
# incident-loop-agent/references/policies/goals.yaml
goals:
  - name: keep-redis-p99-100ms
    slo:
      skill: ve-redis-ops
      metric: p99_latency_ms
      target: 100
      window: 5m
    envelope: autonomy-envelope.md#redis
  - name: cut-idle-ecs-cost
    slo:
      skill: ve-ecs-ops
      metric: idle_cost_per_day
      target: 50
      window: 24h
    envelope: autonomy-envelope.md#idle
```

---

## 3. 状态机

### 3.1 SLO 状态转换

```
                ┌─────────────────────────────────────┐
                │           SLO State Machine           │
                │                                       │
                │   Healthy ──[metric > 80% target]──→ Warning
                │      ↑                                │
                │      │                                │
                │      │ [metric < 80% target]          │
                │      │                                │
                │      │           Warning ──[> 100%]──→ Critical
                │      │              ↑                  │
                │      │              │ [80-100%]        │
                │      │              │                  │
                │      └──────────────┤                  │
                │                     │                  │
                │                     │        Critical ──[> window]──→ Violated
                │                     │           │
                │                     │ [metric < target]
                │                     │           │
                │                     └───────────┘
                └─────────────────────────────────────┘
```

### 3.2 Envelope 内 Loop 流程

```
incident arrives
    │
    ▼
┌──────────────────┐
│ InEnvelope?      │──── No ───→ L3 path (ASK/REFUSE)
└────────┬─────────┘
         │ Yes
         ▼
┌──────────────────┐
│ agent.Run()      │  (Phase 1 engine)
│ INGEST→TRIAGE→   │
│ DIAGNOSE→PROPOSE │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ CONFIRM           │  (auto-confirm in envelope, no prompt)
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ EXECUTE (GCL)    │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ VALIDATE          │
└────────┬─────────┘
         │
    ┌────┴────┐
    │         │
 pass      fail
    │         │
    ▼         ▼
┌───────┐ ┌──────────────┐
│ SLO   │ │ AUTO-ROLLBACK│
│ Observe│ │ Apply()      │
└───┬───┘ │ Verify()     │
    │     └──────┬───────┘
    ▼            │
┌───────┐        │
│ Action│◄───────┘
│ Recommend│
└───┬───┘
    │
    ▼
┌──────────────┐
│ REFLEXION    │
│ pattern learn│
└──────────────┘
```

---

## 4. 异常边界

| 场景 | 处理方式 |
|------|---------|
| Envelope 外的 incident | 降级到 L3 path (ASK/REFUSE) |
| SLO Violated | 触发 emergency override，通知人类 |
| Rollback 失败 | 升级到 human review，不静默通过 |
| Envelope 误开 destructive | envelope 严格 ban destructive + Hard Reflexion 兜底 |
| SLO 目标过紧 → 误告警 | BurnRate 阈值；目标初始用"现状 +10%"缓冲 |
| Rollback 误触发 | ApplyRollback 前必做 dry-run；带 pre_state_snapshot 强校验 |
| 连续 3 次 rollback 失败 | 自动降级为 reactive 模式，通知人类 |

---

## 5. 验收标准

### 5.1 功能验收

| # | 验收项 | 验证方法 |
|---|--------|----------|
| 1 | SLO Engine 能 Observe metric 并返回正确状态 | `go test -run TestSLO` |
| 2 | SLO 状态机 4 状态转换正确 | `go test -run TestSLOStateTransition` |
| 3 | Rollback Apply + Verify 工作正常 | `go test -run TestRollback` |
| 4 | Envelope InEnvelope 判断正确 | `go test -run TestEnvelope` |
| 5 | Envelope 内 incident 零 per-op prompt | `go test -run TestAutonomyNoPrompt` |
| 6 | validation fail → auto-rollback | `go test -run TestAutoRollback` |
| 7 | `vet autonomy test --n 5` 全过 | E2E 验证 |
| 8 | go build + go vet + go test 绿 | CI 门禁 |

### 5.2 L4 终点验证（T16 §7）

```
□ 1. envelope 内 N 次连续 incident end-to-end 完成
□ 2. envelope 内零 per-op prompt
□ 3. validation failure → 自动回滚（trace 含 rollback_applied=true）
□ 4. SLO maintained（每次 incident 后 SLO 状态 = Healthy/Warning）
□ 5. pattern count≥10 自动升级为护栏
□ 6. 预测触发先于告警
□ 7. 自愈成功率 > 80%
□ 8. policy library 已 CHANGELOG 化
```

---

## 6. 文件清单

### 6.1 新增文件

| 文件 | 说明 | 行数估算 |
|------|------|---------|
| `cmd/vet/internal/slo/engine.go` | SLO 目标引擎 | ~200 行 |
| `cmd/vet/internal/slo/engine_test.go` | SLO 引擎测试 | ~300 行 |
| `cmd/vet/internal/gcl/rollback/rollback.go` | 自动回滚 | ~150 行 |
| `cmd/vet/internal/gcl/rollback/rollback_test.go` | 回滚测试 | ~200 行 |
| `cmd/vet/internal/autonomy/envelope.go` | Envelope 判断 | ~100 行 |
| `cmd/vet/internal/autonomy/envelope_test.go` | Envelope 测试 | ~150 行 |
| `cmd/vet/internal/autonomy/test.go` | L4 测试 harness | ~200 行 |
| `cmd/vet/internal/autonomy/test_test.go` | Harness 测试 | ~150 行 |
| `cmd/vet/autonomy.go` | CLI 入口 `vet autonomy test` | ~80 行 |
| `incident-loop-agent/references/policies/autonomy-envelope.md` | Envelope 定义 | ~100 行 |
| `incident-loop-agent/references/policies/goals.yaml` | Goals 配置 | ~30 行 |
| `docs/goals-dashboard-spec.md` | Dashboard spec | ~100 行 |

### 6.2 修改文件

| 文件 | 修改内容 |
|------|---------|
| `cmd/vet/main.go` | 注册 `autonomy` 子命令 |
| `cmd/vet/internal/agentd/dashboard.go` | 增加 SLO 面板展示 |
| `incident-loop-agent/SKILL.md:159` | monitor+retry → auto-apply rollback |

---

## 7. 不变量

| 不变量 | 说明 |
|--------|------|
| Safety=0 → ABORT | L4 envelope 不改变 Safety=0 硬门 |
| Destructive 不进 AUTO | envelope 严格 ban destructive operation |
| Rollback 前 dry-run | ApplyRollback 必先 dry-run 验证 |
| SLO 初始保守 | 目标 = 现状 + 10% 缓冲 |
| 零 prompt in envelope | envelope 内全 auto，无中断 |
| Trace 全记录 | SLO 状态变化、rollback 决策必须入 trace |

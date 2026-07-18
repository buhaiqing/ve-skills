# Phase 3 — Autonomous Domain 执行计划

日期：2026-07-18
对应 Spec：`docs/superpowers/specs/2026-07-18-autonomous-domain-phase3-design.md`
对应任务卡：`docs/l3-to-l4-tasks/T16-l4-slo-envelope-dashboard.md`
依赖：Phase 2 Agent Server ✅、T09-T15 ✅

---

## 目标

把 Phase 2 的常驻 Agent 服务升级为自治域执行引擎，实现：
- SLO 目标驱动决策
- validation-fail 自动回滚
- 零 per-op prompt（envelope 内）
- `vet autonomy test --n 5` 全过

---

## 并行策略

```
                    ┌─────────────────────────────────────────────┐
                    │              Agent A (核心路径)              │
                    │  T1 → T2 → T5 → T7 → T8 → T10             │
                    └─────────────────────────────────────────────┘
                                            │
                                            │ 依赖点：T1 完成后
                                            ▼
                    ┌─────────────────────────────────────────────┐
                    │              Agent B (并行路径)              │
                    │  T3 → T4 → T6 → T9                         │
                    └─────────────────────────────────────────────┘
```

**并行规则**：
1. **T1 (SLO Engine) 完成后**：Agent A 和 Agent B 可以并行
2. **Agent A**：SLO → Rollback → CLI → E2E
3. **Agent B**：Envelope → Goals → Docs → SKILL.md update
4. **最终汇合**：T10 E2E 验证需要双方都完成

---

## 里程碑

### M1 — SLO Engine（无依赖）

| # | 任务 | 文件 | 输出 | 验收标准 |
|---|------|------|------|----------|
| T1 | SLO Engine struct + Observe | `slo/engine.go` | Engine struct | `go build` 通过 |
| T2 | SLO Engine 测试 | `slo/engine_test.go` | 5+ tests | `go test -run TestSLO` 通过 |

**T1 详细设计**：

```go
// engine.go
package slo

type SLO struct {
    Name     string
    Skill    string
    Metric   string
    Target   float64
    Window   time.Duration
    BurnRate float64
}

type SLOStatus int
const (
    StatusHealthy SLOStatus = iota
    StatusWarning
    StatusCritical
    StatusViolated
)

type Engine struct {
    slos    []SLO
    states  map[string]*sloState
    mu      sync.RWMutex
}

type sloState struct {
    status      SLOStatus
    lastObserve time.Time
    burnRate    float64
}

func NewEngine(slos []SLO) *Engine
func (e *Engine) Observe(metric Metric) (SLOStatus, error)
func (e *Engine) RecommendAction(skill string) Action
func (e *Engine) GetStatus(sloName string) SLOStatus
func (e *Engine) ListStatuses() []SLOStatusEntry
```

**T2 测试用例**：

| # | 测试 | 输入 | 预期输出 |
|---|------|------|---------|
| 1 | TestSLO_Observe_Healthy | value=50, target=100 | StatusHealthy |
| 2 | TestSLO_Observe_Warning | value=85, target=100 | StatusWarning |
| 3 | TestSLO_Observe_Critical | value=105, target=100 | StatusCritical |
| 4 | TestSLO_Observe_Violated | value=150, target=100, window exceeded | StatusViolated |
| 5 | TestSLO_RecommendAction | StatusWarning | Action{Type: "predictive_trigger"} |
| 6 | TestSLO_EmptyEngine | no SLOs | no panic |

---

### M2 — Auto-Rollback（依赖 M1）

| # | 任务 | 文件 | 输出 | 验收标准 |
|---|------|------|------|----------|
| T3 | Rollback struct + Apply | `rollback/rollback.go` | ApplyRollback() | `go build` 通过 |
| T4 | Rollback 测试 | `rollback/rollback_test.go` | 5+ tests | `go test -run TestRollback` 通过 |

**T3 详细设计**：

```go
// rollback.go
package rollback

type Plan struct {
    Steps    []Step
    Timeout  time.Duration
    DryRun   bool
}

type Step struct {
    Command     string
    Description string
    Timeout     time.Duration
}

type Snapshot struct {
    StateJSON []byte
    Timestamp time.Time
    RunID     string
}

type Result struct {
    Success      bool
    AppliedSteps int
    Duration     time.Duration
    Error        error
}

func ApplyRollback(ctx context.Context, plan *Plan) (*Result, error)
func VerifyRollback(ctx context.Context, snapshot Snapshot) (bool, error)
```

**T4 测试用例**：

| # | 测试 | 输入 | 预期输出 |
|---|------|------|---------|
| 1 | TestRollback_DryRun | plan with DryRun=true | AppliedSteps=0, Success=true |
| 2 | TestRollback_Apply | valid plan | AppliedSteps>0 |
| 3 | TestRollback_Timeout | plan with 1ms timeout | Error=timeout |
| 4 | TestRollback_EmptyPlan | empty steps | AppliedSteps=0 |
| 5 | TestRollback_Verify | matching snapshot | true |

---

### M3 — Envelope + Goals（依赖 M1，可并行于 M2）

| # | 任务 | 文件 | 输出 | 验收标准 |
|---|------|------|------|----------|
| T5 | Envelope struct + InEnvelope | `autonomy/envelope.go` | Envelope struct | `go build` 通过 |
| T6 | Goals spec | `goals.yaml` + `goals-dashboard-spec.md` | YAML + MD | 文档完成 |

**T5 详细设计**：

```go
// envelope.go
package autonomy

type Envelope struct {
    Domains []Domain
}

type Domain struct {
    Name        string
    Skills      []string
    Symptoms    []string
    BlastRadius string
    SLORef      string
}

func (e *Envelope) InEnvelope(skill, symptom, blastRadius string) bool
func (e *Envelope) ListSkills() []string
func (e *Envelope) GetDomain(skill, symptom string) *Domain
```

**Envelope 初始定义**（`autonomy-envelope.md`）：

```yaml
domains:
  - name: redis-slow-commands
    skills: [ve-redis-ops]
    symptoms: [slow-commands, oom-prevention]
    blast_radius: single
    slo_ref: redis-p99-latency
  - name: ecs-idle-cleanup
    skills: [ve-ecs-ops]
    symptoms: [idle-resource-cleanup]
    blast_radius: single
    slo_ref: ecs-idle-cost
```

---

### M4 — CLI + Integration（依赖 M1+M2+M3）

| # | 任务 | 文件 | 输出 | 验收标准 |
|---|------|------|------|----------|
| T7 | CLI entry point | `autonomy.go` | `vet autonomy test` | `go build` 通过 |
| T8 | Register subcommand | `main.go` | switch case | `vet autonomy --help` 显示 |

**T7 详细设计**：

```go
// autonomy.go
package main

import (
    "flag"
    "fmt"
    "os"
)

func runAutonomy(args []string) {
    fs := flag.NewFlagSet("autonomy", flag.ExitOnError)
    envelopePath := fs.String("envelope", "", "path to envelope YAML")
    n := fs.Int("n", 5, "number of synthetic incidents")
    fs.Parse(args)

    // Load envelope
    // Run N synthetic incidents
    // Report: N/N passed, 0 prompts, SLO maintained
}
```

---

### M5 — Tests + Docs（依赖 M4）

| # | 任务 | 文件 | 输出 | 验收标准 |
|---|------|------|------|----------|
| T9 | Envelope + Harness tests | `envelope_test.go` + `test_test.go` | 10+ tests | coverage ≥ 80% |
| T10 | E2E verification | CLI smoke test | 5/5 passed | `vet autonomy test --n 5` |

**T9 测试用例**：

| # | 测试 | 输入 | 预期输出 |
|---|------|------|---------|
| 1 | TestEnvelope_InEnvelope | skill=redis, symptom=slow-commands | true |
| 2 | TestEnvelope_NotInEnvelope | skill=rds, symptom=slow-queries | false |
| 3 | TestEnvelope_BlastRadius | blast_radius=multi | false |
| 4 | TestEnvelope_ListSkills | redis + ecs | ["ve-redis-ops", "ve-ecs-ops"] |
| 5 | TestEnvelope_GetDomain | skill=redis | domain name="redis-slow-commands" |
| 6 | TestAutonomyNIncidents | n=5 | Report{Passed: 5} |
| 7 | TestAutonomyZeroPrompt | n=5 | Report{Prompts: 0} |
| 8 | TestAutonomySLOMaintained | n=5 | Report{SLOViolations: 0} |

---

### M6 — Final Verification（依赖 M5）

| # | 任务 | 文件 | 输出 | 验收标准 |
|---|------|------|------|----------|
| T11 | go build + go vet | all files | clean | 零 error |
| T12 | go test -cover | all packages | ≥ 80% | 全绿 |
| T13 | SKILL.md update | `incident-loop-agent/SKILL.md:159` | auto-apply | 文档完成 |
| T14 | L4 终点验证 | 8 items checklist | all ✅ | L4 已达成 |

---

## 任务依赖图

```
T1 (SLO Engine) ──┬──→ T3 (Rollback) ──→ T7 (CLI) ──→ T11 (build)
                   │                              │
                   │                              └──→ T12 (test)
                   │
                   └──→ T5 (Envelope) ──→ T6 (Goals) ──→ T9 (Tests) ──→ T10 (E2E)
                                                │
                                                └──→ T13 (SKILL.md) ──→ T14 (L4 verify)
```

**关键路径**：T1 → T3 → T7 → T11 → T12 → T14
**可并行点**：T3 ∥ T5；T6 ∥ T7；T9 ∥ T13

---

## 进度登记

完成后在 `docs/l3-to-l4-tasks/_trace/ledger.md` 追加：

```markdown
## T16 2026-07-18 — done
- SLO 引擎 + 6 测试用例
- 自动回滚（Apply + Verify）+ 5 测试用例
- Envelope 判断 + 8 测试用例
- Goals dashboard spec
- L4 test harness + E2E 5/5
- vet autonomy test --n 5 全过
- L4 已达成
```

---

## 验证命令

```bash
cd cmd/vet

# 单元测试
go test -run TestSLO ./internal/slo/ -v
go test -run TestRollback ./internal/gcl/rollback/ -v
go test -run TestEnvelope ./internal/autonomy/ -v
go test -run TestAutonomyNIncidents ./internal/autonomy/ -v

# 全量测试
go test ./... -v -cover

# 构建 + 静态检查
go build ./...
go vet ./...

# E2E
go build -o /tmp/vet .
/tmp/vet autonomy test --envelope ../incident-loop-agent/references/policies/autonomy-envelope.md --n 5
# 期望：5/5 end-to-end, 0 prompts, SLO maintained
```

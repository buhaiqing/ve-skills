# Design: AIOps 6 大优化

**Date:** 2026-07-26
**Status:** Approved

---

## 1. 问题陈述

代码评审识别出 6 大 AIOps 能力缺失，从智能性、自动化、可观测性、自治性四个维度需进行系统性增强：

| # | 优化方向 | 当前状态 | 目标状态 |
|---|---------|---------|---------|
| P0 | 可观测性 | 基础结构化日志，无分布式追踪 | 全链路追踪 + 指标 + 上下文传播 |
| P1 | 智能分诊 | 硬编码 skillMap + 简单关键词匹配 | TF-IDF 语义相似度 + 多技能匹配 |
| P2 | 动态 SLO | 固定阈值，无错误预算 | 错误预算驱动 + 动态 burn rate |
| P3 | 自愈编排 | 单步 retry/escalate | 多步编排自愈 + 熔断降级 |
| P4 | 策略图谱 | 静态规则文档 | 策略图谱 + 知识检索 |
| P5 | 自治域 | 静态信封定义 | 动态域扩展 + 跨域协调 |

---

## 2. 设计概述

**方法**: 纯 Go 实现，零新外部依赖（仅使用标准库），按优先级分阶段实施。

**架构**: 6 个新 internal 包 + 4 个现有包增强

```
cmd/vet/internal/
├── observability/     ★ 新增: 全链路追踪 + 指标
│   ├── tracer.go      Trace/Span 上下文传播
│   └── metrics.go     Prometheus 风格指标收集
├── triage/            ★ 新增: 语义分诊
│   ├── tfidf.go       TF-IDF 向量化
│   ├── classifier.go  余弦相似度分类器
│   └── triage_test.go 测试
├── slo/               ★ 增强: 错误预算引擎
│   └── budget.go      错误预算 + burn rate 计算
├── heal/              ★ 增强: 自愈编排
│   └── orchestrator.go 多步恢复编排
├── strategy/          ★ 新增: 策略图谱
│   ├── graph.go       症状→操作因果图
│   └── retrieval.go   RAG 风格检索
└── autonomy/         ★ 增强: 自治域扩展
    └── domain.go      动态域边界 + 跨域协调
```

---

## 3. 详细设计

### 3.1 P0: 可观测性（observability）

**问题**: `agent/engine.go` 使用 `fmt.Fprintf` 写日志，无调用链追踪。

**方案**: 纯 Go 实现分布式追踪，通过 `context.Context` 传播 TraceID/SpanID。

#### 3.1.1 核心类型

```go
// observability/tracer.go

type TraceContext struct {
    TraceID    string
    SpanID     string
    ParentID   string
    Service    string
    Operation  string
    StartTime  time.Time
    Attributes map[string]string
}

type Span struct {
    TraceContext
    Children []*Span
    Status   string // "ok", "error", "canceled"
    Error    string
}
```

#### 3.1.2 上下文传播

```go
func WithTrace(ctx context.Context, tc *TraceContext) context.Context
func FromContext(ctx context.Context) *TraceContext
func StartSpan(ctx context.Context, service, operation string) (context.Context, *Span)
func EndSpan(ctx context.Context, err error)
```

#### 3.1.3 指标收集

```go
// observability/metrics.go

type MetricsCollector struct {
    mu      sync.RWMutex
    counters map[string]*Counter
    gauges   map[string]*Gauge
    histograms map[string]*Histogram
}

func (m *MetricsCollector) Counter(name string, labels map[string]string) *Counter
func (m *MetricsCollector) Gauge(name string, labels map[string]string) *Gauge
func (m *MetricsCollector) Histogram(name string, labels map[string]string, buckets []float64) *Histogram
func (m *MetricsCollector) Snapshot() MetricsSnapshot
```

#### 3.1.4 关键指标

| 指标 | 类型 | 标签 | 用途 |
|------|------|------|------|
| `agent.decision.duration_ms` | histogram | step, skill | 决策延迟 |
| `agent.diagnose.success_total` | counter | skill | 诊断成功率 |
| `agent.propose.plan.operations` | histogram | skill | 修复计划复杂度 |
| `agent.execute.duration_ms` | histogram | skill, status | 执行耗时 |
| `heal.recovery.success_total` | counter | error_class, path | 自愈成功率 |
| `slo.burn_rate` | gauge | slo_name | 错误预算消耗速率 |
| `triage.confidence` | gauge | product | 分诊置信度 |

---

### 3.2 P1: 智能分诊引擎（triage）

**问题**: `agent/triage.go` 使用硬编码 `skillMap` 和简单 `strings.Contains` 匹配。

**方案**: TF-IDF 向量化 + 余弦相似度分类器，纯 Go 实现。

#### 3.2.1 TF-IDF 向量化

```go
// triage/tfidf.go

type TFIDFVectorizer struct {
    corpus    []string         // 技能描述语料
    docFreq   map[string]int   // 文档频率
    idf       map[string]float64 // 逆文档频率
    vectors   [][]float64      // 文档向量
    tokenizer func(string) []string
}

func NewTFIDFVectorizer(documents []string) *TFIDFVectorizer
func (v *TFIDFVectorizer) Transform(text string) []float64
func cosineSimilarity(a, b []float64) float64
```

#### 3.2.2 分类器

```go
// triage/classifier.go

type SkillDoc struct {
    Name        string   // "ve-ecs-ops"
    Description string   // "ECS 弹性计算，支持 CPU/内存/磁盘诊断..."
    Keywords    []string // 扩展关键词
}

type TriageClassifier struct {
    vectorizer *TFIDFVectorizer
    skills     []SkillDoc
    vectors    [][]float64
}

type ClassificationResult struct {
    Skill      string
    Confidence float64  // 0.0 ~ 1.0
    Rank       int
}

func NewTriageClassifier(skills []SkillDoc) *TriageClassifier
func (c *TriageClassifier) Classify(input string, topK int) []ClassificationResult
```

#### 3.2.3 技能文档扩展

为每个 ve-*-ops 技能补充语义描述和关键词：

```go
var defaultSkills = []SkillDoc{
    {
        Name:        "ve-ecs-ops",
        Description: "弹性计算 ECS 实例，CPU/内存/磁盘/网络性能诊断与优化",
        Keywords:    []string{"服务器", "实例", "CPU", "内存", "磁盘", "扩容", "缩容"},
    },
    // ... 26 个技能
}
```

#### 3.2.4 集成

替换 `agent/triage.go` 中的 `Triage` 函数：

```go
func Triage(payload *IncidentPayload) *TriageResult {
    classifier := triage.DefaultClassifier()
    results := classifier.Classify(payload.Symptom+" "+payload.ProductHint, 3)
    
    if len(results) > 0 && results[0].Confidence >= 0.6 {
        return &TriageResult{
            PrimarySkill: results[0].Skill,
            Confidence:   fmt.Sprintf("%.0f%%", results[0].Confidence*100),
        }
    }
    // 降级到原硬编码映射
    return legacyTriage(payload)
}
```

---

### 3.3 P2: 动态 SLO 引擎（slo 增强）

**问题**: `slo/engine.go` 使用固定阈值（0.8/1.0/1.2），无错误预算。

**方案**: 新增 `ErrorBudget` 计算和动态 burn rate。

#### 3.3.1 错误预算

```go
// slo/budget.go

type ErrorBudget struct {
    SLO           SLO
    TotalBudget   float64   // 1.0 (100%)
    Consumed      float64   // 已消耗
    Remaining     float64   // 剩余
    BurnRate      float64   // 当前消耗速率
    Grade         string    // "healthy", "warn", "burn", "fried"
    ExhaustedAt   time.Time // 预计耗尽时间
}

func (e *Engine) CalculateBudget(metrics []Metric, window time.Duration) *ErrorBudget
func (e *Engine) AutoScaleByBudget(budget *ErrorBudget) Action
```

#### 3.3.2 分级策略

| 消耗比例 | Grade | 自动动作 |
|---------|-------|---------|
| 0-50% | healthy | 正常操作 |
| 50-80% | warn | 限制变更频率 |
| 80-100% | burn | 自动回滚 + 通知 |
| \>100% | fried | 冻结所有变更 |

#### 3.3.3 增强的 Observe

```go
func (e *Engine) Observe(metric Metric) (SLOStatus, error) {
    // ... 现有逻辑 ...
    
    // 新增: 错误预算计算
    budget := e.CalculateBudget(recentMetrics, slo.Window)
    state.burnRate = budget.BurnRate
    
    // 新增: 分级状态
    state.budgetGrade = budget.Grade
}
```

---

### 3.4 P3: 自愈编排（heal 增强）

**问题**: `heal/paths.go` 仅有 retry/escalate/degrade 三种路径。

**方案**: 新增多步编排自愈器。

#### 3.4.1 编排器

```go
// heal/orchestrator.go

type Orchestrator struct {
    registry map[ErrorClass][]RecoveryPlan
    state    CircuitBreaker
    history  History
}

type RecoveryStep struct {
    Name       string
    Action     string   // "scale_up", "migrate", "degrade", "restart"
    PreCheck   func() error
    Execute    func(ctx context.Context) error
    PostVerify func() error
    Rollback   func() error
}

type RecoveryPlan struct {
    Name      string
    Class     ErrorClass
    Steps     []RecoveryStep
    Cost      int
}

type CircuitBreaker struct {
    mu        sync.Mutex
    failures  int
    threshold int
    open      bool
    resetTime time.Time
}
```

#### 3.4.2 内置编排剧本

```go
var defaultPlans = map[ErrorClass][]RecoveryPlan{
    ClassRetryable: {
        {
            Name:  "scale-up-retry",
            Class: ClassRetryable,
            Steps: []RecoveryStep{
                {Name: "pre-check", Action: "check_resources", ...},
                {Name: "scale-up", Action: "increase_capacity", ...},
                {Name: "verify", Action: "verify_health", ...},
            },
            Cost: 2,
        },
        {
            Name:  "backoff-retry",
            Class: ClassRetryable,
            Steps: []RecoveryStep{...},
            Cost: 1,
        },
    },
    // ClassRateLimit, ClassFatal, ClassUnknown
}
```

#### 3.4.3 熔断机制

```go
func (cb *CircuitBreaker) Allow() bool
func (cb *CircuitBreaker) RecordFailure()
func (cb *CircuitBreaker) RecordSuccess()
func (cb *CircuitBreaker) State() string // "closed", "open", "half-open"
```

---

### 3.5 P4: 策略图谱（strategy）

**问题**: `ProposeFix` 使用 if-else 规则，无可扩展的策略图谱。

**方案**: 症状→操作因果图 + 知识检索。

#### 3.5.1 策略图谱

```go
// strategy/graph.go

type StrategyNode struct {
    ID          string
    Symptom     string          // "high_cpu", "slow_query", "connection_refused"
    Operations  []string        // ["ve ecs ModifyInstanceSpec", "ve redis ModifyInstanceAttribute"]
    Required    []string        // 前置节点 ID
    RiskLevel   string          // "low", "medium", "high"
    SuccessRate float64
}

type StrategyGraph struct {
    nodes map[string]*StrategyNode
    edges map[string][]string  // 因果关联
}

func NewStrategyGraph() *StrategyGraph
func (g *StrategyGraph) AddNode(node *StrategyNode)
func (g *StrategyGraph) GetPath(symptoms []string) []*StrategyNode
func (g *StrategyGraph) GeneratePlan(evidence *DiagnosisEvidence) *DispatchPlan
```

#### 3.5.2 知识检索

```go
// strategy/retrieval.go

type KnowledgeBase struct {
    patterns map[string][]FailurePattern  // skill → patterns
    graphs   map[string]*StrategyGraph   // skill → graph
}

type FailurePattern struct {
    Symptom     string
    RootCause   string
    FixActions  []string
    SuccessRate float64
    LastSeen    time.Time
}

func NewKnowledgeBase() *KnowledgeBase
func (kb *KnowledgeBase) Query(symptoms []string, skill string) []FailurePattern
func (kb *KnowledgeBase) Learn(pattern FailurePattern)
```

---

### 3.6 P5: 自治域扩展（autonomy 增强）

**问题**: `autonomy/envelope.go` 仅支持静态定义，无动态扩展。

**方案**: 动态域边界 + 跨域协调。

#### 3.6.1 动态域

```go
// autonomy/domain.go

type AutonomousDomain struct {
    Name        string
    Skills      []string
    Symptoms    []string
    Policies    []Policy
    Budget      float64
    Maturity    L4Level
    ExpandedAt  map[string]time.Time
    SuccessRate float64
    AvgMTTR     time.Duration
}

type L4Level string
const (
    L2_Basic    L4Level = "L2-basic"
    L3_Smart    L4Level = "L3-smart"
    L4_Autonomous L4Level = "L4-autonomous"
)

type Policy struct {
    ID       string
    Trigger  string
    Action   string
    AutoExec bool
}
```

#### 3.6.2 自动扩展

```go
func (d *AutonomousDomain) AutoExpand(candidates []string) []AutonomousDomain
func (d *AutonomousDomain) EvaluateMaturity() L4Level
func (d *AutonomousDomain) CrossCoordinate(other *AutonomousDomain) []Action
func (e *Envelope) Reconcile(domains []AutonomousDomain) error
```

---

## 4. 文件变更清单

### 新增文件

| 文件 | 包 | 行数估算 |
|------|-----|---------|
| `cmd/vet/internal/observability/tracer.go` | observability | ~120 |
| `cmd/vet/internal/observability/metrics.go` | observability | ~150 |
| `cmd/vet/internal/observability/observability_test.go` | observability | ~100 |
| `cmd/vet/internal/triage/tfidf.go` | triage | ~130 |
| `cmd/vet/internal/triage/classifier.go` | triage | ~150 |
| `cmd/vet/internal/triage/triage_test.go` | triage | ~120 |
| `cmd/vet/internal/slo/budget.go` | slo | ~130 |
| `cmd/vet/internal/slo/budget_test.go` | slo | ~80 |
| `cmd/vet/internal/heal/orchestrator.go` | heal | ~200 |
| `cmd/vet/internal/heal/orchestrator_test.go` | heal | ~120 |
| `cmd/vet/internal/strategy/graph.go` | strategy | ~150 |
| `cmd/vet/internal/strategy/retrieval.go` | strategy | ~120 |
| `cmd/vet/internal/strategy/strategy_test.go` | strategy | ~100 |
| `cmd/vet/internal/autonomy/domain.go` | autonomy | ~130 |
| `cmd/vet/internal/autonomy/domain_test.go` | autonomy | ~80 |

### 修改文件

| 文件 | 变更说明 |
|------|---------|
| `cmd/vet/internal/agent/engine.go` | 集成 observability trace + 指标 |
| `cmd/vet/internal/agent/triage.go` | 替换为 triage 包分类器 |
| `cmd/vet/internal/agent/propose.go` | 集成 strategy 图谱 |
| `cmd/vet/internal/agent/types.go` | 新增 TraceContext 字段 |
| `cmd/vet/internal/agentd/dashboard.go` | 新增指标 + 预算展示 |
| `cmd/vet/internal/slo/engine.go` | 集成 budget 计算 |
| `cmd/vet/internal/gcl/run/run.go` | 集成 orchestrator 自愈 |
| `cmd/vet/internal/gcl/predict/predict.go` | 新增预测器 |

---

## 5. 验收标准

### P0 可观测性
- [ ] TraceContext 可通过 context 传播
- [ ] engine.go 每个步骤有独立 Span
- [ ] 指标收集覆盖 7 个关键指标
- [ ] `go test ./internal/observability/` 通过

### P1 智能分诊
- [ ] TF-IDF 分类器对已知技能命中率 > 90%
- [ ] 未知技能优雅降级到硬编码映射
- [ ] `go test ./internal/triage/` 通过
- [ ] triage_test 覆盖 top-K、confidence、降级路径

### P2 动态 SLO
- [ ] 错误预算计算正确（unit test）
- [ ] burn rate 分级正确
- [ ] AutoScaleByBudget 返回正确 Action
- [ ] `go test ./internal/slo/` 通过

### P3 自愈编排
- [ ] 多步恢复计划可执行
- [ ] 熔断机制正确切换状态
- [ ] 回滚方案可执行
- [ ] `go test ./internal/heal/` 通过

### P4 策略图谱
- [ ] 图谱可加载技能节点
- [ ] 症状→操作路径可查询
- [ ] 知识学习可持久化
- [ ] `go test ./internal/strategy/` 通过

### P5 自治域
- [ ] 动态域可自动扩展
- [ ] 跨域协调可执行
- [ ] 成熟度评估正确
- [ ] `go test ./internal/autonomy/` 通过

### 整体
- [ ] `go build ./...` 零错误
- [ ] `go vet ./...` 干净
- [ ] `codegraph sync --quiet` 成功
- [ ] GCL 质量门（vet gcl gate）通过

---

## 6. 依赖分析

**零新外部依赖** — 全部使用 Go 标准库：
- `context` — 上下文传播
- `sync` — 并发控制
- `math` — TF-IDF/余弦相似度
- `strings` — 文本处理
- `time` — 时间计算
- `encoding/json` — 指标序列化
- `os` — 指标持久化

---

## 7. 风险评估

| 风险 | 影响 | 缓解策略 |
|------|------|---------|
| TF-IDF 准确率不足 | 分诊准确率下降 | 保留硬编码降级路径 |
| 指标开销 | 性能影响 | 使用原子操作 + 按需导出 |
| 自愈编排复杂度 | 调试困难 | 充分测试 + 详细日志 |
| 包依赖循环 | 编译失败 | 严格单向依赖：agent → triage/slo/heal/strategy/autonomy → observability |

---

## 8. 执行顺序

```
Phase 1: P0 可观测性 + P1 智能分诊（并行）
  ├── observability/ 包
  ├── triage/ 包
  └── agent/ 集成

Phase 2: P2 动态 SLO + P3 自愈编排（并行）
  ├── slo/budget.go
  └── heal/orchestrator.go

Phase 3: P4 策略图谱 + P5 自治域（并行）
  ├── strategy/ 包
  └── autonomy/domain.go

Phase 4: 集成 + 测试 + GCL 质量门
```

---

## 9. 参考

- [GCL 规范](docs/gcl-spec.md)
- [Agent Runtime Patterns](docs/agent-runtime-patterns.md)
- [Token Efficiency](docs/token-efficiency.md)
- [Autonomous Ops Roadmap](docs/autonomous-ops-roadmap.md)

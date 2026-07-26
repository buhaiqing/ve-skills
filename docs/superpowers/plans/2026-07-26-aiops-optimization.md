# AIOps 6 大优化实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 ve-skills 仓库实现 6 大 AIOps 优化：可观测性、智能分诊、动态 SLO、自愈编排、策略图谱、自治域。

**Architecture:** 纯 Go 实现，零新外部依赖。6 个新 internal 包 + 4 个现有包增强。按 4 阶段串行执行，每阶段内 2 个任务并行。

**Tech Stack:** Go 1.26, 标准库 (context, sync, math, strings, time, encoding/json, crypto/rand, encoding/hex, unicode, fmt)

---

## 复杂度评估

| 维度 | 分数 | 依据 |
|------|------|------|
| 依赖深度 | 3 | Phase 1→4 有顺序依赖；阶段间可并行 |
| 变更范围 | 3 | 6 新包 + 8 文件修改 |
| 风险系数 | 2 | 内部重构，无破坏性操作 |
| **复杂度** | **18** | 必须拆阶段，每阶段 ≤3 任务 |

---

## Phase 1: P0 可观测性 + P1 智能分诊（并行 batch=2）

### Task P0: observability 包

**Files:**
- Create: `cmd/vet/internal/observability/tracer.go`
- Create: `cmd/vet/internal/observability/metrics.go`
- Create: `cmd/vet/internal/observability/observability_test.go`

- [ ] **Step 1: 创建 tracer.go**

实现 TraceContext + Span + context 传播，包含 NewRootTrace, StartSpan, Span.End, Span.Duration, WithTrace, FromContext, traceKey。

- [ ] **Step 2: 创建 metrics.go**

实现 Counter, Gauge, Histogram + MetricsCollector，包含 Snapshot, Persist 方法。

- [ ] **Step 3: 创建 observability_test.go**

覆盖 NewRootTrace, StartSpan, SpanEnd, SpanDuration, WithTrace/FromContext, Counter, Gauge, Snapshot 测试。

- [ ] **Step 4: 运行测试**

```bash
cd cmd/vet && go test ./internal/observability/ -v -count=1
```

- [ ] **Step 5: Commit**

```bash
git add cmd/vet/internal/observability/
git commit -m "feat: add observability package with tracing and metrics"
```

---

### Task P1: triage 包

**Files:**
- Create: `cmd/vet/internal/triage/tfidf.go`
- Create: `cmd/vet/internal/triage/classifier.go`
- Create: `cmd/vet/internal/triage/triage_test.go`

- [ ] **Step 1: 创建 tfidf.go**

实现 TFIDFVectorizer + cosineSimilarity，包含 defaultTokenizer, buildVocab, buildIDF, transform 方法。

- [ ] **Step 2: 创建 classifier.go**

实现 TriageClassifier + SkillDoc + ClassificationResult，包含 26 个技能文档、DefaultClassifier、Classify、Explain 方法。

- [ ] **Step 3: 创建 triage_test.go**

覆盖 CosineSimilarity, TFIDFVectorizer, TriageClassifier (CPU/Redis/MySQL/VPC/Kafka), TopK, Fallback, ConfidenceRange。

- [ ] **Step 4: 运行测试**

```bash
cd cmd/vet && go test ./internal/triage/ -v -count=1
```

- [ ] **Step 5: Commit**

```bash
git add cmd/vet/internal/triage/
git commit -m "feat: add triage package with TF-IDF classifier"
```

---

## Phase 2: P2 动态 SLO + P3 自愈编排（并行 batch=2）

### Task P2: slo/budget.go

**Files:**
- Create: `cmd/vet/internal/slo/budget.go`
- Create: `cmd/vet/internal/slo/budget_test.go`

- [ ] **Step 1: 创建 budget.go**

实现 ErrorBudget + CalculateBudget + AutoScaleByBudget，分级（healthy/warn/burn/fried）+ 错误预算消耗计算。

- [ ] **Step 2: 创建 budget_test.go**

覆盖 CalculateBudget, CalculateBudgetEmpty, BudgetGrade, AutoScaleByBudget。

- [ ] **Step 3: 运行测试**

```bash
cd cmd/vet && go test ./internal/slo/ -v -count=1
```

- [ ] **Step 4: Commit**

```bash
git add cmd/vet/internal/slo/
git commit -m "feat: add error budget and burn rate to SLO engine"
```

---

### Task P3: heal/orchestrator.go

**Files:**
- Create: `cmd/vet/internal/heal/orchestrator.go`
- Create: `cmd/vet/internal/heal/orchestrator_test.go`

- [ ] **Step 1: 创建 orchestrator.go**

实现 CircuitBreaker + Orchestrator + RecoveryPlan + RecoveryStep，包含 defaultPlans、多步编排、回滚、熔断机制。

- [ ] **Step 2: 创建 orchestrator_test.go**

覆盖 CircuitBreaker (Closed/Opens/Resets/Success), Orchestrator (ExecutePlan/NoPlan/CircuitOpen/Rollback)。

- [ ] **Step 3: 运行测试**

```bash
cd cmd/vet && go test ./internal/heal/ -v -count=1
```

- [ ] **Step 4: Commit**

```bash
git add cmd/vet/internal/heal/
git commit -m "feat: add self-healing orchestrator with circuit breaker"
```

---

## Phase 3: P4 策略图谱 + P5 自治域（并行 batch=2）

### Task P4: strategy 包

**Files:**
- Create: `cmd/vet/internal/strategy/graph.go`
- Create: `cmd/vet/internal/strategy/retrieval.go`
- Create: `cmd/vet/internal/strategy/strategy_test.go`

- [ ] **Step 1: 创建 graph.go**

实现 StrategyNode + StrategyGraph，包含 AddNode、GetPath、GeneratePlan、matchesSymptom。

- [ ] **Step 2: 创建 retrieval.go**

实现 FailurePattern + KnowledgeBase，包含 Query、Learn、AddGraph、GetGraph。

- [ ] **Step 3: 创建 strategy_test.go**

覆盖 StrategyGraph、GeneratePlan、EmptyGraph、KnowledgeBase、LearnUpdate。

- [ ] **Step 4: 运行测试**

```bash
cd cmd/vet && go test ./internal/strategy/ -v -count=1
```

- [ ] **Step 5: Commit**

```bash
git add cmd/vet/internal/strategy/
git commit -m "feat: add strategy graph and knowledge base"
```

---

### Task P5: autonomy/domain.go

**Files:**
- Create: `cmd/vet/internal/autonomy/domain.go`
- Create: `cmd/vet/internal/autonomy/domain_test.go`

- [ ] **Step 1: 创建 domain.go**

实现 AutonomousDomain + Policy + Action + L4Level，包含 AutoExpand、canExpandTo、EvaluateMaturity、CrossCoordinate、Reconcile。

- [ ] **Step 2: 创建 domain_test.go**

覆盖 AutoExpand、CanExpandTo、EvaluateMaturity、CrossCoordinate、Reconcile。

- [ ] **Step 3: 运行测试**

```bash
cd cmd/vet && go test ./internal/autonomy/ -v -count=1
```

- [ ] **Step 4: Commit**

```bash
git add cmd/vet/internal/autonomy/
git commit -m "feat: add autonomous domain expansion and cross-domain coordination"
```

---

## Phase 4: 集成 + 测试 + GCL 质量门

### Task P6: 集成现有代码 + GCL 门

**Files:**
- Modify: `cmd/vet/internal/agent/engine.go`
- Modify: `cmd/vet/internal/agent/triage.go`
- Modify: `cmd/vet/internal/agent/propose.go`
- Modify: `cmd/vet/internal/agent/types.go`
- Modify: `cmd/vet/internal/agentd/dashboard.go`
- Modify: `cmd/vet/internal/slo/engine.go`

- [ ] **Step 1: 集成 engine.go**

在 runLoop 中集成 observability trace，每个步骤创建 Span。

- [ ] **Step 2: 集成 triage.go**

将 Triage 函数改为先调用 triage.DefaultClassifier().Classify()，降级到原硬编码。

- [ ] **Step 3: 集成 propose.go**

在 ProposeFix 中集成 strategy 图谱查询。

- [ ] **Step 4: 集成 slo/engine.go**

在 Observe 方法中调用 CalculateBudget。

- [ ] **Step 5: 全量编译测试**

```bash
cd cmd/vet && go build ./... && go test ./... -v -count=1
```

- [ ] **Step 6: go vet**

```bash
cd cmd/vet && go vet ./...
```

- [ ] **Step 7: CodeGraph sync**

```bash
codegraph sync --quiet
```

- [ ] **Step 8: GCL 质量门（eval-optimize-loop）**

独立 Critic subagent 按 rubric 评审，MAX_ITER=3。

- [ ] **Step 9: Final commit**

```bash
git add -A
git commit -m "feat: integrate all AIOps optimizations with GCL quality gate"
```

---

## 依赖关系图

```
Phase 1 ──→ Phase 2 ──→ Phase 3 ──→ Phase 4
  │            │            │            │
  ├── P0       ├── P2       ├── P4       └── 集成
  └── P1       └── P3       └── P5
   (并行)       (并行)       (并行)
```

## 验收标准

- [ ] `go build ./...` 零错误
- [ ] `go vet ./...` 干净
- [ ] 所有新包 `go test` 通过
- [ ] 集成测试通过
- [ ] GCL 质量门 pass
- [ ] CodeGraph sync 成功

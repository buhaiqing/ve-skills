# AIOps 下一批（P0 #1 + #7）设计规格

> **Date**: 2026-07-31 | **Task**: 持久化 (#1) + 配置化 (#7)
> **For worker agents**: Follow `writing-plans` skill; for multi-step, `dispatching-parallel-agents` or `subagent-orchestrator`; for code, use `source-command-sc-implement` or `general_purpose_task` executor.

---

## 1. 目标

修复两个 P0 级生产就绪度缺口：

1. **持久化缺失** — strategy 知识库的 Learn() 结果和 triage 分类器的 TF-IDF 向量全部 in-memory，重启即丢 / 每次重建。
2. **配置硬编码** — 熔断器 threshold=5 / timeout=30s、分诊置信阈值 0.4（agent 层）、最大重试次数等常量散落各处，无法按环境调整。

---

## 2. 边界

### 2.1 范围内

| 编号 | 功能 | 文件 |
|------|------|------|
| #1a | KnowledgeBase JSON 持久化 (Save/Load) | `cmd/vet/internal/strategy/retrieval.go` + 新增 `persistence.go` + 测试 |
| #1b | DefaultClassifier 单例缓存 (sync.Once) | `cmd/vet/internal/triage/classifier.go` + 测试 |
| #7a | HealConfig + NewOrchestratorWithConfig | `cmd/vet/internal/heal/orchestrator.go` + 测试 |
| #7b | AgentConfig + TriageConfig + LoadConfigFromYAML/Env | 新增 `cmd/vet/internal/agent/config.go` + 测试；`agent/triage.go` 集成；`agent/engine.go` 集成 |

### 2.2 范围外（留到下一批）

- SLO 预算分级边界的配置化（BudgetGrade thresholds）
- triage 阈值不一致（0.05 vs 0.4）的统一决策 → 留到 #7b 集成时暴露但不改默认值
- observability MetricsCollector 的持久化 → 下一批 P2

---

## 3. 状态机 / 契约

### 3.1 KnowledgeBase 持久化契约

```
kb.Save(path) error:
  - 写入 JSON: { patterns: [...], graphs_serialized: map[name]→{nodes:[...]} }
  - 原子写：先写 path.tmp，成功后 os.Rename 覆盖 path
  - patterns[i].LastHit 用 RFC3339Nano 序列化
  - graphs 以节点展开形式序列化为 JSON（graph name → nodes）

kb.Load(path) error:
  - 读取并反序列化 JSON，覆盖内部 patterns 和 graphs
  - 空文件 / 不存在时：返回 nil，kb 保持空状态（非 error）
  - JSON 损坏时：返回 error，kb 状态不变
  - 加载后重建 graphs 的双向指针：遍历 nodes，根据 Children IDs 链接
```

### 3.2 DefaultClassifier 缓存契约

```
var (
    defaultClassifierOnce  sync.Once
    defaultClassifierInst  *TriageClassifier
)

func DefaultClassifier() *TriageClassifier {
    defaultClassifierOnce.Do(func() {
        defaultClassifierInst = NewTriageClassifier(defaultSkills)
    })
    return defaultClassifierInst
}
```

约束：`defaultClassifierInst` 在并发调用下线程安全——NewTriageClassifier 创建的所有字段都是初始化时一次性赋值，后续只读。Classify() 内部无共享可变状态（每次调用构造独立 input 向量），因此可安全并发。

### 3.3 HealConfig 契约

```go
type HealConfig struct {
    CircuitThreshold int           // default 5; <1 时用 default
    CircuitTimeout   time.Duration // default 30s; <1ms 时用 default
}

func DefaultHealConfig() HealConfig { return {5, 30*time.Second} }
func (cfg HealConfig) ApplyDefaults() HealConfig { /* 修正无效值 */ }
func NewOrchestratorWithConfig(cfg HealConfig) *Orchestrator {
    o := NewOrchestrator()
    o.circuit.threshold = cfg.CircuitThreshold
    // timeout 通过 RecordFailure 内部存储：扩展 CircuitBreaker 字段
    return o
}
```

扩展 CircuitBreaker：新增 `timeout time.Duration` 字段（原硬编码 30s），`RecordFailure` 和 `Allow` 使用该字段。

向后兼容：`NewOrchestrator()` 保持不变，使用 `DefaultHealConfig()`。

### 3.4 AgentConfig 契约

```go
type TriageConfig struct {
    ConfidenceThreshold  float64  // 语义分诊采用阈值；default 0.4
    StrategyEnable       bool     // 是否查询 strategy 知识库；default true
    TopK                 int      // 分类 Top-K；default 3
    FallbackSkill        string   // 低置信度回退技能；default "ve-cms-ops"
}

type AgentConfig struct {
    Triage       TriageConfig
    MaxStateRetry int     // SaveState 最大重试次数；default 1
    DryRun        bool    // default false
}

func DefaultAgentConfig() AgentConfig { ... }
func LoadConfigFromYAML(path string) (AgentConfig, error) { ... }
func LoadConfigFromEnv(prefix string) AgentConfig {
    // 读 VETR_AGENT_TRIAGE_CONFIDENCE_THRESHOLD 等
}
```

triage.go 的 `Triage` 函数接受一个可选 config 参数（但为了保持 API 兼容，默认使用 `DefaultAgentConfig()`，同时新增 `TriageWithConfig(payload, cfg)` 作为扩展）。

---

## 4. 文件变更

| 操作 | 路径 |
|------|------|
| **New** | `cmd/vet/internal/strategy/persistence.go` — KnowledgeBase.Save/Load + saveState/loadState 序列化类型 |
| **Modify** | `cmd/vet/internal/strategy/retrieval.go` — 导出 patterns/graphs 的辅助函数用于序列化；保留 mutex 保护 |
| **New** | `cmd/vet/internal/strategy/persistence_test.go` — Save/Load round-trip、空文件、损坏文件、原子写、graph 重建 |
| **Modify** | `cmd/vet/internal/triage/classifier.go` — sync.Once 缓存 DefaultClassifier；不改变签名 |
| **New** | `cmd/vet/internal/triage/cache_test.go` — 并发调用验证 instance identity、单例构造只跑一次 |
| **Modify** | `cmd/vet/internal/heal/orchestrator.go` — 新增 HealConfig、扩展 CircuitBreaker.timeout、NewOrchestratorWithConfig、ApplyDefaults |
| **Modify** | `cmd/vet/internal/heal/orchestrator_test.go` — 新增测试：自定义 threshold 打开时机、自定义 timeout 更快重置、applyDefaults 边界 |
| **New** | `cmd/vet/internal/agent/config.go` — AgentConfig / TriageConfig struct、defaults、LoadFromYAML / LoadFromEnv、TriageWithConfig |
| **Modify** | `cmd/vet/internal/agent/triage.go` — Triage() 内部改为使用 DefaultAgentConfig().Triage；新增 TriageWithConfig() |
| **New** | `cmd/vet/internal/agent/config_test.go` — ApplyDefaults、LoadFromEnv（设置 os.Setenv + 读取）、LoadFromYAML 最小样例 |

---

## 5. 异常边界

| 场景 | 行为 |
|------|------|
| `kb.Save("/no/such/dir/x.json")` | 返回 wrapped error；不写临时文件 |
| `kb.Save(path)` 中 os.Rename 失败 | 保留 .tmp 文件并返回 error |
| JSON 反序列化 types mismatch | 返回 error；kb 状态回滚（先 unmarshal 到临时变量再赋值） |
| HealConfig.CircuitThreshold = 0 | ApplyDefaults() 重置为 5 |
| TriageConfig.ConfidenceThreshold > 1.0 | ApplyDefaults() 钳制为 1.0；< 0.0 → 0.0 |
| LoadConfigFromYAML 中字段缺失 | 对应字段填默认值（ApplyDefaults 后合并） |
| 并发调用 DefaultClassifier() | sync.Once 保证单例；Classify() 方法可并发安全 |

---

## 6. 验收标准

- [ ] `go build ./...` 零错误
- [ ] `go vet ./...` 干净
- [ ] 新增测试：strategy/persistence_test.go (Save/Load roundtrip, empty file, corrupt file, roundtrip with 5 patterns, graph rebuild)
- [ ] 新增测试：triage/cache_test.go (identity preserved, concurrent call identity)
- [ ] 新增测试：heal TestHealConfigDefaults、TestOrchestratorCustomThreshold、TestOrchestratorCustomTimeout
- [ ] 新增测试：agent/config_test.go (defaults, loadFromEnv override, loadFromYAML merge)
- [ ] 向后兼容：`Triage(payload)` 行为不变；`NewOrchestrator()` 行为不变（均走 DefaultXxxConfig）
- [ ] Critic PASS: R1 build/test, R2 安全(无文件注入路径), R3 幂等 (Save/Load/Save roundtrip same bytes / same objects)

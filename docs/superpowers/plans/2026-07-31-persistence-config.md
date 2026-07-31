# P0 #1 持久化 + #7 配置化 执行计划

> **For worker agents:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans. This plan has **two independent batches**: Batch A (#1a + #1b persistence) and Batch B (#7a + #7b configuration) can run in parallel.

**Goal**: 解决 2 个 P0 生产就绪度缺口：策略/知识库持久化 + 硬编码配置化。

**Strategy**: 混合批处理 — Batch A 与 Batch B 完全独立（涉及不同包、无文件重叠），可并行执行；Batch 内按依赖顺序（单包内文件修改）。

```
╔══════════════════════════════════════════════════════════╗
║  AIOps 编排图 — Batch A + B (并行)                        ║
╠══════════════════════════════════════════════════════════╣
║                                                           ║
║  ╔══ Batch A: 持久化 ══════════════════════════════╗     ║
║  ║                                                    ║     ║
║  ║  Worker A1                                       ║     ║
║  ║  ├─ strategy/persistence.go (Save/Load)           ║     ║
║  ║  ├─ strategy/retrieval.go (辅助函数导出)          ║     ║
║  ║  ├─ strategy/persistence_test.go (roundtrip 等)   ║     ║
║  ║  └─ go test ./internal/strategy/                  ║     ║
║  ║                                                    ║     ║
║  ║  Worker A2                                       ║     ║
║  ║  ├─ triage/classifier.go (sync.Once cache)        ║     ║
║  ║  ├─ triage/cache_test.go (identity + concurrent)  ║     ║
║  ║  └─ go test ./internal/triage/                    ║     ║
║  ╚════════════════════════════════════════════════════╝     ║
║                          ┃ PARALLEL ┃                       ║
║  ╔══ Batch B: 配置化 ══════════════════════════════╗     ║
║  ║                                                    ║     ║
║  ║  Worker B1                                       ║     ║
║  ║  ├─ heal/orchestrator.go (HealConfig + CB.timeout)║     ║
║  ║  ├─ heal/orchestrator_test.go (3 个新测试)        ║     ║
║  ║  └─ go test ./internal/heal/                      ║     ║
║  ║                                                    ║     ║
║  ║  Worker B2                                       ║     ║
║  ║  ├─ agent/config.go (AgentConfig + 加载器)        ║     ║
║  ║  ├─ agent/triage.go (TriageWithConfig)            ║     ║
║  ║  ├─ agent/config_test.go (defaults + yaml + env)  ║     ║
║  ║  └─ go test ./internal/agent/                     ║     ║
║  ╚════════════════════════════════════════════════════╝     ║
║                        ┃ JOIN ┃                             ║
║  ╔══ Phase 3: 全量验证 + GCL Critic ═══════════════╗      ║
║  ║  go build ./... + go vet ./... + go test ./...   ║      ║
║  ║  GCL Critic 独立只读评审 (R1/R2/R3)              ║      ║
║  ╚═══════════════════════════════════════════════════╝      ║
╚══════════════════════════════════════════════════════════╝
```

---

## Batch A — 持久化（并行独立）

### Task A1: strategy KnowledgeBase 持久化

**Files:**
- Create: `cmd/vet/internal/strategy/persistence.go`
- Modify: `cmd/vet/internal/strategy/retrieval.go`
- Create: `cmd/vet/internal/strategy/persistence_test.go`

**Plan:**

- [ ] **Step 1: Create persistence.go**
  - `savePatterns(kb *KnowledgeBase) []patternState` — serializes patterns (LastHit → RFC3339Nano string)
  - `loadPatterns(states []patternState) []FailurePattern` — 反序列化
  - `saveGraphs(kb *KnowledgeBase) map[string][]nodeState` — 展开 graph 节点为 nodeState 数组 (ID, Symptom, Operation, Condition, Priority, Enabled, ChildrenIDs)
  - `loadGraphs(saved map[string][]nodeState) map[string]*StrategyGraph` — 反序列化并重建 Children 指针链接（先建 map[ID]*StrategyNode，再链接）
  - `kbState` 容器 struct 带 `json:"patterns"` / `json:"graphs"` tag
  - `(kb *KnowledgeBase) Save(path string) error`:
    - marshal kbState → JSON bytes
    - tmpPath := path + ".tmp"
    - ioutil.WriteFile(tmpPath, bytes, 0644) → os.Rename(tmpPath, path)
  - `(kb *KnowledgeBase) Load(path string) error`:
    - 若 os.Stat(path) 显示文件不存在 → return nil (kb stays empty, NOT error)
    - ReadFile → unmarshal 到临时 `kbState`
    - 反序列化 patterns + graphs 到临时变量
    - kb.mu.Lock() → 赋值到 kb.patterns + kb.graphs → Unlock()

- [ ] **Step 2: Modify retrieval.go**（如需要）
  - 确认 fields 为 lowercase 的情况下 persistence.go 同包可直接访问（都是 `strategy` 包）

- [ ] **Step 3: Create persistence_test.go**
  - `TestSaveLoadRoundTrip`: kb.Learn 3 patterns + AddGraph + 2 graph nodes → Save → Load new kb → Query & GetGraph 结果一致
  - `TestSaveLoadEmptyKB`: 空 kb Save → Load → patterns 为空，graphs 为空
  - `TestLoadNoFile`: Load("non-existent-path.json") → err == nil，kb 状态空
  - `TestLoadCorruptFile`: WriteFile(path, []byte("not json")) → Load → err != nil，kb 不变
  - `TestSaveBadDir`: Save("/nonexistent/foo/x.json") → err != nil
  - `TestGraphRebuildChildren`: graph 3 个节点有父子关系 → save → load → children pointers correctly linked (same IDs)

- [ ] **Step 4: 运行测试**
```bash
cd cmd/vet && go test ./internal/strategy/ -v -count=1
```

---

### Task A2: triage DefaultClassifier 单例缓存

**Files:**
- Modify: `cmd/vet/internal/triage/classifier.go`
- Create: `cmd/vet/internal/triage/cache_test.go`

**Plan:**

- [ ] **Step 1: Modify classifier.go**
  - 新增包级变量：
    ```go
    var (
        defaultClassifierOnce sync.Once
        defaultClassifierInst *TriageClassifier
    )
    ```
  - 修改 `DefaultClassifier()`：用 `.Do(func(){ inst = NewTriageClassifier(defaultSkills) })`

- [ ] **Step 2: Create cache_test.go**
  - `TestDefaultClassifierIdentity`: 100 次循环调用 → 返回指针相同 (`==`)
  - `TestDefaultClassifierConcurrent`: 100 goroutines 同时取 → 全部指向同一实例；无 data race (go test -race -count=1 验证)
  - `TestDefaultClassifierCorrectness`: 分类结果与 NewTriageClassifier(defaultSkills) 结果一致（相同输入 → 相同 Top-3）

- [ ] **Step 3: 运行测试**
```bash
cd cmd/vet && go test -race ./internal/triage/ -v -count=1
```

---

## Batch B — 配置化（并行独立）

### Task B1: HealConfig + Orchestrator 自定义配置

**Files:**
- Modify: `cmd/vet/internal/heal/orchestrator.go`
- Modify: `cmd/vet/internal/heal/orchestrator_test.go`

**Plan:**

- [ ] **Step 1: Extend CircuitBreaker struct**
  - 新增 `timeout time.Duration` 字段（原 resetTime = time.Now().Add(30s) → 改为 Add(cb.timeout)）

- [ ] **Step 2: Add HealConfig + helpers**
  ```go
  type HealConfig struct {
      CircuitThreshold int
      CircuitTimeout   time.Duration
  }
  func DefaultHealConfig() HealConfig { return {5, 30*time.Second} }
  func (c HealConfig) ApplyDefaults() HealConfig {
      if c.CircuitThreshold < 1 { c.CircuitThreshold = 5 }
      if c.CircuitTimeout < time.Millisecond { c.CircuitTimeout = 30 * time.Second }
      return c
  }
  func NewOrchestratorWithConfig(cfg HealConfig) *Orchestrator {
      cfg = cfg.ApplyDefaults()
      o := NewOrchestrator()
      o.circuit.threshold = cfg.CircuitThreshold
      o.circuit.timeout = cfg.CircuitTimeout
      return o
  }
  ```
- [ ] **Step 3: Backward compat NewOrchestrator()**
  - 保持其内部逻辑不变，但设置 `cb.timeout = 30s`；等价于 NewOrchestratorWithConfig(DefaultHealConfig())

- [ ] **Step 4: Add tests to orchestrator_test.go**
  - `TestHealConfigDefaults`: zero-value HealConfig.ApplyDefaults() == DefaultHealConfig()
  - `TestOrchestratorCustomThreshold`: threshold=2 → 2 次 failure 后打开（原 test 用 threshold=5）
  - `TestOrchestratorCustomTimeout`: timeout=10ms → 打开后 10ms 即自动恢复（比原 30s 更短）

- [ ] **Step 5: 运行测试**
```bash
cd cmd/vet && go test ./internal/heal/ -v -count=1
```

---

### Task B2: AgentConfig + 加载器 + TriageWithConfig

**Files:**
- Create: `cmd/vet/internal/agent/config.go`
- Modify: `cmd/vet/internal/agent/triage.go`
- Create: `cmd/vet/internal/agent/config_test.go`

**Plan:**

- [ ] **Step 1: Create config.go**
  - 定义 `TriageConfig`、`AgentConfig` struct
  - `DefaultAgentConfig()` 返回默认值
  - `(a *AgentConfig) ApplyDefaults()` — 钳制阈值 [0,1]，TopK < 1 → 3，FallbackSkill 空 → "ve-cms-ops"
  - `LoadConfigFromYAML(path string) (AgentConfig, error)`：
    - 读取 yaml 到临时 cfg
    - 不存在或空：返回 DefaultAgentConfig(), nil
    - ApplyDefaults() 合并
  - `LoadConfigFromEnv(prefix string) AgentConfig`：
    - 查 os.Getenv(prefix + "_TRIAGE_CONFIDENCE_THRESHOLD") 等，解析 float/int/bool/string
    - parse 失败时 fall back 到 default
  - `TriageWithConfig(payload *IncidentPayload, cfg TriageConfig) *TriageResult`：
    - 使用 cfg.ConfidenceThreshold / cfg.TopK / cfg.FallbackSkill
    - cfg.StrategyEnable 预留（Propose 层用，但函数签名里不强制使用）

- [ ] **Step 2: Modify triage.go**
  - `Triage(payload)` 实现改为：
    ```go
    func Triage(payload *IncidentPayload) *TriageResult {
        return TriageWithConfig(payload, DefaultAgentConfig().Triage)
    }
    ```

- [ ] **Step 3: Create config_test.go**
  - `TestDefaultAgentConfig`: defaults 符合 spec (0.4, true, 3, "ve-cms-ops")
  - `TestApplyDefaultsBoundary`: Confidence=1.5 → 1.0; Confidence=-0.2 → 0.0; TopK=0 → 3
  - `TestLoadFromEnv`: os.Setenv("VETR_AGENT_TRIAGE_CONFIDENCE_THRESHOLD", "0.7") → LoadConfigFromEnv("VETR_AGENT").Triage.ConfidenceThreshold == 0.7
  - `TestLoadFromYAML`: 创建 tempfile 写最小 YAML（只用一个字段覆盖）→ 加载后覆盖字段正确，其他为 default
  - `TestTriageWithConfigLowThreshold`: threshold=0.01 → 更激进采用语义匹配结果；对比默认阈值的不同之处（或至少无 panic）

- [ ] **Step 4: 运行测试**
```bash
cd cmd/vet && go test ./internal/agent/ -v -count=1
```

---

## Phase 3 — 全量验证 + GCL 质量门

在 Batch A + B 都完成后，串行执行：

- [ ] **Final Build**
```bash
cd cmd/vet && go build ./... && go vet ./... && go test ./... -count=1
```

- [ ] **GCL Critic Review**
  - 独立 Critic subagent 只读验证：R1 build/test pass、R2 安全(文件写入路径注入、环境变量注入)、R3 幂等 (Save/Load/Save 等价)、R4 可追溯 (Config.ApplyDefaults 日志？可无，但行为要确定)、R5 规格符合（默认值、阈值、单例正确性）

- [ ] **CodeGraph sync**
```bash
codegraph sync --quiet
```

---

## 依赖关系

```
Batch A          Batch B
  │                 │
  ├── A1 (strategy) ├── B1 (heal)
  └── A2 (triage)   └── B2 (agent)
       │ (independent) │
       └──────┬────────┘
              ▼
        Phase 3: Verify + GCL
```

## 验收标准（DoD）

- [ ] `go build ./...` 零错误
- [ ] `go vet ./...` 干净
- [ ] A1: 6 个 strategy persistence 测试全绿
- [ ] A2: 3 个 triage cache 测试全绿（含 -race）
- [ ] B1: 3 个 heal config 测试全绿
- [ ] B2: 5 个 agent config 测试全绿，原有 agent 测试不回归
- [ ] GCL Critic: PASS (0 BLOCKER)
- [ ] CodeGraph sync 成功

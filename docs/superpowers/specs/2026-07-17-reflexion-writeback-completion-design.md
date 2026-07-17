# Reflexion Write-back 链路补全 (SDD Spec)

日期：2026-07-17
对应：`docs/reflexion-memory.md` + `.runtime/memory/` 体系

---

## 1. 功能描述

### 1.1 核心目标

补全 Reflexion 的 write-back 链路，使 GCL 执行轨迹中的 failure pattern 能够：
1. **结构化持久化**：写入 `.runtime/memory/failure-patterns.json`（JSON，带 count）
2. **count 递增**：同一 pattern 重复出现时 count++，而非覆盖
3. **自动触发 T13**：count ≥ 10 时自动调用 T13 转译器生成 guardrails.yaml
4. **Pre-flight 优化**：GCL 执行前从 JSON 加载已知 patterns，而非逐行扫描 markdown

### 1.2 数据流（补全后）

```
GCL 执行失败
    │
    ├── writebackFailurePattern()
    │       │
    │       ├── .runtime/memory/failure-patterns.json  ← 结构化持久化，count++
    │       │       │
    │       │       │  count ≥ 10
    │       │       ▼
    │       │     T13 transpile → guardrails.yaml  ← 自动触发
    │       │
    │       └── docs/failure-patterns.md  ← 保持兼容（markdown 表）
    │
    └── loadKnownFailurePatterns()  ← 从 JSON 加载（结构化查询）
            │
            └── 注入 Generator context（HINT，不强制）
```

---

## 2. API / 组件契约

### 2.1 新增 Go 包：`cmd/vet/internal/memory/store.go`

```go
package memory

type FailurePatternEntry struct {
    Skill    string `json:"skill"`
    Pattern  string `json:"pattern"`
    Category string `json:"category"`
    Fix      string `json:"fix"`
    Source   string `json:"source"`
    Count    int    `json:"count"`
}

// AppendFailurePattern 追加或更新一条 failure pattern。
// 如果 (skill, pattern) 已存在 → count++；否则 → 新增 count=1。
func AppendFailurePattern(root string, entry FailurePatternEntry) error

// LoadFailurePatterns 加载所有 failure patterns。
func LoadFailurePatterns(root string) ([]FailurePatternEntry, error)

// GetPatternsBySkill 按 skill 过滤 patterns。
func GetPatternsBySkill(root, skill string, limit int) ([]FailurePatternEntry, error)
```

### 2.2 修改 `run.go:writebackFailurePattern`

- 保持原有 markdown 写回逻辑
- 新增：同时调用 `memory.AppendFailurePattern(root, entry)` 写入 JSON
- 新增：如果更新后 count ≥ 10，触发 T13 transpile

### 2.3 修改 `run.go:loadKnownFailurePatterns`

- 改为从 `memory.LoadFailurePatterns(root)` 加载
- 按 skill 过滤，返回 top-N 作为 HINT 注入 Generator context

---

## 3. 异常边界

| 场景 | 处理 |
|------|------|
| `.runtime/memory/failure-patterns.json` 不存在 | 创建空文件，追加 entry |
| JSON 格式损坏 | WARN 日志，降级到 markdown 写回 |
| T13 transpile 失败 | WARN 日志，不阻塞主流程 |

---

## 4. 验收标准

- [ ] `cmd/vet/internal/memory/store.go` 实现 Append + Load + GetBySkill
- [ ] `store_test.go` 覆盖：新增、count++、按 skill 过滤、空文件
- [ ] `run.go` 修改 writebackFailurePattern：追加 JSON 写入
- [ ] `run.go` 修改 loadKnownFailurePatterns：从 JSON 加载
- [ ] count ≥ 10 时自动调用 T13 transpile
- [ ] `go build/vet/test` 全绿

# Agent 开发模板 (Agent Template)

> 本文档提炼自 Phase 1 Agent Runtime 开发实战，包含可复用的模式、契约和开发流程。
> 每个新的 Agent 功能模块开发时，以此为起点。

---

## 1. 开发流程模板

### 1.1 Spec + Plan First（铁律）

```
Step 1: 写 Spec  → docs/superpowers/specs/<date>-<feature>-design.md
Step 2: 写 Plan  → docs/superpowers/plans/<date>-<feature>.md
Step 3: 进入 Plan Mode → 用户审批
Step 4: GCL 并行开发 → Generator(s) + Critic (跨厂商)
Step 5: 主 Agent 修复 + CLI 集成 + 测试
Step 6: 复盘 + 存档经验
```

### 1.2 GCL 配置模板

```
╔══════════════════════════════════════════╗
║      Generator-Critic-Loop 模型配置       ║
╠══════════════════════════════════════════╣
║ Generator A: [default] (编码)             ║
║ Generator B: [default] (编码) — 可选     ║
║ Critic:     [reasoning] (跨厂商审计)      ║
║ 并行: 独立模块可并行 Generator             ║
║ 最大轮次: 2                               ║
╚══════════════════════════════════════════╝
```

---

## 2. 代码组织模板

### 2.1 目录结构

```
cmd/vet/internal/<feature>/
├── types.go          # 共享数据结构（最先写）
├── <step1>.go        # 按功能拆文件，每个文件一个核心函数
├── <step2>.go
├── engine.go         # 编排/状态机
├── state.go          # 状态持久化（如果需要断点续跑）
├── <feature>.go      # CLI 层（cmd/vet/ 下，package main）
└── <feature>_test.go # 测试
```

### 2.2 文件命名规则

| 文件类型 | 命名 | 示例 |
|---------|------|------|
| 数据类型 | `types.go` | Step/IncidentPayload/RunResult |
| 输入层 | `ingest.go` / `parse.go` | ParseJSON / ParseNaturalLanguage |
| 业务逻辑 | 按步骤名 | `triage.go` / `diagnose.go` / `propose.go` |
| 编排层 | `engine.go` | Run() 状态机 |
| 持久化 | `state.go` | SaveState / LoadState |
| CLI 层 | `<feature>.go` (package main) | `agent.go` |
| 测试 | `<feature>_test.go` | `engine_test.go` |

---

## 3. 复用清单模板

开发新功能时，**必须先检查**以下现有基础设施是否可用：

| 基础设施 | 导入路径 | 核心函数 | 检查项 |
|---------|---------|---------|--------|
| GCL Runner | `internal/gcl/run` | `Run(Options) Result` | 是否可以直接调用？ |
| Policy Loader | `internal/policy` | `Load(root) (*PolicySet, error)` | 是否需要决策矩阵？ |
| Memory Store | `internal/memory` | `AppendFailurePattern(root, entry)` | 是否需要持久化 failure pattern？ |
| Transpiler | `internal/reflexion/transpile` | `Transpile(FailurePattern) (Guardrail, bool)` | 是否需要 pattern→policy 转译？ |
| Promote | `internal/reflexion/promote` | `LevelOf(Pattern) Level` | 是否需要 count→Level 分级？ |
| Log Package | `internal/log` | `Append(path, runID, level, component, msg, kvs...)` | 是否需要结构化日志？ |
| Heal | `internal/gcl/heal` | `Classify(errorStr) ErrorClass` | 是否需要错误分类？ |
| Trace | `internal/gcl/trace` | `PersistTrace(root, path, t)` | 是否需要 GCL trace？ |
| Secret | `internal/gcl/secret` | `MaskSecrets(text) string` | 是否需要脱敏？ |

**铁律：禁止重新实现上述任何功能。必须复用。**

---

## 4. 日志模板

### 4.1 结构化日志格式（AGENTS.md 铁律）

```
<ISO_8601_ts> | [<run_id>] | <level> | <component> | <message> | <key=value>...
```

### 4.2 关键日志点

```go
// 启动
fmt.Fprintf(os.Stderr, "[%s] [INFO] <component> | start | <k=v...>\n", runID)

// 每步开始
fmt.Fprintf(os.Stderr, "[%s] [INFO] <component> | <step> start | <context>\n", runID)

// 每步结束
fmt.Fprintf(os.Stderr, "[%s] [INFO] <component> | <step> done | <result>\n", runID)

// 错误
fmt.Fprintf(os.Stderr, "[%s] [ERROR] <component> | <step> failed | %v\n", runID, err)

// 终止
fmt.Fprintf(os.Stderr, "[%s] [INFO] <component> | complete | success=%v step=%s\n", runID, success, finalStep)
```

### 4.3 日志 helper 模板

```go
func logStep(runID, step, event, format string, args ...interface{}) {
    fmt.Fprintf(os.Stderr, "[%s] [INFO] <pkg>.engine | %s %s", runID, step, event)
    if format != "" {
        fmt.Fprintf(os.Stderr, " | "+format, args...)
    }
    fmt.Fprintln(os.Stderr)
}

func logError(runID, step, format string, args ...interface{}) {
    fmt.Fprintf(os.Stderr, "[%s] [ERROR] <pkg>.engine | %s ", runID, step)
    fmt.Fprintf(os.Stderr, format, args...)
    fmt.Fprintln(os.Stderr)
}
```

---

## 5. 状态持久化模板

### 5.1 RunState 结构

```go
type RunState struct {
    RunID       string        `json:"run_id"`
    CurrentStep int           `json:"current_step"`
    Payload     InputPayload  `json:"payload"`
    // ... 每步的中间结果
    Error       string        `json:"error,omitempty"`
}
```

### 5.2 持久化函数

```go
func runDir(root, runID string) string {
    return filepath.Join(root, ".runtime", "<feature>", "runs", runID)
}

func SaveState(root string, state *RunState) error {
    dir := runDir(root, state.RunID)
    os.MkdirAll(dir, 0o755)
    data, _ := json.MarshalIndent(state, "", "  ")
    return os.WriteFile(filepath.Join(dir, "state.json"), data, 0o644)
}

func LoadState(root, runID string) (*RunState, error) {
    // 文件不存在返回 nil, nil
}
```

---

## 6. CLI 集成模板

### 6.1 子命令注册 (main.go)

```go
case "<feature>":
    run<Feature>(args)
```

### 6.2 子命令文件 (<feature>.go)

```go
package main

func run<Feature>(args []string) {
    // dispatch sub-subcommands
}

func run<Feature>Run(args []string) {
    fs := flag.NewFlagSet("<feature> run", flag.ExitOnError)
    root := fs.String("root", repoRoot(), "repo root")
    // ... flags
    fs.Parse(args)
    // ... call internal package
}
```

### 6.3 usage() 更新

```
vet <feature> <run|status|resume> [flags]
```

---

## 7. 安全清单模板

每个新功能上线前必须通过：

| 检查项 | 状态 |
|--------|------|
| 命令注入防护（shell 元字符检测） | [ ] |
| 凭据脱敏（MaskSecrets 调用） | [ ] |
| 破坏性操作确认门（SafetyClass 检查） | [ ] |
| 错误不静默丢弃（至少 logError） | [ ] |
| 每步有 checkpoint（断点续跑） | [ ] |
| 结构化日志（[runID] [LEVEL] 格式） | [ ] |
| 复用现有基础设施（不重新实现） | [ ] |
| go build/vet/test 全绿 | [ ] |

---

## 8. Spec 模板

```markdown
# <Feature> (SDD Spec)

日期：YYYY-MM-DD
对应架构：`docs/<architecture>.md`

## 1. 功能描述
### 1.1 核心目标
### 1.2 状态机 / 数据流

## 2. 异常边界
| 场景 | 处理 |
|------|------|

## 3. API / 组件契约

## 4. 验收标准
- [ ] ...
```

## 9. Plan 模板

```markdown
# <Feature> 执行计划

日期：YYYY-MM-DD
对应 Spec：`docs/superpowers/specs/<date>-<feature>-design.md`

## 里程碑拆分
| # | 任务 | 文件 | 验证点 |
|---|------|------|--------|

## 依赖图

## 并行策略

## 新增/修改文件

## 验证命令
```

---

## 10. 实战产物：可复用的代码骨架

### 10.1 7 步编排引擎骨架

从 `cmd/vet/internal/agent/engine.go` 提炼：

```go
// Run executes a multi-step loop with checkpoint at each step.
// Adapt this pattern for any state-machine-based feature.
func Run(root string, input *InputPayload, runID string) *RunResult {
    state := &RunState{RunID: runID, CurrentStep: Step1, Payload: *input}

    logStep(runID, "STEP1", "start", "...")
    result1, err := doStep1(input)
    if err != nil {
        logError(runID, "STEP1", "failed: %v", err)
        return &RunResult{Success: false, Error: err.Error()}
    }
    state.Result1 = result1
    if err := SaveState(root, state); err != nil {
        logError(runID, "STEP1", "save failed: %v", err)
        return &RunResult{Success: false, Error: err.Error()}
    }
    logStep(runID, "STEP1", "done", "...")

    // ... repeat for each step ...

    logStep(runID, "COMPLETE", "done", "success=true")
    return &RunResult{Success: true}
}
```

### 10.2 自然语言输入解析器

从 `cmd/vet/internal/agent/ingest.go` 提炼：

```go
// Keyword-to-output mapping with deterministic iteration.
var keywordMap = map[string]string{
    "keyword1": "output1",
    "关键词2":  "output2",
}

// sortedKeywords ensures deterministic iteration order.
// Longer keywords match first; alphabetical tiebreaker for same length.
func sortedKeywords() []struct{ key, value string } {
    var entries []struct{ key, value string }
    for k, v := range keywordMap {
        entries = append(entries, struct{ key, value string }{k, v})
    }
    sort.Slice(entries, func(i, j int) bool {
        if len(entries[i].key) != len(entries[j].key) {
            return len(entries[i].key) > len(entries[j].key)
        }
        return entries[i].key < entries[j].key
    })
    return entries
}

// Parse extracts structured output from free-text input.
func Parse(input string) (*Output, error) {
    lower := strings.ToLower(strings.TrimSpace(input))
    if lower == "" {
        return nil, fmt.Errorf("input is empty")
    }
    for _, entry := range sortedKeywords() {
        if strings.Contains(lower, strings.ToLower(entry.key)) {
            return &Output{Hint: entry.value}, nil
        }
    }
    return nil, fmt.Errorf("no matching keyword found")
}
```

### 10.3 双 Generator 并行 Prompt 模板

```markdown
## Generator A Prompt 结构
- 明确指定要写的文件和模块路径
- 列出共享类型（types.go）作为对齐点
- 给出完整的函数签名和 import 路径
- 标注复用点（"CRITICAL — do NOT reimplement these"）
- 明确依赖边界（"Do NOT modify any files outside <dir>"）
- 结尾：`go build ./...` 必须通过

## Generator B Prompt 结构
- 引用 Generator A 已写的文件（"Generator A has already written types.go..."）
- 给出需要导入的包的完整路径
- 给出函数签名 + 复用点
- 同样约束修改边界
```

### 10.4 Critic 审计 Prompt 模板

```markdown
## Critic Audit Prompt 结构
- 列出所有待审查文件
- 定义评分维度（Correctness/Safety/Idempotency/Traceability/Spec Compliance）
- 逐个检查点（Specific checks 列表）
- 要求输出分级（CRITICAL/WARNING/SUGGESTION）
- 要求每个 finding 附带：文件 + 行号 + 问题描述 + 修复建议
- 结尾要求 `go build/vet/test` 验证
```

### 10.5 测试模板

从 `cmd/vet/internal/agent/agent_test.go` 提炼：

```go
func TestParse_ValidInput(t *testing.T) {
    result, err := Parse("valid input")
    if err != nil {
        t.Fatalf("Parse failed: %v", err)
    }
    if result.Field != "expected" {
        t.Errorf("expected %q, got %q", "expected", result.Field)
    }
}

func TestParse_InvalidInput(t *testing.T) {
    _, err := Parse("")
    if err == nil {
        t.Error("expected error for empty input, got nil")
    }
}

func TestParse_DeterministicOutput(t *testing.T) {
    // Run N times to verify deterministic behavior
    for i := 0; i < 100; i++ {
        result, err := Parse("ambiguous input")
        if err != nil {
            t.Fatalf("iteration %d failed: %v", i, err)
        }
        if result.Field != "expected" {
            t.Errorf("iteration %d: expected %q, got %q", i, "expected", result.Field)
        }
    }
}

func TestStateMachine_StepOrder(t *testing.T) {
    // Verify steps execute in correct order
    // Verify checkpoint at each step
    // Verify error propagation
}
```

### 10.6 Dry-run 模式

从 `cmd/vet/agent.go:runAgentRun` 提炼：

```go
func runFeatureRun(args []string) {
    // ... flag parsing ...
    if *dryRun {
        // Run all non-destructive steps, log results, skip actual execution
        fmt.Fprintf(os.Stderr, "[%s] [INFO] feature.run | dry-run mode\n", runID)
        step1 := doStep1(input)
        fmt.Fprintf(os.Stderr, "[%s] [INFO] feature.run | step1 | result=%v\n", runID, step1)
        // ... all steps ...
        fmt.Fprintf(os.Stderr, "[%s] [INFO] feature.run | dry-run complete\n", runID)
        return
    }
    // Real execution path
}
```

### 10.7 错误处理模式

```go
// Pattern 1: Return error immediately, log to caller
func doStep(input *Input) (*Output, error) {
    result, err := externalCall(input)
    if err != nil {
        return nil, fmt.Errorf("step failed: %w", err)
    }
    return result, nil
}

// Pattern 2: Log warning but continue (best-effort)
func optionalWriteback(data *Data) {
    if err := store.Write(data); err != nil {
        logError(runID, "WRITEBACK", "failed: %v", err)
        // Continue — this is best-effort
    }
}

// Pattern 3: Fail-safe default
func evaluate(policy *Policy) Decision {
    if policy == nil {
        // Missing policy → fail-safe (most restrictive)
        return DecisionDeny
    }
    return policy.Decide()
}
```

### 10.8 runID 生成

```go
import "time"

// Generate a short hex run ID from the current nanosecond timestamp.
// 8 hex chars = enough to distinguish concurrent runs.
func newRunID() string {
    return fmt.Sprintf("%08x", time.Now().UnixNano()%0x100000000)
}
```

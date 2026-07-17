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

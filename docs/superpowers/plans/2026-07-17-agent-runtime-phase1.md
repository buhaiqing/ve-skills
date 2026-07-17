# Phase 1 — Agent Runtime 执行计划（最终版）

日期：2026-07-17
对应 Spec：`docs/superpowers/specs/2026-07-17-agent-runtime-phase1-design.md`
对应架构：`docs/vet-agent-evolution-architecture.md`

---

## 目标

把 `incident-loop-agent/SKILL.md` 的 7 步编排变成可执行 Go 代码：
```
INGEST → TRIAGE → DIAGNOSE → PROPOSE → CONFIRM → EXECUTE → REFLEXION
```

## 里程碑

### M1 — 数据模型 + Ingest（无依赖）

| # | 任务 | 文件 | 说明 |
|---|------|------|------|
| T1 | `types.go`: IncidentPayload, RunState, Step 枚举 | `internal/agent/types.go` | 共享数据结构 |
| T2 | `ingest.go`: ParseJSON/ParseNaturalLanguage/ParseJIRA | `internal/agent/ingest.go` | 多通道输入标准化 |

### M2 — Triage + Diagnose（依赖 M1）

| # | 任务 | 文件 | 说明 |
|---|------|------|------|
| T3 | `triage.go`: 路由图匹配 + 降级 | `internal/agent/triage.go` | 读 skill-routing-graph.md |
| T4 | `diagnose.go`: 并行 Describe* | `internal/agent/diagnose.go` | cap 15 次 |

### M3 — Propose + Confirm + Execute（依赖 M2）

| # | 任务 | 文件 | 说明 |
|---|------|------|------|
| T5 | `propose.go`: DispatchPlan + GCL | `internal/agent/propose.go` | 复用 vet gcl run |
| T6 | `confirm.go`: 策略评估 | `internal/agent/confirm.go` | 复用 policy.Load() |
| T7 | `execute.go`: 委托执行 + heal | `internal/agent/execute.go` | 复用 vet gcl run |

### M4 — Engine + State + Reflexion（依赖 M3）

| # | 任务 | 文件 | 说明 |
|---|------|------|------|
| T8 | `engine.go`: 7 步状态机 | `internal/agent/engine.go` | Run(payload) → Result |
| T9 | `state.go`: RunState 持久化 | `internal/agent/state.go` | 断点续跑 |
| T10 | `reflexion.go`: 写回 failure-patterns | `internal/agent/reflexion.go` | 复用 memory.Append |

### M5 — CLI + 测试（依赖 M4）

| # | 任务 | 文件 | 说明 |
|---|------|------|------|
| T11 | `cmd/vet/agent.go`: CLI 注册 | `agent.go` + `main.go` | vet agent run/resume/status |
| T12 | `engine_test.go`: 集成测试 | `internal/agent/engine_test.go` | mock payload 端到端 |

### M6 — 收尾

| # | 任务 | 说明 |
|---|------|------|
| T13 | `go build/vet/test` 全绿 | 全仓库 |
| T14 | 端到端验证 | `vet agent run --payload '{"product":"ecs","symptom":"cpu>90%"}' --dry-run` |

## 并行策略

```
M1 (T1→T2) ──┐
              ├── M2 (T3→T4) ── M3 (T5→T6→T7) ── M4 (T8→T9→T10) ── M5 (T11→T12) ── M6 (T13→T14)
              │
M1 可 1 Agent 完成；M2-M3 可 1 Agent 完成；M4 可 1 Agent 完成
或：T1-T4 1 Agent，T5-T10 1 Agent 并行
```

**推荐并行策略**：
- Agent A: M1+M2（types + ingest + triage + diagnose）→ 纯数据模型和解析
- Agent B: M3+M4（propose + confirm + execute + engine + state + reflexion）→ 编排逻辑
- 两者在 M1 的 types.go 上对齐后并行开发

## 关键设计约束

1. **所有文件在 `cmd/vet/internal/agent/` 下，一个 Go package**
2. **不修改任何现有 skill 的 SKILL.md**
3. **复用现有基础设施**：`vet gcl run`、`policy.Load()`、`memory.AppendFailurePattern()`、`docs/skill-routing-graph.md`
4. **IncidentPayload 统一结构**：`ProductHint` + `Symptom` + `RawInput`

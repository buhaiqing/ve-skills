# Vet Agent 演进架构 — 从 CLI 工具到自主 Agent 引擎

> 本文档描述 `vet` CLI 工具从当前"被动命令行工具"演进到"主动 Agent 执行引擎"的完整路线图。
> 三个 Phase 逐步递进，每个 Phase 有独立的 Spec 和 Plan。

---

## 1. 当前架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        vet CLI (被动调用)                         │
│                                                                  │
│  $ vet gcl run --skill ve-ecs-ops --cmd "StopInstance ..."       │
│  $ vet check frontmatter --root .                                │
│  $ vet policy check-changelog --root .                           │
│                                                                  │
│  incident-loop-agent/SKILL.md — 编排逻辑是 prose 描述              │
│  AI 读取 SKILL.md 后手动调用 vet CLI 完成 7 步流程                 │
└─────────────────────────────────────────────────────────────────┘
```

**问题**：
- SKILL.md 是 prose 不是代码，每次执行依赖 AI 重新理解
- 7 步流程没有可执行的状态机，无法断点续跑
- 无法作为常驻服务接收 webhook
- Trace/Policy/Reflexion 基础设施已完备，但缺少编排层来串联

---

## 2. Phase 1 — Agent Runtime（可执行引擎）

**目标**：把 `incident-loop-agent/SKILL.md` 的 7 步编排变成可执行 Go 代码。

### 核心组件

```
cmd/vet/internal/agent/
├── engine.go          # 7-step 编排引擎（状态机）
├── ingest.go          # Incident Payload 解析（CMS webhook / JIRA / patrol）
├── triage.go          # 路由图匹配（docs/skill-routing-graph.md）
├── diagnose.go        # 并行 read-only Describe* 调用
├── propose.go         # 构建 dispatch_plan + GCL loop
├── confirm.go         # 策略评估 + 人机交互门（AUTO/ASK/REFUSE）
├── execute.go         # 委托执行（调用 vet gcl run）
├── reflexion.go       # 写回 failure-patterns.json
├── state.go           # 状态持久化（断点续跑）
└── engine_test.go     # 集成测试
```

### CLI 接口

```bash
# 从 JSON payload 执行
$ vet agent run --payload '{"product":"ecs","symptom":"cpu>90%"}'

# 从 JIRA ticket 执行
$ vet agent run --ticket DOPS-12345

# 带确认模式（ASK 类操作暂停等待输入）
$ vet agent run --ticket DOPS-12345 --confirm
```

### 状态机

```
INGEST → TRIAGE → DIAGNOSE → PROPOSE → CONFIRM → EXECUTE → REFLEXION
   │                   │                    │
   └── FAIL_FAST       └── UNKNOWN_ROUTE    └── REFUSE (skip to REFLEXION)
```

### 与现有代码的关系

```
vet agent run
    │
    ├── 复用 vet gcl run（GCL runner + heal + trace）
    ├── 复用 policy.Load()（执行风险决策）
    ├── 复用 memory.AppendFailurePattern()（自动学习）
    ├── 复用 transpile.Transpile()（自动升级 guardrails）
    └── 复用 docs/skill-routing-graph.md（路由）
```

### 增量

- **新增**: `cmd/vet/internal/agent/` 包（~1500 行 Go）
- **新增**: `vet agent run` 子命令
- **修改**: `cmd/vet/main.go` 注册 agent 子命令
- **不修改**: 任何现有 skill 的 SKILL.md 或 GCL 基础设施

---

## 3. Phase 2 — Agent Server（常驻服务）

**目标**：从一次性 CLI 变为常驻 HTTP 服务，接收 webhook、管理并发任务。

### 核心组件

```
cmd/vet/internal/agentd/
├── server.go          # HTTP server（net/http）
├── handler.go         # REST API handlers
├── pool.go            # Goroutine pool 管理并发执行
├── dashboard.go       # SLO dashboard HTML template
└── server_test.go     # 集成测试
```

### REST API

```
POST   /api/v1/incidents           # 接收 alarm webhook → 创建 run
GET    /api/v1/runs/:id            # 查询执行状态
GET    /api/v1/runs                 # 列出所有 run（分页）
POST   /api/v1/runs/:id/confirm    # 人工确认 ASK 类操作
GET    /api/v1/dashboard           # SLO/统计面板
GET    /api/v1/health              # 健康检查
```

### 并发模型

```
agentd serve (HTTP server)
    │
    ├── POST /incidents → enqueue → goroutine pool (max N)
    │                                │
    │                                ├── agent.Run(payload)
    │                                └── state persisted to .runtime/agent/runs/<id>/
    │
    ├── GET /runs/:id → 读取 .runtime/agent/runs/<id>/state.json
    │
    └── GET /dashboard → 聚合所有 run 的 trace + heal stats
```

### CLI 接口

```bash
# 启动服务
$ vet agentd serve --port 8080 --max-concurrent 5

# 后台运行
$ vet agentd serve --port 8080 --daemon
```

### 增量

- **新增**: `cmd/vet/internal/agentd/` 包（~800 行 Go）
- **新增**: `vet agentd serve` 子命令
- **依赖**: Phase 1 的 `internal/agent/` 包

---

## 4. Phase 3 — Autonomous Domain（自主域）

**目标**：在声明的作用域内，Agent 完全自主执行，人类只设定目标。

### 自主域定义

```yaml
# incident-loop-agent/references/policies/autonomous-domain.yaml
domain:
  skills:
    - ve-ecs-ops
    - ve-redis-ops
    - ve-vpc-ops
  symptoms:
    - cpu>80%
    - mem>85%
    - disk>85%
    - conn_pool_exhausted
  blast_radius: single          # 不允许 multi/account-or-region
  max_concurrent_ops: 3
  slo_targets:
    cpu_p95: 80
    incident_resolution_p50_ms: 300000
```

### 自主行为

| 行为 | 触发条件 | 说明 |
|------|---------|------|
| **监控告警响应** | CMS webhook 到达 | 自动 triage → diagnose → fix |
| **预测式干预** | 指标趋势偏离基线（T12） | 在告警触发前主动干预 |
| **SLO 维持** | SLO 偏离目标 | 自动调整资源（扩缩容、迁移） |
| **修复失败回滚** | validate 不通过（M4-2） | 自动执行 rollback plan |
| **Pattern 升级** | count ≥ 10（T13） | 自动 transpile 为 guardrails |
| **人类介入** | 仅 envelope 外或 Hard guard | dashboard 通知，暂停等待确认 |

### 人类界面

```
┌────────────────────────────────────────────────────┐
│              Agent Dashboard                         │
│                                                      │
│  SLO Overview                                        │
│  ┌──────────────────────────────────────────────┐   │
│  │ ve-ecs-ops  CPU p95: 72% (target 80%) ✅     │   │
│  │ ve-redis-ops Mem p95: 88% (target 85%) ⚠️    │   │
│  │ ve-vpc-ops   Conn: 120 (target 200) ✅       │   │
│  └──────────────────────────────────────────────┘   │
│                                                      │
│  Active Runs                                         │
│  ┌──────────────────────────────────────────────┐   │
│  │ [a1b2] DOPS-12345  triage→diagnose  ⏳       │   │
│  │ [c3d4] DOPS-12346  reflexion  ✅             │   │
│  │ [e5f6] CMS-alarm-7  confirm (ASK)  ⏸️        │   │
│  └──────────────────────────────────────────────┘   │
│                                                      │
│  Guardrails (auto-promoted)                          │
│  ┌──────────────────────────────────────────────┐   │
│  │ evidence_overfetch (count=15) → auto-ASK     │   │
│  │ safety_gate_bypass (count=12) → auto-ASK     │   │
│  │ retry_storm (count=35) → auto-REFUSE 🔴      │   │
│  └──────────────────────────────────────────────┘   │
└────────────────────────────────────────────────────┘
```

### 增量

- **新增**: `incident-loop-agent/references/policies/autonomous-domain.yaml`
- **新增**: `cmd/vet/internal/agentd/dashboard.go`（SLO 面板）
- **新增**: `cmd/vet/internal/agentd/slo.go`（SLO 监控引擎）
- **依赖**: Phase 2 的 agentd + T16 SLO 引擎

---

## 5. 实施优先级

| Phase | 依赖 | 估时 | 状态 |
|-------|------|------|------|
| **Phase 1** — Agent Runtime | T16 (SLO engine) | 5d | 🟡 TODO |
| **Phase 2** — Agent Server | Phase 1 | 3d | 🟡 TODO |
| **Phase 3** — Autonomous Domain | Phase 2 + M4-2 | 5d | 🟡 TODO |

## 6. 前置条件

Phase 1 需要 T16 (SLO engine + dashboard) 完成后才能启动（SLO 是自主域的度量基础）。
当前 L3→L4 进度：T09-T15 ✅，T16 🟡 TODO。

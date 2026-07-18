# Phase 2 — Agent Server (SDD Spec)

日期：2026-07-17
对应架构：`docs/vet-agent-evolution-architecture.md`
依赖：Phase 1 Agent Runtime ✅ 已完成

---

## 1. 功能描述

### 1.1 核心目标

把 Phase 1 的 CLI Agent Runtime 升级为**常驻 HTTP 服务**，能够：
- 接收外部 webhook 告警（CMS/JIRA/自定义）
- 并发管理多个 Agent 执行任务
- 提供 REST API 查询状态
- 提供 Dashboard 监控面板

### 1.2 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                    vet agentd serve                          │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │                   HTTP Server                         │   │
│  │   POST /api/v1/incidents                             │   │
│  │   GET  /api/v1/runs/:id                              │   │
│  │   GET  /api/v1/runs                                  │   │
│  │   POST /api/v1/runs/:id/confirm                      │   │
│  │   GET  /api/v1/dashboard                             │   │
│  │   GET  /api/v1/health                                │   │
│  └──────────────────────────────────────────────────────┘   │
│                              │                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │                Goroutine Pool                         │   │
│  │   max_concurrent: 5 (configurable)                    │   │
│  │   queue: buffered channel                             │   │
│  │                                                       │   │
│  │   ┌─────────┐ ┌─────────┐ ┌─────────┐               │   │
│  │   │ Run #1  │ │ Run #2  │ │ Run #3  │ ...           │   │
│  │   │ agent.  │ │ agent.  │ │ agent.  │               │   │
│  │   │ Run()   │ │ Run()   │ │ Run()   │               │   │
│  │   └─────────┘ └─────────┘ └─────────┘               │   │
│  └──────────────────────────────────────────────────────┘   │
│                              │                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │           State Persistence (Phase 1)                  │   │
│  │   .runtime/agent/runs/<run_id>/state.json              │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 1.3 REST API 详细设计

#### 1.3.1 POST /api/v1/incidents

**用途**：接收告警 webhook，创建新的 Agent Run

**请求体**：
```json
{
  "product_hint": "ecs",
  "symptom": "cpu>90%",
  "ticket_id": "DOPS-12345",  // optional
  "severity": "critical",     // optional
  "region": "cn-beijing",     // optional
  "resource_ids": ["i-xxx"],  // optional
  "source": "cms_webhook"     // optional
}
```

**响应**：
```json
{
  "run_id": "1784301664183030000",
  "status": "queued",
  "message": "Incident received, run created"
}
```

**错误场景**：
- 400: product_hint 缺失
- 429: 并发任务已满
- 500: 服务内部错误

#### 1.3.2 GET /api/v1/runs/:id

**用途**：查询指定 Run 的执行状态

**响应**：
```json
{
  "run_id": "1784301664183030000",
  "status": "running",  // queued|running|paused|completed|failed
  "current_step": "DIAGNOSE",
  "payload": { ... },
  "triage": { "primary_skill": "ve-ecs-ops", "confidence": "high" },
  "evidence": { ... },
  "plan": { ... },
  "confirm": { ... },
  "result": null,
  "created_at": "2026-07-17T14:30:00Z",
  "updated_at": "2026-07-17T14:30:15Z"
}
```

#### 1.3.3 GET /api/v1/runs

**用途**：列出所有 Run（分页）

**查询参数**：
- `status`: 过滤状态（queued|running|paused|completed|failed）
- `page`: 页码（默认 1）
- `limit`: 每页数量（默认 20，最大 100）

**响应**：
```json
{
  "runs": [ ... ],
  "total": 42,
  "page": 1,
  "limit": 20
}
```

#### 1.3.4 POST /api/v1/runs/:id/confirm

**用途**：人工确认 ASK 类操作

**请求体**：
```json
{
  "confirmed": true,  // 或 false 拒绝
  "comment": "允许执行"  // optional
}
```

**响应**：
```json
{
  "run_id": "1784301664183030000",
  "status": "running",
  "message": "Confirmation received, run resumed"
}
```

#### 1.3.5 GET /api/v1/dashboard

**用途**：SLO 监控面板数据

**响应**：
```json
{
  "summary": {
    "total_runs": 100,
    "success_rate": 0.85,
    "avg_duration_ms": 45000,
    "active_runs": 3,
    "queued_runs": 1
  },
  "by_skill": {
    "ve-ecs-ops": { "runs": 50, "success_rate": 0.92 },
    "ve-redis-ops": { "runs": 30, "success_rate": 0.80 }
  },
  "recent_runs": [ ... ]
}
```

#### 1.3.6 GET /api/v1/health

**用途**：健康检查（负载均衡器/监控使用）

**响应**：
```json
{
  "status": "healthy",
  "version": "0.2.0",
  "uptime_seconds": 3600,
  "active_runs": 3,
  "queue_size": 0
}
```

### 1.4 并发模型

```go
// 并发控制
type Pool struct {
    maxConcurrent int            // 最大并发数
    sem           chan struct{}   // 信号量
    queue         chan *Task      // 任务队列
    runs          sync.Map       // run_id -> *RunState
}
```

**关键设计**：
1. **有界并发**：通过 `sem` 信号量限制同时执行的 Agent 数量
2. **任务队列**：超出并发限制的任务进入 buffered channel 排队
3. **状态隔离**：每个 Run 独立的 state.json，互不干扰
4. **优雅关闭**：收到 SIGTERM 后等待正在执行的任务完成，不接受新任务

### 1.5 CLI 接口

```bash
# 启动服务
$ vet agentd serve --port 8080 --max-concurrent 5

# 后台运行（daemon 模式）
$ vet agentd serve --port 8080 --daemon

# 查看服务状态
$ vet agentd status
```

---

## 2. 异常边界

| 场景 | 处理 |
|------|------|
| 并发任务已满（429） | 返回 `Retry-After` 头，建议客户端重试 |
| Agent Run 执行失败 | 记录失败原因，标记 run 为 failed，继续处理队列 |
| HTTP Server 启动失败 | 退出并返回错误码 |
| 优雅关闭超时 | 强制终止正在执行的任务 |
| 状态文件损坏 | 返回 500，日志记录错误 |

---

## 3. 验收标准

### 3.1 功能验收

- [ ] `vet agentd serve` 启动 HTTP 服务器
- [ ] 所有 6 个 API 端点正常工作
- [ ] 并发控制生效（超过 max_concurrent 的任务排队）
- [ ] Agent Run 状态持久化到 `.runtime/agent/runs/`
- [ ] ASK 类操作可以通过 `/confirm` 端点人工确认
- [ ] Dashboard 返回正确的统计数据

### 3.2 非功能验收

- [ ] `go build ./...` 编译通过
- [ ] `go vet ./...` 无警告
- [ ] 单元测试覆盖率 ≥ 70%
- [ ] 集成测试覆盖所有 API 端点
- [ ] API 响应时间 < 100ms（不含 Agent 执行时间）
- [ ] 优雅关闭：收到 SIGTERM 后等待任务完成

### 3.3 测试验收

- [ ] 单元测试：每个 handler 函数有独立测试
- [ ] 集成测试：端到端创建 Run → 查询状态 → 确认
- [ ] 并发测试：验证 max_concurrent 限制
- [ ] 压力测试：100 并发请求不崩溃

---

## 4. 参考文件

| 文件 | 用途 |
|------|------|
| `cmd/vet/internal/agent/engine.go` | Agent Runtime 引擎（Phase 1） |
| `cmd/vet/internal/agent/state.go` | 状态持久化（Phase 1） |
| `cmd/vet/internal/agent/types.go` | 数据模型（Phase 1） |
| `docs/vet-agent-evolution-architecture.md` | 整体架构设计 |

# Phase 2 — Agent Server 执行计划（最终版）

日期：2026-07-17
对应 Spec：`docs/superpowers/specs/2026-07-17-agent-server-phase2-design.md`
依赖：Phase 1 Agent Runtime ✅ 已完成
版本：v3（二次自我复盘后优化）

---

## 目标

把 Phase 1 的 CLI Agent Runtime 升级为常驻 HTTP 服务，支持：
- 接收 webhook 告警（POST /incidents）
- 并发管理多个 Agent 执行任务
- 提供 REST API 查询状态
- 提供 Dashboard 监控面板

---

## 并行策略

```
                    ┌─────────────────────────────────────────────┐
                    │              Agent A (核心路径)              │
                    │  T1 → T2 → T3 → T7 → T8 → T10 → T12       │
                    └─────────────────────────────────────────────┘
                                            │
                                            │ 依赖点：T1 完成后
                                            ▼
                    ┌─────────────────────────────────────────────┐
                    │              Agent B (并行路径)              │
                    │  T4 → T5 → T6 → T9 → T11                   │
                    └─────────────────────────────────────────────┘
```

**并行规则**：
1. **T1 完成后**：Agent A 和 Agent B 可以并行执行
2. **Agent A**：核心 HTTP 服务 + API handlers + CLI 集成
3. **Agent B**：并发池 + Dashboard + 测试
4. **最终汇合**：T12 端到端验证需要双方都完成

---

## 里程碑

### M1 — Server 基础（无依赖）

| # | 任务 | 文件 | 输入 | 输出 | 验收标准 |
|---|------|------|------|------|----------|
| T1 | HTTP Server 结构体 | `server.go` | Phase 1 agent 包 | Server struct | `go build` 通过 |

**T1 - HTTP Server 结构体**：
```go
// 文件：cmd/vet/internal/agentd/server.go

// 输入参考：
// - cmd/vet/internal/agent/engine.go: agent.Run() 函数签名
// - cmd/vet/internal/agent/state.go: LoadState/SaveState 函数

// 输出要求：
// 1. Server struct 定义
type Server struct {
    addr          string
    maxConcurrent int
    pool          *Pool
    runs          sync.Map  // run_id -> *agent.RunState
    startTime     time.Time
    root          string
}

// 2. 构造函数
func NewServer(root, addr string, maxConcurrent int) *Server

// 3. 路由注册
func (s *Server) setupRoutes()

// 4. 启动服务
func (s *Server) Start(ctx context.Context) error

// 5. 优雅关闭
func (s *Server) Stop(ctx context.Context) error

// 验证命令：
cd cmd/vet
go build ./...
go vet ./internal/agentd/...
```

---

### M2 — REST API Handlers（依赖 T1）

| # | 任务 | 文件 | 输入 | 输出 | 验收标准 |
|---|------|------|------|------|----------|
| T2 | 健康检查 + 状态查询 | `handler.go` | T1 完成 | /health + /runs/:id | curl 返回 200 |
| T3 | 创建 Run + 确认 | `handler.go` | T2 完成 | /incidents + /confirm | curl 返回 run_id |

**T2 - 健康检查 + 状态查询**：
```go
// 文件：cmd/vet/internal/agentd/handler.go

// 输入参考：
// - Phase 1 agent/types.go: RunState struct
// - Phase 1 agent/state.go: LoadState()

// 输出要求：
// 1. healthHandler(w, r)
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
    // 返回 JSON: {"status":"healthy","version":"0.2.0","uptime_seconds":...,"active_runs":...}
}

// 2. getRunHandler(w, r)
func (s *Server) getRunHandler(w http.ResponseWriter, r *http.Request) {
    // 从 URL 提取 run_id
    // 调用 agent.LoadState(root, runID)
    // 返回完整 RunState JSON
    // 错误处理：404(run不存在)
}

// 验证命令：
go run . agentd serve --port 8080 &
curl http://localhost:8080/api/v1/health
curl http://localhost:8080/api/v1/runs/nonexistent  # 应返回 404
kill %1
```

**T3 - 创建 Run + 确认**：
```go
// 文件：cmd/vet/internal/agentd/handler.go

// 输入参考：
// - Phase 1 agent/types.go: IncidentPayload struct
// - Phase 1 agent/engine.go: Run() 函数

// 输出要求：
// 1. createIncidentHandler(w, r)
func (s *Server) createIncidentHandler(w http.ResponseWriter, r *http.Request) {
    // 解析 JSON body 到 IncidentPayload
    // 校验 product_hint 必填
    // 调用 s.pool.Submit(payload)
    // 返回: {"run_id":"xxx","status":"queued"}
    // 错误处理：400(参数错误)/429(队列满)
}

// 2. confirmRunHandler(w, r)
func (s *Server) confirmRunHandler(w http.ResponseWriter, r *http.Request) {
    // 从 URL 提取 run_id
    // 解析 JSON body: {"confirmed":bool,"comment":"..."}
    // 更新 state.json
    // 调用 agent.Run() 恢复执行
    // 返回: {"status":"running","message":"..."}
}

// 验证命令：
curl -X POST http://localhost:8080/api/v1/incidents \
  -H "Content-Type: application/json" \
  -d '{"product_hint":"ecs","symptom":"cpu>90%"}'
# 应返回 {"run_id":"xxx","status":"queued"}

curl -X POST http://localhost:8080/api/v1/runs/<run_id>/confirm \
  -H "Content-Type: application/json" \
  -d '{"confirmed":true}'
```

---

### M3 — Run 列表 + Dashboard（依赖 T2，并行于 M2 的 T3）

| # | 任务 | 文件 | 输入 | 输出 | 验收标准 |
|---|------|------|------|------|----------|
| T4 | Run 列表 | `handler.go` | T2 完成 | /runs（分页） | curl 返回 JSON 数组 |
| T5 | Dashboard 模板 + SLO 聚合 | `dashboard.go` | T2 完成 | HTML 页面 | curl 返回 HTML |

**T4 - Run 列表**：
```go
// 文件：cmd/vet/internal/agentd/handler.go

// 输入参考：
// - T2 的 handler 模式
// - .runtime/agent/runs/ 目录结构

// 输出要求：
// 1. listRunsHandler(w, r)
func (s *Server) listRunsHandler(w http.ResponseWriter, r *http.Request) {
    // 扫描 .runtime/agent/runs/ 目录
    // 支持 ?status=running 过滤
    // 支持 ?page=1&limit=20 分页
    // 返回: {"runs":[...],"total":42,"page":1,"limit":20}
}

// 验证命令：
curl "http://localhost:8080/api/v1/runs?status=running&page=1&limit=10"
```

**T5 - Dashboard 模板 + SLO 聚合**：
```go
// 文件：cmd/vet/internal/agentd/dashboard.go

// 输入参考：
// - Spec §1.3.5 GET /api/v1/dashboard 响应格式

// 输出要求：
// 1. dashboardHandler(w, r)
func (s *Server) dashboardHandler(w http.ResponseWriter, r *http.Request) {
    // 聚合统计：total_runs, success_rate, avg_duration_ms
    // 按 skill 分组统计
    // 渲染 HTML 模板
}

// 2. HTML 模板（嵌入 Go 代码）
var dashboardHTML = `
<!DOCTYPE html>
<html>
<head><title>Agent Dashboard</title></head>
<body>
  <h1>SLO Overview</h1>
  <table>...</table>
  
  <h1>Active Runs</h1>
  <table>...</table>
</body>
</html>
`

// 验证命令：
curl http://localhost:8080/api/v1/dashboard
# 应返回 HTML 页面
```

---

### M4 — Goroutine Pool（依赖 T1，并行于 M2）

| # | 任务 | 文件 | 输入 | 输出 | 验收标准 |
|---|------|------|------|------|----------|
| T6 | Pool 结构体 + Submit + 并发控制 | `pool.go` | T1 完成 | 任务提交 + 并发限制 | Submit 后任务被执行，超过 maxConcurrent 时排队 |

**T6 - Pool 结构体 + Submit + 并发控制**：
```go
// 文件：cmd/vet/internal/agentd/pool.go

// 输入参考：
// - Phase 1 agent/engine.go: Run() 函数签名
// - Phase 1 agent/state.go: RunState struct

// 输出要求：
// 1. Pool struct 定义
type Pool struct {
    maxConcurrent int
    sem           chan struct{}
    runs          sync.Map
    root          string
    wg            sync.WaitGroup
}

// 2. 构造函数
func NewPool(root string, maxConcurrent int) *Pool

// 3. Submit 方法（包含并发控制）
func (p *Pool) Submit(payload *agent.IncidentPayload) (string, error) {
    // 生成 runID: fmt.Sprintf("%d", time.Now().UnixNano())
    // 保存初始状态到 state.json
    // 获取信号量: p.sem <- struct{}{}
    // 异步执行 agent.Run()
    // 释放信号量: <-p.sem
    // 返回 runID
}

// 4. Drain 方法（优雅关闭用）
func (p *Pool) Drain(ctx context.Context) error {
    // 等待所有正在执行的任务完成
}

// 验证命令：
go test -v -run TestPoolSubmit ./internal/agentd/...
go test -v -run TestPoolConcurrency ./internal/agentd/...
```

---

### M5 — CLI 集成（依赖 M2 + M4）

| # | 任务 | 文件 | 输入 | 输出 | 验收标准 |
|---|------|------|------|------|----------|
| T7 | CLI 入口 + 参数解析 | `agentd.go` | M2 + M4 | vet agentd serve | `--help` 显示帮助 |
| T8 | main.go 注册 | `main.go` | T7 完成 | vet agentd 子命令 | `vet agentd serve` 可执行 |

**T7 - CLI 入口 + 参数解析**：
```go
// 文件：cmd/vet/agentd.go

// 输入参考：
// - cmd/vet/agent.go: runAgent() 函数模式

// 输出要求：
// 1. runAgentd(args []string) - 路由子命令
func runAgentd(args []string) {
    if len(args) < 1 {
        // 打印帮助
        os.Exit(2)
    }
    switch args[0] {
    case "serve":
        runAgentdServe(args[1:])
    default:
        // 打印帮助
    }
}

// 2. runAgentdServe(args []string) - 解析参数
func runAgentdServe(args []string) {
    fs := flag.NewFlagSet("agentd serve", flag.ExitOnError)
    port := fs.Int("port", 8080, "HTTP port")
    maxConcurrent := fs.Int("max-concurrent", 5, "max concurrent runs")
    root := fs.String("root", repoRoot(), "repo root")
    daemon := fs.Bool("daemon", false, "run as daemon")
    fs.Parse(args)
    
    // 创建并启动 Server
}

// 验证命令：
go run . agentd --help
go run . agentd serve --help
```

**T8 - main.go 注册**：
```go
// 文件：cmd/vet/main.go

// 输入参考：
// - cmd/vet/main.go: 现有子命令注册模式

// 输出要求：
// 1. 在 main() 的 switch 中添加：
case "agentd":
    runAgentd(args)

// 验证命令：
go build ./...
./vet --help  # 应显示 agentd 子命令
./vet agentd serve --port 8080
```

---

### M6 — 单元测试（依赖 M2 + M4）

| # | 任务 | 文件 | 输入 | 输出 | 验收标准 |
|---|------|------|------|------|----------|
| T9 | 单元测试 | `server_test.go` | M2 + M4 | 测试覆盖 | 覆盖率 ≥ 70% |

**T9 - 单元测试**：
```go
// 文件：cmd/vet/internal/agentd/server_test.go

// 输入参考：
// - cmd/vet/internal/agent/agent_test.go: 测试模式

// 输出要求：
// 1. TestHealthHandler
func TestHealthHandler(t *testing.T) {
    // 启动测试服务器
    // 发送 GET /api/v1/health
    // 验证返回 200
    // 验证 JSON 包含 status, version, uptime
}

// 2. TestGetRunHandler
func TestGetRunHandler(t *testing.T) {
    // 测试 404: 不存在的 run_id
    // 测试 200: 存在的 run_id
}

// 3. TestCreateIncidentHandler
func TestCreateIncidentHandler(t *testing.T) {
    // 测试 400: 缺少 product_hint
    // 测试 201: 正常创建
}

// 4. TestListRunsHandler
func TestListRunsHandler(t *testing.T) {
    // 测试空列表
    // 测试分页
    // 测试状态过滤
}

// 5. TestPoolSubmit
func TestPoolSubmit(t *testing.T) {
    // 提交任务
    // 验证任务被执行
}

// 6. TestPoolConcurrency
func TestPoolConcurrency(t *testing.T) {
    // 提交 10 个任务，maxConcurrent=3
    // 验证同时只有 3 个执行
}

// 验证命令：
go test -v -cover ./internal/agentd/...
# 覆盖率应 ≥ 70%
```

---

### M7 — 集成测试（依赖 T9）

| # | 任务 | 文件 | 输入 | 输出 | 验收标准 |
|---|------|------|------|------|----------|
| T10 | 集成测试 | `server_test.go` | T9 完成 | 端到端流程 | 创建→查询→确认 |

**T10 - 集成测试**：
```go
// 文件：cmd/vet/internal/agentd/server_test.go

// 输入参考：
// - T9 的单元测试

// 输出要求：
// 1. TestIntegrationCreateAndQuery
func TestIntegrationCreateAndQuery(t *testing.T) {
    // 启动测试服务器
    // POST /api/v1/incidents 创建 Run
    // GET /api/v1/runs/:id 查询状态
    // 验证状态为 queued 或 running
}

// 2. TestIntegrationConfirmAskRun
func TestIntegrationConfirmAskRun(t *testing.T) {
    // 创建一个 ASK 类 Run
    // POST /api/v1/runs/:id/confirm 确认
    // 验证状态变为 running
}

// 3. TestIntegrationConcurrentRuns
func TestIntegrationConcurrentRuns(t *testing.T) {
    // 并发创建多个 Run
    // 验证并发控制生效
}

// 验证命令：
go test -v -run TestIntegration ./internal/agentd/...
```

---

### M8 — 端到端验证（依赖全部）

| # | 任务 | 文件 | 输入 | 输出 | 验收标准 |
|---|------|------|------|------|----------|
| T11 | 端到端验证 | 手动测试 | T10 完成 | 完整流程 | 所有 curl 命令通过 |
| T12 | 代码复盘 + 优化 | 全部文件 | T11 完成 | 代码质量 | 无重复、无死代码、符合规范 |

**T11 - 端到端验证**：
```bash
# 验证命令（按顺序执行）：

# 1. 构建
cd cmd/vet
go build -o vet .
go vet ./...

# 2. 启动服务
./vet agentd serve --port 8080 &
SERVER_PID=$!
sleep 2

# 3. 健康检查
curl -s http://localhost:8080/api/v1/health | jq .

# 4. 创建 Run
RUN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/incidents \
  -H "Content-Type: application/json" \
  -d '{"product_hint":"ecs","symptom":"cpu>90%"}')
echo $RUN_RESPONSE
RUN_ID=$(echo $RUN_RESPONSE | jq -r '.run_id')

# 5. 查询状态
curl -s http://localhost:8080/api/v1/runs/$RUN_ID | jq .

# 6. 列出所有 Runs
curl -s http://localhost:8080/api/v1/runs | jq .

# 7. 查看 Dashboard
curl -s http://localhost:8080/api/v1/dashboard | head -20

# 8. 测试并发限制（需要同时发送多个请求）
for i in {1..10}; do
  curl -s -X POST http://localhost:8080/api/v1/incidents \
    -H "Content-Type: application/json" \
    -d "{\"product_hint\":\"ecs\",\"symptom\":\"test$i\"}" &
done
wait

# 9. 测试优雅关闭
kill -TERM $SERVER_PID
sleep 5
curl -s http://localhost:8080/api/v1/health  # 应失败

# 10. 清理
kill $SERVER_PID 2>/dev/null
```

**T12 - 代码复盘 + 优化**：
```bash
# 检查项：
# 1. 无重复代码
grep -r "func.*(" cmd/vet/internal/agentd/ | awk -F: '{print $3}' | sort | uniq -d

# 2. 无死代码
go vet ./cmd/vet/internal/agentd/...

# 3. 符合 Go 规范
gofmt -l cmd/vet/internal/agentd/

# 4. 文件职责单一
ls -la cmd/vet/internal/agentd/
```

---

## 关键设计约束

1. **文件位置**：所有代码在 `cmd/vet/internal/agentd/` 下
2. **不修改 Phase 1**：不修改 `internal/agent/` 的任何代码
3. **复用 Phase 1**：调用 `agent.Run()`、`agent.LoadState()`、`agent.SaveState()`
4. **标准库优先**：HTTP Server 使用 `net/http`，不引入外部依赖

---

## 文件清单

```
cmd/vet/
├── agentd.go              # T7: CLI 入口
├── main.go                # T8: 修改，注册 agentd 子命令
└── internal/
    └── agentd/
        ├── server.go      # T1: HTTP Server 结构体
        ├── handler.go     # T2, T3, T4: REST API handlers
        ├── pool.go        # T6: Goroutine Pool
        ├── dashboard.go   # T5: Dashboard 模板
        └── server_test.go # T9, T10: 测试
```

---

## 预估工时

| 里程碑 | 任务数 | 预估 | 可并行 |
|--------|--------|------|--------|
| M1 | 1 | 0.5d | - |
| M2 | 2 | 0.5d | T1 完成后开始 |
| M3 | 2 | 0.5d | 与 M2 并行 |
| M4 | 1 | 0.5d | 与 M2 并行 |
| M5 | 2 | 0.5d | M2 + M4 完成后开始 |
| M6 | 1 | 0.5d | M2 + M4 完成后开始 |
| M7 | 1 | 0.5d | T9 完成后开始 |
| M8 | 2 | 0.5d | T10 完成后开始 |
| **总计** | **12** | **4.5d** | - |

---

## 验证清单

完成所有任务后，逐项验证：

### 编译验证
- [ ] `go build ./cmd/vet/...` 编译通过
- [ ] `go vet ./cmd/vet/...` 无警告

### 测试验证
- [ ] `go test ./cmd/vet/...` 全部通过
- [ ] `go test -cover ./cmd/vet/internal/agentd/...` 覆盖率 ≥ 70%

### API 验证
- [ ] `curl http://localhost:8080/api/v1/health` 返回 200
- [ ] `curl -X POST http://localhost:8080/api/v1/incidents` 返回 run_id
- [ ] `curl http://localhost:8080/api/v1/runs/<id>` 返回状态
- [ ] `curl http://localhost:8080/api/v1/dashboard` 返回 HTML

### 并发验证
- [ ] 并发 10 个请求，只有 5 个同时执行（max_concurrent=5）
- [ ] 超出并发限制的任务排队等待

### 优雅关闭验证
- [ ] 发送 SIGTERM 后，正在执行的任务完成后再退出
- [ ] 不接受新任务

### 代码质量验证
- [ ] 无重复代码
- [ ] 无死代码
- [ ] 符合 Go 规范

---

## 任务依赖图

```
T1 (Server)
├── T2 (health + getRun) ──┬── T3 (create + confirm) ── T7 (CLI) ── T8 (main.go)
│                           │
│                           └── T4 (listRuns) ──────── T7 (CLI)
│
├── T6 (Pool) ──────────────── T5 (Dashboard) ────── T9 (单元测试) ── T10 (集成测试) ── T11 (端到端) ── T12 (复盘)
│
└── T5 (Dashboard) ──────────── T9 (单元测试)
```

**并行点**：
- T2 和 T6 可以并行（都依赖 T1）
- T3 和 T4 可以并行（都依赖 T2）
- T5 和 T6 可以并行（都依赖 T1）

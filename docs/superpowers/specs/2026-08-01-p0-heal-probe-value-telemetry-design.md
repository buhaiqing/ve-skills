# Design: P0 真实恢复验证 + 业务价值遥测

> **Date**: 2026-08-01  
> **Task**: Microsoft Agentic AI 成熟度 P0 — (1) Heal CheckFn 真探针 / stub 禁 AUTO；(2) MTTA/MTTR/人工节省遥测 + 写回  
> **Status**: Approved (user: 落地 P0 两项)

---

## 1. 目标

关闭成熟度诊断中的两个 scale-breaker 缺口：

1. **真实恢复验证** — `internal/heal` 的 `CheckFn: func() bool { return true }` 是假执行；stub plan 不得进入生产 AUTO。
2. **业务价值遥测** — agent run 结束时产出 MTTA / MTTR / 人工节省估计，持久化并写回 ticket（文件门 + 可注入 JIRA writer）。

---

## 2. 边界

### 2.1 范围内

| # | 功能 | 包 |
|---|------|-----|
| P0-1a | RecoveryStep 标记 Stub；IsStubPlan / AllowProductionAuto | `heal` |
| P0-1b | 可注入 ProbeRunner；真实 probe 跑 `ve … Describe*`（argv 直传，无 shell） | `heal` |
| P0-1c | ExecutePlan 在 production 拒绝 stub；测试可 `AllowStub` | `heal` |
| P0-1d | agent Confirm：计划绑定 stub heal → 强制 ASK | `agent` |
| P0-2a | ValueMetrics（MTTA/MTTR/LaborMinutesSaved）计算 + 持久化 | `agent` |
| P0-2b | Writeback 接口（File 默认；可注入） | `agent` |
| P0-2c | engine 集成：记时 → 计算 → Save → Writeback | `agent` |

### 2.2 范围外

- 真实调用云上扩缩容 Action（仅 probe = read_only Describe/metric）
- 真 JIRA HTTP SDK（只留 `TicketWriter` 接口 + file sink；env 注入自定义 writer 留给后续）
- Phase 3 自主域扩权
- Multi-agent CrossCoordinate

---

## 3. 契约

### 3.1 Heal Probe

```go
// ProbeRunner 执行只读探针。MUST 用 argv 直传（P1 Shell 安全），禁止 sh -c。
type ProbeRunner func(ctx context.Context, argv []string) (stdout string, err error)

type RecoveryStep struct {
    Name, Action string
    Params       map[string]interface{}
    ProbeArgv    []string          // e.g. ["ve","ecs","DescribeInstances"]
    Stub         bool              // true = 未接真探针；禁止生产 AUTO
    CheckFn      func() bool       // 兼容测试；生产路径优先 ProbeArgv
    RollbackFn   func() error
}

func (p *RecoveryPlan) IsStub() bool          // 任一步 Stub==true 或 (CheckFn==nil && len(ProbeArgv)==0)
func AllowProductionAuto(p *RecoveryPlan) bool // !p.IsStub()

type ExecuteOpts struct {
    AllowStub bool // 仅测试
    Runner    ProbeRunner // nil → DefaultProbeRunner（exec.CommandContext argv）
}
func (o *Orchestrator) ExecutePlanWithOpts(incidentType string, opts ExecuteOpts) (string, error)
```

默认 plans：全部 `Stub: true`，去掉恒 true 的 CheckFn。提供 `PromotePlan(name, argv[][], runner)` 或构建 helper 把某 plan 升级为非 stub（测试用）。

`ExecutePlan` 保持兼容：内部调 `ExecutePlanWithOpts(..., ExecuteOpts{})`；遇 stub → error `"stub recovery plan %q: refuse production execution"`。

### 3.2 Agent Confirm 联动

```go
// DispatchPlan 新增：
HealIncidentType string `json:"heal_incident_type,omitempty"` // e.g. "cpu_high"

// Confirm：若 HealIncidentType 非空且对应 plan IsStub → Decision=ASK
// Reason: "heal plan <type> is stub (no real probe); production AUTO forbidden"
```

ProposeFix：当 symptom 匹配已知 heal 类型时填 `HealIncidentType`（cpu→cpu_high 等）。

### 3.3 ValueMetrics

```go
type ValueMetrics struct {
    RunID              string  `json:"run_id"`
    TicketID           string  `json:"ticket_id,omitempty"`
    Success            bool    `json:"success"`
    PolicyDecision     string  `json:"policy_decision,omitempty"`
    AlertedAt          string  `json:"alerted_at"`           // RFC3339Nano
    StartedAt          string  `json:"started_at"`
    ResolvedAt         string  `json:"resolved_at"`
    MTTAMs             int64   `json:"mtta_ms"`              // max(0, Started-Alerted)
    MTTRMs             int64   `json:"mttr_ms"`              // max(0, Resolved-Alerted)；失败则为 0 或 -1？→ 失败仍记 elapsed，另用 Success 区分
    AgentDurationMs    int64   `json:"agent_duration_ms"`
    LaborMinutesSaved  float64 `json:"labor_minutes_saved"`  // max(0, BaselineManualMin - agentMin)
    BaselineManualMin  float64 `json:"baseline_manual_min"`  // default 30
}

func ComputeValue(in ValueInput) ValueMetrics
func PersistValue(root string, m ValueMetrics) error  // state dir value.json + append JSONL
```

Baseline：常量 `DefaultBaselineManualMin = 30`；可经 `AgentConfig.Value.BaselineManualMin` 覆盖（可选，缺省即可）。

Writeback：

```go
type TicketWriter interface {
    WriteValueComment(ticketID string, body string) error
}

type FileTicketWriter struct{ Dir string } // writes <dir>/<ticket>.md
func FormatValueComment(m ValueMetrics) string
```

engine：ingest 时记 `StartedAt`；payload 可带 `AlertedAt`（RFC3339）；缺省 AlertedAt=StartedAt（MTTA=0）。结束时 Compute → Persist → 若 TicketID 非空则 Writeback（失败只打 ERROR 日志，不翻转 run Success）。

---

## 4. 异常边界

| 场景 | 行为 |
|------|------|
| stub plan + ExecutePlan | error，不改 Status 为 completed |
| ProbeRunner 超时/非零退出 | Check 失败 → rollback 路径 |
| PersistValue 目录不存在 | MkdirAll |
| TicketWriter 失败 | log ERROR；run 仍按执行结果成功/失败 |
| AlertedAt 解析失败 | 回退 StartedAt |
| HealIncidentType 未知 | Confirm 不因 heal stub 降级（仅已知类型） |

---

## 5. 验收标准

- [ ] `go build ./...` / `go vet ./...` 干净
- [ ] heal: TestIsStubPlan、TestAllowProductionAuto、TestExecutePlanRejectsStub、TestExecutePlanWithRealProbe（假 Runner）、TestDefaultPlansAreStub
- [ ] agent: TestConfirmStubHealForcesASK、TestComputeValueMTTA/MTTR、TestPersistValueRoundtrip、TestFileTicketWriter
- [ ] 向后兼容：`ExecutePlan` 签名不变；`Confirm(root, plan)` 对无 HealIncidentType 的 plan 行为不变
- [ ] 无 `sh -c` 新引入路径

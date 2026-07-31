# L400 Wave B — 嵌入运营 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 完成 L400 Capable 的 Wave B（价值看板、可注入 Ticket 写回、价值优先级周报、ASK ConfirmedBy 硬门、黄金路径），使业务 KPI 与技术成功率同屏，且 ASK 放行可审计。

**Architecture:** 在既有 `ValueMetrics` / `PersistValue` / `FileTicketWriter`、agentd dashboard 雏形、GCL `Options.Confirmed`/`ConfirmedBy` 之上扩展聚合与硬门；不接真实 JIRA SDK；不做 Wave C 自主域 / L500（ADR-0004/0005）。

**Tech Stack:** Go 1.26、`cmd/vet` 模块、`go test` / `httptest`、现有 `agent` / `agentd` / `gcl/run`；禁止 `sh -c`；无新第三方依赖。

**Upstream:**
- Roadmap: `docs/superpowers/plans/2026-08-01-l400-capable-roadmap.md` §3 Wave B
- Spec: `docs/superpowers/specs/2026-08-01-l400-wave-b-ops-embed-design.md`
- ADRs: **Accepted** — 0003（价值一等）、0006（ConfirmedBy）；约束 0002/0005 仍适用
- Wave A 已落地: Promote、eval-report、Makefile `check-ci`、KB load

## Global Constraints

- ADR-0003: Dashboard **必须**展示 Value KPI；TicketWriter 可注入，默认 File
- ADR-0006: ASK 放行必须 `Confirmed && ConfirmedBy≠""`；裸 confirmed 失败
- ADR-0005: 不做预测式 / multi-agent
- ADR-0002: Stub heal 仍禁止生产 AUTO（本波不改 heal 升级集合）
- AGENTS: `go build` + `go vet` 干净；凭证不入库；结构化日志关键路径带 run ID
- 超时升级通知通道推迟 Wave C2；B4 仅硬门 + 可选 deadline 字段

## File map

| File | Responsibility |
|------|----------------|
| `cmd/vet/internal/agent/value_agg.go` | `AggregateValueMetrics(root) ValueDashboard` + p50 helper |
| `cmd/vet/internal/agent/value_agg_test.go` | JSONL fixture → p50 / AUTO rate |
| `cmd/vet/internal/agentd/dashboard.go` | 渲染 Value KPI 卡片 |
| `cmd/vet/internal/agent/value.go` / `engine.go` | TicketWriter 注入点 |
| `cmd/vet/internal/agent/priority.go` | `BuildValuePriorityReport` |
| `cmd/vet/internal/agent/evalreport.go` + `agent.go` | `--priority-out` 或并列写 priority JSON |
| `cmd/vet/internal/gcl/run/run.go` | ASK 要求 ConfirmedBy |
| `cmd/vet/gcl.go` | `--confirmed` 校验 `--confirmed-by` |
| `cmd/vet/internal/agentd/handler.go` | confirm body 要求 `confirmed_by` |
| `docs/runbooks/agentd-value-golden-path.md` | B5 操作说明 |
| `cmd/vet/internal/agentd/golden_path_test.go` | httptest 黄金路径 |

**Out of this plan:** Wave C、真实 JIRA、治理 webhook、autonomous-domain.yaml。

---

### Task 1: AggregateValueMetrics + Dashboard Value KPI（B1）

**Files:**
- Create: `cmd/vet/internal/agent/value_agg.go`
- Create: `cmd/vet/internal/agent/value_agg_test.go`
- Modify: `cmd/vet/internal/agentd/dashboard.go`
- Modify: `cmd/vet/internal/agentd/server_test.go`（或 dashboard 测）

**Interfaces:**
- Consumes: `ValueMetrics`, `audit-results/value-metrics.jsonl`, per-run `value.json`（优先 JSONL）
- Produces:
  - `type ValueDashboard struct { P50MTTAMs, P50MTTRMs int64; LaborMinutesSum, AutoRate float64; Samples int }`
  - `func AggregateValueMetrics(root string) ValueDashboard`
  - DashboardStats 增加 `Value ValueDashboard`（或并列字段）

- [ ] **Step 1: Failing test — p50 + AUTO rate**

```go
func TestAggregateValueMetricsP50AndAutoRate(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "audit-results")
	_ = os.MkdirAll(dir, 0o755)
	lines := []string{
		`{"run_id":"1","success":true,"policy_decision":"AUTO","mtta_ms":100,"mttr_ms":200,"labor_minutes_saved":10}`,
		`{"run_id":"2","success":true,"policy_decision":"ASK","mtta_ms":300,"mttr_ms":400,"labor_minutes_saved":5}`,
		`{"run_id":"3","success":false,"policy_decision":"AUTO","mtta_ms":500,"mttr_ms":600,"labor_minutes_saved":0}`,
	}
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(l + "\n")
	}
	_ = os.WriteFile(filepath.Join(dir, "value-metrics.jsonl"), []byte(b.String()), 0o644)
	d := AggregateValueMetrics(root)
	if d.Samples != 3 {
		t.Fatalf("samples=%d", d.Samples)
	}
	if d.P50MTTAMs != 300 {
		t.Fatalf("p50 mtta=%d", d.P50MTTAMs)
	}
	if d.AutoRate < 0.66 || d.AutoRate > 0.67 {
		t.Fatalf("auto_rate=%v", d.AutoRate)
	}
	if d.LaborMinutesSum != 15 {
		t.Fatalf("labor=%v", d.LaborMinutesSum)
	}
}
```

- [ ] **Step 2: Run — expect FAIL**

Run: `cd cmd/vet && go test ./internal/agent/ -run TestAggregateValueMetricsP50AndAutoRate -count=1`  
Expected: FAIL（undefined）

- [ ] **Step 3: Implement AggregateValueMetrics**

```go
func AggregateValueMetrics(root string) ValueDashboard {
	// read JSONL; skip bad lines; sort mtta/mttr for p50;
	// AutoRate = count(AUTO)/Samples; LaborMinutesSum = sum
}
```

Missing file → zero ValueDashboard（非 error）。

- [ ] **Step 4: Wire dashboard**

`aggregateStats` 调用 `agent.AggregateValueMetrics(s.root)`；`renderDashboard` 增加卡片：P50 MTTA / P50 MTTR / LaborSaved / AUTO%。

- [ ] **Step 5: Tests PASS + commit**

```bash
cd cmd/vet && go test ./internal/agent/ ./internal/agentd/ -count=1
git commit -m "feat(agentd): show ValueMetrics KPIs on dashboard"
```

---

### Task 2: TicketWriter 注入（B2）

**Files:**
- Modify: `cmd/vet/internal/agent/value.go`（若需 opts 类型）
- Modify: `cmd/vet/internal/agent/engine.go`（`emitValue` 接受 writer）
- Create/Modify: `cmd/vet/internal/agent/value_test.go`

**Interfaces:**
- Consumes: `TicketWriter`, `FileTicketWriter`
- Produces: `emitValue` / 导出 `EmitValue(root, state, ..., writer TicketWriter)`；`writer==nil` → `FileTicketWriter{Dir: runDir(root, runID)}`

- [ ] **Step 1: Failing test — fake writer called**

```go
type recordingWriter struct{ n int; lastID, lastBody string }
func (r *recordingWriter) WriteValueComment(id, body string) error {
	r.n++; r.lastID = id; r.lastBody = body; return nil
}

func TestEmitValueUsesInjectedTicketWriter(t *testing.T) {
	// build minimal RunState with TicketID + Confirm AUTO + Success result
	// call EmitValue(..., &recordingWriter{})
	// assert n==1 and body contains "MTTA"
}
```

注意：若 `emitValue` 未导出，可测通过小范围导出 `EmitValue` 或在同包测试。

- [ ] **Step 2: Implement injection；默认 File 行为不变**

- [ ] **Step 3: PASS + commit**

```bash
git commit -m "feat(agent): inject TicketWriter for value writeback"
```

---

### Task 3: 价值优先级周报（B3）

**Files:**
- Create: `cmd/vet/internal/agent/priority.go`
- Create: `cmd/vet/internal/agent/priority_test.go`
- Modify: `cmd/vet/internal/agent/evalreport.go` 和/或 `cmd/vet/agent.go`

**Interfaces:**
- Consumes: `EvalSample` 列表 + 可选 `ValueMetrics` 切片（或从 JSONL 加载）
- Produces:
  - `func BuildValuePriorityReport(samples []EvalSample, values []ValueMetrics) ValuePriorityReport`
  - CLI：`vet agent eval-report --samples X --out Y --priority-out Z`

排序规则（确定性）：
1. `high_false_refuse`：skill 维度 FalseRefuse 率高
2. `low_success`：value Success=false 占比高
3. `high_mtta`：skill/全局 p50 MTTA 高  
Top N=5；同分按 Key 字典序。

- [ ] **Step 1: Failing fixture test**

```go
func TestBuildValuePriorityReportOrdersFalseRefuseFirst(t *testing.T) {
	samples := []EvalSample{
		{PredictedSkill: "ve-ecs-ops", LabeledSkill: "ve-ecs-ops", PolicyDecision: "REFUSE", ShouldRefuse: false},
		{PredictedSkill: "ve-ecs-ops", LabeledSkill: "ve-ecs-ops", PolicyDecision: "REFUSE", ShouldRefuse: false},
		{PredictedSkill: "ve-redis-ops", LabeledSkill: "ve-redis-ops", PolicyDecision: "AUTO", ShouldRefuse: false},
	}
	rep := BuildValuePriorityReport(samples, nil)
	if len(rep.Top) == 0 || rep.Top[0].Key != "ve-ecs-ops" || rep.Top[0].Reason != "high_false_refuse" {
		t.Fatalf("%+v", rep.Top)
	}
}
```

- [ ] **Step 2: Implement + wire CLI flag `--priority-out`**

- [ ] **Step 3: PASS + commit**

```bash
git commit -m "feat(agent): value-priority report for weekly triage"
```

---

### Task 4: ASK ConfirmedBy 硬门（B4）

**Files:**
- Modify: `cmd/vet/internal/gcl/run/run.go`（约 L741 策略门）
- Modify: `cmd/vet/gcl.go`
- Modify: `cmd/vet/internal/gcl/run/run_test.go`
- Modify: `cmd/vet/internal/agentd/handler.go`（`confirmRunHandler`）
- Modify: `cmd/vet/internal/agentd/server_test.go`

**Interfaces:**
- Consumes: `Options.Confirmed`, `Options.ConfirmedBy`
- Produces: ASK 放行条件变为 `Confirmed && strings.TrimSpace(ConfirmedBy) != ""`

- [ ] **Step 1: Failing GCL test**

```go
func TestASKConfirmedWithoutByIsBlocked(t *testing.T) {
	// same ASK fixture as existing test
	res := Run(opts with Confirmed:true, ConfirmedBy:"")
	if res.ExitCode != 4 {
		t.Fatalf("expected POLICY_BLOCK, got %d", res.ExitCode)
	}
}
```

- [ ] **Step 2: Change gate**

```go
askOK := policy == OpAsk && opts.Confirmed && strings.TrimSpace(opts.ConfirmedBy) != ""
if policy != OpAuto && !askOK { /* POLICY_BLOCK */ }
```

- [ ] **Step 3: CLI guard**

在 `gcl.go` parse 后：若 `*confirmed && strings.TrimSpace(*confirmedBy)==""` → `fmt.Fprintln(stderr, "..."); os.Exit(2)`。

- [ ] **Step 4: agentd confirm body**

```go
var req struct {
	Confirmed   bool   `json:"confirmed"`
	ConfirmedBy string `json:"confirmed_by"`
	Comment     string `json:"comment,omitempty"`
}
if req.Confirmed && strings.TrimSpace(req.ConfirmedBy) == "" {
	writeError(w, http.StatusBadRequest, "confirmed_by required when confirmed=true")
	return
}
// persist ConfirmedBy onto state (add field on RunState or ConfirmResult if missing)
```

若 `RunState` / `ConfirmResult` 无 provenance 字段：新增 `ConfirmedBy string \`json:"confirmed_by,omitempty"\`` 并写入 SaveState。

- [ ] **Step 5: PASS + commit**

```bash
cd cmd/vet && go test ./internal/gcl/run/ ./internal/agentd/ -count=1
git commit -m "fix(gcl,agentd): require ConfirmedBy for ASK authorization"
```

---

### Task 5: Golden path runbook + httptest（B5）

**Files:**
- Create: `docs/runbooks/agentd-value-golden-path.md`
- Create: `cmd/vet/internal/agentd/golden_path_test.go`
- Modify: `docs/README.md`（链到 runbook）

**Interfaces:**
- Consumes: `POST /api/v1/incidents`, confirm handler, `PersistValue` 路径
- Produces: 可复现步骤文档 + `TestGoldenPathIncidentToValueJSON`

- [ ] **Step 1: Write runbook**（告警 JSON → agentd → 若 ASK 则带 confirmed_by → 检查 `value.json` / dashboard）

- [ ] **Step 2: httptest golden test**

对 **AUTO** 路径用 promoted heal symptom（`cpu`）+ allowlist fixture，或 dry 路径：至少断言 run 创建且终端 `value.json` 存在。  
若 CONFIRM=ASK：再 POST confirm with `confirmed_by`。

- [ ] **Step 3: PASS + commit**

```bash
git commit -m "docs+test(agentd): golden path incident to value metrics"
```

---

### Task 6: 全量回归

**Files:** none required

- [ ] **Step 1: Verify**

```bash
cd cmd/vet && make check-ci && go test ./internal/agent/ ./internal/agentd/ ./internal/gcl/run/ -count=1
codegraph sync --quiet
```

Expected: all PASS

- [ ] **Step 2: Spec coverage checklist**

| Wave B ID | Task |
|-----------|------|
| B1 Dashboard Value | Task 1 |
| B2 TicketWriter 注入 | Task 2 |
| B3 价值优先级 | Task 3 |
| B4 ConfirmedBy 硬门 | Task 4 |
| B5 黄金路径 | Task 5 |

- [ ] **Step 3: Optional** — 若 Makefile 仍无 `help`，可顺手加（非阻塞）

---

## Self-review

1. **Spec coverage:** Roadmap B1–B5 均有 Task；ADR-0003/0006 契约写入 Interfaces。
2. **Fact-check:** `confirmRunHandler` 当前仅 `confirmed` bool（handler.go）；GCL 仅查 `Confirmed`（run.go:741）— 计划显式修补。
3. **No L500 / Wave C** 泄漏。
4. **TDD** 每 Task 先红后绿；禁止真实 JIRA 凭证。

## Follow-on

- Wave C plan: `docs/superpowers/plans/2026-08-01-l400-wave-c-capable.md`（未写）
- ADR-0004 envelope 落地仅在 Wave C

---

**Plan complete.** Execution: subagent-driven-development（推荐）或 executing-plans。

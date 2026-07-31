# Design: L400 Wave B — 嵌入运营

> **Date**: 2026-08-01  
> **Status**: Ready for implementation  
> **Upstream**: [L400 roadmap §3](../plans/2026-08-01-l400-capable-roadmap.md)、[ADR-0003](../../adr/0003-value-metrics-first-class.md)、[ADR-0006](../../adr/0006-ask-requires-confirmed-by.md)  
> **Prerequisite**: Wave A DoD 已满足（`feat/l400-wave-a`）

## 1. Problem

Wave A 已使 ≥2 heal plan 可生产 AUTO、CI/eval/KB 落地。但：

- agentd `/dashboard` 仅技术成功率，未消费 `ValueMetrics`
- Ticket 写回写死 `FileTicketWriter`，无注入点
- 周报不置顶「低成功率 / 高误 REFUSE」路径
- ASK 放行仍可裸 `--confirmed` / agentd confirm 无 provenance（违反 ADR-0006）
- 缺一条可复现的 webhook→run→价值写回黄金路径

## 2. Goals (B1–B5)

| ID | Goal | Non-goal |
|----|------|----------|
| B1 | Dashboard 展示 p50 MTTA/MTTR、LaborSaved 合计、AUTO 比率 | 不做实时 push / 复杂前端框架 |
| B2 | `TicketWriter` 可注入；默认仍 File | 不把真实 JIRA SDK/凭证入库 |
| B3 | 价值优先级周报：低 path 成功率 / 高误 REFUSE / 高 MTTA 置顶 | 不做自动改 policy |
| B4 | ASK 必须 `Confirmed && ConfirmedBy≠""`；agentd/CLI 硬门 | 不做超时升级 webhook（可 stub 字段） |
| B5 | 1 条 golden-path runbook + 可跑的集成测试（httptest） | 不接真实告警总线 |

## 3. Contracts

### 3.1 Value aggregation (B1)

```go
// AggregateValueMetrics reads audit-results/value-metrics.jsonl (and/or per-run value.json).
type ValueDashboard struct {
    P50MTTAMs       int64   `json:"p50_mtta_ms"`
    P50MTTRMs       int64   `json:"p50_mttr_ms"`
    LaborMinutesSum float64 `json:"labor_minutes_sum"`
    AutoRate        float64 `json:"auto_rate"` // PolicyDecision==AUTO / n
    Samples         int     `json:"samples"`
}
func AggregateValueMetrics(root string) ValueDashboard
```

### 3.2 TicketWriter injection (B2)

```go
// Engine / emitValue uses optional writer; nil → FileTicketWriter{Dir: runDir}.
type EmitValueOpts struct {
    Writer TicketWriter // optional
}
```

Fake writer in tests only; production JIRA adapter out-of-repo or behind build tag — Wave B 只交付接口 + File 默认 + fake 测试。

### 3.3 Priority report (B3)

Extend or sibling of `EvalReport`:

```go
type PriorityItem struct {
    Key     string  `json:"key"`     // skill or heal path
    Reason  string  `json:"reason"`  // low_success | high_false_refuse | high_mtta
    Score   float64 `json:"score"`   // higher = more urgent
}
type ValuePriorityReport struct {
    Top []PriorityItem `json:"top"`
}
```

CLI: `vet agent eval-report` 可增 `--priority-out`，或新子命令 `vet agent value-priority`（plan 任务内二选一，推荐扩展 eval-report 以免新入口膨胀）。

### 3.4 ConfirmedBy hard gate (B4)

- GCL: `policy==ASK && opts.Confirmed` **且** `strings.TrimSpace(opts.ConfirmedBy)!=""` 才放行；否则 POLICY_BLOCK。
- CLI: `--confirmed` 无 `--confirmed-by` → flag 解析后立即错误退出（非 0）。
- agentd `POST /api/v1/runs/{id}/confirm` body:

```json
{"confirmed":true,"confirmed_by":"ticket:DOPS-1|human:alice","comment":"..."}
```

缺 `confirmed_by` 且 `confirmed=true` → 400。

超时升级：Wave B 仅在 state/文档中预留 `ask_deadline` 字段（可选），不实现通知通道（→ Wave C2）。

### 3.5 Golden path (B5)

文档：`docs/runbooks/agentd-value-golden-path.md`  
测试：`agentd` httptest：`POST /incidents` → （若 ASK）`POST .../confirm` with provenance → 读 `value.json` 存在。

## 4. Acceptance

- Dashboard HTML/JSON 含 Value KPI（与 Success Rate 同屏）
- Fake TicketWriter 在单测被调用；默认 File 行为回归绿
- Priority Top 列表对 fixture 确定性排序
- 裸 `--confirmed` / 无 `confirmed_by` 的 agentd confirm **失败**
- Golden path 测试 PASS；runbook 链接进 docs/README

## 5. Out of scope

Wave C（autonomous-domain、治理告警、inventory）、L500（ADR-0005）、真实 JIRA 凭证、预测式干预。

# Plan: P0 Heal Probe + Value Telemetry

> **Date**: 2026-08-01  
> **Spec**: `docs/superpowers/specs/2026-08-01-p0-heal-probe-value-telemetry-design.md`

---

## Milestone A — P0-1 Heal Probe（并行轨 1）

### Tasks

1. **`heal/probe.go`**（new）
   - `ProbeRunner` type
   - `DefaultProbeRunner` via `exec.CommandContext(ctx, argv[0], argv[1:]...)`（无 shell）
   - `RunProbe(ctx, runner, argv) error` — 空 argv → error

2. **`heal/orchestrator.go`**（modify）
   - `RecoveryStep` 加 `ProbeArgv []string`, `Stub bool`
   - `RecoveryPlan.IsStub()` / `AllowProductionAuto(p)`
   - `ExecuteOpts{AllowStub, Runner}`；`ExecutePlanWithOpts`
   - `ExecutePlan` → `ExecutePlanWithOpts(type, ExecuteOpts{})`
   - `defaultPlans()`：全部 `Stub: true`，删除恒 true CheckFn
   - 执行步：若 Stub && !AllowStub → error；否则若 ProbeArgv 非空用 Runner，否则 CheckFn

3. **`heal/orchestrator_test.go`**（modify）
   - 更新现有测试：对需跑通的 plan 设 `AllowStub: true` 或改用非 stub 自定义 plan
   - 新增：IsStub / AllowProductionAuto / RejectStub / RealProbe fake runner

### DoD
```
cd cmd/vet && go test ./internal/heal/ -count=1
```

---

## Milestone B — P0-2 Value Telemetry（并行轨 2）

### Tasks

1. **`agent/value.go`**（new）
   - `ValueMetrics`, `ValueInput`, `ComputeValue`, `PersistValue`, `FormatValueComment`
   - `TicketWriter` + `FileTicketWriter`
   - `DefaultBaselineManualMin = 30`

2. **`agent/types.go` / `state.go`**
   - `IncidentPayload.AlertedAt string`（optional RFC3339）
   - `RunState.Value *ValueMetrics`；`StartedAt` 可放 state 或仅内存

3. **`agent/propose.go`**
   - symptom→heal type map；设 `DispatchPlan.HealIncidentType`

4. **`agent/confirm.go`**
   - 若 HealIncidentType 非空：查 `heal` default orchestrator plan；IsStub → ASK

5. **`agent/engine.go`**
   - 记录 started；结束 ComputeValue → PersistValue → optional FileTicketWriter under runDir
   - RunResult 可附带 Value 指针（可选）

6. **`agent/value_test.go` + confirm/propose 测试**

### DoD
```
cd cmd/vet && go test ./internal/agent/ -count=1
```

---

## Milestone C — 集成门禁

```
cd cmd/vet && go build ./... && go vet ./... && go test ./internal/heal/ ./internal/agent/ -count=1
codegraph sync --quiet
```

---

## 依赖图

```
spec ✅
  ├── A (heal) ──┐
  └── B (agent) ─┴─► C (gate)
```

A 与 B 可并行；B 的 confirm 依赖 A 的 `IsStub` API（B 可 import heal）。
若并行冲突：A 先合 IsStub/AllowProductionAuto；B 后接。

# Spec: L4 Synthetic Load-Test Harness (T16 runtime evidence)

> Date: 2026-08-06
> Parent card: docs/l3-to-l4-tasks/T16-l4-slo-envelope-dashboard.md §7 (L4 exit checklist)
> Status: draft

## 1. Current state

`vet autonomy test` (internal/autonomy/test.go) is a **stub**: `runSyntheticIncident`
hardcodes `Status="passed"`, `SLOStatus=Healthy`, `Prompted=false`, and never
invokes the SLO engine, rollback, heal, predict, or reflexion. The real code paths
exist but are orphaned:

- `RunIncidentWithEngine` (test.go:97) genuinely calls `engine.Observe` +
  `engine.RecommendAction` and sets `RolledBack` when action is `"rollback"` — but the
  CLI never calls it.
- `slo`, `gcl/predict`, `reflexion/transpile`, `gcl/heal` (`heal-stats`) are REAL and
  pure/file-driven; only need synthetic input, no cloud API.
- `gcl/rollback` is an in-memory stub (no live executor). `trace.Iteration` has no
  `rollback_applied` field.

## 2. Goal

Make `vet autonomy test` drive REAL code paths so it produces runtime evidence for L4
exit items ①③④⑤⑥⑦, replacing the hardcoded-passed stub. Keep `vet autonomy test --n 5`
behaving as a PASS/FAIL gate. Add a `vet autonomy loadtest` subcommand that emits a
consolidated per-item L4 evidence report.

## 3. Acceptance criteria (DoD)

- [ ] `RunNConsecutiveIncidents` builds a real `slo.Engine` from envelope `slo_ref`s
      and routes each incident through `RunIncidentWithEngine` (real Observe/RecommendAction).
- [ ] Per incident, `RolledBack` (action `rollback`) flows into `HarnessReport`;
      `IncidentResult` gains `RollbackApplied` mirror for trace.
- [ ] `trace.Iteration` gains `RollbackApplied bool` (omitempty); harness sets it when
      rollback recommended, so "trace 含 rollback_applied=true" (item ③) is satisfiable.
- [ ] SLO status per incident recorded (item ④ SLO maintained); `HarnessReport` gains
      `Rollbacks int` and `SLOViolations` already exists.
- [ ] New `vet autonomy loadtest` runs N incidents through real paths AND drives:
      - ⑤ reflexion transpile with ≥10 patterns → asserts guardrail file written.
      - ⑥ `gcl predict` with a Risk=high recipe → asserts `risk=high`.
      - ⑦ `gcl heal-stats` over a synthetic log (9 ok + 1 fail) → asserts rate 90% > 80%.
- [ ] `go build`/`go vet`/`go test ./...` green; `vet autonomy test --n 5` still PASS
      (0 prompts, SLO maintained); `vet autonomy loadtest` prints per-item PASS/BLOCKED.

## 4. Out of scope (explicit)

- No live cloud API calls (predict/slo/transpile/heal-stats are pure/file-based).
- No production rollback executor change (item ③ evidence = real code path via
  `RunIncidentWithEngine` → `RecommendAction.Type=="rollback"` → trace field; the
  executor remains the in-memory stub by design). Labeled accordingly in report.
- No change to existing `vet autonomy test` PASS/FAIL contract for CI (`--n 5`).

## 5. Interfaces (signatures to implement)

```go
// internal/autonomy/test.go
type HarnessReport struct {
    TotalIncidents  int
    Passed          int
    Failed          int
    Prompts         int
    SLOViolations   int
    Rollbacks       int            // NEW
    Duration        time.Duration
    IncidentResults []IncidentResult
}
type IncidentResult struct {
    RunID           string
    Status          string
    SLOStatus       slo.SLOStatus
    RolledBack      bool
    RollbackApplied bool          // NEW (trace mirror)
    Prompted        bool
    Duration        time.Duration
}

// NEW: consolidated L4 evidence report
type L4Evidence struct {
    Items []L4Item
}
type L4Item struct {
    ID      string // "1".."8"
    Name    string
    Status  string // "PASS" | "BLOCKED" | "PARTIAL"
    Detail  string
}
func RunL4LoadTest(ctx context.Context, n int, envelopePath string) (*L4Evidence, error)

// internal/gcl/trace/trace.go
type Iteration struct {
    // ...existing fields...
    RollbackApplied bool `json:"rollback_applied,omitempty"` // NEW (item ③)
}
```

## 6. Design notes

- Engine construction: map envelope `Domain.SLORef` → a synthetic `SLO` with
  `Skill=Domain.Skills[0]`, `Metric=<derived>`, `Target=100`, `Window=0` (so a
  single >1.2× sample yields Violated → RecommendAction rollback, exercising item ③).
  For healthy incidents, feed `Value ≤ 0.8*Target` → Healthy (item ④).
- Metric→SLO match uses `SLO.Metric == Metric.Name`; the harness sets `Metric.Name`
  from the SLO's metric and `Tags["skill"]` from `SLO.Skill`.
- Loadtest ⑤: write a temp patterns.md (≥10 rows, count≥10) → `reflexion.TranspileFile`
  into a temp yaml → assert file non-empty + contains `severity`.
- Loadtest ⑥: `predict.Registry.Evaluate` with `slow_commands_per_sec` history
  `80,90,100,120,150` current 150 → `Evaluation.Risk == "high"`.
- Loadtest ⑦: write temp heal log (9 ok + 1 fail) → `heal.ParseFile` + `Metrics.SuccessRate`
  → assert `>= 0.80`.
- All temp files use `t.TempDir()` (tests) or `os.TempDir()` (CLI), no repo pollution.

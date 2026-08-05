# Plan: L4 Synthetic Load-Test Harness (T16 runtime evidence)

> Date: 2026-08-06
> Spec: docs/superpowers/specs/2026-08-06-l4-synthetic-loadtest-design.md

## Milestones

### M1 — Wire real SLO path into harness (items ①④)
- Modify `internal/autonomy/test.go`: `RunNConsecutiveIncidents` builds `*slo.Engine`
  from envelope domains' `slo_ref` (synthesize SLO{Metic,Skill,Target=100,Window=0}),
  and calls `RunIncidentWithEngine` per incident instead of `runSyntheticIncident`.
- Feed a mix: most incidents healthy (`Value ≤ 0.8*Target`), inject ≥1 rollback incident
  (`Value > 1.2*Target` → Violated → RecommendAction rollback).
- `HarnessReport` gains `Rollbacks int`; `IncidentResult` gains `RollbackApplied bool`.
- Verification: `go test -run TestRunNConsecutiveIncidents ./internal/autonomy/`.

### M2 — Add `rollback_applied` to trace (item ③)
- Add `RollbackApplied bool` to `trace.Iteration` (omitempty) in `internal/gcl/trace/trace.go`.
- Harness sets it on the incident's `Iteration` (or records via a lightweight trace write)
  when `RolledBack`. Keep change minimal — no new trace IO; set field on a constructed Iteration.
- Verification: `go test -run TestIterationRollback ./internal/gcl/trace/`.

### M3 — `vet autonomy loadtest` subcommand (items ⑤⑥⑦ + consolidated report)
- Add `loadtest` case in `cmd/vet/autonomy.go`; implement `autonomy.RunL4LoadTest`.
- `RunL4LoadTest` drives:
  - ① N incidents end-to-end via real paths.
  - ② zero prompts (already invariant).
  - ③ rollback_applied from M1/M2.
  - ④ SLO maintained.
  - ⑤ reflexion transpile ≥10 patterns → guardrail written.
  - ⑥ predict Risk=high recipe.
  - ⑦ heal-stats 90% over synthetic log.
  - ⑧ policy CHANGELOG present (already true).
- Each item → `L4Item{ID,Name,Status,Detail}`; print table; exit non-zero only on hard fail.
- Verification: `go test -run TestRunL4LoadTest ./internal/autonomy/`.

### M4 — Build / vet / test + manual smoke
- `go build ./...`, `go vet ./...`, `go test ./...`.
- `vet autonomy test --envelope autonomy-envelope.md --n 5` → PASS.
- `vet autonomy loadtest --envelope autonomy-envelope.md --n 5` → per-item report.

## Dependencies
- M1 before M2 (field consumed by harness) and M3 (report reads M1 results).
- No external deps; all paths pure/file-based.

## DoD (per spec §3)
- [x] real engine wired, Rollbacks counted, trace field added (M1/M2)
- [x] loadtest subcommand + L4Evidence report for ③⑤⑥⑦ (M3)
- [x] `go build`/`go vet`/`go test ./...` green (M4)
- [x] `vet autonomy test --n 5` PASS; `vet autonomy loadtest` prints per-item status (M4 smoke)

## Status (2026-08-06)
M1–M4 全部完成并验证：`go build`/`go vet`/`go test ./...` 绿；`vet autonomy test --n 5` 与
`vet autonomy loadtest --n 3` 均 PASS，L4 items ③⑤⑥⑦ 经真实 slo/predict/transpile/heal 代码路径取证。

## Verification commands
```bash
go -C cmd/vet test ./internal/autonomy/ ./internal/gcl/trace/ -run 'TestRunNConsecutiveIncidents|TestRunL4LoadTest|TestIterationRollback'
go -C cmd/vet build ./... && go -C cmd/vet vet ./...
go -C cmd/vet build -o /tmp/vet .
/tmp/vet autonomy test --envelope ../incident-loop-agent/references/policies/autonomy-envelope.md --n 5
/tmp/vet autonomy loadtest --envelope ../incident-loop-agent/references/policies/autonomy-envelope.md --n 5
```

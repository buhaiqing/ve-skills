# Plan: GCL Loop Feedback Closure (P0-1 + P0-2)

> Date: 2026-08-06
> Spec: docs/superpowers/specs/2026-08-06-gcl-loop-feedback-design.md

## Milestones

### M1 — `NO_PROGRESS` status + command hashing (P0-1 detection)
- `internal/gcl/trace/trace.go`: add `"NO_PROGRESS"` to `FinalStatuses` (trace.go:20).
- `internal/gcl/run/run.go`: add `commandHash(cmd string) string` (sha256 hex of the
  masked command) near the other helpers.
- Loop: introduce `effectiveCmd` + `prevHash`; arm no-op detection only when the
  previous iteration's decision was `RETRY`; on hit, build `Final{Status:"NO_PROGRESS"}`
  + `FailurePattern{Category:"no_progress"}`, persist trace, optional writeback, return
  `ExitCode: 1`.
- Verification: `go test -run TestNoProgress ./internal/gcl/run/` + `go build ./...`.

### M2 — `--fix-command` hook (P0-1 remediation path)
- `internal/gcl/run/run.go`: add `Options.FixCommand`; add `runFixCommand(...)` that
  shells out to `opts.FixCommand` with the JSON payload on stdin and parses
  `{"command": "..."}` from stdout. Malformed ⇒ ERROR log + return "" (caller keeps
  previous effective command ⇒ M1 fires `NO_PROGRESS`).
- `runGeneratorWithHeal` gains an `effectiveCmd` parameter; `Options.Command` untouched
  (stays the original for audit delta).
- `cmd/vet/gcl.go`: register `--fix-command` in `runGCLRun`; wire through to `Options`.
- Verification: `go test -run TestFixCommand ./internal/gcl/run/`; then a CLI smoke:
  `/tmp/vet gcl run --skill ve-ecs-ops --command 'false' --fix-command 'echo {"command":"true"}' --critic-json <fixture> --max-iter 3`
  ⇒ PASS at iter 2, and the trace's iter-2 generator command is `true`.

### M3 — Critic history + `prior_suggestions` (P0-2)
- `internal/gcl/run/run.go`: move the current-iteration `append` before the
  `runIsolatedCritic` call; after the critic call, stamp `Decision` onto
  `tr.Iterations[len-1]`. Add `priorSuggestions(iterations)` helper (dedupe, cap 10×200).
- `runIsolatedCritic`: add `"prior_suggestions": priorSuggestions(iterations)` to the
  payload map (run.go:429).
- `runFixCommand` payload also carries `prior_suggestions`.
- Verification: `go test -run TestCriticHistory ./internal/gcl/run/` — asserts a stub
  `--critic-command` receives a payload whose `trace.iterations` length equals the
  current iter and whose `prior_suggestions` contains iter-1's suggestion.

### M4 — Build / vet / test / gate + 2-round self-review
- `cd cmd/vet && go build ./... && go vet ./... && go test ./... -count=1`.
- `go -C cmd/vet build -o /tmp/vet . && /tmp/vet gcl gate --root .` (structural gate).
- Walk AGENTS.md two-round self-review checklist (C1–C17 / F1–F13) — fix every finding.
- Update `docs/gcl-spec.md` §4/§5 (retry semantics + `NO_PROGRESS` termination + new
  exit-status row) and the changelog, since the spec is the single source of truth and
  currently implies the command is rewritten without saying how.

## Dependencies
- M1 before M2 (fix path relies on hash guard for malformed output).
- M1 + M2 before M3 (critic payload adds `prior_suggestions` consumed by both).
- M4 last (green build required).

## DoD (per spec §3)
- [x] `Options.FixCommand` + `--fix-command` registered; empty ⇒ legacy behaviour
- [x] Fix command stdin JSON → stdout `{"command":...}` applied to next iteration
- [x] Effective command tracked; trace records actually-executed masked command
- [x] Identical effective command across a **failing** RETRY ⇒ `Final.Status == "NO_PROGRESS"`, exit 1
- [x] `FinalStatuses` contains `NO_PROGRESS`
- [x] Critic payload includes current iteration + `prior_suggestions`
- [x] `go build` / `go vet` / `go test ./...` green; existing tests unchanged
- [x] New tests: TestNoProgress, TestFixCommand, TestCriticHistory, TestCommandHash
- [x] `docs/gcl-spec.md` §4/§5/§13 updated
- [x] Follow-ups from independent Critic review (2026-10-03): guard gates on
      `gen.ExitCode != 0` (`TestNoProgressSkipsSuccessfulRetry`); guard iteration stamps
      `RequestID` (`TestNoProgressTracePassesCheck`); `Aggregate` counts `NO_PROGRESS`
      (`TestAggregate`); `NO_PROGRESS` `Final` carries `unresolved` dims matching the
      `MAX_ITER` shape (§6.2) — repro final `unresolved:["spec_compliance","correctness"]`,
      asserted in `TestNoProgress`.

## Status (2026-10-03)
M1..M4 complete + Critic follow-ups closed. Verified in worktree `ve-skills-gcl-noprog`:
`go build`/`go vet` clean; `go test ./internal/gcl/... -count=1` all ok;
`vet gcl gate --root .` → 30/30. The only repo-wide test failure is the pre-existing
`TestRepoRootResolvesToRepo`, which hardcodes the repo dir name `ve-skills` and therefore
fails in any worktree (passes in the main repo) — unrelated to this change.

## Verification commands
```bash
cd cmd/vet
go build ./... && go vet ./...
go test ./internal/gcl/run/ -run 'TestNoProgress|TestNoProgressSkipsSuccessfulRetry|TestNoProgressTracePassesCheck|TestFixCommand|TestCriticHistory|TestCommandHash' -v
go test ./internal/gcl/trace/ -run TestAggregate -v
go test ./internal/gcl/... -count=1
go build -o /tmp/vet . && /tmp/vet gcl gate --root ..
# NO_PROGRESS end-to-end: persisted trace must carry final.unresolved (spec §6.2)
# /tmp/vet gcl run --root <tmp> --skill ve-ecs-ops --command <failing> ... ; jq .final <trace>
```

> Note: `go test ./... -count=1` (whole module) still fails only
> `TestRepoRootResolvesToRepo` inside a worktree, because it asserts
> `filepath.Base(repoRoot()) == "ve-skills"`. Run the module-wide suite from the main
> checkout to see it green, or scope to `./internal/...`.

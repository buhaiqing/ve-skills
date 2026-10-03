# Spec: GCL Loop Feedback Closure (P0-1 no-op retry + P0-2 critic history)

> Date: 2026-08-06
> Trigger: Loop Engineering review — P0-1 / P0-2 findings
> Status: draft

## 1. Current state

`cmd/vet/internal/gcl/run/run.go` (`Run`) owns the GCL loop. Two defects make the
RETRY path structurally unable to converge:

**P0-1 — RETRY re-runs an identical command.**
`runGeneratorWithHeal` (run.go:270) builds `env := {"GCL_CRITIC_FEEDBACK": criticFeedback,
"GCL_KNOWN_FAILURE_PATTERNS": knownPatterns}` (run.go:271) and then retries
`opts.Command` **verbatim** (run.go:276, 280, 287). `Options` (run.go:119) has no
`FixCommand` field, and the CLI (`cmd/vet/gcl.go` `runGCLRun`) registers no
`--fix-command` flag. So for a **deterministic** failure (a command whose exit code
and output do not depend on the feedback env var — the overwhelmingly common case)
every iteration produces byte-identical `GeneratorResult`; only
`gen.Args["critic_feedback"]` (run.go:794) differs. The loop burns all `MaxIter`
iterations and terminates `MAX_ITER` (exit 1), with the loop having made **zero
progress and zero observable change between iterations**.

**P0-2 — the Critic cannot see history.**
`runIsolatedCritic` (run.go:422) *does* marshal `trace.iterations` into the payload
(run.go:437), but the call site passes `tr.Iterations` (run.go:834) **before** the
current iteration is appended (append happens at run.go:858). Two consequences:
1. The Critic never sees the iteration it is scoring (offset-by-one at best).
2. No structured *prior suggestions* are passed at all — only the current
   `generator_output`. The Critic therefore cannot tell whether a previous blocking
   suggestion was addressed, so its verdict cannot express "stagnant" and the loop
   has no signal to stop early or escalate.

Also, `trace.FinalStatuses` (trace.go:20) enumerates only
`{"PASS","SAFETY_FAIL","MAX_ITER"}` — there is no status for "the loop stopped because
it stopped making progress".

## 2. Goal

Close the feedback loop so that:
1. An **external Generator can rewrite the command** from Critic feedback between
   iterations (new `--fix-command` hook, symmetric to `--critic-command`).
2. The system **detects a no-op retry** (identical effective command on a retry that
   was supposed to change) and terminates with a stable, auditable `NO_PROGRESS`
   status instead of silently exhausting `MaxIter`.
3. The **Critic receives full history** — prior iterations *including* the current
   one, plus an explicit `prior_suggestions` list — so RETRY is directed and
   stagnation is expressible.

## 3. Acceptance criteria (DoD)

- [ ] `Options` gains `FixCommand string`; `vet gcl run` registers `--fix-command`.
      Empty value ⇒ legacy behaviour (no fix attempt), fully backward compatible.
- [ ] When `FixCommand != ""` and the decision is `RETRY`, the orchestrator invokes
      the fix command with a JSON payload (same shape family as the critic payload:
      `skill`, `operation_intent`, `generator_output`, `critic` = the blocking
      record, `next_iter`). Its stdout is parsed as `{"command": "<new cmd>"}`;
      a non-empty `command` replaces the effective command for the next iteration.
- [ ] The **effective command** is tracked separately from `opts.Command`; every
      `GeneratorResult` in the trace still records the actually-executed (masked)
      command, so the audit trail shows what really ran.
- [ ] If a retry iteration computes a `sha256` of the effective command **equal** to
      the previous iteration's, the loop terminates immediately with
      `Final.Status = "NO_PROGRESS"`, `ExitCode = 1`, and a
      `FailurePattern{Category:"no_progress", Fix:"supply --fix-command or address the blocking suggestion"}`.
- [ ] `trace.FinalStatuses` gains `"NO_PROGRESS"`.
- [ ] `runIsolatedCritic` is called **after** the current iteration is appended, so
      `trace.iterations` includes the iteration under review; payload additionally
      carries `prior_suggestions` (flattened, deduped, capped) from previous iterations.
- [ ] When a fix command yields the same command as the last attempt, that is also a
      no-op ⇒ `NO_PROGRESS` (detection is on the *effective* command, not on who set it).
- [ ] `go build ./...` / `go vet ./...` / `go test ./...` green; existing tests
      (`TestRunResultTimedOut`, `TestDeriveOperationIntent`, `TestScoreDecision_9Cell`)
      unchanged and passing.
- [ ] New tests cover: fix-command rewrite path, no-op detection → `NO_PROGRESS`,
      prior-suggestions in critic payload, backward-compat (no `FixCommand` ⇒ old behaviour).

## 3bis. Addendum — P0-3: NO_PROGRESS iteration records no critic verdict

> Found during M4 verification (2026-10-03) by replaying a real `NO_PROGRESS` trace.
> Not a separate finding from the Loop Engineering review; it is a defect **introduced
> by the P0-1 fix** and must be closed in the same change.

### 3bis.1 Defect

The no-op guard (run.go:919) fires at the **top** of iteration *N*, i.e. *after* iteration
*N-1*'s critic already returned `RETRY`. It appends a fresh `trace.Iteration` for iteration
*N* but never populates `Critic`. Observed trace JSON:

```json
{"iter": 2, "decision": "RETRY", "critic": {"scores": null, "suggestions": null, "blocking": false}}
```

Two problems:
1. **Distorted record** — iteration 2 is stamped `RETRY` with no critic evidence. The
   *verdict* being acted on belongs to iteration *N-1*; iteration *N* is merely the
   duplicate execution that proves the stall.
2. **Schema hazard** — a `retry` decision with `scores: null` is exactly the shape
   `check/trace` and downstream consumers would treat as malformed. Verified 2026-10-03:
   `trace.Check` does **not** tolerate it. Because the guard appends an iteration for a
   command that *did* run, `trace.Check` requires that iteration to carry a `request_id`
   (`iterations[i].request_id is required (ve call not traceable)`) — so the guard must
   stamp one, or the loop writes a trace its own checker rejects.

### 3bis.2 Fix contract

- The no-op guard binds the observed verdict to `lastCritic` and **ignores** the
  duplicate iteration's own generator result, reusing iteration *N-1*'s verdict:
  - `Critic` = `lastCritic` (the real RETRY verdict),
  - `Decision` = `"RETRY"` (unchanged — the stop *is* on a retry),
  - `Generator` = the duplicate execution's result (the evidence of the stall),
  - `Timestamp` / `DurationMs` = the duplicate iteration's own timing.
- `Final.FailurePattern` stays `category: "no_progress"` and gains
  `count: 1, reusable: false` so it matches the shape `extractFailurePattern` emits.
- The guard's duplicate iteration stamps `RequestID: parseRequestID(gen.ResultExcerpt)`
  so the appended iteration satisfies `trace.Check` (see §3bis.1 problem 2).
- The guard gates on `gen.ExitCode != 0`: a **byte-identical** command may legitimately
  succeed on retry (transient failure — rate limit, lock contention, eventual
  consistency), which is precisely what `MaxIter` exists for. Only a *repeated failure*
  is a stall; a successful retry falls through to the Critic and can reach `PASS`.
  (Regression guard: `TestNoProgressSkipsSuccessfulRetry`.)
- `Final.Unresolved` lists the rubric dims still below threshold, computed from the
  guard's captured `criticRec.Scores` via the same `critic.RubricThresholds` loop the
  `MAX_ITER` path uses — so `NO_PROGRESS` and `MAX_ITER` finals share one shape (§6.2).
  (Assertion: `TestNoProgress`.)

### 3bis.3 Acceptance criteria (DoD)

- [x] The `NO_PROGRESS` iteration carries non-nil `critic.scores` equal to the prior
      iteration's scores; `critic.blocking` is `true`. (verified 2026-10-03: iter 2 critic
      `{correctness:0, idempotency:0.5, safety:1, spec_compliance:0, traceability:0.5,
      blocking:true}` — identical to iter 1)
- [x] `TestNoProgress` asserts the above (not merely `Final.Status`).
- [x] `vet gcl trace` aggregates a `NO_PROGRESS` trace without error.
- [x] `vet gcl gate --root .` still reports `30/30` (no regression).
- [x] The `NO_PROGRESS` trace passes `trace.Check` (both iterations carry `request_id`).
      (`TestNoProgressTracePassesCheck`; guard iteration stamps `RequestID`.)
- [x] A byte-identical retry that *succeeds* reaches `PASS`, not `NO_PROGRESS`.
      (`TestNoProgressSkipsSuccessfulRetry`; guard gates on `gen.ExitCode != 0`.)
- [x] `Aggregate` counts `NO_PROGRESS` in both `totals` and `by_skill` (`TestAggregate`),
      so a `NO_PROGRESS` run is no longer invisible to `vet gcl trace` while still
      inflating `total_runs`.
- [x] `NO_PROGRESS` `Final` carries `unresolved` = the below-threshold dims, matching the
      `MAX_ITER` shape in §6.2. (verified 2026-10-03: repro final
      `unresolved:["spec_compliance","correctness"]`; `TestNoProgress` asserts both
      non-emptiness and set-equality against `critic.RubricThresholds`.)

## 4. Out of scope (explicit)

- No change to the rubric thresholds or `critic.Decide` mapping (PASS/RETRY/SAFETY_FAIL
  keep their meaning; `NO_PROGRESS` is an *orchestrator* termination, not a rubric verdict).
- No new cloud API calls; fix command is an arbitrary external process, exactly like
  `--critic-command`.
- No change to exit-code semantics for existing statuses (0/1/2/3/4 unchanged).
  `NO_PROGRESS` reuses exit 1 (same "did not pass" bucket as `MAX_ITER`) — distinct
  only in `Final.Status`, preserving CI contract.
- No persistence of heal metrics (that is P1-5, a separate change).
- No change to `deriveOperationIntent` resource_scope (that is P1-3).

## 5. Interfaces (signatures to implement)

```go
// cmd/vet/internal/gcl/run/run.go
type Options struct {
    // ...existing fields...
    FixCommand string // NEW: external Generator fixer; "" = disabled (legacy)
}

// effective command for the next iteration; "" means "reuse current".
func runFixCommand(opts Options, intent map[string]any, gen trace.GeneratorResult,
    c *critic.CriticResult, iterations []trace.Iteration, nextIter int, runID string) (string, error)

// runGeneratorWithHeal keeps its (result, healClass, healRecord) contract.
// P0-3 does NOT widen this signature — instead the loop separately binds the
// critic verdict it already holds (lastCritic) so the no-op guard can reuse the
// real verdict instead of emitting a null one.
func runGeneratorWithHeal(opts Options, effectiveCmd, criticFeedback, knownPatterns string,
    metrics *heal.Metrics, logPath string, runID string) (trace.GeneratorResult, string, *trace.SelfHealingRecord)

// sha256 of the masked effective command — identical hash across a retry ⇒ no-op.
func commandHash(cmd string) string

// cmd/vet/internal/gcl/trace/trace.go
var FinalStatuses = []string{"PASS", "SAFETY_FAIL", "MAX_ITER", "NO_PROGRESS"} // NEW member

// cmd/vet/gcl.go (runGCLRun)
fs.StringVar(&fixCommand, "fix-command", "", "external fixer: reads critic JSON on stdin, emits {\"command\":\"...\"} on stdout")
```

### Fix-command payload (stdin JSON)

```json
{
  "skill": "ve-ecs-ops",
  "operation_intent": { "...": "..." },
  "generator_output": { "command": "...", "exit_code": 1, "result_excerpt": "..." },
  "critic": { "scores": {"...": 0.5}, "suggestions": ["..."], "blocking": true },
  "prior_suggestions": ["..."],
  "next_iter": 3
}
```

### Fix-command output (stdout JSON)

```json
{ "command": "ve ecs DescribeInstances --InstanceIds '[\"i-xxx\"]' --region cn-beijing" }
```

Malformed output ⇒ log ERROR and keep the previous effective command (so P0-1's
no-op detection will then fire, giving an explicit `NO_PROGRESS` rather than a crash).

### Critic payload (stdin JSON) — extended

```json
{
  "skill": "...",
  "operation_intent": { "...": "..." },
  "generator_output": { "command": "...", "exit_code": 1, "result_excerpt": "..." },
  "trace": { "iterations": [ /* NOW includes the current iteration */ ] },
  "prior_suggestions": ["deduped suggestions from iterations 1..n-1"],
  "rubric_path": "ve-ecs-ops/references/rubric.md"
}
```

## 6. Design notes

### 6.1 Effective-command tracking

Introduce `effectiveCmd := opts.Command` before the loop. Each iteration:
1. Run `runGeneratorWithHeal` against `effectiveCmd` (new param; `Options.Command`
   stays the *original* so the trace/audit can show the delta).
2. Compute `h := commandHash(effectiveCmd)`. If `prevHash != "" && h == prevHash`
   **and the previous decision was RETRY and this attempt also failed
   (`gen.ExitCode != 0`)** ⇒ `NO_PROGRESS` termination. A byte-identical retry that
   succeeds is a transient recovery, not a stall — it must fall through to the Critic.
3. After a RETRY decision, if `opts.FixCommand != ""`, call `runFixCommand`; a
   non-empty returned command sets `effectiveCmd` for the next iteration.
4. `prevHash = h` at the end of the iteration.

Note the hash guard is only armed on the *retry* path (`prevDecision == "RETRY"`), so
a legitimate `MaxIter=1` single-shot run is never mislabelled.

### 6.2 `NO_PROGRESS` trace shape

```
Iteration{iter, decision:"RETRY", Generator:{command:<effective, masked>}, Critic:{...}}
Final{status:"NO_PROGRESS", iter:<n>, output:&lastExcerpt,
      unresolved:[...dims below threshold...],
      failure_pattern:{category:"no_progress", command:<effective>, error:"effective command unchanged across retry", fix:"supply --fix-command / address blocking suggestion"}}
```

`writebackFailurePattern` is called as for `MAX_ITER`, but the pattern's
`Category == "no_progress"` distinguishes it. `StructuralOnly` runs skip writeback
(via the existing guard in `writebackFailurePattern`), so CI smoke tests do not
pollute the pattern store with `no_progress` rows.

### 6.3 Ordering fix for the Critic call

Move the `tr.Iterations = append(...)` (run.go:858) **before** the critic invocation
(run.go:834) so `trace.iterations` seen by the Critic includes the current iteration.
Because `critic.Decide` only needs `c.Scores`, we compute `decision` after the critic
call and then stamp it onto the already-appended iteration (`tr.Iterations[len-1].Decision`).
This keeps a single append site (no double-append, no off-by-one).

Guard: the earlier credential-leak branch (run.go:806-826) still appends its own
`SAFETY_FAIL` iteration *without* a critic — unaffected, since it returns immediately.

### 6.4 `prior_suggestions` derivation

Flatten `tr.Iterations[*].Critic.Suggestions`, dedupe preserving order, cap at 10
entries and 200 chars/entry (mirrors `firstN` usage at run.go:902). Pass to both the
critic payload and the fix payload.

### 6.5 Backward compatibility

- `--fix-command` absent ⇒ `FixCommand == ""` ⇒ step 3 skipped; steps 2/4 still add
  no-op detection. This is a **behaviour change for the degenerate case** (identical
  command retried) but is strictly better than silently burning `MaxIter`: the trace
  now records *why* it stopped. Existing tests do not exercise an identical-command
  multi-iteration retry (they use `StructuralOnly: true` + `MaxIter` default), so they
  remain green; the change is validated by a dedicated new test.
- `--critic-json` / stdin / `--structural-critic-only` paths unchanged.

## 7. Verification commands

```bash
cd cmd/vet
go build ./... && go vet ./...
go test ./internal/gcl/... -count=1
go test ./internal/gcl/run/ -run 'TestNoProgress|TestFixCommand|TestCriticHistory|TestCommandHash' -v
cd .. && go -C cmd/vet build -o /tmp/vet . && /tmp/vet gcl gate --root .   # structural CI gate unchanged
```

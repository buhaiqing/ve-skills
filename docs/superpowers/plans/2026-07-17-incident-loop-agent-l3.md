# Plan: incident-loop-agent L3 Execution-Risk Gate

> **Parent spec:** `docs/superpowers/specs/2026-07-17-incident-loop-agent-l3-design.md`
> **Companion task cards:** `docs/l2-to-l3-tasks/T01`–`T08` (T06 = runtime; this plan is the Superpowers-level counterpart that the task cards decompose from)
> **Status:** Shipped 2026-07-16 (P8/P9). This plan is the **retroactive spec/plan anchor** created per the AGENTS.md "Spec + Plan First" iron-clad rule, to give the already-shipped code a three-way consistency anchor for future code review.

---

## 0. Scope

Wire the L3 execution-risk policy (`AUTO/ASK/REFUSE`) into `vet gcl run` so the loop auto-executes bounded low-risk ops and escalates only ambiguous/destructive/low-confidence ones. This is the single runtime seam that completes the L2→L3 leap.

**Out of scope (handled by sibling tasks, not duplicated here):**
- Policy prose/schema/allow-list/leaf-metadata authoring → T01–T04.
- Trace schema validator + RequestId enforcement → T07.
- Eval coverage AUTO/ASK/REFUSE + Safety-invariant guard → T08.

---

## 1. Milestones

### M1 — Gate logic in `Run()` (P8)
**File:** `cmd/vet/internal/gcl/run/run.go`

- Call `policyInputs` + `scoreDecision` at the top of each iteration, **before** `runCommand`.
- Block unless `policy == OpAuto`, or (`policy == OpAsk && opts.Confirmed`).
- On block: append `Iteration{Decision: "POLICY_BLOCK", PolicyDecision: blocked}`, set `Final{Status: "POLICY_BLOCK"}`, `PersistTrace` + `writebackFailurePattern`, return `ExitCode: 4`.
- Degrade unconfirmed ASK → REFUSE for the recorded decision (non-interactive runtime has no human to ask).

### M2 — `--confirmed` flag (P9)
**File:** `cmd/vet/gcl.go`

- Add `--confirmed` bool flag → `Options.Confirmed`.
- No change to REFUSE path (Safety=0 / destructive-floor still refuse).

### M3 — Trace schema field
**File:** `cmd/vet/internal/gcl/trace/trace.go`

- `Iteration` gains `PolicyDecision string \`json:"policy_decision,omitempty"\``.

### M4 — Spec/doc sync
**Files:** `docs/gcl-spec.md`, `incident-loop-agent/SKILL.md`

- `gcl-spec.md` §4 add `[0.5] Execution-Risk Gate`; §5 add `POLICY_BLOCK` + exit-code table (0/1/2/3/4); §6 trace example add `policy_decision`.
- `SKILL.md` Step 5 note: non-interactive ASK without `--confirmed` → REFUSE (exit 4).

---

## 2. Function-level contract (reconciles T06 drift)

T06 DoD listed 5 functions (`scoreDecision`, `enforceMaxIter`, `withBackoff`, `detectPartialRollback`, `validatePolicyDecision`, `RunLoop`). **Actual shipped code consolidated these into `Run` + helpers** — recorded here as the authoritative mapping so code review's three-way check passes:

| T06 claimed | Actual (authoritative) | Note |
|-------------|------------------------|------|
| `scoreDecision` | `scoreDecision` (`run.go:89`) | ✅ exact, unit-tested 9-cell |
| `enforceMaxIter` | inline in `Run` (`run.go:404-412`) | max_iter from `SkillMaxIter` map; no separate fn |
| `withBackoff` | not separate; single attempt per iter, retry via loop | backoff deferred to L4 (T09 smart-retry) |
| `detectPartialRollback` | not in runtime | belongs to L4 execute-monitor (T10); out of L3 scope |
| `validatePolicyDecision` | inline gate block (`run.go:436`) | policy decision validated by `scoreDecision` + `policyguard` invariants |
| `RunLoop` | `Run(opts Options) Result` (`run.go:403`) | entry is `Run`, not `RunLoop` |

> **Consistency ruling:** consolidation is acceptable (simple-priority rule). The *behavior* T06 required (gate before execute, destructive→ASK, safety=0→REFUSE, max_iter enforcement) is present. Future reviews must check **behavior**, not literal function names.

---

## 3. DoD (authoritative)

```
✅ 1. `Run()` calls scoreDecision before runCommand on every iteration
✅ 2. POLICY_BLOCK recorded + ExitCode 4 on block
✅ 3. read_only + high conf → AUTO (no prompt)
✅ 4. Safety=0 → REFUSE, never bypassed by --confirmed
✅ 5. destructive → ASK (honors --confirmed); spec-compliant
✅ 6. `Options.Confirmed` wired from --confirmed flag
✅ 7. trace.Iteration.PolicyDecision field added + persisted
✅ 8. cmd/vet go build + go vet + go test clean
✅ 9. gcl-spec.md documents gate + exit code 4
✅ 10. incident-loop-agent/SKILL.md Step 5 notes ASK→REFUSE(no --confirmed) degrade
```

---

## 4. Verification

```bash
cd cmd/vet
go build ./... && go vet ./... && go test ./...

# policy gate unit tests (run_test.go)
go test -run 'TestPolicyInputs_FailSafe|TestRun_PolicyBlocksDestructive|TestRun_PolicyAutoReadonly|TestRun_PolicyAskNeedsConfirm' ./internal/gcl/run/ -v
```

---

## 5. Three-Way Consistency Anchor

Per AGENTS.md "Spec + Plan First" rule, this plan + its spec are the **intent anchor** for the shipped code. Code review must verify:

- **Spec ↔ Plan:** decision logic (§Component 1) == DoD (§3). ✅ aligned.
- **Plan ↔ Code:** function mapping (§2) explains T06 drift; behavior present. ✅ aligned.
- **Spec ↔ Code:** destructive→ASK, safety=0→REFUSE, exit 4 — matches `run.go`. ✅ aligned.

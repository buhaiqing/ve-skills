# policyguard — Safety-Invariant Checker

Enforces three policy invariants that must never be violated:

| # | Invariant | Description |
|---|-----------|-------------|
| 1 | `safety = 0 → REFUSE` | Hard floor. Safety=0 means the operation is categorically unsafe — it must always be refused, regardless of other signals. |
| 2 | `safety_class = destructive → not AUTO` | Destructive operations (Delete, Terminate, Stop, etc.) require human confirmation — AUTO is never appropriate. |
| 3 | `metadata_missing → not AUTO` | When operation metadata (safety_class, blast_radius) is absent, the fail-safe default is ASK — never AUTO. |

## CLI

Registered as `vet check policyguard`:

```bash
# Check a repo (walks incident-loop-agent/assets/ for dispatch plans)
vet check policyguard --root /path/to/repo

# Machine-readable JSON output
vet check policyguard --root /path/to/repo --json
```

## Go API

```go
import "github.com/buhaiqing/ve-skills/cmd/vet/internal/check/policyguard"

// Check a dispatch plan file
if err := policyguard.Check("/path/to/dispatch_plan.json"); err != nil {
    fmt.Println(err)
}

// With detailed violation reports
reports, err := policyguard.CheckPlanWithReport(plan)
for _, r := range reports {
    fmt.Printf("invariant-%d: %s\n", r.Invariant, r.Description)
}
```

## Decision Logic (mirrors T06 `scoreDecision`)

```
safety == 0                → REFUSE  (hard floor, overrides everything)
safety_class == "destructive" → ASK    (never AUTO)
metadata_ok == false                 → ASK    (fail-safe, never AUTO)
skill not in allowlist              → ASK    (8 coordinated skills only)
read_only + high confidence         → AUTO
mutating + single + high confidence → AUTO
otherwise                           → ASK
```

## Prompt Examples for AI Agents

**When to invoke policyguard**:

> "Check this dispatch plan for policy violations: vet check policyguard --root ."

**When writing a loop iteration that evaluates a new operation**:

```
Check whether the operation has safety_class=destructive or metadata_ok=false.
If either, set decision=ASK and require {{user.confirm}}.
Never set decision=AUTO when safety=0 or destructive — safety floor is non-negotiable.
```

**When an AI evaluates a GCL Critic scorecard**:

```
Review: Safety dimension score must be 0 → REFUSE.
safety=0 in any dimension is a hard floor, not a rubric signal.
```

## Test Coverage

```bash
go test ./internal/check/policyguard/ -v
```

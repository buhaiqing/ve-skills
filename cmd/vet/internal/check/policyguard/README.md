# policyguard — Safety-Invariant Checker

Enforces three policy invariants that must never be violated:

| # | Invariant | Description |
|---|-----------|-------------|
| 1 | `safety = 0 → REFUSE` | Hard floor. Safety=0 means the operation is categorically unsafe — it must always be refused, regardless of other signals. |
| 2 | `safety_class = destructive → not AUTO` | Destructive operations (Delete, Terminate, Stop, etc.) require human confirmation — AUTO is never appropriate. |
| 3 | `metadata_missing → not AUTO` | When operation metadata (safety_class, blast_radius) is absent, the fail-safe default is ASK — never AUTO. |

## Usage

```go
import "github.com/buhaiqing/ve-skills/cmd/vet/internal/check/policyguard"

// Check a dispatch plan from file
if err := policyguard.Check("/path/to/dispatch_plan.json"); err != nil {
    fmt.Println(err)
}

// Or with detailed violation reports
reports, err := policyguard.CheckPlanWithReport(plan)
```

## CLI

Registered as `vet check policyguard`:

```bash
/tmp/vet check policyguard --root /path/to/repo
```

## Decision Logic (mirrors T06 `scoreDecision`)

```
safety == 0         → REFUSE  (hard floor, overrides everything)
safety_class == "destructive" → ASK    (never AUTO)
metadata_ok == false             → ASK    (fail-safe, never AUTO)
otherwise                        → AUTO   (only when all preconditions met)
```

## Test Coverage

```bash
go test ./internal/check/policyguard/ -v
```

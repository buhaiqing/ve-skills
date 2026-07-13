# trace — Incident Loop Trace Validator

Validates `incident-loop-agent` GCL trace files against the incident-trace schema.

## CLI

Registered as `vet check trace`:

```bash
# Check all traces under audit-results/
vet check trace --root /path/to/repo

# Machine-readable JSON output
vet check trace --root /path/to/repo --json
```

## Trace File Naming

Only files matching `incident-trace-*.json` are checked.
Other files (e.g. `gcl-trace-*.json` from the GCL runner) are skipped.

## Schema Requirements (incident-loop-agent/assets/trace.schema.json)

| Field | Required | Notes |
|-------|----------|--------|
| `ticket_id` | ✅ | JIRA/DOPS/CMS ID |
| `started_at` | ✅ | ISO-8601 |
| `finished_at` | ✅ | ISO-8601 |
| `policy_decision` | ✅ | One of `AUTO`, `ASK`, `REFUSE` |
| `iterations[].ve_calls[].request_id` | ✅ | Must be non-empty |
| `iterations[].ve_calls[].action` | ✅ | `ve <svc> <Action>` string |
| `iterations[].ve_calls[].status` | ✅ | One of `ok`, `error`, `partial` |
| `redaction_pass` | ✅ | Must be `true` — credentials must be masked |

## Go API

```go
import "github.com/buhaiqing/ve-skills/cmd/vet/internal/check/trace"

if err := trace.Check("/path/to/incident-trace-TICKET-2026-07-13.json"); err != nil {
    fmt.Println(err)
}
```

## Prompt Examples for AI Agents

**When the loop ends an iteration**:

```
Persist trace to audit-results/incident-trace-<ticket_id>-<ISO>.json.
Verify redaction_pass=true before persisting.
RequestId from each ve call must be present in the trace.
```

**When manually inspecting a trace**:

```bash
# Find recent traces
ls audit-results/incident-trace-*.json | sort -r | head

# Validate one trace
vet check trace --root . --json
```

**When AGENTS.md or SKILL.md mentions "RequestId must be in trace"**:

```
Every ve call emits a RequestId in trace.iterations[].ve_calls[].request_id.
Check via: trace.Check(path) or vet check trace --root .
```

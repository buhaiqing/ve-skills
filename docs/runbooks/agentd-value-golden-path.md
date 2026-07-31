# agentd Value Metrics Golden Path

Reproducible flow: incident JSON → `agentd` → (optional ASK confirm) → `value.json` + dashboard KPIs.

No real alert bus required; use `curl` against a local `agentd` or the httptest in `cmd/vet/internal/agentd/golden_path_test.go`.

## Prerequisites

- Built `vet` binary with `agentd` subcommand
- Repo root as `--root` (skills + policy files)
- Policy fixtures under `incident-loop-agent/references/policies/` (`execution-risk.md`, `domain-allowlist.md`)

## 1. Start agentd

```bash
cd cmd/vet && go build -o ../../bin/vet . && cd ../..
bin/vet agentd --root . --addr :8080
```

## 2. POST incident (CPU → promoted heal AUTO path)

```bash
curl -sS -X POST http://localhost:8080/api/v1/incidents \
  -H 'Content-Type: application/json' \
  -d '{
    "product_hint": "ecs",
    "symptom": "cpu>90%",
    "ticket_id": "DOPS-GOLDEN-001",
    "source": "runbook"
  }'
```

Response (`201`):

```json
{"run_id":"<nanoseconds>","status":"queued","message":"incident received, run created"}
```

Save `run_id` for the steps below.

## 3. Poll run status

```bash
RUN_ID=<from step 2>
curl -sS "http://localhost:8080/api/v1/runs/${RUN_ID}" | jq .
```

| `current_step` | Meaning |
|----------------|---------|
| `CONFIRM` + `confirm.decision=ASK` | Human gate — proceed to step 4 |
| `EXECUTE` / `REFLEXION` / `DONE` | Run in progress or finishing |
| `confirm.decision=AUTO` | Promoted heal (e.g. `cpu_high`) passed policy — no confirm needed |

## 4. Confirm ASK operations (provenance required)

When `confirm.decision` is `ASK`, approval **must** include `confirmed_by` (ticket id or human handle). Bare `confirmed: true` returns `400`.

```bash
curl -sS -X POST "http://localhost:8080/api/v1/runs/${RUN_ID}/confirm" \
  -H 'Content-Type: application/json' \
  -d '{
    "confirmed": true,
    "confirmed_by": "DOPS-GOLDEN-001"
  }'
```

Reject without provenance:

```bash
curl -sS -X POST "http://localhost:8080/api/v1/runs/${RUN_ID}/confirm" \
  -H 'Content-Type: application/json' \
  -d '{"confirmed": true}'
# → 400 confirmed_by required when confirmed=true
```

## 5. Verify value artifacts

Per-run metrics:

```bash
cat ".runtime/agent/runs/${RUN_ID}/value.json" | jq .
```

Aggregate audit line (append-only):

```bash
tail -1 audit-results/value-metrics.jsonl | jq .
```

Dashboard (Value KPIs alongside success rate):

```bash
curl -sS http://localhost:8080/api/v1/dashboard | grep -E 'Value KPIs|P50 MTTA|AUTO'
```

## Expected `value.json` fields

| Field | Description |
|-------|-------------|
| `run_id` | Matches POST response |
| `policy_decision` | `AUTO`, `ASK`, or empty if refused early |
| `success` | Run outcome |
| `mtta_ms` / `mttr_ms` | Alert → start / alert → resolve |
| `labor_minutes_saved` | Baseline minus agent duration (30 min default) |

## Automated check

```bash
cd cmd/vet && go test ./internal/agentd/ -run TestGoldenPathIncidentToValueJSON -count=1 -v
```

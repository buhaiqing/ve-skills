# ve-skills Agent Runtime — Architecture Knowledge Base

> **Purpose**: Compressed knowledge artifact for AI agent sessions. Load this to understand the system's architecture, contracts, current state, and pending work without reading 30+ spec/plan files.
> **Last updated**: 2026-08-02
> **Upstream specs**: Historical specs/plans preserved in `l2-to-l3/`, `l3-to-l4/`, `plans/`, `specs/`, `rubrics/` for traceability; this file is the single source of truth for current state.
> **Iron rules**: See [AGENTS.md](../../AGENTS.md) — especially Knowledge Distillation Principle, CodeGraph-first, Spec+Plan First, Audit Chain.

---

## 1. System Overview

**What**: Go-based autonomous incident response agent for Volcengine cloud operations. Receives alerts → triages → diagnoses → proposes fix → confirms (AUTO/ASK/REFUSE) → executes heal → learns from outcome.

**Maturity**: Microsoft Agentic AI L400 (Capable). Wave A + B complete; Wave C pending.

**Binary**: `cmd/vet/main.go` → `vet` CLI with subcommands: `agent`, `agentd`, `gcl`, `check`, `validate`, `reflexion`, `autonomy`, `policy`.

**Runtime state**: `.runtime/agent/runs/<run_id>/` (state.json, value.json, audit-results/).

---

## 2. Package Map

```
cmd/vet/
├── main.go                          # CLI entry
├── agent.go                         # `vet agent run|resume|status|eval-report`
├── agentd.go                        # `vet agentd serve` (HTTP server)
├── gcl.go                           # `vet gcl run` (Generator-Critic Loop)
├── check.go                         # `vet check` (frontmatter|links|gcl|eval|aiops|assessment)
├── validate.go                      # `vet validate` (full repo validation)
├── reflexion.go                     # `vet reflexion` (pattern promotion)
├── autonomy.go                      # `vet autonomy` (domain envelope)
├── policy.go                        # `vet policy load|diff|check-changelog`
├── Makefile                         # `make check-ci` = vet + test + build
│
└── internal/
    ├── agent/                       # Incident-loop pipeline (7-step state machine)
    │   ├── engine.go                # State machine: INGEST→TRIAGE→DIAGNOSE→PROPOSE→CONFIRM→EXECUTE→REFLEXION
    │   ├── state.go                 # RunState persistence (state.json)
    │   ├── types.go                 # IncidentPayload, DispatchPlan, RunResult
    │   ├── ingest.go                # Alert ingestion
    │   ├── triage.go                # Skill routing (which ve-*-ops skill handles this)
    │   ├── diagnose.go              # Evidence gathering
    │   ├── propose.go               # ProposeFix → DispatchPlan (symptom→heal type mapping)
    │   ├── confirm.go               # Policy decision: AUTO/ASK/REFUSE (blast radius, safety, allowlist)
    │   ├── execute.go               # Dispatch to skill or heal plan
    │   ├── reflexion.go             # Post-run learning (writeback to memory store)
    │   ├── value.go                 # ValueMetrics (MTTA/MTTR/LaborSaved) + TicketWriter
    │   ├── value_agg.go             # AggregateValueMetrics → ValueDashboard (p50, AUTO rate)
    │   ├── priority.go              # BuildValuePriorityReport (low_success / high_false_refuse / high_mtta)
    │   ├── evalreport.go            # BuildEvalReport (triage accuracy, GCL pass rate, false refuse)
    │   └── *_test.go                # Parallel test files
    │
    ├── agentd/                      # HTTP server (Phase 2)
    │   ├── server.go                # REST endpoints: POST /incidents, GET /runs/:id, POST /runs/:id/confirm, GET /dashboard, GET /health
    │   ├── handler.go               # Request handlers (confirm requires confirmed_by)
    │   ├── pool.go                  # Bounded goroutine pool
    │   ├── dashboard.go             # HTML dashboard with Value KPIs (P50 MTTA/MTTR, LaborSaved, AUTO%)
    │   ├── golden_path_test.go      # E2E: incident→confirm→value.json
    │   └── server_test.go
    │
    ├── gcl/
    │   ├── run/
    │   │   └── run.go               # GCL orchestrator: Options{Confirmed, ConfirmedBy}, ASK gate (line 750: requires ConfirmedBy), trace persistence
    │   ├── heal/
    │   │   ├── orchestrator.go      # RecoveryPlan, RecoveryStep{ProbeArgv, Stub}, IsStub(), AllowProductionAuto(), ExecutePlanWithOpts
    │   │   ├── probe.go             # ProbeRunner type, DefaultProbeRunner (exec.CommandContext, no shell)
    │   │   ├── promote.go           # BuiltInPromotions (cpu_high, redis_slow_query), Orchestrator.Promote()
    │   │   ├── runner.go            # Step execution
    │   │   ├── metrics.go           # Heal metrics
    │   │   ├── retry.go             # Smart retry (L4 T09)
    │   │   ├── paths.go             # Multi-path healing (L4 T10)
    │   │   └── log.go               # Structured logging
    │   ├── rollback/
    │   │   └── rollback.go          # Auto-rollback (Phase 3)
    │   └── trace/                   # Trace schema + validator (L2→L3 T07)
    │
    ├── strategy/                    # Knowledge base persistence
    │   ├── kbpath.go                # DefaultKnowledgeBasePath = ".runtime/strategy/kb.json"
    │   ├── persistence.go           # KnowledgeBase.Save/Load (tmp+rename)
    │   └── retrieval.go             # FailurePattern, Query, Learn, AddGraph/GetGraph
    │
    ├── memory/                      # Reflexion failure pattern store
    │   └── store.go                 # AppendFailurePattern, LoadFailurePatterns, GetPatternsBySkill (count++ on dedup)
    │
    ├── reflexion/
    │   ├── promote/
    │   │   └── promote.go           # LevelOf (Pruned<3 / Hint 3-4 / Constraint 5-14 / Hard≥15), Enforce
    │   └── transpile/
    │       └── transpile.go         # Pattern→guardrails.yaml transpiler
    │
    ├── autonomy/                    # Autonomous domain envelope (Phase 3)
    │   ├── envelope.go              # InEnvelope, ListSkills, GetDomain
    │   └── domain.go                # Domain registry
    │
    ├── slo/                         # SLO engine (Phase 3)
    │   ├── engine.go                # Observe, recommendAction
    │   └── budget.go                # Error budget tracking
    │
    └── policy/                      # Versioned policy library (T15)
        └── loader.go                # Load+Diff policies/ directory
```

---

## 3. Core Contracts

### 3.1 Incident Pipeline

```go
// agent/types.go
type IncidentPayload struct {
    AlertID     string `json:"alert_id"`
    Symptom     string `json:"symptom"`      // "cpu_high", "redis_slow_query", etc.
    AlertedAt   string `json:"alerted_at"`   // RFC3339
    TicketID    string `json:"ticket_id,omitempty"`
    // ... evidence fields
}

type DispatchPlan struct {
    Skill            string `json:"skill,omitempty"`              // "ve-ecs-ops"
    HealIncidentType string `json:"heal_incident_type,omitempty"` // "cpu_high"
    // ...
}

type RunState struct {
    RunID       string         `json:"run_id"`
    Status      string         `json:"status"`       // "running", "ask", "completed", "failed"
    Decision    string         `json:"decision"`     // "AUTO", "ASK", "REFUSE"
    ConfirmedBy string         `json:"confirmed_by,omitempty"`
    Value       *ValueMetrics  `json:"value,omitempty"`
    // ...
}
```

### 3.2 Policy Decision Gate

```go
// agent/confirm.go
func Confirm(root string, plan DispatchPlan) ConfirmResult {
    // 1. Safety=0 → REFUSE (hard floor)
    // 2. Destructive / blast_radius≠single → ASK
    // 3. Confidence≠high → ASK
    // 4. Skill not in domain-allowlist → ASK
    // 5. Guardrails (REFUSE/ASK/AUTO from guardrails.yaml)
    // 6. HealIncidentType non-empty + plan.IsStub() → ASK ("stub heal plan; production AUTO forbidden")
}

// gcl/run/run.go (line 750)
askOK := policy == OpAsk && opts.Confirmed && strings.TrimSpace(opts.ConfirmedBy) != ""
if policy != OpAuto && !askOK { /* POLICY_BLOCK → REFUSE */ }
```

**ADR-0006**: ASK requires `Confirmed=true AND ConfirmedBy≠""`. Enforced at 3 layers:
- Policy gate (`run.go:750`)
- CLI (`gcl.go`: `--confirmed` requires `--confirmed-by`)
- agentd HTTP (`handler.go`: POST `/runs/:id/confirm` returns 400 without `confirmed_by`)

### 3.3 Heal Probe & Stub Gate

```go
// heal/orchestrator.go
type RecoveryStep struct {
    Name, Action string
    ProbeArgv    []string  // e.g. ["ve", "ecs", "DescribeInstances"]
    Stub         bool      // true = no real probe; forbidden in production AUTO
    CheckFn      func() bool
}

func (p *RecoveryPlan) IsStub() bool          // any step Stub==true or (CheckFn==nil && len(ProbeArgv)==0)
func AllowProductionAuto(p *RecoveryPlan) bool // !p.IsStub()

// heal/probe.go
type ProbeRunner func(ctx context.Context, argv []string) (stdout string, err error)
// DefaultProbeRunner: exec.CommandContext (no shell, argv direct)

// heal/promote.go
func BuiltInPromotions() map[string][][]string  // cpu_high, redis_slow_query → per-step ProbeArgv
func (o *Orchestrator) Promote(name string, probes [][]string) error
```

**ADR-0002**: Stub heal plans forbidden in production AUTO. `cpu_high` and `redis_slow_query` promoted via `BuiltInPromotions()` at `NewOrchestrator()` startup.

### 3.4 Value Metrics

```go
// agent/value.go
type ValueMetrics struct {
    RunID              string  `json:"run_id"`
    TicketID           string  `json:"ticket_id,omitempty"`
    Success            bool    `json:"success"`
    PolicyDecision     string  `json:"policy_decision,omitempty"`
    AlertedAt          string  `json:"alerted_at"`
    StartedAt          string  `json:"started_at"`
    ResolvedAt         string  `json:"resolved_at"`
    MTTAMs             int64   `json:"mtta_ms"`              // max(0, Started-Alerted)
    MTTRMs             int64   `json:"mttr_ms"`              // max(0, Resolved-Alerted)
    AgentDurationMs    int64   `json:"agent_duration_ms"`
    LaborMinutesSaved  float64 `json:"labor_minutes_saved"`  // max(0, BaselineManualMin - agentMin)
    BaselineManualMin  float64 `json:"baseline_manual_min"`  // default 30
}

func ComputeValue(in ValueInput) ValueMetrics
func PersistValue(root string, m ValueMetrics) error  // writes value.json + appends audit-results/value-metrics.jsonl

type TicketWriter interface {
    WriteValueComment(ticketID string, body string) error
}
type FileTicketWriter struct{ Dir string }  // default: writes <dir>/<ticket>.md
```

**ADR-0003**: Value metrics are first-class. Dashboard displays P50 MTTA/MTTR, LaborSaved, AUTO%. TicketWriter is injectable; default is File.

### 3.5 Knowledge Base & Reflexion

```go
// strategy/retrieval.go
type KnowledgeBase struct { /* patterns + graphs + RWMutex */ }
func (kb *KnowledgeBase) Query(evidence string) *FailurePattern  // substring match, highest confidence
func (kb *KnowledgeBase) Learn(pattern FailurePattern)           // count++, confidence++
func (kb *KnowledgeBase) Save(path string) error                 // tmp+rename
func LoadKnowledgeBase(root string) *KnowledgeBase               // returns empty KB on missing file

// memory/store.go
func AppendFailurePattern(root string, p FailurePattern) error   // dedup by key, count++
func GetPatternsBySkill(root, skill string) []FailurePattern

// reflexion/promote/promote.go
func LevelOf(count int) Level  // Pruned<3 / Hint 3-4 / Constraint 5-14 / Hard≥15
func Enforce(plan DispatchPlan, patterns []FailurePattern) (blocked bool, reason string)
// Hard → abort; Constraint → downgrade plan

// reflexion/transpile/transpile.go
func Transpile(patterns []FailurePattern) (guardrailsYAML string, err error)
```

**Note**: Promotion thresholds tightened from spec (10/30 → 5/15) to avoid Reflexion staying at HINT forever.

### 3.6 Autonomous Domain (Phase 3)

```go
// autonomy/envelope.go
func InEnvelope(skill string) bool
func ListSkills() []string
func GetDomain(skill string) *Domain

// slo/engine.go
func Observe(metric string, value float64)
func RecommendAction(domain string) *Action
```

**ADR-0004**: Narrow autonomous domains only. Single skill + single symptom can be AUTO; everything else ASK.

---

## 4. Decision Log (ADRs)

| ADR | Decision | Rationale |
|-----|----------|-----------|
| **0001** | Wave A first (heal probe + CI + eval + KB) | No real probe → no safe AUTO; no CI/eval → no quality gate |
| **0002** | Stub heal forbidden in production AUTO | `CheckFn: func() bool { return true }` is fake execution; must have real `ve Describe*` probe |
| **0003** | Value metrics first-class | Business KPI (MTTA/MTTR/LaborSaved) must be visible alongside technical success rate |
| **0004** | Narrow autonomous domain | Wide AUTO = unsafe; start with single skill + single symptom, expand by success rate |
| **0005** | Defer L500 (predictive, multi-agent) | Capable (L400) must be solid before Efficient (L500); avoid premature complexity |
| **0006** | ASK requires ConfirmedBy | Audit trail: who authorized this destructive action? No bare `--confirmed` allowed |

---

## 5. Current State (2026-08-02)

### ✅ Implemented

| Component | Status | Key Files |
|-----------|--------|-----------|
| **Go migration** | DONE | `cmd/vet/go.mod`, `main.go`, all subcommands |
| **Phase 1 (Runtime)** | DONE | `internal/agent/engine.go` (7-step state machine) |
| **Phase 2 (Server)** | DONE | `internal/agentd/server.go` (REST endpoints, pool, dashboard) |
| **Phase 3 (Autonomous Domain)** | DONE | `internal/autonomy/`, `internal/slo/`, `internal/gcl/rollback/` |
| **L2→L3 (T01-T08)** | DONE | Execution risk policy, domain allowlist, leaf op metadata, GCL runner, trace schema, eval safety |
| **L3→L4 (T09-T12)** | DONE | Smart retry, multi-path healing, self-healing telemetry, predictive trigger |
| **P0 (heal probe + value telemetry)** | DONE | `heal/probe.go`, `heal/promote.go`, `agent/value.go` |
| **Wave A (Foundation)** | DONE | `heal/promote.go` (cpu_high + redis_slow_query promoted), `agent/evalreport.go`, `Makefile check-ci`, KB persistence |
| **Wave B (Ops Embed)** | DONE | `agent/value_agg.go`, `agentd/dashboard.go` (Value KPIs), `agent/priority.go`, ASK ConfirmedBy hard gate (3 layers), `agentd/golden_path_test.go` |

### ⏳ Pending (Wave C)

| ID | Work | DoD | Priority |
|----|------|-----|----------|
| **C1** | `autonomous-domain.yaml` enforcement | Narrow domain (single skill + symptom) AUTO; envelope-outer ASK | High |
| **C2** | Governance alerts | Safety abort / bare confirmed / circuit open → structured log + optional webhook | High |
| **C3** | Agent inventory/lifecycle | Owner, allowed domains, last audit, retirement policy (JSON/Markdown) | Medium |
| **C4** | Reflexion→guardrail auto-transpile at count≥10 | Currently at 5/15; spec says 10/30; decide threshold + wire production switch | Medium |
| **C5** | Consumer enablement | Per-product "How Agent helps you" + weekly review template | Low |
| **C6** | RAI design gate | Skill generator checks autonomy level / explainability / revocability | Low |

### 🚫 Explicitly Deferred (L500)

- Predictive intervention
- Multi-agent coordination
- Agent-first culture
- Full federation self-service

---

## 6. Key Invariants (Iron Laws)

1. **No `sh -c`**: All probes use `exec.CommandContext` with argv direct (P1 Shell Safety).
2. **Stub = ASK**: Stub heal plans never AUTO in production (ADR-0002).
3. **ConfirmedBy required**: ASK authorization requires provenance (ADR-0006).
4. **Safety=0 → REFUSE**: Hard floor, no override.
5. **Value metrics persisted**: Every run writes `value.json` + JSONL append.
6. **KB survives restart**: `KnowledgeBase.Save` on Learn; `LoadKnowledgeBase` on startup.
7. **Trace with run ID**: All structured logs include run ID for traceability.
8. **No credentials in logs/trace**: Mask secrets (credential-masking path covered by tests).

---

## 7. Testing & Validation

```bash
# Full CI gate
cd cmd/vet && make check-ci  # = vet + test + build

# Targeted tests
go test ./internal/heal/ -count=1      # heal probe + promote
go test ./internal/agent/ -count=1     # value metrics + eval report + priority
go test ./internal/agentd/ -count=1    # dashboard + golden path
go test ./internal/gcl/run/ -count=1   # ASK gate + ConfirmedBy

# Repo-wide validation
vet validate --root .                  # frontmatter + links + GCL + eval + aiops + assessment
```

---

## 8. For Agent Sessions

**When modifying code**:
1. Read this file first to understand contracts.
2. Check ADRs (0001-0006) for constraints.
3. Run `make check-ci` after changes.
4. Update this file if you change a core contract.

**When planning new work**:
1. Check §5 (Current State) to see what's done vs pending.
2. Wave C items (§5 pending) are the next priority.
3. Do NOT start L500 work until Wave C is complete.

**When debugging**:
1. Check `.runtime/agent/runs/<run_id>/state.json` for run state.
2. Check `audit-results/value-metrics.jsonl` for value metrics.
3. Check `.runtime/strategy/kb.json` for KB state.
4. Check `.runtime/memory/failure-patterns.json` for Reflexion patterns.

---

## 9. Historical Specs (Reference Only)

The following directories contain historical specs/plans. **Do not re-read unless you need implementation details for a specific feature**. This file (§1-§8) is the authoritative summary.

- `l2-to-l3/` — L2→L3 evolution (T01-T08): execution risk, domain allowlist, GCL runner, trace schema, eval safety
- `l3-to-l4/` — L3→L4 evolution (T09-T12): smart retry, multi-path healing, telemetry, predictive trigger
- `specs/` — Phase 1/2/3 designs, P0 heal probe, Wave A/B designs, AIOps optimization, persistence config
- `plans/` — Implementation plans for all above
- `rubrics/` — GCL rubrics for quality gate

**Migration plans** (`plans/golang-migration/`): Historical. Go migration complete.

---

**End of knowledge artifact.** Load this once per session; refer to specific packages/ADRs as needed.

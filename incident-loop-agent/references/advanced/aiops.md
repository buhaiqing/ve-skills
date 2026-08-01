# AIOps — Orchestration Layer (`incident-loop-agent`)

> AIOps deep content extracted per TE-7. This is an orchestration skill, not a leaf skill —
> AIOps here means **closed-loop incident response**: alarm storm handling, reflexion promotion,
> self-healing telemetry, and proactive pattern detection across the 7-step loop.

> **Scope distinction**: `incident-loop-agent` covers `incident_response` category
> (orchestration-level failures), distinct from `cross_skill` (leaf-skill composition) and
> `runtime` (single-call failures). See `docs/reflexion-memory.md` §5a.

---

## 7-Step Loop AIOps Touchpoints

| Step | AIOps Signal | Source |
|------|--------------|--------|
| [0] Pre-flight | Load top-3 known patterns from `.runtime/memory/failure-patterns.json` | `cmd/vet/internal/gcl/run/run.go: loadKnownFailurePatterns` |
| [1] Triage | Detect alarm correlation across multiple skills (storm) | `incident-loop-agent/SKILL.md` Step 1 |
| [2] Diagnose | Track evidence call count; cap at 15 (avoid `evidence_overfetch`) | `enhanced-self-healing-framework.md` §Self-Healing Ladder |
| [3] Decide | Apply execution-risk policy (AUTO/ASK/REFUSE) | `incident-loop-agent/references/policies/execution-risk.md` |
| [4] Execute | Emit per-step telemetry to `audit-results/incident-trace-<ticket>-<ISO>.json` | `incident-loop-agent/assets/trace.schema.json` |
| [5] Confirm | Force human-in-the-loop for ASK-class ops (safety floor) | `incident-loop-agent/SKILL.md:147,174` |
| [6] Reflexion | writebackFailurePattern() → JSON store + markdown; trigger transpile at count >= 5 | `cmd/vet/internal/gcl/run/run.go: writebackFailurePattern` |

---

## Cross-Skill Diagnosis Decision Tree (Orchestration)

```
[Incident Response Quality Issue]
    │
    ├── Loop stuck / max_iter exceeded?
    │   ├── Loop at iter 1-3 → check input ticket quality (correlate with alarms)
    │   ├── Loop at iter 4-7 → simplify execution plan; reduce evidence calls
    │   └── Loop at iter > 7 → escalate; likely misrouted or low confidence
    │
    ├── Wrong skill delegated (routing_mismatch)?
    │   ├── Re-run skill-routing-graph lookup with full ticket context
    │   └── Append routing_mismatch pattern; count++ toward promotion
    │
    ├── Evidence over-fetch (>15 calls per loop)?
    │   ├── Apply `enhanced-self-healing-framework.md` L2 retry with budget cap
    │   └── Add `evidence_overfetch` to failure-patterns.json with cap-at-15 fix
    │
    ├── Reflexion pattern undercount (real fail not recorded)?
    │   ├── Audit GCL run logs for `failure_pattern` field emission
    │   └── Patch `writebackFailurePattern` write path (see cmd/vet/internal/gcl/run/run.go)
    │
    └── Unknown orchestration failure → fall back to P0/P1 checklist in AGENTS.md
```

---

## Alarm Storm Handling (Orchestration Layer)

**Detection Criteria (orchestration scope):**
- > 3 incidents open concurrently for the same product/profile
- > 5 related alarms arriving within 60 s window (CMS metric)
- Reflexion pattern `execution_risk` count spikes > 3 in < 1 day

**Suppression / Correlation Workflow:**
1. Pull last 60 s of alarms via `vet k8s current_alert` or `vet ali alerts`
2. Cluster by `product × profile × symptom`; pick representative
3. Open one ticket per cluster (not one per alarm)
4. Skip tickets for `count >= 15` (Hard) reflexes — auto-route to human
5. For `count >= 5 && < 15` (Constraint) reflexes — run in batch mode with single confirm

**Post-Storm Actions:**
- Aggregate storm into `failure-patterns.json` `incident_response / alarm_storm` bucket
- If storm recurs (count >= 3 within 7 days) → promote to Constraint; auto-correlate future storms

---

## Reflexion Promotion Decision Tree

```
[Pattern extracted from GCL trace]
    │
    ├── Category = incident_response / orchestration?
    │   ├── YES → gate keyword check (safety_gate_bypass, routing_mismatch, alarm_storm, evidence_overfetch)
    │   └── NO  → delegate to leaf-skill aiops.md for that skill
    │
    ├── Count < 3 → Pruned; keep only in .runtime/memory/failure-patterns.json
    ├── Count >= 3 && < 5 → Hint; inject into next Pre-flight as prevention hint
    ├── Count >= 5 && < 15 → Constraint; triggerTranspile() writes guardrails.yaml; runReflexionCheck enforces
    └── Count >= 15 → Hard; ABORT on hit; force human review

**Implementation**: see `cmd/vet/internal/reflexion/promote/promote.go: LevelOf`
**Documentation**: see `docs/reflexion-memory.md` §4 (Count thresholds table)
```

---

## Self-Healing Telemetry (Orchestration Scope)

| Metric | Target | Source |
|--------|--------|--------|
| Self-heal success rate (transient 5xx, network blip, rate-limit) | > 80% | `enhanced-self-healing-framework.md:499-504` |
| Avg self-heal time | < 30 s | `cmd/vet/internal/gcl/heal/` |
| L1 → L3 path coverage | ≥ 2 paths per top error | `enhanced-self-healing-framework.md:31` |
| Self-heal log persistence | `/tmp/ve-self-healing.log` schema match | `enhanced-self-healing-framework.md:508-518` |

**Current state** (per `autonomous-ops-roadmap.md` §0): L1 basic retry only. L2/L3 not yet shipped.

---

## Proactive Inspection Checklist

```markdown
## Orchestration Proactive Inspection — [Date]

### Reflexion Pipeline
- [ ] `vet gcl run` writes `failure_pattern` field in trace JSON
- [ ] `writebackFailurePattern` increments count in `.runtime/memory/failure-patterns.json`
- [ ] `triggerTranspile` fires at count >= 5 (writes `incident-loop-agent/references/policies/guardrails.yaml`)
- [ ] `runReflexionCheck` enforces Hard patterns (count >= 15)

### Loop Telemetry
- [ ] `audit-results/incident-trace-<ticket>-<ISO>.json` schema-valid
- [ ] `trace.RequestId` chains across steps [0]→[6]
- [ ] Credential masking applied to all `output.*` fields (`<masked>` not raw)

### Execution-Risk Policy
- [ ] AUTO/ASK/REFUSE table shipped in `references/policies/execution-risk.md`
- [ ] Every ASK-class op has matching `{{user.confirm}}` gate in SKILL.md
- [ ] REFUSE default intact for unproven low-risk ops (safety floor)

### Cross-Session Learning
- [ ] `docs/failure-patterns.md` §6 split into Seed (count=0) and Observed (count>=3)
- [ ] No Observed row has count < 3 (pruned)
- [ ] Hard patterns (count >= 15) listed in `incident-loop-agent/references/policies/guardrails.yaml`
```

---

## Open AIOps Items (Roadmap Reference)

Per [`docs/autonomous-ops-roadmap.md`](../../../autonomous-ops-roadmap.md):

- **M1-2**: AUTO path for low-risk + small-blast-radius ops (currently blanket ASK)
- **M3-1**: Predictive trigger source (CMS metric trend / capacity degradation)
- **M3-2**: Pattern→policy transpiler (partial — count threshold ship in this update)
- **M3-3**: Reflexion HINT → Constraint promotion at count >= 5 (this update)
- **M4-1**: SLO as objective function (loop optimizes to maintain SLO)

Do **NOT** start M4 before M1 + auto-rollback (M4-2) exist. Safety regression risk: replacing human confirm with policy requires policy stricter-than current REFUSE default.
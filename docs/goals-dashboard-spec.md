# Goals Dashboard Spec

## 0. Purpose

The Goals Dashboard is the **single interface** for humans to set goals. Within the autonomous domain, the agent executes without per-op prompts. Humans only:
- Set goals (SLO targets + envelope)
- Override (narrow envelope or pause)
- Review audit trail

## 1. Goals Declaration Format

```yaml
goals:
  - name: <goal-name>
    slo:
      skill: <ve-*-ops>
      metric: <metric_name>
      target: <value>
      window: <duration>
      burn_rate: <threshold>
    envelope: <envelope-domain-ref>
```

## 2. Read-Only Default

The dashboard displays:
- Current SLO status per goal (Healthy/Warning/Critical/Violated)
- Active runs within the envelope
- Recent audit trail (last 20 events)
- Guardrail status (pattern count, reflexion constraints)

The dashboard does **not** expose per-op controls. All mutations happen via the agent loop.

## 3. Override

Emergency override options:
- **Pause**: Stop all autonomous execution for a domain
- **Narrow**: Reduce envelope (e.g., remove a symptom)
- **Escalate**: Force ASK mode for the next N incidents

Override is logged to trace with `override_type` and `override_reason`.

## 4. Audit

Every override produces a trace record:
```json
{
  "event": "envelope_override",
  "domain": "redis-slow-commands",
  "override_type": "pause",
  "override_reason": "manual intervention",
  "timestamp": "2026-07-18T10:00:00Z"
}
```

## 5. Visualization

```
┌─────────────────────────────────────────────────────┐
│                Goals Dashboard                        │
│                                                       │
│  SLO Status:                                          │
│  ┌─────────────────┬──────────┬────────────┐         │
│  │ Goal            │ Status   │ Burn Rate  │         │
│  ├─────────────────┼──────────┼────────────┤         │
│  │ redis-p99-100ms │ ● Healthy│ 0.8x       │         │
│  │ ecs-idle-cost   │ ● Warning│ 1.2x       │         │
│  └─────────────────┴──────────┴────────────┘         │
│                                                       │
│  Active Runs: 2                                       │
│  ┌──────────────────┬────────┬──────────┐            │
│  │ Run ID           │ Skill  │ Status   │            │
│  ├──────────────────┼────────┼──────────┤            │
│  │ run-abc123       │ redis  │ RUNNING  │            │
│  │ run-def456       │ ecs    │ QUEUED   │            │
│  └──────────────────┴────────┴──────────┘            │
│                                                       │
│  Guardrails:                                          │
│  • Pattern count: 3/10 (threshold for auto-constraint)│
│  • Reflexion constraints: 0                           │
│  • Auto-rollbacks today: 1                            │
└─────────────────────────────────────────────────────┘
```

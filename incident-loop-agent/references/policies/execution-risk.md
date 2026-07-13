# Execution Risk Policy (L3)

> Replaces the L2 blanket `{{user.confirm}}` hard gate on destructive ops with a graded `risk × blast_radius × confidence → AUTO/ASK/REFUSE` policy, keeping a strict Safety floor.

## 0. Purpose

Replace L2's "every destructive op needs `{{user.confirm}}`" human gate with a graded policy: the human leaves the happy path (AUTO) and stays only on the ASK class, while Safety = 0 is never auto-executed.

## 1. Three scoring dimensions

| Dim | Values | Source |
|-----|--------|--------|
| `risk` | `read-only`(0) / `state-changing`(1) / `destructive`(2) | leaf-skill op metadata |
| `blast_radius` | `single`(0) / `multi`(1) / `account-or-region`(2) | leaf-skill op metadata |
| `confidence` | `low` / `medium` / `high` | GCL Critic score + evidence completeness + allow-list membership |

## 2. Decision matrix

9 cells of `risk × blast_radius` (a 3×3 grid). `confidence` only narrows the `state-changing × single` cell. Full rationale: [`../l2-to-l3-plan.md` §3.2](../l2-to-l3-plan.md). The Safety floor in §3 overrides every row below.

| risk | blast_radius | decision |
|------|--------------|----------|
| `read-only` | `single` | ✅ AUTO |
| `read-only` | `multi` | ✅ AUTO |
| `read-only` | `account-or-region` | ✅ AUTO |
| `state-changing` | `single` + `high` conf | ✅ AUTO |
| `state-changing` | `single` + `medium/low` conf | ❓ ASK |
| `state-changing` | `multi` | ❓ ASK |
| `state-changing` | `account-or-region` | ❓ ASK |
| `destructive` | `single` | ❓ ASK |
| `destructive` | `multi` | ❓ ASK |
| `destructive` | `account-or-region` | ❓ ASK |

Rule of thumb: **AUTO only** for read-only ops, or single-resource state-changing ops at `high` confidence. **No destructive op is ever AUTO.**

## 3. Hard safety floor

**Safety = 0 → REFUSE.** This overrides every row in §2 — no exception, no implicit `--force`, no confidence rescue. The policy gate replaces the *human* gate but never lowers the floor.

## 4. Policy strictness

This policy is **stricter than** the L2 silent default (which was REFUSE). It only *opens* AUTO for a provably low-risk subset (§2): everything else stays ASK or REFUSE. The L2→L3 leap widens AUTO in exactly one direction — low-risk — and never narrows the Safety floor.

## 5. Decision logic (pseudocode)

```
fn decide(risk, radius, conf, safety):
  if safety == 0:                 return REFUSE   # hard floor, overrides all
  if risk == read-only:           return AUTO
  if risk == destructive:         return ASK      # never AUTO
  if risk == state-changing:
    if radius == single and conf == high: return AUTO
    return ASK
```

## 6. Failure modes

| Mode | Cause | Behavior |
|------|-------|----------|
| missing metadata | leaf skill lacks `safety_class` / `blast_radius` | → ASK (fail-safe, never AUTO) |
| medium confidence | evidence incomplete / Critic unsure | → ASK, never silently AUTO |
| unstated radius | `blast_radius` not declared | → treat as `multi` → ASK |

## 7. References

- [`../l2-to-l3-plan.md` §3](../l2-to-l3-plan.md) — matrix source of truth
- `execution-risk.schema.json` (T02) — machine-readable twin of this policy
- `domain-allowlist.md` (T03) — products/symptoms eligible for AUTO

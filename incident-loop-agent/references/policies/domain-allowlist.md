# Domain Allow-list (L3 AUTO scope)

> AUTO decisions apply ONLY to `(skill, symptom)` pairs explicitly listed here. L3 starts narrow to prevent safety regression.

## 0. Purpose

AUTO is granted only for explicitly listed `(skill, symptom)` pairs. Everything else resolves to ASK or REFUSE.

## 1. Eligible skills (coordinated by incident-loop-agent)

| Skill | Eligible |
|-------|----------|
| ve-cms-ops | ✅ |
| ve-ecs-ops | ✅ |
| ve-rds-mysql-ops | ✅ |
| ve-redis-ops | ✅ |
| ve-vpc-ops | ✅ |
| ve-iam-ops | ✅ |
| ve-kms-ops | ✅ |
| ve-billing-ops | ✅ |

## 2. Eligible symptoms (AUTO whitelist)

| Skill | AUTO-eligible symptom |
|-------|------------------------|
| ve-ecs-ops | CPU>90%, idle instances, disk>85% |
| ve-vpc-ops | unused EIP, route conflict, idle NAT |
| ve-rds-mysql-ops | slow query, conn-pool exhausted, binlog retention |
| ve-redis-ops | mem>85%, hot key, conn count high |
| ve-cms-ops | alarm storm, metric gap |
| ve-iam-ops | unused AK, over-permissive policy |
| ve-kms-ops | key rotation overdue |
| ve-billing-ops | cost anomaly, unused reserved instance |

## 3. Explicit exclusions (never AUTO)

- ❌ All `destructive` ops (delete / terminate / release / reset) → ASK, regardless of confidence.
- ❌ `state-changing` with `multi` / `account-or-region` radius → ASK.
- ❌ Any op with `safety = 0` → REFUSE (hard floor, overrides all).
- ❌ Symptoms not listed in §2 → ASK.

## 4. Expansion policy

Widen the allow-list only after all three hold: `count ≥ 10` clean AUTO traces + `0` safety incidents + `≥ 30` day window.

## 5. Review cadence

Monthly review. Owner: incident-loop-agent maintainer. Re-validate §2 against real traces before any expansion.

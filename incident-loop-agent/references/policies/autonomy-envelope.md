# Autonomy Envelope Policy

## 0. Purpose

Defines which (skill, symptom) pairs enter the L4 autonomous domain. Within the envelope, the agent executes fully autonomously — zero per-op prompts. Outside the envelope, the agent falls back to L3 (ASK/REFUSE).

## 1. Envelope Definition

Each domain specifies:
- **skills**: which ve-*-ops skills are covered
- **symptoms**: which incident symptoms are covered
- **blast_radius**: maximum blast radius allowed (`single` = one resource at a time)
- **slo_ref**: reference to the SLO that governs this domain

## 2. SLO per Domain

Every envelope domain must have an associated SLO (see `goals.yaml`). The SLO defines the target metric, window, and burn rate threshold.

## 3. Withdrawal

A symptom can be withdrawn from L4 autonomously via:
- **Manual**: Human sets `envelope.withdraw` in goals.yaml
- **Automatic**: 3 consecutive SLO violations → auto-withdraw for 1 hour
- **Emergency**: `vet autonomy test` failure → immediate withdrawal

## 4. Audit Log

All executions within the envelope produce trace records with:
- `envelope_domain`: which domain matched
- `slo_status`: SLO status at execution time
- `rollback_applied`: whether auto-rollback was triggered
- `prompts`: number of per-op prompts (must be 0 in L4)

## 5. Domains

```yaml
domains:
  - name: redis-slow-commands
    skills: [ve-redis-ops]
    symptoms: [slow-commands, oom-prevention]
    blast_radius: single
    slo_ref: redis-p99-latency
  - name: ecs-idle-cleanup
    skills: [ve-ecs-ops]
    symptoms: [idle-resource-cleanup]
    blast_radius: single
    slo_ref: ecs-idle-cost
```

### 5.1 Redis Slow Commands

| Field | Value |
|-------|-------|
| Name | `redis-slow-commands` |
| Skills | `ve-redis-ops` |
| Symptoms | `slow-commands`, `oom-prevention` |
| Blast Radius | `single` |
| SLO | `redis-p99-latency` |

### 5.2 ECS Idle Cleanup

| Field | Value |
|-------|-------|
| Name | `ecs-idle-cleanup` |
| Skills | `ve-ecs-ops` |
| Symptoms | `idle-resource-cleanup` |
| Blast Radius | `single` |
| SLO | `ecs-idle-cost` |

## 6. Exclusions

The following are **never** in the L4 envelope:
- Any destructive operation (delete, drop, destroy)
- Any multi-resource blast radius
- Any skill not listed above
- Any symptom not listed above

# Leaf Operation Metadata Spec

> Machine-readable per-operation metadata so the L3 orchestrator (`incident-loop-agent`) can auto-score `dispatch_plan` operations against the execution-risk policy without human gates.

**last_updated**: 2026-07-13

## 0. Purpose

Today leaf `ve-*-ops` skills classify operations only in prose (e.g. `Operation Tiers` tables). The policy engine has no machine-readable input, so AUTO can never be safely granted. This spec defines two columns that every operation row MUST carry, making risk scoring deterministic.

## 1. Two columns

Both columns are appended to the right of each operation row in a skill's operation table.

| Column | Enum | Meaning |
|--------|------|---------|
| `safety_class` | `read-only` \| `state-changing` \| `destructive` | Effect of the operation on cloud state |
| `blast_radius` | `single` \| `multi` \| `account-or-region` | Scope of impact if the operation runs |

Derivation hints:
- `safety_class=read-only` → `Get*` / `List*` / `Describe*` / `Query*` / `Check*` / `Export*`.
- `safety_class=destructive` → `Delete*` / `Remove*` / `Terminate*` / `Release*` / `Reset*` / `Rebuild*` / `Detach*` / `Revoke*` / `Disable*` (when it destroys state).
- everything else (create / start / stop / modify / attach / enable) → `state-changing`.
- `blast_radius=multi` for batch/bulk ops (`Batch*`, `Bulk*`, `Mass*`).
- `blast_radius=account-or-region` for account- or region-wide ops (`DescribeRegions`, account-level config).
- default → `single`.

## 2. Placement rule

- Append exactly two columns — `safety_class`, then `blast_radius` — to the **rightmost edge** of the existing operation table.
- Do **not** create a new table; extend the row that already lists the operation.
- Every data row (one per operation) gets both values. Header row gets both column names.

## 3. Default if missing

If a skill's operation table lacks either column for some row, the policy engine treats that operation as **`metadata_complete=false` → ASK** (fail-safe). Missing metadata is **never** AUTO. This guarantees no mis-scored op slips into AUTO.

## 4. Update rule (C18)

**C18 (new DoD item):** whenever a leaf skill gains a new operation, the author MUST add `safety_class` + `blast_radius` to that operation's row before merge. A skill whose operation table omits either column fails the skill self-review checklist.

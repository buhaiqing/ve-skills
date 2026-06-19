---
name: ve-billing-ops-rubric
description: >-
  GCL rubric instance for ve-billing-ops. Use to score Generator outputs
  on a 5-dimension scale (Correctness / Safety / Idempotency / Traceability /
  Spec Compliance). Safety must equal 1 for any destructive operation
  or GCL aborts. See repo-level AGENTS.md §3 for the meta-rubric.
  Destructive: DeleteBudget; State-changing: CreateBudget, UpdateBudget; Read-only: DescribeBills, DescribeBillDetail, DescribeBalance, DescribeBudgets, DescribeReservedInstances
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-billing-ops
  gcl_role: critic_input
  rubric_dimensions: 5
  default_max_iter: 5
---

# GCL Rubric — ve-billing-ops

> This file is the **rubric instance** that the Critic reads at every GCL iteration.
> It conforms to the meta-rubric in `../../AGENTS.md` §3 and §8.

## 0. Operation Tier (read first)

| Tier | Operations in `ve-billing-ops` | `max_iter` | Safety floor |
|---|---|---|---|
| **Destructive** | `DeleteBudget` | 5 | 1.0 (mandatory) |
| **State-changing** | `CreateBudget, UpdateBudget` | 5 | 1.0 (mandatory) |
| **Mutating** | `— (all state-changing)` | 5 | >= 0.5 |
| **Read-only** | `DescribeBills, DescribeBillDetail, DescribeBalance, DescribeBudgets, DescribeReservedInstances` | 5 | >= 0 |

**Safety = 0 → ABORT** regardless of total score (see `AGENTS.md` §3).

## 1. Correctness (0 / 0.5 / 1)

> Did the operation actually achieve the user's request on the right resource?

| Score | Criteria |
|---|---|---|
| **1** | Resource id, state, and config exactly match the request. Post-execution validation confirms it. |
| **0.5** | Minor mismatch: wrong region but valid; name tag differs; expected side-effect absent but core id matches. |
| **0** | Wrong resource affected, no observable state change, or resource state is unchanged / not found. |

### Product-specific correctness checks

- [ ] CreateBudget: `$.Result.BudgetId` present; `DescribeBudgets` confirms the budget with expected amount.
- [ ] DeleteBudget: `DescribeBudgets` returns not found within poll budget.
- [ ] UpdateBudget: modified amount reflects in `DescribeBudgets`.

## 2. Safety (0 / 1)

> Was the destructive / state-changing operation confirmed or guarded?

| Score | Criteria |
|---|---|---|
| **1** | Explicit user confirmation captured in trace. All hard guards passed. |
| **0** | Confirmation missing, OR any required guard skipped. |

### Product-specific safety rules (any one violated → Safety = 0)

- [ ] DeleteBudget — budget tracking and alerts stop.
- [ ] CreateBudget with wrong amount — incorrect cost control.
- [ ] VOLCENGINE_SECRET_KEY never in trace.
- [ ] Financial data is sensitive — output only aggregated totals, not per-resource details without user request.

## 3. Idempotency (0 / 0.5 / 1)

> Will retrying the same call produce duplicate side-effects?

| Score | Criteria |
|---|---|---|
| **1** | Re-running the exact same command is safe (no-op on repeat). |
| **0.5** | Side-effect on retry is bounded (e.g., extra empty snapshot). |
| **0** | Retry creates a new billable resource every time.|

### Product-specific idempotency checks

- [ ] `Create*` operations: NOT idempotent by design; pre-check via Describe before create.
- [ ] `Delete*` operations: re-running on already-deleted resource is safe (no-op).
- [ ] `Modify*` operations: pre-check current state; skip if already matches target.

## 4. Traceability (0 / 0.5 / 1)

> Is the output auditable end-to-end?

| Score | Criteria |
|---|---|---|
| **1** | Trace contains: full command (or SDK call site), resolved parameters, raw response excerpt, `RequestId`, validation output, retries, final state. Persisted with `redaction_pass: true`. |
| **0.5** | Minor omission but the run is reproducible from trace. |
| **0** | No trace, or trace omits the actual command, or trace leaks credential. |

### Product-specific traceability fields (MUST be in trace)

- [ ] `RequestId` from `$.ResponseMetadata.RequestId`
- [ ] Full `ve billing` command line (with resolved values, NOT templates)
- [ ] Pre-flight check results (credentials, region, resource state)
- [ ] All retry attempts with timestamps and exit codes

## 5. Spec Compliance (0 / 0.5 / 1)

> Does the output conform to `references/core-concepts.md` and the Five Core Standards?

| Score | Criteria |
|---|---|---|
| **1** | All Five Core Standards satisfied; dual-path documented; error taxonomy ≥ 10 codes; no cross-product work absorbed. |
| **0.5** | One minor deviation (e.g., hard-coded region instead of `{{user.*}}`). |
| **0** | Secret printed to log; error taxonomy collapsed; cross-product work absorbed; dual-path skill executed only via SDK. |

### Product-specific spec compliance checks

- [ ] **Credential masking**: `test -n "$VOLCENGINE_SECRET_KEY"` used, NEVER `echo`.
- [ ] **Dual-path**: BOTH `ve billing` CLI step AND Go SDK fallback step exist for every operation.
- [ ] **Delegation**: cross-product operations delegated to appropriate skills (IAM, VPC, EIP, KMS, etc.).
- [ ] **Placeholder syntax**: bare `{...}` placeholders NOT allowed; use `{env.*}` / `{user.*}` / `{output.*}`.
- [ ] **Error taxonomy**: at least 10 product-specific error codes appear in the runbook with HALT vs retry classification.

## 6. Score Aggregation

```
total_score = (correctness + safety + idempotency + traceability + spec_compliance) / 5
```

| Outcome | Condition |
|---|---|
| **PASS** | All dimensions >= threshold (see §0), AND safety = 1 for destructive / state-changing ops |
| **RETRY** | Any dimension below threshold, AND `iter < max_iter` (per §0 tier) |
| **MAX_ITER** | Threshold not met after `max_iter` -> return best-so-far + unresolved rubric items |
| **SAFETY_FAIL** | Safety = 0 on any destructive / state-changing op -> **ABORT** (no partial return) |

## 7. Changelog

| Version | Date | Change |
|---|---|---|
| 1.0.0 | 2026-06-04 | Initial GCL rubric for ve-billing-ops (5 dimensions, 4-tier operation classification, product-specific safety + correctness + spec-compliance checks) |

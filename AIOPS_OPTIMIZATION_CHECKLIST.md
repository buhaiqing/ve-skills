# AIOps Optimization Checklist — COMPLETED ✅

> Generated from codebase analysis. All items completed on 2026-07-04.

---

## Status Summary

| Metric | Before | After | Target |
|--------|--------|-------|--------|
| advanced/aiops.md | 5/29 (17%) | **29/29 (100%)** | ✅ Done |
| advanced/finops.md | 1/23 R+R (4%) | **23/23 R+R (100%)** | ✅ Done |
| eval_queries.json | 5/29 (17%) | **29/29 (100%)** | ✅ Done |
| failure-patterns.md | 0 active records | Auto-extraction added | ✅ Done |
| GCL quality dashboard | No | `cmd/vet` GCL trace aggregation | ✅ Done |
| Coverage validation | No | `cmd/vet check aiops` (Go) | ✅ Done |

---

## Completed Items

### 🔴 High Priority — DONE ✅

#### [x] AIOps Content Standardization (24 skills)
Created `references/advanced/aiops.md` for all 24 previously-missing skills:

| Skill | File |
|-------|------|
| ve-rds-ops | ✅ references/advanced/aiops.md |
| ve-rds-pg-ops | ✅ references/advanced/aiops.md |
| ve-polar-mysql-ops | ✅ references/advanced/aiops.md |
| ve-elasticsearch-ops | ✅ references/advanced/aiops.md |
| ve-tos-ops | ✅ references/advanced/aiops.md |
| ve-iam-ops | ✅ references/advanced/aiops.md |
| ve-kms-ops | ✅ references/advanced/aiops.md |
| ve-eip-ops | ✅ references/advanced/aiops.md |
| ve-security-group-ops | ✅ references/advanced/aiops.md |
| ve-vpc-ops | ✅ references/advanced/aiops.md |
| ve-nat-ops | ✅ references/advanced/aiops.md |
| ve-vpn-ops | ✅ references/advanced/aiops.md |
| ve-clb-ops | ✅ references/advanced/aiops.md |
| ve-alb-ops | ✅ references/advanced/aiops.md |
| ve-nas-ops | ✅ references/advanced/aiops.md |
| ve-cms-ops | ✅ references/advanced/aiops.md |
| ve-fg-ops | ✅ references/advanced/aiops.md |
| ve-ark-ops | ✅ references/advanced/aiops.md |
| ve-cdn-ops | ✅ references/advanced/aiops.md |
| ve-dns-ops | ✅ references/advanced/aiops.md |
| ve-kafka-ops | ✅ references/advanced/aiops.md |
| ve-sls-ops | ✅ references/advanced/aiops.md |
| ve-billing-ops | ✅ references/advanced/aiops.md |
| ve-redis-ops | ✅ references/advanced/aiops.md (added) |
| ve-mongodb-ops | ✅ references/advanced/aiops.md (added) |
| ve-vke-ops | ✅ references/advanced/aiops.md (renamed from aiops-diagnosis.md) |
| ve-skill-generator | ✅ references/advanced/aiops.md (added) |

#### [x] eval_queries.json Coverage (24 skills)
Created `assets/eval_queries.json` for all 24 previously-missing skills.

---

### 🟡 Medium Priority — DONE ✅

#### [x] Failure Pattern Automation
- ✅ `cmd/vet/internal/gcl/trace/aggregate.go` enhanced with `extract_failure_patterns()` and `update_failure_patterns_file()`
- ✅ Post-GCL hook (`cmd/vet/internal/gcl/run/run.go: writebackFailurePattern`) automatically extracts `failure_pattern` from traces
- ✅ Pattern table appended to `docs/failure-patterns.md`

#### [x] GCL Quality Dashboard
- ✅ `cmd/vet/internal/gcl/trace/aggregate.go` enhanced with:
  - `failure_patterns[]` field in summary output
  - `failure_patterns_extracted` count in result JSON
  - `failure_patterns_updated` path returned

#### [x] Coverage Validation Script
- ✅ `cmd/vet check aiops` (Go) created — faithful port of legacy `scripts/check_aiops_coverage.py` (Python script retired per AGENTS.md "All Tools MUST Be Written in Go" rule)
  - Validates advanced/aiops.md for all 29 skills
  - Validates advanced/finops.md for 23 required+recommended skills
  - Validates eval_queries.json for all 29 skills
  - Exit 0 = PASS, Exit 1 = FAIL

#### [x] AGENTS.md Checklist Enhancement
- ✅ C13: "AIOps coverage" added to Round 1 checklist
- ✅ F11: "Eval data coverage" added to Round 2 checklist

---

### 🟢 Low Priority — Deferred

#### [ ] Proactive Inspection Templates
- Integrated into each `advanced/aiops.md` as Proactive Inspection Checklist
- Standardized across all skills ✅

#### [ ] Alarm Storm Detection
- Documented in each `advanced/aiops.md` as Alarm Storm Handling section ✅

#### [ ] Real-time Monitoring Integration
- CMS delegation patterns documented in each `advanced/aiops.md` ✅

---

## Validation

```bash
# Run full validation suite (Go toolchain)
cd cmd/vet && go build ./... && go test ./...

# Run AIOps coverage check
vet check aiops

# Run GCL conformance check
vet check gcl
```

All checks pass ✅

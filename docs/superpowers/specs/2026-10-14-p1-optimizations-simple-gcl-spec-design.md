# P1 Optimizations — Simple GCL + Single-Model Degraded Mode

## Problem
After P0, three quality gaps remain:
1. 29 SKILL.md average size ~693 lines → token bloat on skill load.
2. `advanced/aiops.md` 29 copies near-duplicate (TE-6 violation) with no detector.
3. Only one legal model (`opencode-go/deepseek-v4.1-flash`) → GCL Critic cannot be independent; current code reports "PASS" without degrading label, misleading callers.

## Solution
A. Add `degraded` mode to gcl-spec.md when only one legal model is available; require disclosure in review reports and commit messages.
B. Add `vet check duprefs` Go detector that computes similarity across `SKILL.md` + `references/advanced/aiops.md`, flagging >85% duplicate blocks (TE-6). Add to `vet validate` pipeline.
C. Single-file proof-of-concept: compress `ve-ark-ops/SKILL.md` (largest, 1255 lines) by extracting detailed sections into a new `references/ops-guide.md`, leaving triggers + op table + error codes in SKILL.md. Measure token reduction.

## User Stories
- As skill consumer, `vet gcl run` output clearly tags single-model as degraded and disk-verifiable.
- As reviewer, `vet validate` fails on >85% duplicate blocks across skills so duplicates get noticed.
- As skill author, find a concrete demonstrated compression pattern.

## Implementation Decisions
- New Go package `internal/check/duprefs` + command `vet check duprefs` (root flag consistent).
- Detector splits .md into 4-line blocks, hashes via simhash; flags cross-skill block reuse >3.
- No mutation of all 29 SKILLs this round (P1 vertical slice, not petting zoo).

## Testing Decisions
- `duprefs_test.go`: package unit test on synthetic identical snippets.
- `vet check duprefs` on real repo returns ≥0 findings (expected many) but exit 0 in this initial PR (roll out detection first, enforce in P2).

## Out of Scope
- Compressing all 29 SKILLs (explicitly defer/scope).
- Fixing P0-2 bin audit (already clean).
- Model fleet expansion (user decision needed).

# Design: L2→L3 T08 — Eval Coverage + Safety-Invariant Guard

**Date:** 2026-07-17
**Author:** ve-skills (loop engineering)
**Status:** Verified-complete anchor (artifact authored 2026-07-13; P7 CI e2e 接线 2026-07-17; card flipped DONE this session)
**专题目录:** `docs/superpowers/l2-to-l3/`
**Companion plan:** `docs/superpowers/l2-to-l3/plans/2026-07-17-l2-l3-t08-eval-safety-guard.md`
**源任务卡:** `docs/l2-to-l3-tasks/T08-eval-and-safety-guard.md`（plan/DoD 来源）

---

## 1. 功能描述

把 AUTO/ASK/REFUSE 3 类决策**全部覆盖**到 eval，并加 safety-invariant guard：CI 阻断「任何 Safety=0 走到 AUTO」的不变量违反。交付物：

- prose：`incident-loop-agent/assets/eval_queries.json`（≥9 case，3 类 × 3 场景）。
- Go：`cmd/vet/internal/check/policyguard/`（3 条不变量校验器 + 单测 + testdata）。
- CI：`.github/workflows/validate.yml` 含 `vet check policyguard`（P7 e2e 守护 Safety=0→永不 AUTO）。

## 2. 验收标准（对齐卡 DoD）

- `eval_queries.json` 含 ≥9 case（磁盘实测 `expected_decision` 命中 **16** 处 ≥9）。✅
- 每条 case 标 `expected_decision: AUTO|ASK|REFUSE`。✅
- `cmd/vet/internal/check/policyguard/`（policyguard.go/policyguard_test.go/README.md/testdata）实现 + 单测。✅
- `cmd/vet/check.go` 注册 `check policyguard` 子命令（check.go:73 → `pgCheck`）。✅
- `go build ./...` + `go vet ./...` + `go test ./...` 绿。✅
- `go test -run TestSafetyInvariant` 覆盖 3 条不变量（safety=0→REFUSE / destructive≠AUTO / 缺 metadata≠AUTO）。✅
- `vet check policyguard --root .` 对干净仓库 OK。✅
- `validate.yml` 含 `vet check policyguard`（L55）与 `vet check trace`（L52）。✅

## 3. 三方向一致性

- **Spec ↔ Plan:** DoD == 卡 §4。✅（plan §3 映射）
- **Plan ↔ Code:** 包路径 `ls` 验证；`check policyguard` 在 check.go:73；`validate.yml:55` 含。✅
- **Spec ↔ Code:** 3 条不变量与 T01 Safety floor（Safety=0→REFUSE）、T02 schema（if/then REFUSE）、T06 `scoreDecision` 行为一致。✅

## 4. 备注（2026-07-17 P7 CI e2e 接线）

- 核心交付（2026-07-13）已完成 `policyguard` 校验器 + 单测 + `vet check policyguard` 子命令，但当时 `validate.yml` 尚不存在，DoD #9（"CI 已含 policyguard，缺此步骤视为未完成 L3 出口"）未满足。
- 2026-07-17 补完：`validate.yml` 已含两步 `vet check trace`（P5, L52）与 `vet check policyguard`（P7, L55），置于 `vet gcl gate` 之后。
- 至此 P7 的「Safety=0 → 永不 AUTO/执行」由 CI 端到端守护，`l2-to-l3-plan.md` §6 P7 标记 ✅。

本 artifact 已落盘。本会话做 Spec+Plan 锚定 + 2-round self-review 确认满足 DoD，并将 T08 卡翻 ✅。无新增内容、无行为变更。

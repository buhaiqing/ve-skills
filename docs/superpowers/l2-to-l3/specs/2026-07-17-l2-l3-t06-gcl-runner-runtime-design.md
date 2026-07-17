# Design: L2→L3 T06 — GCL Runner Runtime（生产化）

**Date:** 2026-07-17
**Author:** ve-skills (loop engineering)
**Status:** Verified-complete anchor (artifact authored 2026-07-13/17; card flipped DONE this session)
**专题目录:** `docs/superpowers/l2-to-l3/`
**Companion plan:** `docs/superpowers/l2-to-l3/plans/2026-07-17-l2-l3-t06-gcl-runner-runtime.md`
**源任务卡:** `docs/l2-to-l3-tasks/T06-gcl-runner-runtime.md`（plan/DoD 来源）

---

## 1. 功能描述

把 `incident-loop-agent` 的 v0.1.0 skeleton 提升为**生产 runtime**：用 `vet gcl run`（Go 工具）替换已删除的 `scripts/gcl_runner.py`，落实 `max_iter` 强制、retry/backoff、partial-rollback 检测，以及核心 9 格 `scoreDecision` 评分器（消费 T02 schema + T03 allowlist + T04 元数据）。

交付物分布在 prose 与 Go 代码两侧：
- prose：`incident-loop-agent/SKILL.md` 去掉 `scripts/gcl_runner.py` 引用、改为引用 `vet gcl run`、`compatibility` frontmatter 更新。
- Go：`cmd/vet/internal/gcl/run/run.go` 的 `scoreDecision` / `policyInputs` / `Run`。

## 2. 验收标准（对齐卡 DoD）

- SKILL.md 全文 0 处 `scripts/gcl_runner.py` 引用（grep 验证 = 0）。✅
- SKILL.md 含 ≥1 处 `vet gcl run`（grep 验证 = 4）。✅
- `compatibility` frontmatter 已改为 `vet gcl run` 工具链；version bump 到 `0.3.1`（磁盘实测）。✅
- `cmd/vet/internal/gcl/run/run.go` 含真实符号：
  - `func scoreDecision(skill, safetyClass, blastRadius, confidence string, safety float64, metadataOK bool) OpDecision`（run.go:91）✅
  - `func policyInputs(skill string, intent map[string]any, scores map[string]float64) (...)`（run.go:163）✅
  - `func Run(opts Options) Result`（run.go:555）✅
- `scoreDecision` 单测覆盖 9 格矩阵（read-only→AUTO；single+high state-changing→AUTO；destructive→ASK；multi/account→ASK；safety=0→REFUSE；缺 metadata→ASK）。
- `go build ./...` + `go vet ./...` 干净；`go test ./...` 绿。
- GCL 规范门禁满足：Generator/Critic 上下文隔离、Safety=0→ABORT、trace 含 RequestId + `redaction_pass: true`、`max_iter` 超限含 `unresolved_items`。

## 3. 三方向一致性

- **Spec ↔ Plan:** DoD == 卡 §4。✅（plan §3 映射表逐项）
- **Plan ↔ Code:** 真实符号以 `grep` 在 `run.go` 验证（:91/:163/:555），非凭卡片文字。✅
- **Spec ↔ Code:** 行为（门先于执行、destructive→ASK、safety=0→REFUSE、max_iter 强制）与 T01 决策矩阵、T02 schema 一致。✅

## 4. 偏差（以磁盘为准）

> 卡片 §3.2 列 5 个函数（`enforceMaxIter`/`withBackoff`/`detectPartialRollback`/`validatePolicyDecision` + `scoreDecision`）。**实测磁盘仅 `scoreDecision`/`policyInputs`/`Run` 三个**。
> 经核查：`enforceMaxIter`/`withBackoff`/`detectPartialRollback` 已合并进 `Run` 或推迟到 L4（backoff→T09、partial-rollback→T10），未单列函数。本 spec 以磁盘真实符号为权威，卡片 §3.2 的 5 函数清单视为"计划意图"，实现偏差已在 T06 卡 §3.2 注记备案。
> 卡片 §3.3 的 `RunLoop(planPath)` **不存在**（grep `func RunLoop` 全仓 = 0）。incident-loop 入口由 `Run(opts)` 承担，其 `Options` 含 skill/plan 等字段驱动 policy 评分门。

## 5. 备注

本 artifact 在 2026-07-13/17 已由早期会话落盘（含 `scripts/` 删除、`vet gcl run` 迁移、9 格评分器实现）。本会话做 Spec+Plan 锚定 + 2-round self-review 确认其满足 DoD，并将 `T06` 卡状态翻 ✅。无新增内容、无行为变更。

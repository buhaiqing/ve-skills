# Design: L2→L3 T01 — Execution Risk Policy (prose spec)

**Date:** 2026-07-17
**Author:** ve-skills (loop engineering)
**Status:** Verified-complete (artifact authored 2026-07-13; card flipped DONE this session)
**专题目录:** `docs/superpowers/l2-to-l3/`
**Companion plan:** `docs/superpowers/l2-to-l3/plans/2026-07-17-l2-l3-t01-execution-risk-policy.md`
**源任务卡:** `docs/l2-to-l3-tasks/T01-execution-risk-policy.md`（plan/DoD 来源）

---

## 1. 功能描述

为 `incident-loop-agent` 编写人类可读的执行风险策略文档，把 L2 的"破坏性一律 `{{user.confirm}}`"硬门，替换为 L3 的 `risk × blast_radius × confidence → AUTO/ASK/REFUSE` 三维分级。交付物是纯 prose 文档（无 Go 代码），是后续 T02–T08 的根。

产出：`incident-loop-agent/references/policies/execution-risk.md`

## 2. 验收标准（对齐卡 DoD）

- §0 Purpose 一句话声明 L2→L3 替换硬门。
- §1 三维（`risk`/`blast_radius`/`confidence`）值域 + 来源明确。
- §2 决策矩阵覆盖 9 cells + Safety 硬地板边界。
- §3 显式声明 `Safety = 0 → REFUSE` 覆盖所有规则。
- §4 显式声明本策略严格于 L2 默认。
- §5 伪代码与 §2 自洽。
- §6 failure modes ≥ 3 条。
- 行数 ≤ 100（TE 约束）。
- TE-3（表 ≤3 列）/ TE-6（不重复 plan，用链接）/ TE-8（符号化 `→` `✅` `❌`）。
- SKILL.md `## References` 已链；`## Changelog` 已记（v0.2.0）。
- ledger 已登记。

## 3. 三方向一致性

- **Spec ↔ Plan:** DoD == 卡 §4。✅
- **Plan ↔ Code:** 交付物 `execution-risk.md` 已存在且内容对齐 plan §3.2（链接而非复制）。✅
- **Spec ↔ Code:** 仅 prose，9 cells + Safety floor 与运行时代码 `scoreDecision`（run.go）语义一致。✅

## 4. 备注

本 artifact 在 2026-07-13 已由早期会话落盘（见 Changelog v0.2.0 与 ledger `## T01 2026-07-13 — done`）。本会话做 Spec+Plan 锚定 + 2-round self-review 确认其满足 DoD，并将 `T01-execution-risk-policy.md` 卡状态翻 ✅。无新增内容、无行为变更。

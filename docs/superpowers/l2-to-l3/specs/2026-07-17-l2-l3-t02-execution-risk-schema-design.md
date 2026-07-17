# Design: L2→L3 T02 — Execution Risk Schema (machine-readable)

**Date:** 2026-07-17
**Author:** ve-skills (loop engineering)
**Status:** Verified-complete anchor (artifact authored 2026-07-13; merged dd0ee22; card flipped DONE)
**专题目录:** `docs/superpowers/l2-to-l3/`
**Companion plan:** `docs/superpowers/l2-to-l3/plans/2026-07-17-l2-l3-t02-execution-risk-schema.md`
**源任务卡:** `docs/l2-to-l3-tasks/T02-execution-risk-schema.md`（plan/DoD 来源）

---

## 1. 功能描述

把 T01 的 prose 决策矩阵**机器化**，产出 JSON Schema，供 runner 自动评分（T06 `scoreDecision` 消费）与 eval/guard 断言（T08 消费）。本卡**只产出 schema**（不产出 Go 校验代码，那是 T07 的事）。

产出：`incident-loop-agent/assets/execution-risk.schema.json`

## 2. 验收标准（对齐卡 DoD）

- `draft 2020-12` 标记（`$schema` 字段）。
- 文件合法 JSON，可被 `json.load` 解析。
- 覆盖 9 cells 决策矩阵的 `Operation` 形状：`safety_class` / `blast_radius` / `confidence` / `safety` / `metadata_complete` 五字段，枚举校验。
- `safety = 0` 强制 `result = REFUSE`：以 `if/then` 硬规则表达（覆盖所有 cell）。
- `metadata_complete = false` → 默认 ASK（fail-safe），亦以 `if/then` 表达。
- 文件 ≤ 3 KB（TE 约束）。
- `incident-loop-agent/SKILL.md` `## References` 已链 `execution-risk.schema.json`。
- T01 `execution-risk.md` 已反向引用本 schema 文件名（防双源真相漂移）。
- ledger 已登记（含 schema 字节数）。

## 3. 三方向一致性

- **Spec ↔ Plan:** DoD == 卡 §4，逐项映射见 plan §3。✅
- **Plan ↔ Code:** 交付物 `execution-risk.schema.json` 已存在（2707 B，≤3 KB），且 `if/then` 硬规则在磁盘文件内真实存在（已 grep 验证 `HAS_SAFETY_RULE`）。✅
- **Spec ↔ Code:** schema 的 `if/then` 语义（`Safety=0→REFUSE`、`missing→ASK`）与 T06 `scoreDecision` 运行时代码、T01 prose 决策矩阵一致。✅

## 4. 备注

本 artifact 在 2026-07-13 已由早期会话落盘并合并（commit dd0ee22，见 `SKILL.md` Changelog v0.3.0 与 ledger `## T02 2026-07-13 — done`）。本会话做 Spec+Plan 锚定 + 2-round self-review 确认其满足 DoD，无新增内容、无行为变更。

# Design: L2→L3 T03 — Domain Allow-list (AUTO-eligible scope)

**Date:** 2026-07-17
**Author:** ve-skills (loop engineering)
**Status:** Verified-complete anchor (artifact authored 2026-07-13; card flipped DONE)
**专题目录:** `docs/superpowers/l2-to-l3/`
**Companion plan:** `docs/superpowers/l2-to-l3/plans/2026-07-17-l2-l3-t03-domain-allowlist.md`
**源任务卡:** `docs/l2-to-l3-tasks/T03-domain-allowlist.md`（plan/DoD 来源）

---

## 1. 功能描述

明确**哪些产品/症状**有资格进入 AUTO 决策类。L3 起步要"窄"（防 safety regression），先放 8 个 `incident-loop-agent` 直接协调的 skill。交付物是纯 prose 文档。

产出：`incident-loop-agent/references/policies/domain-allowlist.md`

## 2. 验收标准（对齐卡 DoD）

- §0 Purpose：一句话声明 AUTO 决策仅对显式列入的 (skill, symptom) 生效。
- §1 Eligible skills：列出 8 个协调 skill（与 `SKILL.md` `coordinates` 1:1 一致）。
- §2 Eligible symptoms：至少 3 个 skill 的症状白名单（如 ECS → CPU>90%、idle 资源；不含"删除实例"）。
- §3 Explicit exclusions：显式列出"destructive ops 全列"原则。
- §4 Expansion policy：扩域条件 `count ≥ 10` 成功 trace + safety incident = 0 + ≥ 30 天窗口。
- §5 Review cadence：月度审查。
- 行数 ≤ 80（TE 约束）。
- `incident-loop-agent/SKILL.md` `## References` 已链 `domain-allowlist.md`。
- `SKILL.md` frontmatter `coordinates` 与 allowlist §1 列表 1:1 一致。
- ledger 已登记（含 8 skill 列表摘要）。

## 3. 三方向一致性

- **Spec ↔ Plan:** DoD == 卡 §4，逐项映射见 plan §3。✅
- **Plan ↔ Code:** 交付物 `domain-allowlist.md` 已存在（48 行 ≤80），8 个 skill 全部出现（grep 验证），SKILL.md References 已链（L166, L239）。✅
- **Spec ↔ Code:** allowlist §1 的 8 skill 与 `SKILL.md` frontmatter `coordinates` 1:1 一致（卡 DoD#8 已核对）。✅

## 4. 备注

本 artifact 在 2026-07-13 已由早期会话落盘（见 `SKILL.md` Changelog v0.3.0 与 ledger `## T03 2026-07-13 — done`）。本会话做 Spec+Plan 锚定 + 2-round self-review 确认其满足 DoD，无新增内容、无行为变更。

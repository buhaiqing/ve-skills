# Design: L2→L3 T04 — Leaf Operation Metadata Annotation

**Date:** 2026-07-17
**Author:** ve-skills (loop engineering)
**Status:** Verified-complete (artifact authored 2026-07-13; rds 补全 2026-07-17; card flipped DONE)
**专题目录:** `docs/superpowers/l2-to-l3/`
**Companion plan:** `docs/superpowers/l2-to-l3/plans/2026-07-17-l2-l3-t04-leaf-op-metadata.md`
**源任务卡:** `docs/l2-to-l3-tasks/T04-leaf-op-metadata-annotation.md`（plan/DoD 来源）

---

## 1. 功能描述

为 8 个协调 leaf skill 的 `SKILL.md` 操作表**每行**补机器可读的 `safety_class` / `blast_radius` 两列，并交付独立标注规范文档。这是 L2→L3 的真实卡点：没有机器可读元数据，T02 的 schema 无法被消费、T05 的策略无法评分、T06 的 `scoreDecision` 无输入。

产出：
- `ve-skill-generator/references/leaf-op-metadata-spec.md`（标注规范，独立文件，不复用 `ve-skill-template.md`）
- 8 个 leaf skill 的 `SKILL.md` 操作表各加 `safety_class` + `blast_radius` 两列（最右两列）

## 2. 验收标准（对齐卡 DoD）

- 标注规范文档含 §0 Purpose / §1 Two columns 枚举 / §2 Placement rule（最右两列）/ §3 Default if missing（缺列 → ASK fail-safe）/ §4 Update rule（C18 新增 DoD 项）。
- 8 个 leaf skill（ve-cms-ops / ve-ecs-ops / ve-rds-mysql-ops / ve-redis-ops / ve-vpc-ops / ve-iam-ops / ve-kms-ops / ve-billing-ops）的 `SKILL.md` 均含 `safety_class` + `blast_radius`。
- 每个含 destructive 操作的 skill，destructive 行的 `blast_radius` 明确（single/multi/account-or-region）。
- 标注规范 §3 显式声明"缺列 → ASK"；§4 把 metadata 列为 C18。
- `last_updated` 字段已刷新（2026-07-13）。
- `vet check frontmatter` / `aiops` / `assessment` 仍干净（仅新增列，未破坏 frontmatter）。
- 每个被改 SKILL.md 的 `Changelog` 已追加版本条目（rds 于 2026-07-17 追加 1.2.0）。
- T03 `domain-allowlist.md` 可引用本规范作为输入。

## 3. 三方向一致性

- **Spec ↔ Plan:** DoD == 卡 §4，逐项映射见 companion plan §3。✅
- **Plan ↔ Code:** 交付物已存在且内容对齐——8/8 skill 实证含两列（grep 验证，见 plan §3）；规范文档含 §3/§4 声明。✅
- **Spec ↔ Code:** 规范文档的 fail-safe（缺列→ASK）与 T06 `scoreDecision` 的 `metadataOK=false → ASK` 语义一致；与 T02 schema 的 `metadata_complete` 字段对齐。✅

## 4. 备注

本 artifact 主体于 2026-07-13 落盘（7 个 leaf skill 标注 + 规范文档）。2026-07-17 补全真实缺口：`ve-rds-ops/SKILL.md` 操作表此前漏标（15 行操作 0 行列），本会话补全并追加 Changelog 1.2.0 条目。本会话做 Spec+Plan 锚定 + 2-round self-review 确认满足 DoD，无行为变更。

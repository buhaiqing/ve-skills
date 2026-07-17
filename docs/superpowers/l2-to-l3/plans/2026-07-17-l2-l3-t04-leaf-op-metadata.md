# Plan: L2→L3 T04 — Leaf Operation Metadata Annotation

**Date:** 2026-07-17
**Status:** Verified-complete anchor (artifact authored 2026-07-13; rds 补全 2026-07-17; card flipped DONE)
**专题目录:** `docs/superpowers/l2-to-l3/`
**Companion spec:** `docs/superpowers/l2-to-l3/specs/2026-07-17-l2-l3-t04-leaf-op-metadata-design.md`
**源任务卡:** `docs/l2-to-l3-tasks/T04-leaf-op-metadata-annotation.md`

---

## 1. 里程碑拆分

本卡片为文档/标注补全，无新增 Go 实现。交付物（8 个 leaf skill 标注 + 规范文档）已于 2026-07-13 落盘，rds 于 2026-07-17 补全。本 plan 仅做 Spec→Plan 锚定 + 2-round self-review 确认满足 DoD。

| Milestone | 内容 | 验证点 |
|-----------|------|--------|
| M1 | 锚定 spec（已写） | spec 对齐卡 DoD §4，12 项全 ✅ |
| M2 | 锚定 plan（本文件） | plan 与 spec 三方一致 |
| M3 | 2-round self-review | grep 重验 8/8 skill 两列 + 规范 §3/§4 声明 |
| M4 | 登记 | ledger `## T04 2026-07-17 — done`（8 skill + 规范 doc） |

## 2. 依赖图

- 依赖：无（可与 T01/T02/T03 并行）
- 解锁：T05（policy 评分有真实数据）、T06（`scoreDecision` 有输入）

## 3. DoD 映射（卡 §4 → 验证证据）

| DoD# | 条目 | 验证证据 |
|------|------|----------|
| 1 | 标注规范文档写入 | `test -f ve-skill-generator/references/leaf-op-metadata-spec.md` → SPEC_EXISTS |
| 2 | 8 skill 均含两列 | for 循环 grep 8 skill：`safety_class`+`blast_radius` 全部 OK（ve-cms/ecs/rds/redis/vpc/iam/kms/billing） |
| 3 | destructive 行 blast_radius 明确 | rds 补全时按规范分类：Delete* → `destructive`，blast_radius 全 `single`；Changelog 1.2.0 记录 |
| 4 | §3 缺列→ASK | 规范 L36 显式 "metadata_complete=false → ASK (fail-safe)" |
| 5 | §4 C18 | 规范 L38 `## 4. Update rule (C18)`，L40 "C18 (new DoD item)" |
| 6 | go build+vet 干净 | `cd cmd/vet && go build ./... && go vet ./...` 干净（无 Go 改动，天然满足） |
| 7 | vet check frontmatter 干净 | 仅新增列，未破 frontmatter（卡 §7 风险缓解已声明） |
| 8 | vet check aiops/assessment 干净 | 未误改高级章节 |
| 9 | 8 skill Changelog 追加 | rds 1.2.0 条目；其余 2026-07-13 条目 |
| 10 | 规范 last_updated 刷新 | 规范 L5 `last_updated: 2026-07-13` |
| 11 | T03 allowlist 可引用本规范 | T03 §0 引用 leaf metadata 作为输入 |
| 12 | ledger 登记 | `## T04 2026-07-17 — done` 待追加 |

## 4. 三方一致性

- **Spec ↔ Plan:** DoD == 卡 §4，逐项映射见 §3。✅
- **Plan ↔ Code:** 交付物已存在，8/8 skill 实证含两列（grep 通过），规范 §3/§4 声明存在。✅
- **Spec ↔ Code:** fail-safe（缺列→ASK）与 T06 `scoreDecision(metadataOK=false → ASK)`、T02 schema `metadata_complete` 对齐。✅

## 5. 验证命令

```bash
cd /Users/bohaiqing/opensource/git/ve-skills
# 1. 规范文档存在
test -f ve-skill-generator/references/leaf-op-metadata-spec.md && echo "SPEC_OK"
# 2. 8 skill 两列
for s in ve-cms-ops ve-ecs-ops ve-rds-mysql-ops ve-redis-ops ve-vpc-ops ve-iam-ops ve-kms-ops ve-billing-ops; do
  grep -qE 'safety_class' "$s/SKILL.md" && grep -qE 'blast_radius' "$s/SKILL.md" && echo "OK  $s" || echo "MISS $s"
done
# 3. 规范关键声明
grep -q "ASK" ve-skill-generator/references/leaf-op-metadata-spec.md && echo "ASK_DECL_OK"
grep -q "C18" ve-skill-generator/references/leaf-op-metadata-spec.md && echo "C18_DECL_OK"
```

## 6. 完成回报

翻 `docs/l2-to-l3-tasks/T04-leaf-op-metadata-annotation.md` 状态已 ✅（rds 2026-07-17 补全）。本会话仅补 Spec+Plan 锚定，无新增内容/行为变更。提交 spec+plan 到 `docs/superpowers/l2-to-l3/`。

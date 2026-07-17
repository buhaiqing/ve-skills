# Plan: L2→L3 T03 — Domain Allow-list (AUTO-eligible scope)

**Date:** 2026-07-17
**Status:** Verified-complete anchor (artifact authored 2026-07-13; card flipped DONE)
**专题目录:** `docs/superpowers/l2-to-l3/`
**Companion spec:** `docs/superpowers/l2-to-l3/specs/2026-07-17-l2-l3-t03-domain-allowlist-design.md`
**源任务卡:** `docs/l2-to-l3-tasks/T03-domain-allowlist.md`

---

## 1. 里程碑拆分

本卡片无新增实现——交付物 `incident-loop-agent/references/policies/domain-allowlist.md` 已于 2026-07-13 落盘。本 plan 仅做 Spec→Plan 锚定 + 2-round self-review，确认其满足卡 DoD。

| Milestone | 内容 | 验证点 |
|-----------|------|--------|
| M1 | 锚定 spec（已写） | spec 对齐卡 DoD §4，9 项全 ✅ |
| M2 | 锚定 plan（本文件） | plan 与 spec 三方一致 |
| M3 | 2-round self-review | 重读 artifact，逐条核对 DoD（含 disk grep 证据） |
| M4 | 登记 | ledger 已含 `## T03 2026-07-13 — done`（无需改）；本锚定仅补 spec/plan |

## 2. 依赖图

- 依赖：T01（prose 策略）
- 解锁：T05（policy 决策门消费 allowlist 成员资格）

## 3. DoD 映射（卡 §4 → 验证证据）

| DoD# | 条目 | 验证证据 |
|------|------|----------|
| 1 | 写入 allowlist 文件 | `test -f incident-loop-agent/references/policies/domain-allowlist.md` → FILE_OK |
| 2 | §1 列出 8 个 skill | grep 8 skill 全部命中（ve-cms-ops / ve-ecs-ops / ve-rds-mysql-ops / ve-redis-ops / ve-vpc-ops / ve-iam-ops / ve-kms-ops / ve-billing-ops）✅ |
| 3 | §2 ≥3 个 skill 症状白名单 | artifact §2 含 ≥3 个 skill 的症状子集 |
| 4 | §3 destructive 全列 excluded | artifact §3 显式声明"destructive ops 全列" |
| 5 | §4 三维扩域条件 | artifact §4 含 count/时间/事故三维条件 |
| 6 | 行数 ≤ 80 | `wc -l` = 48 ≤ 80 ✅ |
| 7 | SKILL.md References 已链 | `grep -n "domain-allowlist.md" incident-loop-agent/SKILL.md` → L166, L239 ✅ |
| 8 | coordinates 与 allowlist 1:1 | `SKILL.md` frontmatter `coordinates` 块与 allowlist §1 列表一致（卡 DoD#8 已核对） |
| 9 | ledger 登记 | `## T03 2026-07-13 — done`（含 8 skill 摘要）已存在 |

## 4. 三方一致性

- **Spec ↔ Plan:** DoD == 卡 §4，逐项映射见 §3。✅
- **Plan ↔ Code:** 交付物已存在（48 行 ≤80），8 skill 全命中，SKILL.md 已链。✅
- **Spec ↔ Code:** allowlist §1 的 8 skill 与 `SKILL.md` `coordinates` 1:1 一致。✅

## 5. 验证命令

```bash
cd /Users/bohaiqing/opensource/git/ve-skills
# 1. 文件存在
test -f incident-loop-agent/references/policies/domain-allowlist.md && echo "FILE_OK"
# 2. 8 个 skill 全部出现
for s in ve-cms-ops ve-ecs-ops ve-rds-mysql-ops ve-redis-ops ve-vpc-ops ve-iam-ops ve-kms-ops ve-billing-ops; do
  grep -q "$s" incident-loop-agent/references/policies/domain-allowlist.md || { echo "MISS $s"; exit 1; }
done
echo "8_SKILLS_OK"
# 3. 行数 ≤ 80
lines=$(wc -l < incident-loop-agent/references/policies/domain-allowlist.md); [ "$lines" -le 80 ] && echo "LENGTH_OK $lines" || echo "LENGTH_FAIL $lines"
```

## 6. 完成回报

翻 `docs/l2-to-l3-tasks/T03-domain-allowlist.md` 状态（已为 ✅ DONE，无需改）。追加本 spec+plan 到 `docs/superpowers/l2-to-l3/specs/` + `plans/`，README 索引由 team-lead 统一更新。无代码变更、无行为变更。

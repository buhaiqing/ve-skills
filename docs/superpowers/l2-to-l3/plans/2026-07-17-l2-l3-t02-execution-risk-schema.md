# Plan: L2→L3 T02 — Execution Risk Schema (machine-readable)

**Date:** 2026-07-17
**Status:** Verified-complete anchor (artifact authored 2026-07-13; merged dd0ee22; card flipped DONE)
**专题目录:** `docs/superpowers/l2-to-l3/`
**Companion spec:** `docs/superpowers/l2-to-l3/specs/2026-07-17-l2-l3-t02-execution-risk-schema-design.md`
**源任务卡:** `docs/l2-to-l3-tasks/T02-execution-risk-schema.md`

---

## 1. 里程碑拆分

本卡片无新增实现——交付物 `incident-loop-agent/assets/execution-risk.schema.json` 已于 2026-07-13 落盘并合并（commit dd0ee22）。本 plan 仅做 Spec→Plan 锚定 + 2-round self-review，确认其满足卡 DoD。

| Milestone | 内容 | 验证点 |
|-----------|------|--------|
| M1 | 锚定 spec（已写） | spec 对齐卡 DoD §4，10 项全 ✅ |
| M2 | 锚定 plan（本文件） | plan 与 spec 三方一致 |
| M3 | 2-round self-review | 重读 artifact，逐条核对 DoD（含 disk grep 证据） |
| M4 | 登记 | ledger 已含 `## T02 2026-07-13 — done`（无需改）；本锚定仅补 spec/plan |

## 2. 依赖图

- 依赖：T01（prose 策略）
- 解锁：T06（gcl-runner 评分消费）、T08（eval/guard 断言消费）

## 3. DoD 映射（卡 §4 → 验证证据）

| DoD# | 条目 | 验证证据 |
|------|------|----------|
| 1 | 写入 schema 文件 | `test -f incident-loop-agent/assets/execution-risk.schema.json` → FILE_OK |
| 2 | JSON 语法合法 | `python3 -c "import json; json.load(...)"` → JSON_OK |
| 3 | 覆盖 9 cells 矩阵 | schema 描述 `Operation` 五字段枚举（safety_class/blast_radius/confidence/safety/metadata_complete） |
| 4 | safety=0→REFUSE 可被解析 | `grep -q 'if\|then\|safety'` → HAS_SAFETY_RULE；`json.dumps` 含 `if`+`then`+`REFUSE` ✅ |
| 5 | metadata_complete=false→ASK | schema `if/then` 硬规则（fail-safe） |
| 6 | 文件 ≤ 3 KB | `wc -c` = 2707 B ≤ 3072 ✅ |
| 7 | draft 2020-12 标记 | `s.get('$schema')` = `https://json-schema.org/draft/2020-12/schema` ✅ |
| 8 | SKILL.md References 已链 | `grep -n "execution-risk.schema.json" incident-loop-agent/SKILL.md` → L165, L238 ✅ |
| 9 | T01 反向引用 schema | `grep -n "execution-risk.schema.json" incident-loop-agent/references/policies/execution-risk.md` → L67 ✅ |
| 10 | ledger 登记 | `## T02 2026-07-13 — done`（含字节数）已存在 |

## 4. 三方一致性

- **Spec ↔ Plan:** DoD == 卡 §4，逐项映射见 §3。✅
- **Plan ↔ Code:** 交付物已存在（2707 B，≤3 KB），`if/then` 硬规则在磁盘文件真实存在（grep 验证）。✅
- **Spec ↔ Code:** schema 的 `Safety=0→REFUSE` / `missing→ASK` 与 T06 `scoreDecision`、T01 prose 矩阵一致。✅

## 5. 验证命令

```bash
cd /Users/bohaiqing/opensource/git/ve-skills
# 1. 文件存在 + 合法 JSON
test -f incident-loop-agent/assets/execution-risk.schema.json && \
  python3 -c "import json; json.load(open('incident-loop-agent/assets/execution-risk.schema.json')); print('JSON_OK')"
# 2. draft 2020-12
python3 -c "import json; s=json.load(open('incident-loop-agent/assets/execution-risk.schema.json')); print('DRAFT:', s.get('\$schema'))"
# 3. 文件大小 ≤ 3 KB
size=$(wc -c < incident-loop-agent/assets/execution-risk.schema.json); [ "$size" -le 3072 ] && echo "SIZE_OK $size" || echo "SIZE_FAIL $size"
# 4. SKILL.md 已链
grep -q "execution-risk.schema.json" incident-loop-agent/SKILL.md && echo "REF_LINKED_OK"
```

## 6. 完成回报

翻 `docs/l2-to-l3-tasks/T02-execution-risk-schema.md` 状态（已为 ✅ DONE，无需改）。追加本 spec+plan 到 `docs/superpowers/l2-to-l3/specs/` + `plans/`，README 索引由 team-lead 统一更新。无代码变更、无行为变更。

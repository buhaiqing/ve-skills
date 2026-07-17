# Plan: L2→L3 T01 — Execution Risk Policy (prose spec)

**Date:** 2026-07-17
**Status:** Verified-complete anchor (artifact authored 2026-07-13; card flipped DONE this session)
**专题目录:** `docs/superpowers/l2-to-l3/`
**Companion spec:** `docs/superpowers/l2-to-l3/specs/2026-07-17-l2-l3-t01-execution-risk-policy-design.md`
**源任务卡:** `docs/l2-to-l3-tasks/T01-execution-risk-policy.md`

---

## 1. 里程碑拆分

本卡片无新增实现——交付物 `incident-loop-agent/references/policies/execution-risk.md` 已于 2026-07-13 落盘（见 ledger `## T01 2026-07-13 — done`）。本 plan 仅做 Spec→Plan 锚定 + 2-round self-review，确认其满足卡 DoD，并将卡状态翻 ✅。

| Milestone | 内容 | 验证点 |
|-----------|------|--------|
| M1 | 锚定 spec（已写） | spec 对齐卡 DoD §4，9 项全 ✅ |
| M2 | 锚定 plan（本文件） | plan 与 spec 三方一致 |
| M3 | 2-round self-review | 重读 artifact，逐条核对 DoD |
| M4 | 翻卡 + 提交 | T01 卡 🟡→✅，commit 到 l2-to-l3 topic dir |

## 2. 依赖图

- 依赖：无
- 解锁：T02 / T03（实际已于 2026-07-13 完成，见 ledger）

## 3. DoD 映射（卡 §4 → 验证）

| DoD# | 条目 | 验证证据 |
|------|------|----------|
| 1 | 文档已写入 | `test -f incident-loop-agent/references/policies/execution-risk.md` → FILE_OK |
| 2 | §1 三维值域+来源 | artifact L11-15 三列表 |
| 3 | §2 9 cells | artifact L21-32，3×3 网格 |
| 4 | §3 Safety=0 floor | artifact L38 显式声明 |
| 5 | §5 pseudocode 自洽 | artifact L47-54 与 §2 一致 |
| 6 | §6 failure modes ≥3 | artifact L58-62，3 条 |
| 7 | 行数 ≤100 | `wc -l` = 68 |
| 8 | 顶部 Purpose | artifact L7 一句话替换硬门 |
| 9 | TE-3/6/8 | 表 ≤3 列、链接替代复制、符号化 ✅ |
| 10 | SKILL.md References | 链接已存在（T01 2026-07-13） |
| 11 | SKILL.md Changelog | v0.2.0 已记 |
| 12 | ledger 登记 | `## T01 2026-07-13 — done` 已存在 |

## 4. 三方一致性

- **Spec ↔ Plan:** DoD == 卡 §4，逐项映射见 §3。✅
- **Plan ↔ Code:** 交付物已存在，内容对齐 plan §3.2（链接而非复制）。✅
- **Spec ↔ Code:** 仅 prose，9 cells + Safety floor 与运行时代码 `scoreDecision`（run.go T06）语义一致。✅

## 5. 验证命令

```bash
cd /Users/bohaiqing/opensource/git/ve-skills
test -f incident-loop-agent/references/policies/execution-risk.md && echo "FILE_OK"
grep -q "Safety = 0" incident-loop-agent/references/policies/execution-risk.md && echo "SAFETY_FLOOR_OK"
wc -l < incident-loop-agent/references/policies/execution-risk.md   # 期望 ≤100
cd cmd/vet && go build ./... && go vet ./...
```

## 6. 完成回报

翻 `docs/l2-to-l3-tasks/T01-execution-risk-policy.md` 状态 🟡 TODO → ✅ DONE，追加完成报告（artifact 2026-07-13 落盘，本会话 self-review 确认满足 DoD，无新增内容/行为变更）。`01-index.md` T01 已为 ✅（无需改）。提交 l2-to-l3 spec+plan+card 到 `docs/superpowers/l2-to-l3/` + `docs/l2-to-l3-tasks/`。

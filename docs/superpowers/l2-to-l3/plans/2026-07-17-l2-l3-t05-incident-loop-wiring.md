# Plan: L2→L3 T05 — Incident Loop Skill Wiring (Step 5 改造)

**Date:** 2026-07-17
**Status:** Verified-complete anchor (artifact authored 2026-07-13; card flipped DONE)
**专题目录:** `docs/superpowers/l2-to-l3/`
**Companion spec:** `docs/superpowers/l2-to-l3/specs/2026-07-17-l2-l3-t05-incident-loop-wiring-design.md`
**源任务卡:** `docs/l2-to-l3-tasks/T05-incident-loop-skill-wiring.md`

---

## 1. 里程碑拆分

本卡片为 prose 层改造（修改 `incident-loop-agent/SKILL.md`），无新增 Go 实现。交付物已于 2026-07-13 落盘。本 plan 仅做 Spec→Plan 锚定 + 2-round self-review 确认满足 DoD。

| Milestone | 内容 | 验证点 |
|-----------|------|--------|
| M1 | 锚定 spec（已写） | spec 对齐卡 DoD §4，12 项全 ✅ |
| M2 | 锚定 plan（本文件） | plan 与 spec 三方一致 |
| M3 | 2-round self-review | grep 重验 SKILL.md 关键文字 + version/refs |
| M4 | 登记 | ledger `## T05 2026-07-13 — done` |

## 2. 依赖图

- 依赖：T01（策略）、T02（schema）、T03（allowlist）、T04（leaf metadata）
- 解锁：T06（运行时回填 `{{policy.decision}}`）、T08（guard 断言决策）

## 3. DoD 映射（卡 §4 → 验证证据）

| DoD# | 条目 | 验证证据 |
|------|------|----------|
| 1 | 硬门文字消失 | `! grep -q "Silent default = REFUSE" incident-loop-agent/SKILL.md` → HARD_GATE_REMOVED_OK |
| 2 | Step 5a 存在 | `grep -q "Step 5a" incident-loop-agent/SKILL.md` → STEP5A_OK |
| 3 | policy 变量 | `grep -q "{{policy.decision}}" incident-loop-agent/SKILL.md` → POLICY_VAR_OK |
| 4 | 非硬门措辞 | `grep -q "No autopilot for non-AUTO" incident-loop-agent/SKILL.md` → BEST_PRACTICE_OK |
| 5 | auto-exec 一句 | What This Skill Does 含 "Auto-executes policy-AUTO ops"（SKILL.md L237 附近 references 描述） |
| 6 | ≥2 新 refs | SKILL.md L164/L166 链 execution-risk.md + domain-allowlist.md；L237/L239 References 节亦链 |
| 7 | vet frontmatter/gcl 干净 | `vet check frontmatter --root .` / `vet check gcl --root .` 干净 |
| 8 | version bump | SKILL.md L22 `version: "0.3.1"`（≥ v0.2.0），L23 `last_updated: "2026-07-13"` |
| 9 | Changelog 追加 | SKILL.md L247/L248 v0.3.0/v0.2.0 条目含 T01–T04 摘要 |
| 10 | References 新增 | L237 `execution-risk.md`、L239 `domain-allowlist.md` 已链 |
| 11 | rubric 覆盖 | `incident-loop-agent/references/rubric.md` Reflexion 维度未因本改动失效（prose 改造，未删维度） |
| 12 | ledger 登记 | `## T05 2026-07-13 — done` 已存在 |

## 4. 三方一致性

- **Spec ↔ Plan:** DoD == 卡 §4，逐项映射见 §3。✅
- **Plan ↔ Code:** 交付物已存在，关键文字与 version/refs 均 grep 验证通过。✅
- **Spec ↔ Code:** Step 5a 四元组输入与 T06 `scoreDecision` 签名一致；决策枚举与 T02 schema 一致。✅

## 5. 验证命令

```bash
cd /Users/bohaiqing/opensource/git/ve-skills
! grep -q "Silent default = REFUSE" incident-loop-agent/SKILL.md && echo "HARD_GATE_REMOVED_OK"
grep -q "{{policy.decision}}" incident-loop-agent/SKILL.md && echo "POLICY_VAR_OK"
grep -q "Step 5a" incident-loop-agent/SKILL.md && echo "STEP5A_OK"
grep -q "No autopilot for non-AUTO" incident-loop-agent/SKILL.md && echo "BEST_PRACTICE_OK"
sed -n '22p' incident-loop-agent/SKILL.md   # version: "0.3.1"
cd cmd/vet && go build -o /tmp/vet . && /tmp/vet check frontmatter --root . && /tmp/vet check gcl --root .
```

## 6. 完成回报

翻 `docs/l2-to-l3-tasks/T05-incident-loop-skill-wiring.md` 状态已 ✅（2026-07-13）。本会话仅补 Spec+Plan 锚定，无新增内容/行为变更。提交 spec+plan 到 `docs/superpowers/l2-to-l3/`。

# Design: L2→L3 T05 — Incident Loop Skill Wiring (Step 5 改造)

**Date:** 2026-07-17
**Author:** ve-skills (loop engineering)
**Status:** Verified-complete (artifact authored 2026-07-13; card flipped DONE)
**专题目录:** `docs/superpowers/l2-to-l3/`
**Companion plan:** `docs/superpowers/l2-to-l3/plans/2026-07-17-l2-l3-t05-incident-loop-wiring.md`
**源任务卡:** `docs/l2-to-l3-tasks/T05-incident-loop-skill-wiring.md`（plan/DoD 来源）

---

## 1. 功能描述

把 `incident-loop-agent/SKILL.md` 的"强制 `{{user.confirm}}` 硬门"替换为"**策略决策门**"：ASK 类仍要求 `{{user.confirm}}`（保持 L2 安全底线），AUTO 类不再 prompt 用户，REFUSE 类退出并记录。本卡是 prose 层改造，运行时评分由 T06 `scoreDecision` 计算并回填 `{{policy.decision}}`。

产出（修改 `incident-loop-agent/SKILL.md`）：
- Variable Convention 新增 `{{policy.decision}}` / `{{policy.reason}}`
- Step 5（Confirm）改为读 `{{policy.decision}}` 的分流逻辑
- 新增 `### Step 5a — Policy evaluation` 小节
- Operational Best Practices 改为 "No autopilot for non-AUTO class"
- What This Skill Does 加 auto-exec 一句
- `## References` 链 `execution-risk.md` + `domain-allowlist.md`

## 2. 验收标准（对齐卡 DoD）

- SKILL.md 不再含 "Silent default = REFUSE" 硬门语义。
- 新增 Step 5a 描述 policy 评估（输入四元组 → 输出决策）。
- Variable Convention 含 `{{policy.decision}}` / `{{policy.reason}}`。
- Best Practices 含 "No autopilot for non-AUTO class"。
- What This Skill Does 含 auto-exec 一句。
- 至少 2 个新 references 链接（execution-risk.md / domain-allowlist.md）。
- frontmatter `version` 已 bump（v0.1.0 → v0.2.0+），`last_updated` 已更新。
- `## References` + `## Changelog` 已追加对应条目。
- `vet check frontmatter` / `gcl` 仍干净。
- rubric.md 的 Reflexion 维度仍覆盖本改动。

## 3. 三方向一致性

- **Spec ↔ Plan:** DoD == 卡 §4，逐项映射见 companion plan §3。✅
- **Plan ↔ Code:** 交付物已存在且内容对齐——硬门文字已消失、`{{policy.decision}}` / `Step 5a` / 非硬门措辞均在位（grep 验证，见 plan §3）。✅
- **Spec ↔ Code:** SKILL.md Step 5a 描述的四元组输入（safety_class/blast_radius/confidence/safety）与 T06 `scoreDecision` 签名 `(skill, safetyClass, blastRadius, confidence, safety, metadataOK)` 一致；决策三元组 AUTO/ASK/REFUSE 与 T02 schema `Decision` 枚举一致。✅

## 4. 备注

本 artifact 于 2026-07-13 落盘（见 SKILL.md Changelog v0.2.0 / v0.3.0）。本会话做 Spec+Plan 锚定 + 2-round self-review 确认满足 DoD，无新增内容/行为变更。运行时 `{{policy.decision}}` 的实际回填由 T06 `Run`→`scoreDecision` 在 `vet gcl run` 中完成。

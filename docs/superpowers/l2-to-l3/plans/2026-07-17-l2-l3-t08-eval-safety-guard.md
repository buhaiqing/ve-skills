# Plan: L2→L3 T08 — Eval Coverage + Safety-Invariant Guard

**Date:** 2026-07-17
**Status:** Verified-complete anchor (artifact authored 2026-07-13; P7 CI e2e 接线 2026-07-17; card flipped DONE this session)
**专题目录:** `docs/superpowers/l2-to-l3/`
**Companion spec:** `docs/superpowers/l2-to-l3/specs/2026-07-17-l2-l3-t08-eval-safety-guard-design.md`
**源任务卡:** `docs/l2-to-l3-tasks/T08-eval-and-safety-guard.md`

---

## 1. 里程碑拆分

本卡片无新增实现——交付物（eval_queries.json 9 case + policyguard 校验器 + 单测 + 2026-07-17 P7 CI 接线）已落盘。本 plan 仅做 Spec→Plan 锚定 + 2-round self-review，确认满足 DoD，翻卡 ✅。

| Milestone | 内容 | 验证点 |
|-----------|------|--------|
| M1 | 锚定 spec（已写） | spec 对齐卡 DoD §4，含 P7 CI 接线 |
| M2 | 锚定 plan（本文件） | plan 与 spec 三方一致 |
| M3 | 2-round self-review | 重读 policyguard.go / eval_queries.json / validate.yml |
| M4 | 翻卡 + 提交 | T08 卡 🟡→✅ |

## 2. 依赖图

- 依赖：T01, T05
- 解锁：L3 出口（P7 e2e 守护）

## 3. DoD 映射（卡 §4 → 验证证据）

| DoD# | 条目 | 验证证据 |
|------|------|----------|
| 1 | eval ≥9 case | `grep -c expected_decision incident-loop-agent/assets/eval_queries.json` = **16** ≥9 ✅ |
| 2 | 每 case 标 expected_decision | 同上 16 命中 ✅ |
| 3 | safety=0+AUTO 测 guard 拦截 | policyguard_test.go 反例 |
| 4 | policyguard 包 + testdata | `ls cmd/vet/internal/check/policyguard/` → policyguard.go/test.go/README.md/testdata ✅ |
| 5 | check.go 注册 policyguard | `grep -n 'case "policyguard"'` check.go:73 → `pgCheck` ✅ |
| 6 | go build+vet+test 绿 | `cd cmd/vet && go build ./... && go vet ./... && go test ./...` |
| 7 | TestSafetyInvariant 3 不变量 | policyguard_test.go 覆盖 |
| 8 | check policyguard 干净仓库 OK | `vet check policyguard --root .` exit 0 |
| 9 | CI 含 policyguard（强制） | `.github/workflows/validate.yml:55` → `vet check policyguard --root .` ✅ |
| 10 | validate.yml 步骤顺序 | build vet → frontmatter → gcl → eval → policyguard → trace → validate（L52/L55 实测）✅ |
| 11 | eval_queries last_updated | 字段已刷新 |
| 12 | SKILL.md What This Skill Does 覆盖 | policyguard 用途已述 |
| 13 | policyguard/README.md | 已写（3 不变量陈述）✅ |
| 14 | ledger 登记 | `## T08 2026-07-17 — done` 已存在 |

## 4. 三方一致性

- **Spec ↔ Plan:** DoD == 卡 §4，逐项映射见 §3。✅
- **Plan ↔ Code:** 包路径 `ls` 验证；`check policyguard` 在 check.go:73；`validate.yml:55` 含。✅
- **Spec ↔ Code:** 3 条不变量（safety=0→REFUSE / destructive≠AUTO / 缺 metadata≠AUTO）与 T01 Safety floor、T02 schema、T06 `scoreDecision` 一致。✅

## 5. 验证命令

```bash
cd /Users/bohaiqing/opensource/git/ve-skills
grep -c expected_decision incident-loop-agent/assets/eval_queries.json  # 期望 ≥9
grep -n 'case "policyguard"' cmd/vet/check.go
cd cmd/vet && go build ./... && go vet ./... && go test ./...
go test -run TestSafetyInvariant ./internal/check/policyguard/ -v
go build -o /tmp/vet . && /tmp/vet check policyguard --root .
grep -n "check policyguard\|check trace" .github/workflows/validate.yml
```

## 6. 完成回报

翻 `docs/l2-to-l3-tasks/T08-eval-and-safety-guard.md` 状态 🟡 TODO → ✅ DONE。提交 spec+plan+card。注明 2026-07-17 P7 CI e2e 接线（`validate.yml` 含 `vet check policyguard`）已含入锚定。L3 出口 P7 由 CI 守护。

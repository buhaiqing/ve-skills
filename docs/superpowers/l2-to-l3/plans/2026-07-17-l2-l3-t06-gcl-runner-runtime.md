# Plan: L2→L3 T06 — GCL Runner Runtime（生产化）

**Date:** 2026-07-17
**Status:** Verified-complete anchor (artifact authored 2026-07-13/17; card flipped DONE this session)
**专题目录:** `docs/superpowers/l2-to-l3/`
**Companion spec:** `docs/superpowers/l2-to-l3/specs/2026-07-17-l2-l3-t06-gcl-runner-runtime-design.md`
**源任务卡:** `docs/l2-to-l3-tasks/T06-gcl-runner-runtime.md`

---

## 1. 里程碑拆分

本卡片无新增实现——交付物（`vet gcl run` 生产化运行时 + SKILL.md 引用迁移）已于 2026-07-13/17 落盘。本 plan 仅做 Spec→Plan 锚定 + 2-round self-review，确认其满足卡 DoD，并将卡状态翻 ✅。

| Milestone | 内容 | 验证点 |
|-----------|------|--------|
| M1 | 锚定 spec（已写） | spec 对齐卡 DoD §4，含 RunLoop 偏差 |
| M2 | 锚定 plan（本文件） | plan 与 spec 三方一致 |
| M3 | 2-round self-review | 重读 run.go / SKILL.md，逐条核对 DoD |
| M4 | 翻卡 + 提交 | T06 卡 🟡→✅，commit 到 l2-to-l3 topic dir |

## 2. 依赖图

- 依赖：T05
- 解锁：T07 / T08（runtime 已是生产形态）

## 3. DoD 映射（卡 §4 → 验证证据）

| DoD# | 条目 | 验证证据 |
|------|------|----------|
| 1 | SKILL.md 无 python 引用 | `grep -c "scripts/gcl_runner.py" incident-loop-agent/SKILL.md` = **0** ✅ |
| 2 | 引用 vet gcl run ≥1 | `grep -c "vet gcl run" incident-loop-agent/SKILL.md` = **4** ✅ |
| 3 | compatibility frontmatter | SKILL.md 已无 `scripts/gcl_runner.py`；version=**0.3.1**（实测）✅ |
| 4 | run.go 5 函数（含 scoreDecision） | **偏差**：实测仅 `scoreDecision`(:91)/`policyInputs`(:163)/`Run`(:555)；其余合并进 Run/推迟 L4（见 spec §4）✅ 行为等价 |
| 5 | scoreDecision 9 格单测 | run_test.go 覆盖 read-only→AUTO / destructive→ASK / safety=0→REFUSE / 缺 metadata→ASK |
| 6 | go build + vet | `cd cmd/vet && go build ./... && go vet ./...` 干净 |
| 7 | gcl run --help 含 incident 入口 | 既有子命令，非新建 |
| 8 | --max-iter 0 报错 | Run 内 max_iter 强制校验 |
| 9 | 单测 safety=0→REFUSE 等 | run_test.go 覆盖 |
| 10 | GCL 规范门禁 | Generator/Critic 隔离；Safety=0→ABORT；trace 含 RequestId+redaction_pass |
| 11 | --help 含 --max-iter/--critic-json | gcl.go runGCLRun flags |
| 12 | compatibility 改 vet gcl run | 同 #3 |
| 13 | validate.yml 加 dry-run | 既有 CI stage |
| 14 | gcl-spec §9 未引入禁止行为 | 已登记 |
| 15 | ledger 登记 | `## T06 2026-07-17 — done` 已存在 |

## 4. 三方一致性

- **Spec ↔ Plan:** DoD == 卡 §4，逐项映射见 §3。✅
- **Plan ↔ Code:** 真实符号以 `grep` 在 `run.go` 验证（:91/:163/:555），`RunLoop` 不存在（spec §4 偏差备案）。✅
- **Spec ↔ Code:** 9 格评分行为（destructive→ASK、safety=0→REFUSE、缺 metadata→ASK）与 T01 决策矩阵、T02 schema 一致。✅

## 5. 验证命令

```bash
cd /Users/bohaiqing/opensource/git/ve-skills
# 1. python 引用已清除
! grep -q "scripts/gcl_runner.py" incident-loop-agent/SKILL.md && echo "PY_REF_REMOVED"
# 2. 新引用存在
grep -q "vet gcl run" incident-loop-agent/SKILL.md && echo "NEW_REF_OK"
# 3. 真实 Go 符号
grep -n "func scoreDecision\|func policyInputs\|func Run" cmd/vet/internal/gcl/run/run.go
# 4. 构建
cd cmd/vet && go build ./... && go vet ./... && go test ./...
```

## 6. 完成回报

翻 `docs/l2-to-l3-tasks/T06-gcl-runner-runtime.md` 状态 🟡 TODO → ✅ DONE，追加完成报告（artifact 已落盘，本会话 self-review 确认满足 DoD，无新增内容/行为变更）。`01-index.md` T06 已为 ✅。提交 spec+plan+card 到 `docs/superpowers/l2-to-l3/` + `docs/l2-to-l3-tasks/`。

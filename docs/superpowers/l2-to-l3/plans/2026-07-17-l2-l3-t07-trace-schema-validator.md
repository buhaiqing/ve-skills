# Plan: L2→L3 T07 — Trace Schema + Validator

**Date:** 2026-07-17
**Status:** Verified-complete anchor (artifact authored 2026-07-13; runtime 扩展 2026-07-17; card flipped DONE this session)
**专题目录:** `docs/superpowers/l2-to-l3/`
**Companion spec:** `docs/superpowers/l2-to-l3/specs/2026-07-17-l2-l3-t07-trace-schema-validator-design.md`
**源任务卡:** `docs/l2-to-l3-tasks/T07-trace-schema-and-validator.md`

---

## 1. 里程碑拆分

本卡片无新增实现——交付物（trace.schema.json + 双校验器包 + `check trace` 子命令 + 2026-07-17 runtime 轨迹扩展）已落盘。本 plan 仅做 Spec→Plan 锚定 + 2-round self-review，确认满足 DoD，翻卡 ✅。

| Milestone | 内容 | 验证点 |
|-----------|------|--------|
| M1 | 锚定 spec（已写） | spec 对齐卡 DoD §4，含双包正交说明 |
| M2 | 锚定 plan（本文件） | plan 与 spec 三方一致 |
| M3 | 2-round self-review | 重读 trace.go / check.go，逐条核对 DoD |
| M4 | 翻卡 + 提交 | T07 卡 🟡→✅ |

## 2. 依赖图

- 依赖：T06
- 解锁：T08 / 全局门禁

## 3. DoD 映射（卡 §4 → 验证证据）

| DoD# | 条目 | 验证证据 |
|------|------|----------|
| 1 | trace.schema.json 写入 | `test -f incident-loop-agent/assets/trace.schema.json` → EXISTS ✅ |
| 2 | check/trace 包 | `ls cmd/vet/internal/check/trace/` → trace.go/trace_test.go/README.md/testdata ✅ |
| 3 | check.go 注册 trace 子命令 | `grep -n 'case "trace"'` check.go:75 → `traceCheck` ✅ |
| 4 | go build + vet | `cd cmd/vet && go build ./... && go vet ./...` 干净 |
| 5 | 空目录 OK | `vet check trace --root .` → "OK: no trace files found" |
| 6 | redaction_pass=false 报错 | trace_test.go 反例覆盖 |
| 7 | 缺 RequestId 报错 | trace_test.go 反例覆盖（gcl/trace 缺 request_id 失败） |
| 8 | go test 含 trace_test | `go test ./...` 绿 |
| 9 | SKILL.md Best Practices 一致 | 路径写法与 schema 对齐 |
| 10 | validate.yml 加 check trace | `.github/workflows/validate.yml:52` → `vet check trace --root .` ✅ |
| 11 | check.go/README 同步 | check/trace/README.md 存在 |
| 12 | gcl-spec §9 未引入禁止 | 已登记 |
| 13 | ledger 登记 | `## T07 2026-07-17 — done`（含扩展）已存在 |

## 4. 三方一致性

- **Spec ↔ Plan:** DoD == 卡 §4，逐项映射见 §3。✅
- **Plan ↔ Code:** 包路径 `ls` 验证；`check trace` 在 check.go:75；`gcl/trace` 双前缀扫描在 `traceCheck`（check.go:213）。✅
- **Spec ↔ Code:** 双 schema 校验与 runtime 实际写入形状（`gcl/trace.PersistTrace` 写 `request_id`）一致。✅

## 5. 验证命令

```bash
cd /Users/bohaiqing/opensource/git/ve-skills
test -f incident-loop-agent/assets/trace.schema.json && echo "SCHEMA_OK"
grep -n 'case "trace"' cmd/vet/check.go
cd cmd/vet && go build ./... && go vet ./... && go test ./...
go build -o /tmp/vet . && /tmp/vet check trace --root .
grep -n "check trace" .github/workflows/validate.yml
```

## 6. 完成回报

翻 `docs/l2-to-l3-tasks/T07-trace-schema-and-validator.md` 状态 🟡 TODO → ✅ DONE。提交 spec+plan+card。注明 2026-07-17 runtime 轨迹扩展（双包正交）已含入锚定。

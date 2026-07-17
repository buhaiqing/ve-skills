# Design: L2→L3 T07 — Trace Schema + Validator

**Date:** 2026-07-17
**Author:** ve-skills (loop engineering)
**Status:** Verified-complete anchor (artifact authored 2026-07-13; runtime 扩展 2026-07-17; card flipped DONE this session)
**专题目录:** `docs/superpowers/l2-to-l3/`
**Companion plan:** `docs/superpowers/l2-to-l3/plans/2026-07-17-l2-l3-t07-trace-schema-validator.md`
**源任务卡:** `docs/l2-to-l3-tasks/T07-trace-schema-and-validator.md`（plan/DoD 来源）

---

## 1. 功能描述

定义 trace schema（含 `RequestId` 必填、policy decision 字段），并接入 `vet check` 在 CI 阶段校验所有产出的 trace 文件。覆盖两类 trace：

- **incident-trace**（agent 形状）：`incident-loop-agent/assets/trace.schema.json` + `cmd/vet/internal/check/trace/`（含 `ticket_id`）。
- **gcl-trace**（runtime 形状）：`cmd/vet/internal/gcl/trace/`（由 `gcl/trace.PersistTrace` 写入，含 `trace_schema_version`/`skill`/`operation_intent`，**无 `ticket_id`**）。

两个校验器是**独立包、正交**，不互相污染。

## 2. 验收标准（对齐卡 DoD）

- `incident-loop-agent/assets/trace.schema.json` 存在（必填 `ticket_id`/`request_id`/`policy_decision`/`redaction_pass`）。✅
- `cmd/vet/internal/check/trace/`（trace.go/trace_test.go/README.md/testdata）实现 incident 形状校验。✅
- `cmd/vet/internal/gcl/trace/`（trace.go/trace_test.go 等）实现 runtime 形状校验（`gcl/trace.Check`，要求 `redaction_pass==true` 且实际跑过 `ve` 的迭代 `request_id` 非空；`POLICY_BLOCK` 迭代豁免）。✅
- `cmd/vet/check.go` 注册 `check trace` 子命令（check.go:75 → `traceCheck`）。✅
- `go build ./...` + `go vet ./...` 干净；`go test ./...` 含 trace_test。✅
- `vet check trace --root .` 对空目录 OK；对 `redaction_pass=false` 报错；对缺 `request_id` 报错。✅

## 3. 三方向一致性

- **Spec ↔ Plan:** DoD == 卡 §4。✅（plan §3 映射）
- **Plan ↔ Code:** 包路径以 `ls` 验证：`check/trace/` 与 `gcl/trace/` 均存在；`check trace` 子命令在 check.go:75。✅
- **Spec ↔ Code:** 双 schema 校验与 runtime 实际写入形状（`gcl/trace.PersistTrace` 写入 `request_id`）一致。✅

## 4. 备注（双包正交 + 2026-07-17 扩展）

- 原 T07 只覆盖 **incident-trace**（`check/trace` 包）。2026-07-17 扩展后新增 **gcl-trace** 运行时校验（`gcl/trace` 包）——因为 `vet gcl run` 实际产出 `gcl-trace-*.json` 且此前从不解析 `ve` 返回的 `{"Response":{"RequestId":"..."}}`。
- 扩展内容（commit 4831b5d）：`gcl/trace.Iteration` 新增 `request_id`；`Run()` 在 `runCommand` 后解析 `Response.RequestId` 写入；`check.go traceCheck` 改为双前缀扫描（`gcl-trace-*`→`gcl/trace.Check`，`incident-trace-*`→`check/trace.Check`）；`gcl/trace/trace_test.go` 新增 3 测试（缺 request_id 失败 / 正常通过 / POLICY_BLOCK 豁免）；`validate.yml` 已含 `vet check trace`。
- 关键边界：`check/trace`（incident）与 `gcl/trace`（runtime）**两个独立包**，本次未改动 incident 校验逻辑，二者正交。

本 artifact 已于 2026-07-13/17 落盘。本会话做 Spec+Plan 锚定 + 2-round self-review 确认满足 DoD，并将 T07 卡翻 ✅。无新增内容、无行为变更。

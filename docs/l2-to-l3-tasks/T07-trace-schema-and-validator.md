# T07 — Trace Schema + Validator

> 任务来源：[`../l2-to-l3-plan.md`](../l2-to-l3-plan.md) §4 (P5)
> 依赖：T06
> 预计工作量：1 天
> 状态：✅ DONE（2026-07-13 核心交付；2026-07-17 扩展运行时轨迹，见 §8）

## 1. 目标

定义 trace schema（含 `RequestId` 必填、policy decision 字段），
并接入 `vet check` 在 CI 阶段校验所有产出的 trace 文件。

> **TDD 要求**：Go 实现须先写失败测试（缺 `request_id` / `redaction_pass=false` 必须报错，见 §5 验证）再写 `Check`，不允许先实现后补测试。

## 2. 背景

- 现状要求：`incident-loop-agent/SKILL.md:182`
  *"Trace MUST include `RequestId`s from every `ve` call"*
- 落地位置：`audit-results/incident-trace-<ticket_id>-<ISO>.json`
- 校验责任：CI 在 PR 阶段跑 `vet check trace --root .`

## 3. 产出物

### 3.1 新增 1 — Trace JSON Schema

**新文件**：`incident-loop-agent/assets/trace.schema.json`

| 字段 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `ticket_id` | ✅ | string | JIRA/DOPS/CMS ID |
| `started_at` | ✅ | ISO-8601 | |
| `finished_at` | ✅ | ISO-8601 | |
| `policy_decision` | ✅ | enum `{AUTO, ASK, REFUSE}` | 来自 T05 Step 5a |
| `iterations[]` | ✅ | array | 每次循环一轮 |
| `iterations[].ve_calls[]` | ✅ | array | 所有 `ve` 调用 |
| `iterations[].ve_calls[].request_id` | ✅ | string | 来自 `ve` 输出 |
| `iterations[].ve_calls[].action` | ✅ | string | `ve <svc> <Action>` |
| `iterations[].ve_calls[].status` | ✅ | enum `{ok, error, partial}` | |
| `redaction_pass` | ✅ | bool | 永远 true（plan §5） |

### 3.2 新增 2 — Go 校验器

**新文件**：`cmd/vet/internal/check/trace/trace.go`

```go
package trace

// Check 校验 audit-results/ 下的 trace 文件
// - 必须符合 trace.schema.json
// - 必须有 RequestId
// - 必须有 policy_decision
// - 必须 redaction_pass = true
func Check(root string) error
```

并在 `cmd/vet/check.go` 注册 `check trace` 子命令。

### 3.3 新增 3 — `vet gcl trace` 也用本 schema

修改 `cmd/vet/gcl.go`：`vet gcl trace` 汇总时按本 schema 校验。

## 4. DoD

```
□ 1. 写入 incident-loop-agent/assets/trace.schema.json
□ 2. 写入 cmd/vet/internal/check/trace/trace.go
□ 3. cmd/vet/check.go 注册 "trace" 子命令
□ 4. cmd/vet 仍可 build + vet
□ 5. cmd/vet check trace --root . 对空目录给"OK"
□ 6. cmd/vet check trace --root . 对伪造文件（含 redaction_pass=false）报错
□ 7. cmd/vet check trace --root . 对伪造文件（缺 RequestId）报错
□ 8. go test ./... 包含 trace_test
□ 9. `incident-loop-agent/SKILL.md` 的 `## Operational Best Practices` 已确认 trace 路径写法与 schema 一致
□ 10. `.github/workflows/validate.yml` 已加入 `vet check trace --root .` 阶段（如尚未）
□ 11. `cmd/vet/check.go` 与 `cmd/vet/internal/check/trace/` 的 README 已同步新子命令
□ 12. `docs/gcl-spec.md` §9 anti-patterns 段落确认未引入新禁止行为；如引入则 §8 同步登记
□ 13. ledger 已登记（含 schema 必填字段摘要 + 反例 case 编号）
```

## 5. 验证命令

```bash
# 1. Go 工具链
cd cmd/vet
go build ./...
go vet ./...
go test ./...

# 2. 编译 vet 并跑 check trace
go build -o /tmp/vet .
/tmp/vet check trace --root .               # 空目录应 OK（无 trace）
echo "OK"

# 3. 构造假 trace 校验
mkdir -p /tmp/trace-test
echo '{"ticket_id":"X","started_at":"2026-07-13T00:00:00Z","finished_at":"2026-07-13T00:01:00Z","policy_decision":"AUTO","iterations":[],"redaction_pass":true}' \
  > /tmp/trace-test/trace.json
/tmp/vet check trace --root /tmp/trace-test    # ✅ 通过

# 缺 RequestId
echo '{"ticket_id":"X","started_at":"2026-07-13T00:00:00Z","finished_at":"2026-07-13T00:01:00Z","policy_decision":"AUTO","iterations":[{"ve_calls":[{"action":"ve ecs DescribeInstances","status":"ok"}]}],"redaction_pass":true}' \
  > /tmp/trace-test/trace.json
! /tmp/vet check trace --root /tmp/trace-test  && echo "REQID_GUARD_OK"

# 缺 redaction_pass
echo '{"ticket_id":"X","started_at":"2026-07-13T00:00:00Z","finished_at":"2026-07-13T00:01:00Z","policy_decision":"AUTO","iterations":[{"ve_calls":[{"request_id":"r1","action":"ve ecs DescribeInstances","status":"ok"}]}],"redaction_pass":false}' \
  > /tmp/trace-test/trace.json
! /tmp/vet check trace --root /tmp/trace-test  && echo "REDACTION_GUARD_OK"

# 清理
rm -rf /tmp/trace-test
```

## 6. 完成回报

```markdown
## T07 2026-07-XX — done
- 写入 trace.schema.json（必填 RequestId + policy_decision + redaction_pass）
- cmd/vet/internal/check/trace/ 实现完整校验
- check trace 子命令已注册
- T08 / 全局门禁可消费
```

## 7. 风险与回滚

| 风险 | 缓解 |
|------|------|
| schema 太严 → 旧 trace 不通过 | schema 加 `additionalProperties: false`；旧 trace 不在本仓库 |
| 校验器开销大 | 校验器跑在 PR 阶段，单文件 < 10ms；不影响 runtime |
| 回滚 | `git checkout cmd/vet/ incident-loop-agent/assets/trace.schema.json` |

## 8. 扩展：运行时轨迹采集 + 双 schema 校验（2026-07-17, commit 4831b5d）

原 T07 只覆盖 **incident-trace**（`check/trace` 包，schema 含 `ticket_id`）。
但 `vet gcl run` 实际产出的是 **gcl-trace-*.json**（runtime 形状，由 `gcl/trace.PersistTrace` 写入，含 `trace_schema_version`/`skill`/`operation_intent`，**无 `ticket_id`**），且此前**从不解析 `ve` 返回的 `{"Response":{"RequestId":"..."}}`**，导致 P5 的"每次 `ve` 调用都记录 RequestId"并未真正满足。

本次扩展（对齐 `l2-to-l3-plan.md` §6 P5 ✅）：

- `cmd/vet/internal/gcl/trace/trace.go`：`Iteration` 新增 `request_id` 字段；**新增 `gcl/trace.Check`**——只校验 runtime 轨迹（`trace_schema_version != ""`），要求 `redaction_pass == true` 且实际跑过 `ve` 调用的迭代 `request_id` 非空（`POLICY_BLOCK` 迭代豁免，因其未执行命令）。
- `cmd/vet/internal/gcl/run/run.go`：`Run()` 在 `runCommand` 后解析 `Response.RequestId` 写入 `Iteration.request_id`。
- `cmd/vet/check.go` `traceCheck`：去掉只扫 `incident-trace-` 的过滤，改为 `gcl-trace-*`（runtime，走 `gcl/trace.Check`）+ `incident-trace-*`（agent，走 `check/trace.Check`）双前缀扫描。
- `cmd/vet/internal/gcl/trace/trace_test.go`：新增 3 个 test（缺 request_id 失败 / 正常通过 / POLICY_BLOCK 豁免）。
- `validate.yml`：已含 `vet check trace` + `vet check policyguard`（P5/P7 e2e）。

> 关键边界：`check/trace`（incident 形状）与 `gcl/trace`（runtime 形状）是**两个独立包**，本次未改动 incident 校验逻辑，仅新增 runtime 校验器，二者正交。

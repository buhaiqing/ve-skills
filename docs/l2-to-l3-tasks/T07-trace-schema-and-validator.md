# T07 — Trace Schema + Validator

> 任务来源：[`../l2-to-l3-plan.md`](../l2-to-l3-plan.md) §4 (P5)
> 依赖：T06
> 预计工作量：1 天
> 状态：🟡 TODO

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

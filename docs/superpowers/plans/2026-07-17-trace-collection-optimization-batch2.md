# 执行轨迹采集优化 — 第二批次执行计划（Plan）

日期：2026-07-17
对应 Spec：docs/superpowers/specs/2026-07-17-trace-collection-optimization-batch2-design.md
范围：#3 时间窗口脆弱 / #4 request_id 提取 / #5 Window 虚假 / #6 incident 聚合

---

## 里程碑拆分

### M1 — #3 文件名时间戳过滤 + #5 Window 真实时间范围（同文件，耦合）

| # | 任务 | 文件 | 验证点 (DoD) |
|---|------|------|--------------|
| T1 | `CollectPaths` 用文件名时间戳过滤替代 ModTime；新增 `collectSince` helper 解析 `20060102-150405` | `internal/gcl/trace/trace.go:207` | 给定 sinceHours 按文件名时间过滤；全量不变 |
| T2 | `Summary.Window` 类型由 `map[string]int` 改为 `Window` struct；**`Aggregate` 签名加 `sinceHours *int` 参数**（R1 修正），内部算 cutoff 填 `since`/`until`；连带 `CmdAggregate`（cmd.go）签名同步加参并透传；`gcl.go` 已持有 `hours` 直接传入无需改 | `internal/gcl/trace/aggregate.go` `internal/gcl/trace/cmd.go` `cmd/vet/gcl.go` | sinceHours 给定→含 since/until；nil→since=nil |
| T3 | 单测：T1 文件名过滤（旧格式跳过+WARN、窗口内外）、T2 Window 两种形态；**同步更新第一批次 `aggregate_test.go` 对 `Aggregate` 的调用签名**（传 nil 保持全量语义） | `trace/trace_test.go` `trace/aggregate_test.go` | `go test` 全绿 |
| T4 | 回归：`vet gcl trace` 全量模式行为不变；`Window` JSON 字段名仍为 `window` | `gcl.go` | 全量聚合输出结构兼容 |

### M2 — #4 parseRequestID 鲁棒化

| # | 任务 | 文件 | 验证点 (DoD) |
|---|------|------|--------------|
| T5 | `parseRequestID` 先 `json.Unmarshal` 到 `struct{Response struct{RequestId string}}`，失败/空降级现有字符串扫描 | `internal/gcl/run/run.go:363` | 标准 JSON 返回 RequestId；非 JSON 降级 |
| T6 | 单测：标准 JSON / 嵌套 / 非 JSON / 截断 四态 | `run/run_test.go` 或新建 | 四态全过 |

### M3 — #6 incident 聚合器

| # | 任务 | 文件 | 验证点 (DoD) |
|---|------|------|--------------|
| T7 | 新增 `IncidentSummary`/`TicketStat`/`Totals`/`PolicyDecision` 类型 + `AggregateIncident(root)` + `PersistIncident` | `internal/gcl/trace/incident.go`（新增） | 纯函数可测 |
| T8 | 复用 M2 的 `incidentTraceLocal` 解析 incident 文件（同文件 link.go 已定义，直接复用） | `trace/incident.go` | 不重复解析逻辑 |
| T9 | `CmdIncident(root)` + gcl.go 接线 `vet gcl trace incident` | `trace/incident.go` `gcl.go` | 输出 `incident-summary-*.json` |
| T10 | 单测：per-ticket 统计（ve_calls/request_ids/policy_decision） | `trace/incident_test.go` | `go test` 全绿 |

### M4 — 收尾

| # | 任务 | 文件 | 验证点 (DoD) |
|---|------|------|--------------|
| T11 | `go build ./... && go vet ./...` 干净 | 全仓库 | exit 0 |
| T12 | `vet validate --root .` 保持绿（本批次只动 Go） | repo root | 无新增失败 |
| T13 | spec ↔ plan ↔ code 三方一致性自检 | — | review 报告 |

---

## 依赖图

```
M1 (T1→T2→T3→T4)
M2 (T5→T6)          ← 独立
M3 (T7→T8→T9→T10)   ← 复用 M2-link 的 incidentTraceLocal（已在 main）
M1 / M2 / M3 相互独立，可并行三 worktree 开发
M4 在三者合入后执行
```

注意 M3 的 `incidentTraceLocal` 是 M2 第一批次已合入 main 的类型，本批次在 main 上开发即可直接复用，无需重复定义。

---

## 风险与回滚

- **T2 Window 类型变更**：`Summary.Window` JSON 字段结构变化（map→struct），但字段名 `window` 不变，消费方按 key 读取仍可兼容；若严格依赖旧 map 形态需评估。风险低。
- **T1 旧格式文件**：无时间戳的旧 trace 在 sinceHours 模式下被跳过——符合"时间不可信则保守跳过"原则，有 WARN 提示。
- 每个 T 独立 commit；出问题 `git revert` 单 commit。

---

## 验证命令

```bash
cd cmd/vet && go build ./... && go vet ./...
cd cmd/vet && go test ./internal/gcl/...
vet validate --root .
```

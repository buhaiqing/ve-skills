# 执行轨迹采集优化 — 第二批次设计（Spec）

日期：2026-07-17
对应第一批次 Spec：docs/superpowers/specs/2026-07-17-trace-collection-optimization-design.md
范围：第一批次遗留的四点采集链路质量缺陷
- #3 `sinceHours` 时间窗口依赖 ModTime，脆弱
- #4 request_id 提取用字符串扫描，鲁棒性弱
- #5 `Summary.Window` 字段虚假（名为 window 实为总量）
- #6 incident trace 只有校验、没有聚合器

不在此批次：跨链关联（已做）、heal 并入（已做）。

---

## 1. 背景与问题

第一批次已完成跨链 link 与 heal 并入。本批次修复采集/聚合链路的四个质量缺陷：

- **P3（时间窗口脆弱）**：`trace.CollectPaths`（`internal/gcl/trace/trace.go:219-233`）用文件 `ModTime` 判断 `sinceHours` 窗口，但 `PersistTrace` 用 `time.Now().UTC()` 命名文件（trace.go:191）。一旦 trace 被 `git checkout`/`cp`/`rsync`，ModTime 改变 → 时间窗口失效、漏采或误采。
- **P4（提取脆弱）**：`parseRequestID`（`internal/gcl/run/run.go:363-383`）用 `strings.Index` 逐字符找 `"RequestId"`，对嵌套 JSON、多 RequestId、输出截断、字段顺序变化脆弱。
- **P5（Window 虚假）**：`Summary.Window` 只有 `{"trace_count": N}`（`aggregate.go:105`），名为 window 实为总量；当传入 `sinceHours` 时没有任何真实时间范围，报告不可解释。
- **P6（incident 无聚合）**：incident trace（`audit-results/incident-trace-*.json`）只有 `vet check trace` 的 `CheckDir` 校验（internal/check/trace/trace.go:87），没有按 ticket/skill 的聚合统计。第一批次 M2 的 `LinkIndex` 已能解析 incident 文件，但未做 per-ticket 聚合。

---

## 2. 设计选型

### #3 选型：用文件名时间戳过滤，不依赖 ModTime

`PersistTrace` 文件名含 `YYYYMMDD-HHMMSS`（`gcl-trace-20060102-150405.json`）。`CollectPaths` 解析文件名时间戳做窗口过滤，替代 ModTime。
- **Why**：文件名时间戳是写入时刻的权威记录，不受后续 cp/checkout 影响，且与 `sinceHours` 语义一致（"N 小时内的 trace"）。
- **Why 不选 mtime 修补**：mtime 不可信，补 mtime 等于在错误基础上打补丁。

### #4 选型：json.Unmarshal 提取，降级到扫描

`parseRequestID` 改为：先 `json.Unmarshal` 到 `struct{ Response struct{ RequestId string } }`，成功即返回；失败/无 `Response` 再降级到现有字符串扫描（保留兼容）。
- **Why**：`ve` 输出形如 `{"Response":{"RequestId":"..."}}`，结构稳定；json 解析对嵌套/多字段/顺序变化鲁棒。降级保证既有行为不回退。

### #5 选型：Window 填真实时间范围

当 `sinceHours != nil` 时，`Summary.Window` 填 `{ "since": <cutoff RFC3339>, "until": <now RFC3339>, "trace_count": N }`；`sinceHours == nil` 时填 `{ "since": null, "until": <now>, "trace_count": N }`（显式表达"全量"而非伪装窗口）。
- **Why**：让聚合报告可解释——消费方知道这份 summary 覆盖了哪个时间区间。

### #6 选型：扩展 M2 link 基础设施，新增 incident 聚合

新增 `trace.IncidentSummary` 聚合（复用 M2 的 `incidentTraceLocal` 解析）：按 `ticket_id` 统计 ve 调用数、request_id 覆盖数、policy_decision 分布；输出 `audit-results/incident-summary-YYYYMMDD-HHMMSS.json`，由新子命令 `vet gcl trace incident` 触发。
- **Why 选扩展 M2 而非独立聚合器**：incident 解析逻辑已在 `link.go` 落地，复用避免重复；且不与 runtime `CmdAggregate`、incident `CheckDir` 冲突（三者正交）。
- **Why 不并入 CmdAggregate**：runtime 聚合与 incident 聚合语义不同（runtime 按 skill、incident 按 ticket），混在一起会污染 `Summary` 结构。独立子命令更清晰。

---

## 3. 契约

### 3.1 `CollectPaths` 过滤（#3）

新增 `collectSince` helper：从文件名提取 `20060102-150405`，`time.Parse` 后比较 `UTC().After(cutoff)`。ModTime 分支删除。`input` 显式路径模式不受影响（仍按 glob）。

**旧格式判定（R4 澄清）**：仅当文件名匹配 `gcl-trace-*.json` 但时间戳部分无法 `time.Parse` 为 `20060102-150405` 时，视为"无时间信息"。`sinceHours` 给定 → 跳过该文件并 `fmt.Fprintf(stderr, "WARN: skip %s (no parseable timestamp)\n", p)`；全量模式 → 保留（不 WARN）。时间提取用正则/前缀切片定位 `gcl-trace-` 之后的 15 位数字+短横，解析失败即旧格式。

### 3.2 `parseRequestID` 返回（#4）

签名不变 `func parseRequestID(output string) string`。内部：`json.Unmarshal` → 取 `Response.RequestId`；为空 → 降级现有 `strings.Index` 扫描。

### 3.3 `Summary.Window`（#5）

```go
type Window struct {
    Since      *string `json:"since"`      // RFC3339 cutoff，nil = 全量
    Until      string  `json:"until"`      // RFC3339 now
    TraceCount int     `json:"trace_count"`
}
```
`Summary.Window` 类型由 `map[string]int` 改为 `Window`。

**签名变更（R1 修正）**：当前 `Aggregate(root string, traces []*Trace) *Summary` **不接收** `sinceHours`——`sinceHours` 只传到 `CollectPaths`。为填真实时间范围，需给 `Aggregate` 增加 `sinceHours *int` 参数：

```go
func Aggregate(root string, traces []*Trace, sinceHours *int) *Summary
```

内部计算：`if sinceHours != nil { cutoff := time.Now().UTC().Add(-time.Duration(*sinceHours) * time.Hour); s := cutoff.Format(time.RFC3339); win.Since = &s }`，`Until` 始终为 `time.Now().UTC().Format(time.RFC3339)`。

**连带改动**：`CmdAggregate`（cmd.go）签名同步加 `sinceHours *int` 并透传；`gcl.go` 已持有 `hours *int` 直接传入，无需改调用方；第一批次新建的 `aggregate_test.go` 调用 `Aggregate` 需同步更新签名（传 `nil` 保持全量语义）。

### 3.4 incident 聚合（#6）

```go
type IncidentSummary struct {
    TraceSchemaVersion string                `json:"trace_schema_version"` // "v1-incident"
    GeneratedAt        string                `json:"generated_at"`
    Totals             map[string]int        `json:"totals"` // tickets, ve_calls, request_ids_covered
    ByTicket           map[string]TicketStat `json:"by_ticket"`
    PolicyDecision     map[string]int        `json:"policy_decision"` // AUTO/ASK/REFUSE 计数
}
type TicketStat struct {
    VeCalls        int    `json:"ve_calls"`
    RequestIDs     int    `json:"request_ids"` // 该 ticket 下非空 request_id 数（incident 内部，非跨链）
    PolicyDecision string `json:"policy_decision"`
}
```
> 注：`request_ids_covered` / `RequestIDs` 指 incident trace **内部**非空 request_id 计数，仅反映采集完整度，不做跨链匹配（跨链关联是 M2 link 的职责，避免语义重叠）。

`CmdIncident(root string) int`：扫描 `incident-trace-*.json` → 聚合 → `PersistIncident(root, *IncidentSummary)` → 打印 `{ "incident_path": ..., "tickets": N }`。

---

## 4. 异常边界

| 场景 | 处理 |
|------|------|
| 文件名无时间戳（旧格式） | #3 过滤时该文件视为"无时间信息"→ 若 `sinceHours` 给定则跳过并 `WARN`；全量模式保留（不 WARN） |
| `ve` 输出非 JSON | #4 降级字符串扫描，行为同现状 |
| 多 RequestId（罕见） | #4 取顶层 `Response.RequestId`（单值），降级扫描取首个 |
| incident 文件损坏 | #6 跳过并 WARN，不中断聚合 |
| `audit-results/` 不存在 | #3/#6 均返回空结果 exit 0 |

---

## 5. 验收标准

1. `CollectPaths` 在 `sinceHours` 给定时按文件名时间戳过滤，不读 ModTime；全量模式行为不变。
2. `parseRequestID` 对标准 `{"Response":{"RequestId":"x"}}` 返回 x；对非 JSON/截断降级扫描仍工作。
3. `Summary.Window` 在 `sinceHours` 给定时含真实 `since`/`until`；全量时 `since=nil`。
4. `vet gcl trace incident` 输出 `incident-summary-*.json`，`ByTicket` 正确按 ticket 统计 ve_calls/request_ids。
5. 三个改动均有 `_test.go`：#3 文件名过滤单测、#4 双路径单测、#6 incident 聚合单测。
6. `go build ./... && go vet ./...` 干净；既有的 runtime `CmdAggregate`、`CheckDir`（两个 checker）行为不变。

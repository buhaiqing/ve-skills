# 执行轨迹采集优化 — 第一批次设计（Spec）

日期：2026-07-17
范围：轨迹采集「生成 → 写入 → 聚合」链路的两个割裂点
- #1 request_id 跨链关联（runtime GCL trace ↔ incident trace）
- #2 heal telemetry 并入 audit-results 并纳入 GCL 聚合

不在此批次：ModTime 时间窗口脆弱性、parseRequestID 字符串扫描、Window 字段虚假、incident 无聚合器（第二批次）。

---

## 1. 背景与问题

仓库现有三套轨迹数据，采集链路三段割裂：

| 链路 | 文件 | 写入方 | 聚合/校验 |
|------|------|--------|-----------|
| Runtime GCL trace | `audit-results/gcl-trace-*.json` | `vet gcl run` | `vet gcl trace` 聚合 |
| Incident trace | `audit-results/incident-trace-*.json` | `incident-loop-agent` | `vet check trace` 仅校验 |
| Heal telemetry | `/tmp/ve-self-healing.log` | `vet gcl run` (heal 包) | `vet gcl heal-stats` 聚合 |

两个核心问题：

- **P1（跨链割裂）**：runtime trace 的 `Iteration.RequestID` 与 incident trace 的 `VeCall.RequestID` 是同一朵云的同一调用，却分属两个 schema、两个文件、两个正交 checker（`internal/gcl/trace/trace.go:153` / `internal/check/trace/trace.go:39`），无法用 request_id 拼出「incident 触发 → GCL 执行 → 自愈」完整链路。
- **P2（heal 脱离质量报告）**：heal telemetry 落在 `/tmp/ve-self-healing.log`（`heal/log.go:14`），不随仓库、不进 `audit-results`、被 `aggregate.go` 完全忽略。自愈是 L4 核心能力，却在 GCL 质量报告中不可见（SLO：成功率/平均时长/人工干预率/回退率）。

---

## 2. 设计选型

### #1 选型：被动索引（A 方案，推荐）

不改动两套 trace 的写入逻辑与 schema，新增聚合层命令 `vet gcl trace link`：
- 扫描 `audit-results/` 下所有 `gcl-trace-*.json` 与 `incident-trace-*.json`
- 以 `request_id` 为键建立交叉引用索引，输出 `audit-results/trace-link-YYYYMMDD-HHMMSS.json`

**Why 选 A 不选 B（主动回填 TicketID）**：
- B 需改 `Options`、trace schema、incident-loop-agent 调用约定；incident 进程与 GCL 执行时序难保证（incident 可能先于/并行于 GCL），回填易空。
- A 零侵入既有 trace 写入、不改 P5 不变量、不改 checker 正交性，直接解决「request_id 拼完整链路」诉求。

### #2 选型：heal telemetry 改写到 audit-results + 纳入聚合

- `heal.DefaultLogPath` 从 `/tmp/ve-self-healing.log` 改为 `audit-results/ve-self-healing.log`（与既有 trace 同目录、同 gitignore 策略）。
- 在 `trace.Aggregate` 中新增可选 heal 维度：解析 heal log，把 `SuccessRate / AvgDurationMs / UserInterventionRate / FallbackRate` 计入 `Summary`（仅当 heal 文件存在时填充，否则留 nil——保持向后兼容）。
- 保留 `vet gcl heal-stats` 不变（仍直接读 heal log），避免破坏既有消费方。

---

## 3. 状态机 / 契约

### 3.1 `vet gcl trace link` 输出契约（`trace-link-*.json`）

```json
{
  "trace_schema_version": "v1-link",
  "generated_at": "<RFC3339 UTC>",
  "request_index": {
    "<request_id>": {
      "gcl_traces": ["audit-results/gcl-trace-xxx.json"],
      "incident_traces": ["audit-results/incident-trace-yyy.json"]
    }
  },
  "unlinked": {
    "gcl_only": ["<request_id>", ...],
    "incident_only": ["<request_id>", ...]
  },
  "counts": { "gcl": N, "incident": M, "linked": K }
}
```

- 仅当 `request_id` 非空时参与索引（与 P5 不变量一致：运行时 ve 调用必有 request_id）。
- `gcl_only` / `incident_only` 列出无法跨链匹配的 request_id，用于发现采集盲区。

### 3.2 heal 维度并入 `Summary`

```go
type Summary struct {
    // ... 既有字段不变 ...
    Heal *HealSummary `json:"heal,omitempty"` // 新增，nil 表示无 heal 数据
}

type HealSummary struct {
    SuccessRate         float64 `json:"success_rate"`
    AvgDurationMs       float64 `json:"avg_duration_ms"`
    UserInterventionRate float64 `json:"user_intervention_rate"`
    FallbackRate        float64 `json:"fallback_rate"`
    TotalEvents         int64   `json:"total_events"`
}
```

---

## 4. 异常边界

| 场景 | 处理 |
|------|------|
| `audit-results/` 不存在 | `trace link` 与 `Aggregate` 均返回空结果，exit 0（非错误） |
| heal log 不存在 / 为空 | `Summary.Heal` 为 nil，不影响既有统计 |
| 同一 request_id 出现在多个 gcl trace | 全部列入 `gcl_traces[]`（一对多允许，同一云调用可能多次 GCL 执行） |
| trace 文件解析失败 | 跳过并 WARN，不中断索引构建 |
| `/tmp` 老 heal log 残留 | 不改写旧路径；首次以新路径写入。旧数据可通过 `--log` 显式指定 |

---

## 5. 验收标准

1. `vet gcl trace link` 能扫描 `audit-results/`，以 request_id 关联 gcl/incident trace，输出 `trace-link-*.json`；`counts.linked` 正确。
2. 改 `DefaultLogPath` 后，`vet gcl run --heal=smart` 的 heal 事件写入 `audit-results/ve-self-healing.log`，且 `vet gcl heal-stats` 仍可读。
3. `vet gcl trace` 聚合时，若 `audit-results/ve-self-healing.log` 存在，`Summary.Heal` 被正确填充；不存在时 `Summary.Heal` 为 nil，既有字段不受影响。
4. 既有 `vet check trace`（incident schema）与 `trace.Check`（runtime schema）正交性不变——本批次不修改两个 checker 的判定逻辑。
5. 三处改动均有 `_test.go` 覆盖：link 索引正确性、heal 路径常量、Aggregate 的 heal 维度。
6. `go build ./... && go vet ./...` 干净。

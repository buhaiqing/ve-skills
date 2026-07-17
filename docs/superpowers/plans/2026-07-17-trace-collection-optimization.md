# 执行轨迹采集优化 — 第一批次执行计划（Plan）

日期：2026-07-17
对应 Spec：docs/superpowers/specs/2026-07-17-trace-collection-optimization-design.md
范围：#1 request_id 跨链关联 + #2 heal telemetry 并入聚合

---

## 里程碑拆分

### M1 — #2 Heal telemetry 并入 audit-results（低风险，先做）

| # | 任务 | 文件 | 验证点 (DoD) |
|---|------|------|--------------|
| T1 | 改 `heal.DefaultLogPath` 常量为 `audit-results/ve-self-healing.log` | `cmd/vet/internal/gcl/heal/log.go:14` | 常量值更新；`go build` 通过 |
| T2 | 确认 `Persist`（`metrics.go:122`）的 append 逻辑对新路径无副作用（目录需 MkdirAll） | `heal/log.go` / `metrics.go` | `Persist` 写 `audit-results/` 成功（补/修测试） |
| T3 | `vet gcl heal-stats` 默认读新路径仍工作 | `heal/log.go` `LoadEvents` | `heal-stats` 读新路径回归通过 |
| T4 | `Summary` 增加 `Heal *HealSummary` 字段；`Aggregate` 解析 heal log 填充 | `trace/aggregate.go` | heal 存在→填充；不存在→nil |
| T5 | 单测：T1/T2/T4 覆盖 | `heal/log_test.go` `trace/aggregate_test.go` | `go test ./internal/gcl/...` 全绿 |

### M2 — #1 request_id 跨链关联（新增 link 聚合器）

| # | 任务 | 文件 | 验证点 (DoD) |
|---|------|------|--------------|
| T6 | 实现 `LinkIndex`（扫描 gcl-trace / incident-trace，按 request_id 建索引） | `trace/link.go`（新增） | 函数纯函数、可测 |
| T7 | request_id 提取：从 runtime `Iteration.RequestID` 与 incident `VeCall.RequestID` 收集 | `trace/link.go` | 仅非空参与索引 |
| T8 | 实现 `vet gcl trace link` CLI 命令，输出 `trace-link-*.json` | `trace/link.go` + `cmd` 接线 | 输出契约符合 spec §3.1 |
| T9 | 单测：link 索引正确性（一对多、unlinked、解析失败跳过） | `trace/link_test.go` | `go test` 全绿 |

### M3 — 收尾

| # | 任务 | 文件 | 验证点 (DoD) |
|---|------|------|--------------|
| T10 | `go build ./... && go vet ./...` 干净 | 全仓库 | exit 0 |
| T11 | `vet validate --root .` 保持绿（本批次只动 Go，不影响 skill markdown，但需确认无副作用） | repo root | 无新增失败 |
| T12 | spec ↔ plan ↔ code 三方一致性自检（见 AGENTS.md 铁律） | — | review 报告 |

---

## 依赖图

```
M1 (T1→T2→T3, T4→T5)  ── 并行可行，但 T4 依赖 T1 路径生效
M2 (T6→T7→T8→T9)
M1 与 M2 相互独立，可并行开发（分两个 worktree/分支）
M3 在 M1+M2 合入后执行
```

---

## 风险与回滚

- **T1 路径变更**：`/tmp` → `audit-results/`。若已有消费方硬编码 `/tmp/ve-self-healing.log`，回退只需还原常量。已确认 `heal-stats` 经 `DefaultLogPath` 读取，无硬编码。
- **M2 纯新增**：不改既有 checker 与写入逻辑，LinkIndex 为只读扫描，零风险。
- 回滚：每个 T 独立 commit；出问题 `git revert` 单 commit 即可。

---

## 验证命令

```bash
cd cmd/vet && go build ./... && go vet ./...
cd cmd/vet && go test ./internal/gcl/...
vet validate --root .
```

# Reflexion Write-back 链路补全执行计划

日期：2026-07-17
对应 Spec：`docs/superpowers/specs/2026-07-17-reflexion-writeback-completion-design.md`

---

## 里程碑拆分

### M1 — Memory Store Go 包

| # | 任务 | 文件 | 验证点 (DoD) |
|---|------|------|--------------|
| T1 | 创建 `cmd/vet/internal/memory/store.go`：`FailurePatternEntry`、`AppendFailurePattern`、`LoadFailurePatterns`、`GetPatternsBySkill` | `memory/store.go` | 纯函数，路径 `.runtime/memory/` |
| T2 | 创建 `memory/store_test.go`：新增、count++、按 skill 过滤、空文件、并发安全 | `memory/store_test.go` | 5+ 测试全绿 |

### M2 — Write-back 链路接线

| # | 任务 | 文件 | 验证点 (DoD) |
|---|------|------|--------------|
| T3 | 修改 `run.go:writebackFailurePattern`：追加调用 `memory.AppendFailurePattern` | `run/run.go` | JSON 写入 + markdown 写入双路径 |
| T4 | 修改 `run.go:loadKnownFailurePatterns`：改为从 `memory.LoadFailurePatterns` 加载 | `run/run.go` | 结构化查询替代逐行扫描 |
| T5 | 新增：count ≥ 10 时自动调用 `transpile.TranspileFile` 生成 guardrails.yaml | `run/run.go` | 达到阈值自动触发 |

### M3 — 收尾

| # | 任务 | 文件 | 验证点 (DoD) |
|---|------|------|--------------|
| T6 | `go build ./... && go vet ./... && go test ./...` | 全仓库 | exit 0 |
| T7 | 端到端验证：GCL 执行 → JSON 写入 → count++ → guardrails 生成 | — | 链路完整 |
| T8 | 更新 `docs/reflexion-memory.md`：补充 write-back 链路描述 | reflexion-memory.md | 文档同步 |

---

## 依赖图

```
M1 (T1 → T2) → M2 (T3 → T4 → T5) → M3 (T6 → T7 → T8)
```

---

## 新增/修改文件

**新增**：
- `cmd/vet/internal/memory/store.go`
- `cmd/vet/internal/memory/store_test.go`

**修改**：
- `cmd/vet/internal/gcl/run/run.go`（writebackFailurePattern + loadKnownFailurePatterns）
- `docs/reflexion-memory.md`（文档同步）

---

## 验证命令

```bash
cd cmd/vet && go build ./... && go vet ./... && go test ./...
go test ./internal/memory/ -v
```

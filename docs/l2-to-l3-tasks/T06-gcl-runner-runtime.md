# T06 — GCL Runner Runtime（生产化）

> 任务来源：[`../l2-to-l3-plan.md`](../l2-to-l3-plan.md) §4 (P4)
> 依赖：T05
> 预计工作量：1–1.5 天（增强既有运行时，非从零构建 — 见 §2）
> 状态：🟡 TODO

## 1. 目标

把 `incident-loop-agent` 的 v0.1.0 skeleton 提升为**生产 runtime**：
用 `vet gcl run`（Go 工具）替换废弃的 `scripts/gcl_runner.py`，
并落实 `max_iter` 强制、retry/backoff、partial-rollback 检测，
以及核心 9 格 `scoreDecision` 评分器（消费 T02 schema + T03 allowlist + T04 元数据）。

> **TDD 要求**：本卡既有/新增函数（含 `scoreDecision`）须先写失败测试再实现，单测见 §5。

> **背景**：AGENTS.md 强制"All Tools MUST Be Go" —
> 当前 `incident-loop-agent` SKILL.md 仍引用 `scripts/gcl_runner.py`（已废弃），
> 必须迁移到 `vet gcl run`（在 `cmd/vet/internal/gcl/run/` 下）。

## 2. 背景

- 现状：`incident-loop-agent/SKILL.md:146` 引用 `scripts/gcl_runner.py`
- 目标：`vet gcl run --skill incident-loop-agent --plan <dispatch_plan>`
- Go 工具入口：`cmd/vet/gcl.go` + `cmd/vet/internal/gcl/run/`
- `max_iter` 策略：`SKILL.md:148` — "max_iter=3 for repair, max_iter=2 for destructive"
- **既有运行时（2026-07-13 评估确认已存在）**：`cmd/vet/internal/gcl/run/run.go`（含 `run_test.go`）、`trace/`、`critic/`、`secret/`、`gate/` 包及真实 trace 样本均已落地；`vet gcl run` 入口 `cmd/vet/gcl.go` 已存在。本卡是**增强既有运行时 + 清除 Python 引用**，不是从 skeleton 起建，故工作量下调。

## 3. 产出物

### 3.1 修改 1 — 删废弃引用

**`incident-loop-agent/SKILL.md`**
- `:146` 改为引用 `vet gcl run`
- 同时更新 `compatibility:` frontmatter（`:18-19`）把 `scripts/gcl_runner.py` 改成 `vet gcl run` 工具链
- 同步 `## References` 部分

### 3.2 新增 1 — max_iter enforcement（Go 代码）

**修改**：`cmd/vet/internal/gcl/run/run.go`（既有文件，非新建）

| 函数 | 行为 |
|------|------|
| `scoreDecision(op)` | **核心 9 格评分器（此前无卡片显式归属）**：读 op 的 `(safety_class, blast_radius, confidence, safety)` + T03 allowlist 成员资格 → 查 T02 schema / T01 §3.2 矩阵 → 输出 `AUTO/ASK/REFUSE`；缺 metadata → ASK（fail-safe）；safety=0 → REFUSE（覆盖一切） |
| `enforceMaxIter(plan, opSafety)` | 读 plan，destructive op → max_iter=2，其余 max_iter=3；超限 → 报错 |
| `withBackoff(call, maxRetry=2)` | retry 2 次，指数退避 200ms/1s |
| `detectPartialRollback(trace)` | 对比 pre/post snapshot，标记 diff 项 |
| `validatePolicyDecision(plan, decision)` | 复核 `scoreDecision` 输出：safety=0 → REFUSE 必生效 |

### 3.3 新增 2 — incident-loop-agent 专用 entry

**扩展既有 `vet gcl run` 子命令**（入口已在 `cmd/vet/gcl.go`，无需新建子命令）

```go
// RunLoop 是 incident-loop-agent 的入口
// 入参：dispatch_plan (JSON)
// 出参：trace + decision (AUTO/ASK/REFUSE per op)
// 接口契约：RunLoop 必须对每个 op 调用 scoreDecision 并把结果写入
//           {{policy.decision}}（T05 在 SKILL.md Step 5a 消费此变量）；
//           allowlist 外（非 T03 列出的 skill/symptom）一律不授予 AUTO。
func RunLoop(planPath string) (*Trace, error)
```

## 4. DoD

```
□ 1. SKILL.md 全文无 "scripts/gcl_runner.py" 引用
□ 2. SKILL.md 引用 vet gcl run 至少 1 处
□ 3. compatibility frontmatter 已更新
□ 4. cmd/vet/internal/gcl/run/run.go 新增/增强 5 个函数（含核心 `scoreDecision`）
□ 5. `scoreDecision` 单测覆盖 9 格矩阵：read-only→AUTO；single+high state-changing→AUTO；destructive→ASK（never AUTO）；multi/account→ASK；safety=0→REFUSE（覆盖一切）；缺 metadata→ASK（fail-safe）
□ 6. cmd/vet 仍可 build + vet
□ 7. `vet gcl run --help` 已含 incident-loop 入口（既有，非新建）
□ 8. vet gcl run --max-iter 0 ... 显式报错（不静默忽略）
□ 9. 单元测试覆盖：safety=0 → REFUSE；destructive → max_iter=2；非 destructive → max_iter=3
```

## 5. 验证命令

```bash
# 1. 废弃引用全清
! grep -r "scripts/gcl_runner.py" incident-loop-agent/ \
  && echo "PYTHON_REF_REMOVED"

# 2. 新引用存在
grep -q "vet gcl run" incident-loop-agent/SKILL.md && echo "NEW_REF_OK"

# 3. Go 工具链
cd cmd/vet
go build ./...
go vet ./...
go test ./...                                 # 全部 ok
go build -o /tmp/vet .
/tmp/vet gcl run --help

# 4. 关键单测（TDD：先写失败测试再实现）
go test -run TestScoreDecision ./internal/gcl/run/ -v
go test -run TestMaxIter ./internal/gcl/run/ -v
go test -run TestSafetyZeroRefuse ./internal/gcl/run/ -v
go test -run TestPartialRollback ./internal/gcl/run/ -v

# 5. CI 模拟
make build && make vet && make test
```

## 6. 完成回报

```markdown
## T06 2026-07-XX — done
- scripts/gcl_runner.py 引用全部清除
- cmd/vet/internal/gcl/run/run.go 新增/增强 5 个函数（含 scoreDecision）
- go build / vet / test 全部绿
- T07 / T08 可消费（runtime 已是生产形态）
```

## 7. 风险与回滚

| 风险 | 缓解 |
|------|------|
| `vet gcl run` 行为与 `scripts/gcl_runner.py` 不一致 | 单测覆盖关键路径；保留 README 写明差异 |
| max_iter enforcement 误伤合法修复 | 默认 max_iter=3，仅 destructive 限到 2 |
| 回滚 | `git checkout cmd/vet/ incident-loop-agent/SKILL.md` |

# T10 — L3 多路径自愈（auto-select best）

> 任务来源：[`../autonomous-ops-roadmap.md`](../autonomous-ops-roadmap.md) §M2 (M2-2)
> 依赖：T09
> 预计工作量：1.5 天
> 状态：🟡 TODO

## 1. 目标

为每个高频错误类别提供 **≥ 2 条** 互不重叠的自愈路径，
自愈引擎按"成本 + 历史成功率"自动选最优。

## 2. 背景

- L2 智能重试已就位（T09）— 本卡把"重试 1 条路"扩到"多条候选路"
- framework §3 自愈决策树（`:36-57`）已定义 6 步流程
- framework 目标（`:31`）："L3 多路径自愈：多种自愈路径，自动选择最优方案"

## 3. 产出物

### 3.1 路径注册表

**新文件**：`cmd/vet/internal/gcl/heal/paths.go`

```go
package heal

type Path struct {
    Name     string
    Class    ErrorClass   // 适用类别
    Cost     int          // 0=最低（如换镜像）, 5=高（如重启）
    Execute  func(ctx context.Context) error
}

var Paths = []Path{
    // NET_TIMEOUT 至少 2 条
    {Name: "switch-mirror",    Class: ClassRetryable, Cost: 1, Execute: ...},
    {Name: "increase-timeout", Class: ClassRetryable, Cost: 2, Execute: ...},
    // PERM_* 至少 2 条
    {Name: "use-user-bin",     Class: ClassFatal,     Cost: 1, Execute: ...},
    {Name: "prompt-sudo",      Class: ClassFatal,     Cost: 4, Execute: ...},
    // ...
}

// SelectBest 按 (Cost, 历史成功率) 选最优
func SelectBest(class ErrorClass) *Path
```

### 3.2 自愈 runner

**新文件**：`cmd/vet/internal/gcl/heal/runner.go`

- 输入：error_code + 当前上下文
- 流程：Classify → SelectBest → Execute → 验证（读 trace 状态）
- 输出：trace 中 `self_healing` 段（含 path_name / cost / result / duration）

### 3.3 至少 2 条路径的错误类别（framework §2 + Go runtime §2.2）

| 错误类别 | 路径 1 | 路径 2 |
|---------|--------|--------|
| NET_TIMEOUT | switch-mirror | increase-timeout |
| GO_DOWNLOAD_FAIL | try-cn-mirror | use-fallback-version |
| PERM_WRITE_FAIL | use-user-bin | prompt-sudo |
| RES_DISK_FULL | clean-tmp | suggest-cleanup |
| GO_VERSION_INCOMPATIBLE | try-lts | try-stable |

## 4. DoD

```
□ 1. 写入 cmd/vet/internal/gcl/heal/paths.go（≥ 5 类错误 × 2+ 路径）
□ 2. 写入 cmd/vet/internal/gcl/heal/runner.go（Classify→Select→Execute→Verify）
□ 3. 路径选择按 (Cost, 历史成功率) 加权；历史数据来自 T11 指标
□ 4. go build + go vet + go test 绿
□ 5. runner_test.go 覆盖：每个 class 选到预期最优路径
□ 6. runner_test.go 覆盖：所有路径失败时 → 下一路径 / 上报
□ 7. trace 中 self_healing 段必填（与 T07 trace schema 兼容）
```

## 5. 验证命令

```bash
cd cmd/vet
go build ./...
go vet ./...
go test -run TestSelectBest ./internal/gcl/heal/ -v
go test -run TestRunnerExhaustion ./internal/gcl/heal/ -v   # 路径全失败场景
go test ./...
```

## 6. 完成回报

```markdown
## T10 2026-07-XX — done
- paths.go：5 类错误 × 2+ 路径
- runner.go：Classify→Select→Execute→Verify 闭环
- trace.self_healing 段必填
- T11 / T16 可消费
```

## 7. 风险与回滚

| 风险 | 缓解 |
|------|------|
| 路径副作用（如 prompt-sudo 卡住） | 每条路径设 timeout（默认 30s）；超时 → 下一条 |
| SelectBest 选择退化 | 历史数据为空时退到 Cost 最低；T11 指标累积后切换 |
| 回滚 | `git checkout cmd/vet/internal/gcl/heal/` |

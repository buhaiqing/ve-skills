# T09 — L2 智能重试（错误分类驱动）

> 任务来源：[`../autonomous-ops-roadmap.md`](../autonomous-ops-roadmap.md) §M2 (M2-1)
> 依赖：T06 (vet gcl run — L2→L3)
> 可并行：T12
> 预计工作量：1 天
> 状态：🟡 TODO

## 1. 目标

把"固定次数重试"（L1）改为"基于错误分类的针对性重试"（L2）：
- 网络类错误 → 退避重试 + 切换镜像/endpoint
- 限流类 → 等待令牌 + 重试
- 权限/参数类 → **不重试**，直接上报

## 2. 背景

- 当前状态：L1 基础重试（`enhanced-self-healing-framework.md:29`）
- 错误码体系：框架已定义 ≥ 10 码（`:65-97`）
- 调用入口：`vet gcl run` 内 retry 路径（T06 已建立）

## 3. 产出物

### 3.1 新增 Go 包

**新文件**：`cmd/vet/internal/gcl/heal/retry.go`

```go
package heal

type ErrorClass int
const (
    ClassRetryable   ErrorClass = iota // NET_TIMEOUT, NET_DNS_FAIL, GO_DOWNLOAD_TIMEOUT
    ClassRateLimit                     // Throttling, RequestLimitExceeded
    ClassFatal                         // PERM_*, RES_*, CONF_*, parameter errors
    ClassUnknown
)

type RetryPolicy struct {
    MaxAttempts int
    BackoffBase time.Duration
    BackoffMax  time.Duration
    Jitter      bool
}

// Classify 根据 GCL trace 中的 error_code 给出分类
func Classify(errorCode string) ErrorClass

// SmartRetry 替代固定 retry：按分类决定是否重试 / 退避 / 切换
func SmartRetry(ctx context.Context, op func() error, policy RetryPolicy) error
```

### 3.2 错误码映射表（来源：framework §2）

| 错误码（节选） | 分类 | 重试策略 |
|---------------|------|---------|
| `NET_TIMEOUT`, `NET_DNS_FAIL`, `NET_CONNECTION_REFUSED` | Retryable | 指数退避 200ms→1.6s，max=3 |
| `DEP_NET_TIMEOUT` | Retryable | 同上 |
| `GO_DOWNLOAD_TIMEOUT`, `GO_DOWNLOAD_FAIL` | Retryable | 同上 + 切换镜像 |
| `RequestLimitExceeded`, `Throttling` | RateLimit | 等待 Retry-After + 1 次重试 |
| `PERM_*`, `RES_VERSION_INCOMPATIBLE`, `InvalidParameter` | Fatal | **不重试** |
| `UNKNOWN_ERROR` | Unknown | 重试 1 次 + 标记 |

## 4. DoD

```
□ 1. 写入 cmd/vet/internal/gcl/heal/retry.go（含 Classify + SmartRetry）
□ 2. 错误码映射覆盖 framework §2 中至少 10 个具体码
□ 3. go build + go vet + go test 全部绿
□ 4. retry_test.go 覆盖：Retryable 重试到 max；Fatal 立即放弃；RateLimit 等待 Retry-After
□ 5. retry_test.go 覆盖：所有重试记录到指标（T11 消费）
□ 6. vet gcl run --heal=smart 与 --heal=none 行为可对比（CI 中加 1 case）
```

## 5. 验证命令

```bash
cd cmd/vet
go build ./...
go vet ./...
go test -run TestSmartRetry ./internal/gcl/heal/ -v
go test -run TestClassify ./internal/gcl/heal/ -v
go test ./...   # 全包不退化
```

## 6. 完成回报

```markdown
## T09 2026-07-XX — done
- 写入 cmd/vet/internal/gcl/heal/retry.go
- Classify 覆盖 10+ framework 错误码
- 单测：Retryable/Fatal/RateLimit 三类行为正确
- T10 / T11 可消费
```

## 7. 风险与回滚

| 风险 | 缓解 |
|------|------|
| 错误码新增未同步 | Classify 维护"未知 → Unknown"分支 + 标记；月度 audit |
| 重试导致下游压力 | BackoffMax + 限速；与 framework §6.1 指标挂钩 |
| 回滚 | `git checkout cmd/vet/internal/gcl/heal/retry.go` |

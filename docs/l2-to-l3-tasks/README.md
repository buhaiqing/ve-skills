# L2 → L3 任务卡片目录（可交给 AI 迭代开发）

> **目的**：把 [`docs/l2-to-l3-plan.md`](../l2-to-l3-plan.md) 里的 P1–P7 拆成**单次可交付**的 Task 卡片。每张卡片：
> - 单一可交付物（一个文件 / 一个改动）
> - 自含验证命令（按 AGENTS.md "All Tools MUST Be Go" 规则，可编译/可 vet/可测试）
> - 自含 DoD（Definition of Done）
> - 显式依赖（前序 Task 编号）
> - 链接回 L2→L3 plan 的对应 §4 任务行
>
> **使用方式**：
> 1. 按 [01-index.md](./01-index.md) 的依赖图领走一张卡片
> 2. 单独打开对应 `Txx-<name>.md`
> 3. 严格按"产出 + DoD + 验证"完成
> 4. 在 [`_trace/ledger.md`](./_trace/ledger.md) 登记进度
>
> **最后更新**：2026-07-13

## 开发规范（强制）

- **TDD 铁律**：所有 Go 实现卡片（T06 / T07 / T08）严格执行测试驱动开发 —— 先写**失败测试**，再写最小实现，再重构（Red→Green→Refactor）。单测是 DoD 的硬证据，不允许"先实现后补测试"。单测命令见各卡片 §5。
- **GCL 规范**：`vet gcl run` 相关改动须遵循 [`../gcl-spec.md`](../gcl-spec.md)：Generator/Critic 隔离上下文、Safety=0 必 ABORT、trace 持久化且凭证脱敏。
- **All Tools MUST Be Go**：新增/修改工具一律 Go 实现，`go build` + `go vet` + `go test` 全绿方可登记完成。
- **凭证安全**：trace / 报告中的凭证一律 `<masked>`，遵循 `cmd/vet/internal/check/assessment` 脱敏路径。
- **进度登记**：完成一条才在 [`_trace/ledger.md`](./_trace/ledger.md) 追加一条，禁止预填。

# Agent Runtime Patterns

> 从 Phase 1 Agent Runtime 开发 + Code Review 中提炼的高阶模式，适用于所有 Go Agent 开发。
> AGENTS.md 中只保留规则摘要，完整实现见本文档。

---

## P1: Shell 安全 — 禁止 `sh -c`

```go
// ❌ P0 违规：shell wrapping 残留
cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)

// ✅ 正确：直接传参，无 shell 解析
args := []string{"--Region", "cn-beijing", "--InstanceId", id}
cmd := exec.CommandContext(ctx, "ve", args...)
```

若必须拼接命令字符串（如外部脚本），**必须**用白名单正则过滤所有参数，且注释说明为什么不能直接传参。

---

## P2: Checkpoint/Resume 模式

有状态引擎**必须**支持断点续跑：

```go
func runLoop(root, payload, runID, dryRun) *RunResult {
    state := &RunState{CurrentStep: StepIngest, Payload: *payload}
    existing, _ := LoadState(root, runID)
    if existing != nil { state = existing }  // 恢复 checkpoint

    if state.CurrentStep <= StepTriage { ... state.CurrentStep = StepDiagnose }
    if state.CurrentStep <= StepDiagnose { ... state.CurrentStep = StepPropose }
    // ... 每步用 <= 判断跳过已完成步骤
}
```

---

## P3: DRY — Dry-Run 共用引擎

Dry-run 和正常执行**必须**走同一条引擎路径，用 flag 区分行为：

```go
func Run(root, payload, runID) *RunResult   { return runLoop(root, payload, runID, false) }
func RunDry(root, payload, runID) *RunResult { return runLoop(root, payload, runID, true) }
```

禁止在 CLI 层重复实现步骤序列。

---

## P4: 提取的数据必须使用

若从输入提取了数据（如 `ResourceIDs`），**必须**在下游使用（如 `BuildDiagnoseArgs` 加 `--InstanceId` 过滤）。提取后不用 = 设计缺陷。

---

## P5: 委托执行必须带超时

调用外部 runner（GCL / shell / HTTP）**必须**传 timeout，防止无限阻塞：

```go
result := run.Run(run.Options{
    Root: root, Skill: op.Skill, Command: op.Command,
    Timeout: 300,  // 秒
})
```

---

## P6: RunID 唯一性

```go
// ❌ 截断到 32 位，高并发碰撞
runID := fmt.Sprintf("%08x", time.Now().UnixNano()%0x100000000)

// ✅ 完整纳秒时间戳，无碰撞
runID := fmt.Sprintf("%d", time.Now().UnixNano())
```

---

## P7: 配置化硬编码值

Region、超时、重试次数等硬编码值**必须**通过参数/字段/环境变量暴露，提供合理默认值。

---

## P8: FlagSet 解析子命令模式

`flag.FlagSet.Parse` 在遇到第一个非 flag 参数时停止解析。对于 `cmd subcmd --flag val` 形式的 CLI，必须先提取子命令，再解析剩余参数：

```go
// ❌ 错误：flag.Parse 在 "test" 处停止，--envelope 永远不会被解析
flag.StringVar(&envelope, "envelope", "", "...")
flag.Parse()  // args = ["test", "--envelope", "x.yaml"]

// ✅ 正确：先提取子命令，再解析剩余 args
func runAutonomy(args []string) {
    if len(args) == 0 { usage(); return }
    subcmd := args[0]
    remaining := args[1:]
    fs := flag.NewFlagSet(subcmd, flag.ExitOnError)
    fs.StringVar(&envelope, "envelope", "", "...")
    fs.Parse(remaining)  // 只解析 "test" 之后的参数
}
```

---

## P9: Range 未使用变量

Go 不允许 `for i, step := range` 中 `step` 未使用。用 `for i := range` 替代：

```go
// ❌ 编译错误：step declared but not used
for i, step := range steps { ... }

// ✅ 正确：只用索引
for i := range steps { ... }
```

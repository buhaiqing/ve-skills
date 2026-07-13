# T14 — Reflexion HINT → Constraint 升级机制

> 任务来源：[`../autonomous-ops-roadmap.md`](../autonomous-ops-roadmap.md) §M3 (M3-3) + `docs/reflexion-memory.md:76`
> 依赖：T11
> 可并行：T13
> 预计工作量：1 天
> 状态：🟡 TODO

## 1. 目标

把 Reflexion 从"HINT 可忽略"升级为"按 count 分级生效"：
- `count < 3` → 不记录（prune）
- `3 ≤ count < 10` → HINT（注入 context，**不强制**）
- `count ≥ 10` → Constraint（policy 必须遵守；触发 T13 转译为护栏）
- `count ≥ 30` → Hard Constraint（命中即 ABORT，强制 human review）

## 2. 背景

- 当前：HINT 恒为"参考不强制"（`reflexion-memory.md:76`）
- 升级信号：M3-3 "High-freq pattern changes auto-exec threshold"
- 与 T13 协作：T13 转译护栏；本卡负责"护栏被触发时如何强制"

## 3. 产出物

### 3.1 升级机制 Go 包

**新文件**：`cmd/vet/internal/reflexion/promote/promote.go`

```go
package promote

type Level int
const (
    LevelPruned   Level = iota // count<3
    LevelHint                   // 3<=count<10
    LevelConstraint             // 10<=count<30
    LevelHard                   // count>=30
)

type Pattern struct {
    Category, Skill, Pattern, Fix string
    Count                         int
}

// LevelOf 根据 count 算 Level
func LevelOf(p Pattern) Level

// Enforce 在 GCL 决策点检查 pattern：
// - Hint：仅注入 context，不阻塞
// - Constraint：plan decision 若违反 → 强制改 ASK
// - Hard：直接 ABORT（覆盖 L4 SLO）
func Enforce(ctx context.Context, p Pattern, plan *Plan) (Level, error)
```

### 3.2 与 plan §5 风险表对齐

- 显式硬约束：count<10 → 永远不升级为 Constraint（plan §4）
- 显式硬约束：Hard 级别只对 Safety=1 的 op 生效（不破坏 L3 决策门）

### 3.3 vet 校验子命令

**新增**：`vet reflexion check --patterns <path> --plan <path>`

- 读 patterns + 当前 plan，输出每条 pattern 的 Level
- 任何 Hard 级别 pattern 命中 plan → exit 1
- Constraint 级别 pattern 违反 → warning

## 4. DoD

```
□ 1. 写入 cmd/vet/internal/reflexion/promote/promote.go
□ 2. cmd/vet 注册 vet reflexion check 子命令
□ 3. go build + go vet + go test 绿
□ 4. promote_test.go 覆盖：4 个 Level 边界（2/3/9/10/29/30）
□ 5. enforce_test.go 覆盖：Hint 不阻塞；Constraint 改 plan；Hard ABORT
□ 6. CI（validate.yml）跑 vet reflexion check（PR 阶段）
□ 7. 与 T13 集成测试：transpile 产出的 guardrail 能被 check 校验
```

## 5. 验证命令

```bash
cd cmd/vet
go build ./...
go vet ./...
go test -run TestLevelOf ./internal/reflexion/promote/ -v
go test -run TestEnforce ./internal/reflexion/promote/ -v
go test -run TestHardABORT ./internal/reflexion/promote/ -v   # count=30 + Safety=1
go test ./...

go build -o /tmp/vet .
# 构造假 patterns + plan
cat > /tmp/p.md <<'EOF'
| category | skill | pattern | fix | count |
| runtime | ve-iam-ops | token-expired | re-auth | 35 |   ← Hard
EOF
echo '{"operations":[{"skill":"ve-iam-ops","safety_class":"state-changing","blast_radius":"single","confidence":"high","safety":1}],"policy_decision":"AUTO"}' > /tmp/plan.json
! /tmp/vet reflexion check --patterns /tmp/p.md --plan /tmp/plan.json && echo "HARD_GUARD_OK"
```

## 6. 完成回报

```markdown
## T14 2026-07-XX — done
- cmd/vet/internal/reflexion/promote/ 4 级 Level + Enforce
- vet reflexion check 子命令
- 硬约束：count<10 不升级；Hard 级别只 Safety=1 生效
- T13 集成测试通过
- T15 可消费
```

## 7. 风险与回滚

| 风险 | 缓解 |
|------|------|
| Hard 级别误伤（数据尖峰导致虚假 count=30） | Hard 触发后必须 human review 才解 ABORT；T03 扩域条件同步 |
| 升级路径被绕过 | 升级只能通过 T13 转译器；手工编辑 guardrails.yaml 在 CI 中拒绝 |
| 回滚 | `git checkout cmd/vet/internal/reflexion/promote/` |

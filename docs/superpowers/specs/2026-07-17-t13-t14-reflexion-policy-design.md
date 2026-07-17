# T13 + T14 — Reflexion HINT→Constraint 升级 + Pattern→Policy 转译器 (SDD Spec)

日期：2026-07-17
对应卡片：`docs/l3-to-l4-tasks/T13-pattern-to-policy.md`、`docs/l3-to-l4-tasks/T14-reflexion-promotion.md`
上游依赖：T11（自愈遥测 + 指标包）已完成
下游消费者：T15（版本化 policy library）、T16（SLO envelope dashboard）

---

## 1. 功能描述

### 1.1 核心目标

将 Reflexion 从"HINT 可忽略"升级为"按 count 分级生效"，并实现从 `failure-patterns.md` 到 `guardrails.yaml` 的自动转译。

### 1.2 四大分级（T14）

| Level | Count 范围 | 行为 |
|-------|-----------|------|
| `LevelPruned` | count < 3 | 不记录，被 prune |
| `LevelHint` | 3 ≤ count < 10 | 注入 context，**不强制** |
| `LevelConstraint` | 10 ≤ count < 30 | Policy 必须遵守，触发 T13 转译为护栏 |
| `LevelHard` | count ≥ 30 | 命中即 ABORT，强制 human review |

### 1.3 转译流程（T13）

```
failure-patterns.md (Section 6: Incident Response + auto-generated)
    │
    ├── count < 10 → 跳过（不升级为护栏）
    └── count ≥ 10 → Transpile() → guardrails.yaml
                           │
                           └── guardrails.schema.json 校验
```

### 1.4 护栏规则结构

```yaml
guardrails:
  - id: "hash-abc123"          # hash(skill+pattern)，幂等
    skill: "ve-ecs-ops"
    trigger: "evidence_overfetch"
    action: "auto-ASK"         # auto-ASK | auto-REFUSE | auto-INFO
    severity: "medium"         # low | medium | high
    source_count: 12
    created_at: "2026-07-17"
```

---

## 2. 状态机定义

### 2.1 Pattern 生命周期

```
                    count < 3
    NEW ──────────────────────► PRUNED (删除)
     │
     │ count ≥ 3
     ▼
    HINT ──────────────────────► 注入 context，不阻塞
     │
     │ count ≥ 10
     ▼
    CONSTRAINT ────────────────► 转译为 guardrails.yaml，policy 强制遵守
     │
     │ count ≥ 30
     ▼
    HARD ──────────────────────► ABORT on hit，强制 human review
```

### 2.2 Transpile 幂等性

- ID = `hash(skill + pattern)` → 同一 pattern 多次升级产生相同 ID
- 已存在的 guardrail 被新升级覆盖（更新 count/created_at）
- 不创建重复条目

---

## 3. 异常边界处理

### 3.1 输入异常

| 场景 | 处理 |
|------|------|
| `failure-patterns.md` 不存在 | 返回 error，exit 1 |
| Section 6 表头不匹配 | 跳过该行，输出 WARN |
| Count 列非数字 | 视为 0，跳过 |
| 空行 / 注释行 | 跳过 |

### 3.2 输出异常

| 场景 | 处理 |
|------|------|
| `guardrails.yaml` 目录不存在 | 自动创建 |
| `guardrails.yaml` 已存在 | 覆盖（幂等） |
| Schema 校验失败 | 输出错误详情，不写入 |

### 3.3 运行时异常（Enforce）

| 场景 | 处理 |
|------|------|
| guardrails.yaml 不存在 | 跳过 enforce（无护栏 = 无强制） |
| Pattern 匹配到 Hard 级别 | 立即 ABORT，exit 1 |
| Constraint 级别违反 | plan decision 强制改为 ASK |
| Hint 级别 | 仅注入 context，不修改 plan |

---

## 4. API / 组件契约

### 4.1 Go 包接口

**`cmd/vet/internal/reflexion/promote/promote.go`** (T14):
```go
package promote

type Level int
const (
    LevelPruned    Level = iota // count < 3
    LevelHint                   // 3 ≤ count < 10
    LevelConstraint             // 10 ≤ count < 30
    LevelHard                   // count ≥ 30
)

type Pattern struct {
    Category, Skill, Pattern, Fix string
    Count                          int
}

func LevelOf(p Pattern) Level
func Enforce(ctx context.Context, p Pattern, plan *Plan) (Level, error)
```

**`cmd/vet/internal/reflexion/transpile/transpile.go`** (T13):
```go
package transpile

type FailurePattern struct {
    Category, Skill, Pattern, Fix string
    Count                          int
}

type Guardrail struct {
    ID          string
    Skill       string
    Trigger     string
    Action      string
    Severity    string
    SourceCount int
    CreatedAt   time.Time
}

func TranspileFile(patternsPath, outPath string) (int, error)
func Transpile(p FailurePattern) (Guardrail, bool)
```

### 4.2 CLI 子命令

```
vet reflexion promote --patterns <path>           # 输出每条 pattern 的 Level
vet reflexion transpile --patterns <path> --out <path>   # 转译 patterns → guardrails.yaml
vet reflexion check --patterns <path> --plan <path>     # 检查 Hard/Constraint 冲突
```

---

## 5. 验收标准

### T13 (Pattern→Policy 转译器)
- [ ] `transpile.go` 实现 `TranspileFile` + `Transpile`
- [ ] `guardrails.schema.json` 定义护栏 schema
- [ ] CLI 注册 `vet reflexion transpile`
- [ ] `go build ./... && go vet ./... && go test ./...` 绿
- [ ] 单测覆盖：count<10 不升级、count≥10 升级、同 ID 幂等、YAML 输出符合 schema
- [ ] 假数据端到端验证

### T14 (HINT→Constraint 升级)
- [ ] `promote.go` 实现 4 级 `LevelOf` + `Enforce`
- [ ] CLI 注册 `vet reflexion check` + `vet reflexion promote`
- [ ] `go build ./... && go vet ./... && go test ./...` 绿
- [ ] 单测覆盖：4 个 Level 边界（2/3/9/10/29/30）
- [ ] enforce 单测：Hint 不阻塞、Constraint 改 plan、Hard ABORT
- [ ] 与 T13 集成测试通过

---

## 6. 风险与缓解

| 风险 | 缓解 |
|------|------|
| Hard 级别误伤（数据尖峰） | Hard 触发后必须 human review 才解 ABORT |
| 升级路径被绕过 | 升级只能通过 T13 转译器，手工编辑 guardrails.yaml 被 CI 拒绝 |
| failure-patterns.md 格式变化 | 表头硬编码匹配，不匹配则 WARN 跳过 |
| ID 漂移 | ID = hash(skill+pattern)，幂等 |

# T13 — Pattern → Policy 转译器

> 任务来源：[`../autonomous-ops-roadmap.md`](../autonomous-ops-roadmap.md) §M3 (M3-2)
> 依赖：T11（指标数据源）, T14（升级机制）
> 可并行：T14
> 预计工作量：1.5 天
> 状态：🟡 TODO

## 1. 目标

把 `docs/failure-patterns.md` 中 `count ≥ 10` 的 pattern **自动转译**为
`incident-loop-agent/references/policies/guardrails.yaml` 的一条护栏规则。
转译器本身是 Go 工具，可单测、可在 CI 中跑。

## 2. 背景

- reflexion spec：`docs/reflexion-memory.md:63`
  *"Patterns with `count >= 10` are candidates for promotion to Anti-Patterns sections"*
- plan §4 风险："Learning feedback (M3-2/3) must not become a hard gate prematurely"
  → `count < 10` 保持 HINT；`count ≥ 10` 才升约束
- 护栏 schema 还没定义（T15 出 yaml schema）

## 3. 产出物

### 3.1 转译器（Go）

**新文件**：`cmd/vet/internal/reflexion/transpile/transpile.go`

```go
package transpile

type FailurePattern struct {
    Category string  // cli_parameter / skill_generation / cross_skill / runtime / token_efficiency / incident_response
    Skill    string
    Pattern  string
    Count    int
    Fix      string
}

type Guardrail struct {
    ID          string  // 稳定 ID（基于 pattern 哈希）
    Skill       string
    Trigger     string  // 触发条件（pattern 描述）
    Action      string  // auto-ASK | auto-REFUSE | auto-INFO
    Severity    string  // low | medium | high
    SourceCount int     // 触发时 count 值（用于追溯）
    CreatedAt   time.Time
}

// TranspileFile 读 failure-patterns.md，输出 guardrails.yaml
func TranspileFile(patternsPath, outPath string) (int, error)
// 单条 transpile，便于单测
func Transpile(p FailurePattern) (Guardrail, bool)
// count < 10 返回 false（不升级）
```

### 3.2 护栏 YAML schema

**新文件**：`incident-loop-agent/assets/guardrails.schema.json`

约束每个 guardrail 必含 `id` / `skill` / `trigger` / `action` / `severity` / `source_count`。

### 3.3 CLI 子命令

**修改**：`cmd/vet/reflexion.go`（或新建）→ `vet reflexion transpile`

```
$ vet reflexion transpile --patterns docs/failure-patterns.md \
                          --out incident-loop-agent/references/policies/guardrails.yaml
# 升级 3 条（count>=10 的 patterns）
```

## 4. DoD

```
□ 1. 写入 cmd/vet/internal/reflexion/transpile/transpile.go
□ 2. 写入 incident-loop-agent/assets/guardrails.schema.json
□ 3. cmd/vet 注册 vet reflexion transpile 子命令
□ 4. go build + go vet + go test 绿
□ 5. transpile_test.go 覆盖：count<10 不升级；count≥10 升级；同 ID 幂等
□ 6. guardrails.yaml 解析后必符合 schema（用 jsonschema 库断言）
□ 7. CI 在 PR 阶段跑 vet reflexion transpile，diff 输出展示给 reviewer
```

## 5. 验证命令

```bash
cd cmd/vet
go build ./...
go vet ./...
go test -run TestTranspile ./internal/reflexion/transpile/ -v
go test -run TestIdempotency ./internal/reflexion/transpile/ -v   # 跑 2 次输出相同

go build -o /tmp/vet .
# 构造假 patterns.md
cat > /tmp/patterns.md <<'EOF'
| category | skill | pattern | fix | count |
| cli_parameter | ve-ecs-ops | InvalidInstanceId | use i-xxx | 15 |
| cli_parameter | ve-redis-ops | OOM | check ttl | 8 |   ← 不到 10，不升级
| runtime | ve-iam-ops | token-expired | re-auth | 12 |
EOF
/tmp/vet reflexion transpile --patterns /tmp/patterns.md \
  --out /tmp/guardrails.yaml
# 期望：2 条 guardrail（count=15 和 count=12），不含 OOM 那条
grep -c "^  - id:" /tmp/guardrails.yaml
python3 -c "import yaml; gs=yaml.safe_load(open('/tmp/guardrails.yaml')); assert len(gs['guardrails'])==2; print('YAML_OK')"
```

## 6. 完成回报

```markdown
## T13 2026-07-XX — done
- cmd/vet/internal/reflexion/transpile/ 转译器
- guardrails.schema.json
- vet reflexion transpile 子命令
- 升级门槛：count≥10（plan §4 风险约束）
- T15 可消费（policy library 用本转译结果作为源）
```

## 7. 风险与回滚

| 风险 | 缓解 |
|------|------|
| 护栏误升（count=11 但实为偶发） | 30 天窗口 + 至少 2 次独立 trace（与 T03 扩域条件一致） |
| ID 漂移（同一 pattern 升级 2 次产生不同 ID） | ID = `hash(skill+pattern)`，幂等 |
| 回滚 | `git checkout cmd/vet/internal/reflexion/transpile/ incident-loop-agent/assets/guardrails.schema.json` |

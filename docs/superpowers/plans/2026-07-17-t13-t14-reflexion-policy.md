# T13 + T14 — Reflexion 升级 + Pattern 转译执行计划

日期：2026-07-17
对应 Spec：`docs/superpowers/specs/2026-07-17-t13-t14-reflexion-policy-design.md`
对应卡片：T13-pattern-to-policy、T14-reflexion-promotion

---

## 里程碑拆分

### M1 — T14 Reflexion 4 级升级机制（基础设施，T13 依赖）

| # | 任务 | 文件 | 验证点 (DoD) |
|---|------|------|--------------|
| T1 | 创建 `cmd/vet/internal/reflexion/promote/promote.go`：`Level` 类型、`LevelOf()`、`Enforce()` | `promote/promote.go` | 纯函数，无外部依赖 |
| T2 | 创建 `promote/promote_test.go`：4 级边界（2/3/9/10/29/30）、Hint/Constraint/Hard 行为 | `promote/promote_test.go` | 7 个测试用例全绿 |
| T3 | 创建 `cmd/vet/reflexion.go`：`runReflexion()` 调度器 + `vet reflexion promote` 子命令 | `reflexion.go` | 输出每条 pattern 的 Level |
| T4 | 创建 `cmd/vet/reflexion.go`：`vet reflexion check` 子命令 | `reflexion.go` | Hard 命中 exit 1；Constraint 违反 warning |
| T5 | 注册到 `main.go`：`case "reflexion": runReflexion(args)` | `main.go` | 单行新增 |

### M2 — T13 Pattern→Policy 转译器

| # | 任务 | 文件 | 验证点 (DoD) |
|---|------|------|--------------|
| T6 | 创建 `cmd/vet/internal/reflexion/transpile/transpile.go`：`TranspileFile` + `Transpile` | `transpile/transpile.go` | 读 markdown → 输出 YAML |
| T7 | 创建 `transpile/transpile_test.go`：count<10 不升级、count≥10 升级、同 ID 幂等、空文件、格式错误 | `transpile/transpile_test.go` | 5+ 测试用例全绿 |
| T8 | 创建 `incident-loop-agent/assets/guardrails.schema.json` | `guardrails.schema.json` | JSON Schema draft 2020-12 |
| T9 | `vet reflexion transpile` 子命令注册（reflexion.go） | `reflexion.go` | 端到端可执行 |
| T10 | T13+T14 集成测试：transpile 产出的 guardrail 能被 promote 的 Enforce 识别 | 集成测试 | Constraint/Hard 级别互通 |

### M3 — 收尾

| # | 任务 | 文件 | 验证点 (DoD) |
|---|------|------|--------------|
| T11 | `go build ./... && go vet ./... && go test ./...` | 全仓库 | exit 0 |
| T12 | `vet validate --root .` 保持绿 | repo root | 无新增失败 |
| T13 | 更新 L3→L4 索引状态：T13 ✅、T14 ✅ | `01-index.md` | 状态更新 |

---

## 依赖图

```
M1 (T1 → T2 → T3 → T4 → T5)   ← T14 基础设施
M2 (T6 → T7 → T8 → T9 → T10)   ← T13 转译器（T10 依赖 M1）
M3 (T11 → T12 → T13)            ← 收尾（依赖 M1 + M2）
```

M1 和 M2 的 T6-T9 可部分并行（T6-T9 不依赖 M1），但 T10 集成测试需要 M1 完成。

---

## 关键文件清单

### 新增文件
```
cmd/vet/internal/reflexion/promote/promote.go       # T14: LevelOf + Enforce
cmd/vet/internal/reflexion/promote/promote_test.go   # T14: 测试
cmd/vet/internal/reflexion/transpile/transpile.go    # T13: TranspileFile + Transpile
cmd/vet/internal/reflexion/transpile/transpile_test.go # T13: 测试
cmd/vet/reflexion.go                                 # CLI 调度器
incident-loop-agent/assets/guardrails.schema.json    # 护栏 schema
```

### 修改文件
```
cmd/vet/main.go                                      # +case "reflexion"
docs/l3-to-l4-tasks/01-index.md                      # T13/T14 状态更新
```

---

## 验证命令

```bash
# 构建 + 检查
cd cmd/vet && go build ./... && go vet ./...

# 单元测试
go test ./internal/reflexion/promote/ -v
go test ./internal/reflexion/transpile/ -v

# 全量测试
go test ./...

# 全局门禁
vet validate --root .

# T13 端到端验证（假数据）
go build -o /tmp/vet .
cat > /tmp/patterns.md <<'EOF'
## 6. Incident Response Failures
| Scenario | Failure Pattern | Root Cause | Fix | Count |
|----------|----------------|------------|-----|-------|
| Alarm triage | evidence_overfetch | AI makes >10 calls | cap at 15 | 15 |
| Diagnosis | data_collection_timeout | timeout on large dataset | paginate | 8 |
| Safety gate | safety_gate_bypass | silent destructive default | explicit ASK | 12 |
EOF
/tmp/vet reflexion transpile --patterns /tmp/patterns.md --out /tmp/guardrails.yaml
# 期望：2 条 guardrail（count=15 和 count=12）
grep -c "^- id:" /tmp/guardrails.yaml  # → 2

# T14 端到端验证
/tmp/vet reflexion promote --patterns /tmp/patterns.md
# 期望：3 条 pattern，Level 分别为 Constraint/Hint/Constraint
```

---

## 风险与回滚

- 每个 T 独立 commit；出问题 `git revert` 单 commit
- 回滚 M1：`git checkout cmd/vet/internal/reflexion/promote/ cmd/vet/reflexion.go cmd/vet/main.go`
- 回滚 M2：`git checkout cmd/vet/internal/reflexion/transpile/ incident-loop-agent/assets/guardrails.schema.json`
- Hard 级别误伤：Hard 触发后必须 human review 才解 ABORT

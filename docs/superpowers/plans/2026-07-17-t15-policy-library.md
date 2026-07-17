# T15 — 版本化 Policy Library 执行计划

日期：2026-07-17
对应 Spec：`docs/superpowers/specs/2026-07-17-t15-policy-library-design.md`
对应卡片：T15-versioned-policy-lib

---

## 里程碑拆分

### M1 — 策略库目录结构 + 文档

| # | 任务 | 文件 | 验证点 (DoD) |
|---|------|------|--------------|
| T1 | 创建 `incident-loop-agent/references/policies/README.md` | README.md | 版本策略 + 回滚方法说明 |
| T2 | 创建 `incident-loop-agent/references/policies/CHANGELOG.md` | CHANGELOG.md | 空表格 + 示例行 |

### M2 — Policy Loader (Go)

| # | 任务 | 文件 | 验证点 (DoD) |
|---|------|------|--------------|
| T3 | 创建 `cmd/vet/internal/policy/loader.go`：PolicySet + Load + DiffPolicySets | `policy/loader.go` | 纯函数，读取 policies/ 目录 |
| T4 | 创建 `policy/loader_test.go`：空目录、1 文件、全量、Diff 一致、Diff 差异 | `policy/loader_test.go` | 5+ 测试用例全绿 |

### M3 — CLI 子命令 + CI

| # | 任务 | 文件 | 验证点 (DoD) |
|---|------|------|--------------|
| T5 | 创建 `cmd/vet/policy.go`：runPolicy() + load/diff/check-changelog 子命令 | `policy.go` | 3 个子命令可用 |
| T6 | 注册到 `main.go`：`case "policy": runPolicy(args)` | `main.go` | 单行新增 |
| T7 | CI：`.github/workflows/validate.yml` 添加 `vet policy check-changelog` | `validate.yml` | PR 阶段检查 |

### M4 — 收尾

| # | 任务 | 文件 | 验证点 (DoD) |
|---|------|------|--------------|
| T8 | `go build ./... && go vet ./... && go test ./...` | 全仓库 | exit 0 |
| T9 | `vet validate --root .` 保持绿 | repo root | 无新增失败 |
| T10 | 端到端验证：load/diff/check-changelog | — | 功能正常 |
| T11 | 更新 `01-index.md` 状态 | `01-index.md` | T15 ✅ DONE |

---

## 依赖图

```
M1 (T1 → T2)    ← 纯文档，独立
M2 (T3 → T4)    ← Go 包，独立
M3 (T5 → T6 → T7) ← 依赖 M2（T5 import policy 包）
M4 (T8 → T9 → T10 → T11) ← 依赖 M1+M2+M3
```

M1 和 M2 可并行。M3 依赖 M2。

---

## 关键文件清单

### 新增文件
```
incident-loop-agent/references/policies/README.md        # T1: 目录说明
incident-loop-agent/references/policies/CHANGELOG.md     # T2: 版本变更记录
cmd/vet/internal/policy/loader.go                        # T3: PolicySet + Load + Diff
cmd/vet/internal/policy/loader_test.go                   # T4: 测试
cmd/vet/policy.go                                        # T5: CLI 调度器
```

### 修改文件
```
cmd/vet/main.go                                          # T6: +case "policy"
.github/workflows/validate.yml                           # T7: +policy check-changelog
docs/l3-to-l4-tasks/01-index.md                          # T11: 状态更新
```

---

## 验证命令

```bash
# 构建 + 测试
cd cmd/vet && go build ./... && go vet ./... && go test ./...

# Policy loader 测试
go test ./internal/policy/ -v

# 端到端
go build -o /tmp/vet .
/tmp/vet policy load --root .          # 输出 JSON
/tmp/vet policy check-changelog --root .  # 初始状态无变更 → 通过
```

# T15 — 版本化 Policy Library

> 任务来源：[`../autonomous-ops-roadmap.md`](../autonomous-ops-roadmap.md) §M3 (M3-4)
> 依赖：T13, T14
> 预计工作量：1 天
> 状态：🟡 TODO

## 1. 目标

把策略 / 护栏做成"as code"：目录化、版本化、可 diff、可回滚。

## 2. 背景

- plan M3-4："Versioned policy library (guardrails as code, reviewable)；Policy diffs tracked in repo"
- 策略来源：T13 转译器（T13 写入 guardrails.yaml）+ T14 升级机制
- 当前状态：所有 policy / 护栏文件零散散在 `incident-loop-agent/references/`

## 3. 产出物

### 3.1 目录结构

```
incident-loop-agent/
├── references/
│   └── policies/                  # ← 本卡建立
│       ├── README.md              # 目录说明 + 版本策略
│       ├── execution-risk.md      # 已有（T01）
│       ├── domain-allowlist.md    # 已有（T03）
│       ├── guardrails.yaml        # T13 转译产物（git 追踪）
│       ├── CHANGELOG.md           # 版本化变更记录
│       └── vN/                    # 未来快照（按需）
└── assets/
    └── guardrails.schema.json     # 已有（T13）
```

### 3.2 Policy Loader（Go）

**新文件**：`cmd/vet/internal/policy/loader.go`

```go
package policy

// Load 读 incident-loop-agent/references/policies/* 并合并
// 返回内部统一 PolicySet（含 execution-risk + domain-allowlist + guardrails）
func Load(rootPath string) (*PolicySet, error)

// DiffPolicySets 比较两个 PolicySet，给出 added/removed/changed
func DiffPolicySets(old, new *PolicySet) []PolicyChange
```

### 3.3 版本化与 changelog

**新增**：`incident-loop-agent/references/policies/CHANGELOG.md`

每次 PR 涉及 policy 改动，CHANGELOG 必加一行：
```
| YYYY-MM-DD | vN | author | summary |
```

**修改**：`incident-loop-agent/references/policies/README.md` 新增一节：
- 版本策略（major.minor）
- 升级窗口（每 N 天审）
- 回滚方法（`git checkout HEAD~N -- incident-loop-agent/references/policies/`）

### 3.4 CI 检查

**修改**：`.github/workflows/validate.yml` 加：
- 任何 policy 文件改动必须同时改 CHANGELOG.md
- 否则 CI 失败（`vet policy check-changelog` 新增子命令）

## 4. DoD

```
□ 1. 创建 incident-loop-agent/references/policies/ 目录（含 README.md + CHANGELOG.md）
□ 2. 写入 cmd/vet/internal/policy/loader.go（Load + DiffPolicySets）
□ 3. cmd/vet 注册 vet policy load / vet policy diff / vet policy check-changelog 子命令
□ 4. go build + go vet + go test 绿
□ 5. loader_test.go 覆盖：读空目录、读 1 文件、读全部、Diff 一致
□ 6. CI：policy 改动无 CHANGELOG 必阻断
□ 7. README.md 解释版本策略 + 回滚方法
```

## 5. 验证命令

```bash
cd cmd/vet
go build ./...
go vet ./...
go test -run TestLoad ./internal/policy/ -v
go test -run TestDiff ./internal/policy/ -v
go test ./...

go build -o /tmp/vet .
/tmp/vet policy load --root .      # 输出 JSON
/tmp/vet policy diff --old HEAD~1 --new HEAD --root .   # 与上次 commit 比

# 故意不带 CHANGELOG 改 policy
cp incident-loop-agent/references/policies/guardrails.yaml /tmp/g.yaml
echo "# unrelated change" >> incident-loop-agent/references/policies/guardrails.yaml
! /tmp/vet policy check-changelog --root . && echo "CHANGELOG_GUARD_OK"
git checkout -- incident-loop-agent/references/policies/guardrails.yaml
```

## 6. 完成回报

```markdown
## T15 2026-07-XX — done
- incident-loop-agent/references/policies/ 目录结构
- cmd/vet/internal/policy/ loader + diff
- vet policy {load,diff,check-changelog} 子命令
- CHANGELOG 强制门禁
- T16 可消费（policy library 是 SLO 引擎的输入）
```

## 7. 风险与回滚

| 风险 | 缓解 |
|------|------|
| 策略库膨胀失控 | 月度归档旧 vN/；护栏总数上限 100 |
| Diff 算法 O(N²) | 现阶段 N<100 够用；>500 重写为哈希索引 |
| 回滚 | `git checkout incident-loop-agent/references/policies/ + cmd/vet/internal/policy/` |

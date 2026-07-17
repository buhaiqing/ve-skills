# T15 — 版本化 Policy Library (SDD Spec)

日期：2026-07-17
对应卡片：`docs/l3-to-l4-tasks/T15-versioned-policy-lib.md`
上游依赖：T13（Pattern→Policy 转译器）+ T14（Reflexion 升级机制）已完成
下游消费者：T16（SLO envelope dashboard）

---

## 1. 功能描述

### 1.1 核心目标

将策略/护栏做成 "as code"：目录化、版本化、可 diff、可回滚。策略来源包括 T13 转译器产出的 guardrails.yaml 和 T14 升级机制。

### 1.2 策略库目录结构

```
incident-loop-agent/references/policies/
├── README.md              # 目录说明 + 版本策略 + 回滚方法
├── CHANGELOG.md           # 版本化变更记录
├── execution-risk.md      # 已有（T01）：3x3 决策矩阵
├── domain-allowlist.md    # 已有（T03）：AUTO 白名单
└── guardrails.yaml        # T13 转译产物（git 追踪）
```

### 1.3 Policy Loader

一个 Go 包 `cmd/vet/internal/policy/loader.go`，提供：
- `Load(rootPath) (*PolicySet, error)` — 读取 policies/ 下所有文件，合并为统一 PolicySet
- `DiffPolicySets(old, new) []PolicyChange` — 比较两个 PolicySet，输出 added/removed/changed

### 1.4 CLI 子命令

```
vet policy load --root .           # 加载并输出 JSON
vet policy diff --old HEAD~1 --new HEAD --root .  # 比较两个版本
vet policy check-changelog --root .              # 检查 policy 变更是否带 CHANGELOG
```

### 1.5 版本化 Changelog

- `incident-loop-agent/references/policies/CHANGELOG.md`：每次 policy 变更必加一行
- CI 门禁：policy 文件改动无对应 CHANGELOG → 阻断

---

## 2. API / 组件契约

### 2.1 PolicySet 类型

```go
type PolicySet struct {
    ExecutionRisk  ExecutionRiskPolicy   `json:"execution_risk"`
    DomainAllowlist []string             `json:"domain_allowlist"`
    Guardrails     []Guardrail           `json:"guardrails"`
}

type ExecutionRiskPolicy struct {
    AutoConditions []string `json:"auto_conditions"`
    AskConditions  []string `json:"ask_conditions"`
    RefuseConditions []string `json:"refuse_conditions"`
}
```

### 2.2 Load 逻辑

1. 读取 `execution-risk.md` → 解析 AUTO/ASK/REFUSE 条件
2. 读取 `domain-allowlist.md` → 提取 8 个 eligible skills
3. 读取 `guardrails.yaml`（如果存在）→ 解析 YAML
4. 合并为 PolicySet

### 2.3 Diff 逻辑

- `added`: new.PolicySet 中存在但 old.PolicySet 中不存在的条目
- `removed`: old.PolicySet 中存在但 new.PolicySet 中不存在的条目
- `changed`: 两者都存在但内容不同的条目

### 2.4 check-changelog 逻辑

1. `git diff --name-only HEAD~1 HEAD` 检查 policies/ 下文件是否变更
2. 如果有变更 → 检查 CHANGELOG.md 是否也变更
3. 如果有 policy 变更但 CHANGELOG.md 未变更 → exit 1

---

## 3. 验收标准

- [ ] `incident-loop-agent/references/policies/README.md` + `CHANGELOG.md` 创建
- [ ] `cmd/vet/internal/policy/loader.go` 实现 Load + Diff
- [ ] `cmd/vet/policy.go` 实现 3 个子命令
- [ ] `main.go` 注册 `vet policy`
- [ ] `go build/vet/test` 全绿
- [ ] loader_test.go 覆盖：空目录、1 文件、全部文件、Diff 一致/差异
- [ ] `.github/workflows/validate.yml` 添加 `vet policy check-changelog` 步骤
- [ ] 端到端验证通过

---

## 4. 风险与缓解

| 风险 | 缓解 |
|------|------|
| 策略库膨胀 | 护栏总数上限 100；月度归档旧 vN/ |
| Diff O(N²) | N<100 够用；>500 重写为哈希索引 |
| 回滚 | `git checkout incident-loop-agent/references/policies/` |

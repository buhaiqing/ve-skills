# Policy Library — 版本化策略库

> 本目录是 incident-loop-agent 的策略库（policy library）。所有策略文件 "as code"，版本化、可 diff、可回滚。

## 目录结构

| 文件 | 用途 | 更新方式 |
|------|------|----------|
| `execution-risk.md` | L3 执行风险决策矩阵（3x3 risk × blast_radius） | 手动维护 |
| `domain-allowlist.md` | AUTO 资格白名单（技能 + 症状） | 手动维护 |
| `guardrails.yaml` | T13 转译器从 failure-patterns.md 自动产出的护栏规则 | 自动生成（`vet reflexion transpile`） |
| `CHANGELOG.md` | 策略变更记录 | 手动追加（PR 强制要求） |
| `README.md` | 本文件 | 手动维护 |

## 版本策略

- 策略库采用 **Semantic Versioning**（major.minor.patch）
- **major**: 决策矩阵核心逻辑变更（如 AUTO 条件扩大/缩小）
- **minor**: 新增策略文件 / 白名单扩展
- **patch**: 护栏规则更新 / 文档修正

当前版本：**v1.0.0**

## 升级窗口

- 策略变更在 **每 30 天评审窗口** 内执行
- 紧急安全修复可随时 bypass 窗口，但需在 CHANGELOG 中标记 `[HOTFIX]`

## 回滚方法

```bash
# 回滚到 N 个 commit 前的版本
git checkout HEAD~N -- incident-loop-agent/references/policies/

# 回滚特定文件
git checkout HEAD~N -- incident-loop-agent/references/policies/guardrails.yaml

# 查看策略变更历史
git log --oneline -- incident-loop-agent/references/policies/
```

## CI 门禁

- PR 中任何 `policies/` 下文件的变更必须伴随 `CHANGELOG.md` 更新
- 门禁命令：`vet policy check-changelog --root .`
- 违反时 CI 阻断（hard gate）

## 护栏上限

- `guardrails.yaml` 中护栏条目上限：**100 条**
- 超出上限时：按 `source_count` 升序淘汰低 count 条目
- 淘汰前保留快照：`cp guardrails.yaml v1/guardrails-$(date +%Y%m%d).yaml`

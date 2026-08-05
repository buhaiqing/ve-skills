# Policy Changelog

> 每次策略文件变更必须在此追加一行记录。CI 强制检查此文件与 policies/ 目录同步。

| 日期 | 版本 | 作者 | 变更摘要 |
|------|------|------|----------|
| 2026-07-17 | v1.0.0 | T15 | 初始版本：execution-risk.md + domain-allowlist.md + guardrails.schema.json |
| 2026-08-05 | v1.1.0 | T16 | autonomy-envelope.md 嵌入 fenced YAML `domains:` 块，使 `vet autonomy test --envelope` 可解析（L4 出口 DoD #10） |

## 变更规则

- 任何 `incident-loop-agent/references/policies/` 下文件变更 → 必须追加一行
- 仅追加，不删除历史行
- `[HOTFIX]` 标记用于紧急安全修复
- `[AUTO]` 标记用于 T13 转译器自动产出的变更

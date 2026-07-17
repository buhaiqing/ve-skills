# Memory Index

> 本索引指向 `.runtime/memory/` 下的所有记忆文件。
> 每次新增或变更 memory 文件时，必须同步更新此索引。

## 记忆文件

- [failure-patterns.json](failure-patterns.json) — GCL trace 自动写回的结构化 failure pattern，dedup by (skill, pattern)
- [execution-stats.json](execution-stats.json) — 执行统计（成功率、平均耗时、按 skill/operation 分组）
- [policy-decisions.jsonl](policy-decisions.jsonl) — 决策轨迹（AUTO/ASK/REFUSE），用于审计和升级

## 更新规则

- failure-patterns.json：GCL MAX_ITER/SAFETY_FAIL 时自动追加
- execution-stats.json：每次 GCL 执行后更新
- policy-decisions.jsonl：incident-loop-agent Step 5a 决策时追加
- count ≥ 10 的 pattern 触发 T13 转译为 guardrails.yaml

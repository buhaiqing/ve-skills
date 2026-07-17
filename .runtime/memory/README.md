# .runtime/memory/ — Agent 运行时长期记忆

> 本目录存放 agent 运行时产生的、跨 session 可复用的结构化记忆数据。
> 与 `docs/failure-patterns.md`（设计文档/种子数据）不同，这里的文件是 **自动生成** 的运行时产物。
>
> **不提交到 git**（`.runtime/` 已在 `.gitignore`）。

## 目录结构

```
.runtime/memory/
├── README.md                   ← 本文件
├── MEMORY.md                   ← 记忆索引（按主题可检索）
├── failure-patterns.json       ← 结构化 failure pattern（GCL trace 自动写回）
├── execution-stats.json        ← 执行统计（成功率、平均耗时等）
├── policy-decisions.jsonl      ← 决策轨迹（AUTO/ASK/REFUSE 决策记录）
└── guardrails-snapshot.yaml    ← 护栏快照（T13 转译前状态）
```

## 数据流

```
GCL 执行 → trace.failure_pattern
    │
    ▼
failure-patterns.json  (自动写回，dedup by skill+pattern, count++)
    │
    │  count ≥ 10
    ▼
T13 transpile → guardrails.yaml → T15 policy library
    │
    ▼
T14 Enforce → incident-loop-agent Step 5a 决策
```

## 与 docs/failure-patterns.md 的关系

| 文件 | 性质 | 写入方式 | 用途 |
|------|------|---------|------|
| `docs/failure-patterns.md` | 设计文档 + 种子数据 | 手动维护 | 新人理解、模式文档 |
| `.runtime/memory/failure-patterns.json` | 运行时数据 | 自动写回 | Agent 决策、模式升级 |

# 运行轨迹自动学习体系 — 架构设计

> 本文档描述 incident-loop-agent 系统如何从 GCL 运行轨迹中自动学习、提炼、升级为策略规则的完整架构。
> 对应 `docs/autonomous-ops-roadmap.md` M3 学习反馈层。

---

## 1. 核心理念

系统通过**三层递进**实现从"运行 → 学习 → 决策"的闭环：

```
L1 记录层 (Record)      →  L2 提炼层 (Refine)      →  L3 应用层 (Apply)
─────────────────────────────────────────────────────────────────────────
运行轨迹 (trace)         →  failure pattern           →  policy guardrail
原始数据                  →  结构化 + count            →  强制执行规则
audit-results/           →  .runtime/memory/           →  incident-loop-agent/references/policies/
```

---

## 2. 完整数据流

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        GCL Runner (vet gcl run)                          │
│                                                                          │
│  每次执行产生:                                                            │
│  ├── GCL Trace (audit-results/gcl-trace-YYYYMMDD-HHMMSS.json)           │
│  │     └── 含 final.failure_pattern (如果执行失败)                        │
│  ├── Self-Healing Log (audit-results/ve-self-healing.log)               │
│  └── writebackFailurePattern()                                          │
│        ├── docs/failure-patterns.md ← 追加 markdown 表 (保持兼容)        │
│        └── .runtime/memory/failure-patterns.json ← 结构化持久化 (新增)    │
│              └── (skill, pattern) 去重，count++                          │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                     GCL Aggregator (vet gcl trace)                       │
│                                                                          │
│  定期聚合:                                                                │
│  ├── 读取所有 audit-results/gcl-trace-*.json                             │
│  ├── 生成 Quality Summary (audit-results/gcl-quality-summary-*.json)     │
│  └── UpdateFailurePatternsFile() → 刷新 docs/failure-patterns.md         │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    Reflexion Promotion (vet reflexion)                    │
│                                                                          │
│  手动/自动触发:                                                           │
│  ├── vet reflexion promote  → 查看每条 pattern 的升级等级                 │
│  ├── vet reflexion check    → 检查 Hard/Constraint 冲突 (CI 门禁)         │
│  └── vet reflexion transpile → count≥10 的 pattern → guardrails.yaml     │
│        └── 输出到 incident-loop-agent/references/policies/guardrails.yaml │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│               Policy Evaluation (incident-loop-agent Step 5a)            │
│                                                                          │
│  决策时读取:                                                              │
│  ├── execution-risk.md     → 3x3 决策矩阵 (AUTO/ASK/REFUSE)               │
│  ├── domain-allowlist.md   → 技能白名单                                   │
│  └── guardrails.yaml       → 从运行轨迹提炼的强制护栏 (T13 产出)           │
│                                                                          │
│  输出: {{policy.decision}} = AUTO | ASK | REFUSE                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 3. 文件角色与生命周期

### 3.1 运行时数据（不提交，gitignored）

| 文件 | 生产者 | 消费者 | 格式 | 生命周期 |
|------|--------|--------|------|----------|
| `audit-results/gcl-trace-*.json` | `vet gcl run` | `vet gcl trace` | JSON | 每次 GCL 执行产出一个 |
| `audit-results/gcl-quality-summary-*.json` | `vet gcl trace` | 人工/CI 审阅 | JSON | 每次聚合产出一个 |
| `audit-results/ve-self-healing.log` | `vet gcl run --heal=smart` | `vet gcl heal-stats` | CSV 行 | 追加模式 |
| `.runtime/memory/failure-patterns.json` | `vet gcl run` (writeback) | `vet gcl run` (pre-flight) | JSON | 持续累积，count 递增 |
| `.runtime/memory/execution-stats.json` | (待实现) | (待实现) | JSON | 持续累积 |
| `.runtime/memory/policy-decisions.jsonl` | incident-loop-agent Step 5a | T15 policy diff | JSONL | 追加模式 |

### 3.2 提炼产物（提交，git tracked）

| 文件 | 生产者 | 消费者 | 格式 | 触发条件 |
|------|--------|--------|------|----------|
| `docs/failure-patterns.md` | 手动 (seed) + GCL writeback (auto) | `vet gcl run` (pre-flight) | Markdown | 手动 + 每次 GCL 失败 |
| `incident-loop-agent/references/policies/guardrails.yaml` | `vet reflexion transpile` | Policy Loader + Step 5a | YAML | count ≥ 10 时手动/自动触发 |
| `incident-loop-agent/references/policies/CHANGELOG.md` | 手动 | `vet policy check-changelog` (CI) | Markdown | 每次 policy 变更 |

### 3.3 设计规范（提交，git tracked）

| 文件 | 用途 |
|------|------|
| `docs/reflexion-memory.md` | Reflexion 记忆系统规范（架构 + 规则） |
| `docs/autonomous-ops-roadmap.md` | 自治运维路线图（M3 学习反馈层） |
| `docs/gcl-spec.md` | GCL 规范（含 Reflexion Integration 章节） |

---

## 4. 升级路径：从 HINT 到 CONSTRAINT

```
                    count < 3
    NEW ──────────────────────► PRUNED (从 memory 删除)
     │
     │ count ≥ 3
     ▼
    HINT ──────────────────────► 注入 Generator context，不强制
     │                             .runtime/memory/failure-patterns.json
     │
     │ count ≥ 10
     ▼
    CONSTRAINT ────────────────► T13 transpile → guardrails.yaml
     │                             incident-loop-agent Step 5a 读取并强制遵守
     │
     │ count ≥ 30
     ▼
    HARD ──────────────────────► ABORT on hit，强制 human review
                                   CI 门禁: vet reflexion check 阻断
```

### 各等级的文件对应

| Level | 存储位置 | 注入方式 | 强制程度 |
|-------|---------|---------|----------|
| Pruned | 从 `.runtime/memory/failure-patterns.json` 删除 | 不注入 | — |
| Hint | `.runtime/memory/failure-patterns.json` (count 3-9) | `loadKnownFailurePatterns()` → GCL_KNOWN_FAILURE_PATTERNS env var | 0% |
| Constraint | `guardrails.yaml` (count 10-29) | Policy Loader → Step 5a decision | 100% (强制改为 ASK) |
| Hard | `guardrails.yaml` (count ≥ 30) | Policy Loader → Step 5a decision | 100% (ABORT) |

---

## 5. 自动学习能产生什么

### 5.1 从 GCL Trace 中提取

| 数据 | 来源字段 | 用途 |
|------|---------|------|
| `FailurePattern.Skill` | trace.final.failure_pattern.skill | 按技能分组统计 |
| `FailurePattern.Pattern` | trace.final.failure_pattern.error | 错误模式去重 |
| `FailurePattern.Category` | trace.final.failure_pattern.category | 分类统计 (cli_parameter/runtime/cross_skill 等) |
| `FailurePattern.Fix` | trace.final.failure_pattern.fix | 修复建议注入 pre-flight |
| `CriticRecord.Scores` | trace.iterations[].critic.scores | 评分趋势分析 |
| `PolicyDecision` | trace.iterations[].policy_decision | 决策审计 + 白名单扩展依据 |

### 5.2 从 Self-Healing Log 中提取

| 数据 | 用途 |
|------|------|
| HealClass (retryable/rate_limit/fatal) | 错误分类准确性统计 |
| PathName (选择的愈合路径) | 路径有效性排名 |
| Cost/DurationMs | 性能基线 + 异常检测 |

### 5.3 最终产出物

| 产出 | 文件 | 触发条件 |
|------|------|----------|
| **Failure Pattern 知识库** | `.runtime/memory/failure-patterns.json` | 每次 GCL 失败自动追加 |
| **Quality Summary** | `audit-results/gcl-quality-summary-*.json` | `vet gcl trace` 聚合 |
| **Guardrails (护栏规则)** | `guardrails.yaml` | count ≥ 10 时 transpile |
| **Decision Audit Trail** | `.runtime/memory/policy-decisions.jsonl` | 每次 Step 5a 决策 |
| **Execution Statistics** | `.runtime/memory/execution-stats.json` | 每次 GCL 执行 |

---

## 6. 当前状态与待补全

### 已实现 (T09-T15)

| 组件 | 状态 | 说明 |
|------|------|------|
| GCL Trace 写入 | ✅ | `PersistTrace()` → `audit-results/gcl-trace-*.json` |
| Failure Pattern 提取 | ✅ | `extractFailurePattern()` → 5 类正则匹配 |
| Markdown 写回 | ✅ | `UpdateFailurePatternsFile()` → `docs/failure-patterns.md` |
| Markdown 读取 (pre-flight) | ✅ | `loadKnownFailurePatterns()` → GCL_KNOWN_FAILURE_PATTERNS env |
| GCL Trace 聚合 | ✅ | `CmdAggregate()` → quality summary |
| Self-Healing Log | ✅ | `recordHealPath()` → `audit-results/ve-self-healing.log` |
| Promotion 机制 | ✅ | `vet reflexion promote/check` → Level 分级 |
| Pattern→Policy 转译 | ✅ | `vet reflexion transpile` → `guardrails.yaml` |
| Policy Library | ✅ | `vet policy load/diff/check-changelog` |
| Policy Evaluation | ✅ | incident-loop-agent Step 5a |

### 已实现 (2026-07-17)

| 组件 | 状态 | 说明 |
|------|------|------|
| JSON 结构化写回 | ✅ | `memory.AppendFailurePattern()` → `.runtime/memory/failure-patterns.json` |
| JSON 读取 (pre-flight) | ✅ | `loadKnownFailurePatterns()` 从 JSON 加载（fallback markdown） |
| count 递增 | ✅ | 去重 + count++ 替代覆盖 |
| 自动触发 T13 | ✅ | count ≥ 10 时 in-process transpile → guardrails.yaml |
| 原子写入 | ✅ | `writeStore` 使用 temp file + rename 防止崩溃损坏 |

### 待实现

| 组件 | 状态 | 说明 |
|------|------|------|
| Decision Audit Trail | 🟡 TODO | Step 5a 决策时写入 `policy-decisions.jsonl` |
| Execution Stats | 🟡 TODO | 每次 GCL 执行后更新统计 |

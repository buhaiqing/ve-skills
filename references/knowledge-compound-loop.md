## Knowledge Compound Loop（知识复利循环）

> **铁律 — 不可打破。** 每次工作都必须让系统变得更好。禁止"只消费知识、不沉淀知识"的行为。

### 复利触发器

| Trigger | Action | Compound Effect |
|---------|--------|-----------------|
| Go interface 变更 | `codegraph impact` → 更新 ARCHITECTURE.md §3 | 文档始终反映当前代码 |
| Runtime failure | Reflexion → `docs/failure-patterns.md` → spec 更新 | 失败不再重复 |
| ADR 决策 | ADR rationale → 代码注释 `// ADR-0001: ...` | 决策理由可追溯 |
| Wave C 完成 | Post-mortem → ARCHITECTURE.md §5 更新 | 经验沉淀 |
| Bug fix | 根因 → `failure-patterns.json` → 测试用例 | 回归免疫 |
| Code review | 发现 → `AGENTS.md` 规则补充 | 规则进化 |

### 复利度量（Compound Interest Metrics）

| Metric | Target | Measurement |
|--------|--------|-------------|
| Knowledge reuse rate | ≥80% | Agent sessions 加载 ARCHITECTURE.md vs 读 raw specs 的比例 |
| Failure recurrence | ≤5% | 匹配 `failure-patterns.md` 已有模式的失败占比 |
| Spec drift | 0 | ARCHITECTURE.md 契约 vs 实际 Go 签名的差异数 |
| Decision traceability | 100% | ADR 有对应代码注释的比例 |
| 文档维护成本 | ↓ 每季 | 每季度文档更新所需的 agent token 数 |

### 反模式（禁止）

- ❌ **知识消费不沉淀**：完成任务后不更新 failure-patterns / ARCHITECTURE.md
- ❌ **重复失败**：同一类 bug 出现第二次（说明 Reflexion 闭环断裂）
- ❌ **文档漂移**：代码改了但 ARCHITECTURE.md 没更新（spec drift > 0）
- ❌ **决策失忆**：ADR 写了但代码里没有注释引用（traceability < 100%）

### 复利检查清单（每次 PR 前）

```bash
# 1. 检查 spec drift
codegraph diff --against ARCHITECTURE.md  # TODO: 待实现

# 2. 检查 failure-patterns 是否更新
git diff --name-only HEAD~1 | grep -q 'failure-patterns' || \
  echo "⚠️ 本次修复是否产生了新的 failure pattern?"

# 3. 检查 ADR traceability
grep -r 'ADR-' cmd/ internal/ | wc -l  # 期望：每个 ADR 至少一处引用
```


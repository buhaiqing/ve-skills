# T01 — Execution Risk Policy (prose spec)

> 任务来源：[`../l2-to-l3-plan.md`](../l2-to-l3-plan.md) §3 + §4 (P1)
> 依赖：无
> 可并行：T03, T04
> 预计工作量：0.5 天
> 状态：🟡 TODO

## 1. 目标（Goal）

为 `incident-loop-agent` 编写**人类可读**的执行风险策略文档，
把 L2 的"破坏性一律 `{{user.confirm}}`"硬门，
替换为 L3 的 `risk × blast_radius × confidence → AUTO/ASK/REFUSE` 三维分级。

## 2. 背景与上下文

- L2 卡点：`incident-loop-agent/SKILL.md:150-154,184`
- 决策矩阵详见 [`../l2-to-l3-plan.md`](../l2-to-l3-plan.md) §3.2
- 这是**后续 7 张卡片**的根，所有机器可读 schema / 接入 / 验证都基于此文档。

## 3. 产出物（Deliverable）

**新文件**：`incident-loop-agent/references/policies/execution-risk.md`

### 3.1 必须包含的章节

| § | 标题 | 必须出现的内容 |
|---|------|---------------|
| 0 | Purpose | 一句话：本策略替代 `{{user.confirm}}` 硬门 |
| 1 | Three scoring dimensions | `risk` / `blast_radius` / `confidence`，每项给出值域 + 来源 |
| 2 | Decision matrix | **9 cells** 的 `AUTO/ASK/REFUSE` 表（含文字说明） |
| 3 | Hard safety floor | 显式声明："Safety = 0 → REFUSE，覆盖所有其他规则" |
| 4 | Policy strictness | 显式声明："本策略严格于旧 L2 默认值：旧默认 REFUSE，本策略对低风险子集放开为 AUTO" |
| 5 | Decision logic pseudocode | ~10 行伪代码，描述如何把 `(risk, blast_radius, confidence, safety)` 映射到决策 |
| 6 | Failure modes | 至少列出 3 个边界：metadata 缺失、confidence 中等、blast_radius 未注明 |
| 7 | References | 反向链回 `l2-to-l3-plan.md §3` |

### 3.2 内容约束（token efficiency / TE-6）

- 表格 ≤ 3 列（按 `docs/token-efficiency.md` TE-3）
- 决策矩阵符号化：用 `→` `✅` `❌` 替代冗余文字（TE-8）
- 不重复 `l2-to-l3-plan.md §3.2` 的矩阵——用**链接**替代（TE-6）
- 行数 ≤ 100

## 4. DoD（Definition of Done）

```
□ 1. 文档已写入 incident-loop-agent/references/policies/execution-risk.md
□ 2. §1 三维定义值域和来源明确
□ 3. §2 决策矩阵覆盖 9 cells（含 0/0/0 的边界 + destructive 的 3 cells）
□ 4. §3 显式声明 Safety=0 hard floor
□ 5. §5 pseudocode 与 §2 表格自洽
□ 6. §6 failure modes 至少 3 条
□ 7. 行数 ≤ 100
□ 8. 顶部 1 句话 Purpose 明确"L2→L3 替换硬门"
□ 9. TE-3/6/8 检查通过（自查）
□ 10. `incident-loop-agent/SKILL.md` 的 `## References` 小节已新增 `execution-risk.md` 路径
□ 11. `incident-loop-agent/SKILL.md` 的 `Changelog` 已追加版本条目（版本/日期/变更说明）
□ 12. ledger 已登记（格式：`## T01 YYYY-MM-DD — done` + 交付物一句话）
```

## 5. 验证命令

无 Go 代码改动，仅 markdown：

```bash
# 1. 文件存在
test -f incident-loop-agent/references/policies/execution-risk.md && echo "FILE_OK"

# 2. 必含章节存在
grep -q "^## 2. Decision matrix" incident-loop-agent/references/policies/execution-risk.md && echo "SECTION_OK"
grep -q "Safety = 0" incident-loop-agent/references/policies/execution-risk.md && echo "SAFETY_FLOOR_OK"

# 3. 行数 ≤ 100
awk 'END{ if (NR<=100) print "LENGTH_OK"; else print "LENGTH_FAIL "NR }' \
  incident-loop-agent/references/policies/execution-risk.md

# 4. 全局门禁仍过（确保没破坏其它 skill）
cd cmd/vet && go build ./... && go vet ./...
```

## 6. 完成回报

完成后在 [`_trace/ledger.md`](./_trace/ledger.md) 追加：

```markdown
## T01 2026-07-XX — done
- 交付：`incident-loop-agent/references/policies/execution-risk.md`（XX 行）
- 9 cells 决策矩阵已对齐 plan §3.2
- Safety=0 hard floor 显式声明
- T02/T03 可解锁
```

并在 [01-index.md](./01-index.md) 把 T01 状态改成 `✅ DONE`。

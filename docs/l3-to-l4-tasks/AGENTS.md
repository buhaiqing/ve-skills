# AGENTS.md — L3→L4 任务目录

> 本目录是 `incident-loop-agent` 从 **L3（条件自主）升级到 L4（高度自主）** 的执行工作区。
> 本文件是 AI Agent 的入口规则，补充而非替代仓库根的 `AGENTS.md` 和 `l2-to-l3-tasks/AGENTS.md`。
>
> **上级文件**：`../../AGENTS.md`（仓库级）、[`../l2-to-l3-tasks/AGENTS.md`](../l2-to-l3-tasks/AGENTS.md)（L2→L3 规则）
>
> **前提**：L2→L3 全部 8 张卡片必须 ✅ DONE 后方可开始本目录任务。
>
> **最后更新**：2026-07-13

---

## Karpathy 行为准则（核心，不可绕过）

> 来源：Andrej Karpathy 关于 LLM 编码陷阱的指导。本目录所有 Agent 必须内化并执行。

### 0. 能力边界（Capability Boundaries）

**不知道就承认不知道，不要假装知道。**

在接受任务或回答问题之前，先评估自己是否具备所需信息：
- 如果问题的领域在你的训练数据之外（特定内部工具、私有 API、未公开版本、非公开仓库），**必须**先声明"该信息可能不准确，需要查证"，再尝试回答。
- 永远不要假装读过你没有读过的文件。如果你知道一个文件存在但没读过，先去读它。
- 不要对未读的代码库结构做假设。说"可能有个 config.py"之前，先去验证它是否存在。
- 如果任务所需的技能/工具不在你当前可用的工具列表中，直接说明能力边界，而不是假装能完成然后编造结果。
- 当用户给出模糊或过于宽泛的需求时，先列出需要澄清的问题，再开始执行——不要自己脑补缺失的细节。

### 1. 编码前先思考（Think Before Coding）

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

### 2. 最简化优先（Simplicity First）

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

### 3. 外科手术式改动（Surgical Changes）

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

### 4. 目标驱动执行（Goal-Driven Execution）

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

---

## 本目录的 Karpathy §0 具体边界（在核心准则之上的扩展）

### X.1 我能做的（In Scope）

- 读写本目录下的任务卡片（`T09.md` ~ `T16.md`）
- 按卡片 §产出 产文件、按 §DoD 自验
- 在 `_trace/ledger.md` 登记进度
- 在仓库根 `cmd/vet/internal/` 写新 Go 包（遵守 All Tools MUST Be Go）
- 扩展 `incident-loop-agent/references/policies/` 下的 policy 文件
- 扩展 `docs/skill-routing-graph.md`（T12 预测触发）
- 修改 `docs/goals-dashboard-spec.md`（T16）
- 修改 `.github/workflows/` CI 配置
- 读取并使用 `cmd/vet/internal/` 下任何现有包

### X.2 我不能做的（Out of Scope）

- **不修改 `docs/autonomous-ops-roadmap.md` M1 部分**（那是 L2→L3 的已完成范围）
- **不修改 `docs/l2-to-l3-plan.md`**（那是 L2→L3 的已完成规划）
- **不在 L4 envelope 中引入 destructive operation 的 AUTO 路径**（L4 仍遵守 destructive 全 ASK 原则）
- **不在 Go 代码中引入新的外部依赖**（go.mod 变更须先有人工批准）
- **不执行任何真实的 `ve` CLI 调用**
- **不绕过 Safety=0 硬门**（L4 的 policy gate 不改变 Safety=0 强制 ABORT 的性质）
- **不在 eval / test fixture 中写入真实凭证**
- **不在 L4 envelope 外擅自宣称 L4 能力**（envelope 外的系统仍是 L2/L3）
- **不在未完成 T09/T10/T11 的情况下声称具备"自愈能力 L3"**
- **不修改其他 Agent 的任务卡片**（除非被明确授权）
- **不在本规则未覆盖时假装能力**（遵循 Karpathy §0）

### X.3 自我评估模板（每次领卡前执行 — Karpathy §0 实证）

```
任务：Txx — <名称>
前置条件检查：
□ L2→L3 全部 T01–T08 已 ✅ DONE
□ 本卡的前置依赖（Txx）已 ✅ DONE
□ 我理解 L3 vs L4 的差异（envelope / SLO / auto-rollback / 预测触发）
□ 我知道本卡是 M2/M3/M4 的哪个阶段
□ 我知道 TDD 要求（§TDD）
□ 我知道 GCL 要求（§GCL）
□ 我知道 All Tools MUST Be Go（§工具）
□ 我知道完成后必须更新相关文档（§文档）
□ 我没有跨越 X.2 的边界
结论：可以开始 / 需要先完成前置任务
```

### X.4 目标驱动执行模板（Karpathy §4 实证）

```
## 目标
<来自 §1 目标的一句话>

## L3 vs L4 区别（执行前必读）
L3: bounded domain, AUTO for read-only, ASK for destructive, 0 human prompt in happy path
L4: bounded domain, AUTO for state-changing + high-confidence, SLO-driven, auto-rollback, predictive trigger

## 关键结果（KR）
KR1: <可验证的产出>
KR2: <TDD 红灯先亮>
KR3: <自愈指标（成功率 >80%、<30s）已定义>
KR4: <文档已更新>

## 与 L2→L3 的差异（必读）
□ 本卡是否涉及新的 vet 子命令？若是 → 同 AGENTS.md L2→L3 §5
□ 本卡是否涉及 GCL 改动？若是 → 同 AGENTS.md L2→L3 §4
□ 本卡是否涉及 safety 不变量？若是 → Safety=0 → REFUSE 仍为硬门
□ 本卡是否涉及 L4 envelope？若是 → destructive 全 ASK 原则不变

## Checkpoint
- [ ] Red: 失败测试已写
- [ ] Green: 测试通过
- [ ] Refactor: 代码可读、无重复
- [ ] 自愈指标定义（成功率 >80% / <30s）已写入 metrics.go
- [ ] L4 envelope 定义（autonomy-envelope.md）已更新
- [ ] 文档已更新
- [ ] ledger 已登记
```

---

## 本目录的工程特定规则（Karpathy 准则在仓库级 AGENTS.md 的具体化）

### T1. TDD 铁律（强制）— 对应 Karpathy §4

> **所有 Go 实现卡片（T09 / T10 / T11 / T12 / T13 / T14 / T15 / T16）必须严格执行 TDD。**

```
Red    → 写一个必然失败的测试（函数未实现，编译不过或断言红）
Green  → 写最小实现让测试通过
Refactor → 消除重复、提升可读性，不改变行为
```

**L3→L4 额外的 TDD 要求**（自愈指标 + envelope）：

- 每个自愈指标（成功率、平均时间）必须有**边界测试**（TotalCount=0、全部失败、全部成功）
- 每个 SLO 状态转换必须有**状态机测试**（Healthy→Warning→Critical→Violated）
- 每个 envelope 边界条件（blast_radius cap）必须有**违反测试**
- 每个预测器的触发/不触发阈值必须有**边界测试**（斜率=0、刚好=阈值、超过阈值）

**禁止**：
- ❌ 先实现后补测试
- ❌ 测试覆盖率 < 100%（每个分支路径至少一个 case）
- ❌ 用 `fmt.Printf` 代替断言
- ❌ 用 `time.Sleep` 代替 channel/condition 同步

### T2. GCL 规范（强制）— 对应 Karpathy §3

> 所有涉及 `vet gcl run` 的改动必须遵循 [`../../docs/gcl-spec.md`](../../docs/gcl-spec.md)。

| 规则 | 说明 |
|------|------|
| **Generator / Critic 隔离** | G 和 C 必须在孤立 prompt context 中 |
| **Safety=0 → ABORT** | L4 的 policy gate 不改变 Safety=0 强制 ABORT 性质 |
| **Trace 持久化** | 自愈、预测触发、rollback 所有决策必须入 trace |
| **凭证脱敏** | `<masked>` 强制 |
| **max_iter 强制** | 破坏性 op → max_iter=2；其余 → max_iter=3 |

**L3→L4 GCL 额外检查项**：
```
□ 自愈决策（path 选择）已记录到 trace.self_healing
□ 预测触发已记录到 trace.trigger_source（reactive|predictive）
□ rollback 已应用 trace.rollback_applied=true/false
□ SLO 状态变化已记录到 trace.slo_status
```

### T3. All Tools MUST Be Go（强制）— 对应 Karpathy §2

> **本仓库所有可执行工具/CLI/自动化脚本必须用 Go 实现。**
> 详见 [`../../AGENTS.md`](../../AGENTS.md) §MANDATORY。

| 门禁 | 验证命令 |
|------|---------|
| Compile | `cd cmd/vet && go build ./...` |
| Static check | `cd cmd/vet && go vet ./...` |
| Test | `cd cmd/vet && go test ./...` |
| Publishable | `make build` 产出二进制 + `make vet` |

**禁止**：Python / Bash / shell 编写新工具。

### T4. L4 专项规则 — 对应 Karpathy §0（不假装 L4）

#### T4.1 Envelope 规则

- L4 envelope 定义了哪些 (skill, symptom, blast_radius) 可以进入 AUTO 决策
- **destructive operation 永远不在 AUTO 类**（L4 envelope 也不例外）
- envelope 初始很窄（2~3 个 symptom），按 M3-4 的"扩域条件"扩展
- envelope 内的决策仍然必须经过 GCL Safety=1 floor

#### T4.2 SLO 规则

- SLO 是 loop 的**目标函数**，不是约束
- SLO 违反 → 系统采取行动维持 SLO（predictive trigger、auto-rollback）
- SLO 目标初始值 = 当前基线 + 10%（保守）

#### T4.3 自愈规则

| 等级 | 要求 | 指标 |
|------|------|------|
| L1 | 固定次数重试 | — |
| L2 | 错误分类 → 针对性重试 | — |
| **L3** | **多路径自愈 + 成功率 >80% / <30s** | **本目录目标** |
| L4 | 预防性自愈（预测触发） | — |
| L5 | 自学习自愈 | — |

#### T4.4 Reflexion 升级规则（count 分级）

| count 范围 | 级别 | 行为 |
|-----------|------|------|
| < 3 | Pruned | 不记录 |
| 3 ≤ count < 10 | HINT | 注入 context，**不强制** |
| **10 ≤ count < 30** | **Constraint** | **policy 必须遵守** |
| ≥ 30 | Hard | 直接 ABORT，强制 human review |

### T5. 凭证安全（强制）

- trace / 报告中的凭证一律 `<masked>`
- 测试 fixture 中用 `<masked>` 占位，不用真实密钥
- 自愈日志（`/tmp/ve-self-healing.log`）不记录凭证

### T6. Token Efficiency（TE）规则 — 对应 Karpathy §2

> 详见 [`../../docs/token-efficiency.md`](../../docs/token-efficiency.md)。

| TE | 要求 |
|----|------|
| TE-3 | 表格 ≤ 3 列 |
| TE-6 | 不跨文件重复内容（用链接替代） |
| TE-8 | 用 `→` `✅` `❌` 符号替代冗余文字 |

### T7. 文档更新规则（强制）— 对应 Karpathy §3

> **每个任务完成后必须更新相关文档。**

| 产出类型 | 必须更新的文档 | 更新内容 |
|---------|-------------|---------|
| 新 Go 包 | 所在模块 + `cmd/vet/` 的 `go.mod`（如有新依赖） | — |
| 新 JSON Schema | 对应 `assets/` 或 `references/` | schema 路径引用 |
| 新 policy 文件 | `incident-loop-agent/SKILL.md` 的 `## References` 小节 | 新增引用路径 |
| 新自愈指标 | `enhanced-self-healing-framework.md` §6.1 | 指标定义更新 |
| 新预测触发器 | `docs/skill-routing-graph.md` §4 | predictive 行加入 |
| 新 guardrail | `incident-loop-agent/references/policies/CHANGELOG.md` | 版本 + 变更说明 |
| CI 新步骤 | `.github/workflows/validate.yml` | 新子命令加入 CI |
| 新 eval case | `assets/eval_queries.json` 的 `last_updated` | 版本更新 |

**禁止**：
- ❌ 只改代码不更新文档
- ❌ 更新了文档但未在 DoD 打勾
- ❌ 在 ledger 登记完成但文档未同步
- ❌ 在未完成 T09~T11 时声称"具备 L3 自愈能力"

### T8. 进度登记规则 — 对应 Karpathy §4（验收闭环）

- 完成一条 → 在 `_trace/ledger.md` 追加**一条**
- 禁止预填、禁止覆盖历史
- 格式：`## Txx YYYY-MM-DD — done` + 交付物一句话描述 + 版本/日期
- 登记是 DoD 的最后一项，不登记 = 未完成

### T9. 异常处理规则 — 对应 Karpathy §1（暴露混乱）

| 情况 | 处理方式 |
|------|---------|
| 自愈路径全失败 | 降级到 human review；不静默通过 |
| 预测器误报 | 计入"误报率"指标；连续 3 次误报 → 自动降级为 reactive 模式 |
| envelope 内的 op 被 ABORT | 按 policy 记录到 trace；human review 后解除 |
| SLO 持续违反 | 触发 M4-4 emergency override；人介入设策略 |
| 前置任务未完成 | 停在依赖检查，不自行跨越 |
| 任务卡片的某条 DoD 解读模糊 | 暂停，报告用户，不自行解释（Karpathy §1） |

---

## 相关文件索引

| 文件 | 作用 |
|------|------|
| [`01-index.md`](./01-index.md) | 任务一览 + 依赖图 + 调度 |
| [`_trace/ledger.md`](./_trace/ledger.md) | 进度台账（审计 trail） |
| [`../l2-to-l3-tasks/AGENTS.md`](../l2-to-l3-tasks/AGENTS.md) | L2→L3 规则（先行完成条件） |
| [`../autonomous-ops-roadmap.md`](../autonomous-ops-roadmap.md) | 完整 L2→L4 路线图 |
| [`../../AGENTS.md`](../../AGENTS.md) | 仓库级规则 |
| [`../../docs/gcl-spec.md`](../../docs/gcl-spec.md) | GCL 规范 |
| [`../../docs/token-efficiency.md`](../../docs/token-efficiency.md) | TE 规则 |
| [`../../incident-loop-agent/SKILL.md`](../../incident-loop-agent/SKILL.md) | 编排器技能 |
| [`../../incident-loop-agent/references/policies/`](../../incident-loop-agent/references/policies/) | Policy 文件目录 |

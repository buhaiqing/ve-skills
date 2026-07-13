# AGENTS.md — L2→L3 任务目录

> 本目录是 `incident-loop-agent` 从 **L2（人工确认执行）升级到 L3（条件自主）** 的执行工作区。
> 本文件是 AI Agent 的入口规则，补充而非替代仓库根的 `AGENTS.md`。
>
> **上级文件**：`../../AGENTS.md`（仓库级规则）、`../l2-to-l3-plan.md`（L2→L3 完整规划）
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

- 读写本目录下的任务卡片（`T*.md`）
- 按卡片 §产出 产文件、按 §DoD 自验
- 在 `_trace/ledger.md` 登记进度
- 在仓库根 `cmd/vet/` 写 Go 代码（按 All Tools MUST Be Go 规则）
- 修改 `incident-loop-agent/SKILL.md` 及其他 `ve-*-ops/SKILL.md`
- 修改 `.github/workflows/` CI 配置
- 修改 `docs/` 下的文档（遵守 TE 规则）
- 读取并使用 `incident-loop-agent/references/policies/` 下任何文件

### X.2 我不能做的（Out of Scope）

- **不创建新的 `ve-*-ops/` 目录**（由 `ve-skill-generator` 元技能生成）
- **不修改 `docs/autonomous-ops-roadmap.md`**（规划文档，由人工或专项任务更新）
- **不修改 `docs/l2-to-l3-plan.md` 的 M2/M3/M4 部分**（那是 L3→L4 的范围）
- **不在 Go 代码中引入新的外部依赖**（go.mod 变更须先有人工批准）
- **不执行任何真实的 `ve` CLI 调用**（只生成命令，不执行）
- **不绕过 Safety=0 硬门**：任何 Safety=0 的操作必须 ABORT，本规则不可违反
- **不在 eval case 中写入真实凭证或密钥**
- **不修改其他 Agent 的任务卡片**（除非被明确授权）
- **不在本规则未覆盖时假装能力**（遵循 Karpathy §0）

### X.3 自我评估模板（每次领卡前执行 — Karpathy §0 实证）

```
任务：Txx — <名称>
自我评估：
□ 我理解目标（§1），知道交付物是什么
□ 我理解 DoD（§4），知道验收标准
□ 我理解依赖（§依赖），知道前置条件是否已满足
□ 我知道 TDD 要求（§TDD），Red→Green→Refactor
□ 我知道 GCL 要求（§GCL），G/C 隔离、Safety=0 必 ABORT
□ 我知道 All Tools MUST Be Go（§工具）
□ 我知道完成后必须更新相关文档（§文档）
□ 我没有跨越 X.2 的边界
结论：可以开始 / 需要先完成前置任务
```

### X.4 目标驱动执行模板（Karpathy §4 实证）

```
## 目标
<来自 §1 目标的一句话>

## 关键结果（KR）
KR1: <可验证的产出>
KR2: <TDD 红灯先亮>
KR3: <文档已更新>

## 障碍与对策
| 障碍 | 对策 |
|------|------|
| 行号漂移 | 用 grep 重新定位，不用硬编码行号 |
| 跨文件改动 | 每次只改一个文件，逐文件验证 |
| Go vet 报错 | 先看 vet 输出，再定位根因 |
| TE 违规 | 用符号替代冗余文字（TE-8）|

## Checkpoint
- [ ] Red: 失败测试已写
- [ ] Green: 测试通过
- [ ] Refactor: 代码可读、无重复
- [ ] 文档已更新
- [ ] ledger 已登记
```

---

## 本目录的工程特定规则（Karpathy 准则在仓库级 AGENTS.md 的具体化）

### T1. TDD 铁律（强制）— 对应 Karpathy §4

> **所有 Go 实现卡片（T06 / T07 / T08）必须严格执行 TDD。**

```
Red    → 写一个**必然失败**的测试（函数未实现，编译不过或断言红）
Green  → 写最小实现让测试通过
Refactor → 消除重复、提升可读性，不改变行为
```

**禁止**：
- ❌ 先实现后补测试
- ❌ 测试覆盖率 < 100%（每个分支路径至少一个 case）
- ❌ 用 `fmt.Printf` 或日志代替断言
- ❌ 用 `time.Sleep` 代替 channel/condition 同步

**强制**：
- ✅ `go test -v -run TestXxx` 可见 Red/Green 过程
- ✅ 每个不变量（invariant）单独一个 `TestXxxInvariant*` 测试
- ✅ 每个边界条件单独一个 `TestXxxEdgeCase*` 测试

### T2. GCL 规范（强制）— 对应 Karpathy §3

> 所有涉及 `vet gcl run` 的改动必须遵循 [`../../docs/gcl-spec.md`](../../docs/gcl-spec.md)。

| 规则 | 说明 |
|------|------|
| **Generator / Critic 隔离** | G 和 C 必须在**孤立 prompt context**中；不得共享内存/状态 |
| **Safety=0 → ABORT** | 任何 GCL 运行中 Safety=0 必须立即 ABORT，不得返回部分结果 |
| **Trace 持久化** | 每次 GCL 运行必须产出 trace（含 RequestId），存到 `audit-results/` |
| **凭证脱敏** | trace 中所有凭证必须 `<masked>`，不得明文 |
| **max_iter 强制** | `max_iter` 是硬约束，超限返回 best-so-far + unresolved items |

**GCL 专有 DoD 检查项**（T06/T07/T08 必填）：
```
□ Generator prompt 不含 Critic 评分逻辑
□ Critic prompt 不含执行命令
□ Safety=0 运行结果 trace 中 Safety=0 已标注
□ 所有 trace 含 RequestId 且凭证已脱敏
□ max_iter 超限 trace 包含 unresolved_items
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

**禁止**：Python / Bash / shell 编写新工具。`scripts/*.py` 仅为参考实现，不再扩展。

### T4. 凭证安全（强制）

- trace / 报告中的凭证一律 `<masked>`
- 不在日志/trace/临时文件中写入明文凭证
- 用 `cmd/vet/internal/check/assessment` 脱敏路径验证
- 测试 fixture 中用 `<masked>` 占位，不用真实密钥

### T5. Token Efficiency（TE）规则 — 对应 Karpathy §2

> 详见 [`../../docs/token-efficiency.md`](../../docs/token-efficiency.md)。任务卡片产出的文档必须遵守。

| TE | 要求 |
|----|------|
| TE-3 | 表格 ≤ 3 列 |
| TE-6 | 不跨文件重复内容（用链接替代） |
| TE-8 | 用 `→` `✅` `❌` 符号替代冗余文字 |
| TE-9 | 按文档用途选压缩级别（Minimal~Emergency） |

### T6. 文档更新规则（强制）— 对应 Karpathy §3

> **每个任务完成后必须更新相关文档。**

| 产出类型 | 必须更新的文档 | 更新内容 |
|---------|-------------|---------|
| 新 `SKILL.md` 改动 | 对应 skill 的 `Changelog` | 版本 + 日期 + 一句话变更 |
| 新 JSON Schema | 对应 skill 的 `references/` 或 `assets/` | schema 路径引用 |
| 新 Go 包 | 所在模块的 `go.mod`（如有新依赖） | — |
| CI 新步骤 | `.github/workflows/validate.yml` | 新子命令加入 CI |
| 新 policy 文件 | `incident-loop-agent/SKILL.md` 的 `## References` 小节 | 新增引用路径 |
| 新 eval case | `assets/eval_queries.json` 的 `last_updated` | 版本更新 |

**禁止**：
- ❌ 只改代码不更新文档
- ❌ 更新了文档但未在 DoD 打勾
- ❌ 在 ledger 登记完成但文档未同步

### T7. 进度登记规则 — 对应 Karpathy §4（验收闭环）

- 完成一条 → 在 `_trace/ledger.md` 追加**一条**
- 禁止预填、禁止覆盖历史
- 登记格式：`## Txx YYYY-MM-DD — done` + 交付物一句话描述 + 版本/日期
- 登记是 DoD 的最后一项，不登记 = 未完成

### T8. 异常处理规则 — 对应 Karpathy §1（暴露混乱）

| 情况 | 处理方式 |
|------|---------|
| grep 找不到锚点行号 | 用更大的上下文（`### Step 5`）或关键字搜索，不死等行号 |
| Go vet 报错 | 先看 `go vet` 输出定位文件，再读文件；不猜 |
| 多文件同时改 | 每次只改一个文件，逐文件跑 `go build` 验证后再改下一个 |
| 前置任务未完成 | 停在依赖检查，不自行跨越；向用户报告 |
| 发现 P0 安全问题 | 立即停止并报告，不尝试自动修复 |
| 行号漂移 | 用 grep 重新定位，不硬依赖行号 |
| 任务卡片的某条 DoD 解读模糊 | 暂停，报告用户，不自行解释（Karpathy §1） |

---

## 相关文件索引

| 文件 | 作用 |
|------|------|
| [`01-index.md`](./01-index.md) | 任务一览 + 依赖图 + 调度 |
| [`_trace/ledger.md`](./_trace/ledger.md) | 进度台账（审计 trail） |
| [`../l2-to-l3-plan.md`](../l2-to-l3-plan.md) | 完整 L2→L3 规划 |
| [`../../AGENTS.md`](../../AGENTS.md) | 仓库级规则 |
| [`../../docs/gcl-spec.md`](../../docs/gcl-spec.md) | GCL 规范 |
| [`../../docs/token-efficiency.md`](../../docs/token-efficiency.md) | TE 规则 |
| [`../../cmd/vet/`](../../cmd/vet/) | Go 工具源码 |
| [`../../incident-loop-agent/SKILL.md`](../../incident-loop-agent/SKILL.md) | 编排器技能 |
